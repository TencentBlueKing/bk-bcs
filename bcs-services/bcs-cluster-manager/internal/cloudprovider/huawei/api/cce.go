package api

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	cce "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/region"
)

// CceClient cce client
type CceClient struct {
	*cce.CceClient
}

// NewCceClient init cce client
func NewCceClient(opt *cloudprovider.CommonOption) (*CceClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}

	auth := basic.NewCredentialsBuilder().WithAk(opt.Account.SecretID).WithSk(opt.Account.SecretKey).
		WithProjectId(opt.Account.HwCCEProjectID).Build()
	// 创建CCE client
	client := cce.NewCceClient(
		cce.CceClientBuilder().WithCredential(auth).WithRegion(region.ValueOf(opt.Region)).Build(),
	)

	return &CceClient{
		CceClient: client,
	}, nil
}

// ListCceCluster get cce cluster list, region parameter init tke client
func (cli *CceClient) ListCceCluster() (*model.ListClustersResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := model.ListClustersRequest{}
	rsp, err := cli.ListClusters(&req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// GetCceCluster get cce cluster
func (cli *CceClient) GetCceCluster(clusterID string) (*model.ShowClusterResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := model.ShowClusterRequest{
		ClusterId: clusterID,
	}
	rsp, err := cli.ShowCluster(&req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}
