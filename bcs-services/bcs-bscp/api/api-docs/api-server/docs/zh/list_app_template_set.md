### 查询服务可见的模版套餐

#### API基本信息

API名称：list_app_template_set

API Path：/api/v1/config/biz/{biz_id}/apps/{app_id}/template_sets

Method：GET

#### 描述

该接口提供版本：v1.0.0+

查询服务可见的所有模版套餐列表

#### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述   |
| -------- | -------- | ---- | ------ |
| biz_id   | uint32   | 是   | 业务ID |
| app_id   | uint32   | 是   | 应用ID |

#### 调用示例

```json

```

#### 响应示例

```json
{
  "data": {
    "details": [
      {
        "id": 1,
        "spec": {
          "name": "template_set_001",
          "memo": "my first template set",
          "template_ids": [
            1,
            2
          ],
          "public": true,
          "bound_apps": []
        },
        "attachment": {
          "biz_id": 2,
          "template_space_id": 1
        },
        "revision": {
          "creator": "bk-user-for-test-local",
          "reviser": "bk-user-for-test-local",
          "create_at": "2023-06-05 21:14:45",
          "update_at": "2023-06-05 21:14:45"
        }
      }
    ]
  }
}
```

#### 响应参数说明

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

| 参数名称     | 参数类型 | 描述                   |
| ------------ | -------- | ---------------------- |
| name         | string   | 模版套餐名称           |
| memo         | string   | 模版套餐描述           |
| template_ids | []uint32 | 引用的模版ID列表       |
| public       | bool     | 是否公开对所有服务可见 |
| bound_apps   | []uint32 | 指定可见的服务列表     |

#### attachment

| 参数名称          | 参数类型 | 描述       |
| ----------------- | -------- | ---------- |
| biz_id            | uint32   | 业务ID     |
| template_space_id | uint32   | 模版空间ID |

#### revision

| 参数名称  | 参数类型 | 描述                 |
| --------- | -------- | -------------------- |
| creator   | string   | 创建者               |
| reviser   | string   | 最后一次修改的修改者 |
| create_at | string   | 创建时间             |
| update_at | string   | 最后一次修改时间     |
