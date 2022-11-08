package manager

type CreateProjectReq struct {
	Name        string `json:"name"`
	EnglishName string `json:"englishName"`
	Creator     string `json:"creator"`
	ProjectType uint32 `json:"projectType"`
	UseBKRes    bool   `json:"useBKRes"`
	BusinessID  string `json:"businessID"`
	Kind        string `json:"kind"`
	DeployType  uint32 `json:"deployType"`
}

type UpdateProjectReq struct {
	ProjectID   string                `json:"projectID"`
	Name        string                `json:"name"`
	Updater     string                `json:"updater"`
	ProjectType uint32                `json:"projectType"`
	UseBKRes    bool                  `json:"useBKRes"`
	BusinessID  string                `json:"businessID"`
	Description string                `json:"description"`
	IsOffline   bool                  `json:"isOffline"`
	Kind        string                `json:"kind"`
	DeployType  uint32                `json:"deployType"`
	BgID        string                `json:"bgID"`
	BgName      string                `json:"bgName"`
	DeptID      string                `json:"deptID"`
	DeptName    string                `json:"deptName"`
	CenterID    string                `json:"centerID"`
	CenterName  string                `json:"centerName"`
	IsSecret    bool                  `json:"isSecret"`
	Credentials map[string]Credential `json:"credentials"`
}

type DeleteProjectReq struct {
	ProjectID string `json:"projectID"`
	IsForce   bool   `json:"isForce"`
}

type GetProjectReq struct {
	ProjectID string `json:"projectID"`
}

type ListProjectReq struct {
}

type GetProjectResp struct {
	Data Project `json:"data"`
}

type ListProjectResp struct {
	Data []*Project `json:"data"`
}

type Project struct {
	ProjectID   string                `json:"projectID"`
	Name        string                `json:"name"`
	EnglishName string                `json:"englishName"`
	Creator     string                `json:"creator"`
	Updater     string                `json:"updater"`
	ProjectType uint32                `json:"projectType"`
	UseBKRes    bool                  `json:"useBKRes"`
	BusinessID  string                `json:"businessID"`
	Description string                `json:"description"`
	IsOffline   bool                  `json:"isOffline"`
	Kind        string                `json:"kind"`
	DeployType  uint32                `json:"deployType"`
	BgID        string                `json:"bgID"`
	BgName      string                `json:"bgName"`
	DeptID      string                `json:"deptID"`
	DeptName    string                `json:"deptName"`
	CenterID    string                `json:"centerID"`
	CenterName  string                `json:"centerName"`
	IsSecret    bool                  `json:"isSecret"`
	Credentials map[string]Credential `json:"credentials"`
	CreatTime   string                `json:"creatTime"`
	UpdateTime  string                `json:"updateTime"`
}

type Credential struct {
	Key               string `json:"key"`
	Secret            string `json:"secret"`
	SubscriptionID    string `json:"subscriptionID"`
	TenantID          string `json:"tenantID"`
	ResourceGroupName string `json:"resourceGroupName"`
	ClientID          string `json:"clientID"`
	ClientSecret      string `json:"clientSecret"`
}

type ProjectMgr interface {
	Create(CreateProjectReq) error
	Update(UpdateProjectReq) error
	Delete(DeleteProjectReq) error
	Get(GetProjectReq) (GetProjectResp, error)
	List(ListProjectReq) (ListProjectResp, error)
}
