### 描述

获取模板文件列表

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| project_code         | string       | 是     | 项目英文名     |
| template_space_id         | string       | 是     | 模板文件文件夹 id     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/clusterresources/v1/projects/{project_code}/template/{template_space_id}/metadatas
```

### 响应示例
```json
{
  "code": 0,
  "message": "OK",
  "requestID": "30c78042-f7b2-4dc8-b66e-20049476e928",
  "data": [
    {
      "createAt": 1717500971,
      "creator": "xxx",
      "description": "",
      "draftContent": "",
      "draftEditFormat": "",
      "draftVersion": "",
      "id": "xxx",
      "isDraft": false,
      "name": "deployment-1.yaml",
      "projectCode": "testprojectli",
      "resourceType": [
        "Deployment"
      ],
      "tags": [],
      "templateSpace": "test_multi",
      "templateSpaceID": "",
      "updateAt": 1717501308,
      "updator": "xxx",
      "version": "v1",
      "versionID": "xxxx",
      "versionMode": 0
    },
    {
      "createAt": 1728526272,
      "creator": "xxxx",
      "description": "",
      "draftContent": "",
      "draftEditFormat": "yaml",
      "draftVersion": "",
      "id": "xxx",
      "isDraft": false,
      "name": "service-ip-test",
      "projectCode": "testprojectli",
      "resourceType": [
        "Service"
      ],
      "tags": [],
      "templateSpace": "test_multi",
      "templateSpaceID": "",
      "updateAt": 1728528940,
      "updator": "xxx",
      "version": "4.0.0",
      "versionID": "xxx",
      "versionMode": 0
    },
    {
      "createAt": 1728554385,
      "creator": "xxx",
      "description": "",
      "draftContent": "",
      "draftEditFormat": "",
      "draftVersion": "",
      "id": "xxx",
      "isDraft": false,
      "name": "svc-test-cl",
      "projectCode": "testprojectli",
      "resourceType": [
        "Service"
      ],
      "tags": [],
      "templateSpace": "test_multi",
      "templateSpaceID": "",
      "updateAt": 1728554414,
      "updator": "xxx",
      "version": "2.0.0",
      "versionID": "xxx",
      "versionMode": 0
    },
    {
      "createAt": 1717660977,
      "creator": "xxx",
      "description": "",
      "draftContent": "",
      "draftEditFormat": "",
      "draftVersion": "",
      "id": "xxx",
      "isDraft": false,
      "name": "template-1717660957",
      "projectCode": "testprojectli",
      "resourceType": [
        "ConfigMap"
      ],
      "tags": [],
      "templateSpace": "test_multi",
      "templateSpaceID": "",
      "updateAt": 1717665403,
      "updator": "xxx",
      "version": "3.0.0",
      "versionID": "xxx",
      "versionMode": 0
    },
    {
      "createAt": 1717664428,
      "creator": "xxx",
      "description": "",
      "draftContent": "",
      "draftEditFormat": "",
      "draftVersion": "",
      "id": "xxx",
      "isDraft": false,
      "name": "template-1717664416",
      "projectCode": "testprojectli",
      "resourceType": [
        "Deployment"
      ],
      "tags": [],
      "templateSpace": "test_multi",
      "templateSpaceID": "",
      "updateAt": 1717664428,
      "updator": "xxx",
      "version": "1.0.0",
      "versionID": "xxx",
      "versionMode": 0
    }
  ],
  "webAnnotations": null
}
```