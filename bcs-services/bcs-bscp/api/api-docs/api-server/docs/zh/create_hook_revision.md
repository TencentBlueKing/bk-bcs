### 描述

该接口提供版本：v1.0.0+


创建hook版本，新建的版本状态为“未上线”，hook下存在一个“未上线”版本时，不能创建新的版本

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述     |
| -------- | -------- | ---- | -------- |
| biz_id   | uint32   | 是   | 业务ID   |
| hook_id  | uint32   | 是   | 脚本ID   |
| name     | string   | 否   | 脚本版本号名称，可选项，不填时系统自动生成，生成格式为v20230904033251。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线、点，且必须以中文、英文、数字开头和结尾 |
| memo     | string   | 是   | 版本日志 |
| content  | string   | 是   | 脚本内容 |

### 调用示例

```json
{
    "name": "v1",
    "content": "#!/bin/bash\n\nnow=$(date +'%Y-%m-%d %H:%M:%S')\necho \"hello, start at $now\"\n",
    "memo": "from datetime import datetime"
}
```

### 响应示例

```json
{
    "data": {
        "id": 24
    }
}
```

### 响应参数说明

#### data