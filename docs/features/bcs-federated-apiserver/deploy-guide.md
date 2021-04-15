> bcs-federated-apiserver 采用helm方式进行部署和升级

federated-apiserver 是 在联邦集群的 host 集群中运行的聚合apiserver，支持 member 集群的Pod资源聚合查询功能。
重要参数说明：

### Value.configmap部分:
* bcsStorageAddress: "http://x.xxx.xx.xxx:yyyy" , 联邦集群资源数据对接的 bcs-storage 地址；不能为空；
* bcsStoragePodUri: "/xxxxxxxxxx/xx/xxxxxxx/xxxxxxxxxxxxxxx/xxx" ， 联邦集群资源数据对接的 bcs-storage 请求路径；不能为空；
* bcsStorageToken: "xxxxxyyyyy" 联邦集群资源数据对接的 bcs-storage apigatewya token (base64 encoded: echo -n "xxxx" | base64)；如果为空，代表不启用；
* memberClusterIgnorePrefix: "xxxxxxx"，从kubefed中获取member集群时，屏蔽集群名中的 "member." 字段；如果为空，代表不屏蔽从字段；
* memberClusterOverride: "xxxxxxxxxxxxx"，指定member集群名称为"xxxxxxxxxxxxx"，以覆盖默认host集群中注册的member集群来提供查询Pod 等资源的集群范围；如果为空，代表从 kube-federation-system 的namespace下的 kubefedclusters 资源中获取；

### Value.secret部分：
* apiserver与kube-apiserver之间的认证信息，因在配置时默认关闭认证，该部分可填写统一的值（聚合apiserver框架暂不支持关闭认证，因此需必填一套认证信息）

### 部署与升级
* 修改value或Chart部分
* 部署或升级
```shell
helm upgrade bcs-federated-apiserver ./bcs-federated-apiserver/ -n bcs-system --install
```

* 查看部署情况
```shell
helm list -n bcs-system
```

* 卸载
```shell
helm uninstall bcs-federated-apiserver -n bcs-system
```
