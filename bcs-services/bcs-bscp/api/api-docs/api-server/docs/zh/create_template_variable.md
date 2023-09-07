#### API基本信息

API名称：create_template_variable

API Path：/api/v1/config/biz/{biz_id}/template_variables

Method：POST

#### 描述

该接口提供版本：v1.0.0+

创建模版变量

#### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述                                                         |
| ----------- | -------- | ---- | ------------------------------------------------------------ |
| biz_id      | uint32   | 是   | 业务ID                                                       |
| name        | string   | 是   | 模版变量名称。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾 |
| type        | string   | 是   | 模版变量类型（枚举值：string、number、bool）                 |
| default_val | string   | 是   | 模版变量默认值的json串                                       |
| memo        | string   | 否   | 模版变量描述。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾 |

#### 调用示例

```json
{
  "name": "bk_bscp_variable001",
  "memo": "my first template space"
}
```

#### 响应示例

```json
{
  "data": {
    "id": 1
  }
}
```

#### 响应参数说明

#### data

| 参数名称 | 参数类型 | 描述       |
| -------- | -------- | ---------- |
| id       | uint32   | 模版变量ID |
