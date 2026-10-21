### 描述

删除共享集群命名空间 ITSM 回调，仅供 ITSM 审批流程回调使用

### 请求参数
| 参数名称              | 参数类型 | 必选 | 描述           |
| --------------------- | -------- | ---- | -------------- |
| title                 | string   | 否   | ITSM 单据标题 |
| id                    | string   | 否   | ITSM 单号 ID   |
| status                | string   | 否   | ITSM 单据状态 |
| approveResult         | bool     | 否   | 单据审批结果   |
| ticketUrl             | string   | 否   | ITSM 单据链接 |
| formData              | object   | 否   | ITSM 单据数据 |
| formData.title        | string   | 否   | ITSM 单据标题 |
| formData.projectCode  | string   | 否   | 项目英文名     |
| formData.clusterId    | string   | 否   | 集群 ID        |
| formData.namespace    | string   | 否   | 命名空间       |

### 调用示例
```sh
curl -X POST -H 'x-bkapi-authorization: {"bk_app_code": "xxx", "bk_app_secret": "***"}' -H 'Content-Type: application/json' -d '{"title": "xxx", "id": "xxx", "status": "FINISHED", "approveResult": true, "ticketUrl": "xxx", "formData": {"title": "xxx", "projectCode": "testproject", "clusterId": "BCS-K8S-12345", "namespace": "test-ns"}}' --insecure https://bcs-api-gateway.apigw.com/prod/bcsproject/v1/projects/sharedclusters/namespaces/callback/delete
```

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": "",
    "requestID": "894f7249f43c4e5xxxx9207058045e8e",
    "result": true
}
```
