### 描述

获取 BCS 项目详情

### 路径参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| all         | bool       | 否     | 是否列举全部项目（包括无权限的项目）     |
| searchKey   | string     | 否     | 按关键字搜索（项目名称，项目英文名，项目ID），仅在 all 为 true 时生效 |
| offset   | int     | 否     | all=true时为偏移量；all=false时为页码，从0开始 |
| limit   | int     | 否     | 每页的数量 |
| kind    | string | 否 | 枚举值 [k8s, mesos]

该接口有两种表现：
1. all 为 `true` 时，返回带分页的全部项目，以及用户对每个项目的权限信息 `web_annotations`
2. all 为 `false` 时，默认返回全部用户有查看权限的项目；默认不分页，当offset不为0或limit不为0时，返回分页数据

> 通过 X-Bcs-Username 请求头代理用户查询用户有权限的项目

### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --header 'X-Bcs-Username: testuser' --insecure https://bcs-api-gateway.apigw.com/prod/bcsproject/v1/projects/authorized_projects
```

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
        "requestID": "894f7249f43c4e5xxxx9207058045e8e"
    }
}
```