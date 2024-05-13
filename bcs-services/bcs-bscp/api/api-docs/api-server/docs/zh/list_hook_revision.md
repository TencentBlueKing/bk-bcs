### 描述

该接口提供版本：v1.0.0+

获取脚本列表

### 输入参数

| 参数名称  | 参数类型 | 必选 | 描述                     |
| --------- | -------- | ---- | ------------------------ |
| biz_id    | uint32   | 是   | 业务ID                   |
| hook_id   | uint32   | 是   | hookID                   |
| searchKey | string   | 否   | 版本号、版本说明、创建人 |
| start     | uint32   | 否   | 分页起始值               |
| limit     | uint32   | 否   | 分页大小                 |
| all       | bool     | 否   | 是否拉取全量数据         |
| state     | string   | 否   | 版本状态查询             |

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
                    "name": "v2",
                    "content": "#!/bin/bash\n\nnow=$(date +'%Y-%m-%d %H:%M:%S')\necho \"hello, start at $now\"\n",
                    "publish_num": 21,
                    "pub_state" : "not_released",
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
        ]
    }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

#### data.details[n]
| 参数名称    | 参数类型 | 描述                                                         |
| ----------- | -------- | ------------------------------------------------------------ |
| biz_id      | uint32   | 业务ID                                                       |
| name        | string   | 版本号                                                       |
| memo        | string   | 版本日志                                                     |
| content     | string   | 脚本内容                                                     |
| publish_num | uint32   | 被引用数，发布一次+1                                         |
| pub_state   | string   | 状态，（枚举值：not_released、partial_released、full_released） |

