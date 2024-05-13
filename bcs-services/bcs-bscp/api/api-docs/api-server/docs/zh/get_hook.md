### 描述

该接口提供版本：v1.0.0+

创建hook脚本

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述     |
| -------- | -------- | ---- | -------- |
| biz_id   | uint32   | 是   | 业务id   |
| hook_id  | uint32   | 是   | 脚本名称 |

### 调用示例

### 响应示例

```json
{
    "data": {
        "id": 66,
        "spec": {
            "name": "myhook002",
            "type": "shell",
            "tag": "自动化脚本",
            "memo": "from datetime import datetime\nprint(\"hello, end at\", datetime.now())\n",
          	"publish_num": 21,
          	"releases": {
            	    "not_release_id": 0
            }
        },
        "attachment": {
            "biz_id": 5
        },
        "revision": {
            "creator": "joelei",
            "reviser": "joelei",
            "create_at": "2023-05-25 17:25:22",
            "update_at": "2023-05-25 17:25:22"
        }
    }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

#### data

| 参数名称    | 参数类型 | 描述       |
| ----------- | -------- | ---------- |
| id          | uint32   | hook脚本ID |
| name        | string   | 脚本名     |
| type        | string   | 脚本类型   |
| tag         | string   | 脚本标签   |
| memo        | string   | 描述       |
| publish_num | uint32   | 发布次数   |