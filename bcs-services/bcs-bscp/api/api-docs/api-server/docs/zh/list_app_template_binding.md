### 描述

该接口提供版本：v1.0.0+

查询服务模版绑定列表

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
    "count": 1,
    "details": [
      {
        "id": 1,
        "spec": {
          "template_space_ids": [
            1
          ],
          "template_set_ids": [
            1,
            2
          ],
          "template_ids": [
            1,
            2,
            3,
            4
          ],
          "template_revision_ids": [
            1,
            2,
            3,
            4
          ],
          "latest_template_revision_ids": [
            1,
            3
          ],
          "bindings": [
            {
              "template_set_id": 1,
              "template_revisions": [
                {
                  "template_id": 1,
                  "template_revision_id": 1,
                  "is_latest": true
                },
                {
                  "template_id": 2,
                  "template_revision_id": 2,
                  "is_latest": false
                }
              ]
            },
            {
              "template_set_id": 2,
              "template_revisions": [
                {
                  "template_id": 3,
                  "template_revision_id": 3,
                  "is_latest": true
                },
                {
                  "template_id": 4,
                  "template_revision_id": 4,
                  "is_latest": false
                }
              ]
            }
          ]
        },
        "attachment": {
          "biz_id": 2,
          "app_id": 1
        },
        "revision": {
          "creator": "bk-user-for-test-local",
          "reviser": "bk-user-for-test-local",
          "create_at": "2023-05-31 15:50:20",
          "update_at": "2023-05-31 16:43:09"
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

| 参数名称   | 参数类型 | 描述           |
| ---------- | -------- | -------------- |
| id         | uint32   | 服务模版绑定ID |
| biz_id     | uint32   | 业务ID         |
| spec       | object   | 资源信息       |
| attachment | object   | 关联信息       |
| revision   | object   | 修改信息       |

#### spec

| 参数名称                     | 参数类型           | 描述                         |
| ---------------------------- | ------------------ | ---------------------------- |
| template_space_ids           | []uint32           | 服务绑定的模版空间列表       |
| template_set_ids             | []uint32           | 服务绑定的模版套餐列表       |
| template_ids                 | []uint32           | 服务绑定的模版列表           |
| template_revision_ids        | []uint32           | 服务绑定的模版版本列表       |
| latest_template_revision_ids | []uint32           | 服务绑定的最新模版版本列表   |
| bindings                     | []template_binding | 服务绑定的套餐和模版版本关系 |

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

