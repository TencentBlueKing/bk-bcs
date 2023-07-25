#### API基本信息

API名称：update_template_variable

API Path：/api/v1/config/biz/{biz_id}/template_variables/{template_variable_id}

Method：PUT

#### 描述

该接口提供版本：v1.0.0+

更新模版变量

#### 输入参数

| 参数名称             | 参数类型 | 必选 | 描述                   |
| -------------------- | -------- | ---- | ---------------------- |
| biz_id               | uint32   | 是   | 业务ID                 |
| template_variable_id | uint32   | 是   | 模版变量ID             |
| default_val          | string   | 是   | 模版变量默认值的json串 |
| memo                 | string   | 否   | 模版变量描述           |

#### 调用示例

```json
{
  "memo": "an update memo"
}
```

#### 响应示例

```json
{
  "data": {}
}
```

#### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

