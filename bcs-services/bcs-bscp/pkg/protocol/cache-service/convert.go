/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package pbcs

import (
	"bscp.io/pkg/dal/table"
	pbapp "bscp.io/pkg/protocol/core/app"
	pbevent "bscp.io/pkg/protocol/core/event"
	pbstrategy "bscp.io/pkg/protocol/core/strategy"
	"bscp.io/pkg/types"
)

// PbAppMetaMap convert app meta map.
func PbAppMetaMap(m map[ /*appID*/ uint32]*types.AppCacheMeta) map[ /*appID*/ uint32]*AppMeta {

	meta := make(map[uint32]*AppMeta)

	if len(m) == 0 {
		return meta
	}

	for key, val := range m {
		meta[key] = &AppMeta{
			Cft:    string(val.ConfigType),
			Mod:    string(val.Mode),
			Reload: pbapp.PbReload(val.Reload),
		}
	}

	return meta
}

// PbAppCRIMetas convert app current released instance meta.
func PbAppCRIMetas(m []*types.AppCRIMeta) []*AppCRIMeta {
	if len(m) == 0 {
		return make([]*AppCRIMeta, 0)
	}

	meta := make([]*AppCRIMeta, len(m))

	for idx := range m {
		meta[idx] = &AppCRIMeta{
			Uid:       m[idx].Uid,
			ReleaseId: m[idx].ReleaseID,
		}
	}

	return meta
}

// PbPublishedStrategies convert published strategies.
func PbPublishedStrategies(ss []*types.PublishedStrategyCache) ([]*PublishedStrategy, error) {

	if len(ss) == 0 {
		return make([]*PublishedStrategy, 0), nil
	}

	ps := make([]*PublishedStrategy, len(ss))
	for idx := range ss {
		scope, err := pbstrategy.PbScopeSelector(ss[idx].Scope)
		if err != nil {
			return nil, err
		}

		ps[idx] = &PublishedStrategy{
			StrategyId: ss[idx].StrategyID,
			ReleaseId:  ss[idx].ReleaseID,
			AsDefault:  ss[idx].AsDefault,
			Scope:      scope,
			Mode:       string(ss[idx].Mode),
			Namespace:  ss[idx].Namespace,
		}
	}

	return ps, nil
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
