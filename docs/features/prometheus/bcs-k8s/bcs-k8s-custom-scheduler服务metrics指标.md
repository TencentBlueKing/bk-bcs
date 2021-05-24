# bcs-k8s-custom-scheduler服务metrics指标

## 指标

### bcs-k8s-custom-scheduler服务metrics指标
####  bkbcs\_k8scustomscheduler\_api\_request\_total\_num
* 标识bcs-k8s-custom-scheduler服务的API请求数指标
* label：version, handler, method, status

#### bkbcs\_k8scustomscheduler\_api\_request\_latency\_time
* 标识访问clustermanager请求的延迟指标
* label：version, handler, method, status

#### bkbcs\_k8scustomscheduler\_filter\_node\_total\_num
* 标识能够调度及不能调度的宿主机数目
* label：handler


## 指标聚合
### bcs-k8s-custom-scheduler服务API接口指标聚合
#### bcs-k8s-custom-scheduler服务API接口维度指标聚合
```
接口维度请求总数
sum(bkbcs_k8scustomscheduler_api_request_total_num{}) by(version, handler)
 
接口维度时延
(sum(bkbcs_k8scustomscheduler_api_request_latency_time_sum) by (version, handler)) / (sum(bkbcs_k8scustomscheduler_api_request_latency_time_count) by (version, handler))
 
接口维度qps
sum(rate(bkbcs_k8scustomscheduler_api_request_total_num[5m])) by (version, handler)
sum(irate(bkbcs_k8scustomscheduler_api_request_total_num[5m])) by (version, handler)
 
接口维度成功率
sum(bkbcs_k8scustomscheduler_api_request_total_num{status="success"}) by(version, handler) / sum(bkbcs_k8scustomscheduler_api_request_total_num) by(version, handler)
``` 

#### bcs-k8s-custom-scheduler服务可访问node数统计
```
bcs-k8s-custom-scheduler服务可访问node数目统计
bkbcs_k8s_customscheduler_filter_node_total_num{} by(version, scheduler)

```
## 创建监控对象
###service
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-k8s-custom-scheduler
    release: po
  name: bcs-k8s-custom-scheduler
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app: bcs-k8s-custom-scheduler
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
    io.tencent.bcs.service_name: bcs-k8s-custom-scheduler
    release: po
  name: bcs-k8s-custom-scheduler
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
      app: bcs-k8s-custom-scheduler
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```
