/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
)

// EksClient aws eks client
type EksClient struct {
	eksClient *eks.EKS
}

// NewEksClient init Eks client
func NewEksClient(opt *cloudprovider.CommonOption) (*EksClient, error) {
	sess, err := NewSession(opt)
	if err != nil {
		return nil, err
	}

	return &EksClient{
		eksClient: eks.New(sess),
	}, nil
}

// ListEksCluster get eks cluster list, region parameter init eks client
func (cli *EksClient) ListEksCluster() ([]*string, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	input := &eks.ListClustersInput{}
	output, err := cli.eksClient.ListClusters(input)
	if err != nil {
		return nil, err
	}

	return output.Clusters, nil
}

// GetEksCluster  gets the eks cluster
func (cli *EksClient) GetEksCluster(clusterName string) (*eks.Cluster, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	input := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}
	output, err := cli.eksClient.DescribeCluster(input)
	if err != nil {
		return nil, err
	}

	return output.Cluster, nil
}

// CreateNodegroup creates eks node group
func (cli *EksClient) CreateNodegroup(input *CreateNodegroupInput) (*eks.Nodegroup, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	newInput := generateAwsCreateNodegroupInput(input)
	blog.Infof("CreateNodegroup request: %s", utils.ToJSONString(newInput))
	output, err := cli.eksClient.CreateNodegroup(newInput)
	if err != nil {
		return nil, err
	}
	if output == nil || output.Nodegroup == nil {
		blog.Errorf("CreateNodegroup resp is nil")
		return nil, fmt.Errorf("CreateNodegroup resp is nil")
	}
	blog.Infof("CreateNodegroup create nodegroup[%s] successful", *output.Nodegroup.NodegroupName)

	return output.Nodegroup, nil
}

// DescribeNodegroup gets eks node group info
func (cli *EksClient) DescribeNodegroup(ngName, clusterName *string) (*eks.Nodegroup, error) {
	blog.Infof("DescribeNodegroup[%s] in cluster %s", *ngName, *clusterName)
	output, err := cli.eksClient.DescribeNodegroup(&eks.DescribeNodegroupInput{
		NodegroupName: ngName,
		ClusterName:   clusterName,
	},
	)
	if err != nil {
		return nil, err
	}
	if output == nil || output.Nodegroup == nil {
		blog.Errorf("DescribeNodegroup resp is nil")
		return nil, fmt.Errorf("DescribeNodegroup resp is nil")
	}
	blog.Infof("DescribeNodegroup[%s] successful", *output.Nodegroup.NodegroupName)

	return output.Nodegroup, nil
}

// UpdateNodegroupConfig gets eks node group info
func (cli *EksClient) UpdateNodegroupConfig(input *eks.UpdateNodegroupConfigInput) (*eks.Update, error) {
	blog.Infof("UpdateNodegroupConfig request: %s", utils.ToJSONString(input))
	output, err := cli.eksClient.UpdateNodegroupConfig(input)
	if err != nil {
		return nil, err
	}
	if output == nil || output.Update == nil {
		blog.Errorf("UpdateNodegroupConfig resp is nil")
		return nil, fmt.Errorf("UpdateNodegroupConfig resp is nil")
	}
	blog.Infof("UpdateNodegroupConfig[%s] successful, update id %s", *input.NodegroupName, *output.Update.Id)

	return output.Update, nil
}

// DeleteNodegroup deletes eks node group
func (cli *EksClient) DeleteNodegroup(input *eks.DeleteNodegroupInput) (*eks.Nodegroup, error) {
	blog.Infof("DeleteNodegroup request: %s", utils.ToJSONString(input))
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}
	output, err := cli.eksClient.DeleteNodegroup(input)
	if err != nil {
		return nil, err
	}
	if output == nil || output.Nodegroup == nil {
		blog.Errorf("DeleteNodegroup resp is nil")
		return nil, fmt.Errorf("DeleteNodegroup resp is nil")
	}
	blog.Infof("DeleteNodegroup delete nodegroup[%s] successful", *output.Nodegroup.NodegroupName)

	return output.Nodegroup, nil
}
