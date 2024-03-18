### 描述

该接口提供版本：v1.0.0+

获取脚本版本引用列表

### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述       |
| ----------- | -------- | ---- | ---------- |
| biz_id      | uint32   | 是   | 业务ID     |
| hook_id     | uint32   | 是   | hookID     |
| revision_id | uint32   | 是   | 版本id     |
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
                "app_id": 1,
                "app_name": "alkaid-test",
                "release_id": 2,
                "release_name": "v1",
                "type": "post_hook"
            },
            {
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

