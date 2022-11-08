package manager

type GetClusterCredentialReq struct {
	ServerKey string `json:"serverKey"`
}

type UpdateClusterCredentialReq struct {
	ClusterID     string `json:"clusterID"`
	ClientModule  string `json:"clientModule"`
	ServerAddress string `json:"serverAddress"`
	CaCertData    string `json:"caCertData"`
	UserToken     string `json:"userToken"`
	ClusterDomain string `json:"clusterDomain"`
}

type DeleteClusterCredentialReq struct {
	ServerKey string `json:"serverKey"`
}

type ListClusterCredentialReq struct {
	Offset uint32 `json:"offset"`
	Limit  uint32 `json:"limit"`
}

type GetClusterCredentialResp struct {
	Data ClusterCredential `json:"data"`
}

type ListClusterCredentialResp struct {
	Data []*ClusterCredential `json:"data"`
}

type ClusterCredential struct {
	ServerKey     string `json:"serverKey"`
	ClusterID     string `json:"clusterID"`
	ClientModule  string `json:"clientModule"`
	ServerAddress string `json:"serverAddress"`
	CaCertData    string `json:"caCertData"`
	UserToken     string `json:"userToken"`
	ClusterDomain string `json:"clusterDomain"`
	ConnectMode   string `json:"connectMode"`
	CreateTime    string `json:"createTime"`
	UpdateTime    string `json:"updateTime"`
	ClientCert    string `json:"clientCert"`
	ClientKey     string `json:"clientKey"`
}

type ClusterCredentialMgr interface {
	//根据提供的ServerKey查询集群凭证详情
	Get(GetClusterCredentialReq) (GetClusterCredentialResp, error)
	//指定更新集群凭证
	Update(UpdateClusterCredentialReq) error
	//删除集群凭证
	Delete(DeleteClusterCredentialReq) error
	//根据查询条件查询集群凭证列表
	List(ListClusterCredentialReq) (ListClusterCredentialResp, error)
}
