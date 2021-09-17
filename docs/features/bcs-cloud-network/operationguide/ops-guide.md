# bcs-cloud-netservice operation guide

## 1. 使用方法

### 1.1 子网操作

#### 添加子网

```shell
curl -X POST 127.0.0.1:8080/v1/subnet -H "accept: application/json" \
  -d '{
  "seq": "1",
  "vpcID": "vpc-xxx",
  "region": "ap-xxx",
  "zone": "ap-xxx-1",
  "subnetID": "subnet-xxxx",
  "subnetCidr": "127.0.0.1/22"
}'
```

#### 开启关闭子网

开启子网，使其可被分配

```shell
curl -X POST 127.0.0.1:8080/v1/subnet/op -H "accept: application/json" \
  -d '{
  "seq": "1",
  "vpcID": "vpc-xxx",
  "region": "ap-xxx",
  "subnetID": "subnet-xxx",
  "state": 1
  }
```

关闭子网，使其不可被分配

```shell
curl -X POST 127.0.0.1:8080/v1/subnet/op -H "accept: application/json" \
  -d '{
  "seq": "1",
  "vpcID": "vpc-xxx",
  "region": "ap-xxx",
  "subnetID": "subnet-xxx",
  "state": 0
  }
```

#### 删除子网

```shell
curl -X DELETE "http://127.0.0.1:8080/v1/subnet?seq=1&vpcID=vpc-xxx&region=ap-xxx&zone=ap-xxx-1&subnetID=subnet-xxx" -H "accept: application/json"
```

注意事项：

* state处于1的子网不可被删除
* 子网中有固定IP地址，active状态IP地址，available状态IP地址和eniprimary状态IP地址，都不可被删除

### 1.2 节点弹性网卡操作

#### 增加弹性网卡

* **nodenetwork.bkbcs.tencent.com: true**表示增加弹性网卡，且默认添加一个弹性网卡
* **eninumber.nodenetwork.bkbcs.tencent.com: 2**表示设置该节点弹性网卡数量为2

```shell
kubectl label node {{ nodeName }} nodenetwork.bkbcs.tencent.com=true
kubectl label node {{ nodeName }} eninumber.nodenetwork.bkbcs.tencent.com=2
```

#### 释放弹性网卡

将节点的label中nodenetwork.bkbcs.tencent.com去掉，则会进行弹性网卡释放操作

```shell
kubectl label node {{ nodeName }} nodenetwork.bkbcs.tencent.com-
```

* 弹性网卡会按照序号从大到小释放
* 如果弹性网卡上存在active状态的IP地址，则不会进行释放操作，等待IP地址被释放之后，才会执行弹性网卡释放过程
* 如果集群节点已经被删除，手动执行kubectl delete nodenetwork {{ nodeName }}，会强制执行弹性网卡释放过程，确保CVM权限归还之前，能够执行delete nodenetwork操作，否则弹性网卡可能游离

### 容器网络设置

#### 申请普通弹性网卡IP地址

* 在pod template的annotations中填入**tke.cloud.tencent.com/networks: bcs-eni-cni**（针对TKE集群）
* 在第一个容器的resources中填入requests和limits**cloud.bkbcs.tencent.com/eip: 1**，相关scheduler extender需要的资源类型来进行调度，如果没有填入，则不会交由相关sheduler extender处理

```yaml
spec:
  template:
    metadata:
      annotations:
        tke.cloud.tencent.com/networks: bcs-eni-cni
      spec:
        containers:
        resources:
          requests:
            cloud.bkbcs.tencent.com/eip: 1
          limits:
            cloud.bkbcs.tencent.com/eip: 1
```

#### 申请固定IP地址

* 在pod template的annotations中填入**tke.cloud.tencent.com/networks: bcs-eni-cni**和**eni.cloud.bkbcs.tencent.com: fixed**
* **keepduration.eni.cloud.bkbcs.tencent.com: 12h**，表示该固定IP在释放后的保留时间，合法的单位为[m, h]，最大不超过500h
* 在第一个容器的resources中填入requests和limits**cloud.bkbcs.tencent.com/eip: 1**，相关scheduler extender需要的资源类型来进行调度，如果没有填入，则不会交由相关sheduler extender处理

```yaml
spec:
  template:
    metadata:
      annotations:
        tke.cloud.tencent.com/networks: bcs-eni-cni
        eni.cloud.bkbcs.tencent.com: fixed
        keepduration.eni.cloud.bkbcs.tencent.com: 24h
      spec:
        containers:
        resources:
          requests:
            cloud.bkbcs.tencent.com/eip: 1
          limits:
            cloud.bkbcs.tencent.com/eip: 1
```

注意事项

* 如果不填**keepduration.eni.cloud.bkbcs.tencent.com**，则默认为48h
* 固定IP的申请只适用于StatefulSet和GameStatefulSet，其它Workload使用则会失败

## 2. 发生IP地址泄漏怎么办？

* cloud-netagent在重启之后，会执行泄漏IP检测，并且释放掉泄漏的active状态的IP地址
