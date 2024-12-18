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

// Package dynamic xxx
package dynamic

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

// PushCreateResourcesToQueue push create resource to queue
func PushCreateResourcesToQueue(data operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}

	// queueFlag true
	go func(data operator.M, featTags []string) {
		err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeUpdate)
		if err != nil {
			blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "putNamespaceResources", err)
		}
	}(data, nsFeatTags)
}

// PushDeleteResourcesToQueue push delete resources to queue
func PushDeleteResourcesToQueue(mList []operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}
	// queueFlag true
	go func(mList []operator.M, featTags []string) {
		for _, data := range mList {
			err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
			if err != nil {
				blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteNamespaceResources", err)
			}
		}
	}(mList, nsFeatTags)
}

// PushDeleteBatchResourceToQueue push delete batch resource to queue
func PushDeleteBatchResourceToQueue(mList []operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}
	// queueFlag true
	go func(mList []operator.M, featTags []string) {
		for _, data := range mList {
			err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
			if err != nil {
				blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteBatchNamespaceResource", err)
			}
		}
	}(mList, nsListFeatTags)
}

// PushCreateClusterToQueue push create cluster to queue
func PushCreateClusterToQueue(data operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}
	// queueFlag true
	// 入参nsFeatTags和实参csFeatTags不一样(featTags unused),源码如此,暂时保留
	// nolint
	go func(data operator.M, featTags []string) {
		err := publishDynamicResourceToQueue(data, csFeatTags, msgqueue.EventTypeUpdate)
		if err != nil {
			blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "putClusterResources", err)
		}
	}(data, nsFeatTags)
}

// PushDeleteClusterToQueue push delete cluster to queue
func PushDeleteClusterToQueue(mList []operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}
	// queueFlag true
	go func(mList []operator.M, featTags []string) {
		for _, data := range mList {
			err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
			if err != nil {
				blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteClusterResources", err)
			}
		}
	}(mList, csFeatTags)
}

// PushDeleteBatchClusterToQueue push delete batch cluster to queue
func PushDeleteBatchClusterToQueue(mList []operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}
	// queueFlag true
	go func(mList []operator.M, featTags []string) {
		for _, data := range mList {
			err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
			if err != nil {
				blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteClusterNamespaceResource", err)
			}
		}
	}(mList, csListFeatTags)
}
