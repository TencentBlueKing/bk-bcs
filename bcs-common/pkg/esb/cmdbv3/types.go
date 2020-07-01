/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmdbv3

import (
	"encoding/json"
)

// BasePage for paging query
type BasePage struct {
	Sort  string `json:"sort,omitempty" mapstructure:"sort"`
	Limit int    `json:"limit,omitempty" mapstructure:"limit"`
	Start int    `json:"start" mapstructure:"start"`
}

// CreatePod request for CreatePod
type CreatePod struct {
	Pod MapStr `json:"pod" mapstructure:"pod"`
}

// CreateOneDataResult the data struct definition in create one function result
type CreateOneDataResult struct {
	Created CreatedDataResult `json:"created"`
}

// CreatedOneOptionResult create One api http response return this result struct
type CreatedOneOptionResult struct {
	BaseResp `json:",inline"`
	Data     CreateOneDataResult `json:"data"`
}

// ExceptionResult exception info
type ExceptionResult struct {
	Message     string      `json:"message"`
	Code        int64       `json:"code"`
	Data        interface{} `json:"data"`
	OriginIndex int64       `json:"origin_index"`
}

// CreatedDataResult common created result definition
type CreatedDataResult struct {
	OriginIndex int64  `json:"origin_index"`
	ID          uint64 `json:"id"`
}

// RepeatedDataResult repeated data
type RepeatedDataResult struct {
	OriginIndex int64  `json:"origin_index"`
	Data        MapStr `json:"data"`
}

// CreateManyInfoResult create many function return this result struct
type CreateManyInfoResult struct {
	Created    []CreatedDataResult  `json:"created"`
	Repeated   []RepeatedDataResult `json:"repeated"`
	Exceptions []ExceptionResult    `json:"exception"`
}

// CreateManyDataResult the data struct definition in create many function result
type CreateManyDataResult struct {
	CreateManyInfoResult `json:",inline"`
}

// CreateManyPod request for CreateManyPod
type CreateManyPod struct {
	PodList []MapStr `json:"pod_list" mapstructure:"pod_list"`
}

// CreatedManyOptionResult create many api http response return this result struct
type CreatedManyOptionResult struct {
	BaseResp `json:",inline"`
	Data     CreateManyDataResult `json:"data"`
}

// UpdateOption common update options
type UpdateOption struct {
	Data      MapStr `json:"data" mapstructure:"data"`
	Condition MapStr `json:"condition" mapstructure:"condition"`
}

// UpdatePod parameter for UpdatePod
type UpdatePod struct {
	UpdateOption
}

// UpdatedCount created count struct
type UpdatedCount struct {
	Count uint64 `json:"updated_count"`
}

// UpdatedOptionResult common update result
type UpdatedOptionResult struct {
	BaseResp `json:",inline"`
	Data     UpdatedCount `json:"data" mapstructure:"data"`
}

// BaseResp common result struct
type BaseResp struct {
	Result bool   `json:"result" mapstructure:"result"`
	Code   int    `json:"bk_error_code" mapstructure:"bk_error_code"`
	ErrMsg string `json:"bk_error_msg" mapstructure:"bk_error_msg"`
}

// DeletePod parameter for DeletePod
type DeletePod struct {
	DeleteOption
}

// DeleteOption common delete condition options
type DeleteOption struct {
	Condition MapStr `json:"condition"`
}

// DeletedCount created count struct
type DeletedCount struct {
	Count uint64 `json:"deleted_count"`
}

// DeletedOptionResult delete  api http response return result struct
type DeletedOptionResult struct {
	BaseResp `json:",inline"`
	Data     DeletedCount `json:"data"`
}

// ListPods request for ListPods
type ListPods struct {
	BizID     int64    `json:"bk_biz_id"`
	SetIDs    []int64  `json:"bk_set_ids"`
	ModuleIDs []int64  `json:"bk_module_ids"`
	Fields    []string `json:"fields"`
	Page      BasePage `json:"page"`
}

// QueryResult common query result
type QueryResult struct {
	Count uint64            `json:"count"`
	Info  []json.RawMessage `json:"info"`
}

// ListPodsResult response for ListPod
type ListPodsResult struct {
	BaseResp `json:",inline"`
	Data     *QueryResult `json:"data"`
}
