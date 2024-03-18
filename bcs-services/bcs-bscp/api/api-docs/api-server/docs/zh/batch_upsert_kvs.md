### 描述

该接口提供版本：v1.0.0+

批量导入｜更新 kv

**kv_type**类型

- string
- text
- number
- json
- xml
- yaml



### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述   |
| -------- | -------- | ---- | ------ |
| biz_id   | uint32   | 是   | 业务id |
| app_id   | uint32   | 是   | 服务id |
| kvs      | obj      | 是   |        |
| key      | string   | 是   | 配置键 |
| kv_type  | string   | 是   | kv类型 |
| value    | string   | 是   | 配置值 |

### 调用示例

```json
{
    "kvs": [
        {
            "key": "key_1",
            "kv_type": "string",
            "value": "11112"
        },
        {
            "key": "key_2",
            "kv_type": "string",
            "value": "111"
        },
        {
            "key": "key_3",
            "kv_type": "number",
            "value": "111"
        }
    ]
}
```

### 响应示例

```json
{
    "data": {}
}
```

### 响应参数说明

#### data

| 参数名称 | 参数类型 | 描述 |
| -------- | -------- | ---- |