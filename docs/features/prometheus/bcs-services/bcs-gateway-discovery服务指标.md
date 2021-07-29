# bcs-gateway-discovery服务metrics指标

## 指标

### 访问bcs-gateway-disovery服务metrics指标
####  bkbcs_gatewaydiscovery\_api\_request\_total\_num
* 标识bcs-gateway-discovery服务访问外部API请求数指标
* label：system, handler, method, status

#### bkbcs_gatewaydiscovery\_api\_request\_latency\_time
* 标识bcs-gateway-discovery服务访问外部API的延迟指标
* label：system, handler, method, status 

### bcs-gateway-discovery访问外部服务metrics指标
#### bkbcs_gatewaydiscovery\_register\_request\_total\_num
* 标识register接口服务访问请求数指标
* label：system, handler, status

#### bkbcs_gatewaydiscovery\_register\_request\_latency\_time
* 标识register接口服务的请求延迟指标
* label：system, handler, status 

#### bkbcs_gatewaydiscovery\_eventchan\_length
* 标识服务发现队列的长度指标
* label：无 

## 指标聚合
### bcs-gateway-discovery服务API接口指标聚合
#### bcs-gateway-discovery API接口维度指标聚合
```
访问外部system接口维度请求总数
sum(bkbcs_gatewaydiscovery_api_request_total_num{}) by(system, handler)
 
访问外部system接口维度时延
(sum(bkbcs_gatewaydiscovery_api_request_latency_time_sum) by (system, handler)) / (sum(bkbcs_gatewaydiscovery_api_request_latency_time_count) by (system, handler))
 
访问外部system接口维度qps
sum(rate(bkbcs_gatewaydiscovery_api_request_total_num[5m])) by (system, handler)
 
访问外部system接口维度成功率
sum(bkbcs_gatewaydiscovery_api_request_total_num{status="success"}) by(system, handler) / sum(bkbcs_gatewaydiscovery_api_request_total_num) by(system, handler)
``` 

#### bcs-gateway-discovery API接口方法维度指标聚合
```
访问外部system接口方法维度请求总数
sum(bkbcs_gatewaydiscovery_api_request_total_num{}) by(system, handler,method)

访问外部system接口方法维度时延
(sum(bkbcs_gatewaydiscovery_api_request_latency_time_sum) by (system, handler,method)) / (sum(bkbcs_gatewaydiscovery_api_request_latency_time_count) by (system, handler,method))

访问外部system接口方法维度qps
sum(rate(bkbcs_gatewaydiscovery_api_request_total_num[5m])) by (system, handler,method)

访问外部system接口方法维度成功率
sum(bkbcs_gatewaydiscovery_api_request_total_num{status="success"}) by(system, handler,method) / sum(bkbcs_gatewaydiscovery_api_request_total_num) by(system, handler,method)
```
### bcs-gateway-discovery 服务register接口的请求指标聚合
#### bcs-gateway-discovery register接口的请求指标
```
方法维度请求总数
sum(bkbcs_gatewaydiscovery_register_request_total_num{}) by(system, handler)
 
方法维度时延
(sum(bkbcs_gatewaydiscovery_register_request_latency_time_sum) by (system, handler)) / (sum(bkbcs_gatewaydiscovery_register_request_latency_time_count) by (system, handler))
 
方法维度qps
sum(rate(bkbcs_gatewaydiscovery_register_request_total_num[5m])) by (system, handler)
 
方法维度成功率
sum(bkbcs_gatewaydiscovery_register_request_total_num{status="success"}) by(system, handler) / sum(bkbcs_gatewaydiscovery_register_request_total_num) by(system, handler)
``` 

### bcs-gateway-discovery服务服务发现模块队列指标

#### bcs-gateway-discovery 服务发现模块队列长度
```
服务发现模块队列长度指标
bkbcs_gatewaydiscovery_eventchan_length{}

```

## 创建监控对象
### service和servicemonitor对象
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-gateway-discovery
    release: po
  name: bcs-gateway-discovery
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app.kubernetes.io/instance: bcs-services
    app.kubernetes.io/name: bcs-gateway-discovery
  sessionAffinity: None
  type: NodePort
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-gateway-discovery
    release: po
  name: bcs-gateway-discovery
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
      app: bcs-gateway-discovery
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
