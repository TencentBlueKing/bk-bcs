### 描述
该接口提供版本：v1.0.0+


更新应用。

### 输入参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| id         | uint32       | 是     | 应用ID     |
| biz_id         | uint32       | 是     | 业务ID     |
| name         | string       | 是     | 应用名称。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾     |
| memo         | string       | 否     | 备注。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾    |
| reload_type         | string       | 选填，仅在 config_type 为 file 类型下使用    | sidecar通知app重新Reload配置的方式（枚举值：file，目前仅支持file类型）    |
| reload_file_path         | string       | 选填，仅在 reload_type 为 file 类型下使用     | Reload文件绝对路径（绝对路径 + 文件名），最大长度为128字节  |
| alias            | string   | 必填                                      | 服务别名。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾 |
| data_type        | string   | 选填，仅在config_type为kv 类型下生效      | 数据类型（枚举值：any、string、number、string、text、yaml、json、xml） |

#### 参数说明：
##### reload_type:
如果应用的配置是文件类配置，且通过 bscp sidecar 绑定的应用实例，获取应用实例匹配的最新配置版本，并进行下载的话。那么，在 bscp sidecar
下载完成之后，需要通过一种方式 (reload_type) 去通知应用实例去加载配置文件，以及相关的配置信息。

###### File类型：
bscp sidecar 会将下载好的配置信息，写到用户指定的 reload 文件（reload_file_path）当中，应用程序通过这个 reload 文件来获取最新的配置信息。



### 调用示例（config_items）
```json
{
    "name": "update_app",
    "memo": "my_update_app",
 		"alias":"appAlias"
}
```

### 调用示例（kv）

```json
{
    "name": "update_app",
    "memo": "my_update_app",
  	"config_type":"any",
 		"alias":"appAlias"
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
