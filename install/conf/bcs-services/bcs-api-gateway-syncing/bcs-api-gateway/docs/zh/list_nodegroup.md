### 描述

节点池列表。根据查询条件过滤节点池（NodeGroup）列表，条件为空时忽略该条件；所有条件都为空时查询全部节点池

### 调用示例

```bash
curl -X GET \
-H 'X-Bkapi-Authorization: {"bk_app_code": "bk_apigw_test", "bk_app_secret": "***"}' \
'http://bcs-api.bkdomain/bcsapi/clustermanager/v1/nodegroup?clusterID=BCS-K8S-00000&projectID=xxx'
```

### 请求参数

| 名称        | 位置    | 类型     | 必选 | 说明              |
|-----------|-------|--------|----|-----------------|
| name      | query | string | 否  | 节点池名称           |
| clusterID | query | string | 否  | 节点池所属 BCS 集群 ID  |
| region    | query | string | 否  | 节点池所属地域         |
| projectID | query | string | 否  | 节点池所属项目 ID      |

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": true,
  "data": [
    {
      "nodeGroupID": "BCS-ng-xxxxxxxx",
      "name": "test-nodegroup",
      "clusterID": "BCS-K8S-00000",
      "region": "ap-guangzhou",
      "enableAutoscale": true,
      "autoScaling": {
        "autoScalingID": "asg-xxxxxxxx",
        "minSize": 0,
        "maxSize": 10,
        "desiredSize": 1,
        "subnetIDs": [
          "subnet-xxxxxxxx"
        ]
      },
      "nodeOS": "TencentOS Server 3.1",
      "creator": "admin",
      "updater": "admin",
      "createTime": "2026-01-01T00:00:00+08:00",
      "updateTime": "2026-01-01T00:00:00+08:00",
      "projectID": "xxx",
      "provider": "tencentCloud",
      "status": "RUNNING",
      "cloudNodeGroupID": "np-xxxxxxxx",
      "nodeGroupType": "normal",
      "nodeTemplate": {},
      "area": {
        "bkCloudID": 0,
        "bkCloudName": "Default Area"
      }
    }
  ]
}
```

### 响应参数说明

| 名称                   | 类型      | 说明                |
|----------------------|---------|-------------------|
| code                 | integer | 返回错误码，0 表示成功      |
| message              | string  | 返回错误信息            |
| result               | boolean | 返回结果              |
| data                 | array   | 节点池列表             |
| data.nodeGroupID     | string  | 节点池 ID            |
| data.name            | string  | 节点池名称             |
| data.clusterID       | string  | 所属 BCS 集群 ID      |
| data.region          | string  | 所属地域              |
| data.enableAutoscale | boolean | 是否开启自动扩缩容         |
| data.autoScaling     | object  | 扩缩容配置             |
| data.nodeOS          | string  | 节点操作系统            |
| data.projectID       | string  | 所属项目 ID           |
| data.provider        | string  | 云提供商              |
| data.status          | string  | 节点池状态             |
| data.cloudNodeGroupID | string | 云上节点池 ID          |
| data.nodeGroupType   | string  | 节点池类型             |
| data.nodeTemplate    | object  | 节点模板配置            |
| data.area            | object  | 云区域信息             |
