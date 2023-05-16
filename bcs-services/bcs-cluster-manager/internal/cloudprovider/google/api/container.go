/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package api

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"

	"golang.org/x/oauth2"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

// ContainerServiceClient container service client
type ContainerServiceClient struct {
	gkeProjectID           string
	region                 string
	containerServiceClient *container.Service
}

// NewContainerServiceClient create container service client
func NewContainerServiceClient(opt *cloudprovider.CommonOption) (*ContainerServiceClient, error) {
	if opt == nil || opt.Account == nil {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if opt.Region == "" {
		return nil, cloudprovider.ErrCloudRegionLost
	}
	if len(opt.Account.ServiceAccountSecret) == 0 || opt.Account.GkeProjectID == "" {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	containerServiceClient, err := GetContainerServiceClient(context.Background(), opt.Account.ServiceAccountSecret)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}
	return &ContainerServiceClient{
		gkeProjectID:           opt.Account.GkeProjectID,
		region:                 opt.Region,
		containerServiceClient: containerServiceClient,
	}, nil
}

// GetContainerServiceClient get google container service client
func GetContainerServiceClient(ctx context.Context, credentialContent string) (*container.Service, error) {
	ts, err := GetTokenSource(ctx, credentialContent)
	if err != nil {
		return nil, fmt.Errorf("GetContainerServiceClient GetTokenSource failed: %v", err)
	}

	service, err := container.NewService(ctx, option.WithHTTPClient(oauth2.NewClient(ctx, ts)))
	if err != nil {
		return nil, fmt.Errorf("GetContainerServiceClient create servcie failed: %v", err)
	}

	return service, nil
}

// ListCluster list clusters
func (cs *ContainerServiceClient) ListCluster(ctx context.Context) ([]*container.Cluster, error) {
	parent := "projects/" + cs.gkeProjectID + "/locations/" + cs.region
	clusters, err := cs.containerServiceClient.Projects.Locations.Clusters.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("gke client ListCluster failed: %v", err)
	}

	return clusters.Clusters, nil
}

// GetCluster get cluster
func (cs *ContainerServiceClient) GetCluster(ctx context.Context, clusterName string) (*container.Cluster, error) {
	parent := "projects/" + cs.gkeProjectID + "/locations/" + cs.region + "/clusters/" + clusterName
	cluster, err := cs.containerServiceClient.Projects.Locations.Clusters.Get(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("gke client GetCluster failed: %v", err)
	}

	return cluster, nil
}

// DeleteCluster delete cluster
func (cs *ContainerServiceClient) DeleteCluster(ctx context.Context, clusterName string) error {
	parent := "projects/" + cs.gkeProjectID + "/locations/" + cs.region + "/clusters/" + clusterName
	o, err := cs.containerServiceClient.Projects.Locations.Clusters.Delete(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("gke client DeleteCluster failed: %v", err)
	}
	blog.Infof("gke client DeleteCluster[%s] successful, operation ID: %s", clusterName, o.SelfLink)

	return nil
}

// CreateClusterNodePool create a node pool
func (cs *ContainerServiceClient) CreateClusterNodePool(ctx context.Context, req *CreateNodePoolRequest,
	clusterName string) (string, error) {
	parent := "projects/" + cs.gkeProjectID + "/locations/" + cs.region + "/clusters/" + clusterName
	req.Parent = parent
	newReq := genCreateNodePoolRequest(req)
	o, err := cs.containerServiceClient.Projects.Locations.Clusters.NodePools.Create(parent, newReq).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gke client CreateClusterNodePool failed: %v", err)
	}
	blog.Infof("gke client CreateClusterNodePool[%s] successful, operation ID: %s", req.NodePool.Name, o.SelfLink)

	return o.SelfLink, nil
}

// GetClusterNodePool create the node pool
func (cs *ContainerServiceClient) GetClusterNodePool(ctx context.Context, clusterName, nodePoolName string) (
	*container.NodePool, error) {
	parent := "projects/" + cs.gkeProjectID + "/locations/" + cs.region + "/clusters/" + clusterName +
		"/nodePools/" + nodePoolName
	np, err := cs.containerServiceClient.Projects.Locations.Clusters.NodePools.Get(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("gke client GetClusterNodePool failed: %v", err)
	}

	return np, nil
}

// UpdateClusterNodePool update the node pool
func (cs *ContainerServiceClient) UpdateClusterNodePool(ctx context.Context,
	req *container.UpdateNodePoolRequest, clusterName, nodePoolName string) (string, error) {
	parent := "projects/" + cs.gkeProjectID + "/locations/" + cs.region + "/clusters/" + clusterName +
		"/nodePools/" + nodePoolName
	o, err := cs.containerServiceClient.Projects.Locations.Clusters.NodePools.Update(parent, req).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gke client UpdateClusterNodePool failed: %v", err)
	}
	blog.Infof("gke client UpdateClusterNodePool[%s] successful, operation ID: %s", nodePoolName, o.SelfLink)

	return o.SelfLink, nil
}

// DeleteClusterNodePool update the node pool
func (cs *ContainerServiceClient) DeleteClusterNodePool(ctx context.Context, clusterName, nodePoolName string) (
	string, error) {
	parent := "projects/" + cs.gkeProjectID + "/locations/" + cs.region + "/clusters/" + clusterName +
		"/nodePools/" + nodePoolName
	o, err := cs.containerServiceClient.Projects.Locations.Clusters.NodePools.Delete(parent).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("gke client DeleteClusterNodePool failed: %v", err)
	}
	blog.Infof("gke client DeleteClusterNodePool[%s] successful, operation ID: %s", nodePoolName, o.SelfLink)

	return o.SelfLink, nil
}
