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

import "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"

// GenerateStaticInstanceSelections return need register instance selection.
func GenerateStaticInstanceSelections() []client.InstanceSelection {
	return []client.InstanceSelection{
		{
			ID:     BusinessSelection,
			Name:   "业务列表",
			NameEn: "Business List",
			ResourceTypeChain: []client.ResourceChain{
				{
					SystemID: SystemIDCMDB,
					ID:       Business,
				},
			},
		},
		{
			ID:     ApplicationSelection,
			Name:   "服务列表",
			NameEn: "Application List",
			ResourceTypeChain: []client.ResourceChain{
				{
					SystemID: SystemIDCMDB,
					ID:       Business,
				},
				{
					SystemID: SystemIDBSCP,
					ID:       Application,
				},
			},
		},
	}
}
