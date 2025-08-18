### 描述

获取模板文件版本列表

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| project_code         | string       | 是     | 项目英文名     |
| template_id         | string       | 是     | 模板文件 id     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/uat/clusterresources/v1/projects/{project_code}//{template_id}/versions
```

### 响应示例
```json
{
  "code": 0,
  "message": "OK",
  "requestID": "c726c81c-b369-48b0-9647-598cc4f4c490",
  "data": [
    {
      "content": "---\napiVersion: apps/v1 # apiVersion is related to the cluster version(e.g extensions/v1beta1 in 1.8)\nkind: Deployment\nmetadata:\n  name: nginx-deployment\n  labels:\n    app: nginx\nspec:\n  replicas: 2\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:1.7.9 # replace with image url of BCS dept\n        ports:\n        - containerPort: 80\n---\napiVersion: apps/v1 # apiVersion is related to the cluster version(e.g extensions/v1beta1 in 1.8)\nkind: Deployment\nmetadata:\n  name: nginx-deployment2\n  labels:\n    app: nginx\nspec:\n  replicas: 2\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:1.7.9 # replace with image url of BCS dept\n        ports:\n        - containerPort: 80\n---\napiVersion: apps/v1 # apiVersion is related to the cluster version(e.g extensions/v1beta1 in 1.8)\nkind: Deployment\nmetadata:\n  name: nginx-deployment3\n  labels:\n    app: nginx\nspec:\n  replicas: 2\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:1.7.9 # replace with image url of BCS dept\n        ports:\n        - containerPort: 80",
      "createAt": 1717501308,
      "creator": "xxxx",
      "description": "",
      "draft": false,
      "editFormat": "yaml",
      "id": "xxxx",
      "latest": true,
      "projectCode": "testprojectli",
      "templateName": "deployment-1.yaml",
      "templateSpace": "test_multi",
      "version": "v1"
    }
  ],
  "webAnnotations": null
}
```