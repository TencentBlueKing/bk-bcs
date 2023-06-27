#### 描述

该接口提供版本：v1.0.0+

更新服务模版绑定

#### 输入参数

| 参数名称             | 参数类型 | 必选 | 描述                            |
| -------------------- | -------- | ---- | ------------------------------- |
| biz_id               | uint32   | 是   | 业务ID                          |
| app_id               | uint32   | 是   | 应用ID                          |
| binding_id           | uint32   | 是   | 服务模版绑定ID                  |
| template_ids         | []uint32 | 是   | 绑定的模版ID集合，最多500个     |
| template_release_ids | []uint32 | 是   | 绑定的模版版本ID集合，最多500个 |

#### 调用示例

```json
{
  "bindings": [
  {
    "template_set_id": 1,
    "template_release_ids": [
      1,
      2
    ]
  },
  {
    "template_set_id": 3,
    "template_release_ids": [
      5,
      6
    ]
  }
]
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
