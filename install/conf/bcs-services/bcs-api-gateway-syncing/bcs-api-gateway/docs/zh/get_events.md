### 描述

查询集群事件，按事件时间倒序返回。该接口需要用户认证，且用户需具备项目查看权限；独立集群需属于该项目。

### 路径参数
| 参数名称     | 参数类型 | 必选 | 描述 |
| ------------ | -------- | ---- | ---- |
| project_code | string   | 是   | 项目英文名 |
| cluster_id   | string   | 是   | 集群 ID |

### 查询参数
| 参数名称            | 参数类型 | 必选 | 描述 |
| ------------------- | -------- | ---- | ---- |
| env                 | string   | 否   | 事件环境，K8S 集群为 k8s |
| kind                | string   | 否   | 资源类型过滤，如 Pod、Deployment，多个以英文逗号分隔 |
| level               | string   | 否   | 事件级别过滤，如 Normal、Warning，多个以英文逗号分隔 |
| component           | string   | 否   | 事件来源组件过滤，多个以英文逗号分隔 |
| type                | string   | 否   | 事件类型过滤，多个以英文逗号分隔 |
| extraInfo.namespace | string   | 否   | 命名空间过滤，多个以英文逗号分隔；共享集群需指定 |
| extraInfo.name      | string   | 否   | 资源名称过滤，多个以英文逗号分隔 |
| extraInfo.kind      | string   | 否   | 资源类型过滤（extraInfo 中的 kind），多个以英文逗号分隔 |
| timeBegin           | int      | 否   | 开始时间，Unix 秒级时间戳 |
| timeEnd             | int      | 否   | 结束时间，Unix 秒级时间戳 |
| offset              | int      | 否   | 分页偏移量 |
| length              | int      | 否   | 返回记录数量上限 |
| field               | string   | 否   | 返回字段，多个以英文逗号分隔，不传返回全部字段 |

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure 'https://bcs-api-gateway.apigw.com/prod/clusterresources/api/v1/projects/testproject/clusters/BCS-K8S-00000/events?env=k8s&kind=Pod&extraInfo.namespace=default&offset=0&length=10'
```

### 响应示例
```json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": [
        {
            "env": "k8s",
            "kind": "Pod",
            "level": "Warning",
            "component": "kubelet",
            "type": "BackOff",
            "describe": "Back-off restarting failed container",
            "clusterId": "BCS-K8S-00000",
            "eventTime": "2026-01-01T00:00:00Z",
            "extraInfo": {
                "name": "nginx-7d9c6b5f4-abcde",
                "namespace": "default",
                "kind": "Pod"
            }
        }
    ],
    "total": 1
}
```

### 响应参数说明
| 参数名称  | 参数类型 | 描述 |
| --------- | -------- | ---- |
| data      | array    | 事件列表，按事件时间倒序；每条事件另含 `id` 及原始 K8S 事件对象 `data`，可通过 `field` 参数裁剪 |
| total     | int      | 符合条件的事件总数 |
