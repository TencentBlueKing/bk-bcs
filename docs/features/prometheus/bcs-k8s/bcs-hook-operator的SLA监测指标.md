# bcs-hook-operator的SLA监测指标

指标详情可参考prom指标定义文件：

bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/controllers/hook/metrics.go

## 指标聚合

### 执行HookRun的成功率

```
# 执行HookRun的成功率>99%
(sum(bkbcs_hook_hookrun_exec_duration_seconds_count{action="executeHookRun"}) - sum(sum(bkbcs_hook_hookrun_exec_duration_seconds_count{action="executeHookRun",status="error"} >= bool 1)  or vector(0)))/sum(bkbcs_hook_hookrun_exec_duration_seconds_count{action="executeHookRun"})
```

