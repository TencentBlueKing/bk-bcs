# bcs-k8s-watch服务metrics指标

## 指标

### 访问bcs-storage服务metrics指标
####  bcs\_k8s\_watch\_storage\_request\_total_num
* 标识API访问请求数指标
* label： cluster_id, handler, namespace, resource_type, method, status 

#### bcs\_k8s\_watch\_storage\_request\_latency\_time
* 标识API访问请求的延迟指标
* label：cluster_id, handler, namespace, resource_type, method, status 


#### bcs\_k8s\_watch\_queue\_queue\_total\_num
* 标识handler队列长度指标
* label：cluster_id、handler

#### bcs\_k8s\_watch\_queue\_handler\_discard\_events 
* 标识handler队列丢弃事件数
* label：cluster_id、handler 

## 指标聚合
### 队列指标聚合
```
handler类型队列长度
sum(bcs_k8s_watch_queue_queue_total_num{handler!~"writer_normal_queue|writer_alarm_queue"}) by(handler)
 
writer队列长度
sum(bcs_k8s_watch_queue_queue_total_num{handler=~"writer_normal_queue|writer_alarm_queue"}) by(handler)
 
队列丢弃事件
sum(bcs_k8s_watch_queue_handler_discard_events{}) by(handler)

```   
### bcs-storage服务API接口指标聚合
#### storage API 全量聚合
```
storage接口请求总数
sum(bcs_k8s_watch_storage_request_total_num)
 
storage API 每秒请求数
sum(rate(bcs_k8s_watch_storage_request_total_num[5m]))
 
storage API 时延
sum(bcs_k8s_watch_storage_request_latency_time_sum{}) / sum(bcs_k8s_watch_storage_request_latency_time_count{})
 
storage API 成功率
sum(bcs_k8s_watch_storage_request_total_num{status="success"}) / sum(bcs_k8s_watch_storage_request_total_num)
```

#### storage API接口维度指标聚合
```
storage接口维度请求总数
sum(bcs_k8s_watch_storage_request_total_num{}) by(handler)
 
storage接口维度时延
(sum(bcs_k8s_watch_storage_request_latency_time_sum) by (handler)) / (sum(bcs_k8s_watch_storage_request_latency_time_count) by (handler))
 
storage接口维度qps
sum(rate(bcs_k8s_watch_storage_request_total_num[5m])) by (handler)
 
storage接口维度成功率
sum(bcs_k8s_watch_storage_request_total_num{status="success"}) by(handler) / sum(bcs_k8s_watch_storage_request_total_num) by(handler)
``` 

#### storage API接口方法维度指标聚合
```
storage接口方法维度请求总数
sum(bcs_k8s_watch_storage_request_total_num{}) by(handler,method)

storage接口方法维度时延
(sum(bcs_k8s_watch_storage_request_latency_time_sum) by (handler,method)) / (sum(bcs_k8s_watch_storage_request_latency_time_count) by (handler,method))

storage接口方法维度qps
sum(rate(bcs_k8s_watch_storage_request_total_num[5m])) by (handler,method)

storage接口方法维度成功率
sum(bcs_k8s_watch_storage_request_total_num{status="success"}) by(handler,method) / sum(bcs_k8s_watch_storage_request_total_num) by(handler,method)
```
 
#### 其他维度指标聚合
```
请求类型维度
sum(bcs_k8s_watch_storage_request_total_num{status="success"}) by (method)
sum(bcs_k8s_watch_storage_request_total_num{status!="success"}) by (method)
 
请求类型每秒请求数
sum(irate(bcs_k8s_watch_storage_request_total_num{status="success"}[10m])) by (method)
sum(irate(bcs_k8s_watch_storage_request_total_num{status!="success"}[10m])) by (method)
 
接口请求频率
sum(irate(bcs_k8s_watch_storage_request_total_num{}[10m])) by (handler)
 
url请求总量
sum(bcs_k8s_watch_storage_request_total_num{status!="success"}) by (handler)
sum(bcs_k8s_watch_storage_request_total_num{status="success"}) by (handler)
```
  