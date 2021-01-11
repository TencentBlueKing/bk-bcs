### 功能描述

report_reload, 上报reload结果

#### 接口方法

- Path: /instance/v2/report_reload
- Method: POST

#### 接口参数

| 字段             |  类型     | 必选   |  描述      |
|------------------|-----------|--------|------------|
| seq              |  string   | 是     | 请求序列号 |
| biz_id           |  string   | 是     | 业务ID     |
| app_id           |  string   | 是     | 应用ID     |
| path             |  string   | 是     | 模块实例配置缓存路径 |
| release_id       |  string   | 是     | 若上报的为非multi的版本时使用(根据watch_reload接口返回的ID) |
| multi_release_id |  string   | 是     | 若上报的为multi的版本时使用(根据watch_reload接口返回的ID) |
| reload_time      |  string   | 是     | reload时间, 2019-08-29 17:18:22 |
| reload_code      |  int      | 是     | reload执行结果，0: 未执行reload，1：reload成功  2：rollback reload成功  其他值：业务自定义 |
| reload_msg       |  string   | 是     | reload执行结果信息, SUCCESS |

#### 请求参数示例

```json
{
    "seq": "xxx",
    "biz_id": "xxx",
    "app_id": "xxx",
    "path": "/data/etc",
    "multi_release_id": "xxx",
    "reload_time": "2019-08-29 17:18:22",
    "reload_code": 1,
    "reload_msg": "SUCCESS"
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
