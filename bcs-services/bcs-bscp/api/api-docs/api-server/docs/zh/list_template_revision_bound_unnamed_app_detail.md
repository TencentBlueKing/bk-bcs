#### 描述

该接口提供版本：v1.0.0+

查询模版版本被未命名版本服务引用详情

#### 输入参数

| 参数名称            | 参数类型 | 必选 | 描述       |
| ------------------- | -------- | ---- | ---------- |
| biz_id              | uint32   | 是   | 业务ID     |
| template_space_id   | uint32   | 是   | 模版空间ID |
| template_id         | uint32   | 是   | 模版ID     |
| template_revision_id | uint32   | 是   | 模版版本ID |
| start               | uint32   | 是   | 分页起始值 |
| limit               | uint32   | 是   | 分页大小   |

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
        "app_name": "service001"
      },
      {
        "app_id": 2,
        "app_name": "service002"
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

| 参数名称 | 参数类型 | 描述             |
| -------- | -------- | ---------------- |
| app_id   | uint32   | 被引用的服务ID   |
| app_name | uint32   | 被引用的服务名称 |

