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

package backend

import (
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// GetClusterResources get cluster resources
func (b *backend) GetClusterResources() (*commtypes.BcsClusterResource, error) {
	return b.sched.GetClusterResource()
}

// GetClusterEndpoints get cluster endpoints
func (b *backend) GetClusterEndpoints() *commtypes.ClusterEndpoints {
	endpoints := new(commtypes.ClusterEndpoints)

	for _, sched := range b.sched.Schedulers {
		endpoints.MesosSchedulers = append(endpoints.MesosSchedulers, *sched)
	}
	for _, master := range b.sched.Memsoses {
		endpoints.MesosMasters = append(endpoints.MesosMasters, *master)
	}

	return endpoints
}

// GetCurrentOffers get current offers of cluster
func (b *backend) GetCurrentOffers() ([]*types.OfferWithDelta) {
	return b.sched.GetCurrentOffers()
}
