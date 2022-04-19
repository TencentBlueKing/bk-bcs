# bcs-gamestatefulset-operator的prom指标

指标详情可参考prom指标定义文件：

bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/controllers/metrics.go

# bcs-gamestatefulset-operator prom指标监测场景

### gsts controller调协情况

| 监测场景名称               | PromQL语句                                                   |
| -------------------------- | ------------------------------------------------------------ |
| 协调成功耗时分布           | sum(bkbcs_gamestatefulset_reconcile_duration_seconds_bucket{status="success"}) by (le) |
| 协调失败耗时分布           | sum(bkbcs_gamestatefulset_reconcile_duration_seconds_bucket{status="failure"}) by (le) |
| 各gsts下协调成功数量       | sum(bkbcs_gamestatefulset_reconcile_duration_seconds_count{status="success"}) by (namespace,name) |
| 所有gsts下协调成功数量     | sum(bkbcs_gamestatefulset_reconcile_duration_seconds_count{status="success"}) |
| 各gsts下协调失败数量       | sum(bkbcs_gamestatefulset_reconcile_duration_seconds_count{status="failure"}) by (namespace,name) |
| 所有gsts下协调失败数量     | sum(bkbcs_gamestatefulset_reconcile_duration_seconds_count{status="failure"}) |
| 协调成功耗时前十的gsts资源 | topk(10, sum(bkbcs_gamestatefulset_reconcile_duration_seconds_sum{status="success"}) by (namespace,name)/sum(bkbcs_gamestatefulset_reconcile_duration_seconds_count{status="success"}) by (namespace,name)) |

### 副本情况

| 监测场景名称                    | PromQL语句                                                   |
| ------------------------------- | ------------------------------------------------------------ |
| 期望的副本总数                  | sum(bkbcs_gamestatefulset_replicas)                          |
| 当前副本总数                    | sum(bkbcs_gamestatefulset_current_replicas)                  |
| 处于ready状态的副本数           | sum(bkbcs_gamestatefulset_ready_replicas)                    |
| 已更新的副本数                  | sum(bkbcs_gamestatefulset_updated_replicas)                  |
| 处于Updatedready状态的副本数    | sum(bkbcs_gamestatefulset_updated_ready_replicas)            |
| 期望的副本数占前十的gsts资源    | topk(10,sum(bkbcs_gamestatefulset_replicas)  by (namespace,name)) |
| unready的副本数占前十的gsts资源 | topk(10,sum(abs(bkbcs_gamestatefulset_replicas  - bkbcs_gamestatefulset_ready_replicas)) by (namespace,name)) |

### pod创建情况

| 监测场景名称              | PromQL语句                                                   |
| ------------------------- | ------------------------------------------------------------ |
| 各gsts下pod创建成功数量   | sum(bkbcs_gamestatefulset_pod_create_duration_seconds_count{status="success"}) by (namespace,name) |
| 各gsts下pod创建失败数量   | sum(bkbcs_gamestatefulset_pod_create_duration_seconds_count{status="failure"}) by (namespace,name) |
| 所有gsts下pod创建成功数量 | sum(bkbcs_gamestatefulset_pod_create_duration_seconds_count{status="success"}) |
| 所有gsts下pod创建失败数量 | sum(bkbcs_gamestatefulset_pod_create_duration_seconds_count{status="failure"}) |
| pod创建成功耗时分布       | sum(bkbcs_gamestatefulset_pod_create_duration_seconds_bucket{status="success"}) by (le) |
| pod创建失败耗时分布       | sum(bkbcs_gamestatefulset_pod_create_duration_seconds_bucket{status="failure"}) by (le) |
| 单个pod创建最大耗时       | max(bkbcs_gamestatefulset_pod_create_duration_seconds_max) by (status) |
| 单个pod创建最小耗时       | min(bkbcs_gamestatefulset_pod_create_duration_seconds_min) by (status) |

### pod删除情况

| 监测场景名称              | PromQL语句                                                   |
| ------------------------- | ------------------------------------------------------------ |
| 各gsts下pod删除成功数量   | sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{status="success"}) by (namespace,name) |
| 各gsts下pod删除失败数量   | sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{status="failure"}) by (namespace,name) |
| 所有gsts下pod删除成功数量 | sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{status="success"}) |
| 所有gsts下pod删除失败数量 | sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{status="failure"}) |
| pod删除成功耗时分布       | sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_bucket{status="success"}) by (le) |
| pod删除失败耗时分布       | sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_bucket{status="failure"}) by (le) |
| 单个pod删除最大耗时       | max(bkbcs_gamestatefulset_pod_delete_duration_seconds_max) by (status) |
| 单个pod删除最小耗时       | min(bkbcs_gamestatefulset_pod_delete_duration_seconds_min) by (status) |

### pod更新情况

| 监测场景名称              | PromQL语句                                                   |
| ------------------------- | ------------------------------------------------------------ |
| 各gsts下pod更新成功数量   | sum(bkbcs_gamestatefulset_pod_update_duration_seconds_count{status="success"}) by (namespace,name) |
| 各gsts下pod更新失败数量   | sum(bkbcs_gamestatefulset_pod_update_duration_seconds_count{status="failure"}) by (namespace,name) |
| 所有gsts下pod更新成功数量 | sum(bkbcs_gamestatefulset_pod_update_duration_seconds_count{status="success"}) |
| 所有gsts下pod更新失败数量 | sum(bkbcs_gamestatefulset_pod_update_duration_seconds_count{status="failure"}) |
| pod更新成功耗时分布       | sum(bkbcs_gamestatefulset_pod_update_duration_seconds_bucket{status="success"}) by (le) |
| pod更新失败耗时分布       | sum(bkbcs_gamestatefulset_pod_update_duration_seconds_bucket{status="failure"}) by (le) |
| 单个pod更新最大耗时       | max(bkbcs_gamestatefulset_pod_update_duration_seconds_max) by (status) |
| 单个pod更新最小耗时       | min(bkbcs_gamestatefulset_pod_update_duration_seconds_min) by (status) |

