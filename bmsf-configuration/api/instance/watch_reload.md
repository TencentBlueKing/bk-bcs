### 功能描述

watch_reload, 监听reload事件

#### 接口方法

- Path: /instance/v2/watch_reload
- Method: POST

#### 接口参数

| 字段             |  类型     | 必选   |  描述      |
|------------------|-----------|--------|------------|
| seq              |  string   | 是     | 请求序列号 |
| biz_id           |  string   | 是     | 业务ID     |
| app_id           |  string   | 是     | 应用ID     |
| path             |  string   | 是     | 模块实例配置缓存路径 |

#### 请求参数示例

```json
{
    "seq": "xxx",
    "biz_id": "xxx",
    "app_id": "xxx",
    "path": "/data/etc"
}
```

#### 返回结果示例

```json
{
    "seq": "xxx",
    "code": 0,
    "message": "OK",
    "release_name": "new release",
    "multi_release_id" "xxx",
    "release_id": "xxx",
    "reload_type": 0,
    "root_path": "/data/etc",
    "metadatas": [
        {
            "name": "common.yaml",
            "fpath": "/data/etc/"
        },
        {
            "name": "server.yaml",
            "fpath": "/data/etc/server/"
        }
    ]
}
```

#### 返回结果参数

| 字段             | 类型   | 描述     |
|------------------|--------|----------|
| release_name     | string | 版本名称 |
| multi_release_id | string | 混合版本ID, 混合版本回滚时返回  |
| release_id       | string | 版本ID, 非混合版本时返回  |
| reload_type      | int    | reload类型, 0: 发布更新reload  1：回滚reload  2：首次启动同步reload |
| root_path        | string | 模块实例配置缓存路径 |
| metadatas        | array  | 需要reload的配置信息列表 |

##### metadatas[n]

| 字段  | 类型   | 描述     |
|-------|--------|----------|
| name  | string | 配置名称 |
| fpath | string | 配置相对路径 |
