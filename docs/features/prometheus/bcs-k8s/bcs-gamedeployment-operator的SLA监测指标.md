# bcs-gamedeployment-operator的SLA prom指标

指标详情可参考prom指标定义文件：

bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/controllers/metrics.go和bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/metrics/metrics.go

## 指标聚合

### 创建/删除GameDeployment

#### GameDeployment变更成功率

```
# GameDeployment变更成功率
sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{status="success"} or bkbcs_gamedeployment_pod_delete_duration_seconds_count{action="deletePod",status="success"})/sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{} or  bkbcs_gamedeployment_pod_delete_duration_seconds_count{action="deletePod"})
# 创建Pod成功率
sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{status="success",gd=~"$gd"}) /sum
(bkbcs_gamedeployment_pod_create_duration_seconds_count{gd=~"$gd"}) 
# 删除Pod成功率
sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{action="deletePod",status="success",gd=~"$gd"}) / sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{action="deletePod",gd=~"$gd"}) 
```

#### GameDeployment变更生效时间

```
# GameDeployment变更生效时间<10s
sum(bkbcs_gamedeployment_pod_create_duration_seconds_bucket{status="success",le="10"} or bkbcs_gamedeployment_pod_delete_duration_seconds_bucket{action="deletePod",status="success",le="10"}) /sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{} or bkbcs_gamedeployment_pod_delete_duration_seconds_count{action="deletePod"})
# 创建pod生效时间<10s
sum(bkbcs_gamedeployment_pod_create_duration_seconds_bucket{status="success",gd=~"$gd",le="10"}) /sum(bkbcs_gamedeployment_pod_create_duration_seconds_count{gd=~"$gd"}) 
# 删除pod生效时间<10s
sum(bkbcs_gamedeployment_pod_delete_duration_seconds_bucket{action="deletePod",status="success",gd=~"$gd",le="10"})/sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{action="deletePod",status="success",gd=~"$gd"}) 
```

### 原地更新

#### 更新对象的成功率

```
# 原地更新的成功率
sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{status="success",action="inplaceUpdate"})/sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{action="inplaceUpdate"})
# 容器被kill或create的成功率
(sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{status="success",gd=~"$gd",action="inplaceUpdate"}) /(sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{action="inplaceUpdate"}) 
```

#### 更新对象的生效时间

```
# 更新对象的生效时间<10s
sum(bkbcs_gamedeployment_pod_update_duration_seconds_bucket{status="success",action="inplaceUpdate",le="10"})/sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{status="success",action="inplaceUpdate"})
# 容器被kill或create的延迟<10s
(sum(bkbcs_gamedeployment_pod_update_duration_seconds_bucket{status="success",gd=~"$gd",action="inplaceUpdate",le="10"})/(sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{status="success",gd=~"$gd",action="inplaceUpdate"}) 
```

### Pod优雅删除/更新

#### 优雅删除或更新的成功率

```
# 优雅删除或更新的成功率
sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{action="rollingUpdate",grace="true",status="success"} or bkbcs_gamedeployment_pod_update_duration_seconds_count{action="inplaceUpdate",grace="true",status="success"} or bkbcs_gameworkload_hookrun_create_duration_seconds_count{status="success",objectKind="GameDeployment"})/ sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{action="rollingUpdate",grace="true"} or bkbcs_gamedeployment_pod_update_duration_seconds_count{action="inplaceUpdate",grace="true"} or bkbcs_gameworkload_hookrun_create_duration_seconds_count{objectKind="GameDeployment"})
# hookrun创建成功率
sum(bkbcs_gameworkload_hookrun_create_duration_seconds_count{status="success",gd=~"$gd",action=~"predelete|preinplace",objectKind="GameDeployment"})/ sum(bkbcs_gameworkload_hookrun_create_duration_seconds_count{gd=~"$gd",objectKind="GameDeployment"})
# 优雅删除pod的成功率
sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{grace="true",status="success",gd=~"$gd"}) / sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{grace="true",gd=~"$gd"})
# 优雅更新容器重建的成功率
sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{grace="true",status="success",gd=~"$gd"}) / sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{grace="true",gd=~"$gd"})
```

#### 优雅删除或更新的延迟

```
# 优雅删除或更新的延迟<10s
sum(bkbcs_gamedeployment_pod_delete_duration_seconds_bucket{grace="true",status="success",le="10"} or bkbcs_gamedeployment_pod_update_duration_seconds_bucket{action="inplaceUpdate",grace="true",status="success",le="10"} or bkbcs_gameworkload_hookrun_create_duration_seconds_bucket{action=~"preinplace|predelete",status="success",le="10",objectKind="GameDeployment"})/ sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{grace="true"} or bkbcs_gamedeployment_pod_update_duration_seconds_count{action="inplaceUpdate",grace="true"} or bkbcs_gameworkload_hookrun_create_duration_seconds_count{action=~"preinplace|predelete",objectKind="GameDeployment"})
# hookrun创建的延迟<10s
sum(bkbcs_gameworkload_hookrun_create_duration_seconds_bucket{action=~"preinplace|predelete",status="success",gd=~"$gd",le="10",objectKind="GameDeployment"}) / sum(bkbcs_gameworkload_hookrun_create_duration_seconds_count{action=~"preinplace|predelete",gd=~"$gd",objectKind="GameDeployment"}) 
# 删除pod的延迟<10s
sum(bkbcs_gamedeployment_pod_delete_duration_seconds_bucket{grace="true",status="success",gd=~"$gd",le="10"}) / sum(bkbcs_gamedeployment_pod_delete_duration_seconds_count{grace="true",gd=~"$gd"}) 
# 容器重建的延迟<10s
sum(bkbcs_gamedeployment_pod_update_duration_seconds_bucket{action="inplaceUpdate",grace="true",status="success",gd=~"$gd",le="10"}) / sum(bkbcs_gamedeployment_pod_update_duration_seconds_count{action="inplaceUpdate",grace="true",gd=~"$gd"})
```
