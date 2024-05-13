### 描述

该接口提供版本：v1.0.0+

更新未命名版本服务变量

### 输入参数

| 参数名称  | 参数类型   | 必选 | 描述                 |
| --------- | ---------- | ---- | -------------------- |
| biz_id    | uint32     | 是   | 业务ID               |
| app_id    | uint32     | 是   | 应用ID               |
| variables | []variable | 是   | 更新后的服务模版变量 |

#### variable

| 参数名称    | 参数类型 | 必选 | 描述                                                         |
| ----------- | -------- | ---- | ------------------------------------------------------------ |
| name        | string   | 是   | 模版变量名称。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾 |
| type        | string   | 是   | 模版变量类型（枚举值：string、number）                       |
| default_val | string   | 是   | 模版变量默认值                                               |
| memo        | string   | 否   | 模版变量描述。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾 |


### 调用示例

```json
{
  "variables": [
    {
      "name": "bk_bscp_variable001",
      "type": "number",
      "default_val": "3",
      "memo": "my first app template variable"
    },
    {
      "name": "bk_bscp_variable002",
      "type": "string",
      "default_val": "hello",
      "memo": "my second app template variable"
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

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

