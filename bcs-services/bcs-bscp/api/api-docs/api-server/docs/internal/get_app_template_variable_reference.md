### 描述

该接口提供版本：v1.0.0+

查询未命名版本服务变量

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述   |
| -------- | -------- | ---- | ------ |
| biz_id   | uint32   | 是   | 业务ID |
| app_id   | uint32   | 是   | 应用ID |

### 调用示例

```json

```

### 响应示例

```json
{
  "data": {
    "details": [
      {
        "variable_name": "bk_bscp_variable001",
        "references": [
          {
            "id": 1,
            "template_revision_id": 1,
            "name": "server.yaml",
            "path": "/etc"
          },
          {
            "id": 2,
            "template_revision_id": 0,
            "name": "server.yaml",
            "path": "/etc"
          }
        ]
      },
      {
        "variable_name": "bk_bscp_variable002",
        "templates": [
          {
            "id": 3,
            "template_revision_id": 1,
            "name": "server3.yaml",
            "path": "/etc"
          },
          {
            "id": 4,
            "template_revision_id": 0,
            "name": "server4.yaml",
            "path": "/etc"
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

| 参数名称      | 参数类型 | 描述                 |
| ------------- | -------- | -------------------- |
| variable_name | string   | 模版变量名称         |
| references    | array    | 引用该变量的模版信息 |

#### references

| 参数名称             | 参数类型 | 描述                                                         |
| -------------------- | -------- | ------------------------------------------------------------ |
| id                   | uint32   | 配置项ID，template_revision_id为0时为非模版配置项ID，大于0时为模版ID |
| template_revision_id | uint32   | 模版版本ID                                                   |
| name                 | string   | 配置项名称                                                   |
| path                 | string   | 配置项路径                                                   |

