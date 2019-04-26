# Bcs-Container-Executor工作机制信息

bcs-container-executor自研mesos executor，用于对接自研bcs-scheduler，基于mesos协议实现容器任务管理。

## taskgroup特性支持与设置约定

executor支持taskgroup特性（mesos 1.1.0加入），针对taskgroup中多个task，按照Pod的概念进行处理，
一个task对应一个容器，多个容器共享生命周期。默认把第一个TaskInfo作为主Task，所有的其他
TaskInfo默认与第一个共享网络参数参数。

启动行为：

* 如果无法链接Docker daemon，设置为状态失败，FAILED
* 默认先启动第一个TaskInfo，如果启动失败，整个TaskGroup失败（FAILED）
* 第一个TaskInfo启动正常（正常向docker提交容器）即可获取到容器ID，此时不会检测容器是否在运行
* 依次启动后续的TaskInfo,使用第一个容器ID作为依据共享网络信息
  * 如果Attach第一个容器的网络失败（说明第一个容器已经失败），整个TaskGroup失败
  * 如果attach成功，继续启动下一个TaskInfo

如果TaskInfo状态切换为RUNNING，在恢复TaskStatus信息时，默认将容器信息作为json写入Data字段。
字段类型如下：

```text
//BcsContainerInfo only for BcsExecutor
type BcsContainerInfo struct {
ID       string    `json:"ID,omitempty"`       //container ID
Name     string    `json:"Name,omitempty"`     //container name
Pid      int       `json:"Pid,omitempty"`      //container pid
StartAt  time.Time `json:"StartAt,omitempty"`  //startting time
FinishAt time.Time `json:"FinishAt,omitempty"` //Exit time
//Status
Status   string `json:"Status,omitempty"`   //status string, paused, restarting, running, dead, created, exited
ExitCode int    `json:"ExitCode,omitempty"` //container exit code

//Network info
Hostname    string `json:"Hostname,omitempty"`    //container host name
NetworkMode string `json:"NetworkMode,omitempty"` //Network mode for container
IPAddress   string `json:"IPAddress,omitempty"`   //Contaienr IP address
}
```

### 容器状态监控行为

* TaskGroup中共有一个失败，所有设置为失败，强杀所有容器
* 所有容器退出后，Executor默认退出

## executor随机端口机制

随机端口机制当前仅限CNM模式下的Host/Bridge，其他模式暂时无效。

* docker -P参数随机，该方式随机executor不做任何干预，默认需要镜像支持
* Host方式随机，bcs json中containerPort填写0，scheduler会分配随机端口，该端口会写入环境变量PORT_name中
* Bridge方式随机，该模式下如果hostPort为0，支持随机，该端口会写入环境变量PORT_name中

## CNI特性

executor命令行参数支持--network-mode参数，--network-mode默认为空值，如果要
开启cni特性，--network-mode需要填写字符串"cni"。

CNI目录定义为**/data/bcs/bcs-cni/bin**。

CNI特性实现依赖于executor对Pod的封装。

Pod定义：bcs-container-executor/container/pod.go。Pod工作流程

* NewPod
* Pod.Init
* Pod.Start
* Pod.containerMonitor

**Pod实现**：

* CNI pod实现：container/cni/cni.pod
  * Init：创建网络容器
  * Start：启动业务容器
  * Stop：关闭业务容器
  * Finit：关闭网络容器
* CNM pod实现：container/cnm/cnm.pod
  * Init：启动第一个业务容器
  * start：启动其他业务容器，网络与第一个共享
  * Stop：关闭所有业务容器
  * Finit：空

CNI pod的实现有一个较大的风险，当前网络容器如果被人为杀掉，无法回收网络资源。

CNI工具调用实现container/network/cni:

* cni.go：插件调用实现
  * SetupPod：针对Pod调用cni插件申请资源
  * TeardownPod：针对Pod调用cni插件回收资源
* manager：Pod网络流程管理
  * SetupPod：针对pod设置网络资源
  * TeardownPod：针对Pod回收网络资源

## CPU绑定和NUMA特性约束

* docker CPU绑定使用说明

docker api中，如果要绑定CPU所使用的核，使用参数cpuset-cpus，
例如--cpuset-cpus 0-3,6,7 ，逗号或者-进行分割。

* docker NUMA特性使用说明

docker参数中，提供cpuset-mems对使用的内存节点进行设定。
直接设置numa节点数字即可，例如0,1,2。如果设定的NUMA节点不存在，
docker会直接报错。

设定规则

* CPU数大于机器已有的CPU数，忽略
* 不设置CPU绑定标识，忽略
* 所需NUMA节点等于系统节点，忽略
* 如果NUMA节点只有一个，不做NUMA设置
* 随机算定一个NUMA节点，根据需求CPU个数使用随机选择NUMA节点中的CPU
* 暂时不支持跨NUMA节点使用

## FrameworkMessage约定

* Msg_LocalFile
  * Right：r，rw
  * User： root

## Docker启动参数设定约定

所使用的数据结构

* Environments：设置Docker启动时所需要环境变量
* Value：设置Docker启动时命令
* Arguments：设置Docker启动命令时的参数，如果有设置，默认拼接在Value之后

行为规定：

* 如果build镜像时有设置entrypoint：
  * 设置Value则会覆盖entrypoint
  * 不设置Value，默认使用Arguments作为entrypoint参数
* 如果镜像没有设置entrypoint，：
  * Value为空字符串，启动失败
  * 默认使用Value拼接Arguments作为CMD进行启动

## 网络参数

**类型设置**说明

**位置**：TaskInfo.Container.Docker.Network

设置类型（必须大写，建议直接引用Mesos定义）：

* HOST(ContainerInfo_DockerInfo_HOST)
* NONE(ContainerInfo_DockerInfo_NONE)
* BRIDGE(ContainerInfo_DockerInfo_BRIDGE)
* USER(ContainerInfo_DockerInfo_USER)

**USER**参数设置

使用USER模式时，默认从docker parameter中寻找--net参数，并作为网络参数进行提交

## 资源参数

Mesos数据类型Resource，需要设置字段

* Name：cpus,mem,gpus,
* Value_Type: 设置资源类型，默认Value_Scalar
* Value_Scalar.Value: 设置需要用到的值

**docker CPU**资源限定

* cpu-shares：CPU使用权重，默认权重（1个核为1024）
* cpuset-cpus：调度容器到指定CPU，例如[1,3]，在1，3号CPU运行
* cpuset-mems：指定容器使用的内存节点，NUMA架构才生效，需要写NUMA节点数

**docker MEM**内存限定，默认没有限制

* memory：最小4M。
* memory-swap：默认不限制，小于memory无效（无限制）。
* memory-swappiness：0为禁用是swap功能

**docker IO**资源限定

* device-write/read-bps: device-path:limit（unit），单位可以是 kb、mb、gb
* device-read/write-iops: 针对设备读写次数限制。格式同上，没有单位

golang针对NUMA信息提取范例：https://github.com/google/cadvisor/pull/1537/files

## secret实现

* 默认采用fetch api将远端https资源提取到本地，落地为文件，默认路径为mesos创建的SANDBOX目录
* bcs-container-executor通过TaskInfo.Data确认secret信息
  * 如果是自定义文件，指定文件路径与文件名，默认推入容器中，业务自行加载
  * 如果设置为环境变量类型，executor loading文件内容，设置到环境变量中（返回结果需要确保为key=value格式）

**要求**：

* 远端http/https接口提供文件下载功能

## Health Check

HealthCheck定义的维度是基于Application。

* 当Application状态为task_running时，HealthCheck才正式工作
* HealthCheck状态
  * healthy
  * unhealthy

HealthCheck状态检查规则

* 超时(timeoutSeconds)之前，必须要收到返回结果
* 如果是HTTP，返回码必须是200-399
* tcp端口检测需要能联通
* 失败次数达到maxConsecutiveFailures，需要杀死Task（待定）

Mesos TaskInfo中HealthCheck字段如下：

```golang
type HealthCheck struct {
// Amount of time to wait until starting the health checks.
DelaySeconds *float64 `protobuf:"fixed64,2,opt,name=delay_seconds,json=delaySeconds,def=15" json:"delay_seconds,omitempty"`
// Interval between health checks.
IntervalSeconds *float64 `protobuf:"fixed64,3,opt,name=interval_seconds,json=intervalSeconds,def=10" json:"interval_seconds,omitempty"`
// Amount of time to wait for the health check to complete.
TimeoutSeconds *float64 `protobuf:"fixed64,4,opt,name=timeout_seconds,json=timeoutSeconds,def=20" json:"timeout_seconds,omitempty"`
// Number of consecutive failures until signaling kill task.
ConsecutiveFailures *uint32 `protobuf:"varint,5,opt,name=consecutive_failures,json=consecutiveFailures,def=3" json:"consecutive_failures,omitempty"`
// Amount of time to allow failed health checks since launch.
GracePeriodSeconds *float64 `protobuf:"fixed64,6,opt,name=grace_period_seconds,json=gracePeriodSeconds,def=10" json:"grace_period_seconds,omitempty"`
// The type of health check.
Type *HealthCheck_Type `protobuf:"varint,8,opt,name=type,enum=mesos.v1.HealthCheck_Type" json:"type,omitempty"`
// Command health check.
Command *CommandInfo `protobuf:"bytes,7,opt,name=command" json:"command,omitempty"`
// HTTP health check.
Http *HealthCheck_HTTPCheckInfo `protobuf:"bytes,1,opt,name=http" json:"http,omitempty"`
// TCP health check.
Tcp  *HealthCheck_TCPCheckInfo `protobuf:"bytes,9,opt,name=tcp" json:"tcp,omitempty"`
XXX_unrecognized []byte                    `json:"-"`
}
```

状态数据通过**TaskStatus.Healthy**回发到scheduler，以TaskInfo为单位。Application的Health状态需要scheduler进行组装。

## 环境变量支持$操作符

executor默认会注入以下环境变量：

* HOST: cni模式下容器IP地址
* BCS_CNI_NAME: cni模式下容器网卡名字
* BCS_NODE_ADDR: 容器所在物理机IP地址
* BCS_POD_ID：容器POD ID

业务容器需要是使用HOST或者BCS_POD_ID信息，但是docker在实现环境变量时没有bash环境，不支持已有环境变量再次赋值。
为了方便操作，executor针对默认注入的环境变量支持再次赋值。例如在application json中：

```json
"env": [
  {"name": "test_env", "value": "test_env"},
  {"name": "CONTAINERIP", "value": "$HOST"},
  {"name": "RANDOMID", "value": "$BCS_POD_ID"}
]
```

上述实例中，默认提取HOST以及BCS_POD_ID信息，在容器启动前，直接将具体的值替换给CONTAINERIP与RANDOMID.

代码实现位置bcs-container-executor/container/task.go#EnvOperCopy

**注意**：$ 操作不支持嵌套。以下为错误，无法支持：

```json
{"name": "HOST", "value": "$OtherENV"},
{"name": "CONTAINERIP", "value": "$HOST"},
{"name": "RANDOMID", "value": "$BCS_POD_ID"}
```