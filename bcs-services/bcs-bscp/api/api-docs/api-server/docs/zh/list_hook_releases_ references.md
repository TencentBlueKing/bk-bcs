#### 描述

该接口提供版本：v1.0.0+

获取脚本引用列表

#### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述       |
| ----------- | -------- | ---- | ---------- |
| biz_id      | uint32   | 是   | 业务ID     |
| hook_id     | uint32   | 是   | hookID     |
| releases_id | uint32   | 是   | 版本id     |
| start       | uint32   | 否   | 分页起始值 |
| limit       | uint32   | 否   | 分页大小   |

#### 调用示例

```json

```

#### 响应示例

```json
{
    "data": {
        "count": 1,
        "details": [
            {
                "hook_release_name": "v1",
                "app_name": "demo-34",
                "config_release_name": "v3",
                "config_release_id": 1,
                "state": "not_released"
            }
        ]
    }
}
```

#### 响应参数说明

