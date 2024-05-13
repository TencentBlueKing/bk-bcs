### 描述

该接口提供版本：v1.0.0+


更新hook版本，新建的版本状态为“未上线”，hook下存在一个“未上线”版本时，不能创建新的版本

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述     |
| -------- | -------- | ---- | -------- |
| biz_id   | uint32   | 是   | 业务ID   |
| hook_id  | uint32   | 是   | 脚本ID   |
| name     | string   | 是   | 版本号   |
| memo     | string   | 是   | 版本日志 |
| content  | string   | 是   | 脚本内容 |

### 调用示例

```json
{
    "name": "myhook001",
    "content": "#!/bin/bash\n\nnow=$(date +'%Y-%m-%d %H:%M:%S')\necho \"hello, start at $now\"\n",
    "memo": "from datetime import datetime"
}
```

### 响应示例

```json
{
    "data": {}
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |