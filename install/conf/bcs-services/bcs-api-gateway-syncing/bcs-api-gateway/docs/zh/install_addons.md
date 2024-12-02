### 描述

安装集群组件，该接口是异步接口，需要通过 addons 列表接口（`list_addons`）或查看 addons 详情接口（`get_addons_details`）获取
addons 的状态

### 路径参数

| 参数名称        | 参数类型   | 必选 | 描述     |
|-------------|--------|----|--------|
| projectCode | string | 是  | 项目英文名  |
| clusterID   | string | 是  | 目标集群ID |

### Body

body 的字段：

| 参数名称    | 参数类型   | 必选 | 描述     |
|---------|--------|----|--------|
| name    | string | 是  | 组件名称   |
| version | string | 是  | 组件版本   |
| value   | string | 否  | values |

示例：

```json
{
  "name": "bk-log-collector",
  "version": "0.3.5-alpha-x86.12",
  "values": ""
}
```

### 调用示例

```sh
curl -X POST \
-d '{"name": "bk-log-collector", "version": "0.3.5-alpha-x86.12", "values": ""}' \
-H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' \
--insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/clusters/clustertest/addons
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": true,
  "requestID": "9d68e4ebf74ded7e02c69c9441100339",
  "web_annotations": null
}
```
