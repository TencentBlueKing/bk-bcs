### 描述

该接口提供版本：v1.0.0+

更新模版变量

### 输入参数

| 参数名称             | 参数类型 | 必选 | 描述           |
| -------------------- | -------- | ---- | -------------- |
| biz_id               | uint32   | 是   | 业务ID         |
| template_variable_id | uint32   | 是   | 模版变量ID     |
| default_val          | string   | 是   | 模版变量默认值 |
| memo                 | string   | 否   | 模版变量描述   |

### 调用示例

```json
{
  "name": "template_variable_001_update",
  "default_val": "5",
  "memo": "an update memo"
}
```

### 响应示例

```json
{
  "data": {}
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

