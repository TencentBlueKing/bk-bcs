#### 描述

该接口提供版本：v1.0.0+

更新credential的规则

#### 输入参数

| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| credential_id    | uint32       | 是     | credential的ID |
| biz_id | uint32 | 是 | 业务ID |
| add_scope | []string | 否 | 增加的规则 |
| del_id | []uint32 | 否 | 删除规则的id |

#### 调用示例

```json
{
  "credential_id": 6,
  "biz_id":5,
  "add_scope":["XXX","AAAA"],
  "del_id":[9,10]
}
```

#### 响应示例

```json
{
  "data": {}
}
```

#### 响应参数说明

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|       data       |      object      |            响应数据                  |


