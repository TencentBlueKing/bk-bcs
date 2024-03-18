### 描述

该接口提供版本：v1.0.0+

获取hook下的版本信息

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述   |
| ---------- | -------- | ---- | ------ |
| biz_id     | uint32   | 是   | 业务ID |
| hook_id    | uint32   | 是   | hookID |
| revision_id | uint32   | 是   | 版本ID |

### 调用示例

```json

```

### 响应示例

```json
{
    "data": {
        "details": {
                "id": 21,
                "spec": {
                    "name": "v2",
                    "content": "#!/bin/bash\n\nnow=$(date +'%Y-%m-%d %H:%M:%S')\necho \"hello, start at $now\"\n",
                    "publish_num": 21,
                    "state" : "deployed",
                    "momo": "v2 memo",
                },
                "attachment": {
                    "biz_id": 5,
                    "hook_id": 21
                },
                "revision": {
                    "creator": "joelei",
                    "reviser": "joelei",
                    "create_at": "2023-05-19 17:32:09",
                    "update_at": "2023-05-19 17:32:09"
                }
            }
    }
}
```

### 响应参数说明

