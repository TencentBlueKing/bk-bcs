# bcs-webhook-server服务metrics指标

## 指标

###  bkbcs_webhookserver\_api\_request\_total\_num
* 标识bcs-webhook-server服务API请求数指标
* label：handler, method, status

### bkbcs_webhookserver\_api\_request\_latency\_time
* 标识bcs-webhook-server服务API的延迟指标
* label：handler, method, status 

### bkbcs_webhookserver\_plugin\_request\_latency\_time
* 标识bcs-webhook-server服务插件执行时延
* label：pluginName, status

## 指标聚合
### bcs-webhook-server服务API接口指标聚合
#### bcs-webhook-server API接口维度指标聚合
```
接口维度请求总数
sum(bkbcs_webhookserver_api_request_total_num{}) by(handler)
 
接口维度时延
(sum(bkbcs_webhookserver_api_request_latency_time_sum) by (handler)) / (sum(bkbcs_webhookserver_api_request_latency_time_count) by (handler))
 
接口维度qps
sum(rate(bkbcs_webhookserver_api_request_total_num[5m])) by (handler)

接口维度成功率
sum(bkbcs_webhookserver_api_request_total_num{status="success"}) by(handler) / sum(bkbcs_webhookserver_api_request_total_num) by(handler)
``` 

#### bcs-webhook-server API接口方法维度指标聚合
```
接口方法维度请求总数
sum(bkbcs_webhookserver_api_request_total_num{}) by(handler,method)
 
接口方法维度时延
(sum(bkbcs_webhookserver_api_request_latency_time_sum) by (handler,method)) / (sum(bkbcs_webhookserver_api_request_latency_time_count) by (handler,method))
 
接口方法维度qps
sum(rate(bkbcs_webhookserver_api_request_total_num[5m])) by (handler,method)

接口方法维度成功率
sum(bkbcs_webhookserver_api_request_total_num{status="success"}) by(handler,method) / sum(bkbcs_webhookserver_api_request_total_num) by(handler,method)

```
### bcs-webhook-server服务插件执行指标聚合
#### bcs-webhook-server 插件执行指标聚合
```
插件运行时延
(sum(bkbcs_webhookserver_plugin_request_latency_time_sum) by (pluginName)) / (sum(bkbcs_webhookserver_plugin_request_latency_time_count) by (pluginName))

``` 

## 创建监控对象
### service和servicemonitor对象
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-webhook-server
    release: po
  name: bcs-webhook-server
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app.kubernetes.io/instance: bcs-services
    app.kubernetes.io/name: bcs-webhook-server
  sessionAffinity: None
  type: NodePort
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-webhook-server
    release: po
  name: bcs-webhook-server
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
      app: bcs-webhook-server
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
