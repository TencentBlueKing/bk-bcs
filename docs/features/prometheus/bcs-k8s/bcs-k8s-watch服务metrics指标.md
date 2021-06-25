# bcs-k8s-watch服务metrics指标

## 指标

### 访问bcs-k8s-watch服务metrics指标
####  bkbcs\_k8swatch\_storage\_request\_total_num
* 标识API访问请求数指标
* label： cluster_id, handler, namespace, resource_type, method, status 

#### bkbcs_k8swatch\_storage\_request\_latency\_time
* 标识API访问请求的延迟指标
* label：cluster_id, handler, namespace, resource_type, method, status 


#### bkbcs\_k8swatch\_queue\_handler\_total\_num
* 标识handler队列长度指标
* label：cluster_id、handler

#### bkbcs\_k8swatch\_queue\_handler\_discard\_events 
* 标识handler队列丢弃事件数
* label：cluster_id、handler

## 指标聚合
### 队列指标聚合
```
handler类型队列长度
sum(bkbcs_k8swatch_queue_handler_total_num{handler!~"writer_normal_queue|writer_alarm_queue"}) by(handler)
 
writer队列长度
sum(bkbcs_k8swatch_queue_handler_total_num{handler=~"writer_normal_queue|writer_alarm_queue"}) by(handler)
 
队列丢弃事件
sum(bkbcs_k8swatch_queue_handler_discard_events{}) by(handler)

```   
### bcs-storage服务API接口指标聚合
#### storage API 全量聚合
```
storage接口请求总数
sum(bkbcs_k8swatch_storage_request_total_num)
 
storage API 每秒请求数
sum(rate(bkbcs_k8swatch_storage_request_total_num[5m]))
 
storage API 时延
sum(bkbcs_k8swatch_storage_request_latency_time_sum{}) / sum(bkbcs_k8swatch_storage_request_latency_time_count{})
 
storage API 成功率
sum(bkbcs_k8swatch_storage_request_total_num{status="success"}) / sum(bkbcs_k8swatch_storage_request_total_num)
```

#### storage API接口维度指标聚合
```
storage接口维度请求总数
sum(bkbcs_k8swatch_storage_request_total_num{}) by(handler)
 
storage接口维度时延
(sum(bkbcs_k8swatch_storage_request_latency_time_sum) by (handler)) / (sum(bkbcs_k8swatch_storage_request_latency_time_count) by (handler))
 
storage接口维度qps
sum(rate(bkbcs_k8swatch_storage_request_total_num[5m])) by (handler)
 
storage接口维度成功率
sum(bkbcs_k8swatch_storage_request_total_num{status="success"}) by(handler) / sum(bkbcs_k8swatch_storage_request_total_num) by(handler)
``` 

#### storage API接口方法维度指标聚合
```
storage接口方法维度请求总数
sum(bkbcs_k8swatch_storage_request_total_num{}) by(handler,method)

storage接口方法维度时延
(sum(bkbcs_k8swatch_storage_request_latency_time_sum) by (handler,method)) / (sum(bkbcs_k8swatch_storage_request_latency_time_count) by (handler,method))

storage接口方法维度qps
sum(rate(bkbcs_k8swatch_storage_request_total_num[5m])) by (handler,method)

storage接口方法维度成功率
sum(bkbcs_k8swatch_storage_request_total_num{status="success"}) by(handler,method) / sum(bkbcs_k8swatch_storage_request_total_num) by(handler,method)
```
 
#### 其他维度指标聚合
```
请求类型维度
sum(bkbcs_k8swatch_storage_request_total_num{status="success"}) by (method)
sum(bkbcs_k8swatch_storage_request_total_num{status!="success"}) by (method)
 
请求类型每秒请求数
sum(irate(bkbcs_k8swatch_storage_request_total_num{status="success"}[10m])) by (method)
sum(irate(bkbcs_k8swatch_storage_request_total_num{status!="success"}[10m])) by (method)
 
接口请求频率
sum(irate(bkbcs_k8swatch_storage_request_total_num{}[10m])) by (handler)
 
url请求总量
sum(bkbcs_k8swatch_storage_request_total_num{status!="success"}) by (handler)
sum(bkbcs_k8swatch_storage_request_total_num{status="success"}) by (handler)
```
## 创建监控对象
###service
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-k8s-watch
    release: po
  name: bcs-k8s-watch
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app: bcs-k8s-watch
    platform: bk-bcs
  sessionAffinity: None
  type: NodePort
```
### servicemonitor

```
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-k8s-watch
    release: po
  name: bcs-k8s-watch
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
      app: bcs-k8s-watch
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
  