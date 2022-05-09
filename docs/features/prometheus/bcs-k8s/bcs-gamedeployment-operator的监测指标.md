

# bcs-gamedeployment-operator  prom指标

指标详情可参考prom指标定义文件：

bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/metrics/metrics.go





# bcs-gamedeployment-operator prom指标监测场景



### gd controller调协情况

| 监测场景名称              | PromQL语句                                                   |
| ------------------------- | ------------------------------------------------------------ |
| 调协成功总次数            | sum(bkbcs_gamedeployment_reconcile_duration_seconds_count{status="success"}) |
| 调协失败总次数            | sum(bkbcs_gamedeployment_reconcile_duration_seconds_count{status="failure"}) |
| 调协成功耗时(s)分布       | sum(bkbcs_gamedeployment_reconcile_duration_seconds_bucket{status="success"}) by(le) |
| 调协失败耗时(s)分布       | sum(bkbcs_gamedeployment_reconcile_duration_seconds_bucket{status="failure"}) by(le) |
| 调协成功耗时top10的gd资源 | topk(10, sum(bkbcs_gamedeployment_reconcile_duration_seconds_sum{status="success"}) by(gd)/sum(bkbcs_gamedeployment_reconcile_duration_seconds_count{status="success"}) by(gd)) |
| 调协失败耗时top10的gd资源 | topk(10, sum(bkbcs_gamedeployment_reconcile_duration_seconds_sum{status="failure"}) by(gd)/sum(bkbcs_gamedeployment_reconcile_duration_seconds_count{status="failure"}) by(gd)) |
| 各gd调协成功次数          | sum(bkbcs_gamedeployment_reconcile_duration_seconds_count{status="success"}) by(gd) |
| 各gd调协失败次数          | sum(bkbcs_gamedeployment_reconcile_duration_seconds_count{status="failure"}) by(gd) |





### 副本情况

| 监测场景名称               | PromQL语句                                                   |
| -------------------------- | ------------------------------------------------------------ |
| 各状态下的副本总数         | DESIRED: sum(bkbcs_gamedeployment_replicas) READY: sum(bkbcs_gamedeployment_ready_replicas) AVAILABLE: sum(bkbcs_gamedeployment_available_replicas) UPDATED: sum(bkbcs_gamedeployment_updated_replicas) UPDATED_READY: sum(bkbcs_gamedeployment_updated_ready_replicas) |
| DESIRED副本数top10的gd资源 | topk(10,sum(bkbcs_gamedeployment_replicas) by(gd))           |
| UNREADY副本数top10的gd资源 | topk(10,sum(abs(bkbcs_gamedeployment_replicas-bkbcs_gamedeployment_ready_replicas)) by(gd)) |





### pod创建情况

| 监测场景名称                 | PromQL语句                                                   |
| ---------------------------- | ------------------------------------------------------------ |
| pod创建成功总次数            | sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{status="success"}) |
| pod创建失败总次数            | sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{status="failure"}) |
| pod创建成功耗时(s)分布       | sum(bkbcs_gamedeployment_pod_create_duration_seconds_bucket{status="success"}) by(le) |
| pod创建失败耗时(s)分布       | sum(bkbcs_gamedeployment_pod_create_duration_seconds_bucket{status="failure"}) by(le) |
| pod创建成功耗时top10的gd资源 | topk(10, sum(bkbcs_gamedeployment_pod_create_duration_seconds_sum{status="success"}) by(gd)/sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{status="success"}) by(gd)) |
| pod创建失败耗时top10的gd资源 | topk(10, sum(bkbcs_gamedeployment_pod_create_duration_seconds_sum{status="failure"}) by(gd)/sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{status="failure"}) by(gd)) |
| 各gd创建pod成功次数          | sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{status="success"}) by(gd) |
| 各gd创建pod失败次数          | sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{status="failure"}) by(gd) |
| pod创建耗时极值情况          | {{status}}_max: max(bkbcs_gamedeployment_pod_create_duration_seconds_max) by(status) {{status}}_min: min(bkbcs_gamedeployment_pod_create_duration_seconds_min) by(status) |





### pod删除情况

| 监测场景名称                 | PromQL语句                                                   |
| ---------------------------- | ------------------------------------------------------------ |
| pod删除成功总次数            | sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{status="success"}) |
| pod删除失败总次数            | sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{status="failure"}) |
| pod删除成功耗时(s)分布       | sum(bkbcs_gamedeployment_pod_delete_duration_seconds_bucket{status="success"}) by(le) |
| pod删除失败耗时(s)分布       | sum(bkbcs_gamedeployment_pod_delete_duration_seconds_bucket{status="failure"}) by(le) |
| pod删除成功耗时top10的gd资源 | topk(10, sum(bkbcs_gamedeployment_pod_delete_duration_seconds_sum{status="success"}) by(gd)/sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{status="success"}) by(gd)) |
| pod删除失败耗时top10的gd资源 | topk(10, sum(bkbcs_gamedeployment_pod_delete_duration_seconds_sum{status="failure"}) by(gd)/sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{status="failure"}) by(gd)) |
| 各gd删除pod成功次数          | sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{status="success"}) by(gd) |
| 各gd删除pod失败次数          | sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{status="failure"}) by(gd) |
| pod删除耗时极值情况          | {{status}}_max: max(bkbcs_gamedeployment_pod_delete_duration_seconds_max) by(status) {{status}}_min: min(bkbcs_gamedeployment_pod_delete_duration_seconds_min) by(status) |





### pod更新情况

| 监测场景名称                 | PromQL语句                                                   |
| ---------------------------- | ------------------------------------------------------------ |
| pod更新成功总次数            | sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{status="success"}) |
| pod更新失败总次数            | sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{status="failure"}) |
| pod更新成功耗时(s)分布       | sum(bkbcs_gamedeployment_pod_update_duration_seconds_bucket{status="success"}) by(le) |
| pod更新失败耗时(s)分布       | sum(bkbcs_gamedeployment_pod_update_duration_seconds_bucket{status="failure"}) by(le) |
| pod更新成功耗时top10的gd资源 | topk(10, sum(bkbcs_gamedeployment_pod_update_duration_seconds_sum{status="success"}) by(gd)/sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{status="success"}) by(gd)) |
| pod更新失败耗时top10的gd资源 | topk(10, sum(bkbcs_gamedeployment_pod_update_duration_seconds_sum{status="failure"}) by(gd)/sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{status="failure"}) by(gd)) |
| 各gd更新pod成功次数          | sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{status="success"}) by(gd) |
| 各gd更新pod失败次数          | sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{status="failure"}) by(gd) |
| pod更新耗时极值情况          | {{status}}_max: max(bkbcs_gamedeployment_pod_update_duration_seconds_max) by(status) {{status}}_min: min(bkbcs_gamedeployment_pod_update_duration_seconds_min) by(status) |

