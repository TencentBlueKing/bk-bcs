### 描述

获取 集群管理 节点池列表

### 查询参数
| 参数名称     | 参数类型     | 必选   | 描述             |
| ------------ | ------------ | ------ | ---------------- |
| clusterID         | string       | 是     | 集群ID     |


### 调用示例
```sh
curl -X GET -H 'x-bkapi-authorization: {"bk_ticket": "xxx", "bk_app_code": "xxx", "bk_app_secret": "***"}' --insecure https://bcs-api-gateway.apigw.com/uat/clustermanager/v1/nodegroup
```

### 依赖结构
- 节点池
```json
type NodeGroup struct {
	NodeGroupID          string               `json:"nodeGroupID,omitempty"`
	Name                 string               `json:"name,omitempty"`
	ClusterID            string               `json:"clusterID,omitempty"`
	Region               string               `json:"region,omitempty"`
	EnableAutoscale      bool                 `json:"enableAutoscale,omitempty"`
	AutoScaling          *AutoScalingGroup    `json:"autoScaling,omitempty"`
	LaunchTemplate       *LaunchConfiguration `json:"launchTemplate,omitempty"`
	Labels               map[string]string    `json:"labels,omitempty"`
	Taints               map[string]string    `json:"taints,omitempty"`
	NodeOS               string               `json:"nodeOS,omitempty"`
	Creator              string               `json:"creator,omitempty"`
	Updater              string               `json:"updater,omitempty"`
	CreateTime           string               `json:"createTime,omitempty"`
	UpdateTime           string               `json:"updateTime,omitempty"`
	ProjectID            string               `json:"projectID,omitempty"`
	Provider             string               `json:"provider,omitempty"`
	Status               string               `json:"status,omitempty"`
	ConsumerID           string               `json:"consumerID,omitempty"`
	NodeTemplate         *NodeTemplate        `json:"nodeTemplate,omitempty"`
	CloudNodeGroupID     string               `json:"cloudNodeGroupID,omitempty"`
	Tags                 map[string]string    `json:"tags,omitempty"`
	NodeGroupType        string               `json:"nodeGroupType,omitempty"`
}

type AutoScalingGroup struct {
	AutoScalingID         string       `json:"autoScalingID,omitempty"`
	AutoScalingName       string       `json:"autoScalingName,omitempty"`
	MinSize               uint32       `json:"minSize,omitempty"`
	MaxSize               uint32       `json:"maxSize,omitempty"`
	DesiredSize           uint32       `json:"desiredSize,omitempty"`
	VpcID                 string       `json:"vpcID,omitempty"`
	DefaultCooldown       uint32       `json:"defaultCooldown,omitempty"`
	SubnetIDs             []string     `json:"subnetIDs,omitempty"`
	Zones                 []string     `json:"zones,omitempty"`
	RetryPolicy           string       `json:"retryPolicy,omitempty"`
	MultiZoneSubnetPolicy string       `json:"multiZoneSubnetPolicy,omitempty"`
	ReplaceUnhealthy      bool         `json:"replaceUnhealthy,omitempty"`
	ScalingMode           string       `json:"scalingMode,omitempty"`
}

//LaunchConfigure template for scaling node
type LaunchConfiguration struct {
	LaunchConfigurationID string              `json:"launchConfigurationID,omitempty"`
	LaunchConfigureName   string              `json:"launchConfigureName,omitempty"`
	ProjectID             string              `json:"projectID,omitempty"`
	CPU                   uint32              `json:"CPU,omitempty"`
	Mem                   uint32              `json:"Mem,omitempty"`
	GPU                   uint32              `json:"GPU,omitempty"`
	InstanceType          string              `json:"instanceType,omitempty"`
	InstanceChargeType    string              `json:"instanceChargeType,omitempty"`
	SystemDisk            *DataDisk           `json:"systemDisk,omitempty"`
	DataDisks             []*DataDisk         `json:"dataDisks,omitempty"`
	InternetAccess        *InternetAccessible `json:"internetAccess,omitempty"`
	InitLoginPassword     string              `json:"initLoginPassword,omitempty"`
	SecurityGroupIDs      []string            `json:"securityGroupIDs,omitempty"`
	ImageInfo             *ImageInfo          `json:"imageInfo,omitempty"`
	IsSecurityService     bool                `json:"isSecurityService,omitempty"`
	IsMonitorService      bool                `json:"isMonitorService,omitempty"`
	UserData              string              `json:"userData,omitempty"`
}

// DataDisk 数据盘定义
type DataDisk struct {
	DiskType             string   `json:"diskType,omitempty"`
	DiskSize             string   `json:"diskSize,omitempty"`
}

// InternetAccessible 公网带宽设置
type InternetAccessible struct {
	InternetChargeType   string   `json:"internetChargeType,omitempty"`
	InternetMaxBandwidth string   `json:"internetMaxBandwidth,omitempty"`
	PublicIPAssigned     bool     `json:"publicIPAssigned,omitempty"`
}

// ImageInfo 创建cvm实例的镜像信息
type ImageInfo struct {
	ImageID              string   `json:"imageID,omitempty"`
	ImageName            string   `json:"imageName,omitempty"`
}

type NodeTemplate struct {
	NodeTemplateID       string            `json:"nodeTemplateID,omitempty"`
	Name                 string            `json:"name,omitempty"`
	ProjectID            string            `json:"projectID,omitempty"`
	Labels               map[string]string `json:"labels,omitempty"`
	Taints               []*Taint          `json:"taints,omitempty"`
	Module               *ModuleInfo       `json:"module,omitempty"`
}

// ModuleInfo 业务模块信息,主要涉及到节点模块转移
type ModuleInfo struct {
	ScaleOutModuleID     string   `json:"scaleOutModuleID,omitempty"`
	ScaleInModuleID      string   `json:"scaleInModuleID,omitempty"`
	ScaleOutBizID        string   `json:"scaleOutBizID,omitempty"`
	ScaleInBizID         string   `json:"scaleInBizID,omitempty"`
}


```

### 响应示例
```json
{
    "code": uint,
    "message": string,
    "data": []NodeGroup,
    "requestID": string
}
```