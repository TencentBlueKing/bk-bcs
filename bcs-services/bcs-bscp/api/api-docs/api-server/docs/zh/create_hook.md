#### 描述

该接口提供版本：v1.0.0+

创建hook脚本，为某个app设置对应的前置、后置hook脚本。
创建hook脚本时，会默认创建一个脚本版本。

#### 输入参数

| 参数名称     | 参数类型     | 必选   | 描述                                  |
| ------------ | ------------ | ------ |-------------------------------------|
| biz_id       | uint32   | 是   | 业务id                                |
| name         | string   | 是   | 脚本名称                              |
| type         | string   | 是   | hook脚本类型，当前类型有shell、python |
| tag          | string   | 否   | 脚本标签                              |
| memo         | string   | 否   | 脚本描述                              |
| content      | string   | 是   | 脚本内容                              |
| revision_name     | string   | 否   | 脚本版本号名称，可选项，不填时系统自动生成，生成格式为v20230904033251。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线、点，且必须以中文、英文、数字开头和结尾 |

#### 调用示例

```json
{
    "name": "myhook003",
    "type": "shell",
    "tag": "自动化脚本2",
    "content": "#!/bin/bash\n\nnow=$(date +'%Y-%m-%d %H:%M:%S')\necho \"hello, start at $now\"\n",
    "memo": "from datetime import datetime\nprint(\"hello, end at\", datetime.now())\n",
    "revision_name": "v1"
}
```

#### 响应示例

```json
{
    "data": {
        "id": 24
    }
}
```

#### 响应参数说明

#### data

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      id        |      uint32      |            hook脚本ID            |
