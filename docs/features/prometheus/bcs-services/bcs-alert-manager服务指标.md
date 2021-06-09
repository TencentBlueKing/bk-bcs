# bcs-alert-manager服务metrics指标

## 指标

### 访问bcs-alert-manager服务metrics指标
####  bkbcs\_alertmanager\_api\_request\_total_num
* 标识bcs-alert-manager服务API访问请求数指标
* label： handler, method, status

#### bkbcs\_alertmanager\_api\_request\_latency\_time
* 标识API访问请求的延迟指标
* label：handler, method, status 

####  bkbcs\_alertmanager\_lib\_request\_total_num
* 标识alertmanager组件API访问请求数指标
* label： handler, method, status 

#### bkbcs\_alertmanager\_lib\_request\_latency\_time
* 标识alertmanager组件API访问请求的延迟指标
* label：handler, method, status 

#### bkbcs\_alertmanager\_handler\_queue\_total\_num
* 标识handler队列长度指标
* label：handler

#### bkbcs\_alertmanager\_handler\_request\_latency\_time
* 标识队列handler处理event事件延迟指标
* label：handler、name、status

## 指标聚合
### eventhandler队列指标聚合
```
instance实例handler类型队列长度
sum(bkbcs_alertmanager_handler_queue_total_num{cluster_id=~"$cluster_id",instance=~"$instance"}) by(cluster_id, instance, handler)

```   

### 队列handler处理event事件延迟指标
```
instance实例event事件处理平均延迟
(sum(bkbcs_alertmanager_handler_request_latency_time_sum{cluster_id=~"$cluster_id",instance=~"$instance"}) by(instance, handler, name, status)) / (sum(bkbcs_alertmanager_handler_request_latency_time_count{cluster_id=~"$cluster_id",instance=~"$instance"}) by(instance, handler, name, status))
```

### bcs-alert-manager服务API接口指标聚合
#### alert-manager API 全量聚合
```
alertmanager接口请求总数
sum(bkbcs_alertmanager_api_request_total_num)
 
alertmanager API 每秒请求数
sum(rate(bkbcs_alertmanager_api_request_total_num[5m]))
 
alertmanager API 时延
sum(bkbcs_alertmanager_api_request_latency_time_sum{}) / sum(bkbcs_alertmanager_api_request_latency_time_count{})
 
alertmanager API 成功率
sum(bkbcs_alertmanager_api_request_total_num{status="success"}) / sum(bkbcs_alertmanager_api_request_total_num)
```

#### alertmanager API接口维度指标聚合
```
alertmanager接口维度请求总数
sum(bkbcs_alertmanager_api_request_total_num{}) by(handler)
 
alertmanager接口维度时延
(sum(bkbcs_alertmanager_api_request_latency_time_sum) by (handler)) / (sum(bkbcs_alertmanager_api_request_latency_time_count) by (handler))
 
alertmanager接口维度qps
sum(rate(bkbcs_alertmanager_api_request_total_num[5m])) by (handler)
 
alertmanager接口维度成功率
sum(bkbcs_alertmanager_api_request_total_num{status="success"}) by(handler) / sum(bkbcs_alertmanager_api_request_total_num) by(handler)
``` 

#### alertmanager API接口方法维度指标聚合
```
alertmanager接口方法维度请求总数
sum(bkbcs_alertmanager_api_request_total_num{}) by(handler,method)

alertmanager接口方法维度时延
(sum(bkbcs_alertmanager_api_request_latency_time_sum) by (handler,method)) / (sum(bkbcs_alertmanager_api_request_latency_time_count) by (handler,method))

alertmanager接口方法维度qps
sum(rate(bkbcs_alertmanager_api_request_total_num[5m])) by (handler,method)

alertmanager接口方法维度成功率
sum(bkbcs_alertmanager_api_request_total_num{status="success"}) by(handler,method) / sum(bkbcs_alertmanager_api_request_total_num) by(handler,method)
```
#### 外部接口指标聚合
 
```
alert接口维度请求总数
sum(bkbcs_alertmanager_lib_request_total_num{}) by(handler)
 
alert接口维度时延
(sum(bkbcs_alertmanager_lib_request_latency_time_sum) by (handler)) / (sum(bkbcs_alertmanager_lib_request_latency_time_count) by (handler))
 
alert接口维度qps
sum(rate(bkbcs_alertmanager_lib_request_total_num[2m])) by (handler)
 
alert接口维度成功率
sum(bkbcs_alertmanager_lib_request_total_num{status="success"}) by(handler) / sum(bkbcs_alertmanager_lib_request_total_num) by(handler)
```   

## 创建监控对象
### service和servicemonitor对象
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-alert-manager
    release: po
  name: bcs-alert-manager
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  - name: https
    port: 50029
    protocol: TCP
    targetPort: 50029
  selector:
    app.kubernetes.io/instance: bcs-alert-manager
    app.kubernetes.io/name: bcs-alert-manager
  sessionAffinity: None
  type: NodePort
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-alert-manager
    release: po
  name: bcs-alert-manager
  namespace: bcs-system
spec:
  endpoints:
  - interval: 30s
    params: {}
    path: /alertmanager/metrics
    port: http
  sampleLimit: 100000
  selector:
    matchLabels:
      app: bcs-alert-manager
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
