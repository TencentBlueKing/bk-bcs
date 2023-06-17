#### 描述

该接口提供版本：v1.0.0+

更新模版套餐

#### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述         |
| ----------------- | -------- | ---- | ------------ |
| biz_id            | uint32   | 是   | 业务ID       |
| template_space_id | uint32   | 是   | 模版空间ID   |
| template_set_id   | uint32   | 是   | 模版套餐ID   |
| template_ids      | []uint32 | 是   | 引用的模版ID列表，最大限制500个                                        |
| memo              | string   | 否   | 模版套餐描述。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾 |
| public           | bool         | 是     | 是否公开对所有服务可见，为true时，则忽略入参bound_apps                           |
| bound_apps       | []uint32     | 否     | 指定可见的服务列表                                                    |

#### 调用示例

```json
{
  "memo": "an update memo",
  "template_ids": [
    1,
    2
  ],
  "public": true,
  "bound_apps": []
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
