### 描述

该接口提供版本：v1.0.0+

获取脚本列表

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述                           |
| -------- | -------- | ---- | ------------------------------ |
| biz_id   | uint32   | 是   | 业务ID                         |
| name     | string   | 否   | 脚本名称，模糊搜索             |
| tag      | string   | 否   | 启用标签搜索，“”时代表拉取全部 |
| not_tag  | bool     | 否   | 未分类                         |
| all      | bool     | 否   | 是否拉取全量数据               |
| start    | uint32   | 否   | 分页起始值                     |
| limit    | uint32   | 否   | 分页大小                       |

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
                "id": 21,
                "spec": {
                    "name": "myhook001",
                    "type": "shell",
                    "tag": "自动化测试",
                    "publish_num": 21,
                    "momo": "from datetime import datetime\nprint(\"hello, end at\", datetime.now())\n",
                },
                "attachment": {
                    "biz_id": 5
                },
                "revision": {
                    "creator": "joelei",
                    "reviser": "joelei",
                    "create_at": "2023-05-19 17:32:09",
                    "update_at": "2023-05-19 17:32:09"
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

