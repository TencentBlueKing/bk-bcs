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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// RegisterSubclusterCallBackName step name for create cluster
	RegisterSubclusterCallBackName = fedsteps.CallBackNames{
		Name:  "REGISTER_SUBCLUSTER_CALLBACK",
		Alias: "register subcluster call back func",
	}
)

// NewRegisterSubclusterCallBack new install federation callback step
func NewRegisterSubclusterCallBack() *RegisterSubclusterCallBack {
	return &RegisterSubclusterCallBack{}
}

// RegisterSubclusterCallBack step for create cluster
type RegisterSubclusterCallBack struct{}

// GetName get step name
func (s RegisterSubclusterCallBack) GetName() string {
	return RegisterSubclusterCallBackName.Name
}

// Callback step callback function
func (s RegisterSubclusterCallBack) Callback(isSuccess bool, t *types.Task) {
	blog.Infof("register subcluster callback, taskID: %s, isSuccess: %v", t.GetTaskID(), isSuccess)
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

	subClusterId, ok := t.GetCommonParams(fedsteps.SubClusterIdKey)
	if !ok {
		err := fedsteps.ParamsNotFoundError(t.GetTaskID(), fedsteps.SubClusterIdKey)
		blog.Errorf("call back failed, error: %v", err)
		oldMessage := t.GetMessage()
		t.SetMessage(fmt.Sprintf(oldMessage+", callback failed, err: %v", err))
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

	subCluster, err := store.GetStoreModel().GetSubCluster(context.Background(), fedClusterId, subClusterId)
	if err != nil {
		blog.Errorf("call back failed when get subcluster, error: %v", err)
		oldMessage := t.GetMessage()
		t.SetMessage(fmt.Sprintf(oldMessage+", callback failed, err: %v", err))
		return
	}

	// set status to failed
	subCluster.Status = store.CreateFailedStatus
	if err := store.GetStoreModel().UpdateSubCluster(context.Background(), subCluster, creator); err != nil {
		blog.Errorf("call back failed when update subcluster, error: %v", err)
		oldMessage := t.GetMessage()
		t.SetMessage(fmt.Sprintf(oldMessage+", callback failed, err: %v", err))
		return
	}

}
