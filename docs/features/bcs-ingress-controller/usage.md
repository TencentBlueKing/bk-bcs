# bcs-ingress-controller使用

## 特性支持

* 支持HTTPS，HTTP，TCP，UDP协议
* 支持腾讯云Clb健康检查等参数配置
* 支持单个ingress同时控制多个clb实例
* 支持转发到NodePort模式和转发到直通Pod模式
* 支持单端口多Service流量转发，以及WRR负载均衡方法下权重配比
* 直通Pod模式下，支持Service内部通过Label选择Pod，以及WRR负载均衡方法下权重配比
* 支持StatefulSet和GameStatefulSet端口段映射
* 云接口的客户端限流与重试

## 启动bcs-ingress-controller

```shell
# 腾讯云clb的api接口
export TENCENTCLOUD_CLB_DOMAIN="clb.xxxx.tencentcloudapi.com"
# 腾讯云云区域
export TENCENTCLOUD_REGION="ap-shenzhen"
# 腾讯云api的secret id
export TENCENTCLOUD_ACCESS_KEY_ID="AppSecretIDExample2efsdaasdf"
# 腾讯云api的secret key
export TENCENTCLOUD_ACESS_KEY="AppSecretKeyExamplewerdsafasdf"

./bcs-ingress-controller \
  # 云厂商, [tencentcloud, aws(coming soon)]
  --cloud tencentcloud \
  # 默认云区域
  --region ap-xxxxx \
  # 选注configmap所在namespace
  --election_namespace bcs-sytem \
  # kubeconfig文件路径, 不指定即使用InCluster模式
  --kubeconfig /root/.kube/config \
  # address，bcs-ingress-controller服务监听地址
  --address 127.0.0.1 \
  # metric_port metric port
  --metric_port 8081 \
  # log_dir blog日志存放文件夹
  --log_dir ./logs \
  # alsologtostderr 将日志同时打印在标准错误中
  --alsologtostderr \
  # v, 日志级别
  --v 3
```

## 不同场景下的配置实例

### 场景：clb转到service NodePort

背景

* 命名空间 **default** 下有名为 **test-svc** 的NodePort类型的Service
* service **test-svc** 中有端口号为8080的TCP端口，对应的NodePort为31003
* service **test-svc** 8080端口对应后端pod端口同样为8080
* clb实例的id是**lb-xxxxxxx**
* Overlay集群，从集群外无法访问集群内Pod的IP地址

想要达到的效果

* clb流量无法直接转发到Pod上，需要转到对应Service的NodePort上
* 需要对外暴露TCP 38080端口

```text
clb(38080 port) ------> service对应pod所在的node节点(31003 port) -------> pod(8080)
```

配置示例

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: Ingress
metadata:
  name: test1
  annotations:
    networkextension.bkbcs.tencent.com/lbids: lb-xxxxxxx
spec:
  rules:
  - port: 38080
    protocol: TCP
    services:
    - serviceName: test-svc
      serviceNamespace: default
      servicePort: 8080
```

### 场景：clb直通Pod

前提

* 命名空间 **default** 下有名为 **test-svc** 的NodePort类型的Service
* service **test-svc** 中有端口号为8080的TCP端口
* service **test-svc** 8080端口对应后端pod端口同样为8080
* Pod拥有underlay IP地址，从集群外可以访问集群内Pod的IP地址
* clb实例的id是 **lb-xxxxxxx**

想要达到的效果

* 假设集群是Underlay的，流量需要直通Pod
* 需要对外暴露TCP 38080端口

```text
clb(38080 port) ------> pod(8080)
```

配置示例

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: Ingress
metadata:
  name: test2
  annotations:
    networkextension.bkbcs.tencent.com/lbids: lb-xxxxxxx
spec:
  rules:
  - port: 38080
    protocol: TCP
    services:
    - serviceName: test-tcp
      serviceNamespace: default
      servicePort: 8080
      isDirectConnect: true
```

### 场景：clb对外暴露HTTPS（单向验证）协议443端口

前提

* 命名空间 **default** 下有名为 **test-svc** 的NodePort类型的Service
* service **test-svc** 中有端口号为8080的TCP端口，对应的NodePort为31003
* service **test-svc** 8080端口对应后端pod端口同样为8080
* Overlay集群，从集群外无法访问集群内Pod的IP地址
* 腾讯云上HTTPS证书ID为**cert-xxx**，单向验证
* clb示例的id是**lb-xxxxxxx**

想要达到的效果

* clb流量无法直接转发到Pod上，需要转到对应Service的NodePort上
* 需要对外暴露TCP 443端口

```text
clb(443 port)(www.qq.com)(/path1) ------> service对应pod所在的node节点(31003 port) -------> pod(8080)
```

配置示例

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: Ingress
metadata:
  name: test3
  annotations:
    networkextension.bkbcs.tencent.com/lbids: lb-xxxxxxx
spec:
  rules:
  - port: 443
    protocol: HTTPS
    certificate:
      mode: UNIDIRECTIONAL
      certID: cert-xxx
    layer7Routes:
    - domain: www.qq.com
      path: /path1
      services:
      - serviceName: test-tcp
        serviceNamespace: default
        servicePort: 8080
```

### 场景：StatefulSet端口段映射

前提

* 命名空间 **game** 下有名为 **gameserver** StatefulSet，有5个Pod
* **gameserver** 的每个Pod都需要单独直接对外提供UDP服务
* **gameserver** 每个Pod的端口号与Pod需要有一定对应关系
* **gameserver** 拥有underlay IP地址或者是Host模式
* clb实例的id是**lb-xxxxxxx**

想要达到的效果

* 端口映射
  * gameserver-0对应30000~30004端口
  * gameserver-1对应30005~30009端口
  * .....
  * gameserver-5对应30020~30024端口

配置示例

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: Ingress
metadata:
  name: test4
  annotations:
    networkextension.bkbcs.tencent.com/lbids: lb-xxxxxxx
spec:
  portMappings:
  - startPort: 30000
    protocol: UDP
    startIndex: 0
    endIndex: 6
    segmentLength: 5
    workloadKind: StatefulSet
    workloadName: gameserver
    workloadNamespace: game
```

### 场景：NodePort模式下，不同Service之间进行权重配比

前提

* 命名空间 **test** 下有名为 **svc1** 和 **svc2**的两个Service
* service **svc1** 和 **svc2** 中都有端口号为8080的TCP端口，**svc1** 对应的NodePort为31003， **svc2** 对应的NodePort为31004
* 80%的流量转给svc1，20%的流量转给svc2
* clb实例的id是**lb-xxxxxxx**

想要达到的效果

```text
clb(38080 port) -----80%---> service svc1 对应pod所在的node节点(31003 port) -----> svc1 pods(8080)
                  |--20%---> service svc2 对应pod所在的node节点(31004 port) -----> svc2 pod(8080)
```

配置示例

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: Ingress
metadata:
  name: test1
  annotations:
    networkextension.bkbcs.tencent.com/lbids: lb-xxxxxxx
spec:
  rules:
  - port: 38080
    protocol: TCP
    services:
    - serviceName: svc1
      serviceNamespace: default
      servicePort: 8080
      weight:
        value: 80
    - serviceName: svc2
      serviceNamespace: default
      servicePort: 8080
      weight:
        value: 20
```

## 更多参数解释

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: Ingress
metadata:
  # ingress的名字
  name: bcs-ingress1
  # ingress的命名空间
  namespace: test
  annotations:
    # 可以通过id和name来关联clb实例，优先匹配id，以下annotation二选一
    # 通过此annotation来关联多个clb实例，value的形式为"lb1-xxxxx,lb2-xxxxx,ap-xxxx:lb3-xxxxx"
    networkextension.bkbcs.tencent.com/lbids: lb-xxxxxx
    # 通过此annotation来关联多个clb实例，value的形式为"name1,name2,ap-xxx:name3"
    networkextension.bkbcs.tencent.com/lbnames: name1
spec:
  # rules是一个数组，每个数组元素定义单个端口到一个或者多个Service的映射
  rules:
    # rule定义的端口号，rule之间的端口号不能冲突
  - port: 1111
    # rule定义的端口协议，可选有[HTTP, HTTPS, TCP, UDP]
    protocol: TCP
    # TCP端口或者UDP端口映射的Service的数组
    services:
      # service的名字
    - serviceName: test-tcp
      # service的命名空间的名字
      serviceNamespace: default
      # service定义的service port的名字
      servicePort: 1111
      # 是否直通Pod，默认为False
      isDirectConnect: true
      # 属于该service的监听器后端的权重（只有负载均衡为WRR时才生效）
      weight:
        value: 10
      # TCP或者UDP监听器转发属性
      # 详情参考链接 https://cloud.tencent.com/document/product/214/30693
      listenerAttribute:
      # 会话保持时间，单位：秒。可选值：30~3600，默认 0，表示不开启。
      sessionTime: 100
      # 负载均衡策略：可选值：WRR, LEAST_CONN
      lbPolicy: WRR
      # 健康检查策略
      healthCheck:
        # 是否开启健康检查，如果不设置该选项，默认开启，采用腾讯云默认值
        enabled: true
        # 健康检查的响应超时时间（仅适用于四层监听器），可选值：2~60，默认值：2，单位：秒。响应超时时间要小于检查间隔时间。
        timeout: 2
        # 健康检查探测间隔时间，默认值：5，可选值：5~300，单位：秒。
        intervalTime: 15
        # 健康阈值，默认值：3，表示当连续探测三次健康则表示该转发正常，可选值：2~10，单位：次。
        healthNum: 3
        # 不健康阈值，默认值：3，表示当连续探测三次不健康则表示该转发异常，可选值：2~10，单位：次。
        unHealthNum: 3
    # HTTP端口的映射规则
  - port: 2222
    protocol: HTTP
    # 只对HTTP或者HTTPS端口生效，7层协议端口映射规则数组（port+domain+path确定唯一的规则）
    layer7Routes:
      # 7层规则的域名
    - domain: www.qq.com
      # 7层规则的url
      path: /test1
      # 7层规则的转发详细属性
      # 详情参考链接 https://cloud.tencent.com/document/product/214/30691
      listenerAttribute:
        # 会话保持时间，单位：秒。可选值：30~3600，默认 0，表示不开启。
        sessionTime: 100
        # 负载均衡策略：可选值：WRR, LEAST_CONN, IP_HASH
        lbPolicy: LEAST_CONN
        # 健康检查策略, 如果不设置该选项，默认开启，采用腾讯云默认值
        healthCheck:
          # 是否开启健康检查
          enabled: true
          # 健康检查的响应超时时间（仅适用于四层监听器），可选值：2~60，默认值：2，单位：秒。响应超时时间要小于检查间隔时间。
          timeout: 2
          intervalTime: 15
          healthNum: 3
          unHealthNum: 3
          # 健康检查状态码（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）。可选值：1~31，默认 31。1 表示探测后返回值 1xx 代表健康，2 表示返回 2xx 代表健康，4 表示返回 3xx 代表健康，8 表示返回 4xx 代表健康，16 表示返回 5xx 代表健康。若希望多种返回码都可代表健康，则将相应的值相加。注意：TCP监听器的HTTP健康检查方式，只支持指定一种健康检查状态码。
          httpCode: 31
          # 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）。
          httpCheckPath: /
          # 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET。
          httpCheckMethod: HEAD
      # HTTP或者HTTPS规则映射的后端数组
      services:
      - serviceName: test2
        serviceNamespace: default
        servicePort: 2222
        # 属于该service的监听器后端的权重（只有负载均衡为WRR时才生效）
        weight:
          value: 10
```

## 如何从bcs-clb-controller升级至bcs-ingress-controller

* 1. 删除clb-controller
  * 确保期间与clb-controller控制的workload相关的pod不发生漂移
* 2. 将旧版的clbingress的内容转换成ingress.networkextension.bkbcs.tencent.com的内容
  * 应用新版的ingress
* 3. 启动bcs-ingress-controller

说明：bcs-ingress-controller在同步监听器数据的时候，不会删除已有监听器，而是在已有监听器的基础上做修改。
所以不需要提前进行listener数据的转换。
