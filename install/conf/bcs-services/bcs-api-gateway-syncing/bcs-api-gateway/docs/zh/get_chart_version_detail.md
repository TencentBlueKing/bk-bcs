### 描述

获取 chart 某个版本的详情，包含文件信息

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectCode         | string       | 是     | 项目英文名     |
| repoName         | string       | 是     | 仓库名，可以从 list_repos 获取     |
| name         | string       | 是     | chart 名称     |
| version         | string       | 是     | chart 版本     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"access_token": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/helmmanager/v1/projects/projecttest/repos/projecttest/charts/nginx/versions/0.0.1
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "result": true,
    "data": {
        "name": "nginx",
        "version": "0.0.1",
        "readme": "nginx/README.md",
        "valuesFile": [
            "nginx/values.yaml"
        ],
        "contents": {
            "nginx/Chart.yaml": {
                "name": "Chart.yaml",
                "path": "nginx/Chart.yaml",
                "content": "apiVersion: v2\nappVersion: 0.0.11\ndescription: A Helm chart\nname: nginx\ntype: application\nversion: 0.0.1\n"
            },
            "nginx/README.md": {
                "name": "README.md",
                "path": "nginx/README.md",
                "content": ""
            },
            "nginx/values.yaml": {
                "name": "values.yaml",
                "path": "nginx/values.yaml",
                "content": "replicas: 1"
            }
        }
    },
    "requestID": "d00bd380-2427-4cec-8404-fe23488475e2"
}
```

### 响应参数说明
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|   name           |     string       |          chart 名称                      |
|   version           |     string       |          chart 版本                      |
|   readme           |     string       |          chart README 文件路径                      |
|   valuesFile           |     string array       |          chart values 文件路径                      |
|   contents           |     object array       |          chart 所有文件内容                      |