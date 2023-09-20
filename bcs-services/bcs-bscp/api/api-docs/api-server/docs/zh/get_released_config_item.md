### 描述

该接口提供版本：v1.0.0+

查询单个已命名版本服务的配置项

### 输入参数

| 参数名称       | 参数类型 | 必选 | 描述     |
| -------------- | -------- | ---- | -------- |
| biz_id         | uint32   | 是   | 业务ID   |
| app_id         | uint32   | 是   | 应用ID   |
| release_id     | uint32   | 是   | 发版ID   |
| config_item_id | uint32   | 是   | 配置项ID |

### 调用示例

```json

```

### 响应示例

```json
{
  "data": {
    "config_item": {
      "id": 1,
      "release_id": 1,
      "commit_id": 1,
      "commit_spec": {
        "content_id": 2,
        "content": {
          "signature": "11e3a57c479ebfce651c5871ee70bf61dca74b8e4590b79954126c497a3bfe6b",
          "byte_size": 1675,
          "origin_signature": "53e3a57c479ebfce651c5871ee70bf61dca74b8e4590b79954126c497a3bf23t",
          "origin_byte_size": 2351
        },
        "memo": ""
      },
      "config_item_id": 1,
      "spec": {
        "name": "bscp-config.json",
        "path": "/etc",
        "file_type": "json",
        "file_mode": "unix",
        "memo": "bscp dev config",
        "permission": {
          "user": "root",
          "user_group": "root",
          "privilege": "755"
        }
      },
      "attachment": {
        "biz_id": 2,
        "app_id": 1
      },
      "revision": {
        "creator": "admin",
        "create_at": "2023-02-02 21:01:31"
      }
    }
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| data     | object   | 响应数据 |

#### data

| 参数名称             | 参数类型 | 描述                     |
| -------------------- | -------- | ------------------------ |
| released_config_item | object   | 已命名版本服务配置项详情 |

#### config_item

| 参数名称         | 参数类型 | 描述                                           |
| ---------------- | -------- | ---------------------------------------------- |
| id               | uint32   | 已生成版本配置项记录唯一ID(不是config_item_id) |
| release_id       | uint32   | 版本ID                                         |
| commit_id        | uint32   | 提交ID                                         |
| commit_spec      | object   | 提交详情                                       |
| config_item_id   | uint32   | 配置项ID                                       |
| config_item_spec | object   | 配置项详情                                     |
| attachment       | object   | 关联信息                                       |
| revision         | object   | 修改信息                                       |

#### commit_spec

| 参数名称   | 参数类型 | 描述     |
| ---------- | -------- | -------- |
| content_id | uint32   | 内容ID   |
| content    | object   | 内容详情 |
| memo       | string   | 内容描述 |

#### content

| 参数名称         | 参数类型 | 描述                   |
| ---------------- | -------- | ---------------------- |
| signature        | string   | 渲染后文件内容的sha256 |
| byte_size        | uint64   | 渲染后文件内容的字节数 |
| origin_signature | string   | 渲染前文件内容的sha256 |
| origin_byte_size | uint64   | 渲染前文件内容的字节数 |


#### spec

| 参数名称   | 参数类型 | 描述           |
| ---------- | -------- | -------------- |
| name       | string   | 配置项名称     |
| path       | string   | 配置项路径     |
| file_type  | string   | 文件格式       |
| file_mode  | string   | 文件模式       |
| memo       | string   | 备注           |
| permission | object   | 配置项权限信息 |

#### spec.permission

| 参数名称   | 参数类型 | 描述                     |
| ---------- | -------- | ------------------------ |
| user       | string   | 归属用户信息, 例如root   |
| user_group | string   | 归属用户组信息, 例如root |
| privilege  | string   | 文件权限，例如755        |

#### attachment

| 参数名称 | 参数类型 | 描述   |
| -------- | -------- | ------ |
| biz_id   | uint32   | 业务ID |
| app_id   | uint32   | 应用ID |

#### revision

| 参数名称  | 参数类型 | 描述     |
| --------- | -------- | -------- |
| creator   | string   | 创建者   |
| create_at | string   | 创建时间 |
