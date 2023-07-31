package api

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
)

// EksClient eks client
type EksClient struct {
	*eks.EKS
	Session *session.Session
}

// NewEksClient init Eks client
func NewEksClient(opt *cloudprovider.CommonOption) (*EksClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}

	awsConf := &aws.Config{}
	awsConf.Region = aws.String(opt.Region)
	awsConf.Credentials = credentials.NewStaticCredentials(opt.Account.SecretID, opt.Account.SecretKey, "")

	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, err
	}

	return &EksClient{
		Session: sess,
		EKS:     eks.New(sess),
	}, nil
}

// ListEksCluster get tke cluster list, region parameter init tke client
func (cli *EksClient) ListEksCluster() ([]*string, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	input := &eks.ListClustersInput{}
	output, err := cli.ListClusters(input)
	if err != nil {
		return nil, err
	}

	return output.Clusters, nil
}

// GetEksCluster get eks cluster
func (cli *EksClient) GetEksCluster(clusterName string) (*eks.Cluster, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	input := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}
	output, err := cli.DescribeCluster(input)
	if err != nil {
		return nil, err
	}

	return output.Cluster, nil
}
