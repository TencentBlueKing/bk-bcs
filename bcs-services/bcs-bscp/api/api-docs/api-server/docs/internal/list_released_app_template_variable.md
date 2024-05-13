### 描述

该接口提供版本：v1.0.0+

查询已命名版本服务变量

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述   |
| ---------- | -------- | ---- | ------ |
| biz_id     | uint32   | 是   | 业务ID |
| app_id     | uint32   | 是   | 应用ID |
| release_id | uint32   | 是   | 发版ID |

### 调用示例

```json

```

### 响应示例

```json
{
  "data": {
    "details": [
      {
        "name": "bk_bscp_variable001",
        "type": "number",
        "default_val": "3",
        "memo": "my first template variable"
      },
      {
        "name": "bk_bscp_variable002",
        "type": "string",
        "default_val": "hello",
        "memo": "my second template variable"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述           |
| -------- | -------- | -------------- |
| detail   | array    | 查询返回的数据 |

#### data.details[n]

| 参数名称    | 参数类型 | 描述                                         |
| ----------- | -------- | -------------------------------------------- |
| name        | string   | 模版变量名称                                 |
| type        | string   | 模版变量类型（枚举值：string、number、bool） |
| default_val | string   | 模版变量默认值                               |
| memo        | string   | 模版变量描述                                 |

