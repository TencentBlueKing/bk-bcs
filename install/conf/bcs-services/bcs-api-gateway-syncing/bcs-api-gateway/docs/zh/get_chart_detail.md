### 描述

获取 chart 详情

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectCode         | string       | 是     | 项目英文名     |
| repoName         | string       | 是     | 仓库名，可以从 list_repos 获取     |
| name         | string       | 是     | chart 名称     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/repos/projecttest/charts/nginx
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "result": true,
    "data": {
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
        "createTime": "2022-11-16T18:00:06.996",
        "updateTime": "2022-11-18T16:05:33.78",
        "projectCode": "projecttest",
        "icon": ""
    },
    "requestID": "653c7755-e40d-42ae-ba25-93aef823cf65"
}
```