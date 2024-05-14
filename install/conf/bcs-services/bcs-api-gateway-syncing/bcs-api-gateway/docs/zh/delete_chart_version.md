### 描述

删除 chart 版本

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectCode         | string       | 是     | 项目英文名     |
| repoName         | string       | 是     | 仓库名，可以从 list_repos 获取     |
| name         | string       | 是     | chart 名称     |
| version         | string       | 是     | chart 版本     |


### 调用示例
```sh
curl -X DELETE -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/repos/projecttest/charts/nginx/versions/0.0.1
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "result": true,
    "requestID": "a2a46e9a-9c4d-4547-8b40-82373ce0b9ff"
}
```