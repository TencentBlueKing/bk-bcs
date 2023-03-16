#### 描述

该接口提供版本：v1.0.0+

创建hook脚本，为某个app设置对应的前置、后置hook脚本。

#### 输入参数

| 参数名称     | 参数类型     | 必选   | 描述                                  |
| ------------ | ------------ | ------ |-------------------------------------|
| app_id    | uint32       | 是     | 应用ID                                |
| release_id | uint32 | 否 | 版本ID，可选项，不传或传0时则为没有发布的脚本            |
| name | string | 是 | hook脚本名称                            |
| pre_type | string | 否 | 前置hook脚本类型，当前类型有shell、python        |
| pre_hook | string | 否 | 前置hook脚本内容，内容格式需和脚本类型相匹配，为内容字节流对应的字符串 |
| post_type | string | 否    | 后置hook脚本类型，当前类型有shell、python        |
| post_hook | string       | 否     | 后置hook脚本内容，内容格式需和脚本类型相匹配，为内容字节流对应的字符串 |

#### 调用示例

```json
{
  "name": "myhook001",
  "pre_type": "shell",
  "pre_hook": "#!/bin/bash\n\nnow=$(date +'%Y-%m-%d %H:%M:%S')\necho \"hello, start at $now\"\n",
  "post_type": "python",
  "post_hook": "from datetime import datetime\nprint(\"hello, end at\", datetime.now())\n"
}
```

#### 响应示例

```json
{
  "data": {
    "id": 1
  }
}
```

#### 响应参数说明

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      code        |      int32      |            错误码                   |
|      message        |      string      |             请求信息                  |
|       data       |      object      |            响应数据                  |

#### data

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      id        |      uint32      |            hook脚本ID            |
