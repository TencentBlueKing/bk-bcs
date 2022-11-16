### 描述
该接口提供版本：v1.0.0+
 
查询应用最新的版本信息。
匹配策略优先级为： 实例发布 > 策略子策略 > 策略主策略
单文件下载路径为：data.repository + data.config_items[n].repository_spec.path

### 输入参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| biz_id         | uint32       | 是     | 业务ID     |
| app_id         | uint32       | 是     | 应用ID     |
| uid         | string       | 是     | App下该实例的唯一标识。最大长度64个字符，仅允许使用英文、数字、下划线、中划线，且必须以英文、数字开头  |
| namespace         | string       | 否     | 命名空间，应用策略集是 namespace 类型，该值必填。最大长度128个字符，仅允许使用中文、英文、数字、下划线、中划线，且必须以中文、英文、数字开头 |
| labels         | map<string, string>   | 否  | 标签  |

#### 字段说明
##### uid
uid 是app下该实例的唯一标识，必填。

该字段会用于查询当前时刻是否有该实例的实例发布。
##### namespace
namespace 是命名空间，选填。

如果当前应用策略集是工作在Namespace模式下，该字段会用于匹配策略集下策略的 namespace 字段，且该字段为必填参数。
##### labels
labels 是标签，选填。

如果当前应用策略集是工作在Namespace模式下，该字段只能用于匹配策略集下策略的子策略的 selector 字段。

如果当前应用策略集是工作在Normal模式下，该字段会用于匹配策略集下策略的主/子策略的 selector 字段。

### 调用示例
```json
{
  "biz_id": 1,
  "app_id": 1,
  "uid": "4fc82b26aecb47d2868c4efbe3581732a3e7cbcc6c2efb32062c08170a05eeb8",
  "namespace": "module1.set1",
  "labels": {
    "module": "1",
    "name": "game",
    "set": "1"
  }
}
```

### 响应示例
```json
{
  "code": 0,
  "message": "",
  "data": {
    "release_id": 4,
    "repository": {
      "root": "http://127.0.0.1:8888/generic/bscp/bscp-v1-310"
    },
    "config_items": [{
      "rci_id": 20,
      "commit_spec": {
        "content_id": 20,
        "content": {
          "signature": "c7d78b78205a2619eb2b80558f85ee18a8836ef5f4f317f8587ee38bc3712a8a",
          "byte_size": 11
        }
      },
      "config_item_spec": {
        "name": "da73bd0c-c6c0-11ec-bed7-5254007a012b.yaml",
        "path": "/etc",
        "file_type": "yaml",
        "file_mode": "unix",
        "permission": {
          "user": "root",
          "user_group": "root",
          "privilege": "0755"
        }
      },
      "repository_spec": {
        "path": "/4/c7d78b78205a2619eb2b80558f85ee18a8836ef5f4f317f8587ee38bc3712a8a"
      }
    }]
  }
}
```

### 响应参数说明
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      code        |      int32      |            状态码                   |
|      message        |      string      |             请求信息                  |
|       data       |      object      |            响应数据                  |

#### data
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      release_id        |      uint32      |            版本ID                  |
|      repository        |      object      |            仓库信息                  |
|      config_items        |      array object      |             配置项列表                  |

#### repository
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      root        |      string      |            配置下载根路径                  |

#### config_items
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|      released_config_item_id        |      uint32      |            版本快照ID                  |
|      commit_id        |      uint32      |             提交记录ID                  |
|      commit_spec        |      object      |             提交记录资源信息                 |
|      config_item_id        |      uint32      |             配置项ID                  |
|      config_item_spec        |      object      |           配置项资源信息                 |
|      repository_spec        |      object      |             仓库元数据信息                 |

#### commit_spec
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
| content_id         | uint32            | 配置内容ID     |
|      content        |      object      |            配置项元数据信息                   |

#### content
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
| signature         | string          | 配置内容的SHA256     |
| byte_size         | uint32           | 配置内容的大小，单位：字节     |

#### config_item_spec
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|name	|string	|	配置项名称 |
|path	|string	|	配置项路径 |
|file_type	|string	|	文件格式 |
|file_mode	|string	|	文件模式 |
|      permission        |      object      |            配置项权限信息                   |

#### permission
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
|user	|string	|	归属用户信息, 例如root |
|user_group|	string	|	归属用户组信息, 例如root |
|privilege|	string	|	文件权限，例如755 |

#### repository_spec
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
| path	|string	|	配置下载子路径  |
