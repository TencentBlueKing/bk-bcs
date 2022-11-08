package manager

type CreateNamespaceReq struct {
	Name                string            `json:"name"`
	FederationClusterID string            `json:"federationClusterID"`
	ProjectID           string            `json:"projectID"`
	BusinessID          string            `json:"businessID"`
	Labels              map[string]string `json:"labels"`
}

type UpdateNamespaceReq struct {
	Name                string            `json:"name"`
	FederationClusterID string            `json:"federationClusterID"`
	Labels              map[string]string `json:"labels"`
}

type DeleteNamespaceReq struct {
	Name                string `json:"name"`
	FederationClusterID string `json:"federationClusterID"`
	IsForced            bool   `json:"isForced"`
}

type GetNamespaceReq struct {
	Name                string `json:"name"`
	FederationClusterID string `json:"federationClusterID"`
}

type ListNamespaceReq struct {
	FederationClusterID string `json:"federationClusterID"`
	ProjectID           string `json:"projectID"`
	BusinessID          string `json:"businessID"`
	Offset              uint32 `json:"offset"`
	Limit               uint32 `json:"limit"`
}

type GetNamespaceResp struct {
	Data *Namespace `json:"data"`
}

type ListNamespaceResp struct {
	Data []*Namespace `json:"data"`
}

type Namespace struct {
	Name                string            `json:"name"`
	FederationClusterID string            `json:"federationClusterID"`
	ProjectID           string            `json:"projectID"`
	BusinessID          string            `json:"businessID"`
	Labels              map[string]string `json:"labels"`
}

type NamespaceMgr interface {
	//创建命名空间
	Create(CreateNamespaceReq) error
	//更新命名空间
	Update(UpdateNamespaceReq) error
	//删除命名空间
	Delete(DeleteNamespaceReq) error
	//查询命名空间
	Get(GetNamespaceReq) (GetNamespaceResp, error)
	//查询命名空间列表
	List(ListNamespaceReq) (ListNamespaceResp, error)
}
