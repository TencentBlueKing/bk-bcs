# bcs-log-webhook-server 原理和实现

## 1 背景
bcs-log-webhook-server 是 bcs 内部的一个用于向容器中注入日志配置信息的 webhook server 组件，详情请先阅读 [ bcs 日志方案](./log-controller.md)

## 2 实现原理
依赖于 k8s 的 Admission Webhook 机制以及 bcs mesos 的 Webhook 特性实现，bcs 在 k8s 和 mesos 两种容器编排引擎上都能支持容器的 webhook 注入，详情可参考：  
[kubernetes Admission Webhook 机制](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)  
[bcs mesos Webhook 特性](../bcs-mesos-driver/driver-implement.md)  

## 3 日志配置注入方案

### 3.1 日志配置 crd 
日志采集的信息通过 crd 的形式创建并存储在 k8s 或 mesos 集群中，需要注入到容器当中的日志采集信息主要包括 stdOut, logPath, dataId, appId, clusterId 。  
bcs-log-webhook-server 在拦截到容器的创建请求时，会从集群中查询 BcsLogConfig 类型的 crd ，如果查询到了匹配的 BcsLogConfig ，会把这个 BcsLogConfig 中的日志采集信息以环境变量的形式注入到容器当中。  
对于一个业务集群，在配置日志采集时，主要要考虑到三方面的情况：  

- 系统组件的日志采集配置  
系统组件一般都运行在 kube-system, kube-public, bcs-system 这几个命名空间下。对于这些命名空间下的系统组件，bcs-log-webhook-server 默认不会注入任何日志采集配置。  
如果要对系统组件的容器也进行日志采集，需要在这些系统组件的 pod 中加入指定的 annotation : sidecar.log.conf.bkbcs.tencent.com。当 pod 中这个指定的 annotation 的值为 "y", "yes", "true", "on" 时，bcs-log-webhook-server 才会对这些 pod 中的容器进行环境变量的注入。  
目前，针对系统组件的日志采集，不支持针对不同容器使用不同的采集配置，统一使用一个固定的配置。  
系统组件的日志采集配置需要创建一个 configType 类型为 bcs-system 的 BcsLogConfig。  


```
apiVersion: bkbcs.tencent.com/v2
kind: BcsLogConfig
metadata:
  name: standard-bcs-log-conf
  namespace: default
spec:
  configType: bcs-system
  appId: "10000"
  stdOut: false
  logPath: /data/logs/app.log
  clusterId: bcs-k8s-00001
  dataId: "10001"
```
  
- 标准的日志采集配置  
对于大多数业务集群，所有业务容器的日志输出是同样的格式，只需配置一种固定的日志清洗规则即可。对于这种需求，为了减少用户的负担，bk-bcs-saas 层可以默认创建一种 configType 类型为 standard 的 BcsLogConfig 资源，bcs-log-webhook-server 默认会对所有容器注入这个标准的日志采集配置：  
```
apiVersion: bkbcs.tencent.com/v2
kind: BcsLogConfig
metadata:
  name: standard-bcs-log-conf
  namespace: default
spec:
  configType: standard
  appId: "10000"
  stdOut: false
  logPath: /data/logs/app.log
  clusterId: bcs-k8s-00001
  dataId: "20001"
```

- 特殊的日志采集配置
如果一个业务集群中除了标准的日志采集配置外，还有某些容器需要配置特殊的日志采集规则，此时，可由用户在 bk-bcs-saas 层创建特定的日志采集配置 BcsLogConfig , 在 BcsLogConfig 中指定需要使用这种规则的容器名。创建完后，bcs-log-webhook-server 会对这些容器使用指定的 BcsLogConfig 配置来进行注入，以满足特定的需求：  
```
apiVersion: bkbcs.tencent.com/v2
kind: BcsLogConfig
metadata:
  name: proxy-bcs-log-conf
  namespace: default
spec:
  appId: "20000"
  stdOut: false
  logPath: /data/bcs
  clusterId: bcs-k8s-15049
  dataId: "20001"
  containers:
    - istio-proxy
    - proxy
```
例如，在创建这个名为 proxy-bcs-log-conf 的 BcsLogConfig 后，在集群中已经有 standard 类型的 BcsLogConfig 下，针对容器名为 istio-proxy 或 proxy 的容器，bcs-log-webhook-server 将使用这个名为 proxy-bcs-log-conf 的 BcsLogConfig 来注入采集信息。

### 3.2 webhook 配置

- k8s  
在 k8s 集群中创建 Kind 为 MutatingWebhookConfiguration 的 webhook 配置，添加对 pod 创建的 webhook 拦截，所有 pod 的创建请求都须先经过 bcs-log-webhook-server ，由 bcs-log-webhook-server 对容器注入日志配置信息。  
```
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: bcs-log-webhook-cfg
  labels:
    app: bcs-log-webhook
webhooks:
  - name: bcs-log-webhook.blueking.io
    clientConfig:
      # bcs-log-webhook-server 如果部署在 k8s 集群外，则配置 bcs-log-webhook-server 的调用 url
      url: https://x.x.x.x:443/bcs/log_inject/v1/k8s
      # bcs-log-webhook-server 如果以容器的形式部署在 k8s 集群中，则配置 service 来供 k8s 调用。
      service:
        name: bcs-log-webhook-svc
        namespace: default
        path: "/bcs/log_inject/v1/k8s
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
在 mesos 集群中创建 kind 为 admissionwebhook 的 webhook 配置，添加对 application 或 deployment 创建请求的拦截，由 bcs-log-webhook-server 对容器注入日志配置信息。  
```
{
  "apiVersion":"v4",
  "kind":"admissionwebhook",
  "metadata":{
    "name":"bcs-log-webhook",
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
        # bcs-log-webhook-server 如果部署在 mesos 集群外，则配置 bcs-log-webhook-server 的调用 url
        "url": "https://x.x.x.x:443/bcs/log_inject/v1/mesos",
        # bcs-log-webhook-server 如果部署在 mesos 集群内，则配置 service 名称和所在的 namespace
        "namespace": "bmsf-system",
        "name": "bmsf-mesos-injector"
      }
    }
  ]
}
```

## 4 部署
bcs-log-webhook-server 既可以容器的形式部署在 k8s 或 mesos 集群中，也可以进程的形式部署在集群外。如果是用于 k8s 集群的容器注入，则启动参数 engine_type 指定为 kubernetes , 如果用于 mesos 集群的容器注入，则启动参数 engine_type 指定为 mesos 。  

下面是进程部署方式：  

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
   "kubeconfig": "/data/home/kubeconfig"
}
```
启动命令：  
```
./bcs-log-webhook-server -f=config_file.json
```