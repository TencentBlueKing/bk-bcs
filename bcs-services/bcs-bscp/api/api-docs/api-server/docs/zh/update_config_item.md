### 描述
该接口提供版本：v1.0.0+
 

更新配置项的属性信息。

### 输入参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| biz_id         | uint32       | 是     | 业务ID     |
| app_id         | uint32       | 是     | 应用ID     |
| id         | uint32       | 是     | 配置项ID     |
| name         | string       | 否     | 配置项名称。最大长度64个字符，仅允许使用英文、数字、下划线、中划线、点，且必须以英文、数字开头和结尾    |
| path         | string       | 是     | 配置项存储的绝对路径。最大长度256个字符，目前仅支持linux路径校验。    |
| file_type         | string       | 否     | 文件格式（枚举值：text、binary）    |
| file_mode         | string       | 否     | 文件模式（枚举值：win、unix）     |
| user         | string       | 否     | 归属用户信息, 例如root    |
| user_group         | string       | 否     | 归属用户组信息, 例如root     |
| privilege         | string       | 否    | 文件权限，例如 755     |
| memo         | string       | 否     | 备注。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾    | 
| sign         | string       | 是     | 配置内容的SHA256，合法长度为64位     |
| byte_size         | uint64       | 是     | 配置内容的大小，单位：字节     |

### 调用示例
```json
{
  "name": "config",
  "path": "/etc",
  "file_type": "text",
  "file_mode": "unix",
  "user": "root",
  "user_group": "root",
  "privilege": "755",
  "memo": "my update app"
}
```

### 响应示例
```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      code        |      int32      |            错误码                   |
|      message        |      string      |             请求信息                  |
