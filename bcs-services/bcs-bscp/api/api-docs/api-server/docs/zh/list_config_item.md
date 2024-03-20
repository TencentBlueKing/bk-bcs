### 描述

该接口提供版本：v1.0.0+

查询未命名版本服务的配置项

### 输入参数

| 参数名称      | 参数类型 | 必选 | 描述                                                                                                                                   |
| ------------- | -------- | ---- |--------------------------------------------------------------------------------------------------------------------------------------|
| biz_id        | uint32   | 是   | 业务ID                                                                                                                                 |
| app_id        | uint32   | 是   | 应用ID                                                                                                                                 |
| search_fields | string   | 否   | 要搜索的字段，search_value有设置时才生效；<br>可支持的字段有name(配置项名称)、path(配置项路径)、memo(配置项描述)、creator(创建人)、reviser(更新人)，默认为name；指定多个字段时以逗号分隔，比如name,path |
| search_value  | string   | 否   | 要搜索的值                                                                                                                                |
| start         | uint32   | 否   | 分页起始值，默认为0                                                                                                                           |
| limit         | uint32   | 否   | 分页大小，all参数设为true时可以不设置，否则必须设置                                                                                                        |
| all           | bool     | 否   | 是否查询全量，默认为false，为true时忽略分页相关参数并获取全量数据                                                                                                |
| with_status   | bool     | 否   | 是否返回模版配置项相对上个版本的状态，默认为否                                                                                                              |

### 响应示例

```json
{
  "data": {
    "count": 1,
    "details": [
      {
        "id":1,
        "config_item_id":1,
        "spec": {
          "name": "server.yaml",
          "path": "/etc",
          "file_type": "yaml",
          "file_mode": "unix",
          "memo": "my—first-config",
          "permission": {
            "user": "root",
            "user_group": "root",
            "privilege": "755"
          }
        },
        "commit_spec": {
          "content_id": 2,
          "content": {
            "signature": "11e3a57c479ebfce651c5871ee70bf61dca74b8e4590b79954126c497a3bfe6b",
            "byte_size": 1675
          },
          "memo": ""
        },
        "file_state": "ADD",
        "attachment": {
          "biz_id": 1,
          "app_id": 1
        },
        "revision": {
          "creator": "tom",
          "reviser": "tom",
          "create_at": "2019-07-29 11:57:20",
          "update_at": "2019-07-29 11:57:20"
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

#### data.details[n]

| 参数名称       | 参数类型 | 描述             |
| -------------- | -------- | ---------------- |
| id             | uint32   | 配置项记录唯一ID |
| config_item_id | uint32   | 配置项ID         |
| spec           | object   | 配置项详情       |
| commit_spec    | object   | 提交详情         |
| attachment     | object   | 配置项关联信息   |
| revision       | object   | 修改信息         |
| file_state     | string   | 配置文件状态     |

##### file_state 的字段说明

	// 增加
	ADD = "ADD"
	//删除
	DELETE = "DELETE"
	//修改
	REVISE = "REVISE"
	//不变
	UNCHANGE = "UNCHANGE"

#### spec

| 参数名称   | 参数类型 | 描述           |
| ---------- | -------- | -------------- |
| name       | string   | 配置项名称     |
| path       | string   | 配置项路径     |
| file_type  | string   | 文件格式       |
| file_mode  | string   | 文件模式       |
| memo       | string   | 备注           |
| permission | object   | 配置项权限信息 |

#### permission

| 参数名称   | 参数类型 | 描述                     |
| ---------- | -------- | ------------------------ |
| user       | string   | 归属用户信息, 例如root   |
| user_group | string   | 归属用户组信息, 例如root |
| privilege  | string   | 文件权限，例如755        |

#### commit_spec

| 参数名称   | 参数类型 | 描述     |
| ---------- | -------- | -------- |
| content_id | uint32   | 内容ID   |
| content    | object   | 内容详情 |
| memo       | string   | 内容描述 |

#### content

| 参数名称  | 参数类型 | 描述             |
| --------- | -------- | ---------------- |
| signature | string   | 文件内容的sha256 |
| byte_size | uint64   | 文件内容的字节数 |

#### attachment

| 参数名称 | 参数类型 | 描述   |
| -------- | -------- | ------ |
| biz_id   | uint32   | 业务ID |
| app_id   | uint32   | 应用ID |

#### revision

| 参数名称  | 参数类型 | 描述                 |
| --------- | -------- | -------------------- |
| creator   | string   | 创建者               |
| reviser   | string   | 最后一次修改的修改者 |
| create_at | string   | 创建时间             |
| update_at | string   | 最后一次修改时间     |
