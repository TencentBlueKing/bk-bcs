### 描述

获取 BCS 项目列表

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/prod/bcsproject/v1/projects
```

### query 参数
| 参数名称      | 参数类型       | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| projectIDs   | string       | 否     | 项目ID，多个以半角逗号分隔    |
| projectCode  | string       | 否     | 项目英文名，多个以半角逗号分隔 |
| businessID   | string       | 否     | 项目业务ID |
| names        | string       | 否     | 项目中文名称，多个以半角逗号分隔 |
| searchName   | string       | 否     | 项目中文名称，通过此字段模糊查询 |
| kind         | string       | 否     | 项目集群类型，允许 k8s/mesos |
| offset       | int          | 否     | 分页数据，表示第几页          |
| limit        | int          | 否     | 分页数据，表示每页数量        |
| all          | bool         | 否     | 是否查询全量数据             |

### 响应示例
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "total": 1,
        "results": [
            {
                "createTime": "2006-01-02T15:04:05Z",
                "updateTime": "2006-01-02T15:04:05Z",
                "creator": "testuser",
                "updater": "testuser",
                "managers": "testuser",
                "projectID": "1xxx3xxx5xxx4xxx8xxx1xxx2xxxexxx",
                "name": "testproject",
                "projectCode": "testproject",
                "useBKRes": false,
                "description": "test",
                "isOffline": false,
                "kind": "k8s",
                "businessID": "100000",
                "businessName": "xxxxx"
            }
        ],
        "requestID": "894f7249f43c4e5xxxx9207058045e8e",
        "webAnnotations": {
            "perms": {
                "1xxx3xxx5xxx4xxx8xxx1xxx2xxxexxx": {
                    "project_create": true,
                    "project_delete": false,
                    "project_edit": false,
                    "project_view": true
                }
            }
        }
    }
}
```