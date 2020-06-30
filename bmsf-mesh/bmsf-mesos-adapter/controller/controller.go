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

package controller

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/queue"
	"github.com/Tencent/bk-bcs/bmsf-mesh/bmsf-mesos-adapter/controller/appnode"
	"github.com/Tencent/bk-bcs/bmsf-mesh/bmsf-mesos-adapter/controller/appsvc"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, svcQ queue.Queue, nodeQ queue.Queue) error {
	if err := appsvc.Add(m, svcQ); err != nil {
		return err
	}
	if err := appnode.Add(m, nodeQ); err != nil {
		return err
	}
	return nil
}
