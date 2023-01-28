/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package shared

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/action"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
)

// SharedNamespaceAction namespace action for shared cluster
type SharedNamespaceAction struct {
	model store.ProjectModel
}

// NewSharedNamespaceAction new namespace action for shared cluster
func NewSharedNamespaceAction(model store.ProjectModel) action.NamespaceAction {
	return &SharedNamespaceAction{
		model: model,
	}
}
