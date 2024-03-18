### 描述

该接口提供版本：v1.0.0+

在kv类型的服务中创建kv

**kv_type**类型

- string
- text
- number
- json
- xml
- yaml



### 输入参数

| 参数名称     | 参数类型     | 必选   | 描述                                  |
| ------------ | ------------ | ------ |-------------------------------------|
| biz_id       | uint32   | 是   | 业务id                                |
| app_id  | uint32 | 是   | 服务id                          |
| key       | string   | 是   | 配置键 |
| kv_type    | string   | 是 | kv类型                          |
| value     | string   | 是  | 配置值                           |

### 调用示例

```json
{
    "key": "key_14",
    "kv_type": "text",
    "value": "nchbfdghf"
}
```

### 响应示例

```json
{
    "data": {
        "id": 24
    }
}
```

### 响应参数说明

#### data

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      id        |      uint32      |            kvID            |
