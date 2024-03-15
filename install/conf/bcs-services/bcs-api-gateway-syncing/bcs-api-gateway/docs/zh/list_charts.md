### 描述

获取 chart 列表

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectCode         | string       | 是     | 项目英文名     |
| repoName         | string       | 是     | 仓库名，可以从 list_repos 获取     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/repos/projecttest/charts
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
        "total": 4,
        "data": [
            {
                "projectID": "projecttest",
                "repository": "projecttest",
                "type": "HELM",
                "key": "helm://nginx",
                "name": "nginx",
                "latestVersion": "0.0.1",
                "latestAppVersion": "0.0.1",
                "latestDescription": "A Helm chart",
                "createBy": "admin",
                "updateBy": "admin",
                "createTime": "2022-11-16 18:00:06",
                "updateTime": "2022-11-18 16:05:33",
                "projectCode": "projecttest",
                "icon": ""
            }
        ]
    },
    "requestID": "1bce1136-a00e-47ca-b5e4-6b077a8ee7d4"
}
```