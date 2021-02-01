### 功能描述

inject, 注入模块实例信息

#### 接口方法

- Path: /instance/v2/inject
- Method: POST

#### 接口参数

| 字段     |  类型     | 必选   |  描述      |
|----------|-----------|--------|------------|
| seq      |  string   | 是     | 请求序列号 |
| biz_id   |  string   | 是     | 业务ID     |
| app_id   |  string   | 是     | 应用ID     |
| path     |  string   | 是     | 模块实例配置缓存路径 |
| labels   |  map      | 是     | 模块实例标签KV集合, 例如"version:1.0" |

#### 请求参数示例

```json
{
    "seq": "xxx",
    "biz_id": "xxx",
    "app_id": "xxx",
    "path": "/data/etc",
    "labels": {
        "version":"1.0"
    }
}
```

#### 返回结果示例

```json
{
    "seq": "xxx",
    "code": 0,
    "message": "OK"
}
```
