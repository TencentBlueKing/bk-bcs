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



## 如何编译

执行下述命令编译 Client 工具
```
make bin
```