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

// Package pod xxx
package pod

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// PodsAction pod action
type PodsAction interface {
	GetPodContainers(ctx context.Context, projectId, clusterId string) (*types.SampleResponse, error)
}

// Action action for pod
type Action struct {
	model storage.Storage
}

// NewPodAction new pod action
func NewPodAction(model storage.Storage) PodsAction {
	return &Action{
		model: model,
	}
}

// GetPodContainers get business info
func (a *Action) GetPodContainers(ctx context.Context, projectId, clusterId string) (*types.SampleResponse, error) {

	audit, err := a.model.GetAudit(ctx, projectId, clusterId)
	if err != nil {
		return nil, err
	}

	sr := &types.SampleResponse{
		Id:                  audit.ID.Hex(),
		CollectorConfigName: audit.CollectorConfigName,
	}
	return sr, nil
}
