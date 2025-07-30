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

// Package project xxx
package project

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

// PushCreateProjectInfoToQueue push create project info to queue
func PushCreateProjectInfoToQueue(data operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}

	// queueFlag true
	go func(data operator.M, featTags []string) {
		err := publishProjectInfoToQueue(data, featTags, msgqueue.EventTypeAdd)
		if err != nil {
			blog.Errorf("func[%s] call publishProjectInfoToQueue failed: err[%v]", "publishProjectInfoToQueue", err)
		}
	}(data, proFeatTags)
}
