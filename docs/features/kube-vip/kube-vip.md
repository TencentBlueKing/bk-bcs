# kube-vip

Kube-Vip最初是为Kubernetes控制平面提供高可用解决方案而创建的，随着时间的推移，它已经发展成将同样的功能纳入Kubernetes的LoadBalancer类型的服务中。

## 架构

kube-vip项目旨在为底层网络服务提供高可用的网络端点和负载均衡功能。

### Cluster

kube-vip服务构建一个多节点或多Pod集群以提供高可用性。在ARP模式下，选举出一个领导者，该领导者将继承虚拟IP并成为集群内负载平衡的领导者，而在BGP模式下，所有节点将广播VIP地址。

当使用ARP或layer2时，它将使用领导者选举。

### 虚拟IP

在群集内部，领导者将被分配VIP，并将其绑定到在配置中声明的网卡接口。当领导者更改时，它将首先撤销VIP，或者在故障场景中，VIP将直接分配给下一个选举的领导者。

当VIP从一个主机移动到另一个主机时，使用VIP的任何主机都将保留先前的VIP到MAC地址映射，直到旧的ARP记录过期（通常在30秒内）并检索新的映射。启用Gratuitous ARP广播可以改善此情况。

### ARP

kube-vip可以选择配置广播 Gratuitous ARP，通常会立即通知所有本地主机 VIP-to-MAC 地址映射已更改。

## 功能
- VIP 地址可以是 IPv4 或 IPv6
- 支持使用ARP（第二层）或BGP（第三层）的控制平面
- 支持使用领导选举或raft的控制平面
- 支持使用kubeadm（静态 Pod）的控制平面 HA
- 支持使用K3s/和其他（DaemonSets）的控制平面 HA
- 支持使用ARP领导者选举的Service LoadBalancer（第 2 层）
- 支持使用BGP的多个节点的Service LoadBalancer
- 按命名空间或全局的Service LoadBalancer地址池
- 地址通过UPNP暴露给网关的Service LoadBalancer

## 部署

这里使用static pod部署,注意要在所有master节点上配置

### ARP模式

```
mkdir -p /etc/kubernetes/manifests/
# 配置vip地址
export VIP=192.168.xx.xx
# 设置网卡名称
export INTERFACE=ens192
ctr image pull docker.io/plndr/kube-vip:v0.5.12
# 静态Pod资源yaml
ctr run --rm --net-host docker.io/plndr/kube-vip:v0.5.12 vip \
/kube-vip manifest pod \
--interface $INTERFACE \
--vip $VIP \
--controlplane \
--services \
--arp \
--leaderElection | tee  /etc/kubernetes/manifests/kube-vip.yaml
```

### BGP模式

```
mkdir -p /etc/kubernetes/manifests/
# 配置vip地址
export VIP=192.168.xx.xx
# 设置网卡名称
export INTERFACE=lo
ctr run --rm --net-host docker.io/plndr/kube-vip:v0.5.12 vip \
/kube-vip manifest pod \
    --interface $INTERFACE \
    --address $VIP \
    --controlplane \
    --services \
    --bgp \
    --localAS 65000 \
    --bgpRouterID 192.168.xx.xx \
    --bgppeers 192.168.xx.xx:65000::false,192.168.xx.xx:65000::false | tee /etc/kubernetes/manifests/kube-vip.yaml
```

## 总结

1. kube-vip只部署在control-plane上

2. 如果考虑集群外访问，需要一个与apiserver相同子网的可用IP

3. 如果只是集群内使用，可以自定义IP，但是要在所有节点上创建可以访问该IP的规则route
