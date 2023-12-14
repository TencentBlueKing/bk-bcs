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

package types

import "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/types"

// TmplBoundUnnamedAppDetail defines struct of template bound unnamed app detail.
type TmplBoundUnnamedAppDetail struct {
	AppID               uint32            `json:"app_id"`
	TemplateRevisionIDs types.Uint32Slice `json:"template_revision_ids"`
}

// TmplBoundNamedAppDetail defines struct of template bound named app detail.
type TmplBoundNamedAppDetail struct {
	AppID              uint32 `json:"app_id"`
	ReleaseID          uint32 `json:"release_id"`
	TemplateRevisionID uint32 `json:"template_revision_id"`
}

// TmplRevisionBoundNamedAppDetail defines struct of template release bound named app detail.
type TmplRevisionBoundNamedAppDetail struct {
	AppID     uint32 `json:"app_id"`
	ReleaseID uint32 `json:"release_id"`
}

// TmplSetBoundNamedAppDetail defines struct of template set bound named app detail.
type TmplSetBoundNamedAppDetail struct {
	AppID     uint32 `json:"app_id"`
	ReleaseID uint32 `json:"release_id"`
}
