### 描述

该接口提供版本：v1.0.0+

创建应用。

### 输入参数

| 参数名称                | 参数类型     | 必选   | 描述             |
|---------------------| ------------ | ------ | ---------------- |
| biz_id              | uint32       | 是                                        | 业务ID     |
| name                | string       | 是                                        | 应用名称。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾     |
| config_type         | string       | 是                                        | 应用配置类型（枚举值：file、kv，目前仅支持file、kv类型）。应用配置类型限制了本应用下所有配置的类型，如果选择file类型，本应用下所有配置只能为file类型 |
| mode                | string       | 是                                        | app的实例在消费配置时的模式（枚举值：normal、namespace），详情见下方描述。    |
| memo                | string       | 否                                        | 备注。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾    |
| reload_type         | string       | 选填，仅在 config_type 为 file 类型下使用    | sidecar通知app重新Reload配置的方式（枚举值：file，目前仅支持file类型）    |
| reload_file_path    | string       | 选填，仅在 reload_type 为 file 类型下使用     | Reload文件绝对路径（绝对路径 + 文件名），最大长度为128字节 |
| alias            | string   | 必填                                      | 服务别名。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头和结尾 |
| data_type | string | 选填，仅在config_type为kv 类型下生效 | 数据类型（枚举值：any、string、number、string、text、yaml、json、xml） |

#### 参数说明：

##### deploy_type:

**Common**：应用程序可以直接从BSCP服务器拉取配置信息。

##### mode:

app工作的工作模式决定了该app下的实例消费配置数据的方式，支持两种工作模式：

- Normal模式： 提供基础的范围发布能力，用户的管理成本比较低，发布过程简单。这也是bscp推荐用户使用的方式。

  该模式是最通用的管理模式，在该管理模式下所有的策略均不使用namespace。在该模式下限制策略集下的策略最大数量为5个，包括兜底策略。

- Namespace模式： 提供复杂的，大批量的范围发布的管理模式。但用户的管理成本略高，适合场景特别复杂，策略集下的策略特别多的场景。 具体特点为：
    1. 在该模式下策略集下的所有策略都必须有一个独立的namespace，且所有的namespace值在该策略集下都是唯一的。
    2. 实例在拉取配置时，请求中必须带所属的namespace信息，如果不带，则bscp会直接拒绝该请求。
    3. 该模式下，提供兜底策略管理能力，每个策略集下有且只能有一个兜底策略。
    4. 该模式下，策略集下策略的总量限制为<=200。

##### reload_type:

如果应用的配置是文件类配置，且通过 bscp sidecar 绑定的应用实例，获取应用实例匹配的最新配置版本，并进行下载的话。那么，在 bscp sidecar 下载完成之后，需要通过一种方式 (reload_type)
去通知应用实例去加载配置文件，以及相关的配置信息。

###### File类型：

bscp sidecar 会将下载好的配置信息，写到用户指定的 reload 文件（reload_file_path）当中，应用程序通过这个 reload 文件来获取最新的配置信息。

### 调用示例（config_item）

```json
{
  "name": "myapp",
  "config_type": "file",
  "mode": "normal",
  "memo": "my_first_app",
  "reload_type": "file",
  "reload_file_path": "/data/reload.json",
  "alias":"appAlias"
}
```

### 调用示例（kv）

```json
{
    "biz_id": "myapp",
    "name": "kv",
    "config_type": "kv",
    "mode": "normal",
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
