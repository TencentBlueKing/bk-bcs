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

import "bscp.io/pkg/iam/client"

// GenerateCommonActions generate all the common actions registered to IAM.
func GenerateCommonActions() []client.CommonAction {
	CommonActions := make([]client.CommonAction, 0)

	CommonActions = append(CommonActions, genBizCommonActions()...)

	return CommonActions

}

// genBizCommonActions 推荐权限，业务查看、业务运维
func genBizCommonActions() []client.CommonAction {
	return []client.CommonAction{
		{
			Name:        "业务查看",
			EnglishName: "business view",
			Actions: []client.ActionWithID{
				{
					BusinessViewResource,
				},
				{
					AppView,
				},
				{
					CredentialView,
				},
			},
		},
		{
			Name:        "业务运维",
			EnglishName: "business ops",
			Actions: []client.ActionWithID{
				{
					BusinessViewResource,
				},
				{
					AppCreate,
				},
				{
					AppView,
				},
				{
					AppEdit,
				},
				{
					AppDelete,
				},
				{
					ReleaseGenerate,
				},
				{
					ReleasePublish,
				},
				{
					CredentialView,
				},
				{
					CredentialManage,
				},
			},
		},
	}

}
