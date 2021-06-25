# bcs-gamestatefulset/deployment/hook-operator服务metrics指标

## controller指标
operator组件的metrics指标主要分为两类

* workqueue队列性能指标
* k8s-api的性能指标(请求量、延迟、成功率)

### workqueue性能指标
#### bkbcs\_workqueue_depth
* 观测队列长度metric指标
* label：name

#### bkbcs\_workqueue\_adds\_total
* 队列添加对象数目
* label：name

#### bkbcs\_workqueue\_queue\_duration\_seconds
* 队列元素添加至队列到获取元素耗时统计
* label：name

#### bkbcs\_workqueue\_work\_duration\_seconds
* 队列获取元素至处理结束的耗时统计
* label：name

#### bkbcs\_workqueue\_unfinished\_work\_seconds
* 统计controller未完成task运行的总耗时
* label：name

#### bkbcs\_workqueue\_longest\_running\_processor\_seconds
* 统计controller运行task的最长耗时
* label：name

### k8s-api性能指标
#### bkbcs\_rest\_client\_requests\_total
* rest_api请求总数统计
* label: code, method, host

#### bkbcs\_rest\_client\_request\_duration\_seconds
* rest_api请求延迟
* label: verb, url

#### bkbcs\_rest\_client\_rate\_limiter\_duration\_seconds
* api限速延迟耗时
* label: verb，url

## 指标聚合
### workqueue队列指标聚合
```
WorkQueue Add Rate 队列元素增长速率
sum(rate(bkbcs_workqueue_adds_total{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance"}[5m])) by (instance, name)

WorkQueue Depth 队列实时长度增长速率
sum(rate(bkbcs_workqueue_depth{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance"}[5m])) by (instance, name)

WorkQueue Depth 队列长度
bkbcs_workqueue_depth{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance"} by (instance, name)

Work Queue Latency 队列添加元素到处理耗时
histogram_quantile(0.99, sum(rate(bkbcs_workqueue_queue_duration_seconds_bucket{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance"}[5m])) by (instance, name, le))

Work Queue Latency 队列处理元素到处理完成耗时
histogram_quantile(0.99, sum(rate(bkbcs_workqueue_work_duration_seconds_bucket{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance"}[5m])) by (instance, name, le))

```   

### rest API指标聚合
```
k8s restAPI请求增长速率
sum(rate(bkbcs_rest_client_requests_total{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance",code=~"2.."}[5m]))

sum(rate(bkbcs_rest_client_requests_total{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance",code=~"3.."}[5m]))

sum(rate(bkbcs_rest_client_requests_total{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance",code=~"4.."}[5m]))

sum(rate(bkbcs_rest_client_requests_total{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance",code=~"5.."}[5m]))


histogram_quantile(0.99, sum(rate(bkbcs_rest_client_request_duration_seconds_bucket{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance", verb="POST"}[5m])) by (verb, url, le))

histogram_quantile(0.99, sum(rate(bkbcs_rest_client_request_duration_seconds_bucket{cluster_id=~"$cluster_id",job="kube-controller-manager", instance=~"$instance", verb="GET"}[5m])) by (verb, url, le))


```

## 创建监控对象
### bcs-gamestatefulset-operator
####service
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-gamestatefulset-operator
    release: po
  name: bcs-gamestatefulset-operator
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app.kubernetes.io/platform: bk-bcs
    app.kubernetes.io/name: bcs-gamestatefulset-operator
  sessionAffinity: None
  type: NodePort
```
#### servicemonitor

```
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-gamestatefulset-operator
    release: po
  name: bcs-gamestatefulset-operator
  namespace: bcs-system
spec:
  endpoints:
  - interval: 30s
    params: {}
    path: /metrics
    port: http
  sampleLimit: 100000
  selector:
    matchLabels:
      app: bcs-gamestatefulset-operator
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
### bcs-gamedeployment-operator
####service
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-gamedeployment-operator
    release: po
  name: bcs-gamedeployment-operator
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app.kubernetes.io/platform: bk-bcs
    app.kubernetes.io/name: bcs-gamedeployment-operator
  sessionAffinity: None
  type: NodePort
```
#### servicemonitor

```
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-gamestatefulset-operator
    release: po
  name: bcs-gamestatefulset-operator
  namespace: bcs-system
spec:
  endpoints:
  - interval: 30s
    params: {}
    path: /metrics
    port: http
  sampleLimit: 100000
  selector:
    matchLabels:
      app: bcs-gamestatefulset-operator
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
### bcs-hook-operator
####service
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-hook-operator
    release: po
  name: bcs-hook-operator
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app.kubernetes.io/platform: bk-bcs
    app.kubernetes.io/name: bcs-hook-operator
  sessionAffinity: None
  type: NodePort
```
#### servicemonitor

```
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-hook-operator
    release: po
  name: bcs-hook-operator
  namespace: bcs-system
spec:
  endpoints:
  - interval: 30s
    params: {}
    path: /metrics
    port: http
  sampleLimit: 100000
  selector:
    matchLabels:
      app: bcs-hook-operator
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
## 参考文档
[如何监测kubernetes控制组件](https://sysdig.com/blog/monitor-kubernetes-control-plane/)

