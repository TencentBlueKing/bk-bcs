### 描述

该接口提供版本：v1.0.0+

创建应用。

### 输入参数

| 参数名称                | 参数类型     | 必选   | 描述             |
|---------------------| ------------ | ------ | ---------------- |
| biz_id              | uint32       | 是                                        | 业务ID     |
| name                | string       | 是                                        | 应用名称。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾     |
| config_type         | string       | 是                                        | 应用配置类型（枚举值：file、kv，目前仅支持file、kv类型）。应用配置类型限制了本应用下所有配置的类型，如果选择file类型，本应用下所有配置只能为file类型 |
| memo                | string       | 否                                        | 备注。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾    |
| alias            | string   | 必填                                      | 服务别名。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾 |
| data_type | string | 选填，仅在config_type为kv 类型下生效 | 数据类型（枚举值：any、string、number、string、text、yaml、json、xml） |


### 调用示例（config_item）

```json
{
  "name": "myapp",
  "config_type": "file",
  "memo": "my_first_app",
  "alias":"appAlias"
}
```

### 调用示例（kv）

```json
{
    "biz_id": "myapp",
    "name": "kv",
    "config_type": "kv",
    "memo": "",
    "data_type":"any",
    "alias":"appAlias"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1
  }
}
```

### 响应参数说明

| 参数名称     | 参数类型         | 描述                           |
| ------------ |--------------| ------------------------------ |
|      code        | int32        |            错误码                   |
|      message     | string       |             请求信息                  |
|       data       | object       |            响应数据                  |

#### data

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      id        |      uint32      |            应用ID                    |
