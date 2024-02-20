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

// Package mongo xxx
package mongo

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

type server struct {
	*ModelCluster
	*ModelNamespace
	*ModelProject
	*ModelWorkload
	*ModelPublic
	*ModelPodAutoscaler
	*ModelPowerTrading
	*ModelCloudNative
	*ModelOperationData
	*ModelWorkloadRequest
	*ModelWorkloadOriginRequest
}

// NewServer new db server
func NewServer(db drivers.DB, bkbaseConf *types.BkbaseConfig) store.Server {
	return &server{
		ModelCluster:               NewModelCluster(db),
		ModelNamespace:             NewModelNamespace(db),
		ModelWorkload:              NewModelWorkload(db),
		ModelProject:               NewModelProject(db),
		ModelPublic:                NewModelPublic(db),
		ModelPodAutoscaler:         NewModelPodAutoscaler(db),
		ModelPowerTrading:          NewModelPowerTrading(db, bkbaseConf),
		ModelCloudNative:           NewModelCloudNative(db, bkbaseConf),
		ModelOperationData:         NewModelOperationData(db, bkbaseConf),
		ModelWorkloadRequest:       NewModelWorkloadRequest(db),
		ModelWorkloadOriginRequest: NewModelWorkloadOriginRequest(db),
	}
}
