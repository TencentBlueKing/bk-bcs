### 描述

该接口提供版本：v1.0.0+

查询业务下的所有模版套餐

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述                                                     |
| -------- | -------- | ---- | -------------------------------------------------------- |
| biz_id   | uint32   | 是   | 业务ID                                                   |
| app_id   | uint32   | 否   | 应用ID，可选项，如果设置，则返回该应用可见的所有模版套餐 |

### 调用示例

```json

```

### 响应示例

```json
{
  "data": {
    "details": [
      {
        "template_space_id": 1,
        "template_space_name": "default_space",
        "template_sets": [
          {
            "template_set_id": 1,
            "template_set_name": "template_set_001",
            "template_ids": [
              1,
              2
            ]
          },
          {
            "template_set_id": 2,
            "template_set_name": "template_set_002",
            "template_ids": [
              3,
              4
            ]
          }
        ]
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

#### data.details[n]

| 参数名称            | 参数类型 | 描述         |
| ------------------- | -------- | ------------ |
| template_space_id   | uint32   | 模版空间ID   |
| template_space_name | string   | 模版空间名称 |
| template_sets       | array    | 模版套餐信息 |

#### template_sets

| 参数名称          | 参数类型 | 描述                     |
| ----------------- | -------- | ------------------------ |
| template_set_id   | uint32   | 模版套餐ID               |
| template_set_name | string   | 模版套餐名称             |
| template_ids      | []uint32 | 模版套餐包含的模版ID列表 |

