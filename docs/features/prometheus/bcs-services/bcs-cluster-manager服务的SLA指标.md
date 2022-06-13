# bcs-cluster-manager服务的SLA指标

指标详情可参考prom指标定义文件：

github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics/metrics.go

## 指标聚合

### API请求成功率

```
sum(bkbcs_clustermanager_api_request_total_num{status="200"} or bkbcs_clustermanager_api_request_total_num{status="success"})/sum(bkbcs_clustermanager_api_request_total_num{} unless bkbcs_clustermanager_api_request_total_num{status=""})
```

### API请求延迟<100ms

```
sum(bkbcs_clustermanager_api_request_latency_time_bucket{status="200",le="0.01"} or bkbcs_clustermanager_api_request_latency_time_bucket{status="success",le="0.01"})/sum(bkbcs_clustermanager_api_request_latency_time_count{status="200"} or bkbcs_clustermanager_api_request_latency_time_count{status="success"} unless bkbcs_clustermanager_api_request_latency_time_count{status=""})
```

