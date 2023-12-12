# BCS-Cluster-Reporter
> 蓝鲸容器服务（BCS）集群巡检工具，检测集群运行状态并生成巡检指标以供监控采集使用

## 指标


### systemappcheck
system_app_chart_version 系统应用的chart部署情况
```
system_app_chart_version{chart="bcs-kube-agent",namespace="kube-system",status="deployed",target="BCS-K8S-XXXXXX",target_biz="XXXXXX",version="v1.27.0"} 1
```

system_app_image_version 系统应用的workload部署情况
```
system_app_image_version{chart="bcs-kube-agent",component="bcs-kube-agent",container="bcs-kube-agent",namespace="kube-system",resource="Deployment",status="ready",target="BCS-K8S-XXXXXX",target_biz="XXXXXX",version="xxxxxx:xxx"} 1
```

### clustercheck
cluster_availability 集群可用性检测情况
```
cluster_availability{status="访问集群失败",target="BCS-K8S-XXXXX",target_biz="XXXXX"} 1
cluster_availability{status="ok",target="BCS-K8S-YYYYY",target_biz="YYYYY"} 1
```




