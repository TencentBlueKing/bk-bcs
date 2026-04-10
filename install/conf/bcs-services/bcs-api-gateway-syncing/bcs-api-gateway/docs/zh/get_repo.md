### 描述

获取仓库详情

### 路径参数
| 参数名称     | 参数类型 | 必选 | 描述     |
| ------------ | -------- | ---- | -------- |
| projectCode  | string   | 是   | 项目英文名 |
| name         | string   | 是   | 仓库名称   |

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/repos/projectrepo
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "result": true,
    "data": {
        "projectCode": "projecttest",
        "name": "projectrepo",
        "type": "HELM",
        "repoURL": "http://helm.dev.bkrepo.example.com/projecttest/projectrepo/",
        "username": "user",
        "password": "password",
        "createBy": "user",
        "updateBy": "user",
        "createTime": "2022-08-26 14:51:03 +0800 CST",
        "updateTime": "2022-08-26 14:53:25 +0800 CST",
        "displayName": "项目仓库",
        "public": false
    },
    "requestID": "1b38642f-32e1-4842-8f06-d1c7e076f05e"
}
```
