# bcs application配置说明

bcs application实现Pod的含义，并与k8s的RC，Mesos的app概念等价。

## json配置模板

```json
{
	"apiVersion": "v4",
	"kind": "application",
	"restartPolicy": {
		"policy": "Never | Always | OnFailure",
		"interval": 5,
		"backoff": 10,
		"maxtimes": 10
	},
	"killPolicy": {
		"gracePeriod": 10
	},
	"constraint": {
		"intersectionItem": [
            { 
				"unionData": [{
					"name": "key1",
					"operate": "TOLERATION",
					"type": 3,
					"text": {
                            "value": "value1"
                    }
				}]
			},
			{
				"unionData": [{
					"name": "label",
					"operate": "EXCLUDE",
					"type": 4,
					"set": {
						"item": [
							"appname:gamesvr",
							"appname:dbsvr"
						]
					}
				}]
			},
			{
				"unionData": [{
					"name": "hostname",
					"operate": "CLUSTER",
					"type": 4,
					"set": {
						"item": [
							"slave3",
							"slave4",
							"slave5"
						]
					}
				}]
			},
			{
				"unionData": [{
					"name": "district",
					"operate": "GROUPBY",
					"type": 4,
					"set": {
						"item": [
							"shanghai",
							"shenzhen",
							"tianjin"
						]
					}
				}]
			},
			{
				"unionData": [{
					"name": "hostname",
					"operate": "LIKE",
					"type": 3,
					"text": {
						"value": "slave[3-5]"
					}
				}]
			},
			{
				"unionData": [{
					"name": "hostname",
					"operate": "UNLIKE",
					"type": 3,
					"text": {
						"value": "slave[1-2]"
					}
				}]
			},
			{
				"unionData": [{
					"name": "hostname",
					"operate": "UNIQUE"
				}]
			},
			{
				"unionData": [{
					"name": "idc",
					"operate": "MAXPER",
					"type": 3,
					"text": {
						"value": "5"
					}
				}]
			}
		]
	},
	"metadata": {
		"labels": {
			"test_label": "test_label",
			"io.tencent.bcs.netsvc.requestip.0": "127.0.0.1|InnerIp=127.0.0.1;127.0.0.2",
			"io.tencent.bcs.netsvc.requestip.1": "127.0.0.2|InnerIp=127.0.0.1;127.0.0.2"
		},
		"annotations": {
            "io.tencent.bcs.netsvc.requestip.0": "127.0.0.1|InnerIp=127.0.0.1;127.0.0.2",
            "io.tencent.bcs.netsvc.requestip.1": "127.0.0.2|InnerIp=127.0.0.1;127.0.0.2"
        },
		"name": "ri-test-rc-001",
		"namespace": "nfsol"
	},
	"spec": {
		"instance": 1,
		"template": {
			"metadata": {
				"labels": {
					"test_label": "test_label"
				}
			},
			"spec": {
				"containers": [{
					"hostname": "container-hostname",
					"command": "bash",
					"args": [
						"args1",
						"args2"
					],
					"parameters": [{
							"key": "rm",
							"value": "false"
						},
						{
							"key": "ulimit",
							"value": "nproc=8092"
						},
						{
							"key": "ulimit",
							"value": "nofile=65535"
						},
						{
							"key": "ip",
							"value": "127.0.0.1"
						}
					],
					"type": "MESOS",
					"env": [
                        {
                            "name": "test_env",
                            "value": "test_env"
                        },
                        {
                            "name": "namespace",
                            "value": "${bcs.namespace}"
                        },
                        {
                            "name": "http-port",
                            "value": "${bcs.ports.http_port}"
                        },
                        {
                            "name": "requests.cpu",
                            "valueFrom": {
                              "resourceFieldRef": {
                                  "resource": "requests.cpu"
                              }
                            }
                        },
                        {
                            "name": "requests.memory",
                            "valueFrom": {
                              "resourceFieldRef": {
                                  "resource": "requests.memory"
                              }
                            }
                        },
                        {
                            "name": "limits.cpu",
                            "valueFrom": {
                              "resourceFieldRef": {
                                  "resource": "limits.cpu"
                              }
                            }
                        },
                        {
                            "name": "limits.memory",
                            "valueFrom": {
                              "resourceFieldRef": {
                                  "resource": "limits.memory"
                              }
                            }
                        }
					],
					"image": "docker.hub.com/nfsol/log:92763",
					"imagePullUser": "userName",
					"imagePullPasswd": "passwd",
					"imagePullPolicy": "Always|IfNotPresent",
					"privileged": false,
					"ports": [{
							"containerPort": 8090,
							"hostPort": 8090,
							"name": "test-tcp",
							"protocol": "TCP"
						},
						{
							"containerPort": 8080,
							"hostPort": 8080,
							"name": "http-port",
							"protocol": "http"
						}
					],
					"healthChecks": [{
						"type": "HTTP|TCP|COMMAND|REMOTE_HTTP|REMOTE_TCP",
						"intervalSeconds": 30,
						"timeoutSeconds": 5,
						"consecutiveFailures": 3,
						"gracePeriodSeconds": 5,
						"http": {
							"port": 8080,
							"portName": "test-http",
							"scheme": "http|https",
							"path": "/check"
						},
						"tcp": {
							"port": 8090,
							"portName": "test-tcp"
						},
						"command": {
                            "value": "ls /"
                        }
					}],
					"resources": {
						"limits": {
							"cpu": "2",
							"memory": "8"
						},
						"requests": {
							"cpu": "2",
							"memory": "8"
						}
					},
					"volumes": [{
						"volume": {
							"hostPath": "/data/host/path",
							"mountPath": "/container/path",
							"readOnly": false
						},
						"name": "test-vol"
					}],
					"secrets": [{
						"secretName": "mySecret",
						"items": [{
								"type": "env",
								"dataKey": "abc",
								"keyOrPath": "SRECT_ENV"
							},
							{
								"type": "file",
								"dataKey": "abc",
								"keyOrPath": "/data/container/path/filename.conf",
								"subPath": "relativedir/",
								"readOnly": false,
								"user": "root"
							}
						]
					}],
					"configmaps": [{
						"name": "template-configmap",
						"items": [{
								"type": "env",
								"dataKey": "config-one",
								"keyOrPath": "SECRET_ENV"
							},
							{
								"type": "file",
								"dataKey": "config_two",
								"dataKeyAlias": "config-two",
								"KeyOrPath": "/data/contianer/path/filename.txt",
								"readOnly": false,
								"user": "root"
							}
						]
					}]
				}],
				"networkMode": "BRIDGE",
				"networkType": "cnm",
				"netLimit": {
					"egressLimit": 100
				}
			}
		}
	}
}
```

## KillPolicy机制

* gracePeriod：宽限期描述在强制kill container之前等待多久，单位秒。 默认为1

## RestartPolicy机制

restartPolicy字段说明：

* policy：支持Never Always OnFailure三种配置(默认为OnFailure),OnFailure表示在失败的情况下重新调度,Always表示在失败和Lost情况下重新调度, Never表示任何情况下不重新调度
* interval: 失败后到执行重启的间隔(秒),默认为0
* backoff：多次失败时,每次重启间隔增加秒,默认为0.如果interval为5,backoff为10,则首次失败时5秒后重新调度,第二次失败时15秒后重新调度,第三次失败时25秒后重新调度
* maxtimes: 最多重新调度次数,默认为0表示不受次数限制.容器正常运行30分钟后重启次数清零重新计算

## constraint调度约束

constraint字段用于定义调度策略

* IntersectionItem

该字段为数组，策略为：所有元素同时满足时才进行调度。使用多个IntersectionItem可以同时满足多个条件限制。

每个IntersectionItem中支持多个UnionData为规则细节字段，用于填写调度匹配规则，也为数组，该数组含义为：所有元素只需要满足一个即可。使用UnionData可以实现多个条件满足其一即可。
UnionData字段说明：

```json
{
"name": "hostname",
"operate": "ClUSTER",
"type": 3|4,
"set": {
"item": ["mesos-slave-1", "mesos-slave-2", "mesos-slave-3"]
},
"text":{
"value": "mesos-slave-1"
}
}
```

* name: 调度字段key（如主机名，主机类型，主机IDC等）用来调度的字段需要mesos slave通过属性的方式进行上报，hostname参数自动获取无需属性上报
* operator：调度算法，当前支持以下5种调度算法（大写）：
  * UNIQUE: 每个实例的name的取值唯一：如果name为主机名则表示每台主机只能部署一个实例，如果name为IDC则表示每个IDC只能部署一个实例。UNIQUE算法无需参数。
  * MAXPER: name同一取值下最多可运行的实例数，为UNIQUE的增强版（数量可配置），MAXPER算法需通过参数text(type为3)指定最多运行的实例数。
  * CLUSTER: 配合set字段（type为4），要求name的取值必须是set中一个，可以限定实例部署在name的取值在指定set范围。
  * LIKE: 配合text字段（type为3）或者set字段（type为4），name与text（或者set）中的内容进行简单的正则匹配，可以限定实例部署时name的取值。如果是参数是set（type为4），只要和set中某一个匹配即可。
  * UNLIKE: LIKE取反。如果是参数为set（type为4），必须和set中所有项都不匹配。
  * GROUPBY: 根据name的目标个数，实例被均匀调度在目标上，与set一起使用，如果实例个数不能被set的元素个数整除，则会存在差1的情况，例如：name为IDC，实例数为3,set为["idc1","idc2"],则会在其中一个idc部署两个实例。
  * EXCLUDE: 和具有指定标签的application不部署在相同的机器上，即：如果该主机上已经部署有这些标签（符合一个即可）的application的实例，则不能部署该application的实例。目前name只支持"label",label的k:v在set数组中指定。
  * GREATER: 配合scaler字段（type为1），要求name的取值必须大于scalar的值。
  * TOLERATION: 容忍被打taint的node
* type: 参数的数据类型，决定operator所操作key为name的值的范围
1: scaler: float64
3：text：字符串。
4：set：字符串集合。

案例说明：

* 要求各实例运行在不同主机上

```json
{
"name": "hostname",
"operate": "UNIQUE"
}
```

* 要求实例限制运行在深圳和东莞地区（要求slave上报section字段）

```json
{
"name": "section",
"operate": "ClUSTER",
"type": 4,
"set": {
"item": ["shenzhen", "dongguan"]
}
}
```

* 要求实例运行在mesos-slave-1，mesos-slave-2上

```json
{
"name": "hostname",
"operate": "LIKE",
"type": 3,
"text": {
"value": "mesos-slave-[1-2]"
}
}

```

## meta元数据

* name: Application名字，小写字母与数字构成，但不能完全由数字构成，不能数字开头
* namespace：App命名空间，小写字母与数字构成，不能完全由数字构成，不能数字开头；不同业务必然不同，默认值为defaultGroup
* label：app的lable信息，对应k8s RC label

### Label或Annotations特殊字段说

当容器网络使用bcs-cni方案的时，如果想针对容器指定IP，可以使用以下label

* io.tencent.bcs.netsvc.requestip.i：针对Pod申请IP，i代表Pod的实例，从0开始计算

当容器指定Ip时，不同的taskgroup需要调度到特定的宿主机上面，宿主机的制定方式与constraint调度约束一致，如下是使用InnerIp：
io.tencent.bcs.netsvc.requestip.i: "127.0.0.1|InnerIp=127.0.0.1;127.0.0.2"
使用分隔符"|"分隔，"|"前面为容器Ip，后面为需要调度到的宿主机Ip，多个宿主机之间使用分隔符";"分隔，宿主机Ip支持正则表示式，方式如下：
io.tencent.bcs.netsvc.requestip.i: "127.0.0.1|InnerIp=127.0.0.[12-25];127.0.0.[11-13]"

## 容器字段信息

* instance：运行实例个数
* label：运行时容器label信息，对应k8s pod label
* name: pod名字，mesos中不启用
* type：DOCKER/MESOS
* hostname: 容器的hostname设置，如果网络模式为"HOST"，则该字段无效
* command：字符串，容器启动命令，例如/bin/bash
* args：command的参数，例如 ["-c", "echo hello world"]
* env: key/value格式，环境变量。针对BCS默认注入的环境变量（BCS_CONTAINER_IP,BCS_NODE_IP）支持赋值操作
* parameters：docker参数，当前以下docker参数已支持
  * oom-kill-disable：有效值true或false；设置为true后，如果容器内存资源超限不会进行强杀
  * ulimit：可以设置ulimit参数，例如core=-1
  * rm：有效值为true，容器退出后，是否直接删除容器
  * shm-size: 有效值为自然数，可以设置/dev/shm大小，单位MB，默认是64MB。例如：128
  * ip: 用户自定义网络模式时，可以设置容器ip。只有自定义网络模式时生效
* image：镜像链接
* imagePullSecrets：存储仓库鉴权信息的secret名字
* imagePullPolicy：拉取容器策略
  * Always：每次都重新从仓库拉取
  * IfNotPresent：如果本地没有，则尝试拉取（默认值）
* privileged：容器特权参数，默认为false
* resources：容器使用资源
  * request.cpu:字符串，可以填写小数，1为使用1核，cpu软限制，对应cpu_shares。调度分配的值，如果不填写默认为limit.cpu
  * request.memory：内存使用，字符串，单位默认为M，memory下限。调度分配的值，如果不填写默认为limit.memory。注意：当memory >= 4Mb, 使用memory的值限制内存；否则，不对memory做limits
  * request.storage：磁盘使用大小，默认单位M
  * limits.cpu:字符串，可以填写小数，1为使用1核，cpu硬限制，对应cpu_quota、cpu_period
  * limits.memory：内存使用，字符串，单位默认为M，memory上限。
  * limits.storage：磁盘使用大小，默认单位M
* cpuset: 是否cpu绑定核，此参数与resources.request.cpu配合使用，并且cpu必须为整数，对应docker参数--cpuset-cpus
* networkMode：网络模式
  * HOST: docker原生网络模式，与宿主机共用一个Network Namespace，此模式下需要自行解决网络端口冲突问题
  * BRIDGE: docker原生网络模式，此模式会为每一个容器分配Network Namespace、设置IP等，并将一个主机上的Docker容器连接到一个虚拟网桥上，通过端口映射的方式对外提供服务
  * NONE: 除了lo网络之外，不配置任何网络
  * ~~USER: 用户深度定制的网络模式，支持macvlan、calico等网络方案。建议不再使用~~
  * 自定义: 用户深度定制的网络模式，例如cni标准，支持macvlan、calico等网络方案；支持docker原声的自定义网络模式
* networkType：容器网络标准
  * cni(小写): 使用bcs提供的cni来构建网络，具体cni的类型是由配置决定
  * cnm：使用docker原生或用户自定义的方式来构建网络

### 容器env环境变量支持bcs系统常量
application容器中env环境变量的配置，支持几种bcs系统常量
- ${bcs.namespace}  //namespace
- ${bcs.appname}  //application name
- ${bcs.instanceid}   //pod index，例如：0
- ${bcs.hostip}  //pod所调度、部署的物理机ip
- ${bcs.ports.port-name}  //bcs支持系统分配随机端口，端口范围在31000-32000，port-name表示ports端口的name。${ports.port-name}变量用于表示随机分配的端口号，业务可以通过启动脚本参数或环境变量的方式获取该端口号。例如：上述json文件中的${ports.http-port}变量表示系统随机分配的ports name为http-port的端口号，业务可以使用环境变量，启动参数等方式获取
- ${bcs.taskgroupid}  //application taskgroup id
- ${bcs.taskgroupname}  //application taskgroup name

例如
```json
"env": [
    {
        "name": "namespace",
        "value": "${bcs.namespace}"
    },
    {
        "name": "http-port",
        "value": "${bcs.ports.http_port}"
    }
]
```

### **netLimit说明**

对容器的网络流量进行限制，包括ingress(进流量)和egress(出流量)，默认情况下不限制。
暂时只支持对egress的限制。

* egressLimit: 容器出流量的限制，value为大于0的整数，单位为Mbps。

### **volume说明**

支持主机磁盘挂载

* name：挂载名，在app中需要保持唯一
* volume.hostPath: 主机目录
* 当目录没填写时，默认为创建一个随机目录，该目录Pod唯一
* 该目录路径可以支持变量$BCS_POD_ID
* volume.mountPath: 需要挂载的容器目录，需要其父目录存在，否则报错
* readOnly: true/false, 是否只读，默认false

### **configmap说明**

主要功能是引用configmap数据，并作为环境变量/文件注入容器中。

* name：configmap索引名字

**环境变量**注入

* items[x].type: env
* items[x].dataKey: configmap子项索引名
* items[x].keyOrPath: 需要注入环境变量名

**文件**注入

* items[x].type: file
* items[x].dataKey: configmap子项索引名，默认为文件名
* items[x].dataKeyAlias: 对于k8s configmap，如果原始文件名带有"_"，则需要对文件进行重命名，使用keyAlias对子项进行索引。也可以增加前缀目录，keyAlias拼接默认构成完成容器路径。
* items[x].keyOrPath: 文件在容器中路径，需要保证父目录存在
* items[x].readOnly: true/false，默认false，文件是否只读
* items[x].user: 文件用户设置，默认root，k8s不生效

### **secrets机制**

secret在k8s和mesos中实现存在差异。在k8s中，即为默认支持的secret数据，并存储在etcd中；在mesos中，secret为bcs-scheduler增加
的数据结构，数据默认存储在vault中，读写控制需要通过bcs-authserver。secrets的数据默认可以注入环境变量/文件。

在k8s中，secret只能存储一项数据，所以不存在子项数据结构。mesos下，一个secret可以存储多项数据。

* secretName: 引用的secret名字

注入**环境变量**：

* items[x].type: env
* items[x].dataKey: secret中子项索引，mesos有效
* items[x].keyOrPath： 环境变量KEY

注入**文件**

* items[x].type: file
* items[x].dataKey：secret中子项索引，mesos有效
* items[x].keyOrPath：需要挂载的容器目录
* items[x].subPath：需要挂载子目录，仅k8s有效
* items[x].readOnly： true/false，文件是否只读，默认为false
* items[x].user：root，文件属主，mesos有效

### **容器Ports机制说明**

ports字段说明：

* protocol：协议，必填，http，tcp，udp
* name：标识port信息，唯一，必填
* containerPort：容器中服务使用端口，必填
* hostPort: 物理主机使用的端口，0代表scheduler进行随机选择

特别说明：

* host模式下，containerPort即代表hostPort
  * 填写固定端口，需要业务自行确认是否产生冲突
  * 填写0，意味着scheduler进行随机选择
* bridge模式下，hostPort代表物理主机上的端口
* hostPort填写固定端口，业务自行解决冲突的问题
  * 填写0，scheduler默认进行端口随机
  * 小于0，不进行端口映射
* 自定义模式下，hostPort代表物理主机上的端口
* hostPort填写固定端口，业务自行解决冲突的问题
  * 填写0，scheduler默认进行端口随机
  * 小于0，不进行端口映射

**端口随机**的状态下，scheduler会根据ports字段序号，生成PORT0 ~ n的环境变量，以便业务读取该随机端口。不支持PORT_NAME的方式

### **容器Health Check机制说明**

#### 通过mesos协议下发检测机制到executor，通过executor执行检测

* 支持的类型为HTTP,TCP两种
* scheduler根据application定义的检测机制，启动进程时下发到executor
* executor根据检测配置实施检测（并在多次检测失败的情况下kill进程，可配置）
* executor将检测结果通过TaskStatus中的healthy（bool）上报到scheduler
* scheduler根据healthy的值以及进程的其他数据(配置数据和动态数据)来确定后续行为：
* 状态修改，数据记录，触发告警等
* 重新调度

#### mesos scheduler根据检测机制直接远程执行检测

* 支持的类型为REMOTE_HTTP,REMOTE_TCP

#### health check Type说明

* health check可以同时支持多种类型的check
* HTTP,TCP两种类型，最多只能同时支持一种
* REMOTE_HTTP,REMOTE_TCP两种类型可以同时支持

#### **healthChecks 字段说明**

* type: 检测方式，目前支持HTTP,TCP,COMMAND,REMOTE_TCP和REMOTE_HTTP五种
* intervalSeconds：前后两次执行健康监测的时间间隔.
* timeoutSeconds: 健康监测可允许的等待超时时间。在该段时间之后，不管收到什么样的响应，都被认为健康监测是失败的，**timeoutSeconds需要小于intervalSeconds**
* consecutiveFailures: 当该参数配置大于0时，在健康检查连续失败次数大于该配置时，scheduler将task设置为Failed状态并下发kill指令（设置为Failed状态后会出发重新调度检测，如果配置了Failed状态下重新调度，则scheduler会重新调度对应的taskgroup）。目前该配置项只在executor本地check有效。如果不需要此功能，请配置为0。
* gracePeriodSeconds：启动之后在该时段内不进行健康检查
* command: type为COMMAND时有效
  * value: 需要执行的命令,value中支持环境变量.mesos协议中区分是否shell,这里不做区分,如果为shell命令,需要包括"/bin/bash ‐c",系统不会自动添加(参考marathon)
  * 后续可能需要补充其他参数如USER
* http: type为HTTP和REMOTE_HTTP时有效
  * port: 检测的端口,如果配置为0,则该字段无效
  * portName: 检测端口名字(替换marathon协议中的portIndex)
    * portName在port配置大于0的情况下,该字段无效
    * portName在port配置不大于0的情况下,检测的端口通过portName从ports配置中获取（scheduler处理）
    * 根据portName获取端口的时候,需要根据不同的网络模型获取不同的端口，目前规则(和exportservice保持一致)如下：
      * BRIDGE模式下如果HostPort大于零则为HostPort,否则为ContainerPort
      * 其他模式为ContainerPort
  * scheme： http和https(https不会做认证的处理)
  * path：请求路径
  * headers: http消息头，为了支持health check时，需要认证的方式，例如：Host: www.xxxx.com。NOTE:目前只支持REMOTE_HTTP。
  * 检测方式:
    *  Sends a GET request to scheme://<host>:port/path.
    *  Note that host is not configurable and is resolved automatically, in most cases to 127.0.0.1.
    *  Default executors treat return codes between 200 and 399 as success; custom executors may employ a different strategy, e.g. leveraging the `statuses` field.
    *  bcs executor需要根据网络模式等情况再具体确认规则
* tcp： type为TCP和REMOTE_TCP的情况下有效：
  * port: 检测的端口,如果配置为0,则该字段无效
  * portName: 检测端口名字(替换marathon协议中的portIndex)
    * protName在port配置大于0的情况下,该字段无效
    * portName在port配置不大于0的情况下,检测的端口通过portName从ports配置中获取（scheduler处理）
    * 根据portName获取端口的时候,需要根据不同的网络模型获取不同的端口，目前规则(和exportservice保持一致)如下：
      * BRIDGE模式下如果HostPort大于零则为HostPort,否则为ContainerPort
      * 其他模式为ContainerPort
  * 检测方式： tcp连接成功即表示健康，需根据不同网络模型获取不同的地址
