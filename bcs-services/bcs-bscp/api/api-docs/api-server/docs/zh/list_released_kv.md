### 描述

该接口提供版本：v1.0.0+

查询已命名版本服务的kv配置项

### 输入参数

| 参数名称      | 参数类型 | 必选 | 描述                                                         |
| ------------- | -------- | ---- | ------------------------------------------------------------ |
| biz_id        | uint32   | 是   | 业务ID                                                       |
| app_id        | uint32   | 是   | 应用ID                                                       |
| release_id    | uint32   | 是   | 发版ID                                                       |
| start      | uint32   | 否   | 分页起始值，默认为0                                          |
| limit      | uint32   | 否   | 分页大小，all参数设为true时可以不设置，否则必须设置          |
| all        | bool     | 否   | 是否查询全量，默认为false，为true时忽略分页相关参数并获取全量数据 |

### 调用示例

```json

```

### 响应示例

```json
{
    "data": {
        "count": 3,
        "details": [
            {
                "id": 4,
                "release_id": 11,
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
                    "create_at": "2023-11-15T07:51:35Z",
                    "update_at": "2023-11-15T07:51:35Z"
                }
            },
            {
                "id": 3,
                "release_id": 11,
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
                    "create_at": "2023-11-15T07:51:35Z",
                    "update_at": "2023-11-15T07:51:35Z"
                }
            },
            {
                "id": 2,
                "release_id": 11,
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
                    "create_at": "2023-11-15T07:51:35Z",
                    "update_at": "2023-11-15T07:51:35Z"
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

