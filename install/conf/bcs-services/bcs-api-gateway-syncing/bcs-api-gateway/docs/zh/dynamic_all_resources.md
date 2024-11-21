### 描述

根据 clusterId，查询集群资源，并返回 list

### 请求参数

| 名称               | 位置    | 类型             | 必选 | 说明                           |
|------------------|-------|----------------|----|------------------------------|
| clusterId        | path  | string         | 是  | BCS内部使用集群ID，格式为BCS-K8S-XXXXX |
| resourceType     | path  | string         | 是  | 资源类型                         |
| extra            | query | string         | 否  | extra. 扩展字段                  |
| labelSelector    | query | string         | 否  | labelSelector. 选择器           |
| updateTimeBefore | query | string(int64)  | 否  | updateTimeBefore. 更新时间之前     |
| offset           | query | string(uint64) | 否  | offset. 查询偏移量                |
| limit            | query | string(uint64) | 否  | limit. 查询限制数量                |
| fields           | query | array[string]  | 否  | fields. 额外字段                 |

### 调用示例

```sh
curl -X GET \
--header 'Authorization: Bearer xxx' \
'http://bcs-api.bkdomain/bcsapi/v4/storage/k8s/dynamic/all_resources/clusters/{clusterId}/Deployment?offset=0&limit=10'
```

### 响应示例

```json
{
  "code": 0,
  "message": "string",
  "result": true,
  "data": [
    {}
  ]
}
```