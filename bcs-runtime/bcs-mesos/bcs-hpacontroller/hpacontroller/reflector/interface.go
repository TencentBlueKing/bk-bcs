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

package reflector

import (
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	schedtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// Reflector watches a specified resource and causes all changes to be reflected in the given store.
type Reflector interface {
	//list all namespace autoscaler
	ListAutoscalers() ([]*commtypes.BcsAutoscaler, error)

	// store autoscaler in zk
	StoreAutoscaler(autoscaler *commtypes.BcsAutoscaler) error

	// update autoscaler in zk
	UpdateAutoscaler(autoscaler *commtypes.BcsAutoscaler) error

	//fetch deployment info, if deployment status is not Running, then can't autoscale this deployment
	FetchDeploymentInfo(namespace, name string) (*schedtypes.Deployment, error)

	//fetch application info, if application status is not Running or Abnormal, then can't autoscale this application
	FetchApplicationInfo(namespace, name string) (*schedtypes.Application, error)

	//list selectorRef deployment taskgroup
	ListTaskgroupRefDeployment(namespace, name string) ([]*schedtypes.TaskGroup, error)

	//list selectorRef application taskgroup
	ListTaskgroupRefApplication(namespace, name string) ([]*schedtypes.TaskGroup, error)
}
