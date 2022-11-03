package manager

type GetNodeReq struct {
	InnerIP string `json:"innerIP"`
}

type UpdateNodeReq struct {
	InnerIPs    []string `json:"innerIPs"`
	Status      string   `json:"status"`
	NodeGroupID string   `json:"nodeGroupID"`
	ClusterID   string   `json:"clusterID"`
}

type CheckNodeInClusterReq struct {
	InnerIPs []string `json:"innerIPs"`
}

type CordonNodeReq struct {
	InnerIPs  []string `json:"innerIPs"`
	ClusterID string
}

type UnCordonNodeReq struct {
	InnerIPs  []string `json:"innerIPs"`
	ClusterID string   `json:"clusterID"`
}

type DrainNodeReq struct {
	InnerIPs  []string `json:"innerIPs"`
	ClusterID string   `json:"clusterID"`
}

type GetNodeResp struct {
	Data []*Node `json:"data"`
}

type CheckNodeInClusterResp struct {
	Data map[string]NodeResult `json:"data"`
}

type CordonNodeResp struct {
	Data []string `json:"data"`
}

type UnCordonNodeResp struct {
	Data []string `json:"data"`
}

type DrainNodeResp struct {
	Data []string `json:"data"`
}

type Node struct {
	NodeID       string `json:"nodeID"`
	InnerIP      string `json:"innerIP"`
	InstanceType string `json:"instanceType"`
	CPU          uint32 `json:"cpu"`
	Mem          uint32 `json:"mem"`
	GPU          uint32 `json:"gpu"`
	Status       string `json:"status"`
	ZoneID       string `json:"zoneID"`
	NodeGroupID  string `json:"nodeGroupID"`
	ClusterID    string `json:"clusterID"`
	VPC          string `json:"vpc"`
	Region       string `json:"region"`
	Passwd       string `json:"passwd"`
	Zone         uint32 `json:"zone"`
	DeviceID     string `json:"deviceID"`
}

type NodeResult struct {
	IsExist     bool   `json:"isExist"`
	ClusterID   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
}

type NodeOperationStatus struct {
	Fail    []NodeOperationStatusInfo `json:"fail"`
	Success []NodeOperationStatusInfo `json:"success"`
}

type NodeOperationStatusInfo struct {
	NodeName string `json:"nodeName"`
	Message  string `json:"message"`
}

type NodeMgr interface {
	// Get 查询指定InnerIP的节点信息
	Get(GetNodeReq) (GetNodeResp, error)
	// Update 更新node信息
	Update(UpdateNodeReq) error
	// CheckNodeInCluster 检查节点是否存在bcs集群
	CheckNodeInCluster(CheckNodeInClusterReq) (CheckNodeInClusterResp, error)
	// Cordon 节点设置不可调度状态
	Cordon(CordonNodeReq) (CordonNodeResp, error)
	// UnCordon 节点设置可调度状态
	UnCordon(UnCordonNodeReq) (UnCordonNodeResp, error)
	// Drain 节点pod迁移,将节点上的Pod驱逐
	Drain(DrainNodeReq) (DrainNodeResp, error)
}
