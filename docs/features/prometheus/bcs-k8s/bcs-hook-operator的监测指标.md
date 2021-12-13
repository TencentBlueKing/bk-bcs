

# bcs-hook-operator  prom指标

指标详情可参考prom指标定义文件：

bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/controllers/hook/metrics.go



# bcs-hook-operator prom指标监测场景



### hook controller调协情况

| 监测场景名称                | PromQL语句                                                   |
| --------------------------- | ------------------------------------------------------------ |
| 调协成功总次数              | sum(bkbcs_hook_reconcile_duration_seconds_count{status="success"}) |
| 调协失败总次数              | sum(bkbcs_hook_reconcile_duration_seconds_count{status="failure"}) |
| 调协成功耗时(s)分布         | sum(bkbcs_hook_reconcile_duration_seconds_bucket{status="success"}) by(le) |
| 调协失败耗时(s)分布         | sum(bkbcs_hook_reconcile_duration_seconds_bucket{status="failure"}) by(le) |
| 调协成功耗时top10的ownerRef | topk(10, sum(bkbcs_hook_reconcile_duration_seconds_sum{status="success"}) by(namespace, owner)/sum(bkbcs_hook_reconcile_duration_seconds_count{status="success"}) by(namespace, owner)) |
| 调协失败耗时top10的ownerRef | topk(10, sum(bkbcs_hook_reconcile_duration_seconds_sum{status="failure"}) by(namespace, owner)/sum(bkbcs_hook_reconcile_duration_seconds_count{status="failure"}) by(namespace, owner)) |
| 各ownerRef调协成功次数      | sum(bkbcs_hook_reconcile_duration_seconds_count{status="success"}) by(namespace, owner) |
| 各ownerRef调协失败次数      | sum(bkbcs_hook_reconcile_duration_seconds_count{status="failure"}) by(namespace, owner) |





### hookrun执行情况

| 监测场景名称                       | PromQL语句                                                   |
| ---------------------------------- | ------------------------------------------------------------ |
| hookrun执行成功总次数              | sum(bkbcs_hook_hookrun_exec_duration_seconds_count{status="success"}) |
| hookrun执行失败总次数              | sum(bkbcs_hook_hookrun_exec_duration_seconds_count{status="failure"}) |
| hookrun执行成功耗时(s)分布         | sum(bkbcs_hook_hookrun_exec_duration_seconds_bucket{status="success"}) by(le) |
| hookrun执行失败耗时(s)分布         | sum(bkbcs_hook_hookrun_exec_duration_seconds_bucket{status="failure"}) by(le) |
| hookrun执行成功耗时top10的ownerRef | topk(10, sum(bkbcs_hook_hookrun_exec_duration_seconds_sum{status="success"}) by(namespace, owner)/sum(bkbcs_hook_hookrun_exec_duration_seconds_count{status="success"}) by(namespace, owner)) |
| hookrun执行失败耗时top10的ownerRef | topk(10, sum(bkbcs_hook_hookrun_exec_duration_seconds_sum{status="failure"}) by(namespace, owner)/sum(bkbcs_hook_hookrun_exec_duration_seconds_count{status="failure"}) by(namespace, owner)) |
| 各ownerRef调协成功次数             | sum(bkbcs_hook_hookrun_exec_duration_seconds_count{status="success"}) by(namespace, owner) |
| 各ownerRef调协失败次数             | sum(bkbcs_hook_hookrun_exec_duration_seconds_count{status="failure"}) by(namespace, owner) |
| 正在运行的hookrun存活时间          | bkbcs_hook_hookrun_survive_time_seconds{phase="Running"}     |
| hookrun执行耗时极值情况            | {{status}}_max: max(bkbcs_hook_hookrun_exec_duration_seconds_max) by(status) {{status}}_min: min(bkbcs_hook_hookrun_exec_duration_seconds_max) by(status) |





### hookrun下的metric任务执行情况

| 监测场景名称                                   | PromQL语句                                                   |
| ---------------------------------------------- | ------------------------------------------------------------ |
| hookrun下的metric任务执行成功总次数            | sum(bkbcs_hook_metric_exec_duration_seconds_count{phase="Successful"}) |
| hookrun下的metric任务执行失败总次数            | sum(bkbcs_hook_metric_exec_duration_seconds_count{phase=~"Error\|Failed"} ) |
| hookrun下的metric任务执行成功耗时(s)分布       | sum(bkbcs_hook_metric_exec_duration_seconds_bucket{phase="Successful"}) by(le) |
| hookrun下的metric任务执行失败耗时(s)分布       | sum(bkbcs_hook_metric_exec_duration_seconds_bucket{phase=~"Error\|Failed"}) by(le) |
| hookrun下的metric任务执行成功耗时top10的metric | topk(10, sum(bkbcs_hook_metric_exec_duration_seconds_sum{phase="Successful"}) by(namespace, owner, metric)/sum(bkbcs_hook_metric_exec_duration_seconds_count{phase="Successful"}) by(namespace, owner, metric)) |
| hookrun下的metric任务执行失败耗时top10的metric | topk(10, sum(bkbcs_hook_metric_exec_duration_seconds_sum{phase=~"Error\|Failed"}) by(namespace, owner, metric)/sum(bkbcs_hook_metric_exec_duration_seconds_count{phase=~"Error\|Failed"}) by(namespace, owner, metric)) |
| 各metric执行成功次数                           | sum(bkbcs_hook_metric_exec_duration_seconds_count{phase="Successful"}) by(namespace, owner, metric) |
| 各metric执行失败次数                           | sum(bkbcs_hook_metric_exec_duration_seconds_count{phase=~"Error\|Failed"}) by(namespace, owner, metric) |
| hookrun下的metric任务执行耗时极值情况          | {{phase}}_max: max(bkbcs_hook_metric_exec_duration_seconds_max) by(phase) {{phase}}_min: min(bkbcs_hook_metric_exec_duration_seconds_min) by(phase) |


