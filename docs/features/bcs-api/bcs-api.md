# bcs-api 
bcs 支持 kubernetes 和 mesos 两种容器服务编排系统，bcs-api 是 bcs 系统对外提供服务的 api 接入层，bcs-api 纳管多个 kubernetes 和
mesos 集群，外部服务通过调用 bcs-api 提供的 api 来访问不同的 kubernetes 或 mesos 集群，再由 bcs-api 把 api 调用下发到相应的集群。

## bcs-api 架构
融合 kubernetes 和 mesos 的 bcs-api 的架构图如下：  

![image.png](./image/bcs-api.png)

### kubernetes 

#### 服务发现  
bcs-api 使用 bcs-kube-agent 上报的方式来实现 kubernetes 集群的服务发现。  
bcs-kube-agent 以 deployment 的形式运行在每个 kubernetes 集群中，它周期性地收集集群的 master 地址以及证书 token 等信息，
通过调用 bcs-api 的接口上报给 bcs-api。bcs-api 把集群的地址及证书信息存入 db 中。  
bcs-kube-agent 在上报 kubernetes 集群信息时，首先需要知道该集群的 clusterId 和 用于验证合法性的 register_token ，因此需要先在 bcs-api 
上创建集群及其 register_token 。

#### api 类别
bcs-api 可以看作成 kubernetes 集群上的一个反向代理，它包含两种 api ：

- rest-api
rest-api 是用于 kubernetes 集群管理的 api ，如在 bcs-api 上创建集群，创建集群的 register_token，提供给 bcs-kube-agent 
调用以注册集群的 api 等等。

- kubernetes api
kubernetes 的反向代理 api ，外部服务通过调用 bcs-api 的这些 api ，来调用相应的各个 kubernetes 集群的 api 。

#### 认证鉴权
bcs-api 打通了 kubernetes 的 rbac 体系，bcs-api 会为用户生成 userToken。  
在调用 bcs-api 的 kubernetes api 时，需要传递 userToken， 
bcs-api 验证 userToken 的合法性后，如果 userToken 对应的用户非超级用户，会在下发调用具体的 kubernetes 时，在 request header 中加入用户名。
因此，使用普通用户的 userToken 调用 bcs-api 的 kubernetes api 时，首先需要在 kubernetes 集群中已经为该用户创建了相应的 rbac 配置，
如果没有创建任何的 rbac 配置，那么就无权调用 kubernetes 集群的 api。  
bcs-api 实现了权限订阅的功能，如果用户有自己的 auth 模块，可以与 bcs-api 的 rbac 订阅模块对接，bcs-api 会通过订阅获取到权限数据后，
解析成 kubernetes 的 rbac 并下发到 kubernetes 集群当中。  
使用 admin 用户的 userToken 调用时，bcs-api 会直接以 cluster-admin 的角色调用 kubernetes 集群，拥有 kubernetes 集群的所有权限。  

### mesos

#### 服务发现
bcs-api 使用 mesos-driver 上报给 zookeeper 的方式实现 mesos 集群的服务发现。mesos-driver 在每个 mesos 集群中部署3台，
它会上报自己的地址及集群 id 到 zookeeper。在使用 bcs-api 调用 mesos 的 api 时，bcs-api 在 zookeeper 中随机选择一台对应的 mesos-driver, 
下发调用到 mesos-driver，最后由 mesos-driver 转发 api 调用给 mesos scheduler 。

#### 其它
具体的 mesos 信息可参考文档 [mesos 文档](../mesos)

### bcs-api 部署

#### 配置文件
bcs-api.yaml
```
edition: ee
service_config:
  address: 0.0.0.0
  port: 8443
  insecure_address: 0.0.0.0
  insecure_port: 8080
metric_config:
  metric_port: 8087
zk_config:
  bcs_zookeeper: 
cert_config:
  ca_file: ""
  server_cert_file: ""
  server_key_file: ""
  client_cert_file: ""
  client_key_file: ""
license_server_config:
  ls_address:
  ls_ca_file:
  ls_client_cert_file:
  ls_client_key_file:
log_config:
  log_dir: ./logs
  log_max_size: 500
  log_max_num: 10
  logtostderr: true
  alsologtostderr: true
  v: 0
  stderrthreshold: "2"
  vmodule:
  log_backtrace_at:
process_config:
  pid_dir:
local_config:
  local_ip: ""

# auth 订阅
bkiam_auth:
  auth:
  apigw_rsa_file:
  auth_token_sync_time:
  bkiam_auth_host: 
  bkiam_auth_app_code: 
  bkiam_auth_app_secret: 
  bkiam_auth_system_id:
  bkiam_auth_scope_id:
  bkiam_auth_zookeeper:
  bkiam_auth_sub_server:
  bkiam_auth_sub_server:
  bkiam_auth_token_whitelist:
    - 

core_database:
  dsn: "root:xx@tcp(x.x.x.x:3306)/bke_core?charset=utf8mb4&parseTime=True&loc=Local"

# bootstrap_users defines some admin users, it can be useful for bootstraping a new bke environment where no
# users exists in database
bootstrap_users:
  - name: "admin"
    is_super_user: true
    tokens:
      - ""

# rbac data
rbac:
  # 是否开启rbac
  turn_on_rbac: false
  # 是否从paas-auth订阅 rbac 数据，默认true
  turn_on_auth: false
  # 是否从配置文件读取 rbac 数据，默认false
  turn_on_conf: false

# mesos webconsole proxy 监听的端口
consoleproxy_port:
  8080
```

#### 启动服务
./bcs-api --config=bcs-api.yaml  

### bcs-api kubernetes 服务使用说明

#### 创建 user 及其 user_token
bcs-api 在启动时在配置文件中指定了一个 admin user 及其 user_token。 使用该 user_token 可以创建普通 user 及其 user_token。  

创建user:  
```
curl  -X POST -H "Content-Type: application/json" --cacert ./bcs-inner-ca.crt -H  "Authorization: Bearer {admin user_token}" https://bcs_server:8443/rest/users/ -d '{"user_name": "xxxx"}'
```
获取user_id:
```
curl  -X GET -H "Content-Type: application/json" --cacert ./bcs-inner-ca.crt -H  "Authorization: Bearer {admin user_token}" https://bcs_server:8443/rest/users/xxxx
```

创建user_token：
```
curl  -X POST -H "Content-Type: application/json" --cacert ./bcs-inner-ca.crt -H  "Authorization: Bearer {admin user_token}" https://bcs_server:8443/rest/users/{user_id}/tokens
```

#### 创建集群
首先，我们需要使用 user_token 作为作证调用 `POST /rest/clusters/bcs` 接口来创建一个集群。调用该接口需要提供一个集群 ID 字段。创建完成后，服务端将会返回新创建的集群信息。

#### 创建集群 RegisterToken
为了通过反向代理访问 k8s 集群的 apiserver，我们需要将集群鉴权信息注册到 bcs-api 中。 这个注册过程需要由 RegisterToken 来保证安全。

访问 `POST /rest/clusters/{{ cluster_id  }}/register_tokens` 可以为集群创建新的 RegisterToken。

一个集群只能有一个 RegisterToken，在创建完 token 后，你可以通过访问 `GET /rest/clusters/{{ cluster_id  }}/register_tokens` 来查看 token 内容。

#### 注册集群鉴权信息
kubernetes 集群地址及鉴权信息由 bcs-kube-agent 上报给 bcs-api。部署好 bcs-kube-agent 后，bcs-kube-agent 会调用 bcs-api 的 `PUT /clusters/{cluster_id}/credentials` 自动上报到 bcs-api 

#### 获取集群 id 和 证书信息
访问 `GET /clusters/{cluster_id}/client_credentials` 获取集群信息：  
```
{
  cluster_id:
  server_address:
  user_token:
  cacert_data:
}
```
server_address 为通过 bcs-api 来调用 kubernetes 集群的地址。

#### 调用 kubernetes api
```
curl -H  "Authorization: Bearer {user_token}" {server_address}/version
```
注： 在开启了 rbac 的情况下(配置文件中 turn_on_rbac 为 true)，如果使用普通用户的 user_token 调用 kubernetes api，是没有权限的，需要先在 kubernetes 集群中为该用户创建 rbac。使用 admin 用户的 user_token 具有所有权限。

### bcs-kube-agent 部署
bcs-kube-agent 以 deployment 方式部署在 kubernetes 集群当中。

- Dockerfile
```
FROM centos:latest

ADD ./bcs-kube-agent /bcs-kube-agent

ENTRYPOINT ["/bcs-kube-agent"]
```
在 bcs-api 中创建了集群和 register_token 后，使用以下的 yaml 文件在 kubernetes 中部署 bcs-kube-agent 。
- install yaml 
```
apiVersion: v1
kind: Secret
metadata:
  name: bke-info
  namespace: kube-system
type: Opaque
data:
  # register_token
  token: 
  # ca 证书
  bke-cert:
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: bcs-kube-agent
  namespace: kube-system
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: bcs-kube-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: bcs-kube-agent
  namespace: kube-system
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: bcs-kube-agent
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bcs-kube-agent
  template:
    metadata:
      labels:
        app: bcs-kube-agent
    spec:
      containers:
      - name: bcs-kube-agent
        image: 
        imagePullPolicy: IfNotPresent
        args:
        # bcs-api 地址
        - --bke-address=
        # 集群 id
        - --cluster-id=
        env:
          - name: REGISTER_TOKEN
            valueFrom:
              secretKeyRef:
                name: bke-info
                key: token
          - name: SERVER_CERT
            valueFrom:
              secretKeyRef:
                name: bke-info
                key: bke-cert
      serviceAccountName: bcs-kube-agent

```