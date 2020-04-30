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

**Commit**：针对ConfigSet内容的变化的提交，需要提交具体配置内容或者配置模板。可以连续多次提交，可以选取其中的某次提
交确认生效。非生效的Commit仅做流水记录，用于历史查阅。

**Release**：配置发布，关联已生效Commit，并关联发布策略，例如需要发布到某个cluster，zone，或者匹配节点自定义的kv
对，说明此次发布细节。提交的release也需要确认(confirm)才能正式发布到目标终端。

## bscp-client安装

client配置文件，client连入BSCP后台服务需要相关配合。默认配置路径`/etc/bscp/client.yaml`。该参数可以通过client
的参数--configfile进行指定。

```yaml
kind: bscp-client
version: 0.1.1
access:
  servicename: bk-bscp-accessserver
  timeout: 3s
etcd:
  endpoints:
    - 127.0.0.1:2379
  dialtimeout: 3s
```

配置简单说明：
* etcd：client通过BSCP系统的etcd完成服务发现。需要填写etcd IP列表。
* access：BSCP系统接入层。通过名字进行服务发现。

配置完成即可进行client使用。

```shell
$ bk-bscp-client --help
bscp-client controls the BlueKing Service Configuration Platform.

Usage:
  bscp-client [command]

Available Commands:
  confirm     confirm resource like commit/release
  create      create new resource
  delete      delete resource
  get         get specified resource
  help        Help about any command
  list        list resources
  lock        lock resource
  unlock      unlock resource
  update      update resource

Flags:
      --business string     Business Name to operate. Also comes from ENV BSCP_BUSINESS
      --configfile string   BlueKing Service Configuration Platform CLI configuration. (default "/etc/bscp/client.yaml")
  -h, --help                help for bscp-client
      --operator string     user name for operation, use for audit, Also comes from ENV BSCP_OPERATOR
      --version             version for bscp-client

Use "bscp-client [command] --help" for more information about a command.
```

client默认有三个全局的命令行参数：
* configfile：client接入BSCP的相关配置。建议采用默认路径，简化命令行输入；
* business：需要操作业务名称，所有操作命令都需要输入该参数。该参数也可以会读取环境变量BSCP_BUSINESS，为简化输入，建议使用环境变量输入。
* operator：命令操作者，为实现命令审计，所有操作命令都需要输入操作者英文ID。该参数也可以会读取环境变量BSCP_OPERATOR，为简化输入，建议使用环境变量输入。

**使用案例说明**

```shell
#设置必须的命令行参数
> export BSCP_BUSINESS=X-Game
> export BSCP_OPERATOR=MrMGXXXX

#查看当前业务下模块(app)列表
> bk-bscp-client list app
+----------------------------------------+--------------+----------+------+---------+-------+--------------+
|                   ID                   |      NAME    | BUSINESS | TYPE | CREATOR | STATE | LASTMODIFYBY |
+----------------------------------------+--------------+----------+------+---------+-------+--------------+
| A-33bd6ae7-cee0-11e9-9e47-5254000ea971 |  gameserver  |  X-Game  |    0 | MrMGXXXX|     0 |   MrMGXXXX   |
+----------------------------------------+--------------+----------+------+---------+-------+--------------+
```

## 业务配置命令

业务配置命令，主要对BSCP用户开放，主要涉及类型如下：
* 创建/查看Application(app)
* 创建/查看app下的Cluster、Zone
* 创建/查看app关联的ConfigSet
* 创建/查看/更新Commit
* 创建/查看Strategy
* 创建/查看/更新Release

### Application操作

```shell
export BSCP_BUSINESS=X-Game
export BSCP_OPERATOR=MrMGXXXX

> bk-bscp-client create app --name gameserver
create Application A-eed8cbcf-cfb7-11e9-8163-5254000ea971 successfully.
> bk-bscp-client list app
+----------------------------------------+------------+----------+------+---------+-------+--------------+
|                    ID                  |    NAME    | BUSINESS | TYPE | CREATOR | STATE | LASTMODIFYBY |
+----------------------------------------+------------+----------+------+---------+-------+--------------+
| A-eed8cbcf-cfb7-11e9-8163-5254000ea971 | gameserver |  X-Game  |    0 | MrMGXXXX|     0 |  MrMGXXXX    |
+----------------------------------------+------------+----------+------+---------+-------+--------------+
```

### Cluster和Zone操作

Cluster和Zone是为了定义业务逻辑架构，便于在配置模板渲染，生成和存储时能为业务不同的逻辑架单元
生成隔离独立的配置。建议业务在Application创建之后就完成该逻辑层次划分。

**创建cluster**

```shell
> bk-bscp-client create cluster --app gameserver --name cluster-shenzhen
create Cluster C-d9c129a9-cfe1-11e9-8163-5254000ea971/cluster-shenzhen successfully.

> bk-bscp-client create cluster --app gameserver --name cluster-shanghai
create Cluster C-e2ebd517-cfe1-11e9-8163-5254000ea971/cluster-shanghai successfully.

> bk-bscp-client list cluster --app gameserver
+----------------------------------------+------------------+----------+------------+---------+-------+--------------+------------+
|                   ID                   |       NAME       | BUSINESS |     APP    | CREATOR | STATE | LASTMODIFYBY | RCLUSTERID |
+----------------------------------------+------------------+----------+------------+---------+-------+--------------+------------+
| C-e2ebd517-cfe1-11e9-8163-5254000ea971 | cluster-shanghai |  X-Game  | gameserver | MrMGXXXX|     0 | MrMGXXXX     |            |
| C-d9c129a9-cfe1-11e9-8163-5254000ea971 | cluster-shenzhen |  X-Game  | gameserver | MrMGXXXX|     0 | MrMGXXXX     |            |
+----------------------------------------+------------------+----------+------------+---------+-------+--------------+------------+
```

**创建zone**

```shell
> bk-bscp-client create zone --app gameserver --cluster cluster-shenzhen --name zone-tel-1
create Zone Z-894595a0-cfe3-11e9-8163-5254000ea971/zone-tel-1 successfully.

> bk-bscp-client create zone --app gameserver --cluster cluster-shenzhen --name zone-tel-2
create Zone Z-90da0f19-cfe3-11e9-8163-5254000ea971/zone-tel-2 successfully.

> bk-bscp-client create zone --app gameserver --cluster cluster-shanghai --name zone-mob-1
create Zone Z-a0f88189-cfe3-11e9-8163-5254000ea971/zone-mob-1 successfully.

> bk-bscp-client create zone --app gameserver --cluster cluster-shanghai --name zone-mob-2
create Zone Z-a29276aa-cfe3-11e9-8163-5254000ea971/zone-mob-2 successfully.

> bk-bscp-client list zone --app gameserver --cluster cluster-shenzhen
+----------------------------------------+------------+----------+------------+------------------+---------+-------+--------------+---------------------+
|                   ID                   |    NAME    | BUSINESS |     APP    |     CLUSTER      | CREATOR | STATE | LASTMODIFYBY |      UPDATEAT       |
+----------------------------------------+------------+----------+------------+------------------+---------+-------+--------------+---------------------+
| Z-894595a0-cfe3-11e9-8163-5254000ea971 | zone-tel-1 | X-Game   | gameserver | cluster-shenzhen | MrMGXXXX|     0 | MrMGXXXX     | 2019-09-05 21:46:31 |
| Z-90da0f19-cfe3-11e9-8163-5254000ea971 | zone-tel-2 | X-Game   | gameserver | cluster-shenzhen | MrMGXXXX|     0 | MrMGXXXX     | 2019-09-05 21:46:18 |
+----------------------------------------+------------+----------+------------+------------------+---------+-------+--------------+---------------------+

> bk-bscp-client list zone --app gameserver --cluster cluster-shanghai
+----------------------------------------+------------+----------+------------+------------------+---------+-------+--------------+---------------------+
|                    ID                  |    NAME    | BUSINESS |     APP    |     CLUSTER      | CREATOR | STATE | LASTMODIFYBY |      UPDATEAT       |
+----------------------------------------+------------+----------+------------+------------------+---------+-------+--------------+---------------------+
| Z-a0f88189-cfe3-11e9-8163-5254000ea971 | zone-mob-1 | X-Game   | gameserver | cluster-shanghai |MrMGXXXX |     0 | MrMGXXXX     | 2019-09-05 21:47:00 |
| Z-a29276aa-cfe3-11e9-8163-5254000ea971 | zone-mob-2 | X-Game   | gameserver | cluster-shanghai |MrMGXXXX |     0 | MrMGXXXX     | 2019-09-05 21:46:58 |
+----------------------------------------+------------+----------+------------+------------------+---------+-------+--------------+---------------------+
```

**多App同时创建Cluster/Zone**

考虑业务模块架构的复杂度，Cluster和Zone虽然有助于配置渲染生成和隔离存储，但是业务模块众多的情况下，逐一创建较为繁琐。
client提供了批量创建的命令。

```shell
为App [gameserver], [loginserver], [chatserver]同时创建集群cluster-shenzhen

> bk-bscp-client create cluster-list --name cluster-shenzhen --for-apps gameserver, loginserver --for-apps chatserver

为App [gameserver], [loginserver]下的cluster-shenzhen集群同时创建大区zone-tel-1

> bk-bscp-client create zone-list --name zone-tel-1 --for-apps gameserver, loginserver --cluster cluster-shenzhen

为App [gameserver]下cluster同时创建多个zone

> bk-bscp-client create zone-list --names zone-tel-1, zone-tel-2 --app gameserver --cluster cluster-shenzhen
```

### ConfigSet与Commit操作

**创建ConfigSet**

```shell
> export BSCP_BUSINESS=X-Game
> export BSCP_OPERATOR=MrMGXXXX

> bk-bscp-client create cfgset --name server.yaml --app gameserver
Create ConfigSet successfully: F-be31a375-d090-11e9-a11f-5254000ea971

> bk-bscp-client create cfgset --name local.yaml --app gameserver
Create ConfigSet successfully: F-ca99342b-d090-11e9-a11f-5254000ea971
```

**查看ConfigSet**

```shell
> bk-bscp-client list cfgset --app gameserver
+----------------------------------------+-------------+----------+------------+---------+-------+--------------+---------------------+
|                    ID                  |     NAME    | BUSINESS |     APP    | CREATOR | STATE | LASTMODIFYBY |      UPDATEAT       |
+----------------------------------------+-------------+----------+------------+---------+-------+--------------+---------------------+
| F-ca99342b-d090-11e9-a11f-5254000ea971 | local.yaml  | X-Game   | gameserver | MrMGXXXX|     0 | MrMGXXXX     | 2019-09-06 18:26:30 |
| F-be31a375-d090-11e9-a11f-5254000ea971 | server.yaml | X-Game   | gameserver | MrMGXXXX|     0 | MrMGXXXX     | 2019-09-06 18:26:10 |
+----------------------------------------+-------------+----------+------------+---------+-------+--------------+---------------------+
```

**提交Commit**

提交commit是生成具体配置文件的第一步，提交Commit有两种形式：
* `迭代中`提交模板，通过预设定的变量针对不同的cluster与zone生成不同的配置实例
* 直接提交配置内容文件，该文件所有App实例共用

```shell
#设置基础环境
> export BSCP_BUSINESS=X-Game
> export BSCP_OPERATOR=MrMGXXXX

#确认配置文件(示例，实际请遵守YAML规则)
> cat server.yaml
<侠客行> 唐·李白
赵客缦胡缨，吴钩霜雪明。银鞍照白马，飒沓如流星。
十步杀一人，千里不留行。事了拂衣去，深藏身与名。
闲过信陵饮，脱剑膝前横。将炙啖朱亥，持觞劝侯嬴。
三杯吐然诺，五岳倒为轻。眼花耳热后，意气素霓生。
救赵挥金锤，邯郸先震惊。千秋二壮士，烜赫大梁城。
纵死侠骨香，不惭世上英。谁能书阁下，白首太玄经。

#为指定ConfigSet创建Commit
> bk-bscp-client create commit --app gameserver --cfgset server.yaml --config-file ./new-release-server.yaml
Create Commit successfully: M-2ef13220-d142-11e9-a11f-5254000ea971
```

**查看Commit**

```shell
#查看Commit
> bk-bscp-client list commit --app gameserver --cfgset server.yaml
+----------------------------------------+----------+-------------+-------------+---------+-------+---------------------+---------------------+
|                    ID                  | BUSINESS | APPLICATION |  CONFIGSET  | CREATOR | STATE |      CREATEDAT      |      UPDATEDAT      |
+----------------------------------------+----------+-------------+-------------+---------+-------+---------------------+---------------------+
| M-2ef13220-d142-11e9-a11f-5254000ea971 | X-Game   | gameserver  | server.yaml |MrMGXXXX |     0 | 2019-09-07 15:36:20 | 2019-09-07 15:36:20 |
+----------------------------------------+----------+-------------+-------------+---------+-------+---------------------+---------------------+

#查看具体详情
> bk-bscp-client get commit --Id M-2ef13220-d142-11e9-a11f-5254000ea971
CommitID: M-2ef13220-d142-11e9-a11f-5254000ea971
BusinessID: B-b9f8492c-cedf-11e9-9e47-5254000ea971
AppID: A-eed8cbcf-cfb7-11e9-8163-5254000ea971
ConfigSetID: F-be31a375-d090-11e9-a11f-5254000ea971
Creator: MrMGXXXX
ReleaseID: Not Released
State: 0
CreatedAt: 2019-09-07 15:36:20
UpdatedAt: 2019-09-07 15:36:20
PrevConfigs:

Configs:
    <侠客行> 唐·李白
    赵客缦胡缨，吴钩霜雪明。银鞍照白马，飒沓如流星。
    十步杀一人，千里不留行。事了拂衣去，深藏身与名。
    闲过信陵饮，脱剑膝前横。将炙啖朱亥，持觞劝侯嬴。
    三杯吐然诺，五岳倒为轻。眼花耳热后，意气素霓生。
    救赵挥金锤，邯郸先震惊。千秋二壮士，烜赫大梁城。
    纵死侠骨香，不惭世上英。谁能书阁下，白首太玄经。

Changes:

```

**Commit状态**

当前Commit仍然处于编辑状态，如果提交的是模板，当前模板仍然未开始渲染。我们针对文件内容第一行`增加空格`在重新提交。

```shell
> bk-bscp-client update commit --Id M-2ef13220-d142-11e9-a11f-5254000ea971 --config-file ./new-release-server.yaml
Update Commit successfully: M-2ef13220-d142-11e9-a11f-5254000ea971

> bk-bscp-client get commit --Id M-2ef13220-d142-11e9-a11f-5254000ea971
CommitID: M-2ef13220-d142-11e9-a11f-5254000ea971
BusinessID: B-b9f8492c-cedf-11e9-9e47-5254000ea971/melobu
AppID: A-eed8cbcf-cfb7-11e9-8163-5254000ea971
ConfigSetID: F-be31a375-d090-11e9-a11f-5254000ea971
Creator: MrMGXXXX
ReleaseID: Not Released
State: 0
CreatedAt: 2019-09-07 15:36:20
UpdatedAt: 2019-09-07 16:20:24
PrevConfigs:

Configs:
        <侠客行> 唐·李白
    赵客缦胡缨，吴钩霜雪明。银鞍照白马，飒沓如流星。
    十步杀一人，千里不留行。事了拂衣去，深藏身与名。
    闲过信陵饮，脱剑膝前横。将炙啖朱亥，持觞劝侯嬴。
    三杯吐然诺，五岳倒为轻。眼花耳热后，意气素霓生。
    救赵挥金锤，邯郸先震惊。千秋二壮士，烜赫大梁城。
    纵死侠骨香，不惭世上英。谁能书阁下，白首太玄经。

Changes:


> bk-bscp-client confirm commit --Id M-2ef13220-d142-11e9-a11f-5254000ea971
Confirm Commit successfully: M-2ef13220-d142-11e9-a11f-5254000ea971

```

Commit confirm之后：
* 如果是纯配置内容，则已经生效，无法再进行update操作，如果需要继续调整内容，需要提交新的Commit。
* 如果提交的是模板，则已经开始进行渲染，亦无法进行Update，如需调整重新提交Commit

**持续Commit**

### Release配置发布

**`无发布策略`**

针对配置发布，可以定制相关发布策略。如果不做任何发布策略，则发布的配置认为是Application(app)全局配置，
所有的App都会启用该配置。

```shell
> bk-bscp-client list release --app gameserver --cfgset server.yaml
Found no Release resource.

> bk-bscp-client list commit --app gameserver --cfgset server.yaml
+----------------------------------------+----------+-------------+-------------+---------+-------+---------------------+---------------------+
|                   ID                   | BUSINESS | APPLICATION |  CONFIGSET  | CREATOR | STATE |      CREATEDAT      |      UPDATEDAT      |
+----------------------------------------+----------+-------------+-------------+---------+-------+---------------------+---------------------+
| M-2ef13220-d142-11e9-a11f-5254000ea971 | X-Game   | gameserver  | server.yaml |MrMGXXXX |     1 | 2019-09-07 15:36:20 | 2019-09-07 16:22:41 |
+----------------------------------------+----------+-------------+-------------+---------+-------+---------------------+---------------------+

# 创建release
> bk-bscp-client create release --app gameserver --commitId M-2ef13220-d142-11e9-a11f-5254000ea971 --name new-release-v1
Create Release successfully: R-85a74cc6-d14b-11e9-a11f-5254000ea971

> bk-bscp-client list release --app gameserver --cfgset server.yaml
+----------------------------------------+------------------+----------+------------+-------------+----------------------------------------+-------+---------+---------------------+--------------+---------------------+
|                   ID                   |      NAME        | BUSINESS |     APP    |  CONFIGSET  |                 COMMITID               | STATE | CREATOR |      CREATEDAT      | LASTMODIFYBY |      UPDATEDAT      |
+----------------------------------------+------------------+----------+------------+-------------+----------------------------------------+-------+---------+---------------------+--------------+---------------------+
| R-85a74cc6-d14b-11e9-a11f-5254000ea971 | new-release-v1   | X-Game   | gameserver | server.yaml | M-2ef13220-d142-11e9-a11f-5254000ea971 |     0 |MrMGXXXX | 2019-09-07 16:43:11 | MrMGXXXX     | 2019-09-07 16:43:11 |
+----------------------------------------+------------------+----------+------------+-------------+----------------------------------------+-------+---------+---------------------+--------------+---------------------+

> bk-bscp-client get release --Id R-85a74cc6-d14b-11e9-a11f-5254000ea971
ReleaseID: R-85a74cc6-d14b-11e9-a11f-5254000ea971
Name: new-release-v1
BusinessID: B-b9f8492c-cedf-11e9-9e47-5254000ea971
AppID: A-eed8cbcf-cfb7-11e9-8163-5254000ea971
State: 0
ConfigSetID: F-be31a375-d090-11e9-a11f-5254000ea971
ConfigSetName: server.yaml
CommitID: M-2ef13220-d142-11e9-a11f-5254000ea971
StrategyID:
Strategies:
    {}
Creator: MrMGXXXX
CreatedAt: 2019-09-07 16:43:11
LastModifyBy: MrMGXXXX
UpdatedAt: 2019-09-07 16:43:11
```

**`release状态`**

与Commit类似，release被创建之后，在未confirm之前，都处于可编辑状态，可以随时调整发布策略（如有）。
一旦进行confirm之后，release则会进入生效状态，BSCP会将具体配置发布(publish)至远端App实例。

```shell
> bk-bscp-client confirm release --Id R-85a74cc6-d14b-11e9-a11f-5254000ea971
release R-85a74cc6-d14b-11e9-a11f-5254000ea971 confirm to publish successfully
```

**`策略发布`**

如果进行策略发布，在创建/更新Release时，需要提交Strategy的信息。策略信息需要提前创建。

```shell
bk-bscp-client create strategy --help
create strategy for application configuration release

Usage:
  bscp-client create strategy [flags]

Examples:

	bscp-client create strategy --app gameserver --json ./somefile.json --name bluestrategy
	json template as followed:
		{
			"Appid":"appid",
			"Clusterids": ["clusterid01","clusterid02","clusterid03"],
			"Zoneids": ["zoneid01","zoneid02","zoneid03"],
			"Dcs": ["dc01","dc02","dc03"],
			"IPs": ["X.X.X.1","X.X.X.2","X.X.X.3"],
			"Labels": {
				"k1":"v1",
				"k2":"v2",
				"k3":"v3"
			}
		}


Flags:
  -a, --app string    settings application name that strategy belongs to
  -h, --help          help for strategy
  -j, --json string   json details for strategy

Global Flags:
      --business string     Business Name to operate. Also comes from ENV BSCP_BUSINESS
      --configfile string   BlueKing Service Configuration Platform CLI configuration. (default "/etc/bscp/client.yaml")
      --operator string     user name for operation, use for audit, Also comes from ENV BSCP_OPERATOR
```

发布策略描述json文件。相关字段说明：

* Appid：AppID，必填字段
* Clusterids：cluster ID列表，来源于系统，用于过滤sidecar的属性
* Zoneids：zone ID列表，来源系统，用于过滤sidecar的属性
* Dcs：数据中心，该数据用于过滤sidecar属性
* IPs：IP列表，用于过滤sidecar的属性
* Labels：自定义kv段，用于过滤sidecar对应的属性

Clusterids、Zoneids、Dcs、IPs、Labels不能同时为空。至少有一项要有值，否则该策略不会命中任何区域, 相关联版本也无法正常下发。

BSCP sidecar启动时，默认需要填入cluster，zone，dc，IP和labels，以上字段主要是用于对sidecar进行过滤。

```shell

#创建策略
> bscp-client create strategy --app gameserver --json ./shenzhen-strategy.json --name shenzhen-strategy

#release 关联策略
> bk-bscp-client create release --app gameserver --cfgset server.yaml --commitId M-2ef13220-d142-11e9-a11f-5254000ea971 --name new-release-v1 --strategy shenzhen-strategy
```

## 服务配置平台sidecar使用

为方便容器集成，解决容器迁移问题，BSCP为容器环境提供了sidecar。BSCP sidecar与业务容器组成一个Pod，协助使用业务将配置实时拉取至容器本地。

bscp sidecar启动说明，sidecar启动需要以下环境变量完成系统接入：

* BSCP_BCSSIDECAR_APPCFG_PATH: 应用配置生效路径，例如/app/(默认)，命中发布规则的cfgset会以文件形式存储在该目录中
* BSCP_BCSSIDECAR_APPINFO_BUSINESS：sidecar所在Pod所属业务名，必填
* BSCP_BCSSIDECAR_APPINFO_APP：sidecar所在Pod所属应用名，必填
* BSCP_BCSSIDECAR_APPINFO_CLUSTER: sidecar所在Pod所属集群名称，必填
* BSCP_BCSSIDECAR_APPINFO_ZONE: Sidecar所在Pod所属大区名称，必填
* BSCP_BCSSIDECAR_APPINFO_DC: Sidecar所在Pod物理机房标识，用作内网IP命名空间，必填

sidecar自动注入功能需要Deployment或者Application在annotation中加入：

>

其他相关可选值

|                  环境变量名                     |          默认值         |                       备注                       |
| :---------------------------------------------- | ---------------------:  | :----------------------------------------------: |
| BSCP_BCSSIDECAR_PULL_CFG_INTERVAL               | 60s                     | 自动同步最新配置版本间隔                         |
| BSCP_BCSSIDECAR_SYNC_CFGSETLIST_INTERVAL        | 10m                     | 自动同步配置集合列表间隔                         |
| BSCP_BCSSIDECAR_REPORT_INFO_INTERVAL            | 10m                     | 自动上报本地信息间隔                             |
| BSCP_BCSSIDECAR_ACCESS_INTERVAL                 | 3s                      | 接入链接会话服务等待间隔                         |
| BSCP_BCSSIDECAR_SESSION_TIMEOUT                 | 5s                      | 链接会话超时时间                                 |
| BSCP_BCSSIDECAR_SESSION_COEFFICIENT             | 2                       | 链接会话超时时间系数                             |
| BSCP_BCSSIDECAR_CFGSETLIST_SIZE                 | 1000                    | 拉取最大配置集合列表大小                         |
| BSCP_BCSSIDECAR_HANDLER_CH_SIZE                 | 10000                   | main处理协程管道大小                             |
| BSCP_BCSSIDECAR_HANDLER_CH_TIMEOUT              | 1s                      | main处理协程管道超时时间                         |
| BSCP_BCSSIDECAR_CFG_HANDLER_CH_SIZE             | 10000                   | 配置处理协程管道大小                             |
| BSCP_BCSSIDECAR_CFG_HANDLER_CH_TIMEOUT          | 1s                      | 配置处理协程管道超时时间                         |
| BSCP_BCSSIDECAR_CONNSERVER_HOSTNAME             | conn.bscp.bk.com    | 链接会话服务域名                                 |
| BSCP_BCSSIDECAR_CONNSERVER_PORT                 | 9516                    | 链接会话服务端口                                 |
| BSCP_BCSSIDECAR_CONNSERVER_DIALTIMEOUT          | 3s                      | 链接会话服务建立链接超时时间                     |
| BSCP_BCSSIDECAR_CONNSERVER_CALLTIMEOUT          | 3s                      | 链接会话服务请求超时时间                         |
| BSCP_BCSSIDECAR_APPINFO_LABELS                  | {"k1": "v1"}            | 当前Sidecar附带labels, json字符串KV格式          |
| BSCP_BCSSIDECAR_APPINFO_IP_ETH                  | eth1                    | 网卡名称，用于获取本地IP信息作为Sidecar身份标识  |
| BSCP_BCSSIDECAR_APPCFG_PATH                     | /my/app/config/dir/     | 应用配置路径                                     |
| BSCP_BCSSIDECAR_FILE_CACHE_PATH                 | ./cache/fcache/         | 生效信息文件缓存路径                             |
| BSCP_BCSSIDECAR_CONTENT_CACHE_PATH              | ./cache/ccache/         | 内容缓存路径                                     |
| BSCP_BCSSIDECAR_CONTENT_CACHE_EXPIRATION        | 168h                    | 内容缓存过期时间                                 |
| BSCP_BCSSIDECAR_CONTENT_EXPCACHE_PATH           | /tmp/                   | 过期内容缓存回收路径                             |
| BSCP_BCSSIDECAR_CONTENT_MCACHE_SIZE             | 1000                    | 内存内容缓存大小                                 |
| BSCP_BCSSIDECAR_CONTENT_MCACHE_EXPIRATION       | 10m                     | 内存内容缓存过期时间                             |
| BSCP_BCSSIDECAR_CONTENT_CACHE_PURGE_INTERVAL    | 30m                     | 内容缓存清理间隔                                 |
