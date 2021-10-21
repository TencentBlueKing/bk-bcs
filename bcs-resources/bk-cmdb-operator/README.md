## bk-cmdb-operator 使用指南

### 构建镜像

```
# make docker-build
……
……
Successfully built ec9d56eb5d3f
Successfully tagged bk-cmdb-operator:test
```

### 创建crd 

```
# kubectl apply -f config/crd/bases/bkcmdb.bkbcs.tencent.com_bkcmdbs.yaml
```

### 部署 bk-cmdb-operator

```
# kubectl apply -f config/install/operator-deploy.yaml
```

### 拉起 bk-cmdb
```
# kubectl apply -f config/install/sample.yaml
```

### 确认部署状态
```
# kubectl get pods
NAME                                             READY   STATUS      RESTARTS   AGE
bk-cmdb-operator-5758f65b99-84sz8                1/1     Running     0          19h
bkcmdb-sample-adminserver-7974fd658f-wcnhk       1/1     Running     0          19h
bkcmdb-sample-apiserver-5fb54c7647-6lj86         1/1     Running     0          19h
bkcmdb-sample-bootstrap-7jpdc                    0/1     Completed   3          19h
bkcmdb-sample-coreservice-c786dd449-6tjrh        1/1     Running     0          19h
bkcmdb-sample-datacollection-57fc4b4c7f-trnwb    1/1     Running     1          19h
bkcmdb-sample-eventserver-5dd878db79-9h6hn       1/1     Running     0          19h
bkcmdb-sample-hostserver-68f87c88db-chwwr        1/1     Running     0          19h
bkcmdb-sample-mongodb-75dddc8f54-vc274           1/1     Running     0          19h
bkcmdb-sample-operationserver-6b47c7f9c6-jfzhz   1/1     Running     0          19h
bkcmdb-sample-procserver-7d6c64c485-q279r        1/1     Running     0          19h
bkcmdb-sample-redis-master-0                     1/1     Running     0          19h
bkcmdb-sample-redis-slave-699cb76c68-jqvtb       1/1     Running     0          19h
bkcmdb-sample-taskserver-5d56d98bf8-vw4vq        1/1     Running     0          19h
bkcmdb-sample-tmserver-78dcc59cf5-s8zxw          1/1     Running     0          19h
bkcmdb-sample-toposerver-674dc44485-6w5zc        1/1     Running     0          19h
bkcmdb-sample-webserver-7b7fd9b4d4-q24x2         1/1     Running     0          19h
bkcmdb-sample-zookeeper-0                        1/1     Running     0          19h
```