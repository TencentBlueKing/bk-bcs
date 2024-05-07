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
	"strconv"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt" // use init func
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	v1 "k8s.io/api/core/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

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
	oldNodePool *model.NodePool) *model.UpdateNodePoolRequest {
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
	cluster *proto.Cluster) (*CreateNodePoolRequest, error) {
	var (
		clusterId               = cluster.SystemID
		subnetId                = ""
		securityGroups          = make([]string, 0)
		az                      = "random" // 随机选择可用区
		sshKey                  = ""
		dataVolumes             = make([]*Volume, 0)
		taints                  = make([]v1.Taint, 0)
		period           uint32 = 0
		renewFlag               = ""
		containerRuntime        = ""
		password                = ""
	)

	if group.AutoScaling != nil {
		// 指定可用区
		if len(group.AutoScaling.Zones) > 0 {
			az = group.AutoScaling.Zones[0]
		}
		// 华为云只支持设置一个子网
		if len(group.AutoScaling.SubnetIDs) > 0 {
			subnetId = group.AutoScaling.SubnetIDs[0]
		}
	}

	for _, v := range group.LaunchTemplate.SecurityGroupIDs {
		securityGroups = append(securityGroups, v)
	}

	diskSize, err := strconv.Atoi(group.LaunchTemplate.SystemDisk.DiskSize)
	if err != nil {
		return nil, err
	}

	for _, v := range group.NodeTemplate.DataDisks {
		size, err := strconv.Atoi(v.DiskSize)
		if err != nil {
			return nil, err
		}
		dataVolumes = append(dataVolumes, &Volume{
			Size:       int32(size),
			VolumeType: v.DiskType,
		})
	}

	for _, v := range group.NodeTemplate.Taints {
		taints = append(taints, v1.Taint{
			Key:    v.Key,
			Value:  v.Value,
			Effect: v1.TaintEffect(v.Effect),
		})
	}

	if group.LaunchTemplate.Charge != nil {
		period = group.LaunchTemplate.Charge.Period
		renewFlag = group.LaunchTemplate.Charge.RenewFlag
	}

	if group.NodeTemplate.Runtime != nil {
		containerRuntime = group.NodeTemplate.Runtime.ContainerRuntime
	}

	if group.LaunchTemplate.KeyPair != nil && group.LaunchTemplate.KeyPair.KeyID != "" {
		sshKey = group.LaunchTemplate.KeyPair.KeyID
	} else if group.LaunchTemplate.InitLoginPassword != "" {
		password, err = Crypt(group.LaunchTemplate.InitLoginPassword)
		if err != nil {
			return nil, err
		}
	}

	return &CreateNodePoolRequest{
		ClusterId: clusterId,
		Name:      group.NodeGroupID,
		Spec: CreateNodePoolSpec{
			Template: CreateNodePoolTemplate{
				Flavor: group.LaunchTemplate.InstanceType,
				Az:     az,
				Os:     group.NodeOS,
				Login: Login{
					SshKey: sshKey,
					Passwd: password,
				},
				RootVolume: &Volume{
					Size:       int32(diskSize),
					VolumeType: group.LaunchTemplate.SystemDisk.DiskType,
				},
				DataVolumes: dataVolumes,
				Charge: ChargePrepaid{
					ChargeType: group.LaunchTemplate.InstanceChargeType,
					Period:     period,
					RenewFlag:  renewFlag,
				},
				Taints:           taints,
				Labels:           group.Tags,
				ContainerRuntime: containerRuntime,
				MaxPod:           int32(group.NodeTemplate.MaxPodsPerNode),
				PreScript:        group.NodeTemplate.PreStartUserScript,
				PostScript:       group.NodeTemplate.UserScript,
			},
			SecurityGroups: securityGroups,
			SubnetId:       subnetId,
		},
	}, nil
}
