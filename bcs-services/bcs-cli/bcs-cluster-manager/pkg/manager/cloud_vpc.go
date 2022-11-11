package manager

type CreateCloudVPCReq struct {
	CloudID     string `json:"cloudID"`
	NetworkType string `json:"networkType"`
	Region      string `json:"region"`
	VPCName     string `json:"vpcName"`
	VPCID       string `json:"vpcID"`
}

type UpdateCloudVPCReq struct {
	CloudID     string `json:"cloudID"`
	NetworkType string `json:"networkType"`
	Region      string `json:"region"`
	RegionName  string `json:"regionName"`
	VPCName     string `json:"vpcName"`
	VPCID       string `json:"vpcID"`
	Available   string `json:"available"`
	Updater     string `json:"updater"`
}

type DeleteCloudVPCReq struct {
	CloudID string `json:"cloudID"`
	VPCID   string `json:"vpcID"`
}

type ListCloudRegionsReq struct {
	CloudID string `json:"cloudID"`
}

type GetVPCCidrReq struct {
	VPCID string `json:"vpcID"`
}

type ListCloudVPCResp struct {
	Data []*CloudVPC `json:"data"`
}

type ListCloudRegionsResp struct {
	Data []*CloudRegion `json:"data"`
}

type GetVPCCidrResp struct {
	Data []*VPCCidr `json:"data"`
}

type CloudVPC struct {
	CloudID     string `json:"cloudID"`
	Region      string `json:"region"`
	RegionName  string `json:"regionName"`
	NetworkType string `json:"networkType"`
	VPCID       string `json:"vpcID"`
	VPCName     string `json:"vpcName"`
	Available   string `json:"available"`
	Extra       string `json:"extra"`
	Creator     string `json:"creator"`
	Updater     string `json:"updater"`
	CreatTime   string `json:"creatTime"`
	UpdateTime  string `json:"updateTime"`
}

type CloudRegion struct {
	CloudID    string `json:"cloudID"`
	Region     string `json:"region"`
	RegionName string `json:"regionName"`
}

type VPCCidr struct {
	VPC      string `json:"vpc"`
	Cidr     string `json:"cidr"`
	IPNumber uint32 `json:"ipNumber"`
	Status   string `json:"status"`
}

type CloudVPCMgr interface {
	// Create 创建云VPC管理信息
	Create(CreateCloudVPCReq) error
	// Update 更新云vpc信息
	Update(UpdateCloudVPCReq) error
	// Delete 删除特定cloud vpc信息
	Delete(DeleteCloudVPCReq) error
	// List 查询Cloud VPC列表
	List() (ListCloudVPCResp, error)
	// ListCloudRegions 根据cloudID获取所属cloud的地域信息
	ListCloudRegions(ListCloudRegionsReq) (ListCloudRegionsResp, error)
	// GetVPCCidr 根据vpcID获取所属vpc的cidr信息
	GetVPCCidr(GetVPCCidrReq) (GetVPCCidrResp, error)
}
