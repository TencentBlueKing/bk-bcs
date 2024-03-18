### 描述

该接口提供版本：v1.0.0+

创建模版版本

### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述                                                         |
| ----------------- | -------- | ---- | ------------------------------------------------------------ |
| biz_id            | uint32   | 是   | 业务ID                                                       |
| template_space_id | uint32   | 是   | 模版空间ID                                                   |
| template_id       | uint32   | 是   | 模版ID                                                       |
| revision_name     | string   | 否   | 模版版本名称，可选项，不填时系统自动生成，生成格式为v20230904033251。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线、点，且必须以中文、英文、数字开头和结尾 |
| revision_memo     | string   | 否   | 版本描述。最大长度256个字符，仅允许使用中文、英文、数字、下划线、中划线、空格，且必须以中文、英文、数字开头和结尾 |
| file_type         | string   | 是   | 文件格式（枚举值：text、binary）                  |
| file_mode         | string   | 是   | 文件模式（枚举值：win、unix）                                |
| user              | string   | 是   | 文件所属的用户, 例如root                                     |
| user_group        | string   | 是   | 文件所属的用户组, 例如root                                   |
| privilege         | string   | 是   | 文件的权限，例如 755                                         |
| sign              | string   | 是   | 配置内容的SHA256，合法长度为64位                             |
| byte_size         | uint64   | 是   | 配置内容的大小，单位：字节                                   |


### 调用示例

```json
{
  "revision_memo": "my second version",
  "file_type": "text",
  "file_mode": "unix",
  "user": "mysql",
  "user_group": "mysql",
  "privilege": "755",
  "sign": "11e3a57c479ebfae641c5821ee70bf61dca74b8e6596b78950526c397a3b1234",
  "byte_size": 2067
}
```

### 响应示例

```json
{
  "data": {
    "id": 1
  }
}
```

### 响应参数说明

#### data

| 参数名称 | 参数类型 | 描述       |
| -------- | -------- | ---------- |
| id       | uint32   | 模版版本ID |

