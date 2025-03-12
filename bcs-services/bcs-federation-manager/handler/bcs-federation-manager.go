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

// Package handler xxx
package handler

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/thirdparty"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/user"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
)

var _ federationmgr.FederationManagerHandler = &FederationManager{}

// FederationManagerOptions defines federation manager options
type FederationManagerOptions struct {
	BcsGateway        *GatewayConfig
	Store             store.FederationMangerModel
	ClusterManagerCli cluster.Client
	HelmManagerCli    helm.Client
	UserManagerCli    user.Client
	ProjectManagerCli project.Client
	ThirdManagerCli   thirdparty.Client
	TaskManager       *task.TaskManager
}

// GatewayConfig bcs gateway config
type GatewayConfig struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

// FederationManager defines federation manager handler
type FederationManager struct {
	bcsGateWay  *GatewayConfig
	store       store.FederationMangerModel
	clusterCli  cluster.Client
	helmCli     helm.Client
	userCli     user.Client
	projectCli  project.Client
	thirdCli    thirdparty.Client
	taskmanager *task.TaskManager
}

// NewFederationManager create federationmanager handler with store and k8s client
func NewFederationManager(o *FederationManagerOptions) *FederationManager {
	return &FederationManager{
		bcsGateWay:  o.BcsGateway,
		store:       o.Store,
		clusterCli:  o.ClusterManagerCli,
		helmCli:     o.HelmManagerCli,
		userCli:     o.UserManagerCli,
		projectCli:  o.ProjectManagerCli,
		thirdCli:    o.ThirdManagerCli,
		taskmanager: o.TaskManager,
	}
}

// IntToUint32Ptr convert int to *uint32
func IntToUint32Ptr(num int) *uint32 {
	uint32Num := uint32(num)
	return &uint32Num
}
