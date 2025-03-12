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

// Package steps include all steps for federation manager
package steps

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// InstallFederationCallBackName step name for create cluster
	InstallFederationCallBackName = fedsteps.CallBackNames{
		Name:  "INSTALL_FEDERATION_CALLBACK",
		Alias: "install federation call back func",
	}
)

// NewInstallFederationCallBack new install federation callback step
func NewInstallFederationCallBack() *InstallFederationCallBack {
	return &InstallFederationCallBack{}
}

// InstallFederationCallBack step for create cluster
type InstallFederationCallBack struct{}

// GetName get step name
func (s InstallFederationCallBack) GetName() string {
	return InstallFederationCallBackName.Name
}

// Callback step callback function
func (s InstallFederationCallBack) Callback(isSuccess bool, t *types.Task) {
	blog.Infof("install federation callback, taskID: %s, isSuccess: %t", t.GetTaskID(), isSuccess)
	if isSuccess {
		// do nothing, all success relate operation should be execute in last step
		return
	}

	fedClusterId, ok := t.GetCommonParams(fedsteps.FedClusterIdKey)
	if !ok {
		err := fedsteps.ParamsNotFoundError(t.GetTaskID(), fedsteps.FedClusterIdKey)
		blog.Errorf("call back failed, error: %v", err)
		oldMessage := t.GetMessage()
		t.SetMessage(fmt.Sprintf(oldMessage+", callback failed, err: %v", err))
		return
	}

	// update federation cluster status to failure
	err := cluster.GetClusterClient().UpdateFederationClusterStatus(context.Background(), fedClusterId, cluster.ClusterStatusCreateFailure)
	if err != nil {
		blog.Errorf("update federation cluster status failed when callback, clusterID: %s, err: %v", fedClusterId, err)
		oldMessage := t.GetMessage()
		t.SetMessage(fmt.Sprintf(oldMessage+", update federation cluster status failed when callback, err: %v", err))
		return
	}

	fedCluster, err := store.GetStoreModel().GetFederationCluster(context.Background(), fedClusterId)
	if err != nil {
		blog.Errorf("get federation cluster failed when callback, clusterID: %s, err: %v", fedClusterId, err)
		oldMessage := t.GetMessage()
		t.SetMessage(fmt.Sprintf(oldMessage+", get federation cluster failed when callback, err: %v", err))
		return
	}

	creator, ok := t.GetCommonParams(fedsteps.CreatorKey)
	if !ok {
		// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
		err := fedsteps.ParamsNotFoundError(t.GetTaskID(), fedsteps.CreatorKey)
		blog.Errorf("call back failed, error: %v", err)
		oldMessage := t.GetMessage()
		t.SetMessage(fmt.Sprintf(oldMessage+", callback failed, err: %v", err))
		return
	}

	fedCluster.Status = store.CreateFailedStatus
	if err = store.GetStoreModel().UpdateFederationCluster(context.Background(), fedCluster, creator); err != nil {
		blog.Errorf("update federation cluster failed when callback, clusterID: %s, err: %v", fedClusterId, err)
		oldMessage := t.GetMessage()
		t.SetMessage(fmt.Sprintf(oldMessage+", update federation cluster failed when callback, err: %v", err))
		return
	}
}
