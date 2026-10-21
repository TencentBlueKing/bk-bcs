### 描述

获取节点详情接口，通过节点内网 IP 查询节点信息

### 调用示例

```bash
curl -X GET \
-H 'X-Bkapi-Authorization: {"bk_app_code": "bk_apigw_test", "bk_app_secret": "***"}' \
'http://bcs-api.bkdomain/bcsapi/clustermanager/v1/node/127.0.0.1'
```

### 请求参数

| 名称      | 位置    | 类型      | 必选 | 说明                    |
|---------|-------|---------|----|-----------------------|
| nodeIP  | path  | string  | 是  | 节点内网 IP               |
| showPwd | query | boolean | 否  | 节点密码是否展示，默认不展示        |

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": true,
  "data": [
    {
      "nodeID": "ins-xxxxxxxx",
      "innerIP": "127.0.0.1",
      "instanceType": "S5.MEDIUM4",
      "CPU": 2,
      "mem": 4,
      "GPU": 0,
      "status": "RUNNING",
      "zoneID": "ap-guangzhou-3",
      "nodeGroupID": "",
      "clusterID": "BCS-K8S-00000",
      "VPC": "vpc-xxxxxxxx",
      "region": "ap-guangzhou",
      "passwd": "",
      "zone": 100003,
      "deviceID": "",
      "nodeTemplateID": "",
      "nodeType": "CVM",
      "nodeName": "",
      "innerIPv6": "",
      "zoneName": "广州三区",
      "taskID": "",
      "failedReason": "",
      "chargeType": "POSTPAID_BY_HOUR",
      "dataDiskNum": 1,
      "isGpuNode": false
    }
  ]
}
```

### 响应参数说明

| 名称                | 类型      | 说明                                                     |
|-------------------|---------|--------------------------------------------------------|
| code              | integer | 返回错误码，0 表示成功                                           |
| message           | string  | 返回错误信息                                                 |
| result            | boolean | 返回结果                                                   |
| data              | array   | 节点信息列表                                                 |
| data.nodeID       | string  | 节点云 instanceID                                         |
| data.innerIP      | string  | 节点内网主 IP                                               |
| data.instanceType | string  | 机型信息，可能为空                                              |
| data.CPU          | integer | 节点 CPU 核数                                              |
| data.mem          | integer | 节点内存                                                   |
| data.GPU          | integer | 节点 GPU 数量，可能为 0                                        |
| data.status       | string  | 节点状态：RUNNING、DELETING、CREATING、FAILED、INITIALIZATION    |
| data.zoneID       | string  | 节点可用区，可能为空                                             |
| data.nodeGroupID  | string  | 节点所属节点池 ID，仅节点池节点有值                                    |
| data.clusterID    | string  | 节点所属集群 ID                                              |
| data.VPC          | string  | 节点所属 VPC ID                                            |
| data.region       | string  | 节点所属地域                                                 |
| data.passwd       | string  | 实例初始化密码，仅 showPwd=true 时返回                             |
| data.nodeType     | string  | 节点类型，主要包含 CVM/IDC，默认 CVM                               |
| data.nodeName     | string  | 节点名称                                                   |
| data.innerIPv6    | string  | 节点 IPv6 地址                                             |
| data.zoneName     | string  | 节点所在可用区名称                                              |
| data.chargeType   | string  | 计费模式：POSTPAID_BY_HOUR 按小时后付费，PREPAID 包年包月             |
| data.isGpuNode    | boolean | 是否 GPU 节点                                              |
