### 描述
该接口提供版本：v1.0.0+
 
客户端去制品库下载应用配置时，制品库(bk_repo)回调BSCP进行回调鉴权API， 仅限制品库(bk_repo)使用。

### 输入参数
| 参数名称      | 参数类型   | 必选  | 描述        |
|-----------|--------|-----|-----------|
| userId    | string | 是   | 需要鉴权的用户ID |
| type      | string | 是   | 需要鉴权的资源类型 |
| action    | string | 是   | 需要鉴权的资源操作 |
| projectId | string | 是   | 需要鉴权的项目ID |
| repoName  | string | 是   | 需要鉴权的仓库名  |
| nodes     | array  | 是   | 需要鉴权的仓库节点 |

#### 字段说明
##### nodes
nodes 是需要鉴权的仓库节点数组，其中节点数据说明如下。

| 参数名称     | 参数类型                | 必选  | 描述                                                                 |
|----------|---------------------|-----|--------------------------------------------------------------------|
| fullPath | string              | 是   | 配置文件在制品库存储的节点路径                                                    |
| metadata | map<string, string> | 否   | 该metadata数据由bscp在上传文件数据时写入到制品库，并由制品库带回给bscp用于鉴权，目前只有biz_id与app_id。 |

### 调用示例
```json
{
  "userId": "xxx",
  "type": "NODE",
  "action": "READ",
  "projectId": "bscp",
  "repoName": "bscp-v1-2",
  "nodes": [
    {
      "fullPath": "/file/xxxxxx",
      "metadata": {
        "app_id": "[5]",
        "biz_id": "2"
      }
    },
    {
      "fullPath": "/file/yyyyyy",
      "metadata": {
        "app_id": "[5]",
        "biz_id": "2"
      }
    }
  ]
}
```

### 响应示例
```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明
| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
