# bk-bscp-client使用手册

BSCP(Blueking Service Configuration Platform)蓝鲸服务配置平台。

## 基础概念说明

**Business**：业务划分

**Application(app)**：业务之下具体模块的划分。一个App只能属于一个业务。

**Cluster**：App之下的逻辑划分，可以映射为模块之下的`逻辑可用区`、`物理集群`或者`区域`

**Zone**：cluster之下细分一级逻辑集合。建议可以映射为业务的大区。

**ConfigSet(cfgset)**：App下一个配置集合，可以映射为单一配置文件。如果有多个配置文件，建议创建多个cfgset；
针对kv形式的配置，configset包含多个键值对。

**Template**: 单一配置文件的模板。由于业务针对不同地域、可用区、大区等存在差异，可以通过从cluster，zone中选取特性，
渲染出目标具体配置。`迭代中`。

**Commit**：对多个ConfigSet内容的提交，提交前需要先将提交的配置文件添加到扫描区。发起提交，会自动以提交文件的相对路径创建对应的ConfigSet，并生成对应的提交记录，提交之后，扫描区提交文件列表会被清空。一个提交记录是由多个模块组成，一个模块对应一个ConfigSet，可以通过提交记录取获取单模块提交详情消息。也可以连续多次提交，并查阅历史提交记录。

**Release**：配置发布，关联已生效Commit，并关联发布策略，例如需要发布到某个cluster，zone，或者匹配节点自定义的kv对，说明此次发布细节。一个Release也是有多个模块组成，每个模块与它所关联的Commit模块相关联。提交的release也需要发布(publish)才能正式发布到目标终端。

## bscp-client安装

执行安装脚本，指定连入BSCP后台服务地址。

```
sh install.sh 127.0.0.1:9510
```

会自动生成默认配置 `/etc/bscp/client.yaml`。

```yaml
kind: bscp-client
version: 0.1.1
host: 127.0.0.1:9510
```

配置简单说明：

* kind：应用类型。
* version：应用版本。
* host：BSCP系统接入的地址。

初始化完成之后，即可进行client使用。

```shell
$ bk-bscp-client --help
bk-bscp-client controls the BlueKing Service Configuration Platform.

Publishing ConfigSet steps：
    bk-bscp-client init  -> bk-bscp-client add -> bk-bscp-client commit -> bk-bscp-client release -> bk-bscp-client publish

Explanation:
    First initialize the configuration file repository, and then add the configuration files to be submitted to the scanning area. Then, use the commit command to submit the scan area file (after submission, the content in the scan area will be cleared). Use the release command to select the commit record submitted and generate the corresponding release version. Finally, use the publish command to select the release version to be published for publication.

Usage:
  bk-bscp-client [command]

Available Commands:
  add         Add the configuration file to the scan area
  cancel      Cancel specified release
  checkout    Checkout the file from the scan area
  commit      Submit the files in the scan area
  config      Set or Modify the default parameter of the configuration file repository
  create      Create new resource
  delete      Delete resource
  get         Get specified resource detail
  help        Help about any command
  info        Get repository init info
  init        Initialize the application configuration file repository
  list        List resources
  lock        Lock resource
  publish     Publish release
  release     Create release
  reload      Reload release
  rollback    Rollback release
  status      Show the working tree status
  unlock      Unlock resource
  update      Update resource

Flags:
      --business string   business Name to operate. Get parameter priority: command -> env -> .bscp/desc
  -h, --help              help for bk-bscp-client
      --operator string   user name for operation.  Get parameter priority: command -> env -> .bscp/desc
      --token string      user token for operation. Get parameter priority: command -> env -> .bscp/desc
      --version           version for bk-bscp-client

Use "bk-bscp-client [command] --help" for more information about a command.
```

## 本地应用仓库配置命令

本地应用仓库可以设置 business、app、operator、token 默认参数值。后续执行其他命令时，如果在命令行和环境变量中没有设置这些参数，将读取当前执行命令仓库设置的默认参数值。并且，客户端业务配置命令集只有在初始化过后的仓库下才可以执行。

### 操作步骤：

**创建本地应用配置仓库**：

```shell
mkdir -p /data/bscp/X-GameConfigRepo && cd /data/bscp/X-GameConfigRepo
```

**执行初始化操作**：

```shell
> bk-bscp-client init --business X-Game --app gameserver --operator guohu --token admin
Initialize empty bscp operation directory successfully in /data/bscp/X-GameConfigRepo/.bscp/
```

**查看当前仓库默认配置**：

```shell
> bk-bscp-client info
Current repository[/data/bscp/X-GameConfigRepo] init info:
Business: X-Game
App: gameserver
Operator: guohu
Token: guohu:admin
```

**修改当前仓库默认配置：**

```shell
> bk-bscp-client config --local app game
Set the default parameter successfully! app: game
```

**初始化生成文件介绍：**

```
> tree tree .bscp/
.bscp/
├── desc
└── record
```

- desc：用于存放默认参数
- record：用于记录添加到扫描区的文件列表

**环境变量设置：**

```shell
> export BSCP_BUSINESS=X-Game
> export BSCP_APP=game
> export BSCP_OPERATOR=MrMGXXXX
> export BSCP_TOKEN=guohu:admin
```

**默认参数读取优先级：**

命令行输入 -> 环境变量设置 -> 当前应用仓库默认配置

### 推荐使用方式：

**建立如下目录结构：**

```shell
X-Game/								# business 名称
├── game							# X-Game业务下所属 application 名称
└── gameserver				# X-Game业务下所属 application 名称
```

在业务目录下执行初始化命令时，设置 business、operator、token 参数，作为业务层级管理目录，执行系统管理命令。在应用目录下执行初始化操作时，设置 business、operator、token、app 参数，作为应用层级管理目录，执行业务配置命令。

## 业务配置命令

业务配置命令，主要对BSCP用户开放，主要涉及类型如下：

* 创建/列表查询/通过id-name查询/更新Application(app)
* 创建/列表查询/通过id-name查询/更新App下的Cluster、Zone
* 创建/列表查询/通过id-name查询app关联的ConfigSet
* 创建/列表查询/通过id-name查询app关联的Strategy
* 创建/列表查询/通过id查询Commit/通过mid查询Commit模块
* 创建/发布/列表查询/通过id查询Release/通过mid查询Release模块
* 通过 App、Cluster、Zone、实例状态进行过滤的实例列表查询
* 添加文件到扫描区/撤销扫描区文件/查看本地文件列表和扫描区文件列表
* 应用仓库初始化/修改应用仓库默认配置/查看配置信息
* 通过 strategy 查询可触达的应用实例列表
* 通过 release 查询 reload生效在线实例、发布生效在线实例、未生效在线实例、reload生效下线实例、发布生效下线实例列表及其数量

### Application操作

```shell
> bk-bscp-client create app --name gameserver
Create Application successfully: A-6fcccf08-ccb4-11ea-9cfe-5254006865b1

> bk-bscp-client list app
+----------------------------------------+------------+-----------+------------+----------+
|                   ID                   |    NAME    |   TYPE    |   STATE    | BUSINESS |
+----------------------------------------+------------+-----------+------------+----------+
| A-6fcccf08-ccb4-11ea-9cfe-5254006865b1 | gameserver | container | AFFECTIVED | X-Game   |
+----------------------------------------+------------+-----------+------------+----------+

> bk-bscp-client get app --name gameserver       # 通过 --id appId 也可查询，其他提供名称查询的命令集同样支持
Name: 		gameserver
ApplicationID: 	A-6fcccf08-ccb4-11ea-9cfe-5254006865b1
Business: 	B-48de67cb-b6d5-11ea-90b2-5254006865b1 - X-Game
DeployType: 	Container
State:		Affectived
Memo:
Creator: 	guohu
LastModifyBy: 	guohu
CreatedAt: 	2020-07-23 15:16:33
UpdatedAt: 	2020-07-23 15:16:33

> bk-bscp-client update app --id A-6fcccf08-ccb4-11ea-9cfe-5254006865b1 --name gameser 
Update resources successfully
```

### Cluster和Zone操作

Cluster和Zone是为了定义业务逻辑架构，便于在配置模板渲染，生成和存储时能为业务不同的逻辑架单元
生成隔离独立的配置。建议业务在Application创建之后就完成该逻辑层次划分。

**修改应用仓库默认配置 app 信息：**

```shell
> bk-bscp-client config --local app gameserver
Set the default parameter successfully! app: gameserver
```

修改完成之后，执行后续其他命令时，app参数可以忽略，默认读取此处设置的 app 参数（命令行和环境变量不存在此参数）。

#### Cluster操作

```shell
> bk-bscp-client create cluster --name cluster-shenzhen
create Cluster successfully: C-4169f103-ccb6-11ea-9cfe-5254006865b1

> bk-bscp-client create cluster --name cluster-shanghai
create Cluster successfully: C-584a03c6-ccb6-11ea-9cfe-5254006865b1

> bk-bscp-client list cluster
+----------------------------------------+------------------+------------+----------+------------+------------+
|                   ID                   |       NAME       |   STATE    | BUSINESS |    APP     | RCLUSTERID |
+----------------------------------------+------------------+------------+----------+------------+------------+
| C-584a03c6-ccb6-11ea-9cfe-5254006865b1 | cluster-shanghai | AFFECTIVED | X-Game   | gameserver |            |
| C-4169f103-ccb6-11ea-9cfe-5254006865b1 | cluster-shenzhen | AFFECTIVED | X-Game   | gameserver |            |
| C-6fcd9555-ccb4-11ea-9cfe-5254006865b1 | default          | AFFECTIVED | X-Game   | gameserver |            |
+----------------------------------------+------------------+------------+----------+------------+------------+

> bk-bscp-client list cluster --app game 
+----------------------------------------+----------+------------+----------+------+------------+
|                   ID                   |   NAME   |   STATE    | BUSINESS | APP  | RCLUSTERID |
+----------------------------------------+----------+------------+----------+------+------------+
| C-8e520964-c667-11ea-9cfe-5254006865b1 | xian     | AFFECTIVED | X-Game   | game |            |
| C-b071c279-c667-11ea-9cfe-5254006865b1 | shenzhen | AFFECTIVED | X-Game   | game |            |
+----------------------------------------+----------+------------+----------+------+------------+

> bk-bscp-client get cluster --name cluster-shenzhen
Name: 		cluster-shenzhen
ClusterID: 	C-4169f103-ccb6-11ea-9cfe-5254006865b1
Business: 	B-48de67cb-b6d5-11ea-90b2-5254006865b1 - X-Game
Application: 	A-6fcccf08-ccb4-11ea-9cfe-5254006865b1 - gameserver
RClusterid:
State:		Affectived
Memo:
Creator: 	guohu
LastModifyBy: 	guohu
CreatedAt: 	2020-07-23 15:29:34
UpdatedAt: 	2020-07-23 15:33:26

> bk-bscp-client update cluster --id C-4169f103-ccb6-11ea-9cfe-5254006865b1 --name shenzhen
Update resources successfully
```

#### 创建zone

```shell
> bk-bscp-client create zone --cluster cluster-shenzhen --name zone-tel-1
Create Zone successfully: Z-de896c2c-ccb6-11ea-9cfe-5254006865b1

> bk-bscp-client create zone --cluster cluster-shenzhen --name zone-tel-2
Create Zone successfully: Z-e9abb192-ccb6-11ea-9cfe-5254006865b

> bk-bscp-client list zone --cluster cluster-shenzhen
+----------------------------------------+------------+------------+----------+------------+------------------+
|                   ID                   |    NAME    |   STATE    | BUSINESS |    APP     |     CLUSTER      |
+----------------------------------------+------------+------------+----------+------------+------------------+
| Z-e9abb192-ccb6-11ea-9cfe-5254006865b1 | zone-tel-2 | AFFECTIVED | X-Game   | gameserver | cluster-shenzhen |
| Z-de896c2c-ccb6-11ea-9cfe-5254006865b1 | zone-tel-1 | AFFECTIVED | X-Game   | gameserver | cluster-shenzhen |
+----------------------------------------+------------+------------+----------+------------+------------------+

> bk-bscp-client get zone --name zone-tel-2
Name: 		zone-tel-2
ZoneID: 	Z-e9abb192-ccb6-11ea-9cfe-5254006865b1
Business: 	B-48de67cb-b6d5-11ea-90b2-5254006865b1 - X-Game
Application: 	A-6fcccf08-ccb4-11ea-9cfe-5254006865b1 - gameserver
Cluster: 	C-4169f103-ccb6-11ea-9cfe-5254006865b1 - cluster-shenzhen
State:		Affectived
Memo:
Creator: 	guohu
LastModifyBy: 	guohu
CreatedAt: 	2020-07-23 15:34:16
UpdatedAt: 	2020-07-23 15:34:16

> bk-bscp-client update zone --id Z-de896c2c-ccb6-11ea-9cfe-5254006865b1 --name zone-1
Update resources successfully
```

#### 多App同时创建Cluster/Zone

考虑业务模块架构的复杂度，Cluster和Zone虽然有助于配置渲染生成和隔离存储，但是业务模块众多的情况下，逐一创建较为繁琐。
client提供了批量创建的命令。

```shell
# 为App [gameserver], [loginserver], [chatserver]同时创建集群cluster-shenzhen

> bk-bscp-client create cluster-list --name cluster-shenzhen --for-apps gameserver, loginserver --for-apps chatserver

为App [gameserver], [loginserver]下的cluster-shenzhen集群同时创建大区zone-tel-1

> bk-bscp-client create zone-list --name zone-tel-1 --for-apps gameserver, loginserver --cluster cluster-shenzhen

为App [gameserver]下cluster同时创建多个zone

> bk-bscp-client create zone-list --names zone-tel-1, zone-tel-2 --app gameserver --cluster cluster-shenzhen
```

### ConfigSet操作

**创建ConfigSet**

```shell
> bk-bscp-client create cfgset --path /etc --name server.yaml
Create ConfigSet successfully: F-0a403b7a-ccc2-11ea-9cfe-5254006865b1

> bk-bscp-client create cfgset --path /tmp --name local.yaml
Create ConfigSet successfully: F-5f17767b-ccc3-11ea-9cfe-5254006865b1
```

**查看ConfigSet**

```shell
> bk-bscp-client list cfgset
+----------------------------------------+-------+-------------+------------+----------+------------+
|                   ID                   | FPATH |    NAME     |   STATE    | BUSINESS |    APP     |
+----------------------------------------+-------+-------------+------------+----------+------------+
| F-5f17767b-ccc3-11ea-9cfe-5254006865b1 | /tmp  | local.yaml  | AFFECTIVED | X-Game   | gameserver |
| F-0a403b7a-ccc2-11ea-9cfe-5254006865b1 | /etc  | server.yaml | AFFECTIVED | X-Game   | gameserver |
+----------------------------------------+-------+-------------+------------+----------+------------+

> bk-bscp-client get cfgset --path /etc --name server.yaml
ConfigSetId: 	F-0a403b7a-ccc2-11ea-9cfe-5254006865b1
Business: 	B-48de67cb-b6d5-11ea-90b2-5254006865b1 - X-Game
Application: 	A-6fcccf08-ccb4-11ea-9cfe-5254006865b1 - gameserver
Fpath: 		/etc
Name: 		server.yaml
State:		Affectived
Memo:
Creator: 	guohu
LastModifyBy: 	guohu
CreatedAt: 	2020-07-23 16:53:56
UpdatedAt: 	2020-07-23 16:53:56
```

### Strategy 操作

如果进行策略发布，在创建Release时，需要提交Strategy的信息。策略信息需要提前创建。

```shell
> bk-bscp-client create strategy --help
Create strategy for application configuration release

Usage:
  bk-bscp-client create strategy [flags]

Aliases:
  strategy, str

Examples:

	bk-bscp-client create strategy --json ./somefile.json --name strategyName
	json template as followed:
	{
		"App":"appName",
		"Clusters": ["cluster01","cluster02","cluster03"],
		"Zones": ["zone01","zone02","zone03"],
		"Dcs": ["dc01","dc02","dc03"],
		"IPs": ["X.X.X.1","X.X.X.2","X.X.X.3"],
		"Labels": {
			"k1":"v1",
			"k2":"v2",
			"k3":"v3"
		},
		"LabelsAnd": {
			"k3":"1",
			"k4":"1,2,3"
		}
	}

Flags:
  -a, --app string    settings application name that strategy belongs to
  -h, --help          help for strategy
  -j, --json string   json details for strategy
  -m, --memo string   settings memo for strategy
  -n, --name string   settings strategy name.

Global Flags:
      --business string   business Name to operate. Get parameter priority: command -> env -> .bscp/desc
      --operator string   user name for operation.  Get parameter priority: command -> env -> .bscp/desc
      --token string      user token for operation. Get parameter priority: command -> env -> .bscp/desc
```

发布策略描述json文件。相关字段说明：

* Appid：AppID，必填字段
* Clusterids：cluster ID列表，来源于系统，用于过滤sidecar的属性
* Zoneids：zone ID列表，来源系统，用于过滤sidecar的属性
* Dcs：数据中心，该数据用于过滤sidecar属性
* IPs：IP列表，用于过滤sidecar的属性
* Labels：自定义kv段，用于过滤sidecar对应的属性, 'OR' 关系
* LabelsAnd: 自定义kv段，用于过滤sidecar对应的属性, 'AND' 关系

Clusterids、Zoneids、Dcs、IPs、Labels、LabelsAnd 不能同时为空。至少有一项要有值，否则该策略不会命中任何区域, 相关联版本也无法正常下发。

BSCP sidecar启动时，默认需要填入cluster，zone，dc，IP、labels、LabelsAnd，以上字段主要是用于对sidecar进行过滤。

**创建Strategy**

```shell
> bk-bscp-client create strategy --json strategy.json --name shenzhen-strategy
Create Strategy successfully: S-a7841383-cfdf-11ea-9cfe-5254006865b1
```

**查看Strategy**

```
> bk-bscp-client list strategy
+----------------------------------------+-------------------+------------+------+------------+---------+
|                   ID                   |       NAME        |   STATE    | MEMO |    APP     | CREATOR |
+----------------------------------------+-------------------+------------+------+------------+---------+
| S-a7841383-cfdf-11ea-9cfe-5254006865b1 | shenzhen-strategy | AFFECTIVED |      | gameserver | guohu   |
+----------------------------------------+-------------------+------------+------+------------+---------+

> bk-bscp-client get strategy --name shenzhen-strategy
Name: 		shenzhen-strategy-update
StrategyID: 	S-a7841383-cfdf-11ea-9cfe-5254006865b1
App: 		A-6fcccf08-ccb4-11ea-9cfe-5254006865b1 - gameserver
Status:		AFFECTIVED
Memo:
Creator: 	guohu
LastModifyBy: 	guohu
CreatedAt: 	2020-07-27 16:03:28
UpdatedAt: 	2020-07-27 16:03:28
Content:
{
    "App": "gameserver",
    "Clusters": [
        "cluster-shenzhen"
    ],
    "Zones": [
        "zone-tel-2"
    ],
    "Dcs": [
        "dc01"
    ],
    "IPs": [
        "127.0.0.1"
    ],
    "Labels": {
        "k1": "v1",
        "k2": "v2",
        "k3": "v3"
    },
    "LabelsAnd": {
        "k3": "1",
        "k4": "1,2,3"
    }
}
```

### Commit 操作

#### 1. 添加文件到扫描区

支持指定多文件名提交和全部提交两种方式。

```shell
> tree .
.
└── etc
    ├── local.yaml
    └── server.yaml
    
> bk-bscp-client add etc/local.yaml etc/server.yaml

> bk-bscp-client add .
```

**撤销已提交到扫描区的文件**

```shell
> bk-bscp-client checkout etc/local.yaml etc/server.yaml

> bk-bscp-client checkout .
```

**查看扫描区文件列表**

```shell
> bk-bscp-client st
Scan area file list:
  (use "bk-bscp-client checkout <file>..." to remove file from scan area)
  (use "bk-bscp-client commit" to submit the configuration files in the scan area)
	new file:	etc/server.yaml

The local repository is not added to the scan area file list:
  (use "bk-bscp-client add <file>..." to add file to scan area)
	etc/local.yaml
```

#### 2. 提交扫描区中的文件

如果 commit 提交时，扫描区文件对应的 ConfigSet 没有创建会自动创建，如果已经创建会自动选择已经创建的 ConfigSet 相关联。

**扫描区文件与ConfigSet关联关系：**

etc/server.yaml 对应 ConfigSet (path=/etc，name=server.yaml)

```shell
#确认配置文件(示例，实际请遵守YAML规则)
> cat etc/server.yaml
<侠客行> 唐·李白
赵客缦胡缨，吴钩霜雪明。银鞍照白马，飒沓如流星。
十步杀一人，千里不留行。事了拂衣去，深藏身与名。
闲过信陵饮，脱剑膝前横。将炙啖朱亥，持觞劝侯嬴。
三杯吐然诺，五岳倒为轻。眼花耳热后，意气素霓生。
救赵挥金锤，邯郸先震惊。千秋二壮士，烜赫大梁城。
纵死侠骨香，不惭世上英。谁能书阁下，白首太玄经。

> bk-bscp-client commit --memo "this is a example"
Commit successfully! commitid: MM-b115d865-ccc5-11ea-9cfe-5254006865b1

	(use "bk-bscp-client get commit --id <commitid>" to get commit detail)
	(use "bk-bscp-client release --name <releaseName> --commitid <commitid> --strategy <strategyName> --memo "this is a example"" to create release to publish)

```

**查看Commit**

```shell
# 查询提交记录
> bk-bscp-client list commit
+-----------------------------------------+--------------+-----------+-------------------+----------+------------+---------+
|                COMMITID                 |  RELEASEID   |   STATE   |       MEMO        | BUSINESS |    APP     | CREATOR |
+-----------------------------------------+--------------+-----------+-------------------+----------+------------+---------+
| MM-b115d865-ccc5-11ea-9cfe-5254006865b1 | Not Released | CONFIRMED | this is a example | X-Game   | gameserver | guohu   |
+-----------------------------------------+--------------+-----------+-------------------+----------+------------+---------+

	(use "bk-bscp-client get commit --id <commitid>" to get commit detail)
	(use "bk-bscp-client release --name <releaseName> --commitid <commitid> --strategy <strategyName> --memo "this is a example"" to create release to publish)

# 查看commit具体详情
> bk-bscp-client get commit --id MM-b115d865-ccc5-11ea-9cfe-5254006865b1
CommitID: 	MM-b115d865-ccc5-11ea-9cfe-5254006865b1
Business: 	B-48de67cb-b6d5-11ea-90b2-5254006865b1  -  X-Game
Application: 	A-6fcccf08-ccb4-11ea-9cfe-5254006865b1  -  gameserver
ReleaseID: 	MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1
State:		CONFIRMED
Memo: 		this is a example
Creator: 	guohu
CreatedAt: 	2020-07-23 17:20:04
UpdatedAt: 	2020-07-23 19:13:10
Metadatas:
+----------------------------------------+------------------+----------+--------------+
|                MODULEID                |      CFGSET      | TEMPLATE | TEMPLATERULE |
+----------------------------------------+------------------+----------+--------------+
| M-b1166157-ccc5-11ea-9cfe-5254006865b1 | /etc/server.yaml |          |              |
+----------------------------------------+------------------+----------+--------------+

	(use "bk-bscp-client get commit --mid <moduleId>" to get commit module detail)
	(use "bk-bscp-client release --name <releaseName> --commitid <commitid> --strategy <strategyName> --memo "this is a example"" to create release to publish)

# 查看commit模块具体详情
> bk-bscp-client get commit --mid M-b1166157-ccc5-11ea-9cfe-5254006865b1
ModuleID: 	M-b1166157-ccc5-11ea-9cfe-5254006865b1
CommitID: 	MM-b115d865-ccc5-11ea-9cfe-5254006865b1
Business: 	B-48de67cb-b6d5-11ea-90b2-5254006865b1  -  X-Game
Application: 	A-6fcccf08-ccb4-11ea-9cfe-5254006865b1  -  gameserver
ConfigSet: 	F-0a403b7a-ccc2-11ea-9cfe-5254006865b1  -  /etc/server.yaml
ReleaseID: 	Not Released
State:		CONFIRMED
Creator: 	guohu
CreatedAt: 	2020-07-23 17:20:04
UpdatedAt: 	2020-07-23 17:20:04
Configs:
    <侠客行> 唐·李白
    赵客缦胡缨，吴钩霜雪明。银鞍照白马，飒沓如流星。
    十步杀一人，千里不留行。事了拂衣去，深藏身与名。
    闲过信陵饮，脱剑膝前横。将炙啖朱亥，持觞劝侯嬴。
    三杯吐然诺，五岳倒为轻。眼花耳热后，意气素霓生。
    救赵挥金锤，邯郸先震惊。千秋二壮士，烜赫大梁城。
    纵死侠骨香，不惭世上英。谁能书阁下，白首太玄经。


	(use "bk-bscp-client release --name <releaseName> --commitid <commitid> --strategy <strategyName> --memo "this is a example"" to create release to publish)
```

### Release操作

**`无发布策略`**

针对配置发布，可以定制相关发布策略。如果不做任何发布策略，则发布的配置认为是Application(app)全局配置，
所有的App都会启用该配置。

```shell
> bk-bscp-client list release
Found no Release resource.

> bk-bscp-client list commit
+-----------------------------------------+--------------+-----------+-------------------+----------+------------+---------+
|                COMMITID                 |  RELEASEID   |   STATE   |       MEMO        | BUSINESS |    APP     | CREATOR |
+-----------------------------------------+--------------+-----------+-------------------+----------+------------+---------+
| MM-b115d865-ccc5-11ea-9cfe-5254006865b1 | Not Released | CONFIRMED | this is a example | X-Game   | gameserver | guohu   |
+-----------------------------------------+--------------+-----------+-------------------+----------+------------+---------+

	(use "bk-bscp-client get commit --id <commitid>" to get commit detail)
	(use "bk-bscp-client release --name <releaseName> --commitid <commitid> --strategy <strategyName> --memo "this is a example"" to create release to publish)

# 创建release
> bk-bscp-client release --commitid MM-b115d865-ccc5-11ea-9cfe-5254006865b1 --name new-release-v1
Create Release successfully: MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1

	 (use "bk-bscp-client get release --id <releaseid>" to get release detail)
	（use "bk-bscp-client publish --id <releaseid>" to confrim release to publish)

```

**查询release**

```shell
# 查询版本记录
> bk-bscp-client list release
+-----------------------------------------+----------------+-------+--------------+----------+------------+---------+
|                   ID                    |      NAME      | STATE |   STRATEGY   | BUSINESS |    APP     | CREATOR |
+-----------------------------------------+----------------+-------+--------------+----------+------------+---------+
| MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1 | new-release-v1 | INIT  | No Strategy! | X-Game   | gameserver | guohu   |
+-----------------------------------------+----------------+-------+--------------+----------+------------+---------+

	(use "bk-bscp-client get release --id <releaseid>" to get release detail)
	(use "bk-bscp-client publish --id <releaseid>" to confrim release to publish)

# 查看release具体详情
> bk-bscp-client get release --id MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1
Name: 			new-release-v1
ReleaseID: 		MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1
CommitID: 		MM-b115d865-ccc5-11ea-9cfe-5254006865b1
Business: 		B-48de67cb-b6d5-11ea-90b2-5254006865b1  -  X-Game
Application: 	A-6fcccf08-ccb4-11ea-9cfe-5254006865b1  -  gameserver
Strategy: 		No Strategy!
State:			INIT
Memo:
Creator: 		guohu
LastModifyBy:   guohu
CreatedAt: 		2020-07-23 19:05:31
UpdatedAt: 		2020-07-23 19:05:31
Metadatas:
+----------------------------------------+------------------+----------------------------------------+
|                MODULEID                |      CFGSET      |             COMMITMODULEID             |
+----------------------------------------+------------------+----------------------------------------+
| R-6c6cc301-ccd4-11ea-9cfe-5254006865b1 | /etc/server.yaml | M-b1166157-ccc5-11ea-9cfe-5254006865b1 |
+----------------------------------------+------------------+----------------------------------------+

	（use "bk-bscp-client get release --mid <moduleId>" to get release detail）
	（use "bk-bscp-client publish --id <releaseid>" to confrim release to publish）
	
# 查看release模块具体详情
> bk-bscp-client get release --mid R-6c6cc301-ccd4-11ea-9cfe-5254006865b1
Name: 		new-release-v1
ModuleId: 	R-6c6cc301-ccd4-11ea-9cfe-5254006865b1
ReleaseID:
CommitModuleID: M-b1166157-ccc5-11ea-9cfe-5254006865b1
Business: 	B-48de67cb-b6d5-11ea-90b2-5254006865b1  -  X-Game
Application: 	A-6fcccf08-ccb4-11ea-9cfe-5254006865b1  -  gameserver
ConfigSet: 	F-0a403b7a-ccc2-11ea-9cfe-5254006865b1  -  /etc/server.yaml
Strategy: 	No Strategy!
State:		INIT
Creator: 	guohu
LastModifyBy: 	guohu
CreatedAt: 	2020-07-23 19:05:31
UpdatedAt: 	2020-07-23 19:05:31

	（use "bk-bscp-client publish --id <releaseid>" to confrim release to publish)
	
```

**`策略发布`**

```shell
#release 关联策略
> bk-bscp-client release --commitid MM-b115d865-ccc5-11ea-9cfe-5254006865b1 --name new-release-v1 --strategy shenzhen-strategy
Create Release successfully: MR-cb4b0010-ccd8-11ea-9cfe-5254006865b1

	 (use "bk-bscp-client get release --id <releaseid>" to get release detail)
	（use "bk-bscp-client publish --id <releaseid>" to confrim release to publish)
```

### Publish操作

release创建之后，还未生效，只是生成版本记录。使用 publish 命令指定 release 进行发布，BSCP会将具体配置发布至远端App实例。

```shell
> bk-bscp-client publish --id MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1
Publish successfully: MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1

	use "bk-bscp-client get release --id <releaseid>" to get release detail
> bk-bscp-client list release
+-----------------------------------------+----------------+-----------+-------------------+----------+------------+---------+
|                   ID                    |      NAME      |   STATE   |     STRATEGY      | BUSINESS |    APP     | CREATOR |
+-----------------------------------------+----------------+-----------+-------------------+----------+------------+---------+
| MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1 | new-release-v1 | PUBLISHED | No Strategy!      | X-Game   | gameserver | guohu   |
+-----------------------------------------+----------------+-----------+-------------------+----------+------------+---------+
```

### Rollbcak操作

rollback的命令集支持指定版本号，进行回滚操作，回滚之后会重新发布指定 release 。

```shell
> bk-bscp-client rollback --id MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1
Rollback successfully: MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1

> bk-bscp-client list release
+-----------------------------------------+----------------+------------+-------------------+----------+------------+---------+
|                   ID                    |      NAME      |   STATE    |     STRATEGY      | BUSINESS |    APP     | CREATOR |
+-----------------------------------------+----------------+------------+-------------------+----------+------------+---------+
| MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1 | new-release-v1 | ROLLBACKED | No Strategy!      | X-Game   | gameserver | guohu   |
+-----------------------------------------+----------------+------------+-------------------+----------+------------+---------+
```

### Reload操作

重新加载，支持对已发布未生效的版本和已回滚的版本进行重新加载。

```shell
> bk-bscp-client reload --id MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1
Reload successfully: MR-6c6c3579-ccd4-11ea-9cfe-5254006865b1
```

### Instance操作

**通过 app、cluster、zone、实例状态的实例列表过滤查询（默认只显示在线实例，可以通过 status 选择实例状态）**

```shell
> bk-bscp-client list inst --cluster shenzhen --status 1
+----+-------------+----------+------+----------+--------+------------+--------+--------+
| ID |     IP      | BUSINESS | APP  | CLUSTER  |  ZONE  | DATACENTER | LABLES | STATE  |
+----+-------------+----------+------+----------+--------+------------+--------+--------+
|  2 |  127.0.0.1  | X-Game   | game | shenzhen | zone-1 | bscp_xgame | null   | Online |
+----+-------------+----------+------+----------+--------+------------+--------+--------+
```

**通过 strategy 查看发布策略可触达的实例列表**

```shell
> bk-bscp-client list inst --strategyid S-719cbd79-c66c-11ea-9cfe-5254006865b1
+----+-------------+----------+------+----------+--------+------------+--------+--------+
| ID |     IP      | BUSINESS | APP  | CLUSTER  |  ZONE  | DATACENTER | LABLES | STATE  |
+----+-------------+----------+------+----------+--------+------------+--------+--------+
|  2 |  127.0.0.1  | X-Game   | game | shenzhen | zone-1 | bscp_xgame | null   | Online |
+----+-------------+----------+------+----------+--------+------------+--------+--------+
```

**通过 release 查看各状态的实例列表**

可以在发布前通过该命令查看可触达的实例，发布后查看实例具体生效状态（默认只显示在线实例，可以通过 status 选择实例状态）。

```shell
> bk-bscp-client list inst --releaseid MR-cd9c8e09-cbfc-11ea-9cfe-5254006865b1
+----+-------------+----------+------+----------+--------+------------+--------+--------+--------------------------------+
| ID |     IP      | BUSINESS | APP  | CLUSTER  |  ZONE  | DATACENTER | LABLES | STATE  |      EFFECTSTATUS - TIME       |
+----+-------------+----------+------+----------+--------+------------+--------+--------+--------------------------------+
|  2 |  127.0.0.1  | X-Game   | game | shenzhen | zone-1 | bscp_xgame | null   | Online | RollBackReload - 2020-07-24    |
|    |             |          |      |          |        |            |        |        | 14:52:27                       |
+----+-------------+----------+------+----------+--------+------------+--------+--------+--------------------------------+
    ConfigSet: /etc/local.yaml    OnlineReloadInstance: 1    OnlineEffectInstance: 0    OnlineUnEffectInstance: 0
			OfflineReloadInstance: 0    OfflineEffectInstance: 0

+----+-------------+----------+------+----------+--------+------------+--------+--------+--------------------------------+
| ID |     IP      | BUSINESS | APP  | CLUSTER  |  ZONE  | DATACENTER | LABLES | STATE  |      EFFECTSTATUS - TIME       |
+----+-------------+----------+------+----------+--------+------------+--------+--------+--------------------------------+
|  2 |  127.0.0.1  | X-Game   | game | shenzhen | zone-1 | bscp_xgame | null   | Online | RollBackReload - 2020-07-24    |
|    |             |          |      |          |        |            |        |        | 14:52:27                       |
+----+-------------+----------+------+----------+--------+------------+--------+--------+--------------------------------+
    ConfigSet: /etc/server.yaml    OnlineReloadInstance: 1    OnlineEffectInstance: 0    OnlineUnEffectInstance: 0
			OfflineReloadInstance: 0    OfflineEffectInstance: 0
```