package manager

type InitFederationClusterReq struct {
	FederationClusterID string `json:"federationClusterID"`
	ClusterID           string `json:"clusterID"`
}

type AddFederatedClusterReq struct {
	FederationClusterID string `json:"federationClusterID"`
	ClusterID           string `json:"clusterID"`
}

type FederationClusterMgr interface {
	Init(InitFederationClusterReq) error
	Add(AddFederatedClusterReq) error
}
