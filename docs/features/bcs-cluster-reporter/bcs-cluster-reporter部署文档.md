# BCS-Cluster-Reporter
> 蓝鲸容器服务（BCS）集群巡检工具，检测集群运行状态并生成巡检指标以供监控采集使用

## 部署


### 集中化部署
1.集中化部署开启的检测模块
```
systemappcheck，clustercheck，masterpodcheck
```

2.关闭单集群部署
```
Values.inCluster: false
```

3.集群信息来自kubeconfig文件则配置，默认不需要配置
```
Values.kubeconfigs
```

4.集群信息来自bcs则配置
```
Values.bcsAPIGateway.host
Values.bcsAPIGateway.bcsCAToken
```

### 单集群部署
1.配置需要开启的检测模块

目前集群内部署推荐开启dnscheck，netcheck，eventrecorder，logrecorder模块
```
Values.plugins
```

2.配置集群信息
```
Values.clusterId
Values.bizId
```

3.如果需要采集集中化指标需要配置
```
Values.centerReporterAddress
```

4.配置开启集群内部署，用来检测当前所在集群
```
Values.inCluster: true
```

