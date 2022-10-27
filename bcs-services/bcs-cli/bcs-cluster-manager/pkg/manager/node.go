package manager

type GetNodeReq struct {
	InnerIP string
}

type UpdateNodeReq struct {
	InnerIPs    []string
	Status      string
	NodeGroupID string
	ClusterID   string
	Updater     string
}

type CordonNodeReq struct {
	InnerIPs  []string
	ClusterID string
	Updater   string
}

type UnCordonNodeReq struct {
	InnerIPs  []string
	ClusterID string
	Updater   string
}

type DrainNodeReq struct {
	InnerIPs  []string
	ClusterID string
	Updater   string
}

type GetNodeResp struct {
	Code    uint32    `json:"code"`
	Message string    `json:"message"`
	Result  bool      `json:"result"`
	Data    *NodeInfo `json:"data"`
}

type UpdateNodeResp struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
	Result  bool   `json:"result"`
}

type CordonNodeResp struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
	Result  bool   `json:"result"`
}

type UnCordonNodeResp struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
	Result  bool   `json:"result"`
}

type DrainNodeResp struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
	Result  bool   `json:"result"`
}

type NodeInfo struct {
	NodeID       string `json:"nodeID"`
	InnerIP      string `json:"innerIP"`
	InstanceType string `json:"InstanceType"`
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

type Node interface {
	//查询指定InnerIP的节点信息
	Get(GetNodeReq) GetNodeResp
	//更新node信息
	Update(UpdateNodeReq) UpdateNodeResp
	//节点设置不可调度状态
	Cordon(CordonNodeReq) CordonNodeResp
	//节点设置可调度状态
	UnCordon(UnCordonNodeReq) UnCordonNodeResp
	//节点pod迁移,将节点上的Pod驱逐
	Drain(DrainNodeReq) DrainNodeResp
}
