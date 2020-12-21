[TOC]

# bcs-webhook-server 部署

## 0. bcs-webhook-server参数说明

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
   "alsologtostderr": true,
   "v": 5,
   "server_cert_file": "/data/home/cert.pem",
   "server_key_file": "/data/home/key.pem",
   "engine_type": "kubernetes",
   "plugin_dir": "./plugins",
   "plugins": "dbpriv,bscp"
}
```

参数解析

* **address**: bcs-webhook-server监听的地址
* **port**: bcs-webhook-server监听的端口
* **log_dir**: 日志存放文件夹
* **log_max_size**: 单个日志文件大小，单位MB
* **log_max_num**: 最大的日志数量
* **alsologtostderr**: 是否同时将日志输出到标准错误
* **v**: 日志级别
* **server_cert_file**: bcs-webhook-server server端证书
* **server_key_file**: bcs-webhook-server server端密钥
* **engine_type**: hook引擎类型，kubernetes或者mesos
* **plugin_dir**: 插件配置所在文件夹
* **plugins**: 需要激活的插件名字，hook插件将按次顺序调用

启动命令：  

```shell
./bcs-webhook-server -f=config_file.json
```

## 1. K8S部署

### 1.1 插件配置初始化

bcs-webhook-server插件配置文件以“插件名字.conf”的形式集中存储在bcs-webhook-server启动参数指定的文件夹下

#### 1.1.1 DB授权

dbpriv.conf

```json
{
    "kube_master": "{{ k8s apiserver地址 }}",
    "kubeconfig": "{{ kubeconfig路径 }}",
    "network_type": "{{ 网络类型，可选[overlay, underlay] }}",
    "esb_url": "{{ esb接口地址 }}",
    "init_container_image": "{{ init容器镜像地址 }}"
}
```

因为调用 ESB 接口时，需要提供 app_code 和 app_secret 信息，因此在启用 db 授权注入时，部署 bcs-webhook-server 前，还需要在 k8s 集群中部署包含 app_secret 、app_code 和 operator 的 secret。

```yaml
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

dbprivilege注入的init容器会从集群中的这个 secret 中去获取这3个数据。

#### 1.1.2 BSCP自动注入

bscp.conf

```json
[
  {
    "name": "bscp-side-car",
    "image": "{{ bscp side car镜像地址 }}",
    "imagePullPolicy": "IfNotPresent",
    "env":[
      {
        "name": "BSCP_BCSSIDECAR_CONNSERVER_HOSTNAME",
        "value": "{{ bscp connection server地址 }}"
      },
      {
        "name": "BSCP_BCSSIDECAR_CONNSERVER_PORT",
        "value": "{{ bscp connection server端口 }}"
      },
      {
        "name": "BSCP_BCSSIDECAR_APPINFO_IP_ETH",
        "value": "{{ bscp sidecar网卡名字 }}"
      }
    ]
  }
]
```

#### 将插件配置文件写成configmap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: bcs-webhook-server-plugin-confs
  namespace: kube-system
data:
  bscp.conf: '[
  {
    "name": "bscp-side-car",
    "image": "{{ bscp side car镜像地址 }}",
    "imagePullPolicy": "IfNotPresent",
    "env":[
      {
        "name": "BSCP_BCSSIDECAR_CONNSERVER_HOSTNAME",
        "value": "{{ bscp connection server地址 }}"
      },
      {
        "name": "BSCP_BCSSIDECAR_CONNSERVER_PORT",
        "value": "{{ bscp connection server端口 }}"
      },
      {
        "name": "BSCP_BCSSIDECAR_APPINFO_IP_ETH",
        "value": "{{ bscp sidecar网卡名字 }}"
      }
    ]
  }
]'
  dbpriv.conf: '{
    "kube_master": "{{ k8s apiserver地址 }}",
    "kubeconfig": "{{ kubeconfig路径 }}",
    "network_type": "{{ 网络类型，可选[overlay, underlay] }}",
    "esb_url": "{{ esb接口地址 }}",
    "init_container_image": "{{ init容器镜像地址 }}"
}'
```

### 1.2 bcs-webhook-server部署

#### 1.2.1 bcs-webhook-server deployment 配置

创建ClusterRole

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: bcs-webhook-server
rules:
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - get
  - list
  - watch
  - update
  - create
- apiGroups:
  - bkbcs.tencent.com
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
```

创建ClusterRoleBinding

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: bcs-webhook-server
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: bcs-webhook-server
subjects:
- kind: ServiceAccount
  name: bcs-webhook-server
  namespace: kube-system
```

创建bcs-webhook-server所需证书的secret

```yaml
apiVersion: v1
data:
  cert.pem: 'xxxxxxxxx'
  key.pem: 'xxxxxxxxx'
kind: Secret
metadata:
  name: bcs-webhook-server-certs
  namespace: kube-system
type: Opaque
```

创建bcs-webhook-server service

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-webhook-server
  name: bcs-webhook-server
  namespace: kube-system
spec:
  ports:
  - port: 443
    protocol: TCP
    targetPort: 443
  selector:
    app: bcs-webhook-server
  type: ClusterIP
```

创建bcs-webhook-server deployment

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    app: bcs-webhook-server
  name: bcs-webhook-server
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bcs-webhook-server
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: bcs-webhook-server
    spec:
      containers:
      - args:
        - --address=0.0.0.0
        - --port=443
        - --log_dir=./logs
        - --log_max_size=500
        - --log_max_num=10
        - --logtostderr=true
        - --alsologtostderr=true
        - --v=3
        - --stderrthreshold=2
        - --server_cert_file=/data/bcs/cert/cert.pem
        - --server_key_file=/data/bcs/cert/key.pem
        - --engine_type=kubernetes
        - --plugin_dir=/data/bcs/plugins
        - --plugins
        - dbpriv,bscp
        # bcs-webhook-server镜像地址
        image: bcs-webhook-server:1.3.0
        imagePullPolicy: IfNotPresent
        name: bcs-webhook-server
        volumeMounts:
        - mountPath: /data/bcs/cert
          name: webhook-certs
          readOnly: true
        - mountPath: /data/bcs/plugins
          name: plugin-confs
      dnsPolicy: ClusterFirst
      nodeSelector:
        node-role.kubernetes.io/master: "true"
      restartPolicy: Always
      serviceAccount: bcs-webhook-server
      serviceAccountName: bcs-webhook-server
      terminationGracePeriodSeconds: 30
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
        operator: Exists
      volumes:
      - name: webhook-certs
        secret:
          defaultMode: 420
          secretName: bcs-webhook-server-certs
      - name: plugin-confs
        configMap:
          name: bcs-webhook-server-plugin-confs
          items:
          - key: "bscp.conf"
            path: "bscp.conf"
          - key: "dbpriv.conf"
            path: "dbpriv.conf"
```

#### 1.2.2 注册所需要的AdmissionWebhook

```yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  labels:
    app: bcs-webhook-server
  # 采用zz名字开头，努力保证bcs-webhook-server是最后一个被调用的。
  name: zz-bcs-webhook-server-cfg
webhooks:
- admissionReviewVersions:
  - v1beta1
  clientConfig:
    # bcs-webhook-server服务的CA证书
    caBundle: 'xxxxxxxxxxxxxxxxxxxxxx'
    service:
      name: bcs-webhook-server
      namespace: kube-system
      path: /bcs/webhook/inject/v1/k8s
  failurePolicy: Fail
  name: bcs-webhook-server.blueking.io
  namespaceSelector:
    matchExpressions:
    - key: bcs-webhook
      operator: NotIn
      values:
      - "false"
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - pods
    scope: '*'
  sideEffects: Unknown
  timeoutSeconds: 30
```
