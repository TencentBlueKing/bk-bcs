# bcs-user-manager

## 用户管理

### 创建admin user

示例：创建一个名为 "admin1" 的 admin 用户，必须使用一个已有的 admin 用户的 usertoken 才有权限调用。
bcs-user-manager 在启动配置文件中会配置一个初始化的 admin 用户及其 usertoken 。

```shell
curl -X POST -H "Authorization: Bearer {admin-user-token}" http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/users/admin/admin1
```

若调用成功，返回的 code 为 0 ：

```json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": {
        "id": 2,
        "name": "admin1",
        "user_type": 1,
        "user_token": "xxxxxxxxxxx",
        "created_at": "2020-04-16T15:21:52+08:00",
        "updated_at": "2020-04-16T15:21:52+08:00",
        "expires_at": "2030-04-14T15:21:52+08:00"
    }
}
```

### 查询 admin user

示例：查询一个名为 "admin1" 的 admin 用户，必须使用一个已有的 admin 用户的 usertoken 才有权限调用。

```shell
curl -X GET -H "Authorization: Bearer {admin-user-token}" http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/users/admin/admin1
```

若调用成功，返回的 code 为 0 ：

```json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": {
        "id": 2,
        "name": "admin1",
        "user_type": 1,
        "user_token": "JzapN1zU4sjdFVNpOemz0TA7hdaqfilv",
        "created_at": "2020-04-16T15:21:52+08:00",
        "updated_at": "2020-04-16T15:21:52+08:00",
        "expires_at": "2030-04-14T15:21:52+08:00"
    }
}
```

### 创建 saas user

示例：创建一个名为 "saas1" 的 saas 用户，必须使用一个已有的 admin 用户的 usertoken 才有权限调用。

```shell
    curl -X POST -H "Authorization: Bearer {admin-user-token}" http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/users/saas/saas1
```

返回同上。

### 查询 saas user

示例：查询一个名为 "saas1" 的 saas 用户，必须使用一个已有的 admin 用户的 usertoken 才有权限调用。

```shell
curl -X GET -H "Authorization: Bearer {admin-usertoken}" http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/users/saas/saas1
```

返回同上。

### 创建普通用户

示例：创建一个名为 "xxxx" 的普通用户，必须使用一个已有的 admin 或 saas 用户的 usertoken 才有权限调用。

```shell
curl -X POST -H "Authorization: Bearer {admin-usertoken or saas-user-token}" http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/users/plain/xxxx
```

返回同上。

### 查询普通用户

示例：查询一个名为 "xxxx" 的普通用户，必须使用一个已有的 admin 或 saas 用户的 usertoken 才有权限调用。

```shell
curl -X GET -H "Authorization: Bearer {admin-usertoken or saas-user-token}" http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/users/plain/xxxx
```

返回同上。

## usertoken

### 刷新 saas 用户的 usertoken

创建 saas 用户后，会为该用户生成一个 usertoken ，默认的有效期是长期有效(10 years)。可以调用接口刷新该 saas 用户的 usertoken ，必须使用一个已有的
admin 用户的 usertoken 才有权限调用。  

示例： 刷新名为 "saas1" 的 saas 用户的 usertoken ：  

```shell
curl -X PUT -H "Authorization: Bearer {admin-usertoken}" http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/users/saas/saas1/refresh
```

若刷新成功，返回的 code 为 0 ：

```json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": {
        "id": 3,
        "name": "saas1",
        "user_type": 2,
        "user_token": "xxxxxxxxxxxxx",
        "created_at": "2020-04-16T15:22:40+08:00",
        "updated_at": "2020-05-11T20:42:40.115914079+08:00",
        "expires_at": "2030-05-09T20:42:40.114775928+08:00"
    }
}
```

### 刷新普通用户的 usertoken

创建一个普通用户后，会为该用户生成一个 usertoken ，默认的有效期是 24 小时。可以调用接口刷新该普通用户的 usertoken，
可以指定 usertoken 的有效期，调用接口刷新时，如果原有的 usertoken 已经过期，则会生成一个新的 usertoken ，
如果尚未过期，则只刷新过期时间。必须使用一个已有的 admin 或 saas 用户的 usertoken 才有权限调用。  

示例：刷新名为 "xxxx" 的普通用户的 usertoken ，有效期为 2 天:

```shell
curl -X PUT -H "Authorization: Bearer {admin-usertoken or saas-user-token}" http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/users/plain/xxxx/refresh/2
```

返回同上。

## clusters

### 注册集群

cluster_type 可选类型为 k8s, mesos, tke , 当为 tke 类型时，必须同时指定 tke_cluster_id 和 tke_cluster_region 。  

必须使用一个已有的 admin 用户的 usertoken 才有权限调用，使用示例：

```shell
curl -X POST -H "Authorization: Bearer {admin-usertoken}" \
    -H 'content-type: application/json' http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/clusters \
    -d '{"cluster_id":"BCS-K8S-001", "cluster_type":"k8s", "tke_cluster_id":"xxxx", "tke_cluster_region":"shanghai"}'
```

若注册成功，返回的 code 为 0 ：

```json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": {
        "id": "BCS-K8S-001",
        "cluster_type": 1,
        "tke_cluster_id": "",
        "tke_cluster_region": "",
        "creator_id": 1,
        "created_at": "2020-05-11T20:45:51.595077513+08:00"
    }
}
```

### 创建集群的 register-token

必须使用一个已有的 admin 用户的 usertoken 才有权限调用。  
使用示例，为名为 BCS-K8S-001 的集群创建 register-token:  

```shell
    curl -X POST -H "Authorization: Bearer {admin-usertoken}" -H 'content-type: application/json' http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/clusters/BCS-K8S-001/register_tokens
```

若创建成功，返回的 code 为 0 ：

``` json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": {
        "id": 2,
        "cluster_id": "BCS-K8S-001",
        "token": "qL8BiOcYjco2ZJmCPEp0nNmLZ5ITZMeFC0VTIJmLyY1iDDGJUwrNwmZLHCf0fRAPX8Duknn5SJgHnbEiP1GATk3uNGv55J12b7R4i4DUv4MghL4UCfKxLG9iTNrCknnd",
        "created_at": "2020-05-11T20:48:05+08:00"
    }
}
```

### 查询集群的 register-token

必须使用一个已有的 admin 用户的 usertoken 才有权限调用。  
使用示例，查询名为 BCS-K8S-001 的集群的 register-token:  

```shell
    curl -X GET -H "Authorization: Bearer {admin-usertoken}" -H 'content-type: application/json' http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/clusters/BCS-K8S-001/register_tokens
```

返回同上。

### 更新集群的 crendentials

更新集群的 master 地址、证书和 token 信息，用于 bcs-kube-agent 上报集群信息。  

示例：  

```shell
    curl -X PUT -H 'content-type: application/json' http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/clusters/BCS-K8S-001/credentials -d '{"register_token":"qL8BiOcYjco2ZJmCPEp0nNmLZ5ITZMeFC0VTIJmLyY1iDDGJUwrNwmZLHCf0fRAPX8Duknn5SJgHnbEiP1GATk3uNGv55J12b7R4i4DUv4MghL4UCfKxLG9iTNrCknnd", "server_addresses":"https://x.x.x.x:8443", "cacert_data": "xxxx", "user_token":"xxxx"}'
```

若更新成功，返回 code 为 0：

``` json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": null
}
```

### 查询集群的 credentials

查询集群的信息，必须使用一个已有的 admin 用户的 usertoken 才有权限调用，示例：  

```shell
    curl -X GET -H "Authorization: Bearer {admin-usertoken}" http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/clusters/BCS-K8S-001/credentials
```

若查询成功，返回 code 为 0 ：

``` json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": {
        "id": 2,
        "cluster_id": "BCS-K8S-001",
        "server_addresses": "https://x.x.x.x:8443",
        "ca_cert_data": "xxxx",
        "user_token": "xxxx",
        "cluster_domain": "",
        "created_at": "2020-05-11T21:04:43+08:00",
        "updated_at": "2020-05-11T21:04:43+08:00"
    }
}
```

### list 所有集群的 credentials

必须使用一个已有的 admin 用户的 usertoken 才有权限调用，示例：  

```shell
    curl -X GET -H "Authorization: Bearer {admin-usertoken}" http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/clusters/credentials
```

若 list 成功，返回 code 为 0 ：

```json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": {
        "BCS-K8S-100": {
            "server_addresses": "https://x.x.x.x:6553",
            "ca_cert_data": "xxxxxxx",
            "user_token": "xxxxxxx",
            "cluster_domain": ""
        },
        "BCS-K8S-101": {
            "server_addresses": "https://x.x.x.x:8443",
            "ca_cert_data": "xxxx",
            "user_token": "xxxx",
            "cluster_domain": ""
        }
    }
}
```

## 权限

### 授权

必须使用一个已有的 admin 用户的 usertoken 才有权限调用，示例：  
给 xx 授予 BCS-K8S-001 集群的只读角色，给 yy 授予 BCS-K8S-001 集群的只读角色：

```shell
    curl -X POST -H "Authorization: Bearer {admin-usertoken}" -H 'content-type: application/json' http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/permissions -d '{"apiVersion":"v1", "kind":"permissions", "name":"my-permission", "spec":{"permissions":[{"user_name":"xx", "resource_type":"cluster", "resource":"BCS-K8S-001", "role":"viewer"}, {"user_name":"yy", "resource_type":"cluster", "resource":"BCS-K8S-001", "role":"viewer"}]}}'
```

若授权成功，返回 code 为 0 ：

```json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": null
}
```

### 查询权限

查询某个用户对某种类型的资源的权限列表，必须使用一个已有的 admin 用户的 usertoken 才有权限调用。  
示例，查询 xx 用户对 cluster 资源的权限列表：  

```shell
    curl -X GET -H "Authorization: Bearer {admin-usertoken}" -H 'content-type: application/json' http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/permissions -d '{"user_name":"xx", "resource_type":"cluster"}'
```

若查询成功，返回 code 为 0 ：

``` json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": [{
        "resource_type": "cluster",
        "resource": "BCS-K8S-001",
        "role": "manager"
    }, {
        "resource_type": "cluster",
        "resource": "BCS-K8S-002",
        "role": "viewer"
    }]
}
```

### 回收权限

必须使用一个已有的 admin 用户的 usertoken 才有权限调用，示例：  

```shell
    curl -X DELETE -H "Authorization: Bearer {admin-usertoken}" -H 'content-type: application/json' http://0.0.0.0:8080/bcsapi/v4/usermanager/v1/permissions -d '{"apiVersion":"v1", "kind":"permissions", "name":"my-permission", "spec":{"permissions":[{"user_name":"xx", "resource_type":"cluster", "resource":"BCS-K8S-001", "role":"viewer"}, {"user_name":"yy", "resource_type":"cluster", "resource":"BCS-K8S-001", "role":"viewer"}]}}'
```

若回收成功，返回 code 为 0 ：

``` json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": null
}
```

### 校验权限

校验某个用户对某个资源是否有执行某操作的权限，必须使用一个已有的 admin 用户的 usertoken 才有权限调用。  
示例，校验 usertoken 为 xxxxxxx 所对应的用户是否对名为 BCS-K8S-001 的 cluster 有 GET 的权限：  

```shell
    curl -i -X GET -H "Authorization: Bearer {admin-usertoken}" -H 'content-type: application/json' http://0.0.0.0:8081/bcsapi/v4/usermanager/v1/permissions/verify -d '{"user_token":"xxxxxxx", "resource_type":"cluster", "resource":"BCS-K8S-001", "action":"GET"}'
```

若调用成功，返回 code 为 0 ：

```json
{
    "result": true,
    "code": 0,
    "message": "success",
    "data": {
        "allowed": false,
        "message": "no permission"
    }
}
```
