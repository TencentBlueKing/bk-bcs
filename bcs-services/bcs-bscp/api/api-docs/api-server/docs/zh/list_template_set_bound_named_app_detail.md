#### 描述

该接口提供版本：v1.0.0+

查询模版套餐被已命名版本服务引用详情

#### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述       |
| ----------------- | -------- | ---- | ---------- |
| biz_id            | uint32   | 是   | 业务ID     |
| template_space_id | uint32   | 是   | 模版空间ID |
| template_set_id   | uint32   | 是   | 模版套餐ID |
| start             | uint32   | 是   | 分页起始值 |
| limit             | uint32   | 是   | 分页大小   |

#### 调用示例

```json

```

#### 响应示例

```json
{
  "data": {
    "count": 2,
    "details": [
      {
        "app_id": 1,
        "app_name": "service001",
        "release_id": 1,
        "release_name": "v1.0"
      },
      {
        "app_id": 2,
        "app_name": "service002",
        "release_id": 2,
        "release_name": "v2.0"        
      }
    ]
  }
}
```

#### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述                         |
| -------- | -------- | ---------------------------- |
| count    | uint32   | 当前规则能匹配到的总记录条数 |
| detail   | array    | 查询返回的数据               |

#### data.detail[n]

| 参数名称     | 参数类型 | 描述                 |
| ------------ | -------- | -------------------- |
| app_id       | uint32   | 被引用的服务ID       |
| app_name     | uint32   | 被引用的服务名称     |
| release_id   | uint32   | 被引用的服务版本ID   |
| release_name | uint32   | 被引用的服务版本名称 |

