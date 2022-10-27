package manager

type GetClusterCredentialReq struct {
}

type UpdateClusterCredentialReq struct {
}

type DeleteClusterCredentialReq struct {
}

type ListClusterCredentialReq struct {
}

type GetClusterCredentialResp struct {
}

type UpdateClusterCredentialResp struct {
}

type DeleteClusterCredentialResp struct {
}

type ListClusterCredentialResp struct {
}

type ClusterCredential interface {
	//根据提供的ServerKey查询集群凭证详情
	Get(GetClusterCredentialReq) GetClusterCredentialResp
	//指定更新集群凭证
	Update(UpdateClusterCredentialReq) UpdateClusterCredentialResp
	//删除集群凭证
	Delete(DeleteClusterCredentialReq) DeleteClusterCredentialResp
	//根据查询条件查询集群凭证列表
	List(ListClusterCredentialReq) ListClusterCredentialResp
}
