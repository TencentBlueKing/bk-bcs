package clustermanager

// NodeGroup 节点池定义
type NodeGroup struct {
	NodeGroupID          string               `protobuf:"bytes,1,opt,name=nodeGroupID,proto3" json:"nodeGroupID,omitempty"`
	Name                 string               `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	ClusterID            string               `protobuf:"bytes,3,opt,name=clusterID,proto3" json:"clusterID,omitempty"`
	Region               string               `protobuf:"bytes,4,opt,name=region,proto3" json:"region,omitempty"`
	EnableAutoscale      bool                 `protobuf:"varint,5,opt,name=enableAutoscale,proto3" json:"enableAutoscale,omitempty"`
	AutoScaling          *AutoScalingGroup    `protobuf:"bytes,6,opt,name=autoScaling,proto3" json:"autoScaling,omitempty"`
	LaunchTemplate       *LaunchConfiguration `protobuf:"bytes,7,opt,name=launchTemplate,proto3" json:"launchTemplate,omitempty"`
	Labels               map[string]string    `protobuf:"bytes,8,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Taints               map[string]string    `protobuf:"bytes,9,rep,name=taints,proto3" json:"taints,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	NodeOS               string               `protobuf:"bytes,10,opt,name=nodeOS,proto3" json:"nodeOS,omitempty"`
	Creator              string               `protobuf:"bytes,11,opt,name=creator,proto3" json:"creator,omitempty"`
	Updater              string               `protobuf:"bytes,12,opt,name=updater,proto3" json:"updater,omitempty"`
	CreateTime           string               `protobuf:"bytes,13,opt,name=createTime,proto3" json:"createTime,omitempty"`
	UpdateTime           string               `protobuf:"bytes,14,opt,name=updateTime,proto3" json:"updateTime,omitempty"`
	ProjectID            string               `protobuf:"bytes,15,opt,name=projectID,proto3" json:"projectID,omitempty"`
	Provider             string               `protobuf:"bytes,16,opt,name=provider,proto3" json:"provider,omitempty"`
	Status               string               `protobuf:"bytes,17,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-" bson:"-"`
	XXX_unrecognized     []byte               `json:"-" bson:"-"`
	XXX_sizecache        int32                `json:"-" bson:"-"`
}

// AutoScalingGroup 自动伸缩定义
type AutoScalingGroup struct {
	AutoScalingID         string       `protobuf:"bytes,1,opt,name=autoScalingID,proto3" json:"autoScalingID,omitempty"`
	AutoScalingName       string       `protobuf:"bytes,2,opt,name=autoScalingName,proto3" json:"autoScalingName,omitempty"`
	MinSize               uint32       `protobuf:"varint,3,opt,name=minSize,proto3" json:"minSize,omitempty"`
	MaxSize               uint32       `protobuf:"varint,4,opt,name=maxSize,proto3" json:"maxSize,omitempty"`
	DesiredSize           uint32       `protobuf:"varint,5,opt,name=desiredSize,proto3" json:"desiredSize,omitempty"`
	VpcID                 string       `protobuf:"bytes,6,opt,name=vpcID,proto3" json:"vpcID,omitempty"`
	DefaultCooldown       uint32       `protobuf:"varint,7,opt,name=defaultCooldown,proto3" json:"defaultCooldown,omitempty"`
	SubnetIDs             []string     `protobuf:"bytes,8,rep,name=subnetIDs,proto3" json:"subnetIDs,omitempty"`
	Zones                 []string     `protobuf:"bytes,9,rep,name=zones,proto3" json:"zones,omitempty"`
	RetryPolicy           string       `protobuf:"bytes,10,opt,name=retryPolicy,proto3" json:"retryPolicy,omitempty"`
	MultiZoneSubnetPolicy string       `protobuf:"bytes,11,opt,name=multiZoneSubnetPolicy,proto3" json:"multiZoneSubnetPolicy,omitempty"`
	ReplaceUnhealthy      bool         `protobuf:"varint,12,opt,name=replaceUnhealthy,proto3" json:"replaceUnhealthy,omitempty"`
	ScalingMode           string       `protobuf:"bytes,13,opt,name=scalingMode,proto3" json:"scalingMode,omitempty"`
	TimeRanges            []*TimeRange `protobuf:"bytes,14,rep,name=timeRanges,proto3" json:"timeRanges,omitempty"`
	XXX_NoUnkeyedLiteral  struct{}     `json:"-" bson:"-"`
	XXX_unrecognized      []byte       `json:"-" bson:"-"`
	XXX_sizecache         int32        `json:"-" bson:"-"`
}

// LaunchConfiguration 节点模板定义
type LaunchConfiguration struct {
	LaunchConfigurationID string              `protobuf:"bytes,1,opt,name=launchConfigurationID,proto3" json:"launchConfigurationID,omitempty"`
	LaunchConfigureName   string              `protobuf:"bytes,2,opt,name=launchConfigureName,proto3" json:"launchConfigureName,omitempty"`
	ProjectID             string              `protobuf:"bytes,3,opt,name=projectID,proto3" json:"projectID,omitempty"`
	CPU                   uint32              `protobuf:"varint,4,opt,name=CPU,proto3" json:"CPU,omitempty"`
	Mem                   uint32              `protobuf:"varint,5,opt,name=Mem,proto3" json:"Mem,omitempty"`
	GPU                   uint32              `protobuf:"varint,6,opt,name=GPU,proto3" json:"GPU,omitempty"`
	InstanceType          string              `protobuf:"bytes,7,opt,name=instanceType,proto3" json:"instanceType,omitempty"`
	InstanceChargeType    string              `protobuf:"bytes,8,opt,name=instanceChargeType,proto3" json:"instanceChargeType,omitempty"`
	SystemDisk            *DataDisk           `protobuf:"bytes,9,opt,name=systemDisk,proto3" json:"systemDisk,omitempty"`
	DataDisks             []*DataDisk         `protobuf:"bytes,10,rep,name=dataDisks,proto3" json:"dataDisks,omitempty"`
	InternetAccess        *InternetAccessible `protobuf:"bytes,11,opt,name=internetAccess,proto3" json:"internetAccess,omitempty"`
	InitLoginPassword     string              `protobuf:"bytes,12,opt,name=initLoginPassword,proto3" json:"initLoginPassword,omitempty"`
	SecurityGroupIDs      []string            `protobuf:"bytes,13,rep,name=securityGroupIDs,proto3" json:"securityGroupIDs,omitempty"`
	ImageInfo             *ImageInfo          `protobuf:"bytes,14,opt,name=imageInfo,proto3" json:"imageInfo,omitempty"`
	IsSecurityService     bool                `protobuf:"varint,15,opt,name=isSecurityService,proto3" json:"isSecurityService,omitempty"`
	IsMonitorService      bool                `protobuf:"varint,16,opt,name=isMonitorService,proto3" json:"isMonitorService,omitempty"`
	XXX_NoUnkeyedLiteral  struct{}            `json:"-" bson:"-"`
	XXX_unrecognized      []byte              `json:"-" bson:"-"`
	XXX_sizecache         int32               `json:"-" bson:"-"`
}

// DataDisk 数据盘定义
type DataDisk struct {
	DiskType             string   `protobuf:"bytes,1,opt,name=diskType,proto3" json:"diskType,omitempty"`
	DiskSize             string   `protobuf:"bytes,2,opt,name=diskSize,proto3" json:"diskSize,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// InternetAccessible 公网带宽设置
type InternetAccessible struct {
	InternetChargeType   string   `protobuf:"bytes,1,opt,name=internetChargeType,proto3" json:"internetChargeType,omitempty"`
	InternetMaxBandwidth string   `protobuf:"bytes,2,opt,name=internetMaxBandwidth,proto3" json:"internetMaxBandwidth,omitempty"`
	PublicIPAssigned     bool     `protobuf:"varint,3,opt,name=publicIPAssigned,proto3" json:"publicIPAssigned,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// ImageInfo 镜像定义
type ImageInfo struct {
	ImageID              string   `protobuf:"bytes,1,opt,name=imageID,proto3" json:"imageID,omitempty"`
	ImageName            string   `protobuf:"bytes,2,opt,name=imageName,proto3" json:"imageName,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// Node 节点定义
type Node struct {
	NodeID               string   `protobuf:"bytes,1,opt,name=nodeID,proto3" json:"nodeID,omitempty"`
	InnerIP              string   `protobuf:"bytes,2,opt,name=innerIP,proto3" json:"innerIP,omitempty"`
	InstanceType         string   `protobuf:"bytes,3,opt,name=instanceType,proto3" json:"instanceType,omitempty"`
	CPU                  uint32   `protobuf:"varint,4,opt,name=CPU,proto3" json:"CPU,omitempty"`
	Mem                  uint32   `protobuf:"varint,5,opt,name=mem,proto3" json:"mem,omitempty"`
	GPU                  uint32   `protobuf:"varint,6,opt,name=GPU,proto3" json:"GPU,omitempty"`
	Status               string   `protobuf:"bytes,7,opt,name=status,proto3" json:"status,omitempty"`
	ZoneID               string   `protobuf:"bytes,8,opt,name=zoneID,proto3" json:"zoneID,omitempty"`
	NodeGroupID          string   `protobuf:"bytes,9,opt,name=nodeGroupID,proto3" json:"nodeGroupID,omitempty"`
	ClusterID            string   `protobuf:"bytes,10,opt,name=clusterID,proto3" json:"clusterID,omitempty"`
	VPC                  string   `protobuf:"bytes,11,opt,name=VPC,proto3" json:"VPC,omitempty"`
	Region               string   `protobuf:"bytes,12,opt,name=region,proto3" json:"region,omitempty"`
	Passwd               string   `protobuf:"bytes,13,opt,name=passwd,proto3" json:"passwd,omitempty"`
	Zone                 uint32   `protobuf:"varint,14,opt,name=zone,proto3" json:"zone,omitempty"`
	DeviceID             string   `protobuf:"bytes,15,opt,name=deviceID,proto3" json:"deviceID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// Task 任务定义
type Task struct {
	TaskID               string            `protobuf:"bytes,1,opt,name=taskID,proto3" json:"taskID,omitempty"`
	TaskType             string            `protobuf:"bytes,2,opt,name=taskType,proto3" json:"taskType,omitempty"`
	Status               string            `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Message              string            `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
	Start                string            `protobuf:"bytes,5,opt,name=start,proto3" json:"start,omitempty"`
	End                  string            `protobuf:"bytes,6,opt,name=end,proto3" json:"end,omitempty"`
	ExecutionTime        uint32            `protobuf:"varint,7,opt,name=executionTime,proto3" json:"executionTime,omitempty"`
	CurrentStep          string            `protobuf:"bytes,8,opt,name=currentStep,proto3" json:"currentStep,omitempty"`
	StepSequence         []string          `protobuf:"bytes,9,rep,name=stepSequence,proto3" json:"stepSequence,omitempty"`
	Steps                map[string]*Step  `protobuf:"bytes,10,rep,name=steps,proto3" json:"steps,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ClusterID            string            `protobuf:"bytes,11,opt,name=clusterID,proto3" json:"clusterID,omitempty"`
	ProjectID            string            `protobuf:"bytes,12,opt,name=projectID,proto3" json:"projectID,omitempty"`
	Creator              string            `protobuf:"bytes,13,opt,name=creator,proto3" json:"creator,omitempty"`
	LastUpdate           string            `protobuf:"bytes,14,opt,name=lastUpdate,proto3" json:"lastUpdate,omitempty"`
	Updater              string            `protobuf:"bytes,15,opt,name=updater,proto3" json:"updater,omitempty"`
	ForceTerminate       bool              `protobuf:"varint,16,opt,name=forceTerminate,proto3" json:"forceTerminate,omitempty"`
	CommonParams         map[string]string `protobuf:"bytes,17,rep,name=commonParams,proto3" json:"commonParams,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-" bson:"-"`
	XXX_unrecognized     []byte            `json:"-" bson:"-"`
	XXX_sizecache        int32             `json:"-" bson:"-"`
}

// Step 任务步骤定义
type Step struct {
	Name                 string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	System               string            `protobuf:"bytes,2,opt,name=system,proto3" json:"system,omitempty"`
	Link                 string            `protobuf:"bytes,3,opt,name=link,proto3" json:"link,omitempty"`
	Params               map[string]string `protobuf:"bytes,4,rep,name=params,proto3" json:"params,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Retry                uint32            `protobuf:"varint,5,opt,name=retry,proto3" json:"retry,omitempty"`
	Start                string            `protobuf:"bytes,6,opt,name=start,proto3" json:"start,omitempty"`
	End                  string            `protobuf:"bytes,7,opt,name=end,proto3" json:"end,omitempty"`
	ExecutionTime        uint32            `protobuf:"varint,8,opt,name=executionTime,proto3" json:"executionTime,omitempty"`
	Status               string            `protobuf:"bytes,9,opt,name=status,proto3" json:"status,omitempty"`
	Message              string            `protobuf:"bytes,10,opt,name=message,proto3" json:"message,omitempty"`
	LastUpdate           string            `protobuf:"bytes,11,opt,name=lastUpdate,proto3" json:"lastUpdate,omitempty"`
	TaskMethod           string            `protobuf:"bytes,12,opt,name=taskMethod,proto3" json:"taskMethod,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-" bson:"-"`
	XXX_unrecognized     []byte            `json:"-" bson:"-"`
	XXX_sizecache        int32             `json:"-" bson:"-"`
}

// GetNodeGroupResponse 获取 NodeGroup 响应
type GetNodeGroupResponse struct {
	Code                 uint32     `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string     `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Result               bool       `protobuf:"varint,3,opt,name=result,proto3" json:"result,omitempty"`
	Data                 *NodeGroup `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-" bson:"-"`
	XXX_unrecognized     []byte     `json:"-" bson:"-"`
	XXX_sizecache        int32      `json:"-" bson:"-"`
}

// ListNodesInGroupResponse 获取节点池节点响应
type ListNodesInGroupResponse struct {
	Code                 uint32   `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Result               bool     `protobuf:"varint,3,opt,name=result,proto3" json:"result,omitempty"`
	Data                 []*Node  `protobuf:"bytes,4,rep,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// GetNodeResponse 获取节点详情响应
type GetNodeResponse struct {
	Code                 uint32   `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Result               bool     `protobuf:"varint,3,opt,name=result,proto3" json:"result,omitempty"`
	Data                 []*Node  `protobuf:"bytes,4,rep,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// UpdateGroupDesiredNodeRequest 节点池扩容请求
type UpdateGroupDesiredNodeRequest struct {
	NodeGroupID          string   `protobuf:"bytes,1,opt,name=nodeGroupID,proto3" json:"nodeGroupID,omitempty"`
	DesiredNode          uint32   `protobuf:"varint,2,opt,name=desiredNode,proto3" json:"desiredNode,omitempty"`
	Operator             string   `protobuf:"bytes,3,opt,name=operator,proto3" json:"operator,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// UpdateGroupDesiredNodeResponse 节点池扩容响应
type UpdateGroupDesiredNodeResponse struct {
	Code                 uint32   `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Result               bool     `protobuf:"varint,3,opt,name=result,proto3" json:"result,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// CleanNodesInGroupRequest 缩容节点请求
type CleanNodesInGroupRequest struct {
	ClusterID            string   `protobuf:"bytes,1,opt,name=clusterID,proto3" json:"clusterID,omitempty"`
	Nodes                []string `protobuf:"bytes,2,rep,name=nodes,proto3" json:"nodes,omitempty"`
	NodeGroupID          string   `protobuf:"bytes,3,opt,name=nodeGroupID,proto3" json:"nodeGroupID,omitempty"`
	Operator             string   `protobuf:"bytes,4,opt,name=operator,proto3" json:"operator,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// CleanNodesInGroupResponse 缩容节点响应
type CleanNodesInGroupResponse struct {
	Code                 uint32   `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Result               bool     `protobuf:"varint,3,opt,name=result,proto3" json:"result,omitempty"`
	Data                 *Task    `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// UpdateGroupDesiredSizeRequest 更新节点期望数请求
type UpdateGroupDesiredSizeRequest struct {
	DesiredSize          uint32   `protobuf:"varint,1,opt,name=desiredSize,proto3" json:"desiredSize,omitempty"`
	Operator             string   `protobuf:"bytes,2,opt,name=operator,proto3" json:"operator,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// UpdateGroupDesiredSizeResponse 更新节点期望数响应
type UpdateGroupDesiredSizeResponse struct {
	Code                 uint32   `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Result               bool     `protobuf:"varint,3,opt,name=result,proto3" json:"result,omitempty"`
	Data                 []*Node  `protobuf:"bytes,4,rep,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

// TimeRange 定时规则定义
type TimeRange struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Schedule             string   `protobuf:"bytes,2,opt,name=schedule,proto3" json:"schedule,omitempty"`
	Zone                 string   `protobuf:"bytes,3,opt,name=zone,proto3" json:"zone,omitempty"`
	DesiredNum           uint32   `protobuf:"varint,4,opt,name=desiredNum,proto3" json:"desiredNum,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}
