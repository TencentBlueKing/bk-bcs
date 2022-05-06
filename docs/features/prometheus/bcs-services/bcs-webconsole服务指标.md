# bcs-webconsole服务metrics指标

## 指标

###  bkbcs_webconsole_api_request_total_num
* 标识bcs-web-console服务API请求数指标
* label：handler, method, status

### bkbcs_webconsole_api_request_latency_time
* 标识bcs-web-console服务API的延迟指标
* label：handler, method, status 

### bkbcs_webconsole_pod_create_total_num
* 标识bcs-web-console创建pod数量
* label:namespace, name, status

### bkbcs_webconsole_pod_create_duration_seconds_max
* 创建pod最大时间
* lable:namespace, name, status

### bkbcs_webconsole_pod_create_duration_seconds_min
* 创建pod最小时间
* lable:namespace, name, status

### bkbcs_webconsole_pod_create_duration_seconds_bucket
* 标识bcs-web-console pod创建耗费时间
* lable:namespace, name, status

### bkbcs_webconsole_pod_delete_total_num
* 标识bcs-web-console清除pod数量
* label:namespace, name, status

### bkbcs_webconsole_pod_delete_duration_seconds_bucket
* 清除pod耗费时间
* lable:namespace, name, status

### bkbcs_webconsole_pod_delete_duration_seconds_max
* 清除pod最大时间
* lable:namespace, name, status

### bkbcs_webconsole_pod_delete_duration_seconds_min
* 清除pod最小时间
* lable:namespace, name, status

### bkbcs_webconsole_ws_connection_total_num
* 标识 websocke 连接数
* label: namespace, name

### bkbcs_webconsole_ws_close_total_num
* 标识 websocke 断开连接数量
* label: namespace, name

### bkbcs_webconsole_ws_connection_duration_seconds_bucket
* 标识 websocke 连接耗费时间
* label: namespace, name

## 指标聚合
### bcs-webconsole服务API接口指标聚合
#### bcs-webconsole API接口维度指标聚合
```
接口维度请求总数
sum(bkbcs_webconsole_api_request_total_num{}) by(handler)
 
接口维度时延
(sum(bkbcs_webconsole_api_request_latency_time) by  (handler)) / (sum(bkbcs_webconsole_api_request_total_num) by 
(handler))
 
接口维度qps
sum(rate(bkbcs_webconsole_api_request_total_num[5m])) by (handler)

接口维度成功率
sum(bkbcs_webconsole_api_request_total_num{status="success"}) by(handler) / sum(bkbcs_webconsole_api_request_total_num) by(handler)
``` 

#### bcs-webconsole API接口方法维度指标聚合
```
接口方法维度请求总数
sum(bkbcs_webconsole_api_request_total_num{}) by(handler,method)
 
接口方法维度时延
(sum(bkbcs_webconsole_api_request_latency_time) by (handler,method)) / (sum(bkbcs_webconsole_api_request_total_num) by (handler,method))
 
接口方法维度qps
sum(rate(bkbcs_webconsole_api_request_total_num[5m])) by (handler,method)

接口方法维度成功率
sum(bkbcs_webconsole_api_request_total_num{status="success"}) by(handler,method) / sum(bkbcs_webconsole_api_request_total_num) by(handler,method)
```

#### bcs-webconsole pod维度指标聚合
```
创建pod成功率
sum(bkbcs_webconsole_pod_create_total_num{status="success"}) / sum(bkbcs_webconsole_pod_create_total_num{status="failure"})

pod 创建成功耗时(s)分布
sum(bkbcs_webconsole_pod_create_duration_seconds_bucket{status="success"}) by(le)

pod 创建失败耗时(s)分布
sum(bkbcs_webconsole_pod_create_duration_seconds_bucket{status="failure"}) by(le)

pod 删除成功耗时(s)分布
sum(bkbcs_webconsole_pod_create_duration_seconds_bucket{status="failure"}) by(le)

pod 删除失败耗时(s)分布
sum(bkbcs_webconsole_pod_create_duration_seconds_bucket{status="failure"}) by(le)

当前pod数量
sum(bkbcs_webconsole_pod_create_duration_seconds_count{status="success"}) - 
sum(bkbcs_webconsole_pod_create_duration_seconds_count{status="failure"})

pod创建耗时极值情况
{{status}}_max: max(bkbcs_webconsole_pod_create_duration_seconds_max) by(status) {{status}}_min: min
(bkbcs_webconsole_pod_create_duration_seconds_min) by(status) 

pod删除耗时极值情况
{{status}}_max: max(bkbcs_webconsole_pod_delete_duration_seconds_max) by(status) {{status}}_min: min
(bkbcs_webconsole_pod_delete_duration_seconds_min) by(status) 
```

#### bcs-webconsole ws连接维度指标聚合
```
bcs-webconsole websocket连接次数
sum(bkbcs_webconsole_ws_connection_num{}) by(namespace, name)

ws 连接延时
(sum(bkbcs_webconsole_ws_connection_duration_seconds_bucket) by (namespace, name)) / (sum
(bkbcs_webconsole_ws_connection_total_num) by (namespace, name))

ws连接存活数量
sum(bkbcs_webconsole_ws_connection_total_num) - sum(bkbcs_webconsole_ws_close_total_num)
```

## 创建监控对象
### service和servicemonitor对象
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-web-console
    release: po
  name: bcs-web-console
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app.kubernetes.io/instance: bcs-services
    app.kubernetes.io/name: bcs-web-console
  sessionAffinity: None
  type: NodePort
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    io.tencent.bcs.service_name: bcs-web-console
    release: po
  name: bcs-web-console
  namespace: bcs-system
spec:
  endpoints:
  - interval: 30s
    params: {}
    path: /-/metrics
    port: http
  sampleLimit: 100000
  selector:
    matchLabels:
      app: bcs-web-console
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
