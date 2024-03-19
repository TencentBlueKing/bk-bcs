### 描述

该接口提供版本：v1.0.0+

获取脚本引用列表

### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述       |
| ----------- | -------- | ---- | ---------- |
| biz_id      | uint32   | 是   | 业务ID     |
| hook_id     | uint32   | 是   | hookID     |
| start       | uint32   | 否   | 分页起始值 |
| limit       | uint32   | 否   | 分页大小   |

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
                "hook_revision_id": 5,
                "hook_revision_name": "v20230727155603",
                "app_id": 1,
                "app_name": "alkaid-test",
                "release_id": 2,
                "release_name": "v1",
                "type": "post_hook"
            },
            {
                "hook_revision_id": 2,
                "hook_revision_name": "v20230727155603",
                "app_id": 1,
                "app_name": "alkaid-test",
                "release_id": 0,
                "release_name": "未命名版本",
                "type": "post_hook"
            }
        ]
    }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

