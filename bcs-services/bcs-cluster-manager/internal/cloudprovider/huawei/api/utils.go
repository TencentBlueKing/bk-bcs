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
 */

// Package api xxx
package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	iamModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// GetInternalClusterKubeConfig get cce cluster kebeconfig
func GetInternalClusterKubeConfig(client *CceClient, clusterId string) (string, error) {
	req := model.CreateKubernetesClusterCertRequest{
		ClusterId: clusterId, // 集群ID，可在CCE管理控制台中查看
		Body: &model.CertDuration{
			Duration: int32(-1), // 集群证书有效时间，单位为天，最小值为1，最大值为10950(30*365，1年固定计365天，忽略闰年影响)；若填-1则为最大值30年。
		},
	}

	rsp, err := client.CreateKubernetesClusterCert(&req)
	if err != nil {
		return "", err
	}

	currentContext := "internal"
	clusters := make([]model.Clusters, 0)
	contexts := make([]model.Contexts, 0)

	for _, v := range *rsp.Clusters {
		if *v.Name == "internalCluster" {
			clusters = append(clusters, v)
		}
	}

	if len(clusters) == 0 {
		return "", fmt.Errorf("internal cluster not found")
	}

	for _, v := range *rsp.Contexts {
		if *v.Name == "internal" {
			contexts = append(contexts, v)
		}
	}

	if len(contexts) == 0 {
		return "", fmt.Errorf("internal context not found")
	}

	kubeCfg := &model.CreateKubernetesClusterCertResponse{
		Kind:           rsp.Kind,
		ApiVersion:     rsp.ApiVersion,
		Preferences:    rsp.Preferences,
		Clusters:       &clusters,
		Users:          rsp.Users,
		Contexts:       &contexts,
		CurrentContext: &currentContext,
		PortID:         rsp.PortID,
	}

	bt, err := json.Marshal(kubeCfg)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bt), nil
}

// GetClusterKubeConfig get cce cluster kebeconfig
func GetClusterKubeConfig(client *CceClient, clusterId string) (string, error) {
	req := model.CreateKubernetesClusterCertRequest{
		ClusterId: clusterId, // 集群ID，可在CCE管理控制台中查看
		Body: &model.CertDuration{
			Duration: int32(-1), // 集群证书有效时间，单位为天，最小值为1，最大值为10950(30*365，1年固定计365天，忽略闰年影响)；若填-1则为最大值30年。
		},
	}

	rsp, err := client.CreateKubernetesClusterCert(&req)
	if err != nil {
		return "", err
	}

	kubeCfg := &model.CreateKubernetesClusterCertResponse{}
	if len(*rsp.Clusters) == 1 {
		kubeCfg = rsp
	} else if len(*rsp.Clusters) > 1 && *rsp.CurrentContext == "external" {
		curContext := "externalTLSVerify"
		clusters := make([]model.Clusters, 0)
		contexts := make([]model.Contexts, 0)

		for _, v := range *rsp.Clusters {
			if *v.Name == "externalClusterTLSVerify" {
				clusters = append(clusters, v)
			}
		}

		for _, v := range *rsp.Contexts {
			if *v.Name == "externalTLSVerify" {
				contexts = append(contexts, v)
			}
		}
		kubeCfg = &model.CreateKubernetesClusterCertResponse{
			Kind:           rsp.Kind,
			ApiVersion:     rsp.ApiVersion,
			Preferences:    rsp.Preferences,
			Clusters:       &clusters,
			Users:          rsp.Users,
			Contexts:       &contexts,
			CurrentContext: &curContext,
			PortID:         rsp.PortID,
		}
	}

	bt, err := json.Marshal(kubeCfg)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bt), nil
}

// GetProjectIDByRegion get project ID by region
func GetProjectIDByRegion(opt *cloudprovider.CommonOption) (string, error) {
	client, err := GetIamClient(opt)
	if err != nil {
		return "", err
	}

	req := iamModel.KeystoneListProjectsRequest{Name: &opt.Region}
	rsp, err := client.KeystoneListProjects(&req)
	if err != nil {
		return "", err
	}

	if len(*rsp.Projects) == 0 {
		return "", fmt.Errorf("project not found")
	} else if len(*rsp.Projects) > 1 {
		return "", fmt.Errorf("the number of project is greater than one")
	}

	return (*rsp.Projects)[0].Id, nil
}

// Crypt encryption node password
func Crypt(password string) (string, error) {
	str, err := crypt.SHA512.New().Generate([]byte(password), []byte("$6$tM3|cY3+tI4)"))
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString([]byte(str)), nil
}

// GenerateModifyClusterNodePoolInput get cce update node pool input
func GenerateModifyClusterNodePoolInput(group *proto.NodeGroup, clusterID string,
	oldNodePool *model.ShowNodePoolResponse) *model.UpdateNodePoolRequest {
	// cce nodePool名称以小写字母开头，由小写字母、数字、中划线(-)组成，长度范围1-50位，且不能以中划线(-)结尾
	name := strings.ToLower(group.Name)

	req := &model.UpdateNodePoolRequest{
		NodepoolId: group.CloudNodeGroupID,
		ClusterId:  clusterID,
		Body: &model.NodePoolUpdate{
			Metadata: &model.NodePoolMetadataUpdate{
				Name: name,
			},
			Spec: &model.NodePoolSpecUpdate{
				NodeTemplate: &model.NodeSpecUpdate{
					Taints:   make([]model.Taint, 0),
					K8sTags:  map[string]string{},
					UserTags: make([]model.UserTag, 0),
				},
				//更新节点池不能更新节点数量,只能通过UpdateDesiredNodes方法更新,会影响互斥性
				InitialNodeCount: *oldNodePool.Spec.InitialNodeCount,
				Autoscaling:      &model.NodePoolNodeAutoscaling{},
			},
		},
	}

	if group.NodeTemplate != nil {
		for _, v := range group.NodeTemplate.Taints {
			effect := model.GetTaintEffectEnum().NO_SCHEDULE
			if v.Effect == "PreferNoSchedule" {
				effect = model.GetTaintEffectEnum().PREFER_NO_SCHEDULE
			} else if v.Effect == "NoExecute" {
				effect = model.GetTaintEffectEnum().NO_EXECUTE
			}
			value := v.Value
			req.Body.Spec.NodeTemplate.Taints = append(req.Body.Spec.NodeTemplate.Taints, model.Taint{
				Key:    v.Key,
				Value:  &value,
				Effect: effect,
			})
		}
	}

	if group.Tags != nil {
		req.Body.Spec.NodeTemplate.K8sTags = group.Tags

		for k, v := range group.Tags {
			key := k
			value := v
			req.Body.Spec.NodeTemplate.UserTags = append(req.Body.Spec.NodeTemplate.UserTags, model.UserTag{
				Key:   &key,
				Value: &value,
			})
		}
	}

	if len(req.Body.Spec.NodeTemplate.Taints) == 0 && oldNodePool.Spec.NodeTemplate.Taints != nil {
		req.Body.Spec.NodeTemplate.Taints = *oldNodePool.Spec.NodeTemplate.Taints
	}

	if len(req.Body.Spec.NodeTemplate.K8sTags) == 0 && oldNodePool.Spec.NodeTemplate.K8sTags != nil {
		req.Body.Spec.NodeTemplate.K8sTags = oldNodePool.Spec.NodeTemplate.K8sTags
	}

	if len(req.Body.Spec.NodeTemplate.UserTags) == 0 && oldNodePool.Spec.NodeTemplate.UserTags != nil {
		req.Body.Spec.NodeTemplate.UserTags = *oldNodePool.Spec.NodeTemplate.UserTags
	}

	return req
}

// GenerateCreateNodePoolRequest get cce nodepool request
func GenerateCreateNodePoolRequest(group *proto.NodeGroup,
	cluster *proto.Cluster) (*model.CreateNodePoolRequest, error) {
	var (
		initialNodeCount  int32 = 0
		clusterId               = cluster.SystemID
		podSecurityGroups []model.SecurityId
	)

	nodeTemplate, err := GenerateNodeSpec(group)
	if err != nil {
		return nil, err
	}

	if group.LaunchTemplate != nil {
		for _, v := range group.LaunchTemplate.SecurityGroupIDs {
			id := v
			podSecurityGroups = append(podSecurityGroups, model.SecurityId{Id: &id})
		}
	}

	return &model.CreateNodePoolRequest{
		ClusterId: clusterId,
		Body: &model.NodePool{
			Kind:       "NodePool",
			ApiVersion: "v3",
			Metadata: &model.NodePoolMetadata{
				Name: group.NodeGroupID,
			},
			Spec: &model.NodePoolSpec{
				InitialNodeCount:  &initialNodeCount,
				NodeTemplate:      nodeTemplate,
				PodSecurityGroups: &podSecurityGroups,
			},
		},
	}, nil
}

// GenerateNodeSpec get node spec
func GenerateNodeSpec(nodeGroup *proto.NodeGroup) (*model.NodeSpec, error) {
	if nodeGroup.LaunchTemplate == nil {
		return nil, fmt.Errorf("node group launch template is nil")
	}

	var (
		nodeBillingMode   int32 = 0
		maxPod            int32 = 64
		periodType              = "month"
		periodNum         int32 = 1
		az                      = "random" // 随机选择可用区
		subnetId                = ""
		isAutoRenew             = "false"
		isAutoPay               = "true"
		runtimeName             = model.GetRuntimeNameEnum().CONTAINERD
		metadataEncrypted       = "0"
		matchCount              = "1"
		cceManaged              = true
	)

	if nodeGroup.LaunchTemplate != nil {
		if nodeGroup.LaunchTemplate.InstanceChargeType == common.PREPAID && nodeGroup.LaunchTemplate.Charge != nil {
			nodeBillingMode = 1
			periodNum = int32(nodeGroup.LaunchTemplate.Charge.Period)
			if nodeGroup.LaunchTemplate.Charge.Period >= 12 {
				periodType = "year"
				periodNum = int32(nodeGroup.LaunchTemplate.Charge.Period / 12)
			}
			if nodeGroup.LaunchTemplate.Charge.RenewFlag == common.NOTIFYANDAUTORENEW {
				isAutoRenew = "true"
			}
		}
	}

	if nodeGroup.AutoScaling != nil {
		// 指定可用区
		if len(nodeGroup.AutoScaling.Zones) > 0 {
			az = nodeGroup.AutoScaling.Zones[0]
		}
		// 华为云只支持设置一个子网
		if len(nodeGroup.AutoScaling.SubnetIDs) > 0 {
			subnetId = nodeGroup.AutoScaling.SubnetIDs[0]
		}
	}

	if nodeGroup.LaunchTemplate.InstanceType == "" {
		return nil, fmt.Errorf("the node specifications cannot be empty")
	}

	if nodeGroup.LaunchTemplate.SystemDisk == nil {
		return nil, fmt.Errorf("the system disk information of a node cannot be empty")
	}

	if len(nodeGroup.LaunchTemplate.DataDisks) == 0 {
		return nil, fmt.Errorf("the data disk information of a node cannot be empty")
	}

	if nodeGroup.NodeTemplate != nil && nodeGroup.NodeTemplate.MaxPodsPerNode != 0 {
		maxPod = int32(nodeGroup.AutoScaling.MaxSize)
	}

	if nodeGroup.NodeTemplate != nil && nodeGroup.NodeTemplate.Runtime != nil {
		if nodeGroup.NodeTemplate.Runtime.ContainerRuntime == common.DockerContainerRuntime {
			runtimeName = model.GetRuntimeNameEnum().DOCKER
		}
	}

	diskSize, err := strconv.Atoi(nodeGroup.LaunchTemplate.SystemDisk.DiskSize)
	if err != nil {
		return nil, err
	}

	dataVolumes := make([]model.Volume, 0)
	storageSelectors := make([]model.StorageSelectors, 0)
	storageGroups := make([]model.StorageGroups, 0)
	for k, v := range nodeGroup.NodeTemplate.DataDisks {
		var size int
		size, err = strconv.Atoi(v.DiskSize)
		if err != nil {
			return nil, err
		}

		dataVolumes = append(dataVolumes, model.Volume{
			Volumetype: v.DiskType,
			Size:       int32(size),
		})

		selectorName := fmt.Sprintf("selector%d", k)
		storageSelectors = append(storageSelectors, model.StorageSelectors{
			Name:        selectorName,
			StorageType: "evs",
			MatchLabels: &model.StorageSelectorsMatchLabels{
				Size:              &v.DiskSize,
				VolumeType:        &v.DiskType,
				MetadataEncrypted: &metadataEncrypted,
				Count:             &matchCount,
			},
		})

		if k == 0 {
			storageGroups = append(storageGroups, model.StorageGroups{
				Name:          "vgpaas", // 当cceManaged=ture时，name必须为：vgpaas
				SelectorNames: []string{selectorName},
				CceManaged:    &cceManaged, // k8s及runtime所属存储空间。有且仅有一个group被设置为true，不填默认false
				VirtualSpaces: []model.VirtualSpace{
					{
						Name: "kubernetes",
						Size: "10%",
						LvmConfig: &model.LvmConfig{
							LvType: "linear",
						},
					},
					{
						Name: "runtime",
						Size: "90%",
						RuntimeConfig: &model.RuntimeConfig{
							LvType: "linear",
						},
					},
				},
			})
		} else {
			storageGroup := model.StorageGroups{
				Name:          fmt.Sprintf("group%d", k),
				SelectorNames: []string{selectorName},
				VirtualSpaces: []model.VirtualSpace{
					{
						Name: "user",
						Size: "100%",
						LvmConfig: &model.LvmConfig{
							LvType: "linear",
						},
					},
				},
			}
			if v.FileSystem != "" {
				storageGroup.VirtualSpaces[0].LvmConfig.Path = &v.FileSystem
			}
			storageGroups = append(storageGroups, storageGroup)
		}
	}

	password, err := Crypt(nodeGroup.LaunchTemplate.InitLoginPassword)
	if err != nil {
		return nil, err
	}

	return &model.NodeSpec{
		Flavor: nodeGroup.LaunchTemplate.InstanceType,
		Az:     az,
		Os:     &nodeGroup.NodeOS,
		Login: &model.Login{
			UserPassword: &model.UserPassword{
				Password: password, //username不填默认为root，password必须加盐并base64加密
			},
		},
		RootVolume: &model.Volume{
			Volumetype: nodeGroup.LaunchTemplate.SystemDisk.DiskType,
			Size:       int32(diskSize),
		},
		DataVolumes: dataVolumes,
		Storage: &model.Storage{
			StorageSelectors: storageSelectors,
			StorageGroups:    storageGroups,
		},
		BillingMode: &nodeBillingMode,
		Runtime: &model.Runtime{
			Name: &runtimeName,
		},
		InitializedConditions: &[]string{
			"NodeInitial", // 新增节点调度策略: 设置为不可调度
		},
		ExtendParam: &model.NodeExtendParam{
			MaxPods:             &maxPod,
			PeriodType:          &periodType,
			PeriodNum:           &periodNum,
			IsAutoRenew:         &isAutoRenew,
			IsAutoPay:           &isAutoPay,
			AlphaCcePreInstall:  &nodeGroup.NodeTemplate.PreStartUserScript,
			AlphaCcePostInstall: &nodeGroup.NodeTemplate.UserScript,
		},
		NodeNicSpec: &model.NodeNicSpec{
			PrimaryNic: &model.NicSpec{SubnetId: &subnetId},
		},
		HostnameConfig: &model.HostnameConfig{
			Type: model.GetHostnameConfigTypeEnum().PRIVATE_IP, // 节点名称默认与节点私有ip保持一致
		},
	}, nil
}
