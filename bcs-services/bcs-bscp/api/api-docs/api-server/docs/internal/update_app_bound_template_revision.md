### 描述

该接口提供版本：v1.0.0+

更新服务绑定的模版版本

### 输入参数

| 参数名称   | 参数类型           | 必选 | 描述                                                         |
| ---------- | ------------------ | ---- | ------------------------------------------------------------ |
| biz_id     | uint32             | 是   | 业务ID                                                       |
| app_id     | uint32             | 是   | 应用ID                                                       |
| binding_id | uint32             | 是   | 服务模版绑定ID                                               |
| bindings   | []template_binding | 是   | 服务绑定的套餐和模版版本关系，绑定模板版本最多500个，属于存量更新，只更新指定套餐下的模版版本，套餐下存在其他未指定的模版版本将保持不变，将更新的模版版本在套餐下不存在将报错 |

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
  "data": {}
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

