### 资源描述
注册资源的简单描述

### 输入参数说明
|   参数名称   |    参数类型  |  必须  |     参数说明     |
| ------------ | ------------ | ------ | ---------------- |
| app_code   | string | 是 | 应用ID(app id)，可以通过 蓝鲸开发者中心 -> 应用基本设置 -> 基本信息 -> 鉴权信息 获取 |
| app_secret | string | 否 | 安全秘钥(app secret)，可以通过 蓝鲸开发者中心 -> 应用基本设置 -> 基本信息 -> 鉴权信息 获取 |
| namespace | string | 是| 容器的命名空间 |
| pod_name| string | 是| 容器的pod |
| container_name| string | 是| 容器的container名称 |
| operator| string | 是 | 当前使用者|
| command | string | 否 | 自定义web-console启动命令，格式如`"sh -c \"ls -lh && echo test\"` |
| conn_idle_timeout| int64 | 否| 连接空闲时间，超过这个时间会自动断开，单位分钟，必须小于24小时，不填默认30分钟 |
| session_timeout | int64 | 否| 创建的session过期时间，单位分钟，必须小于24小时，不填默认30分钟 |
| viewers | []string |  否 | 共享查看人 |

注：
- container_id 和 namespace, pod_name, container_name 不能同时为空
- 推荐使用namespace, pod_name, container_name，减少检索时间，效率更高

### 请求示例
```json
curl -X POST https://bcs-app.apigw.com/{stag|prod}/apis/projects/BCS项目Code或ID/clusters/集群ID/web_console/sessions/  -d '{"access_token": "access_token", "container_id":"容器ID", "operator": "操作者"}'
```

### 返回结果
```json
{
  "data": {
    "session_id": "5cf0c428ab4a43ecbc61f07595e4694e",
    "web_console_url": "http://bcs.devops.com/backend/web_console/?session_id=5cf0c428ab4a43ecbc61f07595e4694e&container_name=prometheus"
  },
  "code": 0,
  "message": "创建session成功",
  "request_id": "41a2a94799c100cca66329182210441a"
}
```

### 返回结果说明
|   参数名称   |  参数类型  |           参数说明             |
| ------------ | ---------- | ------------------------------ |
|   web_console_url    |    str        |           通过                   |

再跳转请求web_console_url，即可进入容器

注意：
- session_id 有效期通过 session_timeout 设置，有效期内，可重复使用；过期后，请重新申请session_id