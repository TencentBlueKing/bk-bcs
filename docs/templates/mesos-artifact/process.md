# bcs process配置说明

bcs process是对进程服务的抽象，用于进程服务的描述。

## json配置模板

```json
{
	"apiVersion": "v4",
	"kind": "process",
	"restartPolicy": {
		"policy": "Never | Always | OnFailure",
		"interval": 5,
		"backoff": 10,
		"maxtimes": 10
	},
	"killPolicy": {
		"gracePeriod": 10
	},
	"metadata": {
		"labels": {
			"test_label": "test_label"
		},
		"name": "gamesvc",
		"namespace": "ns-a"
	},
	"spec": {
		"instance": 3,
		"template": {
			"spec": {
				"processes": [{
					"procName": "gamesvr",
					"user": "user",
					"workPath": "${work_base_dir}/${namespace}.${processname}.${instanceid}/gamesvc",
					"pidFile": "${run_base_dir}/${namespace}.${processname}.${instanceid}/pid/gamesvc.pid",
					"ports": [{
							"hostPort": 8899,
							"name": "http_port",
							"protocol": "HTTP"
						},
						{
							"hostPort": 0,
							"name": "metric_port",
							"protocol": "TCP"
						}
					],
					"uris": [{
						"value": "http://xx.xxx.xxxx.com/xxxx/gamesvc-v1.tar.gz",
						"user": "xxxx",
						"pwd": "xxxxxx",
						"pullPolicy": "Always|IfNotPresent",
						"outputDir": "${work_base_dir}/${namespace}.${processname}.${instanceid}"
					}],
					"startCmd": "./start.sh --address ${hospip} --service-port ${ports.http-service}",
					"startGracePeriod": 10,
					"stopCmd": "./stop.sh --pidfile ${pidFile}",
					"reloadCmd": "./reload.sh --pidfile ${pidFile}",
					"healthChecks": [{
						"type": "HTTP|TCP|COMMAND|REMOTE_HTTP|REMOTE_TCP",
						"intervalSeconds": 60,
						"timeoutSeconds": 20,
						"consecutiveFailures": 3,
						"gracePeriodSeconds": 300,
						"command": {
							"value": "./check.sh"
						},
						"http": {
							"port": 8080,
							"portName": "test-http",
							"scheme": "http|https",
							"path": "/check",
							"headers": {
								"key1": "value1",
								"key2": "value2"
							}
						},
						"tcp": {
							"port": 8090,
							"portName": "test-tcp"
						}
					}],
					"resources": {
						"limits": {
							"cpu": "2",
							"memory": "500"
						}
					},
					"env": [{
							"name": "http_port",
							"value": "${ports.http_port}"
						},
						{
							"name": "metric_port",
							"value": "${ports.metric_port}"
						},
						{
							"name": "hostip",
							"value": "${hostip}"
						},
						{
							"name": "work_dir",
							"value": "${workPath}/"
						},
						{
							"name": "pid_file",
							"value": "${pidFile}/"
						},
						{
							"name": "namespace",
							"value": "${namespace}"
						},
						{
							"name": "processname",
							"value": "${processname}"
						},
						{
							"name": "instanceid",
							"value": "${instanceid}"
						}
					],
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
								"readOnly": false,
								"user": "user"
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
				}]
			}
		}
	},
	"constraint": {
		"intersectionItem": [{
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
					},
					{
						"unionData": [{
							"name": "label",
							"operate": "EXCLUDE",
							"type": 4,
							"set": {
								"item": [
									"processname:gamesvr",
									"processname:dbsvr"
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
					}
				]
			}
		]
	}
}
```

## KillPolicy机制

* gracePeriod：宽限期描述在强制kill process之前等待多久，单位秒。 默认为1

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

* name: process名字，小写字母与数字构成，但不能完全由数字构成，不能数字开头
* namespace：process命名空间，小写字母与数字构成，不能完全由数字构成，不能数字开头；不同业务必然不同，默认值为defaultGroup
* label：process的lable信息，对应k8s RC label

## 常用变量
process的定义支持变量方式，bcs提供几种常用的变量，在json填写时可以使用。
- ${namespace}  //namespace
- ${processname}  //process name
- ${instanceid}   //pod index，例如：0
- ${hostip}  //pod所调度、部署的物理机ip
- ${ports.port-name}  //bcs支持系统分配随机端口，端口范围在31000-32000，port-name表示ports端口的name。${ports.port-name}变量用于表示随机分配的端口号，业务可以通过启动脚本参数或环境变量的方式获取该端口号。例如：上述json文件中的${ports.http-service}变量表示系统随机分配的ports name为http-service的端口号，业务可以使用环境变量，启动参数等方式获取
- ${work_base_dir}  //进程的工作根目录，由bcs分配，业务通过变量方式获取
- ${run_base_dir}  //进程的运行根目录，每一个pod都会有一个运行目录，该目录主要用于进程相关的文件的存储，例如：pid文件，log日志等
- ${workPath} //进程工作目录，即业务自定义的spec.processes.workPath的value
- ${pidFile} //进程的pid文件，即业务自定义的spec.processes.pidFile的value

例如
```json
 "pidFile": "${run_base_dir}/${namespace}.${processname}.${instanceid}/pid/lgamesvr.pid",
 "startCmd": "./bin/start.sh --address ${hospip} --service-port ${ports.http-service}",
 "env": {
    "hostip": "${hostip}"
 }
```

## 进程字段信息
- instance: pod数量
- user: 进程启动的用户，例如：user
- workPath: 字符串，进程的工作目录，启动进程时，首先会cd到该目录，再执行命令，支持变量；一般情况下每个pod都有一个独立的工作目录，例如：${work_base_dir}/${namespace}.${processname}.${instanceid}/lgamesvr
- procName: 字符串，进程名称，检查进程存活状态时，会通过pid+procName来判断进程是否存活
- pidFile: 进程的pid文件，文件内为进程的pid，int类型，例如：3245，支持变量；例如：${run_base_dir}/${namespace}.${processname}.${instanceid}/pid/gamesvc.pid
- startCmd: 进程的启动命令
- startGracePeriod：启动等待时间，执行startCmd命令后，等待进程的启动时间，单位s，默认值1
- stopCmd: 进程的停止命令
- restartCmd: 进程的重启命令
- reloadCmd：进程的reload命令
- uris  //表示业务发布的程序包存储在bcs的仓库。调度时，bcs会去相应的地址拉去程序包，部署到物理机上面，然后再启动进程
   - value: 程序包在bcs仓库的路径，例如：/prod/lgamesvr-v1.tar.gz
   - user: 仓库用户名
   - pwd: 仓库密码
   - imagePullPolicy：拉取容器策略。Always：每次都重新从仓库拉取；IfNotPresent：如果本地没有，则尝试拉取（默认值）
   - outputDir: fetch文件的解压目录。例如：${work_base_dir}/${namespace}.${processname}.${instanceid}，该目录一般与workpath配合使用
- environment: 设置系统环境变量
- resources
   - limits.cpu: 字符串，可以填写小数，1为使用1核，如果-1表示不限制
   - limits.memory：内存使用，字符串，单位默认为M. 如果-1表示不限制
- ports
   - port: 服务端口
   - name：标识port信息，唯一，必填
   - protocol: 协议，http、tcp
特别说明：
此种模式下使用的是宿主机的网络命名空间
     - port填0，表示scheduler随机选择，范围：31000-32000。此情境下，会通过环境变量或启动脚本参数的方式告知业务进程该端口的值
     - port填非0值，需自行解决端口冲突问题

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
* items[x].user：user，文件属主，mesos有效

### **容器Health Check机制说明**

#### 通过mesos协议下发检测机制到executor，通过executor执行检测

* 支持的类型为HTTP,TCP和COMMAND三种
* scheduler根据application定义的检测机制，启动进程时下发到executor
* executor根据检测配置实施检测（并在多次检测失败的情况下kill进程，可配置）
* executor将检测结果通过TaskStatus中的healthy（bool）上报到scheduler
* scheduler根据healthy的值以及进程的其他数据(配置数据和动态数据)来确定后续行为：
  * 状态修改，数据记录，触发告警等
  * 重新调度

#### mesos scheduler根据检测机制直接远程执行检测

* 支持的类型为REMOTE_HTTP,REMOTE_TCP

#### health check Type说明

* health check可以同时支持多种类型的check，目前最多为三种
* HTTP,TCP和COMMAND三种类型，最多只能同时支持一种
* REMOTE_HTTP,REMOTE_TCP两种类型可以同时支持

#### **healthChecks 字段说明**

* type: 检测方式，目前支持HTTP,TCP,COMMAND,REMOTE_TCP和REMOTE_HTTP五种
* delaySeconds：容器启动之后到开始进行健康检测的等待时长(mesos协议中有,marathon协议中不支持,因为有gracePeriodSeconds,该参数好像意义不大,可能被废弃)
* intervalSeconds：前后两次执行健康监测的时间间隔.
* timeoutSeconds: 健康监测可允许的等待超时时间。在该段时间之后，不管收到什么样的响应，都被认为健康监测是失败的，**timeoutSeconds需要小于intervalSeconds**
* consecutiveFailures: 在一个不健康的任务被杀掉之前，连续的健康监测失败次数，如果值设为0，tasks如果不通过健康监测，则它不会被杀掉。进程被杀掉后scheduler根据application的restartpolicy来决定是否重新调度. marathon协议中为maxConsecutiveFailures
* gracePeriodSeconds：启动之后在该时段内健康监测失败会被忽略。或直到任务首次变成健康状态.
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
