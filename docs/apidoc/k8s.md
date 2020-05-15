# k8s api 使用说明

## 前置条件
请参考 [基于 k8s 的容器编排](../features/k8s/基于k8s的容器编排.md)，确保已完成 bcs k8s 集群的初始化。  
集群在 bcs 上完成初始化后，集群信息会被同步到 bcs-api-gateway ，通过 bcs-api-gateway 即能管理和调用集群。  

## 认证鉴权
对任意资源的访问都需要经过 bcs-user-manager 的认证和鉴权。  

管理员可以直接使用 bcs-user-manager 的 API，或者使用 bcs-client 命令来创建用户，创建成功后会返回为该用户签发的 usertoken 。  
```
./bcs-client create --usertype=plain --username=xxx --type=user
{
  "created_at": "2020-05-12T10:59:40.060551306+08:00",
  "expires_at": "2020-05-13T10:59:40.032198191+08:00",
  "id": 9,
  "name": "xxx",
  "updated_at": "2020-05-12T10:59:40.060551306+08:00",
  "user_token": "zYCZ962Ex5KW0L7D1WmdR0ijZkEkNIE6",
  "user_type": 3
}
```

管理员可以直接调用  bcs-user-manager 的 API ，或者使用 bcs-client 命令来给某用户授予某种资源的权限。  
目前 bcs-user-manager 支持对某一资源授予管理员和只读者的角色，其中管理员具有 GET\POST\PUT\DELETE\PATCH 等所有操作的权限，只读者只有 GET 的权限。  

授予 xxx 用户 BCS-K8S-100 集群管理 (manager) 的权限，为 yy 用户授权 BCS-K8S-101 集群 (viewer) 的权限。
```
# cat crd_permission.json
{
  "apiVersion":"v1",
  "kind":"permission",
  "metadata": {
     "name":"my-permission"
  },
  "spec":{
     "permissions":[
       {"user_name":"xxx", "resource_type":"cluster", "resource":"BCS-K8S-100", "role":"manager"},
       {"user_name":"yy", "resource_type":"cluster", "resource":"BCS-K8S-101", "role":"viewer"}
     ]
  }
}

# ./bcs-client grant --type=permission --from-file=crd_permission.json
success to grant permission
```


用户拿到管理员签发的 usertoken 后，并被授予了某个 k8s 集群的权限后，就能通过 bcs-api-gateway 提供的 API 来调用集群， bcs-api-gateway 会
调用 bcs-user-manager 完成认证鉴权工作，实现 k8s 的原生 API 访问。    

## k8s API

通过 k8s API 操作集群：  
```
curl -k -X GET -H "Authorization: Bearer zYCZ962Ex5KW0L7D1WmdR0ijZkEkNIE6" https://0.0.0.0:8443/tunnels/clusters/BCS-K8S-100/version
{
  "major": "1",
  "minor": "14+",
  "gitVersion": "v1.14.3-tk8s-v1.1-1",
  "gitCommit": "5f5efb53cb900bc0bccbe1bc9c57c7e3f55290db",
  "gitTreeState": "clean",
  "buildDate": "2019-08-27T06:46:41Z",
  "goVersion": "go1.12.9",
  "compiler": "gc",
  "platform": "linux/amd64"
}
```

通过 kubeconfig 操作集群：  
```
# cat kubeconfig
apiVersion: v1
kind: Config
clusters:
- cluster:
    api-version: v1
    insecure-skip-tls-verify: true
    server: https://0.0.0.0:8443/tunnels/clusters/BCS-K8S-100/
  name: BCS-K8S-100
contexts:
- context:
    cluster: BCS-K8S-100
    user: xxx
  name: BCS-K8S-100-xxx
current-context: BCS-K8S-100-xxx
users:
- name: xxx
  user:
    token: zYCZ962Ex5KW0L7D1WmdR0ijZkEkNIE6

# kubectl get node --kubeconfig=./kubeconfig
NAME      STATUS   ROLES    AGE    VERSION
master1   Ready    master   139d   v1.14.3-tk8s-v1.1-1
node1     Ready    node     137d   v1.14.3-tk8s-v1.1-1
```