### 描述

该接口提供版本：v1.0.0+

查询模版版本被引用计数

### 输入参数

| 参数名称              | 参数类型 | 必选 | 描述                              |
| --------------------- | -------- | ---- | --------------------------------- |
| biz_id                | uint32   | 是   | 业务ID                            |
| template_space_id     | uint32   | 是   | 模版空间ID                        |
| template_id           | uint32   | 是   | 模版ID                            |
| template_revision_ids | []uint32 | 是   | 要查询的模版版本ID列表，最多200个 |

### 调用示例

```json
{
  "template_revision_ids": [
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
        "template_revision_id": 1,
        "bound_unnamed_app_count": 2,
        "bound_named_app_count": 3,
      },
      {
        "template_revision_id": 2,
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
| template_id             | uint32   | 模版ID                             |
| bound_unnamed_app_count | uint32   | 模版版本被未命名版本服务引用的数量 |
| bound_named_app_count   | uint32   | 模版版本被已命名版本服务引用的数量 |

