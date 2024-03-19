### 描述

该接口提供版本：v1.0.0+

创建credential

### 输入参数

| 参数名称   | 参数类型     | 必选 | 描述                  |
|--------| ------------ |----|---------------------|
| biz_id | uint32       | 是  | 业务ID                |
| name   | string | 是  | credential的名称，业务下唯一 |
| memo   | string | 否  | credential的描述       |
| scope  | []string | 否  | credential的匹配规则     |

### 调用示例

```json
{
  "biz_id":5,
  "scope":[
    "mysql",
    "origin/*"
  ],
  "name":"test_name",
  "memo":"test"
}
```

### 响应示例

```json
{
  "data": {
    "id": 1
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

#### data

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      id        |      uint32      |            credential的ID            |