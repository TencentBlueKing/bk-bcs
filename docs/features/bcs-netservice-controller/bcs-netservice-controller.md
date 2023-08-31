# bcs-netservice-controller

## 背景

原bcs-netservice部署复杂，多用于Mesos集群，且需要使用ZK作为后端存储。现在通过创建K8S CRD（BCSNetPool、BCSNetIP、BCSNetIPClaim），并使用bcs-netservice-controller进行维护，来替换和优化原有功能。

## 功能

- 同步BCSNetPool
- 根据BCSNetPool的配置创建和同步BCSNetIP
- 同步BCSNetIPClaim，用于申请固定IP
- 提供接口供cni插件申请和删除IP

## 架构设计

![image-20230803153333537](./bcs-netservice-controller.png)

## 数据结构

如前面所述，我们定义了BCSNetPool、BCSNetIP和BCSNetIPClaim三种资源。bcs-netservice-controller会根据用户创建的BCSNetPool实例，自动创建BCSNetIP实例，用户不会直接创建BCSNetIP资源。当用户申请固定IP时，可以创建BCSNetIPClaim来实现。

### BCSNetPool

表明网络池的基本信息和状态。用户使用kubectl创建具体的BCSNetPool实例，实例创建以后BCSNetPool的AvailableIPs不会再随BCSNetIP的改变而发生变动，而是只保存其被创建时的初始信息。

bcs-netservice-controller会监听BCSNetPools的变动（如增加Host或AvailableIP），更新Pool或IP资源。在删除Pool的时候，如果该Pool中有处于Active状态的IP，则不会进行删除Pool。

```
// BCSNetPoolSpec defines the desired state of BCSNetPool
type BCSNetPoolSpec struct {
	// 网段
	Net string `json:"net"`
	// 网段掩码
	Mask int `json:"mask"`
	// 网段网关
	Gateway string `json:"gateway"`
	// 对应主机列表
	Hosts []string `json:"hosts,omitempty"`
	// 可用的IP
	AvailableIPs []string `json:"availableIPs,omitempty"`
}

// BCSNetPoolStatus defines the observed state of BCSNetPool
type BCSNetPoolStatus struct {
	// Initializing --初始化中，Normal --正常
	Phase      string      `json:"phase,omitempty"`
	UpdateTime metav1.Time `json:"updateTime,omitempty"`
}
```

### BCSNetIP

表明某个IP的基本信息和状态。cni插件通过bcs-netservice-controller暴露的接口进行IP的申请和删除。

bcs-netservice-controller会根据cni插件请求中传递的podNamespace、podName获取集群Pod的详细信息，从而可以知道Pod是否使用了固定IP，最后会更新BCSNetIPStatus中的Fixed、PodPrimaryKey、PodNamespace、PodName等字段。

```
// BCSNetIPSpec defines the desired state of BCSNetIP
type BCSNetIPSpec struct {
	// 所属网段
	Net string `json:"net"`
	// 网段掩码
	Mask int `json:"mask"`
	// 网段网关
	Gateway string `json:"gateway"`
}

// BCSNetIPStatus defines the observed state of BCSNetIP
type BCSNetIPStatus struct {
	// Active --已使用，Available --可用, Reserved --保留
	Phase string `json:"phase,omitempty"`
	// 对应主机信息
	Host string `json:"host,omitempty"`
	// 是否被用作固定IP
	Fixed bool `json:"fixed,omitempty"`
	// 容器ID
	ContainerID string `json:"containerID,omitempty"`
	// BCSNetIPClaim信息，格式为"命名空间/名称"
	IPClaimKey   string      `json:"ipClaimKey,omitempty"`
	PodName      string      `json:"podName,omitempty"`
	PodNamespace string      `json:"podNamespace,omitempty"`
	UpdateTime   metav1.Time `json:"updateTime,omitempty"`
	KeepDuration string      `json:"keepDuration,omitempty"`
}
```

### BCSNetIPClaim

用于为Pod申请固定IP时，绑定BCSNetIP实例。可以在创建BCSNetIPClaim时通过spec.bcsNetIPName指定想要绑定的IP地址，如果不指定，controller会自动绑定一个可用的BCSNetIP。当Pod被删除后，如果BCSNetIPClaim设置了spec.expiredDuration，则会为Pod保留使用的IP，直到超过了spec.expiredDuration设置的时间。如果不设置超时时间，默认永久保留IP。

```
// BCSNetIPClaimSpec defines the desired state of BCSNetIPClaim
type BCSNetIPClaimSpec struct {
	// BCSNetIPName sets the name for BCSNetIP will be bounded with this claim
	BCSNetIPName string `json:"bcsNetIPName,omitempty"`
	// ExpiredDuration defines expired duration for this claim after claimed IP is released
	ExpiredDuration string `json:"expiredDuration,omitempty"`
}

// BCSNetIPClaimStatus defines the observed state of BCSNetIPClaim
type BCSNetIPClaimStatus struct {
	// BCSNetIPName is name for BCSNetIP bounded with this claim
	BoundedIP string `json:"boundedIP"`
	// Phase represents the state of this claim
	Phase string `json:"phase,omitempty"`
}
```

## 接口

### 申请IP地址

Method：POST

URL： http://localhost:8090/netservicecontroller/v1/allocator

Body：

```
{
	"host":"10.xx.xx.11",
	"containerID":"xxx",
	"ipAddr":"10.xx.xx.21",
	"podName":"xxx",
	"podNamespace":"xxx"
}
```

请求说明:

**host**: 在哪台主机上申请IP地址

**containerID**: 容器id

**ipAddr**: 指定申请IP的地址（可选）

**podName**：容器所属pod名称

**podNamespace**：容器所属pod命名空间

Response:

```
{
    "code": 0,
    "message": "xxx",
    "result": true,
    "data": [],
    "request_id": "xxx"
}
```

### 释放IP地址

Method：DELETE

URL： http://localhost:8090/netservicecontroller/v1/allocator

Body：

```
{
    "host":"10.xx.xx.11",
    "containerID":"xxx",
    "podName":"xxx",
    "podNamespace":"xxx"
}
```

Response:

```
{
    "code": 0,
    "message": "xxx",
    "result": true,
    "data": [],
    "request_id": "xxx"
}
```

## 部署及使用

推荐使用helm部署：

离线helm包：https://github.com/TencentBlueKing/bk-bcs/tree/master/install/helm/bcs-netservice-controller

### 场景一：创建BCSNetPool及BCSNetIP

1. 创建BCSNetPool

   ```
   kubectl apply -f bcsnetpool.yaml
   ```

   bcsnetpool.yaml参考

   ```
   apiVersion: networkextension.bkbcs.tencent.com/v1
   kind: BCSNetPool
   metadata:
     name: 10.xx.xx.0
   spec:
     net: 10.xx.xx.0
     mask: 24
     gateway: 10.xx.xx.1
     hosts:
     - 10.xx.xx.11
     - 10.xx.xx.12
     availableIPs:
     - 10.xx.xx.20
     - 10.xx.xx.21
     - 10.xx.xx.22
   ```

2. 查看BCSNetPool和BCSNetIP

   ```
   kubectl get bcsnetpools
   kubectl get bcsnetips
   ```

### 场景二：Pod申请固定IP

1. 创建BCSNetIPClaim

   ```
   kubectl apply -f bcsnetipclaim.yaml
   ```

   bcsnetipclaim.yaml参考，其中命名空间bcs-demo需要与Pod的命名空间相同

   ```
   apiVersion: netservice.bkbcs.tencent.com/v1
   kind: BCSNetIPClaim
     name: bcsnetipclaim-sample
     namespace: bcs-demo
   spec:
     expiredDuration: 10h
   ```

2. 查看BCSNetIPClaim状态

   ```
   kubectl get bcsnetipclaims -n bcs-demo bcsnetipclaim-sample -oyaml
   ```

   确保claim处于Bound状态，实例

   ```
   apiVersion: netservice.bkbcs.tencent.com/v1
   kind: BCSNetIPClaim
   metadata:
     finalizers:
     - netservicecontroller.bkbcs.tencent.com
     name: bcsnetipclaim-sample
     namespace: default
   spec:
     expiredDuration: 10h
   status:
     boundedIP: 10.xx.xx.20
     phase: Bound
   ```

3. 在Deployment/StatefulSet的Pod模版spec.template.metadata.annotations中添加申请固定IP的annotation

   ```
   netservicecontroller.bkbcs.tencent.com/ipclaim: <claimName>
   ```

4. Pod创建完成后，查看其IP是否与claim绑定的IP地址一致