# bcs-user-manager 命令行工具

## 配置文件

配置文件默认放在 `/etc/bcs/bcs-user-manager.yaml` 文件：
```yaml
config:
  apiserver: "${BCS APISERVER地址}"
  bcs_token: "${Token信息}"
```

## 使用文档

### 创建Admin用户 - CreateAdminUser

```bash
kubectl-bcs-user-manager create admin-user --help
```

参数详情:

```yaml 
-u, --user_name string   "用户名称，通过该字段创建admin用户信息"
```

示例:

```
kubectl-bcs-user-manager create admin-user -u [user_name to create]
kubectl-bcs-user-manager create au -u [user_name to create]
```



### 获取Admin用户 - GetAdminUser

```bash
kubectl-bcs-user-manager get admin-user --help
```
参数详情:
```yaml 
-u, --user_name string   "用户名称，通过该字段查询admin用户信息"
```

示例:

```
kubectl-bcs-user-manager get admin-user -u [user_name to query]
kubectl-bcs-user-manager get au -u [user_name to query]
```



### 创建Saas用户 - CreateSaasUser

```bash
kubectl-bcs-user-manager create saas-user --help
```

参数详情:

```yaml 
-u, --user_name string   "用户名称，通过该字段创建saas用户信息"
```

示例:

```
kubectl-bcs-user-manager create saas-user -u [user_name to create]
kubectl-bcs-user-manager create su -u [user_name to create]
```



### 获取Saas用户 - GetSaasUser

```bash
kubectl-bcs-user-manager get saas-user --help
```

参数详情:

```yaml 
-u, --user_name string   "用户名称，通过该字段查询saas用户信息"
```

示例:

```
kubectl-bcs-user-manager get saas-user -u [user_name to query]
kubectl-bcs-user-manager get su -u [user_name to query]
```





### 刷新Saas用户token - RefreshSaasToken

```bash
kubectl-bcs-user-manager update saas-token --help
```

参数详情:

```yaml 
-u, --user_name string   "用户名称，通过该字段刷新saas用户token信息"
```

示例:

```
kubectl-bcs-user-manager update saas-token -u [user_name]
kubectl-bcs-user-manager update st -u [user_name]
```





### 创建PlainUser - CreatePlainUser

### 

```bash
kubectl-bcs-user-manager create plain-user --help
```

参数详情:

```yaml 
-u, --user_name string   "用户名称，通过该字段创建plain用户信息"
```

示例:

```
kubectl-bcs-user-manager create plain-user -u [user_name to create]
kubectl-bcs-user-manager create pu -u [user_name to create]
```



### 获取PlainUser - GetPlainUser

```bash
kubectl-bcs-user-manager get plain-user --help
```

参数详情:

```yaml 
-u, --user_name string   "用户名称，通过该字段查询plain用户信息"
```

示例:

```
kubectl-bcs-user-manager get plain-user -u [user_name to query]
kubectl-bcs-user-manager get pu -u [user_name to query]
```





### 刷新Plain用户Token - RefreshPlainToken

### 

```bash
kubectl-bcs-user-manager update plain-token --help
```

参数详情:

```yaml 
-u, --user_name string   "用户名称，通过该字段创建admin用户信息"
-t, --expire_time string   "过期时间，过期天数，整数 >=0,0为立即过期"
```

示例:

```
kubectl-bcs-user-manager update plain-token -u [user_name] -t [expire_time]
kubectl-bcs-user-manager update pt -u [user_name] -t [expire_time]
```









### 创建集群 - CreateCluster

### 

```bash
kubectl-bcs-user-manager create cluster --help
```

参数详情:

```yaml 
-b, --cluster-body string   "json类型"

cluster-body json说明
{
  "cluster_id": "",
  "cluster_type": "",    //string类型,值范围["k8s","mesos","tke"]
  "tke_cluster_id": "",
  "tke_cluster_region": ""
}


```

示例:

```bash
kubectl-bcs-user-manager create cluster --cluster-body '{"cluster_id":"","cluster_type":"", "tke_cluster_id":"","tke_cluster_region":""}' 
```



### 创建集群token - CreateRegisterToken

### 

```bash
kubectl-bcs-user-manager create register-tokenr --help
```

参数详情:

```yaml 
-i, --cluster_id string   "集群id，通过该字段创建register-token信息"
```

示例:

```
kubectl-bcs-user-manager create register-token --cluster_id [cluster_id]
kubectl-bcs-user-manager create rt --cluster_id [cluster_id]
```



### 获取集群token - GetRegisterToken

### 

```bash
kubectl-bcs-user-manager get register-token --help
```

参数详情:

```yaml 
-i, --cluster_id string   "集群id，通过该字段获取register-token信息"
```

示例:

```
kubectl-bcs-user-manager get register-token --cluster_id [cluster_id]
kubectl-bcs-user-manager get rt --cluster_id [cluster_id]
```



### 更新credentials- UpdateCredentials

### 

```bash
kubectl-bcs-user-manager update credentials --help
```

参数详情:

```yaml 
-i, --cluster_id       string   "集群id，通过该字段更新credential信息"
-f, --credentials_form string   "json类型"

credentials_form  json说明
{
  "register_token": "",
  "server_addresses": "",
  "cacert_data": "",
  "user_token": ""
}


```

示例:

```bash
kubectl-bcs-user-manager update credentials --cluster_id [cluster_id] --credentials_form ' {"register_token":"","server_addresses":"","cacert_data":"","user_token":""}' 
```



### 获取credentials- GetCredentials

### 

```bash
kubectl-bcs-user-manager get credentials --help
```

参数详情:

```yaml 
-i, --cluster_id       string   "集群id，通过该字段获取credential信息"
```

示例:

```
kubectl-bcs-user-manager get credentials --cluster_id [cluster_id]
kubectl-bcs-user-manager get c --cluster_id [cluster_id]
```



### 获取credentials列表- ListCredentials

### 

```bash
kubectl-bcs-user-manager list credentials --help
```

参数详情:

```yaml 
无参数
```

示例:

```
kubectl-bcs-user-manager list credentials
kubectl-bcs-user-manager list c
```





### 授权 - GrantPermission

### 

```bash
kubectl-bcs-user-manager grant permission --help
```

参数详情:

```yaml 
-f, --permission_form string   "json类型"

permission_form  json说明
{
  "apiVersion": "",
  "kind": "",
  "metadata": {
    "name": "",
    "namespace": "",
    "creationTimestamp": "0001-01-01T00:00:00Z",
    "labels": {          //map[string]string
      "a": "a"
    },
    "annotations": {     //map[string]string
      "a": "a"
    },
    "clusterName": ""
  },
  "spec": {
    "permissions": [
      {
        "user_name": "",
        "resource_type": "",
        "resource": "",
        "role": ""
      }
    ]
  }
}
```

示例:

```
kubectl-bcs-user-manager grant permission --permission_form '{
  "apiVersion": "",
  "kind": "",
  "metadata": {
    "name": "",
    "namespace": "",
    "creationTimestamp": "0001-01-01T00:00:00Z",
    "labels": {
      "a": "a"
    },
    "annotations": {
      "a": "a"
    },
    "clusterName": ""
  },
  "spec": {
    "permissions": [
      {
        "user_name": "",
        "resource_type": "",
        "resource": "",
        "role": ""
      }
    ]
  }
}' 
```



### 获取权限 - GetPermission

### 

```bash
kubectl-bcs-user-manager get permission --help
```

参数详情:

```yaml 
-f, --permission_form string   "json类型"

permission_form  json说明
{
  "user_name": "",
  "resource_type": ""
}
```

示例:

```
kubectl-bcs-user-manager get permission -f '{"user_name":"","resource_type":""}' 
```



### 撤销权限 - RevokePermission

### 

```bash
kubectl-bcs-user-manager delete permission --help
```

参数详情:

```yaml 
-f, --permission_form string   "json类型"

permission_form  json说明
{
  "apiVersion": "",
  "kind": "",
  "metadata": {
    "name": "",
    "namespace": "",
    "creationTimestamp": "0001-01-01T00:00:00Z",
    "labels": {          //map[string]string
      "a": "a"
    },
    "annotations": {     //map[string]string
      "a": "a"
    },
    "clusterName": ""
  },
  "spec": {
    "permissions": [
      {
        "user_name": "",
        "resource_type": "",
        "resource": "",
        "role": ""
      }
    ]
  }
}
```

示例:

```
kubectl-bcs-user-manager delete permission --permission_form '{
  "apiVersion": "",
  "kind": "",
  "metadata": {
    "name": "",
    "namespace": "",
    "creationTimestamp": "0001-01-01T00:00:00Z",
    "labels": {
      "a": "a"
    },
    "annotations": {
      "a": "a"
    },
    "clusterName": ""
  },
  "spec": {
    "permissions": [
      {
        "user_name": "",
        "resource_type": "",
        "resource": "",
        "role": ""
      }
    ]
  }
}' 
```



### 验证权限 - VerifyPermission

### 

```bash
kubectl-bcs-user-manager verify permissions  --help
```

参数详情:

```yaml 
-f, --form string   "json类型"

form  json说明

{
  "user_token": "",
  "resource_type": "",
  "resource": "",
  "action": ""
}
```

示例:

```
kubectl-bcs-user-manager verify permissions --form '{"user_token":"","resource_type":"","resource":"","action":""}'
```



### 验证权限V2 - VerifyPermissionV2

### 

```bash
kubectl-bcs-user-manager verify permissionsv2  --help
```

参数详情:

```yaml 
-f, --form string   "json类型"

form  json说明

{
  "user_token": "",
  "resource_type": "",
  "resource": "",
  "action": ""
}
```

示例:

```
kubectl-bcs-user-manager verify permissionsv2 --form '{"user_token":"","resource_type":"","resource":"","action":""}'
```







### 创建token - CreateToken

### 

```bash
kubectl-bcs-user-manager create token --help
```

参数详情:

```yaml 
-f, --token_form string   "json类型"

form  json说明

{
 "usertype":1,   //int类型 AdminUser=1 SaasUser=2 PlainUser=3 ClientUser=4
 "username":"", 
 "expiration":-1  //token expiration second, -1: never expire
}
```

示例:

```
kubectl-bcs-user-manager create token --token_form '{"usertype":1,"username":"", "expiration":-1}'
```



### 获取token - GetToken

### 

```bash
kubectl-bcs-user-manager get token --help
```

参数详情:

```yaml 
-n, --user_name string   "用户名称，通过该字段创建admin用户信息"
```

示例:

```
kubectl-bcs-user-manager get token -u [user_name to create]
kubectl-bcs-user-manager get t -u [user_name to create]
```



### 删除token - DeleteToken

### 

```bash
kubectl-bcs-user-manager delete token --help
```

参数详情:

```yaml 
-t, --token string   "token"
```

示例:

```
kubectl-bcs-manager delete token -t  [token]
```



### 更新token - UpdateToken

### 

```bash
kubectl-bcs-user-manager update token --help
```

参数详情:

```yaml 
-t, --token string   "token"
-f, --token_form string   "json类型"

token_form  json说明

{
 "expiration":-1  //token expiration second, -1: never expire
}


```

示例:

```
kubectl-bcs-manager update token --token [token] --form '{"expiration":-1}'
```



### 获取临时token - CreateTempToken

### 

```bash
kubectl-bcs-user-manager create temp-token --help
```

参数详情:

```yaml 
-f, --token_form string   "json类型"

token_form  json说明

{
  "usertype": 1,      //int类型 AdminUser=1 SaasUser=2 PlainUser=3 ClientUser=4
  "username": "",
  "expiration": -1   //int类型  token expiration second, -1: never expire
}
```

示例:

```
kubectl-bcs-user-manager create temp-token --token_form '{"usertype":1,"username":"", "expiration":-1}' 
```



### 创建客户端Token- CreateClientToken

### 

```bash
kubectl-bcs-user-manager create client-token --help
```

参数详情:

```yaml 
-f, --token_form string   "json类型"

token_form  json说明

{
  "clientName": "",    
  "clientSecret": "",
  "expiration": -1   //int类型  token expiration second, -1: never expire
}
```

示例:

```
kubectl-bcs-user-manager create client-token --token_form '{"clientName":"","clientSecret":"", "expiration":-1}'
```



### 根据user_name、cluster_id、business_id获取Token - GetTokenByUserAndClusterID

### 

```bash
kubectl-bcs-user-manager get extra-token --help
```

参数详情:



```yaml 
注意：三个参数需要同时传递

-n, --user_name string   "用户名称，通过该字段获取token信息"
    --cluster_id string   "集群id，通过该字段获取token信息"
    --business_id string   "业务id，通过该字段获取token信息"
```



示例:

```
kubectl-bcs-user-manager get extra-token -u [user_name] --cluster_id [cluster_id] --business_id [business_id]
```



### 新增TkeCidr - AddTkeCidr

### 

```bash
kubectl-bcs-user-manager create tkecidrs --help
```

参数详情:

```yaml 
-f, --tkecidr_form string   "json类型"

tkecidr_form  json说明
{
  "vpc": "",
  "tke_cidrs": [
    {
      "cidr": "",
      "ip_number": "",
      "status": ""   //"string ["available","used","reserved"]
    }
  ]
}
```

示例:

```
kubectl-bcs-user-manager create tkecidrs --tkecidr_form '{
  "vpc": "",
  "tke_cidrs": [
    {
      "cidr": "",
      "ip_number": "",
      "status": "available"
    }
  ]
}'
```



### 申请TkeCidr - ApplyTkeCidr

### 

```bash
kubectl-bcs-user-manager apply tkecidrs --help
```

参数详情:

```yaml 
-f, --tkecidr_form string   "json类型"

tkecidr_form  json说明
{
  "vpc": "",
  "cluster": "", 
  "ip_number": 1 //uint 正整数
}
```

示例:

```
kubectl-bcs-user-manager apply tkecidrs --tkecidr_form '{\"vpc\":\"\",\"cluster\":\"\", \"ip_number\":}' 
```



### 发布TkeCidr- ReleaseTkeCidr

### 

```bash
kubectl-bcs-user-manager release tkecidrs --help
```

参数详情:

```yaml 
-f, --tkecidr_form string   "json类型"

tkecidr_form  json说明
{
  "vpc": "",
  "cidr": "", 
  "cluster": "", 
}
```

示例:

```bash
kubectl-bcs-user-manager release tkecidrs --tkecidr_form '{"vpc":"","cidr":"","cluster":""}'
```



### 获取TkeCidr列表- ListTkeCidr

### 

```bash
kubectl-bcs-user-manager list tkecidrs --help
```

参数详情:

```yaml 
无参数
```

示例:

```
kubectl-bcs-user-manager list tkecidrs
```



### 同步Tke集群凭证- SyncTkeClusterCredentials

### 

```bash
kubectl-bcs-user-manager sync tkecidrs --help
```

参数详情:

```yaml 
-i, --cluster_id       string   "集群id，通过该字段同步tke集群redential信息"
```

示例:

```
kubectl-bcs-user-manager sync tkecidrs --cluster_id [cluster_id]
```







## 如何编译

执行下述命令编译 Client 工具
```
make bin
```