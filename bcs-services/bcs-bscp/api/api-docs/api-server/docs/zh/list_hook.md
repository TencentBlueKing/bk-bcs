### 描述

该接口提供版本：v1.0.0+

获取脚本列表

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述             |
| -------- | -------- | ---- |----------------|
| biz_id   | uint32   | 是   | 业务ID           |
| name     | string   | 否   | 脚本名称，模糊搜索      |
| tag      | string   | 否   | 启用标签搜索，“”时代表拉取全部 |
| not_tag  | bool     | 否   | 未分类            |
| all      | bool     | 否   | 是否拉取全量数据       |
| start    | uint32   | 否   | 分页起始值          |
| limit    | uint32   | 否   | 分页大小           |

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
        "hook": {
          "id": 11,
          "spec": {
            "name": "hook3",
            "type": "shell",
            "tags": [
              "tag1",
              "tag2"
            ],
            "memo": "",
            "content": "",
            "revision_name": ""
          },
          "attachment": {
            "biz_id": 2
          },
          "revision": {
            "creator": "admin",
            "reviser": "admin",
            "create_at": "2024-04-20T06:26:21Z",
            "update_at": "2024-04-20T06:26:21Z"
          }
        },
        "bound_num": 0,
        "confirm_delete": false,
        "published_revision_id": 7
      },
      {
        "hook": {
          "id": 7,
          "spec": {
            "name": "prehook001",
            "type": "shell",
            "tags": [
              "tag1",
              "tag2"
            ],
            "memo": "an update memo",
            "content": "",
            "revision_name": ""
          },
          "attachment": {
            "biz_id": 2
          },
          "revision": {
            "creator": "admin",
            "reviser": "admin",
            "create_at": "2024-04-17T08:31:23Z",
            "update_at": "2024-04-21T14:54:52Z"
          }
        },
        "bound_num": 4,
        "confirm_delete": true,
        "published_revision_id": 8
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

