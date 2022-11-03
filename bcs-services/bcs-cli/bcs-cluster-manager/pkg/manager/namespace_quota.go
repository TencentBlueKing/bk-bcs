package manager

type CreateNamespaceQuotaReq struct {
}

type UpdateNamespaceQuotaReq struct {
}

type DeleteNamespaceQuotaReq struct {
}

type GetNamespaceQuotaReq struct {
}

type ListNamespaceQuotaReq struct {
}

type CreateNamespaceWithQuotaReq struct {
}

type CreateNamespaceQuotaResp struct {
}

type UpdateNamespaceQuotaResp struct {
}

type DeleteNamespaceQuotaResp struct {
}

type GetNamespaceQuotaResp struct {
}

type ListNamespaceQuotaResp struct {
}

type CreateNamespaceWithQuotaResp struct {
}

type NamespaceQuota interface {
	//创建资源配额
	Create(CreateNamespaceQuotaReq) CreateNamespaceQuotaResp
	//更新资源配额
	Update(UpdateNamespaceQuotaReq) UpdateNamespaceQuotaResp
	//删除资源配额
	Delete(DeleteNamespaceQuotaReq) DeleteNamespaceQuotaResp
	//获取资源配额
	Get(GetNamespaceQuotaReq) GetNamespaceQuotaResp
	//获取资源配额列表
	List(ListNamespaceQuotaReq) ListNamespaceQuotaResp
	//创建带命名空间的资源配额
	CreateNamespaceWithQuota(CreateNamespaceWithQuotaReq) CreateNamespaceWithQuotaResp
}
