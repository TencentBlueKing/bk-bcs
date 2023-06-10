### 描述

该接口提供版本：v1.0.0+

获取服务配置脚本

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述   |
| -------- | -------- | ---- | ------ |
| biz_id   | uint32   | 是   | 业务ID |
| app_id   | uint32   | 是   | 应用ID |

### 调用示例

### 响应示例

```json
{
    "data": {
        "config_hook": {
            "id": 1,
            "spec": {
                "pre_hook_id": 1,
                "pre_hook_release_id": 1,
                "post_hook_id": 0,
                "post_hook_release_id": 0
            },
            "attachment": {
                "biz_id": 5,
                "app_id": 34
            },
            "revision": {
                "creator": "joelei",
                "reviser": "joelei",
                "create_at": "2023-06-09 16:37:43",
                "update_at": "2023-06-09 16:55:27"
            }
        }
    }
}
```

### 响应参数说明

data

| 参数名称             | 参数类型 | 必选 | 描述           |
| -------------------- | -------- | ---- | -------------- |
| pre_hook_id          | int      | 否   | 前置脚本ID     |
| pre_hook_release_id  | int      | 否   | 前置脚本版本ID |
| post_hook_id         | int      | 否   | 后置脚本ID     |
| post_hook_release_id | int      | 否   | 后置脚本版本ID |

