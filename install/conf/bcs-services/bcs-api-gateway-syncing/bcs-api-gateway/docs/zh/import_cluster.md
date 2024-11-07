### 描述

导入用户集群(支持多云集群导入功能: 集群ID/kubeConfig)

### 调用示例

```sh
curl -X 'POST' \
-H 'Cookie: bk_token=xxx' \
-H 'User-Agent: xxx' \
-d '{}' \
'http://bcs-api.bkdomain/bcsapi/v4/clustermanager/v1/cluster/import'
```

### 请求体参数

| 名称              | 类型      | 必选 | 说明                                                                       |
|-----------------|---------|----|--------------------------------------------------------------------------|
| clusterID       | string  | 否  | 集群ID，例如BCS-K8S-000000(手动录入信息时直接指定clusterID,自动导入会默认生成集群ID)                |
| clusterName     | string  | 是  | 集群名称                                                                     |
| description     | string  | 否  | 集群简要描述                                                                   |
| provider        | string  | 是  | 云模版ID, 集群所属云进行流程管理                                                       |
| region          | string  | 否  | 集群所在地域                                                                   |
| projectID       | string  | 是  | 集群所属项目                                                                   |
| businessID      | string  | 是  | CMDB业务ID                                                                 |
| environment     | string  | 是  | 集群环境, 例如[prod, debug, stag]                                              |
| engineType      | string  | 是  | 引擎类型，默认k8s                                                               |
| isExclusive     | boolean | 是  | 是否为业务独占集群,默认为true                                                        |
| clusterType     | string  | 是  | 集群类型, 例如[federation, single], federation表示为联邦集群，single表示独立集群，默认为single   |
| labels          | object  | 否  | 集群的labels，用于携带额外的信息，最大不得超过20个                                            |
| creator         | string  | 是  | 创建人                                                                      |
| cloudMode       | string  | 否  | 云provider导入集群模式                                                          |
| manageType      | string  | 否  | 集群管理类型，公有云时生效，MANAGED_CLUSTER(云上托管集群)，默认是 INDEPENDENT_CLUSTER(独立集群，自行维护) |
| networkType     | string  | 否  | 集群网络类型(underlay/overlay),默认是overlay                                      |
| extraInfo       | object  | 否  | 存储集群扩展信息, 例如esb_url/webhook_image/priviledge_image等扩展信息                  |
| extraClusterID  | string  | 否  | 导入集群的额外集群ID标识信息,默认时空值                                                    |
| clusterCategory | string  | 否  | 集群类别，主要用于区分该集群是否是自建、导入(builder/importer), 默认是自建                          |
| is_shared       | boolean | 否  | 是否为共享集群,默认false                                                          |
| version         | string  | 否  | 导入集群版本信息                                                                 |
| accountID       | string  | 否  | 导入集群关联的cloud凭证信息, 可以为空; 当为空时, 可以通过cloud获取cloud凭证信息                       |
| area            | object  | 否  | 云区域                                                                      |

请求体示例：

```json
{
  "clusterID": "string",
  "clusterName": "string",
  "description": "string",
  "provider": "string",
  "region": "string",
  "projectID": "string",
  "businessID": "string",
  "environment": "string",
  "engineType": "string",
  "isExclusive": true,
  "clusterType": "string",
  "labels": {
    "property1": "string",
    "property2": "string"
  },
  "creator": "string",
  "cloudMode": {
    "cloudID": "string",
    "kubeConfig": "string",
    "inter": true,
    "resourceGroup": "string",
    "nodeIps": [
      "string"
    ]
  },
  "manageType": "string",
  "networkType": "string",
  "extraInfo": {
    "property1": "string",
    "property2": "string"
  },
  "extraClusterID": "string",
  "clusterCategory": "string",
  "is_shared": true,
  "version": "string",
  "accountID": "string",
  "area": {
    "bkCloudID": 0,
    "bkCloudName": "string"
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "string",
  "result": true,
  "data": {}
}
```