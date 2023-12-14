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

package lcache

import (
	"fmt"

	clientset "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/client-set"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// NewLocalCache initial the cache instance.
func NewLocalCache(cs *clientset.ClientSet) (*Cache, error) {

	mc := initMetric()

	return &Cache{
		App:           newApp(mc, cs),
		ReleasedCI:    newReleasedCI(mc, cs),
		ReleasedKv:    newReleasedKv(mc, cs),
		ReleasedGroup: newReleasedGroup(mc, cs),
		ReleasedHook:  newReleasedHook(mc, cs),
		Credential:    newCredential(mc, cs),
		Auth:          newAuth(mc, cs.Authorizer()),
	}, nil
}

// Cache defines a cache instance.
type Cache struct {
	App           *App
	ReleasedCI    *ReleasedCI
	ReleasedKv    *ReleasedKv
	ReleasedGroup *ReleasedGroup
	Credential    *Credential
	ReleasedHook  *ReleasedHook
	Auth          *Auth
}

// Purge is used to clean the resource's cache with events.
func (c *Cache) Purge(kt *kit.Kit, es []*types.EventMeta) {
	if len(es) == 0 {
		return
	}

	logs.Infof("received events, start to purge local cache, rid: %s", kt.Rid)
	for _, one := range es {

		logs.Infof("handle event: %s, rid: %s", formatEvent(one), kt.Rid)

		// no matter what kind of event type, remove the resource from
		// local cache directly to force update the local cache immediately.
		switch one.Spec.Resource {
		case table.Publish:
			switch one.Spec.OpType {
			case table.InsertOp, table.DeleteOp:
				c.ReleasedGroup.client.Purge()
			default:
				logs.V(1).Infof("skip publish strategy event op, %s, rid: %s", formatEvent(one), kt.Rid)
				continue
			}

		case table.Application:
			switch one.Spec.OpType {
			case table.InsertOp:
				// ignore app insert event, refresh app meta cache when query app meta.
			case table.UpdateOp, table.DeleteOp:
				c.App.delete(one.Attachment.AppID)
			default:
				logs.V(1).Infof("skip app event op, %s, rid: %s", formatEvent(one), kt.Rid)
				continue
			}

		case table.EventResource(table.CredentialTable):
			switch one.Spec.OpType {
			case table.UpdateOp, table.DeleteOp:
				c.Auth.client.Purge()
			default:
				logs.V(1).Infof("skip credential event op, %s, rid: %s", formatEvent(one), kt.Rid)
				continue
			}

		default:
			logs.V(1).Infof("skip resource event, %s, rid: %s", formatEvent(one), kt.Rid)
			continue
		}
	}
}

func formatEvent(meta *types.EventMeta) string {
	return fmt.Sprintf("id: %d, biz: %d, app: %d, resource: %s, op: %s, resource_id: %d, uid: %s", meta.ID,
		meta.Attachment.BizID, meta.Attachment.AppID, meta.Spec.Resource, meta.Spec.OpType, meta.Spec.ResourceID,
		meta.Spec.ResourceUid)
}
