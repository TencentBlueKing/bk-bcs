#### 描述

该接口提供版本：v1.0.0+

更新hook脚本

#### 输入参数

| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| app_id    | uint32       | 是     | 应用ID |
| id | uint32 | 是 | 脚本ID |
| release_id | uint32 | 否 | 版本ID |
| name | string | 否 | hook脚本名称 |
| pre_type | string | 否 | 前置hook脚本类型，当前类型有shell、python |
| pre_hook | string | 否 | 前置hook脚本内容，内容格式需和脚本类型相匹配，为内容字节流对应的字符串 |
| post_type | string | 否    | 后置hook脚本类型，当前类型有shell、python |
| post_hook | string       | 否     | 后置hook脚本内容，内容格式需和脚本类型相匹配，为内容字节流对应的字符串 |

#### 调用示例

```json
{
  "name": "myhook001_update",
  "pre_type": "shell",
  "pre_hook": "#!/bin/bash\n\nnow=$(date +'%Y-%m-%d %H:%M:%S')\necho \"hello, update test, start at $now\"\n",
  "post_type": "python",
  "post_hook": "from datetime import datetime\nprint(\"hello, update test, end at\", datetime.now())\n"
}
```

#### 响应示例

```json
{
  "data": {}
}
```

#### 响应参数说明

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|       data       |      object      |            响应数据                  |


