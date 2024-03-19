### 描述

该接口提供版本：v1.0.0+

查询模版被已命名版本服务引用详情

### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述                                                         |
| ----------------- | -------- | ---- | ------------------------------------------------------------ |
| biz_id            | uint32   | 是   | 业务ID                                                       |
| template_space_id | uint32   | 是   | 模版空间ID                                                   |
| template_id       | uint32   | 是   | 模版ID                                                       |
| search_fields     | string   | 否   | 要搜索的字段，search_value有设置时才生效；<br>可支持的字段有app_name(服务名称)、template_revision_name(模版版本名称)、release_name（服务发版名称），默认为app_name；指定多个字段时以逗号分隔，如app_name,template_revision_name |
| search_value      | string   | 否   | 要搜索的值                                                   |
| start             | uint32   | 否   | 分页起始值，默认为0                                          |
| limit             | uint32   | 否   | 分页大小，all参数设为true时可以不设置，否则必须设置          |
| all               | bool     | 否   | 是否查询全量，默认为false，为true时忽略分页相关参数并获取全量数据 |

### 调用示例

```json

```

### 响应示例

```json
{
  "data": {
    "count": 2,
    "details": [
      {
        "template_revision_id": 1,
        "template_revision_name": "v1",
        "app_id": 1,
        "app_name": "service001",
        "release_id": 1,
        "release_name": "v1.0"
      },
      {
        "template_revision_id": 2,
        "template_revision_name": "v2",
        "app_id": 2,
        "app_name": "service002",
        "release_id": 2,
        "release_name": "v2.0"
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

| 参数名称               | 参数类型 | 描述                 |
| ---------------------- | -------- | -------------------- |
| template_revision_id   | uint32   | 模版版本ID           |
| template_revision_name | string   | 模版版本名称         |
| app_id                 | uint32   | 被引用的服务ID       |
| app_name               | string   | 被引用的服务名称     |
| release_id             | uint32   | 被引用的服务版本ID   |
| release_name           | string   | 被引用的服务版本名称 |

