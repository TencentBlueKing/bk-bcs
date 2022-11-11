package manager

type CreateClusterReq struct {
	ProjectID            string               `json:"projectID"`
	BusinessID           string               `json:"businessID"`
	EngineType           string               `json:"engineType"`
	IsExclusive          bool                 `json:"isExclusive"`
	ClusterType          string               `json:"clusterType"`
	Creator              string               `json:"creator"`
	ManageType           string               `json:"manageType"`
	ClusterName          string               `json:"clusterName"`
	Environment          string               `json:"environment"`
	Provider             string               `json:"provider"`
	Description          string               `json:"description"`
	ClusterBasicSettings ClusterBasicSettings `json:"clusterBasicSettings"`
	NetworkType          string               `json:"networkType"`
	Region               string               `json:"region"`
	VpcID                string               `json:"vpcID"`
	NetworkSettings      NetworkSettings      `json:"networkSettings"`
	Master               []string             `json:"master"`
}

type CreateClusterResp struct {
	ClusterID string `json:"clusterID"`
	TaskID    string `json:"taskID"`
}

type DeleteClusterResp struct {
	ClusterID string `json:"clusterID"`
	TaskID    string `json:"taskID"`
}

type DeleteClusterReq struct {
	ClusterID string `json:"clusterID"`
}

type UpdateClusterReq struct {
	ClusterID            string               `json:"clusterID"`
	ProjectID            string               `json:"projectID"`
	BusinessID           string               `json:"businessID"`
	EngineType           string               `json:"engineType"`
	IsExclusive          bool                 `json:"isExclusive"`
	ClusterType          string               `json:"clusterType"`
	Updater              string               `json:"updater"`
	ManageType           string               `json:"manageType"`
	ClusterName          string               `json:"clusterName"`
	Environment          string               `json:"environment"`
	Provider             string               `json:"provider"`
	Description          string               `json:"description"`
	ClusterBasicSettings ClusterBasicSettings `json:"clusterBasicSettings"`
	NetworkType          string               `json:"networkType"`
	Region               string               `json:"region"`
	VpcID                string               `json:"vpcID"`
	NetworkSettings      NetworkSettings      `json:"networkSettings"`
	Master               []string             `json:"master"`
}

type ListClusterReq struct {
	Offset uint32 `json:"offset"`
	Limit  uint32 `json:"limit"`
}

type GetClusterReq struct {
	ClusterID string `json:"clusterID"`
}

type RetryCreateClusterReq struct {
	ClusterID string `json:"clusterID"`
}

type AddNodesClusterReq struct {
	ClusterID string   `json:"clusterID"`
	Nodes     []string `json:"nodes"`
}

type DeleteNodesClusterReq struct {
	ClusterID string   `json:"clusterID"`
	Nodes     []string `json:"nodes"`
}

type CheckCloudKubeconfigReq struct {
	Kubeconfig string `json:"kubeconfig"`
}

type ImportClusterReq struct {
	ClusterID   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
	Provider    string `json:"provider"`
	ProjectID   string `json:"projectID"`
	BusinessID  string `json:"businessID"`
	Environment string `json:"environment"`
	EngineType  string `json:"engineType"`
	IsExclusive bool   `json:"isExclusive"`
	ClusterType string `json:"clusterType"`
}

type ListClusterNodesReq struct {
	Offset uint32 `json:"offset"`
	Limit  uint32 `json:"limit"`
}

type ListCommonClusterReq struct {
}

type GetClusterResp struct {
	Data Cluster `json:"data"`
}

type ListClusterResp struct {
	Data []*Cluster `json:"data"`
}

type RetryCreateClusterResp struct {
	ClusterID string `json:"clusterID"`
	TaskID    string `json:"taskID"`
}

type AddNodesClusterResp struct {
	TaskID string `json:"taskID"`
}

type DeleteNodesClusterResp struct {
	TaskID string `json:"taskID"`
}

type ListClusterNodesResp struct {
	Data []*ClusterNode `json:"data"`
}

type ListCommonClusterResp struct {
	Data []*Cluster `json:"data"`
}

type Cluster struct {
	ClusterID            string               `json:"clusterID"`
	ProjectID            string               `json:"projectID"`
	BusinessID           string               `json:"businessID"`
	EngineType           string               `json:"engineType"`
	IsExclusive          bool                 `json:"isExclusive"`
	ClusterType          string               `json:"clusterType"`
	Creator              string               `json:"creator"`
	Updater              string               `json:"updater"`
	ManageType           string               `json:"manageType"`
	ClusterName          string               `json:"clusterName"`
	Environment          string               `json:"environment"`
	Provider             string               `json:"provider"`
	Description          string               `json:"description"`
	ClusterBasicSettings ClusterBasicSettings `json:"clusterBasicSettings"`
	NetworkType          string               `json:"networkType"`
	Region               string               `json:"region"`
	VpcID                string               `json:"vpcID"`
	NetworkSettings      NetworkSettings      `json:"networkSettings"`
	Master               []string             `json:"master"`
}

type ClusterBasicSettings struct {
	Version string `json:"version"`
}

type NetworkSettings struct {
	CidrStep      uint32 `json:"cidrStep"`
	MaxNodePodNum uint32 `json:"maxNodePodNum"`
	MaxServiceNum uint32 `json:"maxServiceNum"`
}

type ImportCloudMode struct {
	CloudID    string `json:"cloudID"`
	KubeConfig string `json:"kubeConfig"`
}

type ClusterNode struct {
	NodeID  string `json:"nodeID"`
	InnerIP string `json:"innerIP"`
}

type ClusterMgr interface {
	// Create 创建集群
	Create(CreateClusterReq) (CreateClusterResp, error)
	// Delete 删除集群
	Delete(DeleteClusterReq) (DeleteClusterResp, error)
	// Update 更新集群
	Update(UpdateClusterReq) error
	// Get 获取集群
	Get(GetClusterReq) (GetClusterResp, error)
	// List 获取集群列表
	List(ListClusterReq) (ListClusterResp, error)
	// Retry 重试创建集群
	RetryCreate(RetryCreateClusterReq) (RetryCreateClusterResp, error)
	// AddNodes 添加节点到集群
	AddNodes(AddNodesClusterReq) (AddNodesClusterResp, error)
	// DeleteNodes 从集群中删除节点
	DeleteNodes(DeleteNodesClusterReq) (DeleteNodesClusterResp, error)
	// CheckCloudKubeConfig kubeConfig连接集群可用性检测
	CheckCloudKubeconfig(CheckCloudKubeconfigReq) error
	// Import 导入用户集群(支持多云集群导入功能: 集群ID/kubeConfig)
	Import(ImportClusterReq) error
	// ListNodes 查询集群下所有节点列表
	ListNodes(ListClusterNodesReq) (ListClusterNodesResp, error)
	// ListCommon 查询公共集群及公共集群所属权限
	ListCommon(ListCommonClusterReq) (ListCommonClusterResp, error)
}
