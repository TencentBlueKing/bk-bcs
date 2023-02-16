

# bcs-monitor 本地开发10分钟启动教程(helm) / 文档



## 机器环境依赖

```
go               1.17
Docker           20.10.17
k8s(minikube)    1.23.5
git              2.38.1      
helm             3.9.4
```



## 操作步骤

### 1.下载源代码

git clone https://github.com/Tencent/bk-bcs.git

### 2.切换目录，进入到bcs-monitor目录

```
[root@control-plane helm]# cd /data/workspace/src/deploy/bcs-monitor/bcs-monitor/
[root@control-plane bcs-monitor]# ll
total 372
drwxr-xr-x.  2 root root     25 Feb 13 11:48 bin
-rw-r--r--.  1 root root      0 Feb 13 11:37 CHANGELOG.md
drwxr-xr-x.  3 root root     25 Feb 13 11:37 cmd
-rw-r--r--.  1 root root    148 Feb 13 11:37 Dockerfile
drwxr-xr-x.  2 root root     41 Feb 13 11:37 docs
drwxr-xr-x.  2 root root     78 Feb 13 11:37 etc
-rw-r--r--.  1 root root  11602 Feb 13 11:37 go.mod
-rw-r--r--.  1 root root 342884 Feb 13 11:37 go.sum
-rw-r--r--.  1 root root   1088 Feb 13 11:37 LICENSE
-rw-r--r--.  1 root root   2141 Feb 13 11:37 Makefile
drwxr-xr-x. 13 root root    163 Feb 13 11:37 pkg
-rw-r--r--.  1 root root     16 Feb 13 11:37 README.md
-rw-r--r--.  1 root root   1141 Feb 13 11:37 run-test.sh
drwxr-xr-x.  4 root root     54 Feb 13 11:37 test
-rw-r--r--.  1 root root      5 Feb 13 11:37 VERSION

```

### 3.构建docker镜像

```
注意： dockerfile中复制bin目录下的可行性文件，所以需要先make build，后make docker
make build
make docker 
```

### 4.修改values.yaml，按照自己的配置执行helm chat install 安装bcs-monitor

```
helm install --generate-name  --debug  ./bcs-monitor   -n bcs-monitor
```

## 要点

### 1.关于storegw

#### config.storegw.type 与config.storegw.config.url

storegw 对接 prometheus，所以确定自己value.yaml中声明configmap的内容中config.storegw.type值为PROMETHEUS,并在config.storegw.config.url设置

prometheus的访问url

```

config
  storegw:
    #- type: BCS_SYSTEM
    - type: PROMETHEUS
      config: 
      	url: "http://192.168.37.133:30090"

```

```
[root@control-plane bcs-monitor]# k get svc -n monitoring
NAME                                           TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
alertmanager-operated                          ClusterIP   None             <none>        9093/TCP,9094/TCP,9094/UDP   5h1m
prometheus-operated                            ClusterIP   None             <none>        9090/TCP                     5h1m
prometheus-operator-grafana                    ClusterIP   10.107.72.98     <none>        80/TCP                       5h1m
prometheus-operator-kube-p-alertmanager        ClusterIP   10.96.18.143     <none>        9093/TCP                     5h1m
prometheus-operator-kube-p-operator            ClusterIP   10.98.193.108    <none>        443/TCP                      5h1m
prometheus-operator-kube-p-prometheus          NodePort    10.107.100.32    <none>        9090:30090/TCP               5h1m
prometheus-operator-kube-state-metrics         ClusterIP   10.111.36.218    <none>        8080/TCP                     5h1m
prometheus-operator-prometheus-node-exporter   ClusterIP   10.102.102.230   <none>        9100/TCP                     5h1m

```
<font color='red'> 
如上prometheus servie nodeport为30090
</font>





### 2.关于权限

value.yaml中默认dev: false不开启权限校验,该字段控制deployment-xxx.yaml中关于权限校验的configmap挂载

```
credentials:
  enabled: false
```

Error: Invalid Key: Key must be PEM encoded PKCS1 or PKCS8 private key如果容器中出现这个错误，修改配置文件中关于configmap的配置，确保最后挂载到容器中的配置文件不包含以下三种权限认证的配置

```
config:
  auth_conf:
    host: ""
    is_gateway: ""
    ssm_host: ""
  bkapigw_conf:
    host: ""
    jwt_public_key: ""

  bcs_conf:
    verify: false

 

```

同时需要进入本地开发模式

values.global.env.BCS_MONITOR_USERNAME:  "test_dev"   ##（不为空的字符串）

values.config.run_env: dev

## 补充

### 1.prometheus安装

storegw对接 prometheus，所以需要提供prometheus



安装方式1.单机版

github搜索prometheus release包自行解压安装



安装方式2.prometheus-operator（推荐）

```
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
kubectl create ns monitoring
helm install prometheus-operator prometheus-community/kube-prometheus-stack --prometheus.service.type=NodePort --namespace monitoring 
```



有个别镜像需要翻墙不好下载，如下

```
docker pull paymanfu/ingress-nginx_kube-webhook-certgen:v1.3.0
docker tag  paymanfu/ingress-nginx_kube-webhook-certgen:v1.3.0  registry.k8s.io/ingress-nginx/kube-webhook-certgen:v1.3.0



docker pull k8scopy/kube-state-metrics:v2.7.0
docker tag  k8scopy/kube-state-metrics:v2.7.0 registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.7.0
```







## 部署效果图

![image-20230216135432844](E:\goworkspace\src\bk-bcs\bcs-services\bcs-monitor\docs\helm_deploy.assets\image-20230216135432844.png)



![image-20230216135517887](E:\goworkspace\src\bk-bcs\bcs-services\bcs-monitor\docs\helm_deploy.assets\image-20230216135517887.png)

![image-20230216135455203](E:\goworkspace\src\bk-bcs\bcs-services\bcs-monitor\docs\helm_deploy.assets\image-20230216135455203.png)

![image-20230216135556366](E:\goworkspace\src\bk-bcs\bcs-services\bcs-monitor\docs\helm_deploy.assets\image-20230216135556366.png)