### 描述

该接口提供版本：v1.0.0+

更新credential

### 输入参数

| 参数名称   | 参数类型     | 必选 | 描述                  |
|--------| ------------ |----|---------------------|
| biz_id | uint32       | 是  | 业务ID                |
| id     | uint32 | 是  | credential的ID       |
| name   | string | 是  | credential的名称，业务下唯一 |
| memo   | string | 否  | credential的描述       |
| enable | bool | 是  | credential的是否启用     |

### 调用示例

```json
{
  "id": 6,
  "biz_id":5,
  "enable":false,
  "name": "XXXXXX",
  "memo": "XXXXXX"
}
```

### 响应示例

```json
{
  "data": {}
}
```

### 响应参数说明

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|       data       |      object      |            响应数据                  |


