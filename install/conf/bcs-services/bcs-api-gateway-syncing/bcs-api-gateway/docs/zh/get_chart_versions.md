### 描述

获取 chart 版本列表

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectCode         | string       | 是     | 项目英文名     |
| repoName         | string       | 是     | 仓库名，可以从 list_repos 获取     |
| name         | string       | 是     | chart 名称     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/repos/projecttest/charts/nginx/versions
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "result": true,
    "data": {
        "page": 1,
        "size": 1000,
        "total": 2,
        "data": [
            {
                "name": "nginx",
                "version": "0.0.2",
                "appVersion": "0.0.2",
                "description": "A Helm chart",
                "createBy": "admin",
                "updateBy": "admin",
                "createTime": "2022-11-18T16:05:33.758",
                "updateTime": "2022-11-18T16:05:33.758"
            },
            {
                "name": "nginx",
                "version": "0.0.1",
                "appVersion": "0.0.1",
                "description": "A Helm chart",
                "createBy": "admin",
                "updateBy": "admin",
                "createTime": "2022-11-17T12:06:02.107",
                "updateTime": "2022-11-17T12:06:02.107"
            }
        ]
    },
    "requestID": "0eb1451a-19d0-4002-bd26-a634af5f59e2"
}
```