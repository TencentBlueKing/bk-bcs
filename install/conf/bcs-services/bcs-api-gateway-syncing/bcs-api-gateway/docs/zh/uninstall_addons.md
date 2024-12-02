### 描述

卸载集群组件，该接口是异步接口，需要通过 addons 列表接口（`list_addons`）或查看 addons 详情接口（`get_addons_details`）获取 addons 的状态

### 路径参数

| 参数名称        | 参数类型   | 必选 | 描述      |
|-------------|--------|----|---------|
| projectCode | string | 是  | 项目代码    |
| clusterID   | string | 是  | 所在的集群ID |
| name        | string | 是  | 组件名称    |

### 调用示例

```sh
curl -X DELETE -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/clusters/cluster-test/addons/test-addons
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": true,
  "requestID": "67833fab2f5279fd34394afcc332d199",
  "web_annotations": null
}
```
