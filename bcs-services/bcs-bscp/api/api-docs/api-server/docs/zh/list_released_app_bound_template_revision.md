### 描述

该接口提供版本：v1.0.0+

查询已命名版本服务绑定的模版版本

### 输入参数

| 参数名称      | 参数类型 | 必选 | 描述                                                         |
| ------------- | -------- | ---- | ------------------------------------------------------------ |
| biz_id        | uint32   | 是   | 业务ID                                                       |
| app_id        | uint32   | 是   | 应用ID                                                       |
| release_id    | uint32   | 是   | 发版ID                                                       |
| search_fields | string   | 否   | 要搜索的字段，search_value有设置时才生效；<br>可支持的字段有revision_name(版本名称)、revision_memo(版本描述)、name(模版配置项名称)、path(模版配置项路径)、creator(创建人)，默认为revision_name；指定多个字段时以逗号分隔，比如revision_name,revision_memo |
| search_value  | string   | 否   | 要搜索的值                                                   |

### 调用示例

```json

```

### 响应示例

```json
{
  "data": {
    "count": 1,
    "details": [
      {
        "template_space_id": 1,
        "template_space_name": "template_space001",
        "template_set_id": 1,
        "template_set_name": "template_set001",
        "template_revisions": [
          {
            "template_id": 1,
            "name": "server.yaml",
            "path": "/etc",
            "template_revision_id": 1,
            "is_latest": true,
            "template_revision_name": "v20230712150315",
            "template_revision_memo": "my second version",
            "file_type": "json",
            "file_mode": "unix",
            "user": "mysql",
            "user_group": "mysql",
            "privilege": "755",
            "signature": "11e3a57c479ebfae641c5821ee70bf61dca74b8e6596b78950526c397a3b1234",
            "byte_size": "2067",
            "origin_signature": "11e3a57c479ebfae641c5821ee70bf61dca74b8e6596b78950526c397a3b1234",
            "origin_byte_size": "2067",            
            "creator": "bk-user-for-test-local",
            "create_at": "2023-05-31 16:13:56"
          }
        ]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述           |
| -------- | -------- | -------------- |
| detail   | array    | 查询返回的数据 |

#### data.details[n]

| 参数名称            | 参数类型 | 描述         |
| ------------------- | -------- | ------------ |
| template_sapce_id   | uint32   | 模版空间ID   |
| template_sapce_name | string   | 模版空间名称 |
| template_set_id     | uint32   | 模版套餐ID   |
| template_set_name   | string   | 模版套餐名称 |

#### template_revisions

| 参数名称               | 参数类型 | 描述                     |
| ---------------------- | -------- | ------------------------ |
| template_id            | uint32   | 模版ID                   |
| name                   | string   | 模版文件名称             |
| path                   | string   | 模版配置路径             |
| template_revision_id   | uint32   | 模版版本ID               |
| is_latest              | bool     | 是否为最新模版版本       |
| template_revision_name | string   | 模板文件版本名称         |
| template_revision_memo | string   | 模板文件版本描述         |
| file_type              | string   | 模板文件格式             |
| file_mode              | string   | 模板文件模式             |
| user                   | string   | 归属用户信息, 例如root   |
| user_group             | string   | 归属用户组信息, 例如root |
| privilege              | string   | 文件权限，例如755        |
| signature              | string   | 渲染后文件内容的sha256   |
| byte_size              | uint64   | 渲染后文件内容的字节数   |
| origin_signature       | string   | 渲染前文件内容的sha256   |
| origin_byte_size       | uint64   | 渲染前文件内容的字节数   |
| creator                | string   | 创建者                   |
| create_at              | string   | 创建时间                 |

