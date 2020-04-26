# bcs cloud network agent

## 简介

bcs cloud network agent作为进程或者daemonSet方式常驻于slave节点上，主要作用有：

* 向公有云申请弹性和释放弹性网卡
* 设置虚拟机上的弹性网卡IP地址、创建默认路由以及策略路由
* 设置网卡的IP_FORWARD参数以及RP_FILTER参数
* 对主机网卡设置进行周期性轮训

## 整体架构

![bcs-cloud-network-agent](./bcs-cloud-network-agent.png)

## 参数说明

```json
{
    "cluster": "${bcsCloudNetworkAgentClusterid}",
    "cloud": "bcsCloudNetworkAgentCloud",
    "kubeconfig": "${bcsCloudNetworkAgentKubeconfig}",
    "netserviceZookeeper": "${bcsCloudNetworkAgentNetserviceZookeeper}",
    "netserviceCa": "${bcsCloudNetworkAgentNetserviceCaFile}",
    "netserviceKey": "${bcsCloudNetworkAgentNetserviceClientKeyFile}",
    "netserviceCert": "${bcsCloudNetworkAgentNetserviceClentCertFile}",
    "subnets": "${bcsCloudNetworkAgentSubnets}",
    "eniNum": ${bcsCloudNetworkAgentEniNum},
    "ipNumPerEni": ${bcsCloudNetworkAgentIpNumPerEni},
    "eniMTU": ${bcsCloudNetworkAgentEniMTU},
    "ifaces": "${bcsCloudNetworkAgentIfaces}",
    "v": "${bcsCloudNetworkAgentLogLevel}"
}
```

* bcsCloudNetworkAgentClusterid: BCS集群ID
* bcsCloudNetworkAgentCloud: 公有云类型
* bcsCloudNetworkAgentKubeconfig: kubeconfig位置
* bcsCloudNetworkAgentNetserviceZookeeper: bcs-netservice zk地址
* bcsCloudNetworkAgentNetserviceCaFile: bcs-netservice tls ca证书
* bcsCloudNetworkAgentNetserviceClientKeyFile: bcs-netservice 客户端私钥
* bcsCloudNetworkAgentNetserviceClentCertFile: bcs-netservice 客户端证书
* bcsCloudNetworkAgentSubnets: 公有云VPC中，弹性网卡IP可用子网ID，逗号分割
* bcsCloudNetworkAgentEniNum: 申请的弹性网卡数量, "0"表示申请尽可能多的网卡
* bcsCloudNetworkAgentIpNumPerEni: 每张弹性网卡IP数量: "0"表示尽可能多的申请IP
* bcsCloudNetworkAgentEniMTU: 弹性的网卡MTU，默认为1500，(AWS上推荐使用9001)
* bcsCloudNetworkAgentIfaces: 虚拟机的网卡，用来获取表示主机身份的IP地址，逗号分割
* bcsCloudNetworkAgentLogLevel: 日志级别

