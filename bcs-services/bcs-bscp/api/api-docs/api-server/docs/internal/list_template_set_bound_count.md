### 描述

该接口提供版本：v1.0.0+

查询模版套餐被引用计数

### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述                              |
| ----------------- | -------- | ---- | --------------------------------- |
| biz_id            | uint32   | 是   | 业务ID                            |
| template_space_id | uint32   | 是   | 模版空间ID                        |
| template_set_ids  | []uint32 | 是   | 要查询的模版套餐ID列表，最多200个 |

### 调用示例

```json
{
  "template_set_ids": [
    1,
    2
  ]
}
```

### 响应示例

```json
{
  "data": {
    "details": [
      {
        "template_set_id": 1,
        "bound_unnamed_app_count": 2,
        "bound_named_app_count": 3,
      },
      {
        "template_set_id": 2,
        "bound_unnamed_app_count": 5,
        "bound_named_app_count": 6,
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

#### data.detail[n]

| 参数名称                | 参数类型 | 描述                               |
| ----------------------- | -------- | ---------------------------------- |
| template_set_id         | uint32   | 模版套餐ID                         |
| bound_unnamed_app_count | uint32   | 模版套餐被未命名版本服务引用的数量 |
| bound_named_app_count   | uint32   | 模版套餐被已命名版本服务引用的数量 |

