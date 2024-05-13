### 描述

该接口提供版本：v1.0.0+

更新kv

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
| key      | string   | 是   | 配置键 |
| value    | string   | 是   | 配置值 |

### 调用示例

```json
{
    "value": "nchbfdghf"
}
```

### 响应示例

```json
{}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |