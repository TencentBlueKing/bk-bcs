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

// Package pbcs provides cache service core protocol struct and convert functions.
package pbcs

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbevent "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/event"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// PbAppMetaMap convert app meta map.
func PbAppMetaMap(m map[ /*appID*/ uint32]*types.AppCacheMeta) map[ /*appID*/ uint32]*AppMeta {

	meta := make(map[uint32]*AppMeta)

	if len(m) == 0 {
		return meta
	}

	for key, val := range m {
		meta[key] = &AppMeta{
			Name: val.Name,
			Cft:  string(val.ConfigType),
		}
	}

	return meta
}

// PbEventMeta convert event meta to pb event meta
func PbEventMeta(events []*types.EventMeta) []*EventMeta {

	if len(events) == 0 {
		return make([]*EventMeta, 0)
	}

	metas := make([]*EventMeta, len(events))
	for idx := range events {
		metas[idx] = &EventMeta{
			Id:         events[idx].ID,
			Spec:       pbevent.PbEventSpec(events[idx].Spec),
			Attachment: pbevent.PbEventAttachment(events[idx].Attachment),
		}
	}

	return metas
}

// EventMeta convert pb event meta to type event meta.
func (m *EventMeta) EventMeta() *types.EventMeta {
	if m == nil {
		return nil
	}

	return &types.EventMeta{
		ID: m.Id,
		Spec: &table.EventSpec{
			Resource:    table.EventResource(m.Spec.Resource),
			ResourceID:  m.Spec.ResourceId,
			ResourceUid: m.Spec.ResourceUid,
			OpType:      table.EventType(m.Spec.OpType),
		},
		Attachment: &table.EventAttachment{
			BizID: m.Attachment.BizId,
			AppID: m.Attachment.AppId,
		},
	}
}
