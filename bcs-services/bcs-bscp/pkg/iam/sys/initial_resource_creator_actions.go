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

package sys

import "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"

// GenerateResourceCreatorActions generate all the resource creator actions registered to IAM.
func GenerateResourceCreatorActions() client.ResourceCreatorActions {
	return client.ResourceCreatorActions{
		Config: []client.ResourceCreatorAction{
			{
				ResourceID: Application,
				Actions: []client.CreatorRelatedAction{
					{
						ID:         AppView,
						IsRequired: false,
					},
					{
						ID:         AppEdit,
						IsRequired: false,
					},

					{
						ID:         AppDelete,
						IsRequired: false,
					},

					{
						ID:         ReleaseGenerate,
						IsRequired: false,
					},
					{
						ID:         ReleasePublish,
						IsRequired: false,
					},
				},
				SubResourceTypes: nil,
			},
		},
	}
}
