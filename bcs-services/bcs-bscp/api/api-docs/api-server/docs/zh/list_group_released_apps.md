### 描述
该接口提供版本：v1.0.0+


查询分组所有已上线应用和对应的版本

### 输入参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| biz_id      | uint32      | 是     | 业务ID     |
| group_id    | uint32      | 是     | 分组ID     |
| start       | uint32      | 是     | 分页起始值  |
| limit       | uint32      | 是     | 分页大小    |

### 响应示例
```json
{
    "data": {
        "count": 2,
        "details": [
            {
                "app_id": 1,
                "app_name": "应用一",
                "released_id": 1,
                "released_name": "v1.0.0",
                "edited": true
            },
            {
                "app_id": 2,
                "app_name": "应用二",
                "released_id": 2,
                "released_name": "v2.1.0",
                "edited": false
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
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      details      |      array      |             查询返回的数据                  |

#### data.details[n]
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
| app_id       | uint32       | 应用ID |
| app_name     | string       | 应用名称 |
| release_id   | uint32       | 版本ID |
| release_name | string       | 版本名称 |
| edited       | bool         | 是否为已编辑状态 |