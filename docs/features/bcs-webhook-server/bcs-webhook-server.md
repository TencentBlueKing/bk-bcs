# bcs-webhook-server 原理和实现

## 1 背景
bcs-webhook-server 是 bcs 内部的一个用于向容器中注入相关配置信息的 webhook server 组件，可扩展，支持多种配置信息的注入，如 bcs 日志采集信息的注入，访问 db 授权的 init-container 的注入。  
bcs  日志采集信息的注入方案可阅读 [ bcs 日志方案](./log-controller.md)

## 2 实现原理
依赖于 k8s 的 Admission Webhook 机制以及 bcs mesos 的 Webhook 特性实现，bcs 在 k8s 和 mesos 两种容器编排引擎上都能支持容器的 webhook 注入，详情可参考：  
[kubernetes Admission Webhook 机制](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)  
[bcs mesos Webhook 特性](../bcs-mesos-driver/driver-implement.md)  

## 3 webhook 配置

- k8s  
在 k8s 集群中创建 Kind 为 MutatingWebhookConfiguration 的 webhook 配置，添加对 pod 创建的 webhook 拦截，所有 pod 的创建请求都须先经过 bcs-webhook-server ，由 bcs-webhook-server 对容器注入各种配置信息。  
```
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: bcs-webhook-cfg
  labels:
    app: bcs-webhook
webhooks:
  - name: bcs-webhook.blueking.io
    clientConfig:
      # bcs-webhook-server 如果部署在 k8s 集群外，则配置 bcs-webhook-server 的调用 url
      url: https://x.x.x.x:443/bcs/webhook/inject/v1/k8s
      # bcs-webhook-server 如果以容器的形式部署在 k8s 集群中，则配置 service 来供 k8s 调用。
      service:
        name: bcs-webhook-svc
        namespace: default
        path: "/bcs/webhook/inject/v1/k8s
      # CA 证书
      caBundle: 
    rules:
      - operations: [ "CREATE" ]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    failurePolicy: Fail
```   

- mesos
在 mesos 集群中创建 kind 为 admissionwebhook 的 webhook 配置，添加对 application 或 deployment 创建请求的拦截，由 bcs-webhook-server 对容器注入各种配置信息。  
```
{
  "apiVersion":"v4",
  "kind":"admissionwebhook",
  "metadata":{
    "name":"bcs-webhook-cfg",
    "namespace": "default"
  },
  "resourcesRef": {
    "operation": "Create",
    "kind": "application"
  },
  "admissionWebhooks":[
    {
      "name": "bcs-log-inject",
      "failurePolicy": "Fail",
      "clientConfig": {
        # CA 证书
        "caBundle": "",
        # bcs-webhook-server 如果部署在 mesos 集群外，则配置 bcs-webhook-server 的调用 url
        "url": "https://x.x.x.x:443/bcs/webhook/inject/v1/mesos",
        # bcs-webhook-server 如果部署在 mesos 集群内，则配置 service 名称和所在的 namespace
        "namespace": "bmsf-system",
        "name": "bmsf-mesos-injector"
      }
    }
  ]
}
```

## 4 配置注入方案

### 4.1 日志配置的注入

日志采集的信息通过 crd 的形式创建并存储在 k8s 或 mesos 集群中，需要注入到容器当中的日志采集信息主要包括 stdout, logPaths, dataId, appId, clusterId, namespace, logTags。
bcs-webhook-server 在拦截到容器的创建请求时，会从集群中查询 BcsLogConfig 类型的 crd ，如果查询到了匹配的 BcsLogConfig ，会把这个 BcsLogConfig 中的日志采集信息以环境变量的形式注入到容器当中。
对于一个业务集群，在配置日志采集时，主要要考虑到三方面的情况：

- 系统组件的日志采集配置
系统组件一般都运行在 kube-system, kube-public, bcs-system 这几个命名空间下。对于这些命名空间下的系统组件，bcs-webhook-server 默认不会注入任何日志采集配置。
如果要对系统组件的容器也进行日志采集，需要在这些系统组件的 pod 中加入指定的 annotation : webhook.inject.bkbcs.tencent.com。当 pod 中这个指定的 annotation 的值为 "y", "yes", "true", "on" 时，bcs-webhook-server 才会对这些 pod 中的容器进行环境变量的注入。
目前，针对系统组件的日志采集，不支持针对不同容器使用不同的采集配置，统一使用一个固定的配置。
系统组件的日志采集配置需要在应用所在的 namespace 下创建一个 configType 类型为 bcs-system 的 BcsLogConfig。  

```
apiVersion: bkbcs.tencent.com/v1
kind: BcsLogConfig
metadata:
  name: default-log-conf
  namespace: default
spec:
  configType: bcs-system
  appId: "20000"
  clusterId: bcs-k8s-15049
  stdDataId: "20001"
  nonStdDataId: "20002"
  stdout: true
  logPaths:
    - /data/home/logs
    - /data/logs/default
  logTags:
    platform: bcs
    app: kubelet
```
当 stdout 为 true 时，采集标准日志；当 logPaths 不为空时，采集非标准日志。  

- 默认的日志采集配置  
对于大多数业务集群，所有业务容器的日志输出是同样的格式，只需配置一种固定的日志清洗规则即可。对于这种需求，为了减少用户的负担，bk-bcs-saas 层可以在每个 namespace 下默认创建一个 configType 类型为 default 的 BcsLogConfig 资源，bcs-webhook-server 默认会对该 namespace 所有容器注入这个标准的日志采集配置：    
```
apiVersion: bkbcs.tencent.com/v1
kind: BcsLogConfig
metadata:
  name: default-log-conf
  namespace: default
spec:
  configType: default
  appId: "20000"
  clusterId: bcs-k8s-15049
  stdDataId: "20001"
  nonStdDataId: "20002"
  stdout: true
  logPaths:
    - /data/home/logs
    - /data/logs/default
  logTags:
    platform: bcs
```
当 stdout 为 true 时，采集标准日志；当 logPaths 不为空时，采集非标准日志。

- 自定义的日志采集配置
如果一个业务集群中除了标准的日志采集配置外，还有某些容器需要配置特殊的日志采集规则，此时，可由用户从 bk-bcs-saas 层在该 namespace 下创建特定的类型为 custom 的日志采集配置 BcsLogConfig , 在 BcsLogConfig 中指定需要使用这种规则的 workloads 类型(如 Deployment, Statefulset)、workloads 名、容器名。创建完后，bcs-webhook-server 会对这些容器使用指定的 BcsLogConfig 配置来进行注入，以满足特定的需求：    
```
apiVersion: bkbcs.tencent.com/v1
kind: BcsLogConfig
metadata:
  name: deploy-bcs-log-conf
  namespace: default
spec:
  configType: custom
  clusterId: bcs-k8s-15049
  appId: "20000"
  workloadType: Deployment
  workloadName: python-webhook
  containerConfs:
    - containerName: python
      stdDataId: "2001"
      nonStdDataId: "2002"
      stdout: true
      logPaths:
        - /data/home/logs1
        - /data/home/logs2
      logTags:
        app: python
        platform: bcs
    - containerName: sidecar
      stdDataId: "1001"
      nonStdDataId: "1002"
      stdout: false
      logPaths:
        - /var/log
        - /data/home/logs
      logTags:
        app: sidecar
        platform: bcs
```
例如，在创建这个名为 deploy-bcs-log-conf 的 BcsLogConfig 后，在集群的 default 命名空间中已经有 default 类型的 BcsLogConfig 的情况下，
针对 default 命名空间下名为 python-webhook 的 deployment，bcs-webhook-server 为匹配到 deploy-bcs-log-conf ，对其 pod 中容器名为 python 的容器会使用 containerConfs 中第一组配置来注入，
名为 sidecar 的容器会使用 containerConfs 中第二组配置来注入。如果该 pod 中还有其它名字的容器，因为没有匹配到 containerConfs 信息，
仍然会使用 default 类型的 BcsLogConfig 来注入。  
**注意： 每个 workload 只会对一个自定义的 crd 生效**

### 4.2 db 授权 init-container 的注入
db 授权的配置信息以 crd 的形式创建并下发到 k8s 或 mesos 集群中, 包括 podSelector, appName, targetDb, dbType, callUser, dbName.  
- podSelector: 用于匹配包含有相同 label 的 pod 。 匹配的 pod 将会使用此 crd 中的信息来注入授权的 init-container.  
- appName: db 所属的业务名，须与 db 授权模板中的业务名相同。  
- targerDb: db 的域名。  
- dbType: db 的类型，mysql 或 spider. 
- callUser: db 授权模板中匹配的调用 api 的用户名。  
- dbName: database 名，支持模糊匹配，须与 db 授权模板中配置的数据库名称相同  

示例的 crd 配置如下：  
```
apiVersion: bkbcs.tencent.com/v1
kind: BcsDbPrivConfig
metadata:
  name: bcs-db-privilege
  namespace: default
spec:
  podSelector:
    app: db-privelege
  appName: xxxx
  targetDb: xxxx
  dbType: spider
  callUser: bryanhe
  dbName: db%
```

## 5 部署
### 5.1 bcs-webhook-server 部署
bcs-webhook-server 既可以容器的形式部署在 k8s 或 mesos 集群中，也可以进程的形式部署在集群外。如果是用于 k8s 集群的容器注入，则启动参数 engine_type 指定为 kubernetes , 如果用于 mesos 集群的容器注入，则启动参数 engine_type 指定为 mesos 。  
bcs-webhook-server 是可扩展的，支持多种方案的注入，目前支持日志采集信息和 db 授权 init-container 的注入。可以在启动配置中设置是否要开启各个方案的注入。  

下面是示例的启动配置：  

config_file.json
```json
{
   "address": "0.0.0.0",
   "port": 443,
   "log_dir": "./logs",
   "log_max_size": 500,
   "log_max_num": 10,
   "logtostderr": true,
   "alsologtostderr": true,
   "v": 5,
   "stderrthreshold": "2",
   "server_cert_file": "/data/home/cert.pem",
   "server_key_file": "/data/home/key.pem",
   "engine_type": "kubernetes",
   "kubeconfig": "/data/home/kubeconfig",
   "injects": {
         "log_conf": true,   // 是否开启日志采集的注入
         "db_privilege": {
            "db_privilege_inject": true,    // 是否开启db授权的注入
            "network_type": "overlay",      // 该集群的网络方案，overlay或underlay
            "esb_url": "http://x.x.x.x:8080",   // 调用 db 授权的 ESB 接口地址
            "init_container_image": "db-privilege:test"  // 用于授权的 init-container 的镜像名
         }
      }
}
```
启动命令：  
```
./bcs-webhook-server -f=config_file.json
```

### 5.2 ESB secret 部署
因为调用 ESB 接口时，需要提供 app_code 和 app_secret 信息，因此在启用 db 授权注入时，部署 bcs-webhook-server 前，还需要在 k8s 集群中部署包含 app_secret 、app_code 和 operator 的 secret：  
```
apiVersion: v1
kind: Secret
metadata:
  name: bcs-db-privilege
  namespace: kube-system
type: Opaque
data:
  # go-esb-sdk
  sdk-appCode: ""
  sdk-appSecret: ""
  sdk-operator: ""
```

bcs-webhook-server 会从集群中的这个 secret 中去获取这3个数据。