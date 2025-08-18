### 描述

模板文件部署预览

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| project_code         | string       | 是     | 项目英文名     |

### Body
```json
{
  "templateVersions": ["xxxx"],
  "variables": {
    "port": "80"
  },
  "clusterID": "BCS-K8S-12345",
  "namespace": "hito-test"
}
```


### 调用示例
```sh
curl -X POST -d 'your_body.json' -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/clusterresources/v1/projects/{project_code}/template/preview
```

### 响应示例
```json
{
  "code": 0,
  "message": "OK",
  "requestID": "39012af3-4b12-476c-baa2-669883ad297e",
  "data": {
    "error": "unknown",
    "items": [
      {
        "content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  annotations:\n    io.tencent.paas.creator: xxx\n    io.tencent.paas.source_type: template\n    io.tencent.paas.template_name: test_multi/deployment-1.yaml\n    io.tencent.paas.template_version: v1\n    io.tencent.paas.updator: xxx\n  labels:\n    app: nginx\n  name: nginx-deployment\n  namespace: hito-test\nspec:\n  replicas: 2\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - image: nginx:1.7.9\n        name: nginx\n        ports:\n        - containerPort: 80\n",
        "kind": "Deployment",
        "name": "nginx-deployment",
        "previousContent": ""
      },
      {
        "content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  annotations:\n    io.tencent.paas.creator: xxx\n    io.tencent.paas.source_type: template\n    io.tencent.paas.template_name: test_multi/deployment-1.yaml\n    io.tencent.paas.template_version: v1\n    io.tencent.paas.updator: xxx\n  labels:\n    app: nginx\n  name: nginx-deployment2\n  namespace: hito-test\nspec:\n  replicas: 2\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - image: nginx:1.7.9\n        name: nginx\n        ports:\n        - containerPort: 80\n",
        "kind": "Deployment",
        "name": "nginx-deployment2",
        "previousContent": ""
      },
      {
        "content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  annotations:\n    io.tencent.paas.creator: xxx\n    io.tencent.paas.source_type: template\n    io.tencent.paas.template_name: test_multi/deployment-1.yaml\n    io.tencent.paas.template_version: v1\n    io.tencent.paas.updator: xxx\n  labels:\n    app: nginx\n  name: nginx-deployment3\n  namespace: hito-test\nspec:\n  replicas: 2\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - image: nginx:1.7.9\n        name: nginx\n        ports:\n        - containerPort: 80\n",
        "kind": "Deployment",
        "name": "nginx-deployment3",
        "previousContent": ""
      }
    ]
  },
  "webAnnotations": null
}
```