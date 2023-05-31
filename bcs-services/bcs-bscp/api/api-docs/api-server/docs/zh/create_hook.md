#### 描述

该接口提供版本：v1.0.0+

创建hook脚本，为某个app设置对应的前置、后置hook脚本。

#### 输入参数

| 参数名称     | 参数类型     | 必选   | 描述                                  |
| ------------ | ------------ | ------ |-------------------------------------|
| biz_id       | uint32   | 是   | 业务id                                |
| name         | string   | 是   | 脚本名称                              |
| release_name | string   | 是   | 版本号                                |
| type         | string   | 是   | hook脚本类型，当前类型有shell、python |
| tag          | string   | 否   | 脚本标签                              |
| memo         | string   | 否   | 脚本描述                              |
| content      | string   | 是   | 脚本内容                              |

#### 调用示例

```json
{
    "name": "myhook003",
    "release_name": "v1",
    "type": "shell",
    "tag": "自动化脚本2",
    "content": "#!/bin/bash\n\nnow=$(date +'%Y-%m-%d %H:%M:%S')\necho \"hello, start at $now\"\n",
    "memo": "from datetime import datetime\nprint(\"hello, end at\", datetime.now())\n"
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
