### 描述

获取模板文件版本详情

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| project_code         | string       | 是     | 项目英文名     |

### Body
```json
{
  "templateSpace": "space1",
  "templateName": "template1",
  "version": "0.0.1"
}
```


### 调用示例
```sh
curl -X POST -d 'your_body.json' -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/clusterresources/v1/projects/{project_code}/template/detail
```

### 响应示例
```json
{
  "code": 0,
  "message": "OK",
  "requestID": "202d9cf7-20d2-4ba5-81c8-7d035ae9bd65",
  "data": {
    "content": "xxx",
    "createAt": 1718873051,
    "creator": "xxx",
    "description": "1",
    "draft": false,
    "editFormat": "yaml",
    "id": "6673ebdb1ccfc76b10630c86",
    "latest": false,
    "projectCode": "testprojectli",
    "templateName": "template-1718613671",
    "templateSpace": "boweiguan",
    "version": "7.0.0"
  },
  "webAnnotations": null
}
```