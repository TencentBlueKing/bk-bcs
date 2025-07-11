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

package cloud

import (
	"context"
	"sort"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

var (
	cloudEnable = []string{"true", "false"}
)

// appendConfigNetworkInfoToCloud append network info to cloud
func appendConfigNetworkInfoToCloud(cloud *cmproto.Cloud, config *cmproto.CloudNetworkTemplateConfig) {
	cloudNetworkInfo := cloud.NetworkInfo
	cloudNetworkInfo.CidrSteps = append(cloudNetworkInfo.CidrSteps, config.CidrSteps...)
	cloudNetworkInfo.PerNodePodNum = append(cloudNetworkInfo.PerNodePodNum, config.PerNodePodNum...)
	cloudNetworkInfo.ServiceSteps = append(cloudNetworkInfo.ServiceSteps, config.ServiceSteps...)
	cloudNetworkInfo.UnderlaySteps = append(cloudNetworkInfo.UnderlaySteps, config.UnderlaySteps...)
	cloudNetworkInfo.UnderlayAutoSteps = append(cloudNetworkInfo.UnderlayAutoSteps, config.UnderlayAutoSteps...)
}

// getCloudTemplateConfigInfos get cloud template config infos
func getCloudTemplateConfigInfos(ctx context.Context, model store.ClusterManagerModel, businessID, cloudID string) ([]*cmproto.TemplateConfigInfo, error) {
	// if businessID is empty, no need to get template config
	if businessID == "" {
		return nil, nil
	}

	templateConfigs, err := actions.GetTemplateConfigInfosByBusinessID(ctx, model, businessID,
		cloudID, common.CloudConfigType, nil)
	if err != nil {
		return nil, err
	}

	return templateConfigs, nil
}

// dedupeAndSortNetworkInfo dedupe and sort network info
func dedupeAndSortNetworkInfo(networkInfo *cmproto.CloudNetworkInfo) {
	if networkInfo == nil {
		return
	}

	networkInfo.CidrSteps = dedupeAndSortEnvCidrSteps(networkInfo.CidrSteps)

	dedupeAndSortUint32 := func(slice []uint32) []uint32 {
		if len(slice) == 0 {
			return nil
		}
		exists := make(map[uint32]struct{})
		for _, v := range slice {
			exists[v] = struct{}{}
		}
		result := make([]uint32, 0, len(exists))
		for v := range exists {
			result = append(result, v)
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i] < result[j]
		})
		return result
	}

	networkInfo.PerNodePodNum = dedupeAndSortUint32(networkInfo.PerNodePodNum)
	networkInfo.ServiceSteps = dedupeAndSortUint32(networkInfo.ServiceSteps)
	networkInfo.UnderlaySteps = dedupeAndSortUint32(networkInfo.UnderlaySteps)
	networkInfo.UnderlayAutoSteps = dedupeAndSortUint32(networkInfo.UnderlayAutoSteps)

	return
}

// dedupeAndSortEnvCidrSteps dedupe and sort env cidr steps
func dedupeAndSortEnvCidrSteps(steps []*cmproto.EnvCidrStep) []*cmproto.EnvCidrStep {
	if len(steps) == 0 {
		return nil
	}

	exists := make(map[string]struct{})
	var uniqueSteps []*cmproto.EnvCidrStep

	for _, step := range steps {
		if step == nil {
			continue
		}
		key := step.Env + string(step.Step)
		if _, exist := exists[key]; !exist {
			exists[key] = struct{}{}
			uniqueSteps = append(uniqueSteps, step)
		}
	}

	sort.Slice(uniqueSteps, func(i, j int) bool {
		if uniqueSteps[i].Env != uniqueSteps[j].Env {
			return uniqueSteps[i].Env < uniqueSteps[j].Env
		}
		return uniqueSteps[i].Step < uniqueSteps[j].Step
	})

	return uniqueSteps
}
