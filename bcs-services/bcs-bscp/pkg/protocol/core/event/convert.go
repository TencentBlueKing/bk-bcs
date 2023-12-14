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

// Package pbevent provides event core protocol struct and convert functions.
package pbevent

import "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"

// PbEventSpec convert event spec to pb event spec
func PbEventSpec(spec *table.EventSpec) *EventSpec { //nolint:revive
	if spec == nil {
		return nil
	}

	return &EventSpec{
		Resource:    string(spec.Resource),
		ResourceId:  spec.ResourceID,
		ResourceUid: spec.ResourceUid,
		OpType:      string(spec.OpType),
	}
}

// PbEventAttachment convert event attachment to pb event attachment
func PbEventAttachment(attach *table.EventAttachment) *EventAttachment { //nolint:revive
	if attach == nil {
		return nil
	}

	return &EventAttachment{
		BizId: attach.BizID,
		AppId: attach.AppID,
	}
}
