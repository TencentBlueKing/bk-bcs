### 功能描述

ping, 探测并获取当前已存在的模块信息

#### 接口方法

- Path: /instance/v2/ping
- Method: POST

#### 接口参数

| 字段 |  类型     | 必选   |  描述      |
|------|-----------|--------|------------|
| seq  |  string   | 是     | 请求序列号 |

#### 请求参数示例

```json
{
    "seq": "xxx"
}
```

#### 返回结果示例

```json
{
    "seq": "xxx",
    "code": 0,
    "message": "OK",
    "mods": [
        {
            "biz_id": "xxx",
            "app_id": "xxx",
            "cloud_id": "0",
            "ip": "127.0.0.1",
            "path": "/data/etc/",
            "labels": {
                "k1": "v1",
                "k2": "v2"
            },
            "is_ready": "true"
        },
        {
            "biz_id": "xxx",
            "app_id": "xxx",
            "cloud_id": "0",
            "ip": "127.0.0.1",
            "path": "/data/etc/",
            "labels": {
                "k1": "v3",
                "k2": "v4"
            },
            "is_ready": "false"
        }
    ]
}
```

#### 返回结果参数

##### mods[n]

| 字段      | 类型   | 描述     |
|-----------|--------|----------|
| biz_id    | string | 业务ID   |
| app_id    | string | 应用ID   |
| cloud_id  | string | 云区域/网络ID |
| ip        | string | 模块实例IP |
| path      | string | 模块实例配置缓存路径 |
| labels    | map    | 模块实例标签KV集合, 例如"version:1.0" |
| is_ready  | bool   | 模块实例是否已就绪可以开始同步配置, 为true则表示已经在做相关同步工作，否则可能尚未开始同步直到业务方主动进行信息注入 |
