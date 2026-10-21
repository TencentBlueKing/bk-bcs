# 周边系统集成指南（bcs-k8s-watch 与 CMDB 容器拓扑同步）

蓝鲸容器管理平台（BCS）不仅提供面向终端用户的控制台和 API，还通过 **`bcs-k8s-watch` + `bcs-storage`（RabbitMQ 消息总线）** 以及 **`bcs-bkcmdb-synchronizer`（CMDB 容器拓扑同步器）** 将集群内的容器资源实时开放给蓝鲸配置平台（CMDB）及企业内部第三方周边系统消费。

---

## 一、容器拓扑数据同步至蓝鲸 CMDB（`bcs-bkcmdb-synchronizer`）

为了使周边运维系统能够在蓝鲸配置平台（CMDB）中直接查询和消费容器拓扑数据，BCS 提供了 `bcs-bkcmdb-synchronizer` 服务，对接 `bcs-storage` 的实时 Watch 事件与全量接口，将项目下集群、节点、命名空间、工作负载及 Pod 实例同步至 CMDB 3.0+ 的容器拓扑模型中。

### 1. 同步机制与工作原理
- **实时差量 + 周期全量对账**：
  - 实时监听 `bcs-storage` 的资源变更事件流，按 Pod UUID 哈希保序同步至 CMDB。
  - 默认每隔 `600 秒`（`full_sync_interval`）执行一次集群全量数据对账同步，确保最终一致性。
- **分布式高可用分片**：
  - 多实例通过 ZooKeeper 选主，采用一致性 Hash 算法自动分配各实例负责同步的集群列表。

### 2. 业务 Pod 关联 CMDB 集群（Set）与模块（Module）规则

当业务需要将特定 Pod 挂载到 CMDB 传统业务拓扑的指定 **集群（Set）** 和 **模块（Module）** 下时，可在工作负载的 Pod Template `metadata.annotations` 中添加以下两个注解：

```yaml
metadata:
  annotations:
    # 指定关联的 CMDB 集群（Set）名称
    set.bkcmdb.bkbcs.tencent.com: "game-zone-1"
    # 指定关联的 CMDB 模块（Module）名称
    module.bkcmdb.bkbcs.tencent.com: "login-server"
```

- **默认兜底规则**：若 Pod 未配置上述注解，或注解指定的 Set/Module 在 CMDB 中未同时匹配上，则该 Pod 默认归入名为 `Set: bkbcs`、`Module: bkbcs` 的默认模块下（建议业务在 CMDB 对应业务下预先建立名为 `bkbcs` 的 Set 和 Module）。

---

## 二、通过 RabbitMQ 订阅消费 `bcs-k8s-watch` 资源变更事件

周边系统若需实时感知 BCS 集群中的 Pod、Event、Deployment、Service 等资源变更，可直接订阅 `bcs-storage` 写入 **RabbitMQ** 的 `bcs-storage` 交换机（Exchange）。

### 1. 数据链路与前置配置检查

```mermaid
flowchart LR
    K8S["被纳管 K8S 集群<br/>bcs-k8s-watch"] -->|根据 filter.json 过滤采集| ST["BCS 控制面<br/>bcs-storage"]
    ST -->|写入持久化| MONGO[("MongoDB")]
    ST -->|根据 queue.conf 推送| MQ["RabbitMQ Exchange<br/>name: bcs-storage (topic)"]
    MQ -->|routing_key = 资源类型| Q["周边系统自定义队列<br/>Queue"]
```

**只有同时满足以下两处配置，特定资源类型的事件才会推送至 RabbitMQ**：
1. **集群侧 `bcs-k8s-watch` 的 `filter.json`**：目标资源 GVK 已包含在 `k8sResourceWhiteList` 或 `crdResourceWhiteList` 中（若需采集 CRD，还需确保 `bcs-k8s-watch.json` 中 `watch_resource.disable_crd` 为 `false`）。
2. **服务端 `bcs-storage` 的 `queue.conf`**：在 `[queue]` 段的 `Resource` 列表中包含该资源类型：
   ```ini
   [queue]
   Resource = Pod,Event,Service,Namespace,Node,Deployment,StatefulSet,DaemonSet,GameStatefulSet,GameDeployment,taskgroup
   ```

### 2. 获取 RabbitMQ 连接地址

在 BCS 服务端集群 Master 上可通过以下命令获取 RabbitMQ 连接串：

```bash
kubectl get secret bcs-password -n bcs-system -o jsonpath='{.data.messagequeue_address}' | base64 --decode
# 输出格式示例：amqp://user:password@bcs-rabbitmq.bcs-system.svc.cluster.local:5672
```
（开发调试时可使用 `kubectl port-forward svc/bcs-rabbitmq 5672:5672 15672:15672 -n bcs-system` 映射至本地访问）。

### 3. 消息路由与 Header 结构说明

- **交换机信息**：名称为 `bcs-storage`，类型为 **`topic`**，持久化 `durable: true`。
- **Routing Key**：固定对应资源类型名称（`resourceType`），如 `Pod`、`Event`、`Deployment`、`Service`、`GameStatefulSet` 等。
- **消息 Headers 格式**：每条消息的 AMQP `headers` 中均包含以下标准元数据，可用于在客户端或通过中间 `headers` 类型交换机进一步按集群和命名空间过滤：

```json
{
  "headers": {
    "clusterId": "BCS-K8S-00000",
    "event": "update",
    "namespace": "bcs-system",
    "resourceName": "bcs-rabbitmq-0",
    "resourceType": "Pod"
  }
}
```

### 4. 客户端接入规范与 Python 消费示例

> **生产接入强制规范**：周边系统在声明自身消费队列（`QueueDeclare`）时，**必须设置队列最大长度限制（`x-max-length`）与最大字节数限制（`x-max-length-bytes`）**，防止消费者异常离线时消息无限堆积打满 BCS RabbitMQ 磁盘！

```python
import json
import signal
import sys
import pika

# 1. 配置订阅的资源类型（routing_key）与 Header 过滤条件
RESOURCE_TYPES = ['Event', 'Pod']
HEADER_FILTERS = {
    'clusterId': ['BCS-K8S-00000'],
    'namespace': ['default', 'bcs-system'],
}
AMQP_URL = 'amqp://username:password@localhost:5672'
EXCHANGE_NAME = 'bcs-storage'
QUEUE_NAME = 'bkapp-custom-watcher'


def message_handler(channel, method, properties, body):
    # 按 headers 过滤目标集群与命名空间
    if properties and properties.headers:
        for key, allowed_values in HEADER_FILTERS.items():
            if allowed_values and properties.headers.get(key) not in allowed_values:
                return
    try:
        payload = json.loads(body)
    except json.JSONDecodeError:
        payload = body.decode('utf-8')
    print(json.dumps({'headers': properties.headers, 'body': payload}, indent=2, ensure_ascii=False))


def main():
    connection = pika.BlockingConnection(pika.URLParameters(AMQP_URL))
    channel = connection.channel()
    channel.exchange_declare(exchange=EXCHANGE_NAME, exchange_type='topic', durable=True)

    # 声明队列并设置容量上限保护
    queue_args = {
        'x-max-length': 1000,
        'x-max-length-bytes': 500 * 1024 * 1024,  # 500MB
    }
    q = channel.queue_declare(queue=QUEUE_NAME, durable=True, exclusive=False, auto_delete=False, arguments=queue_args)

    for rt in RESOURCE_TYPES:
        channel.queue_bind(exchange=EXCHANGE_NAME, queue=q.method.queue, routing_key=rt)

    channel.basic_consume(queue=q.method.queue, on_message_callback=message_handler, auto_ack=True)
    signal.signal(signal.SIGINT, lambda s, f: (channel.close(), connection.close(), sys.exit(0)))
    channel.start_consuming()


if __name__ == '__main__':
    main()
```
