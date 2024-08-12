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

package action

import (
	"k8s.io/client-go/util/workqueue"
)

const (
	// SyncDataActionAdd is add action on SyncData.
	SyncDataActionAdd = "Add"

	// SyncDataActionDelete is delete action on SyncData.
	SyncDataActionDelete = "Delete"

	// SyncDataActionUpdate is update action on SyncData.
	SyncDataActionUpdate = "Update"
)

// SyncData is metadata would be synced to storage.
type SyncData struct {
	// Kind is resource kind.
	Kind string
	// Namespace is k8s resource namespace.
	Namespace string
	// Name is resource name.
	Name string
	// Action is SyncDataAction Add/Delete/Update.
	Action string

	// Data is resource metadata.
	Data interface{}

	// OwnerUID is resource owner id.
	OwnerUID string

	// RequeueQ queue for requeue object
	RequeueQ workqueue.RateLimitingInterface
}
