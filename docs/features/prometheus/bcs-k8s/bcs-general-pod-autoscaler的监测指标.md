# bcs-general-pod-autoscaler的prom指标

指标详情可参考prom指标定义文件：

bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/metrics/prometheus_metrics.go

## 指标说明

### 指标中各label的含义

| 标签名       | 含义                                                         |
| :----------- | ------------------------------------------------------------ |
| namespace    | gpa所在的命名空间                                            |
| name         | gpa名称                                                      |
| metric       | metric模式下指缩放参考指标名称(如cpu、memory)；webhook模式下指service的namespace/name或url；time模式下指的是具体的schedule |
| scaledObject | 被缩放的对象(kind/name)                                      |
| scaler       | 缩放模式类型(metric,webhook,time等)                          |

### 指标含义

#### keda_metrics_adapter_scaler_errors_total

- 所有scaler的错误总数

#### keda_metrics_adapter_scaler_errors

- 单个scaler的错误数量
- labels：namespace,name, scaledObject, scaler, metric

#### keda_metrics_adapter_scaled_object_errors

- 被缩放对象的错误数量
- labels：namespace,name, scaledObject

#### keda_metrics_adapter_scaler_target_metrics_value

在metric模式下，该指标的含义：

- gpa中设定的参考指标的值
- labels：namespace,name, scaledObject, scaler, metric

在webhook/time模式下，该指标的含义：

- 相应模式下推荐的副本数
- labels：namespace,name, scaledObject, scaler, metric

#### keda_metrics_adapter_scaler_current_metrics_value

在metric模式下，该指标的含义：

- 参考指标当前的数值
- labels：namespace,name, scaledObject, scaler, metric

在webhook/time模式下，该指标的含义：

- 被缩放对象当前的副本数
- labels：namespace,name, scaledObject, scaler, metric

#### keda_metrics_adapter_scaler_desired_replicas_value

- 相应模式下推荐的副本数
- labels：namespace,name, scaledObject, scaler

#### keda_metrics_adapter_gpa_desired_replicas_value

- gpa推荐的副本数
- labels：namespace,name, scaledObject

#### keda_metrics_adapter_gpa_min_replicas_value

- gpa设置的最小副本数
- labels: namespace,name, scaledObject

#### keda_metrics_adapter_gpa_max_replicas_value

- gpa设置的最大副本数
- labels: namespace,name, scaledObject

## 指标聚合

### 错误

```
# 3分钟报错率
rate(keda_metrics_adapter_scaler_errors_total{}[3m]) 
```

### 延迟

```
# scaler执行平均延迟
sum(keda_metrics_adapter_scaler_exec_duration_sum{namespace=~"$namespace",name=~"$name",scaledObject=~"$scaledObject"}) by (namespace,name,scaledObject)/sum(keda_metrics_adapter_scaler_exec_duration_count{namespace=~"$namespace",name=~"$name",scaledObject=~"$scaledObject"}) by (namespace,name,scaledObject)

# 副本更新平均延迟
sum(keda_metrics_adapter_gpa_update_duration_sum{namespace=~"$namespace",name=~"$name",scaledObject=~"$scaledObject",status="success"}) by (namespace,name,scaledObject)/sum(keda_metrics_adapter_gpa_update_duration_count{namespace=~"$namespace",name=~"$name",scaledObject=~"$scaledObject",status="success"}) by (namespace,name,scaledObject)
```

