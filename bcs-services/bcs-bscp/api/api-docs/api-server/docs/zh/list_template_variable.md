### 描述

该接口提供版本：v1.0.0+

查询模版变量列表

### 输入参数

| 参数名称      | 参数类型 | 必选 | 描述                                                         |
| ------------- | -------- | ---- | ------------------------------------------------------------ |
| biz_id        | uint32   | 是   | 业务ID                                                       |
| search_fields | string   | 否   | 要搜索的字段，search_value有设置时才生效；<br>可支持的字段有name(名称)、memo(描述)，creator(创建人)、reviser(更新人)，默认为name；指定多个字段时以逗号分隔，比如name,memo |
| search_value  | string   | 否   | 要搜索的值                                                   |
| start         | uint32   | 否   | 分页起始值，默认为0                                          |
| limit         | uint32   | 否   | 分页大小，all参数设为true时可以不设置，否则必须设置          |
| all           | bool     | 否   | 是否查询全量，默认为false，为true时忽略分页相关参数并获取全量数据 |

### 调用示例

```json

```

### 响应示例

```json
{
    "data": {
        "count": 1,
        "details": [
            {
                "id": 3,
                "spec": {
                    "name": "bk_bscp_variable001",
                    "type": "number",
                    "default_val": "3",
                    "memo": "my first template variable"
                },
                "attachment": {
                    "biz_id": 2
                },
                "revision": {
                    "creator": "bk-user-for-test-stage",
                    "reviser": "bk-user-for-test-stage",
                    "create_at": "2023-08-21 12:40:16",
                    "update_at": "2023-08-21 12:40:16"
                }
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

| 参数名称   | 参数类型 | 描述     |
| ---------- | -------- | -------- |
| id         | uint32   | 应用ID   |
| biz_id     | uint32   | 业务ID   |
| spec       | object   | 资源信息 |
| attachment | object   | 关联信息 |
| revision   | object   | 修改信息 |

#### spec

| 参数名称    | 参数类型 | 描述                                         |
| ----------- | -------- | -------------------------------------------- |
| name        | string   | 模版变量名称                                 |
| type        | string   | 模版变量类型（枚举值：string、number、bool） |
| default_val | string   | 模版变量默认值                               |
| memo        | string   | 模版变量描述                                 |

#### attachment

| 参数名称 | 参数类型 | 描述   |
| -------- | -------- | ------ |
| biz_id   | uint32   | 业务ID |

#### revision

| 参数名称  | 参数类型 | 描述                 |
| --------- | -------- | -------------------- |
| creator   | string   | 创建者               |
| reviser   | string   | 最后一次修改的修改者 |
| create_at | string   | 创建时间             |
| update_at | string   | 最后一次修改时间     |

