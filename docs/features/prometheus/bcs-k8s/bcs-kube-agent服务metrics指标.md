# bcs-kube-agent服务metrics指标

## 指标

### 访问bcs-kube-agent服务metrics指标
####  bkbcs\_kubeagent\_clustermanager\_request\_total\_num
* 标识访问clustermanager服务的请求数指标
* label：handler, method, code

#### bkbcs\_kubeagent\_clustermanager\_request\_latency\_time
* 标识访问clustermanager请求的延迟指标
* label：handler, method, code


#### bkbcs\_kubeagent\_clustermanager\_ws\_connection\_num
* 标识websocke链接断开重连次数
* label：handler


## 指标聚合
### bcs-kube-agent服务API接口指标聚合
#### bcs-kube-agent访问clustermanager接口维度指标聚合
```
clustermanager接口维度请求总数
sum(bkbcs_kubeagent_clustermanager_request_total_num{}) by(handler)
 
clustermanager接口维度时延
(sum(bkbcs_kubeagent_clustermanager_request_latency_time_sum) by (handler)) / (sum(bkbcs_kubeagent_clustermanager_request_latency_time_count) by (handler))
 
clustermanager接口维度qps
sum(rate(bkbcs_kubeagent_clustermanager_request_total_num[5m])) by (handler)
sum(irate(bkbcs_kubeagent_clustermanager_request_total_num[5m])) by (handler)
 
clustermanager接口维度成功率
sum(bkbcs_kubeagent_clustermanager_request_total_num{code="200"}) by(handler) / sum(bkbcs_kubeagent_clustermanager_request_total_num) by(handler)
``` 

#### bcs-kube-agent访问websocket接口维度指标聚合
```
clustermanager websocket接口请求断开重连次数
sum(bkbcs_kubeagent_clustermanager_ws_connection_num{}) by(handler)

```
## 创建监控对象
###service
```
apiVersion: v1
kind: Service
metadata:
  labels:
    app: bcs-kube-agent
    release: po
  name: bcs-kube-agent
  namespace: bcs-system
spec:
  ports:
  - name: http
    port: {{port}}
    protocol: TCP
    targetPort: {{port}}
  selector:
    app: bcs-kube-agent
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
    io.tencent.bcs.service_name: bcs-kube-agent
    release: po
  name: bcs-kube-agent
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
      app: bcs-kube-agent
      release: po
  namespaceSelector:
    matchNames:
      - bcs-system
```

   