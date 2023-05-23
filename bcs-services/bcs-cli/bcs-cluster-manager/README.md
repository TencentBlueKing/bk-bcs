# bcs-cluster-manager 命令行工具

## 配置文件

配置文件默认放在 `/etc/bcs/bcs-cluster-manager.yaml` 文件：

```yaml
config:
  apiserver: "${BCS APISERVER地址}"
  bcs_token: "${Token信息}"
```

## 使用文档

### 创建云凭证 - CreateCloudAccount

```bash
kubectl-bcs-cluster-manager create cloudAccount --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager create cloudAccount -f [filename]
```
```json
{
	"cloudID": "tencentCloud",
	"accountName": "test001",
	"desc": "腾讯云测试账号",
	"account": {
		"secretID": "xxxxxxxxxx",
		"secretKey": "xxxxxxxxxxxx"
	},
	"enable": true,
	"creator": "bcs",
	"projectID": "b363e23b1b354928xxxxxxxxxxxxxxx"
}
```

### 创建云VPC管理信息 - CreateCloudVPC

```bash
kubectl-bcs-cluster-manager create cloudVPC --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager create cloudVPC -f [filename]
```
```json
{
	"cloudID": "tencentCloud",
	"networkType": "overlay",
	"region": "ap-guangzhou",
	"regionName": "广州",
	"vpcName": "vpc-xxxxxxx-1",
	"vpcID": "vpc-xxx",
	"available": "true",
	"extra": "",
	"creator": "bcs"
}
```

### 创建集群 - CreateCluster

```bash
kubectl-bcs-cluster-manager create cluster --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
-v, --virtual bool   "create virtual cluster"
```

示例:

```
kubectl-bcs-cluster-manager create cluster -f [filename]
kubectl-bcs-cluster-manager create cluster -v -f [filename]
```
```json
{
	"projectID": "b363e23b1b354928axxxxxxxxx",
	"businessID": "3",
	"engineType": "k8s",
	"isExclusive": true,
	"clusterType": "single",
	"creator": "bcs",
	"manageType": "INDEPENDENT_xxx",
	"clusterName": "test001",
	"environment": "prod",
	"provider": "bluekingCloud",
	"description": "创建测试集群",
	"clusterBasicSettings": {
		"version": "v1.20.xxx"
	},
	"networkType": "overlay",
	"region": "default",
	"vpcID": "",
	"networkSettings": {},
	"master": ["xxx.xxx.xxx.xxx", "xxx.xxx.xxx.xxx"]
}

virtual cluster json template
{
	"clusterID": "",
	"clusterName": "test-xxx",
	"provider": "tencentCloud",
	"region": "ap-xxx",
	"vpcID": "vpc-xxx",
	"projectID": "xxx",
	"businessID": "xxx",
	"environment": "debug",
	"engineType": "k8s",
	"isExclusive": true,
	"clusterType": "single",
	"hostClusterID": "BCS-K8S-xxx",
	"hostClusterNetwork": "devnet",
	"labels": {
		"xxx": "xxx"
	},
	"creator": "bcs",
	"onlyCreateInfo": false,
	"master": ["xxx"],
	"networkSettings": {
		"cidrStep": 1,
		"maxNodePodNum": 1,
		"maxServiceNum": 1
	},
	"clusterBasicSettings": {
		"version": "xxx"
	},
	"clusterAdvanceSettings": {
		"IPVS": false,
		"containerRuntime": "xxx",
		"runtimeVersion": "xxx",
		"extraArgs": {
			"xxx": "xxx"
		}
	},
	"nodeSettings": {
		"dockerGraphPath": "xxx",
		"mountTarget": "xxx",
		"unSchedulable": 1,
		"labels": {
			"xxx": "xxx"
		},
		"extraArgs": {
			"xxx": "xxx"
		}
	},
	"extraInfo": {
		"xxx": "xxx"
	},
	"description": "xxx",
	"ns": {
		"name": "xxx",
		"labels": {
			"xxx": "xxx"
		},
		"annotations": {
			"xxx": "xxx"
		}
	}
}
```

### 创建节点池 - CreateNodeGroup

```bash
kubectl-bcs-cluster-manager create nodeGroup --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager create nodeGroup -f [filename]
```
```json
{
	"name": "test001",
	"autoScaling": {
		"maxSize": 10,
		"minSize": 0,
		"scalingMode": "xxx",
		"multiZoneSubnetPolicy": "xxx",
		"retryPolicy": "IMMEDIATE_xxx",
		"subnetIDs": ["subnet-xxxxx"]
	},
	"enableAutoscale": true,
	"nodeTemplate": {
		"unSchedulable": 0,
		"labels": {},
		"taints": [],
		"dataDisks": [],
		"dockerGraphPath": "/var/lib/xxx",
		"runtime": {
			"containerRuntime": "docker",
			"runtimeVersion": "19.x"
		}
	},
	"launchTemplate": {
		"imageInfo": {
			"imageID": "img-xxx"
		},
		"CPU": 2,
		"Mem": 2,
		"instanceType": "S4.xxx",
		"systemDisk": {
			"diskType": "CLOUD_xxx",
			"diskSize": "50"
		},
		"internetAccess": {
			"internetChargeType": "TRAFFIC_POSTPAID_xxx",
			"internetMaxBandwidth": "0",
			"publicIPAssigned": false
		},
		"initLoginPassword": "123456",
		"securityGroupIDs": ["sg-xxx"],
		"dataDisks": [],
		"isSecurityService": true,
		"isMonitorService": true
	},
	"clusterID": "BCS-K8S-xxxxx",
	"region": "ap-shanghai",
	"creator": "bcs"
}
```

### 创建任务 - CreateTask

```bash
kubectl-bcs-cluster-manager create task --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager create task -f [filename]
```
```json
{
	"taskID": "feec6ed2-c3e3-481f-a58b-xxxxxx",
	"taskType": "blueking-xxx",
	"status": "FAILED",
	"message": "xxx",
	"start": "2022-11-11T18:23:32+08:00",
	"end": "2022-11-11T18:24:03+08:00",
	"executionTime": 31,
	"currentStep": "bksopsjob-xxx",
	"stepSequence": ["bksopsjob-xxx", "blueking-xxx"],
	"steps": {
		"bksopsjob-createTask": {
			"name": "bksopsjob-xxx",
			"system": "xxx",
			"link": "",
			"params": {
				"taskUrl": "http://apps.xxx.com"
			},
			"retry": 0,
			"start": "2022-11-11T18:23:32+08:00",
			"end": "2022-11-11T18:24:03+08:00",
			"executionTime": 31,
			"status": "FAILURE",
			"message": "running fialed",
			"lastUpdate": "2022-11-11T18:24:03+08:00",
			"taskMethod": "xxx",
			"taskName": "标准运维任务",
			"skipOnFailed": false
		},
		"blueking-UpdateAddNodeDBInfoTask": {
			"name": "blueking-xxx",
			"system": "api",
			"link": "",
			"params": null,
			"retry": 0,
			"start": "",
			"end": "",
			"executionTime": 0,
			"status": "NOTSTARTED",
			"message": "",
			"lastUpdate": "",
			"taskMethod": "blueking-xxx",
			"taskName": "更新任务状态",
			"skipOnFailed": false
		}
	},
	"clusterID": "BCS-K8S-xxx",
	"projectID": "b363e23b1b354928a0f3exxxxxx",
	"creator": "bcs",
	"lastUpdate": "2022-11-11T18:24:03+08:00",
	"updater": "bcs",
	"forceTerminate": false,
	"commonParams": {
		"jobType": "add-node",
		"nodeIPs": "xxx.xxx.xxx.xxx",
		"operator": "bcs",
		"taskName": "blueking-add nodes: BCS-K8S-xxx",
		"user": "bcs"
	},
	"taskName": "xxx",
	"nodeIPList": ["xxx.xxx.xxx.xxx"],
	"nodeGroupID": ""
}
```

### 删除云凭证 - DeleteCloudAccount

```bash
kubectl-bcs-cluster-manager delete cloudAccount --help
```

参数详情:

```yaml 
-c, --cloudID string   "cloud ID"
-a, --accountID string   "account ID"
```

示例:

```
kubectl-bcs-cluster-manager delete cloudAccount -c [cloudID] -a [accountID]
```

### 删除特定cloud vpc信息 - DeleteCloudVPC

```bash
kubectl-bcs-cluster-manager delete cloudVPC --help
```

参数详情:

```yaml 
-c, --cloudID string   "cloud ID"
--vpcID string   "VPC ID"
```

示例:

```
kubectl-bcs-cluster-manager delete cloudVPC -c [cloudID] -vpcID [VPCID]
```

### 删除集群 - DeleteCluster

```bash
kubectl-bcs-cluster-manager delete cluster --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-v --virtual bool   "whether the cluster is a virtual cluster"
```

示例:

```
kubectl-bcs-cluster-manager delete cluster -v -c [clusterID]
```

### 从集群中删除节点 - DeleteNodesFromCluster

```bash
kubectl-bcs-cluster-manager delete nodesFromCluster --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-n --node []string   "node ip"
```

示例:

```
kubectl-bcs-cluster-manager delete nodesFromCluster -c [clusterID] -n [xxx.xxx.xxx.xxx,xxx.xxx.xxx.xxx]
```

### 删除节点池 - DeleteNodeGroup

```bash
kubectl-bcs-cluster-manager delete nodeGroup --help
```

参数详情:

```yaml 
-n, --nodeGroupID string   "node group ID"
```

示例:

```
kubectl-bcs-cluster-manager delete nodeGroup -c [nodeGroupID]
```

### 删除任务 - DeleteTask

```bash
kubectl-bcs-cluster-manager delete task --help
```

参数详情:

```yaml 
-t, --taskID string   "task ID"
```

示例:

```
kubectl-bcs-cluster-manager delete task -t [taskID]
```

### 更新云凭证 - UpdateCloudAccount

```bash
kubectl-bcs-cluster-manager update cloudAccount --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager update cloudAccount -f [filename]
```

### 更新云vpc信息 - UpdateCloudVPC

```bash
kubectl-bcs-cluster-manager update cloudVPC --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager update cloudVPC -f [filename]
```

### 更新集群 - UpdateCluster

```bash
kubectl-bcs-cluster-manager update cluster --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager update cluster -f [filename]
```

### 更新node信息 - UpdateNode

```bash
kubectl-bcs-cluster-manager update node --help
```

参数详情:

```yaml 
-i, --innerIPs []string   "node inner ip"
-s, --status string   "更新节点状态(INITIALIZATION/RUNNING/DELETING/ADD-FAILURE/REMOVE-FAILURE)"
-n, --nodeGroupID string   "更新节点所属的node group ID"
-c, --clusterID string   "更新节点所属的cluster ID"
```

示例:

```
kubectl-bcs-cluster-manager update node -i [xxx.xxx.xxx.xxx] -s [RUNNING] -n [nodeGroupID] -c [clusterID]
```

### 更新节点池 - UpdateNodeGroup

```bash
kubectl-bcs-cluster-manager update nodeGroup --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager update nodeGroup -f [filename]
```

### 更新节点池DesiredNode信息 - UpdateGroupDesiredNode

```bash
kubectl-bcs-cluster-manager update groupDesiredNode --help
```

参数详情:

```yaml 
-n, --nodeGroupID string   "node group ID"
-d, --desiredNode uint32   "desired node"
```

示例:

```
kubectl-bcs-cluster-manager update groupDesiredNode -n [nodeGroupID] -d 1
```

### 更新节点池DesiredSize信息 - UpdateGroupDesiredSize

```bash
kubectl-bcs-cluster-manager update groupDesiredSize --help
```

参数详情:

```yaml 
-n, --nodeGroupID string   "node group ID"
-d, --desiredSize uint32   "desired node"
```

示例:

```
kubectl-bcs-cluster-manager update groupDesiredSize -n [nodeGroupID] -d 1
```

### 更新任务 - UpdateTask

```bash
kubectl-bcs-cluster-manager update task --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager update task -f [filename]
```

### 查询云凭证列表 - ListCloudAccount

```bash
kubectl-bcs-cluster-manager list cloudAccount --help
```

示例:

```
kubectl-bcs-cluster-manager list cloudAccount
```

### 查询云凭证列表,主要用于权限资源查询 - ListCloudAccountToPerm

```bash
kubectl-bcs-cluster-manager list cloudAccountToPerm --help
```

示例:

```
kubectl-bcs-cluster-manager list cloudAccountToPerm
```

### 根据cloudID获取所属cloud的地域信息 - ListCloudRegions

```bash
kubectl-bcs-cluster-manager list cloudRegions --help
```

参数详情:

```yaml 
-c, --cloudID string   "cloud ID"
```

示例:

```
kubectl-bcs-cluster-manager list cloudRegions -c [cloudID]
```

### 查询Cloud VPC列表 - ListCloudVPC

```bash
kubectl-bcs-cluster-manager list cloudVPC --help
```

参数详情:

```yaml 
-n, --networkType string   "cloud VPC network type (required) overlay/underlay"
```

示例:

```
kubectl-bcs-cluster-manager list cloudVPC -n [networkType]
```

### 查询公共集群及公共集群所属权限 - ListCommonCluster

```bash
kubectl-bcs-cluster-manager list commonCluster --help
```

示例:

```
kubectl-bcs-cluster-manager list commonCluster
```

### 查询集群下所有节点列表 - ListNodesInCluster

```bash
kubectl-bcs-cluster-manager list nodesInCluster --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-s, --offset uint32   "offset number of queries"
-l, --limit uint32   "limit number of queries"
```

示例:

```
kubectl-bcs-cluster-manager list nodesInCluster -c [clusterID] -s 0 -l 10
```

### 获取集群列表 - ListCluster

```bash
kubectl-bcs-cluster-manager list cluster --help
```

参数详情:

```yaml 
-s, --offset uint32   "offset number of queries"
-l, --limit uint32   "limit number of queries"
```

示例:

```
kubectl-bcs-cluster-manager list cluster -s 0 -l 10
```

### 查询节点池的节点列表 - ListNodesInGroup

```bash
kubectl-bcs-cluster-manager list nodesInGroup --help
```

参数详情:

```yaml 
-n, --nodeGroupID string   "node group ID"
```

示例:

```
kubectl-bcs-cluster-manager list nodesInGroup -n [nodeGroupID]
```

### 查询节点池列表 - ListNodeGroup

```bash
kubectl-bcs-cluster-manager list nodeGroup --help
```

示例:

```
kubectl-bcs-cluster-manager list nodeGroup
```

### 查询任务列表 - ListTask

```bash
kubectl-bcs-cluster-manager list task --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-p, --projectID string   "project ID"
```

示例:

```
kubectl-bcs-cluster-manager list task -c [clusterID] -p [projectID]
```

### 根据vpcID获取所属vpc的cidr信息 - GetVPCCidr

```bash
kubectl-bcs-cluster-manager get VPCCidr --help
```

参数详情:

```yaml 
-v, --vpcID string   "VPC ID"
```

示例:

```
kubectl-bcs-cluster-manager get VPCCidr -v [vpcID]
```

### 获取集群 - GetCluster

```bash
kubectl-bcs-cluster-manager get cluster --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
```

示例:

```
kubectl-bcs-cluster-manager get cluster -c [clusterID]
```

### 查询指定InnerIP的节点信息 - GetNode

```bash
kubectl-bcs-cluster-manager get node --help
```

参数详情:

```yaml 
-i, --innerIP string   "inner IP"
```

示例:

```
kubectl-bcs-cluster-manager get node -i [innerIP]
```

### 查询节点池信息 - GetNodeGroup

```bash
kubectl-bcs-cluster-manager get nodeGroup --help
```

参数详情:

```yaml 
-n, --nodeGroupID string   "node group ID"
```

示例:

```
kubectl-bcs-cluster-manager get nodeGroup -n [nodeGroupID]
```

### 查询任务 - GetTask

```bash
kubectl-bcs-cluster-manager get task --help
```

参数详情:

```yaml 
-n, --taskID string   "task ID"
```

示例:

```
kubectl-bcs-cluster-manager get task -n [taskID]
```

### 添加节点到集群 - AddNodesToCluster

```bash
kubectl-bcs-cluster-manager add nodesToCluster --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-n, --node []string   "node ip"
-p, --initPassword string   "init log password"
```

示例:

```
kubectl-bcs-cluster-manager add nodesToCluster -c [clusterID] -n [nodeIp] -p [initPassword]
```

### kubeConfig连接集群可用性检测 - CheckCloudKubeconfig

```bash
kubectl-bcs-cluster-manager check cloudKubeConfig --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager check cloudKubeConfig -c [filename]
```

### 检查节点是否存在bcs集群中 - CheckNodeInCluster

```bash
kubectl-bcs-cluster-manager check nodeInCluster --help
```

参数详情:

```yaml 
-i, --innerIPs []string   "node inner ip"
```

示例:

```
kubectl-bcs-cluster-manager check nodeInCluster -i [innerIPs]
```

### 从节点池移除节点并清理资源回收节点 - CleanNodesInGroup

```bash
kubectl-bcs-cluster-manager clean nodesInGroup --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-n, --nodeGroupID string   "node group ID"
-i, --nodes []string   "node inner ip"
```

示例:

```
kubectl-bcs-cluster-manager clean nodesInGroup -c [clusterID] -n [nodeGroupID] -i [nodes]
```

### 从节点池移除节点并清理资源回收节点V2 - CleanNodesInGroupV2

```bash
kubectl-bcs-cluster-manager clean nodesInGroupV2 --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-n, --nodeGroupID string   "node group ID"
-i, --nodes []string   "node inner ip"
```

示例:

```
kubectl-bcs-cluster-manager clean nodesInGroupV2 -c [clusterID] -n [nodeGroupID] -i [nodes]
```

### 节点设置不可调度状态 - CordonNode

```bash
kubectl-bcs-cluster-manager cordon node --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-i, --innerIPs []string   "node inner ip"
```

示例:

```
kubectl-bcs-cluster-manager cordon node -c [clusterID] -i [nodes]
```

### 关闭节点池自动伸缩功能 - DisableNodeGroupAutoScale

```bash
kubectl-bcs-cluster-manager disable nodeGroupAutoScale --help
```

参数详情:

```yaml 
-n, --nodeGroupID string   "node group ID"
```

示例:

```
kubectl-bcs-cluster-manager disable nodeGroupAutoScale -n [nodeGroupID]
```

### 节点pod迁移,将节点上的Pod驱逐 - DrainNode

```bash
kubectl-bcs-cluster-manager drain node --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-i, --innerIPs []string   "node inner ip"
```

示例:

```
kubectl-bcs-cluster-manager drain node -c [clusterID] -i [innerIPs]
```

### 开启节点池自动伸缩功能 - EnableNodeGroupAutoScale

```bash
kubectl-bcs-cluster-manager enable nodeGroupAutoScale --help
```

参数详情:

```yaml 
-n, --nodeGroupID string   "node group ID"
```

示例:

```
kubectl-bcs-cluster-manager enable nodeGroupAutoScale -n [nodeGroupID]
```

### 导入用户集群(支持多云集群导入功能: 集群ID/kubeConfig) - ImportCluster

```bash
kubectl-bcs-cluster-manager import cluster --help
```

参数详情:

```yaml 
-f, --filename string   "File address Support json file"
```

示例:

```
kubectl-bcs-cluster-manager import cluster -f [filename]
```
```json
{
	"clusterID": "xxx",
	"projectID": "xxx",
	"businessID": "100001",
	"engineType": "k8s",
	"isExclusive": false,
	"clusterType": "single",
	"clusterName": "ceshi",
	"environment": "stag",
	"provider": "tencentCloud"
}
```

### 移动节点到节点池 - MoveNodesToGroup

```bash
kubectl-bcs-cluster-manager move nodesToGroup --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-n, --nodeGroupID string   "node group ID"
-i, --nodes []string   "node inner ip"
```

示例:

```
kubectl-bcs-cluster-manager move nodesToGroup -c [clusterID] -n [nodeGroupID] -i [nodes]
```

### 从节点池移除节点 - RemoveNodesFromGroup

```bash
kubectl-bcs-cluster-manager remove nodesFromGroup --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-n, --nodeGroupID string   "node group ID"
-i, --nodes []string   "node inner ip"
```

示例:

```
kubectl-bcs-cluster-manager remove nodesFromGroup -c [clusterID] -n [nodeGroupID] -i [nodes]
```

### 重试创建任务 - RetryTask

```bash
kubectl-bcs-cluster-manager retry task --help
```

参数详情:

```yaml 
-t, --taskID string   "task ID"
```

示例:

```
kubectl-bcs-cluster-manager retry task -t [taskID]
```

### 重试创建集群 - RetryCreateClusterTask

```bash
kubectl-bcs-cluster-manager retry createClusterTask --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
```

示例:

```
kubectl-bcs-cluster-manager retry createClusterTask -c [clusterID]
```

### 节点设置可调度状态 - UncordonNode

```bash
kubectl-bcs-cluster-manager uncordon node --help
```

参数详情:

```yaml 
-c, --clusterID string   "cluster ID"
-i, --innerIPs []string   "node inner ip"
```

示例:

```
kubectl-bcs-cluster-manager uncordon node -c [clusterID] -i [innerIPs]
```

### 获取项目列表 - ListProjects

```bash
kubectl-bcs-cluster-manager list project --help
```

参数详情:

```yaml 
--kind           "项目中集群类型, 允许k8s/mesos"  
--names          "项目中文名称, 长度不能超过64字符, 多个以半角逗号分隔"
--project_code   "项目编码(英文缩写), 全局唯一, 长度不能超过64字符"
--project_ids    "项目ID, 多个以半角逗号分隔"
--search_name    "项目中文名称, 通过此字段模糊查询项目信息"
```

## 如何编译

执行下述命令编译 Client 工具

```
make bin
```