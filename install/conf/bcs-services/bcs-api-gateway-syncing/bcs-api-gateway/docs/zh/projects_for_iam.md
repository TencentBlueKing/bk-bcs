### 描述

获取 BCS 项目列表

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/bcsproject/v1/projects_for_iam
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "name": "test-project",
            "projectID": "xxxx23c8xxxx499eb3990b6a59a23b9f",
            "projectCode": "project-xxx",
            "businessID": "100",
            "managers": "testuser1,testuser2",
            "bkmSpaceBizID": 123,
            "bkmSpaceName": "test-project"
        }
    ],
    "requestID": "12341234ed2e44012348bf33b054a564"
}
```