### 描述

该接口提供版本：v1.0.0+

批量查询模版的版本名称

### 输入参数

| 参数名称     | 参数类型 | 必选 | 描述                  |
| ------------ | -------- | ---- | --------------------- |
| biz_id       | uint32   | 是   | 业务ID                |
| template_ids | []uint32 | 是   | 模版ID列表，最多200个 |

### 调用示例

```json
{
  "template_ids": [
    1,
    2
  ]
}
```

### 响应示例

```json
{
  "data": {
    "details": [
      [
        {
          "template_id": 1,
          "template_name": "template001",
          "latest_template_revision_id": 2,
          "latest_revision_name": "v1",
          "latest_signature": "11e3a57c479ebfae641c5821ee70bf61dca74b8e6596b78950526c397a3b1234",
          "latest_byte_size": 2067,
          "template_revisions": [
            {
              "template_revision_id": 1,
              "template_revision_name": "v20230815120105",
              "template_revision_memo": "my revision for test1"
            },
            {
              "template_revision_id": 2,
              "template_revision_name": "v20230815130206",
              "template_revision_memo": "my revision for test2"
            }
          ]
        },
        {
          "template_id": 2,
          "template_name": "template002",
          "latest_template_revision_id": 4,
          "latest_revision_name": "v2",
          "latest_signature": "22e3a57c479ebfae641c5821ee70bf61dca74b8e6596b78950526c397a3b1253",
          "latest_byte_size": 1023,
          "template_revisions": [
            {
              "template_revision_id": 3,
              "template_revision_name": "v20230815140307",
              "template_revision_memo": "my revision for test3"
            },
            {
              "template_revision_id": 4,
              "template_revision_name": "v20230815150408",
              "template_revision_memo": "my revision for test2"
            }
          ]
        }
      ]
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

#### data.detail[n]

| 参数名称                    | 参数类型 | 描述                       |
| --------------------------- | -------- | -------------------------- |
| template_id                 | uint32   | 模版ID                     |
| template_name               | string   | 模版名称                   |
| latest_template_revision_id | uint32   | 最新模版版本ID             |
| latest_sinature             | string   | 最新模版版本内容的sha256   |
| latest_byte_size            | uint64   | 最新模版版本内容的字节大小 |
| template_revisions          | object   | 模版版本信息               |

#### template_revisions

| 参数名称               | 参数类型 | 描述         |
| ---------------------- | -------- | ------------ |
| template_revision_id   | uint32   | 模版版本ID   |
| template_revision_name | string   | 模版版本名称 |
| template_revision_memo | string   | 模版版本描述 |
