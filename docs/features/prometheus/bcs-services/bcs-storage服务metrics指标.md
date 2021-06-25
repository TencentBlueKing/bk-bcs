# bcs-storage服务metrics指标

## 指标

### 访问bcs-storage服务metrics指标
####  bkbcs\_storage\_api\_request\_total\_num
* 标识API访问请求数指标
* label：handler,method,code

#### bkbcs\_storage\_api\_request\_duration\_time
* 标识API访问请求的延迟指标
* label：handler, method, code

#### bkbcs\_storage\_api\_requests\_inflight
* 标识接口维度的并发访问数
* label：handler, method

#### bkbcs\_storage\_api\_response\_size\_bytes 
* 标识接口访问返回的body体大小
* label：handler,method,code

#### bkbcs\_storage\_watch\_request\_total\_num 
* 标识watch接口访问的链接数
* label：handler,table

#### bkbcs\_storage\_watch\_response\_size\_bytes 
* 标识watch接口长链接访问的body体大小
* label：handler,table

#### bkbcs\_storage\_queue\_push\_total 
* 标识队列推送数据量指标
* label：name，status

#### bkbcs\_storage\_queue\_latency\_seconds
* 标识队列推送数据的时延指标
* label：name，status


## 指标聚合
### bcs-storage服务API接口指标聚合
#### storage API 全量聚合
```
storage接口请求总数
sum(bkbcs_storage_api_request_total_num)
 
storage API 每秒请求数
sum(rate(bkbcs_storage_api_request_total_num[5m]))
 
storage API 时延
sum(bkbcs_storage_api_request_duration_time_sum{}) / sum(bkbcs_storage_api_request_duration_time_count{})
 
storage API 成功率
sum(bkbcs_storage_api_request_total_num{code="2xx"}) / sum(bkbcs_storage_api_request_total_num)
```

#### storage API接口维度指标聚合
```
storage接口维度请求总数
sum(bkbcs_storage_api_request_total_num{}) by(handler)
 
storage接口维度时延
(sum(bkbcs_storage_api_request_duration_time_sum) by (handler)) / (sum(bkbcs_storage_api_request_duration_time_count) by (handler))
 
storage接口维度qps
sum(rate(bkbcs_storage_api_request_total_num[5m])) by (handler)
sum(irate(bkbcs_storage_api_request_total_num[5m])) by (handler)
 
storage接口维度成功率
sum(bkbcs_storage_api_request_total_num{status="2xx"}) by(handler) / sum(bkbcs_storage_api_request_total_num) by(handler)
``` 

#### storage API接口方法维度指标聚合
```
storage接口方法维度请求总数
sum(bkbcs_storage_api_request_total_num{}) by(handler,method)

storage接口方法维度时延
(sum(bkbcs_storage_api_request_duration_time_sum) by (handler,method)) / (sum(bkbcs_storage_api_request_duration_time_count) by (handler,method))

storage接口方法维度qps
sum(rate(bkbcs_storage_api_request_total_num[5m])) by (handler,method)
sum(irate(bkbcs_storage_api_request_total_num[5m])) by (handler,method)

storage接口方法维度成功率
sum(bkbcs_storage_api_request_total_num{code="2xx"}) by(handler,method) / sum(bkbcs_storage_api_request_total_num) by(handler,method)

```

#### 接口维度并发量聚合
```
接口维度并发量
sum(bkbcs_storage_api_requests_inflight) by(handler, method)
```

#### mongo接口方法访问聚合指标

```
mongo访问请求总数
sum(bkbcs_storage_driver_mongdb_total{cluster_id=~"$cluster_id", instance=~"$instance"}) by(method, instance)

表类型的请求延迟
(sum(bkbcs_storage_driver_mongodb_latency_seconds_sum{cluster_id=~"$cluster_id", instance=~"$instance"}) by (method, instance)) / (sum(bkbcs_storage_driver_mongodb_latency_seconds_count{cluster_id=~"$cluster_id", instance=~"$instance"}) by (method, instance))

mongo请求qps
sum(rate(bkbcs_storage_driver_mongdb_total{cluster_id=~"$cluster_id", instance=~"$instance"}[2m])) by (method,instance)
sum(irate(bkbcs_storage_driver_mongdb_total{cluster_id=~"$cluster_id", instance=~"$instance"}[2m])) by (method,instance)

mongo请求的成功率
sum(bkbcs_storage_driver_mongdb_total{cluster_id=~"$cluster_id", instance=~"$instance", status=~"SUCCESS"}) by(method,instance) / sum(bkbcs_storage_driver_mongdb_total{cluster_id=~"$cluster_id", instance=~"$instance"}) by(method,instance)

```
#### watch接口metrics指标聚合
```
watch表的链接数统计
sum(bkbcs_storage_watch_request_total_num{}) by(handler, table)

watch长链接的reponse数据量统计
sum(bkbcs_storage_watch_response_size_bytes{}) by(handler, table)
```
### rabbitmq队列指标聚合
#### 队列推送数据指标
```
队列推送的数据量指标
sum(bkbcs_storage_queue_push_total{}) by(name)
 
队列推送数据的平均时延
(sum(bkbcs_storage_queue_latency_seconds_sum) by (name)) / (sum(bkbcs_storage_queue_latency_seconds_count) by (name))
 
队列推送数据qps
sum(rate(bkbcs_storage_queue_push_total[5m])) by (name)
sum(irate(bkbcs_storage_queue_push_total[5m])) by (name)
 
队列数据推送的成功率
sum(bkbcs_storage_queue_push_total{status="2xx"}) by(name) / sum(bkbcs_storage_queue_push_total) by(name)

```
## 创建监控对象
### service和servicemonitor对象
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-storage
    release: po
  name: bcs-storage
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app.kubernetes.io/instance: bcs-services
    app.kubernetes.io/name: bcs-storage
  sessionAffinity: None
  type: NodePort
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-storage
    release: po
  name: bcs-storage
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
      app: bcs-storage
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
