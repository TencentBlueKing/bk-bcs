# bcs webhook server Mesos部署

## mesos部署

### 1 插件配置初始化

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

### 2.2 bcs-webhook-server部署

#### 2.2.1 部署bcs-webhook-server

bcs-webhook-server以进程方式与bcs-mesos-driver部署在一起

config_file.json

```json
{
   "address": "0.0.0.0",
   "port": 18443,
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

启动命令：  

```shell
./bcs-webhook-server -f=config_file.json
```

#### 2.2.2 创建相关webhook

创建Application的webhook配置，webhook-conf-dep.json

```json
{
  "apiVersion": "bkbcs.tencent.com/v2",
  "kind":"AdmissionWebhookConfiguration",
  "metadata":{
    "name":"webhook-app",
    "namespace": "bcs-system"
  },
  "spec": {
    "apiVersion": "bkbcs.tencent.com/v2",
    "kind":"AdmissionWebhookConfiguration",
    "metadata":{
      "name":"webhook-app",
      "namespace": "bcs-system"
    },
    "resourcesRef": {
      "operation": "Create",
      "kind": "application"
    },
    "admissionWebhooks":[
      {
        "name": "bcs-webhook-server",
        "failurePolicy": "Fail",
        "clientConfig": {
          "caBundle": "xxxxxxxxxx",
          "url": "https://{{所在节点IP地址}}:18443/bcs/webhook/inject/v1/mesos"
        }
      }
    ]
  }
}
```

创建Deployment的webhook配置webhook-conf-dep.json

```json
{
  "apiVersion": "bkbcs.tencent.com/v2",
  "kind":"AdmissionWebhookConfiguration",
  "metadata":{
    "name":"webhook-dep",
    "namespace": "bcs-system"
  },
  "spec": {
    "apiVersion": "bkbcs.tencent.com/v2",
    "kind":"AdmissionWebhookConfiguration",
    "metadata":{
      "name":"webhook-dep",
      "namespace": "bcs-system"
    },
    "resourcesRef": {
      "operation": "Create",
      "kind": "deployment"
    },
    "admissionWebhooks":[
      {
        "name": "bcs-webhook-server",
        "failurePolicy": "Fail",
        "clientConfig": {
          "caBundle": "xxxxxxxxxx",
          "url": "https://{{bcs-webhook-server所在节点IP地址}}:18443/bcs/webhook/inject/v1/mesos"
        }
      }
    ]
  }
}
```
