### 描述

该接口提供版本：v1.0.0+

查询credential的规则

### 输入参数

| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| biz_id         | uint32       | 是     | 业务ID     |
| credential_id         | uint32       | 是     | credential的ID     |

#### 查询参数介绍

| 参数名称     | 参数类型     | 必选   | 描述                                  |
| ------------ | ------------ | ------ |-------------------------------------|
| biz_id    | uint32       | 是     | 业务ID                                |
| credential_id | uint32 | 是 | credential的ID              |


### 调用示例

```json
```

### 响应示例

```json
{
  "data": {
    "count": 1,
    "details": [
      {
        "id": 7,
        "spec": {
          "credential_scope": "XXXX"
        },
        "attachment": {
          "biz_id": 5,
          "credential_id": 5
        },
        "revision": {
          "creator": "credential_scope_tester",
          "reviser": "credential_scope_tester",
          "create_at": "2023-04-14 10:32:22",
          "update_at": "2023-04-14 10:32:22",
          "expired_at": "2023-04-14 10:32:22"
        }
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|       data       |      object      |            响应数据                  |

#### data

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      count        |      uint32      |            当前规则能匹配到的总记录条数                  |
|      detail        |      array      |             查询返回的数据                  |

#### data.details[n]

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      id        |      uint32      |            应用ID                    |
|      biz_id        |      uint32      |            业务ID                    |
|      spec        |      object      |            资源信息       |
|      revision        |      object      |          修改信息        |

#### spec

| 参数名称     | 参数类型   | 描述                                 |
| ------------ | ---------- |------------------------------------|
| credential_scope | string  | credential的规则       |

{% include '_revision.md.j2' %}
