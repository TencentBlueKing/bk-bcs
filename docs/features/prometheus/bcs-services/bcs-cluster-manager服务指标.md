# bcs-cluster-manager服务metrics指标

## 指标

### 访问bcs-cluster-manager服务metrics指标
####  bkbcs\_clustermanager\_api\_request\_total\_num
* 标识`bcs-cluster-manager`服务API访问请求数指标
* label： handler, method, status

#### bkbcs\_clustermanager\_api\_request\_latency\_time
* 标识API访问请求的延迟指标
* label：handler, method, status 

### bcs-cluster-manager访问外部服务metrics指标
####  bkbcs\_clustermanager\_lib\_request\_total_num
* 标识服务API访问请求数指标
* label：system, handler, method, status

#### bkbcs\_clustermanager\_lib\_request\_latency\_time
* 标识访问API请求的延迟指标
* label：system, handler, method, status 


## 指标聚合
### bcs-cluster-manager服务API接口指标聚合
#### bcs-cluster-manager API接口维度指标聚合
```
clustermanager接口维度请求总数
sum(bkbcs_clustermanager_api_request_total_num{}) by(handler)
 
clustermanager接口维度时延
(sum(bkbcs_clustermanager_api_request_latency_time_sum) by (handler)) / (sum(bkbcs_clustermanager_api_request_latency_time_count) by (handler))
 
clustermanager接口维度qps
sum(rate(bkbcs_clustermanager_api_request_total_num[5m])) by (handler)
 
clustermanager接口维度成功率
sum(bkbcs_clustermanager_api_request_total_num{status="success"}) by(handler) / sum(bkbcs_clustermanager_api_request_total_num) by(handler)
``` 

#### bcs-cluster-manager API接口方法维度指标聚合
```
clustermanager接口方法维度请求总数
sum(bkbcs_clustermanager_api_request_total{}) by(handler,method)

clustermanager接口方法维度时延
(sum(bkbcs_clustermanager_api_request_latency_seconds_sum) by (handler,method)) / (sum(bkbcs_clustermanager_api_request_latency_seconds_count) by (handler,method))

clustermanager接口方法维度qps
sum(rate(bkbcs_clustermanager_api_request_total[5m])) by (handler,method)

clustermanager接口方法维度成功率
sum(bkbcs_clustermanager_api_request_total{status="ok"}) by(handler,method) / sum(bkbcs_clustermanager_api_request_total) by(handler,method)
```
### bcs-cluster-manager服务访问外部接口metrics聚合
#### bcs-cluster-manager 访问外部接口维度指标聚合
```
系统接口维度请求总数
sum(bkbcs_clustermanager_lib_request_total_num{}) by(sysem, handler)
 
系统接口维度时延
(sum(bkbcs_clustermanager_lib_request_latency_time_sum) by (system, handler)) / (sum(bkbcs_clustermanager_lib_request_latency_time_count) by (system, handler))
 
系统接口维度qps
sum(rate(bkbcs_clustermanager_lib_request_total_num[5m])) by (system, handler)
 
系统接口维度成功率
sum(bkbcs_clustermanager_lib_request_total_num{status="success"}) by(system, handler) / sum(bkbcs_clustermanager_lib_request_total_num) by(system, handler)
``` 

#### bcs-cluster-manager 访问外部接口方法维度指标聚合
```
系统接口方法维度请求总数
sum(bkbcs_clustermanager_lib_request_total_num{}) by(system, handler,method)

系统接口方法维度时延
(sum(bkbcs_clustermanager_lib_request_latency_time_sum) by (system, handler,method)) / (sum(bkbcs_clustermanager_lib_request_latency_time_count) by (system, handler,method))

系统接口方法维度qps
sum(rate(bkbcs_clustermanager_lib_request_total_num[5m])) by (system, handler,method)

系统接口方法维度成功率
sum(bkbcs_clustermanager_lib_request_total_num{status="success"}) by(system, handler,method) / sum(bkbcs_clustermanager_lib_request_total_num) by(system, handler,method)
```

## 创建监控对象
### service和servicemonitor对象
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-cluster-manager
    release: po
  name: bcs-cluster-manager
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app.kubernetes.io/instance: bcs-services
    app.kubernetes.io/name: bcs-cluster-manager
  sessionAffinity: None
  type: NodePort
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-cluster-manager
    release: po
  name: bcs-cluster-manager
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
      app: bcs-cluster-manager
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
