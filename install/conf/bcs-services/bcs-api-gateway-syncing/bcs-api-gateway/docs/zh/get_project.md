### 描述

获取 BCS 项目详情

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectIDOrCode         | string       | 是     | 项目ID或项目英文名     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/bcsproject/v1/projects/testproject
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "createTime": "2006-01-02T15:04:05Z",
        "updateTime": "2006-01-02T15:04:05Z",
        "creator": "testuser",
        "updater": "testuser",
        "managers": "testuser",
        "projectID": "1xxx3xxx5xxx4xxx8xxx1xxx2xxxexxx",
        "name": "testproject",
        "projectCode": "testproject",
        "useBKRes": false,
        "description": "test",
        "isOffline": false,
        "kind": "k8s",
        "businessID": "100000",
        "businessName": "xxxxx"
    },
    "requestID": "894f7249f43c4e5xxxx9207058045e8e",
    "webAnnotations": {
        "perms": null
    }
}
```