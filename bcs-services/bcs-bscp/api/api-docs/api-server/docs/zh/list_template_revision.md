### 描述

该接口提供版本：v1.0.0+

查询模版版本列表

### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述                                                         |
| ----------------- | -------- | ---- | ------------------------------------------------------------ |
| biz_id            | uint32   | 是   | 业务ID                                                       |
| template_space_id | uint32   | 是   | 模版空间ID                                                   |
| template_id       | uint32   | 是   | 模版ID                                                       |
| search_fields     | string   | 否   | 要搜索的字段，search_value有设置时才生效；<br>可支持的字段有revision_name(版本名称)、revision_memo(版本描述)，creator(创建人)，默认为revision_name；指定多个字段时以逗号分隔，比如revision_name,revision_memo |
| search_value      | string   | 否   | 要搜索的值                                                   |
| start             | uint32   | 否   | 分页起始值，默认为0                                          |
| limit             | uint32   | 否   | 分页大小，all参数设为true时可以不设置，否则必须设置          |
| all               | bool     | 否   | 是否查询全量，默认为false，为true时忽略分页相关参数并获取全量数据 |

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
        "id": 2,
        "spec": {
          "revision_name": "v20230712150315",
          "revision_memo": "my second version",
          "name": "server.yaml",
          "path": "/etc",
          "file_type": "json",
          "file_mode": "unix",
          "permission": {
            "user": "mysql",
            "user_group": "mysql",
            "privilege": "755"
          },
          "content_spec": {
            "signature": "11e3a57c479ebfae641c5821ee70bf61dca74b8e6596b78950526c397a3b1234",
            "byte_size": "2067"
          }
        },
        "attachment": {
          "biz_id": 2,
          "template_space_id": 1,
          "template_id": 2
        },
        "revision": {
          "creator": "bk-user-for-test-local",
          "create_at": "2023-05-31 16:13:56"
        }
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

| 参数名称 | 参数类型 | 描述                         |
| -------- | -------- | ---------------------------- |
| count    | uint32   | 当前规则能匹配到的总记录条数 |
| detail   | array    | 查询返回的数据               |

#### data.detail[n]

| 参数名称   | 参数类型 | 描述     |
| ---------- | -------- | -------- |
| id         | uint32   | 应用ID   |
| biz_id     | uint32   | 业务ID   |
| spec       | object   | 资源信息 |
| attachment | object   | 关联信息 |
| revision   | object   | 修改信息 |

#### spec

| 参数名称      | 参数类型 | 描述             |
| ------------- | -------- | ---------------- |
| template_name | string   | 模版文件名称     |
| template_path | string   | 模版配置路径     |
| name          | string   | 模板文件版本名称 |
| memo          | string   | 模板文件版本描述 |
| file_type     | string   | 模板文件格式     |
| file_mode     | string   | 模板文件模式     |
| permission    | object   | 模板文件权限信息 |

#### permission

| 参数名称   | 参数类型 | 描述                     |
| ---------- | -------- | ------------------------ |
| user       | string   | 归属用户信息, 例如root   |
| user_group | string   | 归属用户组信息, 例如root |
| privilege  | string   | 文件权限，例如755        |

#### attachment

| 参数名称          | 参数类型 | 描述       |
| ----------------- | -------- | ---------- |
| biz_id            | uint32   | 业务ID     |
| template_space_id | uint32   | 模版空间ID |
| template_id       | uint32   | 模版ID     |

#### revision

| 参数名称  | 参数类型 | 描述     |
| --------- | -------- | -------- |
| creator   | string   | 创建者   |
| create_at | string   | 创建时间 |

