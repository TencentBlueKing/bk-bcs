### 描述

获取模板文件文件夹列表

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| project_code         | string       | 是     | 项目英文名     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/uat/clusterresources/v1/projects/{project_code}/template/spaces
```

### 响应示例
```json
{
  "code": 0,
  "message": "OK",
  "requestID": "9dde39deb338492fae6a919ef9423223",
  "data": [
    {
      "description": "",
      "fav": true,
      "id": "66601e6e608f57b998ce2b15",
      "name": "test_multi",
      "projectCode": "testprojectli",
      "tags": []
    },
    {
      "description": "这是一款开源的吃豆游戏的配置模板集",
      "fav": true,
      "id": "66601e71608f57b998ce2b21",
      "name": "示例模板集",
      "projectCode": "testprojectli",
      "tags": []
    },
    {
      "description": "",
      "fav": false,
      "id": "66d6abc5a20aaf825a74b9d4",
      "name": "/_我是克隆来的",
      "projectCode": "testprojectli",
      "tags": []
    },
    {
      "description": "",
      "fav": false,
      "id": "668cfe58e321922fe90388c1",
      "name": "AIDev-Demo",
      "projectCode": "testprojectli",
      "tags": []
    },
    {
      "description": "",
      "fav": false,
      "id": "66c730695f26f75b94c7c39a",
      "name": "AIDev-Demo_1724330089",
      "projectCode": "testprojectli",
      "tags": []
    }
  ],
  "webAnnotations": null
}
```