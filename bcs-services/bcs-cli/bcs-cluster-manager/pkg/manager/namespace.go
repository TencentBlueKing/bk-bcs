package manager

type CreateNamespaceReq struct {
}

type UpdateNamespaceReq struct {
}

type DeleteNamespaceReq struct {
}

type GetNamespaceReq struct {
}

type ListNamespaceReq struct {
}

type CreateNamespaceResp struct {
}

type UpdateNamespaceResp struct {
}

type DeleteNamespaceResp struct {
}

type GetNamespaceResp struct {
}

type ListNamespaceResp struct {
}

type Namespace interface {
	//创建命名空间
	Create(CreateNamespaceReq) CreateNamespaceResp
	//更新命名空间
	Update(UpdateNamespaceReq) UpdateNamespaceResp
	//删除命名空间
	Delete(DeleteNamespaceReq) DeleteNamespaceResp
	//查询命名空间
	Get(GetNamespaceReq) GetNamespaceResp
	//查询命名空间列表
	List(ListNamespaceReq) ListNamespaceResp
}
