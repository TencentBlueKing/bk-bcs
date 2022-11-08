package manager

type CreateNamespaceQuotaReq struct {
	Namespace           string `json:"namespace"`
	FederationClusterID string `json:"federationClusterID"`
	ResourceQuota       string `json:"resourceQuota"`
}

type UpdateNamespaceQuotaReq struct {
	ClusterID           string `json:"clusterID"`
	Namespace           string `json:"namespace"`
	FederationClusterID string `json:"federationClusterID"`
	ResourceQuota       string `json:"resourceQuota"`
}

type DeleteNamespaceQuotaReq struct {
	ClusterID           string `json:"clusterID"`
	Namespace           string `json:"namespace"`
	FederationClusterID string `json:"federationClusterID"`
}

type GetNamespaceQuotaReq struct {
	ClusterID           string `json:"clusterID"`
	Namespace           string `json:"namespace"`
	FederationClusterID string `json:"federationClusterID"`
}

type ListNamespaceQuotaReq struct {
	Namespace           string `json:"namespace"`
	FederationClusterID string `json:"federationClusterID"`
	Offset              uint32 `json:"offset"`
	Limit               uint32 `json:"limit"`
}

type CreateNamespaceWithQuotaReq struct {
	Name                string            `json:"name"`
	FederationClusterID string            `json:"federationClusterID"`
	ProjectID           string            `json:"projectID"`
	BusinessID          string            `json:"businessID"`
	Labels              map[string]string `json:"labels"`
	Region              string            `json:"region"`
	ResourceQuota       string            `json:"resourceQuota"`
}

type CreateNamespaceQuotaResp struct {
	ClusterID string `json:"clusterID"`
}

type GetNamespaceQuotaResp struct {
	Data ResourceQuota `json:"data"`
}

type ListNamespaceQuotaResp struct {
	Data []*ResourceQuota `json:"data"`
}

type CreateNamespaceWithQuotaResp struct {
	ClusterID string `json:"clusterID"`
}

type ResourceQuota struct {
	Namespace           string `json:"namespace"`
	FederationClusterID string `json:"federationClusterID"`
	ClusterID           string `json:"clusterID"`
	ResourceQuota       string `json:"resourceQuota"`
	Region              string `json:"region"`
	CreateTime          string `json:"createTime"`
	UpdateTime          string `json:"updateTime"`
	Status              string `json:"status"`
	Message             string `json:"message"`
}

type NamespaceQuotaMgr interface {
	//创建资源配额
	Create(CreateNamespaceQuotaReq) (CreateNamespaceQuotaResp, error)
	//更新资源配额
	Update(UpdateNamespaceQuotaReq) error
	//删除资源配额
	Delete(DeleteNamespaceQuotaReq) error
	//获取资源配额
	Get(GetNamespaceQuotaReq) (GetNamespaceQuotaResp, error)
	//获取资源配额列表
	List(ListNamespaceQuotaReq) (ListNamespaceQuotaResp, error)
	//创建带命名空间的资源配额
	CreateNamespaceWithQuota(CreateNamespaceWithQuotaReq) (CreateNamespaceWithQuotaResp, error)
}
