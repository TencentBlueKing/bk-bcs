bcs-k8s-watch
===========================

## TODO

1. for resource (not event), should get a buffer, fow 5 seconds, update once
2. have add health check, but I don't think it is a good idea to use liveness probe in static pod
   if the apiserver down, apiserver make alarm, not via datawatch

## data flow

> datawatch -> filter -> writer -> handler -> action -> cc

## DONE

- 实现list-watch监控k8s原生资源, 并汇总上报到storage(http or https)
- 实现与bcs相关系统交互, zk操作, 获取clusterkeeper/storage地址, 获取clusterID
- 实现资源定期同步逻辑

## TODO

- 高可用部署方案
- 上报 LB 相关数据
- 异常事件上报到bcs-health, 发送告警
- 补全单元测试


## components

- main.go    - enterance
- app/bcs    - access to other bcs components
- app/http   - http client
- app/k8s    - k8s resourcce list-watch
- app/output - output to log or storage, writers and handlers
- app/output/action - specific output action
- app/app.go     - Run, start cluster/writer


## dependency

- Makefile
- mock api: https://github.com/typicode/json-server

- Godep https://github.com/tools/godep / https://devcenter.heroku.com/articles/go-dependencies-via-godep
        add vendor to git too: https://stackoverflow.com/questions/26334220/should-i-commit-godeps-workspace-or-is-godeps-json-enough
- json-iterator https://github.com/json-iterator/go
- gorequest https://github.com/parnurzeal/gorequest https://github.com/parnurzeal/gorequest


- zookeeper https://github.com/paulbrown/docker-zookeeper/blob/master/kube/zookeeper.yaml https://kubernetes.io/docs/tutorials/stateful-application/zookeeper/

## test

https://github.com/astaxie/build-web-application-with-golang/blob/master/zh/11.3.md


