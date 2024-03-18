### 描述

该接口提供版本：v1.0.0+

查询latest版本模版被未命名版本服务引用详情

### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述                                                         |
| ----------------- | -------- | ---- | ------------------------------------------------------------ |
| biz_id            | uint32   | 是   | 业务ID                                                       |
| template_space_id | uint32   | 是   | 模版空间ID                                                   |
| template_id       | uint32   | 是   | 模版ID                                                       |
| start             | uint32   | 否   | 分页起始值，默认为0                                          |
| limit             | uint32   | 否   | 分页大小，all参数设为true时可以不设置，否则必须设置          |
| all               | bool     | 否   | 是否查询全量，默认为false，为true时忽略分页相关参数并获取全量数据 |

### 调用示例

```json

```

### 响应示例

```json
{
  "data": {
    "count": 2,
    "details": [
      {
        "template_set_id": 1,
        "template_set_name": "template_set001",
        "app_id": 1,
        "app_name": "service001"
      },
      {
        "template_set_id": 2,
        "template_set_name": "template_set002",
        "app_id": 2,
        "app_name": "service002"
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

| 参数名称 | 参数类型 | 描述                         |
| -------- | -------- | ---------------------------- |
| count    | uint32   | 当前规则能匹配到的总记录条数 |
| detail   | array    | 查询返回的数据               |

#### data.detail[n]

| 参数名称          | 参数类型 | 描述                                   |
| ----------------- | -------- | -------------------------------------- |
| template_set_id   | uint32   | 被服务引用时，模版所位于的模版套餐ID   |
| template_set_name | string   | 被服务引用时，模版所位于的模版套餐名称 |
| app_id            | uint32   | 被引用的服务ID                         |
| app_name          | string   | 被引用的服务名称                       |

