# bcs-gamestatefulset-operator的SLA prom指标

指标详情可参考prom指标定义文件：

bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/controllers/metrics.go和bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/metrics/metrics.go

## 指标聚合

### 创建/删除GameStatefulSet

#### GamestatefulSet变更成功率

```
# GamestatefulSet变更成功率
sum(bkbcs_gamestatefulset_pod_create_duration_seconds_count{status="success"} or bkbcs_gamestatefulset_pod_delete_duration_seconds_count{action="deletePod",status="success"})/sum(bkbcs_gamestatefulset_pod_create_duration_seconds_count{} or  bkbcs_gamestatefulset_pod_delete_duration_seconds_count{action="deletePod"})

# 创建Pod成功率
sum(bkbcs_gamestatefulset_pod_create_duration_seconds_count{status="success"})/sum(bkbcs_gamestatefulset_pod_create_duration_seconds_count{})

# 删除Pod成功率
sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{action="deletePod",status="success"})/sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{action="deletePod"})
```

#### GamestatefulSet变更生效时间

```
# GamestatefulSet变更生效时间<10s
sum(bkbcs_gamestatefulset_pod_create_duration_seconds_bucket{status="success",le="10"} or bkbcs_gamestatefulset_pod_delete_duration_seconds_bucket{action="deletePod",status="success",le="10"}) /sum(bkbcs_gamestatefulset_pod_create_duration_seconds_count{} or bkbcs_gamestatefulset_pod_delete_duration_seconds_count{action="deletePod"})

# 创建pod生效时间<10s
sum(bkbcs_gamestatefulset_pod_create_duration_seconds_bucket{status="success",le="10"})/sum(bkbcs_gamestatefulset_pod_create_duration_seconds_count{})

# 删除pod生效时间<10s
sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_bucket{action="deletePod",status="success",le="10"})/sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{action="deletePod",status="success"})
```

### 原地更新

#### 更新对象的成功率

```
# 原地更新的成功率
sum(bkbcs_gamestatefulset_pod_update_duration_seconds_count{status="success",action="inplaceUpdate"})/sum(bkbcs_gamestatefulset_pod_update_duration_seconds_count{action="inplaceUpdate"})
```

#### 更新对象的生效时间

```
# 更新对象的生效时间<10s
sum(bkbcs_gamestatefulset_pod_update_duration_seconds_bucket{status="success",action="inplaceUpdate",le="10"})/sum(bkbcs_gamestatefulset_pod_update_duration_seconds_count{status="success",action="inplaceUpdate"})
```

### Pod优雅删除/更新

#### 优雅删除或更新的成功率

```
# 优雅删除或更新的成功率
sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{action="rollingUpdate",grace="true",status="success"} or bkbcs_gamestatefulset_pod_update_duration_seconds_count{action="inplaceUpdate",grace="true",status="success"} or bkbcs_gamestatefulset_hookrun_create_duration_seconds_count{status="success",objectKind="GameStatefulSet"})/ sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{action="rollingUpdate",grace="true"} or bkbcs_gamestatefulset_pod_update_duration_seconds_count{action="inplaceUpdate",grace="true"} or bkbcs_gamestatefulset_hookrun_create_duration_seconds_count{objectKind="GameStatefulSet"})

# hookrun创建成功率
sum(bkbcs_gameworkload_hookrun_create_duration_seconds_count{status="success",action=~"predelete|preinplace",objectKind="GameStatefulSet"})/ sum(bkbcs_gamestatefulset_hookrun_create_duration_seconds_count{objectKind="GameStatefulSet"})

# 优雅删除pod的成功率
sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{grace="true",status="success"})/ sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{grace="true"})

# 优雅更新容器重建的成功率
sum(bkbcs_gamestatefulset_pod_update_duration_seconds_count{grace="true",status="success"})/sum(bkbcs_gamestatefulset_pod_update_duration_seconds_count{grace="true"})
```

#### 优雅删除或更新的延迟

```
# 优雅删除或更新的延迟<10s
sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_bucket{grace="true",status="success",le="10"} or bkbcs_gamestatefulset_pod_update_duration_seconds_bucket{action="inplaceUpdate",grace="true",status="success",le="10"} or bkbcs_gamestatefulset_hookrun_create_duration_seconds_bucket{action=~"preinplace|predelete",status="success",le="10",objectKind="GameStatefulSet"})/ sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{grace="true"} or bkbcs_gamestatefulset_pod_update_duration_seconds_count{action="inplaceUpdate",grace="true"} or bkbcs_gamestatefulset_hookrun_create_duration_seconds_count{action=~"preinplace|predelete",objectKind="GameStatefulSet"})

# hookrun创建的延迟<10s
sum(bkbcs_gameworkload_hookrun_create_duration_seconds_bucket{action=~"preinplace|predelete",status="success",le="10",objectKind="GameStatefulSet"})/sum(bkbcs_gamestatefulset_hookrun_create_duration_seconds_count{action=~"preinplace|predelete",objectKind="GameStatefulSet"})

# 删除pod的延迟<10s
sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_bucket{grace="true",status="success",le="10"})/sum(bkbcs_gamestatefulset_pod_delete_duration_seconds_count{grace="true"})

# 容器重建的延迟<10s
sum(bkbcs_gamestatefulset_pod_update_duration_seconds_bucket{action="inplaceUpdate",grace="true",status="success",le="10"})/sum(bkbcs_gamestatefulset_pod_update_duration_seconds_count{action="inplaceUpdate",grace="true"})
```
