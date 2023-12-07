#### 描述

该接口提供版本：v1.0.0+

获取kv列表

#### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述                                      |
|-------------|--------| ---- |-----------------------------------------|
| biz_id      | uint32 | 是   | 业务ID                                    |
| app_id      | uint32 | 是   | 服务ID                                    |
| key         | string | 否   | 配置键，模糊搜索                                |
| all         | bool   | 否   | 是否拉取全量数据                                |
| start       | uint32 | 否   | 分页起始值                                   |
| limit       | uint32 | 否   | 分页大小                                    |
| with_status | bool   | 否   | 是否返回kv配置项相对上个版本的状态，默认为否                 |
| search_key  | string | 否   | key、更新人、创建人                             |
| kv_type     | string | 否   | 枚举类型，[string、text、number、json、xml、yaml] |

#### 调用示例

```json

```

#### 响应示例

```json
{
    "data": {
        "count": 3,
        "details": [
            {
                "id": 28,
                "spec": {
                    "key": "key_14",
                    "kv_type": "text",
                    "value": "nchbfdghf"
                },
                "attachment": {
                    "biz_id": 2,
                    "app_id": 7
                },
                "revision": {
                    "creator": "canway_demo",
                    "reviser": "canway_demo",
                    "create_at": "2023-11-15T07:50:32Z",
                    "update_at": "2023-11-15T07:50:32Z"
                }
            },
            {
                "id": 27,
                "spec": {
                    "key": "key_13",
                    "kv_type": "string",
                    "value": "1231xddx"
                },
                "attachment": {
                    "biz_id": 2,
                    "app_id": 7
                },
                "revision": {
                    "creator": "canway_demo",
                    "reviser": "canway_demo",
                    "create_at": "2023-11-15T07:50:16Z",
                    "update_at": "2023-11-15T07:50:16Z"
                }
            },
            {
                "id": 26,
                "spec": {
                    "key": "key_12",
                    "kv_type": "number",
                    "value": "111111"
                },
                "attachment": {
                    "biz_id": 2,
                    "app_id": 7
                },
                "revision": {
                    "creator": "canway_demo",
                    "reviser": "canway_demo",
                    "create_at": "2023-11-15T07:50:01Z",
                    "update_at": "2023-11-15T07:50:01Z"
                }
            }
        ]
    }
}
```

#### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

