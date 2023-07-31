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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
)

var (
	zones = map[string][]string{
		//非洲-约翰内斯堡
		"af-south-1": {"af-south-1a", "af-south-1b"},
		//华北-北京四
		"cn-north-4": {"cn-north-4a", "cn-north-4b", "cn-north-4c", "cn-north-4g"},
		//华北-北京一
		"cn-north-1": {"cn-north-1a", "cn-north-1b", "cn-north-1c"},
		//华北-乌兰察布一
		"cn-north-9": {"cn-north-9a", "cn-north-9b"},
		//华东-上海二
		"cn-east-2": {"cn-east-2a", "cn-east-2b", "cn-east-2c", "cn-east-2d"},
		//华东-上海一
		"cn-east-3": {"cn-east-3a", "cn-east-3b", "cn-east-3c"},
		//华南-广州
		"cn-south-1": {"cn-south-1a", "cn-south-2b", "cn-south-1c", "cn-south-1e", "cn-south-1f"},
		//华南-广州-友好用户环境
		"cn-south-4": {"cn-south-4a", "cn-south-4b", "cn-south-4c"},
		//华南-深圳
		"cn-south-2": {"cn-south-2a"},
		//拉美-墨西哥城二
		"la-north-2": {"la-north-2a", "la-north-2c"},
		//拉美-圣地亚哥
		"la-south-2": {"la-south-2a"},
		//欧洲-巴黎
		"eu-west-0": {"eu-west-0a", "eu-west-0b", "eu-west-0c"},
		//欧洲-都柏林
		"eu-west-101": {"eu-west-101a", "eu-west-101b"},
		//土耳其-伊斯坦布尔
		"tr-west-1": {"tr-west-1a", "tr-west-1b", "tr-west-1c"},
		//西南-贵阳一
		"cn-southwest-2": {"cn-southwest-2a", "cn-southwest-2b", "cn-southwest-2c", "cn-southwest-2d",
			"cn-southwest-2e", "cn-southwest-2f"},
		//亚太-曼谷
		"ap-southeast-2": {"ap-southeast-2a", "ap-southeast-2b", "ap-southeast-2c"},
		//亚太-新加坡
		"ap-southeast-3": {"ap-southeast-3a", "ap-southeast-3b", "ap-southeast-3c"},
		//亚太-雅加达
		"ap-southeast-4": {"ap-southeast-4a", "ap-southeast-4b", "ap-southeast-4c"},
		//中国-香港
		"ap-southeast-1": {"ap-southeast-1a", "ap-southeast-1b"},
	}
)

// GetClusterKubeConfig get cce cluster kebeconfig
func GetClusterKubeConfig(client *CceClient, clusterId string) (string, error) {
	req := model.CreateKubernetesClusterCertRequest{
		ClusterId: clusterId, // 集群ID，可在CCE管理控制台中查看
		Body: &model.CertDuration{
			// 集群证书有效时间,单位为天,最小值为1,最大值为10950(30*365,1年固定计365天，忽略闰年影响);若填-1则为最大值30年
			Duration: int32(-1),
		},
	}

	rsp, err := client.CreateKubernetesClusterCert(&req)
	if err != nil {
		return "", err
	}

	bt, err := json.Marshal(rsp)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bt), nil
}

// GetRegionZone region convert to zone
func GetRegionZone(region string) string {
	return zones[region][0]
}

// GenerateCreateNodePoolRequest get cce nodepool request
func GenerateCreateNodePoolRequest(group *proto.NodeGroup,
	cluster *proto.Cluster) (*model.CreateNodePoolRequest, error) {
	var (
		initialNodeCount int32 = 0
		clusterId              = cluster.SystemID
	)

	nodeTemplate, err := GenerateNodeSpec(group)
	if err != nil {
		return nil, err
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
				InitialNodeCount: &initialNodeCount,
				NodeTemplate:     nodeTemplate,
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
		nodeBillingMode int32 = 0
		maxPod          int32 = 110
		az                    = GetRegionZone(nodeGroup.Region)
	)

	if nodeGroup.LaunchTemplate.InstanceType == "" {
		return nil, fmt.Errorf("the node specifications cannot be empty")
	}

	if az == "" {
		return nil, fmt.Errorf("the availability zone cannot be found in [%s]", nodeGroup.Region)
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

	diskSize, err := strconv.Atoi(nodeGroup.LaunchTemplate.SystemDisk.DiskSize)
	if err != nil {
		return nil, err
	}

	dataVolumes := make([]model.Volume, 0)
	for _, v := range nodeGroup.LaunchTemplate.DataDisks {
		var size int
		size, err = strconv.Atoi(v.DiskSize)
		if err != nil {
			return nil, err
		}

		dataVolumes = append(dataVolumes, model.Volume{
			Volumetype: v.DiskType,
			Size:       int32(size),
		})
	}

	password, err := Crypt(nodeGroup.LaunchTemplate.InitLoginPassword)
	if err != nil {
		return nil, err
	}

	return &model.NodeSpec{
		Flavor: nodeGroup.LaunchTemplate.InstanceType,
		Az:     GetRegionZone(nodeGroup.Region),
		Os:     &nodeGroup.NodeOS,
		Login: &model.Login{
			UserPassword: &model.UserPassword{
				//username不填默认为root，password必须加盐并base64加密
				Password: password,
			},
		},
		RootVolume: &model.Volume{
			Volumetype: nodeGroup.LaunchTemplate.SystemDisk.DiskType,
			Size:       int32(diskSize),
		},
		DataVolumes: dataVolumes,
		BillingMode: &nodeBillingMode,
		ExtendParam: &model.NodeExtendParam{
			MaxPods: &maxPod,
		},
	}, nil
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
	group.NodeGroupID = strings.ToLower(group.NodeGroupID)

	req := &model.UpdateNodePoolRequest{
		NodepoolId: group.CloudNodeGroupID,
		ClusterId:  clusterID,
		Body: &model.NodePoolUpdate{
			Metadata: &model.NodePoolMetadataUpdate{
				Name: group.NodeGroupID,
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

		if group.Tags != nil {
			req.Body.Spec.NodeTemplate.K8sTags = group.Tags
		}

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
