# bcs-general-pod-autoscaler的SLA prom指标

指标详情可参考prom指标定义文件：

bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/metrics/prometheus_metrics.go及bk-bcs/docs/features/prometheus/bcs-k8s/bcs-general-pod-autoscaler的监测指标.md

## 指标聚合

### 扩缩容成功率

```
# 扩缩容成功率
sum(keda_metrics_adapter_gpa_update_duration_count{status="success"})/sum(keda_metrics_adapter_gpa_update_duration_count{})
```

### 扩缩容延迟<250ms

```
# 扩缩容延迟<250ms
sum(keda_metrics_adapter_scaler_exec_duration_bucket{le="0.25",status="success"} or keda_metrics_adapter_gpa_update_duration_bucket{le="0.25",status="success"}) /sum(keda_metrics_adapter_scaler_exec_duration_count{status="success"} or keda_metrics_adapter_gpa_update_duration_count{status="success"})

# scaler执行延迟<250ms
sum(keda_metrics_adapter_scaler_exec_duration_bucket{le="0.25"})/sum(keda_metrics_adapter_scaler_exec_duration_count{})

# 副本数更新延迟<250ms
sum(keda_metrics_adapter_gpa_update_duration_bucket{le="0.25",status="success"})/sum(keda_metrics_adapter_gpa_update_duration_count{status="success"})
```

