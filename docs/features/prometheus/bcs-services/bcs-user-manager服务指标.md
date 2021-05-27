# bcs-user-manager服务metrics指标

## 指标

###  bkbcs_usermanager\_api\_request\_total\_num
* 标识bcs-user-manager服务API请求数指标
* label：handler, method, status

### bkbcs_usermanager\_api\_request\_latency\_time
* 标识bcs-user-manager服务API的延迟指标
* label：handler, method, status 

## 指标聚合
### bcs-user-manager服务API接口指标聚合
#### bcs-user-manager API接口维度指标聚合
```
接口维度请求总数
sum(bkbcs_usermanager_api_request_total_num{}) by(handler)
 
接口维度时延
(sum(bkbcs_usermanager_api_request_latency_time_sum) by (handler)) / (sum(bkbcs_usermanager_api_request_latency_time_count) by (handler))
 
接口维度qps
sum(rate(bkbcs_usermanager_api_request_total_num[5m])) by (handler)

接口维度成功率
sum(bkbcs_usermanager_api_request_total_num{status="success"}) by(handler) / sum(bkbcs_usermanager_api_request_total_num) by(handler)
``` 

#### bcs-user-manager API接口方法维度指标聚合
```
接口方法维度请求总数
sum(bkbcs_usermanager_api_request_total_num{}) by(handler,method)
 
接口方法维度时延
(sum(bkbcs_usermanager_api_request_latency_time_sum) by (handler,method)) / (sum(bkbcs_usermanager_api_request_latency_time_count) by (handler,method))
 
接口方法维度qps
sum(rate(bkbcs_usermanager_api_request_total_num[5m])) by (handler,method)

接口方法维度成功率
sum(bkbcs_usermanager_api_request_total_num{status="success"}) by(handler,method) / sum(bkbcs_usermanager_api_request_total_num) by(handler,method)

```

## 创建监控对象
### service和servicemonitor对象
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-user-manager
    release: po
  name: bcs-user-manager
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app.kubernetes.io/instance: bcs-services
    app.kubernetes.io/name: bcs-user-manager
  sessionAffinity: None
  type: NodePort
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-user-manager
    release: po
  name: bcs-user-manager
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
      app: bcs-user-manager
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
