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

// Package handler 提供mesh manager的handler实现
package handler

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

var _ meshmanager.MeshManagerHandler = &MeshManager{}

// MeshManager provides a manager server for mesh resources
type MeshManager struct {
	model store.MeshManagerModel
	opt   *MeshManagerOptions
}

// MeshManagerOptions mesh manager options
type MeshManagerOptions struct {
	IstioConfig *options.IstioConfig
}

// NewMeshManager return a new MeshManager instance
func NewMeshManager(model store.MeshManagerModel, opt *MeshManagerOptions) *MeshManager {
	return &MeshManager{
		model: model,
		opt:   opt,
	}
}
