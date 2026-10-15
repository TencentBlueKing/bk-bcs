### 描述

获取仓库列表

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ----------- | ----------- | ------ | ---------------- |
| projectCode         | string       | 是     | 项目英文名     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/repos
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "result": true,
    "data": [
        {
            "projectCode": "projecttest",
            "name": "projecttest",
            "type": "HELM",
            "repoURL": "http://helm.dev.bkrepo.example.com/projecttest/projecttest/",
            "username": "user",
            "password": "password",
            "createBy": "user",
            "updateBy": "user",
            "createTime": "2022-08-26 14:51:03 +0800 CST",
            "updateTime": "2022-08-26 14:53:25 +0800 CST",
            "displayName": "项目仓库",
            "public": false
        },
        {
            "projectCode": "projecttest",
            "name": "public-repo",
            "type": "HELM",
            "repoURL": "https://dev.bkrepo.xample.com/helm/bcs-public-project/helm-public-repo/",
            "username": "",
            "password": "",
            "createBy": "admin",
            "updateBy": "admin",
            "createTime": "2022-11-11 11:57:04 +0800 CST",
            "updateTime": "2022-11-11 11:57:04 +0800 CST",
            "displayName": "公共仓库",
            "public": true
        }
    ],
    "requestID": "1b38642f-32e1-4842-8f06-d1c7e076f05e"
}
```