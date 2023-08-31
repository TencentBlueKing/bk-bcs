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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	cutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	gceURLPrefix = "https://www.googleapis.com/compute/v1/"
)

const (
	UpdatePolicyProactive = "PROACTIVE"
	// UpdatePolicyOpportunistic update is opportunistic
	UpdatePolicyOpportunistic = "OPPORTUNISTIC"
	UpdatePolicyActionNone    = "None"
	// UpdatePolicyActionRefresh update action refresh
	UpdatePolicyActionRefresh = "REFRESH"
	UpdatePolicyActionRestart = "RESTART"
	UpdatePolicyActionReplace = "REPLACE"
)

// GCPClientSet google cloud platform client set
type GCPClientSet struct {
	*ComputeServiceClient
	*ContainerServiceClient
}

// NewGCPClientSet creates a GCP client set
func NewGCPClientSet(opt *cloudprovider.CommonOption) (*GCPClientSet, error) {
	computeCli, err := NewComputeServiceClient(opt)
	if err != nil {
		return nil, err
	}
	containerCli, err := NewContainerServiceClient(opt)
	if err != nil {
		return nil, err
	}
	return &GCPClientSet{computeCli, containerCli}, nil
}

// GetTokenSource gets token source from provided sa credential
func GetTokenSource(ctx context.Context, credential string) (oauth2.TokenSource, error) {
	ts, err := google.CredentialsFromJSON(ctx, []byte(credential), container.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("GetTokenSource failed: %v", err)
	}
	return ts.TokenSource, nil
}

// GetClusterKubeConfig get cloud cluster's kube config
func GetClusterKubeConfig(ctx context.Context, saSecret, gkeProjectID, region, clusterName string) (string, error) {
	client, err := GetContainerServiceClient(ctx, saSecret)
	if err != nil {
		return "", err
	}

	// Get the kube cluster in given project.
	parent := "projects/" + gkeProjectID + "/locations/" + region + "/clusters/" + clusterName
	gkeCluster, err := client.Projects.Locations.Clusters.Get(parent).Do()
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig list clusters failed, project=%s: %v", gkeProjectID, err)
	}

	name := fmt.Sprintf("%s_%s_%s", gkeProjectID, gkeCluster.Location, gkeCluster.Name)

	cert, err := base64.StdEncoding.DecodeString(gkeCluster.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig invalid certificate failed, cluster=%s: %v", name, err)
	}

	restConfig := &rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			CAData: cert,
		},
		Host: "https://" + gkeCluster.Endpoint,
		AuthProvider: &api.AuthProviderConfig{
			Name: GoogleAuthPlugin,
			Config: map[string]string{
				"scopes":      "https://www.googleapis.com/auth/cloud-platform",
				"credentials": saSecret,
			},
		},
	}

	var saToken string
	saToken, err = cutils.GenerateSATokenByRestConfig(ctx, restConfig)
	if err != nil {
		return "", fmt.Errorf("getClusterKubeConfig generate k8s serviceaccount token failed, "+
			"project=%s cluster=%s: %v", gkeProjectID, clusterName, err)
	}

	typesConfig := &types.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []types.NamedCluster{
			{
				Name: name,
				Cluster: types.ClusterInfo{
					Server:                   "https://" + gkeCluster.Endpoint,
					CertificateAuthorityData: cert,
				},
			},
		},
		AuthInfos: []types.NamedAuthInfo{
			{
				Name: name,
				AuthInfo: types.AuthInfo{
					Token: saToken,
				},
			},
		},
		Contexts: []types.NamedContext{
			{
				Name: name,
				Context: types.Context{
					Cluster:  name,
					AuthInfo: name,
				},
			},
		},
		CurrentContext: name,
	}

	configByte, err := json.Marshal(typesConfig)
	if err != nil {
		return "", fmt.Errorf("GetClusterKubeConfig marsh kubeconfig failed, %v", err)
	}
	return base64.StdEncoding.EncodeToString(configByte), nil
}

// MapTaints map cmproto.Taint to Taint
func MapTaints(cmt []*cmproto.Taint) []*Taint {
	t := make([]*Taint, 0)
	for _, v := range cmt {
		t = append(t, &Taint{
			Key:    v.Key,
			Value:  v.Value,
			Effect: v.Effect,
		})
	}
	return t
}

func genCreateNodePoolRequest(req *CreateNodePoolRequest) *container.CreateNodePoolRequest {
	if req == nil || req.NodePool == nil {
		return nil
	}
	newReq := &container.CreateNodePoolRequest{
		NodePool: &container.NodePool{
			Name:             req.NodePool.Name,
			InitialNodeCount: req.NodePool.InitialNodeCount,
			MaxPodsConstraint: &container.MaxPodsConstraint{
				MaxPodsPerNode: req.NodePool.MaxPodsConstraint.MaxPodsPerNode,
			},
			Autoscaling: &container.NodePoolAutoscaling{
				Enabled: false,
			},
		},
	}
	if req.NodePool.Config != nil {
		newReq.NodePool.Config = generateNodeConfig(req.NodePool.Config)
	}
	if req.NodePool.Management != nil {
		newReq.NodePool.Management = &container.NodeManagement{
			AutoRepair:  req.NodePool.Management.AutoRepair,
			AutoUpgrade: req.NodePool.Management.AutoUpgrade,
		}
	}
	return newReq
}

func generateNodeConfig(nc *NodeConfig) *container.NodeConfig {
	conf := &container.NodeConfig{}
	conf.MachineType = nc.MachineType
	conf.Labels = nc.Labels
	conf.Taints = func(t []*Taint) []*container.NodeTaint {
		nt := make([]*container.NodeTaint, 0)
		for _, v := range t {
			nt = append(nt, &container.NodeTaint{
				Key:    v.Key,
				Value:  v.Value,
				Effect: v.Effect,
			})
		}
		return nt
	}(nc.Taints)
	conf.DiskType = nc.DiskType
	conf.DiskSizeGb = nc.DiskSizeGb
	conf.ImageType = nc.ImageType
	return conf
}

// GetGCEResourceInfo get resource info from url
func GetGCEResourceInfo(url string) ([]string, error) {
	if !strings.HasPrefix(url, gceURLPrefix) {
		return nil, fmt.Errorf("GetGCEResourceInfo failed, %s doesn't start wirth %s", url, gceURLPrefix)
	}
	url = strings.TrimPrefix(url, gceURLPrefix)
	ri := strings.Split(url, "/")
	return ri, nil
}

// GetInstanceGroupManager get zonal/regional InstanceGroupManager
func GetInstanceGroupManager(computeCli *ComputeServiceClient, url string) (*compute.InstanceGroupManager, error) {
	igmInfo, err := GetGCEResourceInfo(url)
	if err != nil {
		blog.Errorf("GetInstanceGroupManager failed: %v", err)
		return nil, err
	}
	var igm *compute.InstanceGroupManager
	if utils.StringInSlice("instanceGroupManagers", igmInfo) && len(igmInfo) >= 6 {
		igm, err = computeCli.GetInstanceGroupManager(context.Background(), igmInfo[2], igmInfo[(len(igmInfo)-1)])
		if err != nil {
			blog.Errorf("GetInstanceGroupManager failed: %v", err)
			return nil, err
		}
		return igm, nil
	}
	return nil, fmt.Errorf("GetInstanceGroupManager failed, incorrect InstanceGroupManager url: %s", url)
}

// PatchInstanceGroupManager patch zonal/regional InstanceGroupManager
func PatchInstanceGroupManager(computeCli *ComputeServiceClient, url string, igm *compute.InstanceGroupManager) (
	*compute.Operation, error) {
	igmInfo, err := GetGCEResourceInfo(url)
	if err != nil {
		blog.Errorf("PatchInstanceGroupManager failed: %v", err)
		return nil, err
	}
	if utils.StringInSlice("instanceGroupManagers", igmInfo) && len(igmInfo) >= 6 {
		o, err := computeCli.PatchInstanceGroupManager(context.Background(), igmInfo[2], igmInfo[(len(igmInfo)-1)], igm)
		if err != nil {
			blog.Errorf("PatchInstanceGroupManager failed, operation: %s, err: %v", o.SelfLink, err)
			return nil, err
		}
		return o, nil
	}
	return nil, fmt.Errorf("PatchInstanceGroupManager failed, incorrect InstanceGroupManager url: %s", url)
}

// ResizeInstanceGroupManager resize zonal/regional InstanceGroupManager
func ResizeInstanceGroupManager(computeCli *ComputeServiceClient, url string, size int64) (*compute.Operation, error) {
	igmInfo, err := GetGCEResourceInfo(url)
	if err != nil {
		blog.Errorf("ResizeInstanceGroupManager failed: %v", err)
		return nil, err
	}
	if utils.StringInSlice("instanceGroupManagers", igmInfo) && len(igmInfo) >= 6 {
		var o *compute.Operation
		o, err = computeCli.ResizeInstanceGroupManager(context.Background(), igmInfo[2],
			igmInfo[(len(igmInfo)-1)], size)
		if err != nil {
			blog.Errorf("ResizeInstanceGroupManager failed, operation: %s, err: %v", o.SelfLink, err)
			return nil, err
		}
		return o, nil
	}

	return nil, fmt.Errorf("ResizeInstanceGroupManager failed, incorrect InstanceGroupManager url: %s", url)
}

// GetInstanceTemplate get zonal/regional InstanceTemplate
func GetInstanceTemplate(computeCli *ComputeServiceClient, url string) (*compute.InstanceTemplate, error) {
	itInfo, err := GetGCEResourceInfo(url)
	if err != nil {
		blog.Errorf("GetInstanceTemplate failed: %v", err)
		return nil, err
	}
	var it *compute.InstanceTemplate
	if utils.StringInSlice("instanceTemplates", itInfo) {
		it, err = computeCli.GetInstanceTemplate(context.Background(), itInfo[(len(itInfo)-1)])
		if err != nil {
			blog.Errorf("GetInstanceTemplate failed: %v", err)
			return nil, err
		}
		return it, nil
	}
	return nil, fmt.Errorf("GetInstanceTemplate failed, incorrect InstanceTemplate url: %s", url)
}

// GetOperation get zonal/regional/global Operation
func GetOperation(computeCli *ComputeServiceClient, url string) (*compute.Operation, error) {
	opInfo, err := GetGCEResourceInfo(url)
	if err != nil {
		blog.Errorf("GetOperation failed: %v", err)
		return nil, err
	}
	var o *compute.Operation
	if utils.StringInSlice("operations", opInfo) && len(opInfo) >= 5 {
		o, err = computeCli.GetOperation(context.Background(), opInfo[2], opInfo[(len(opInfo)-1)])
		if err != nil {
			blog.Errorf("GetOperation failed, operation: %s, err: %v", o, err)
			return nil, err
		}
		return o, nil
	}
	return nil, fmt.Errorf("GetOperation failed, incorrect Operation url: %s", url)
}

// ListInstanceGroupsInstances list zonal/regional InstanceGroupsInstances
func ListInstanceGroupsInstances(computeCli *ComputeServiceClient, url string) (
	[]*compute.InstanceWithNamedPorts, error) {
	igInfo, err := GetGCEResourceInfo(url)
	if err != nil {
		blog.Errorf("ListInstanceGroupsInstances failed: %v", err)
		return nil, err
	}
	var instances []*compute.InstanceWithNamedPorts
	if (utils.StringInSlice("instanceGroups", igInfo) || utils.StringInSlice("instanceGroupManagers", igInfo)) &&
		len(igInfo) >= 6 {
		instances, err = computeCli.ListInstanceGroupsInstances(context.Background(), igInfo[2],
			igInfo[(len(igInfo)-1)])
		if err != nil {
			blog.Errorf("ListInstanceGroupsInstances failed: %v", err)
			return nil, err
		}
		return instances, nil
	}
	return nil, fmt.Errorf("ListInstanceGroupsInstances failed, incorrect InstanceGroup url: %s", url)
}

// GenerateUpdatePolicy generate update policy
func GenerateUpdatePolicy(group *cmproto.NodeGroup) *compute.InstanceGroupManagerUpdatePolicy {
	p := &compute.InstanceGroupManagerUpdatePolicy{
		Type:          UpdatePolicyOpportunistic,
		MinimalAction: UpdatePolicyActionRefresh,
	}
	if group.AutoScaling == nil {
		return p
	}

	return p
}
