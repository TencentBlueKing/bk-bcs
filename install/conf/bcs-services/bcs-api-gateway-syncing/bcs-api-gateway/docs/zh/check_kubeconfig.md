### 描述

kubeConfig 连接集群可用性检测，在纳管已经创建的集群之前，可以先通过该接口，测试集群的可用性，正常返回再进行导入纳管。

### 参数

| 名称         | 位置   | 类型     | 必选 | 说明             |
|------------|------|--------|----|----------------|
| kubeConfig | Body | string | 是  | kubeConfig 字符串 |

### 调用示例

```sh
curl -X 'POST' \
-H 'Cookie: bk_token=xxx' \
-H 'User-Agent: xxx' \
-d '{ "kubeConfig": "apiVersion: v1\nclusters:\n- cluster:\n..." }' \
'http://bcs-api.bkdomain/bcsapi/v4/clustermanager/v1/cloud/kubeConfig'
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "result": true
}
```