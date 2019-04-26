# k8s api 使用说明

## rest api

### CreateUser
#### 描述
创建用户
#### 请求地址
- /rest/users/
#### 请求方式
- POST
#### 请求参数
Content-Type: application/json
```json
{
  "user_name": ""
}
```
#### 请求示例
curl  -X POST -H "Content-Type: application/json" -H  "Authorization: Bearer {admin user_token}" http://bcs_server:8080/rest/users/ -d '{"user_name": "xx"}'
#### 返回结果
```json
{
 "ID": 6,
 "Name": "plain:xx",
 "IsSuperUser": false,
 "CreatedAt": "2019-02-22T17:05:55.611242586+08:00",
 "BackendType": "",
 "BackendCredentials": null
}
```

### QueryBCSUserByName
#### 描述
查询用户
#### 请求地址
- /rest/users/{user_name}
#### 请求方式
- GET
#### 请求参数
Content-Type: application/json
#### 请求示例
curl  -X GET -H "Content-Type: application/json" -H  "Authorization: Bearer {admin user_token}" http://bcs_server:8080/rest/users/xx
#### 返回结果
```json
{
 "ID": 6,
 "Name": "plain:xx",
 "IsSuperUser": false,
 "CreatedAt": "2019-02-22T17:05:56+08:00",
 "BackendType": "",
 "BackendCredentials": null
}
```

### CreateUserToken
#### 描述
创建user_token
#### 请求地址
- /rest/users/{user_id}/tokens
#### 请求方式
- POST
#### 请求参数
Content-Type: application/json
#### 请求示例
curl  -X POST -H "Content-Type: application/json" -H  "Authorization: Bearer {admin user_token}" http://bcs_server:8080/rest/users/6/tokens
#### 返回结果
```json
{
 "ID": 4,
 "UserId": 6,
 "Type": 3,
 "Value": "zztCQlWdGMqM8sxdfWVLjbl6PWb7yMjI",
 "ExpiresAt": "2029-02-19T17:45:46.521486741+08:00",
 "CreatedAt": "2019-02-22T17:45:46.52166867+08:00"
}
```

### CreateBCSCluster
#### 描述
创建集群
#### 请求地址
- /rest/clusters/bcs
#### 请求方式
- POST
#### 请求参数
Content-Type: application/json
```json
{
  "id":"", 
  "project_id":""
}
```
#### 请求示例
curl -X POST -H "Content-Type: application/json" -H  "Authorization: Bearer {user_token}" http://bcs_server:8080/rest/clusters/bcs -d '{"id":"k8s-002", "project_id":"project-002"}'
#### 参数返回
```json
{
 "id": "bcs-k8s-002-P4E2Tj68",
 "provider": 2,
 "creator_id": 1,
 "identifier": "bcs-k8s-002-p4e2tj68-yky9VRNXsLIcv6Yd",
 "created_at": "2019-02-25T10:50:12+08:00",
 "turn_on_admin": false
}
```

### QueryBCSClusterByID
#### 描述
查询集群
#### 请求地址
/rest/clusters/bcs/query_by_id/
#### 请求方式
- GET
#### 请求参数
Content-Type: application/json
#### 请求示例
curl -X GET -H "Content-Type: application/json" -H  "Authorization: Bearer {user_token}" http://bcs_server:8080/rest/clusters/bcs/query_by_id/?project_id=project-002\&cluster_id=k8s-002
#### 返回结果
```json
{
 "id": "bcs-k8s-002-P4E2Tj68",
 "provider": 2,
 "creator_id": 1,
 "identifier": "bcs-k8s-002-p4e2tj68-yky9VRNXsLIcv6Yd",
 "created_at": "2019-02-25T10:50:12+08:00",
 "turn_on_admin": false
}
```

### CreateRegisterToken
#### 描述
创建 register_token
#### 请求地址
/rest/clusters/{cluster_id}/register_tokens
#### 请求方式
- POST
#### 请求参数
Content-Type: application/json
#### 请求示例
curl -X POST -H "Content-Type: application/json" -H  "Authorization: Bearer {user_token}" http://bcs_server:8080/rest/clusters/bcs-k8s-002-P4E2Tj68/register_tokens
#### 返回结果
```json
[
 {
  "id": 2,
  "cluster_id": "bcs-k8s-002-P4E2Tj68",
  "token": "GOSGAZK2ikFL3lxDZeMX2GTQzgT04rc7TQ6xqyjT15wI8aQpeeov9DXIUaEwhVY2JlMHOfl7Zgdt9VFXIDBXzJkANWeAR1OxS8tQQOBzUphaS257evTMarLRWeuDCULj",
  "created_at": "2019-02-25T11:07:40+08:00"
 }
]
```

### UpdateCredentials
#### 描述
更新 k8s 集群的地址和证书信息，用于 bcs-kube-agent 上报集群信息
#### 请求地址
/rest/clusters/{cluster_id}/credentials
#### 请求方式
- PUT
#### 请求参数
Content-Type: application/json
```json
{
  "register_token": "", 
  "server_addresses": "", 
  "cacert_data": "", 
  "user_token": ""
}
```
#### 请求示例
curl -X PUT -H "Content-Type: application/json" -H  "Authorization: Bearer {user_token}" http://bcs_server:8080/rest/clusters/bcs-k8s-002-P4E2Tj68/credentials -d '{"register_token": "", "server_addresses": "http://10.0.0.0:6553", "cacert_data": "xx", "user_token": "xx"}'
#### 返回结果
```json
{}
```

### GetClientCredentials
#### 描述
获取集群证书信息
#### 请求地址
/rest/clusters/{cluster_id}/client_credentials
#### 请求方式
- GET
#### 请求参数
Content-Type: application/json
#### 请求示例
curl -X GET -H "Content-Type: application/json" -H  "Authorization: Bearer {user_token}" http://bcs_server:8080/rest/clusters/bcs-k8s-002-P4E2Tj68/client_credentials
#### 返回结果
```json
{
 "cluster_id": "bcs-k8s-002-P4E2Tj68",
 "server_address": "http://bcs_server:8080/tunnels/clusters/bcs-k8s-002-p4e2tj68-yky9VRNXsLIcv6Yd/",
 "user_token": "y99n7HTyYcrmsYSyXOFZcOBNydRBHtEc",
 "cacert_data": "xx"
}
```

## k8s api
通过 bcs-api 调用 k8s api 时，只需在原有的 k8s api 上加上前缀即可。
### k8s version
例如，查询 k8s version 的 api。
#### 请求方式
- GET
#### 请求示例
curl -X GET -H  "Authorization: Bearer {user_token}" http://bcs_server:8080/tunnels/clusters/bcs-k8s-001-opcnwwki-v6kLYNdxIafSOhJI/version
#### 返回结果
```json
{
  "major": "1",
  "minor": "12",
  "gitVersion": "v1.12.3",
  "gitCommit": "435f92c719f279a3a67808c80521ea17d5715c66",
  "gitTreeState": "clean",
  "buildDate": "2018-11-26T12:46:57Z",
  "goVersion": "go1.10.4",
  "compiler": "gc",
  "platform": "linux/amd64"
}
```