#### 描述

该接口提供版本：v1.0.0+

更新模版空间

#### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述         |
| ----------------- | -------- | ---- | ------------ |
| biz_id            | uint32   | 是   | 业务ID       |
| template_space_id | uint32   | 是   | 模版空间ID   |
| name              | string   | 否   | 模版空间名称 |
| memo              | string   | 否   | 模版空间描述 |

#### 调用示例

```json
{
  "name": "template_space_001_update",
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

| 参数名称 | 参数类型   | 描述   |
|------|--------|------|
| data | object | 响应数据 |
