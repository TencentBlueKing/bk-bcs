### 描述

获取模板文件详情

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| project_code         | string       | 是     | 项目英文名     |
| id         | string       | 是     | 模板文件 id     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/clusterresources/v1/projects/{project_code}/template/metadatas/{id}
```

### 响应示例
```json
{
  "code": 0,
  "message": "OK",
  "requestID": "34f78db7-cf02-4b47-a28d-0b199d252c86",
  "data": {
    "createAt": 1717500971,
    "creator": "xxx",
    "description": "",
    "draftContent": "",
    "draftEditFormat": "",
    "draftVersion": "",
    "id": "xxxx",
    "isDraft": false,
    "name": "deployment-1.yaml",
    "projectCode": "testprojectli",
    "resourceType": [
      "Deployment"
    ],
    "tags": [],
    "templateSpace": "test_multi",
    "templateSpaceID": "xxxx",
    "updateAt": 1717501308,
    "updator": "xxx",
    "version": "v1",
    "versionID": "xxx",
    "versionMode": 0
  },
  "webAnnotations": null
}
```