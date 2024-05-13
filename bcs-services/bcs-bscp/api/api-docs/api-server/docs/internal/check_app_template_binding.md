### 描述

该接口提供版本：v1.0.0+

服务模版绑定冲突检测

### 输入参数

| 参数名称 | 参数类型           | 必选 | 描述                                                         |
| -------- | ------------------ | ---- | ------------------------------------------------------------ |
| biz_id   | uint32             | 是   | 业务ID                                                       |
| app_id   | uint32             | 是   | 应用ID                                                       |
| bindings | []template_binding | 是   | 服务绑定的套餐和模版版本关系，没有指定的，默认使用latest版本，绑定模板版本最多500个 |

##### TemplateBinding

```go
// TemplateBinding is relation between template set id and template revisions
type TemplateBinding struct {
	TemplateSetID     uint32                     `json:"template_set_id"`
	TemplateRevisions []*TemplateRevisionBinding `json:"template_revisions"`
}

// TemplateRevisionBinding is template revision binding
type TemplateRevisionBinding struct {
	TemplateID         uint32 `json:"template_id"`
	TemplateRevisionID uint32 `json:"template_revision_id"`
	IsLatest           bool   `json:"is_latest"`
}
```

### 调用示例

```json
{
  "bindings": [
    {
      "template_set_id": 1,
      "template_revisions": [
        {
          "template_id": 1,
          "template_revision_id": 1,
          "is_latest": true
        },
        {
          "template_id": 2,
          "template_revision_id": 2,
          "is_latest": false
        }
      ]
    },
    {
      "template_set_id": 2,
      "template_revisions": [
        {
          "template_id": 3,
          "template_revision_id": 3,
          "is_latest": true
        },
        {
          "template_id": 4,
          "template_revision_id": 4,
          "is_latest": false
        }
      ]
    }
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
        "template_set_name": "template_set001",
        "tempalte_id": 1,
        "template_name": "template001"
      },
      {
        "template_set_id": 2,
        "template_set_name": "template_set002",
        "tempalte_id": 2,
        "template_name": "template002"
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

| 参数名称            | 参数类型 | 描述         |
| ------------------- | -------- | ------------ |
| template_sapce_id   | uint32   | 模版空间ID   |
| template_sapce_name | string   | 模版空间名称 |
| template_set_id     | uint32   | 模版套餐ID   |
| template_set_name   | string   | 模版套餐名称 |
| template_id         | uint32   | 模版ID       |
| template_name       | string   | 模版名称     |

