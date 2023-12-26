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

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
)

// AppCacheMeta defines app's basic meta info
type AppCacheMeta struct {
	Name       string           `json:"name"`
	ConfigType table.ConfigType `json:"cft"`
	// the current effected strategy set's type under this app.
	// only one strategy set is effected at one time.
	Mode   table.AppMode `json:"mod"`
	Reload *table.Reload `json:"reload"`
}

// ReleasedGroupCache is the released group info which will be stored in cache.
type ReleasedGroupCache struct {
	// ID is an auto-increased value, which is a group app's
	// unique identity.
	ID         uint32             `db:"id" json:"id"`
	GroupID    uint32             `db:"group_id" json:"group_id"`
	AppID      uint32             `db:"app_id" json:"app_id"`
	ReleaseID  uint32             `db:"release_id" json:"release_id"`
	StrategyID uint32             `db:"strategy_id" json:"strategy_id"`
	Mode       table.GroupMode    `db:"mode" json:"mode"`
	Selector   *selector.Selector `db:"selector" json:"selector"`
	UID        string             `db:"uid" json:"uid"`
	BizID      uint32             `db:"biz_id" json:"biz_id"`
	UpdatedAt  time.Time          `db:"updated_at" json:"updated_at"`
}

// EventMeta is an event's meta info which is used by feed server to gc cache.
type EventMeta struct {
	ID         uint32                 `db:"id" json:"id"`
	Spec       *table.EventSpec       `db:"spec" json:"spec"`
	Attachment *table.EventAttachment `db:"attachment" json:"attachment"`
}

// ReleaseCICache is the release config item info which will be stored in cache.
type ReleaseCICache struct {
	ID             uint32                      `json:"id"`
	ReleaseID      uint32                      `json:"reid"`
	CommitID       uint32                      `json:"cid"`
	CommitSpec     *CommitSpecCache            `json:"cspec"`
	ConfigItemID   uint32                      `json:"config_item_id"`
	ConfigItemSpec *ConfigItemSpecCache        `json:"ispec"`
	Attachment     *table.ConfigItemAttachment `json:"am"`
}

// ReleaseKvCache is the release kv info which will be stored in cache.
type ReleaseKvCache struct {
	ID         uint32              `json:"id"`
	ReleaseID  uint32              `json:"reid"`
	Key        string              `json:"key"`
	KvType     string              `json:"kv_type"`
	Reviser    string              `json:"reviser"`
	UpdateAt   string              `json:"update_at"`
	Attachment *table.KvAttachment `json:"am"`
}

// ReleasedHooksCache is the released hooks info which will be stored in cache.
type ReleasedHooksCache struct {
	AppID    uint32             `db:"app_id" json:"app_id"`
	BizID    uint32             `db:"biz_id" json:"biz_id"`
	PreHook  *ReleasedHookCache `db:"pre_hook" json:"pre_hook"`
	PostHook *ReleasedHookCache `db:"post_hook" json:"post_hook"`
}

// ReleasedHookCache is the release hook info which will be stored in cache.
type ReleasedHookCache struct {
	HookID         uint32           `db:"hook_id" json:"hook_id"`
	HookRevisionID uint32           `db:"hook_revision_id" json:"hook_revision_id"`
	Content        string           `db:"content" json:"content"`
	Type           table.ScriptType `db:"type" json:"type"`
}

// CommitSpecCache cache struct.
type CommitSpecCache struct {
	ContentID uint32 `json:"id"`
	Signature string `json:"sign"`
	ByteSize  uint64 `json:"size"`
}

// ConfigItemSpecCache cache struct.
type ConfigItemSpecCache struct {
	Name       string               `json:"name"`
	Path       string               `json:"path"`
	FileType   table.FileFormat     `json:"type"`
	FileMode   table.FileMode       `json:"mode"`
	Permission *FilePermissionCache `json:"pm"`
}

// CommitAttachmentCache cache struct.
type CommitAttachmentCache struct {
	BizID        uint32 `json:"bid"`
	AppID        uint32 `json:"aid"`
	ConfigItemID uint32 `json:"cid"`
}

// FilePermissionCache cache struct.
type FilePermissionCache struct {
	User      string `json:"user"`
	UserGroup string `json:"group"`
	Privilege string `json:"priv"`
}

// CredentialCache cache struct.
type CredentialCache struct {
	Enabled      bool                `json:"enabled"`
	Scope        []string            `json:"scope"`
	scopeMap     map[string][]string `json:"-"` // app:scope
	isPreprocess bool                `json:"-"`
}

// preprocess 预处理数据结构, 格式化为app:scope, 方便鉴权处理
func (c *CredentialCache) preprocess() {
	if c.isPreprocess {
		return
	}

	c.scopeMap = map[string][]string{}
	for _, v := range c.Scope {
		index := strings.Index(v, "/")
		if index == -1 {
			panic(fmt.Errorf("invalid scope %s", v))
		}

		app := v[:index]
		scope := v[index+1:]
		_, ok := c.scopeMap[app]
		if !ok {
			c.scopeMap[app] = []string{scope}
		} else {
			c.scopeMap[app] = append(c.scopeMap[app], scope)
		}
	}

	c.isPreprocess = true
}

// MatchApp 是否匹配 App
func (c *CredentialCache) MatchApp(app string) bool {
	c.preprocess()

	_, ok := c.scopeMap[app]
	return ok
}

// MatchKv 是否匹配kv
func (c *CredentialCache) MatchKv(app, key string) bool {
	scopes, ok := c.scopeMap[app]
	if !ok {
		return false
	}

	for _, v := range scopes {
		ok, err := path.Match(v, key)
		if err != nil {
			return false
		}
		if ok {
			return true
		}
	}

	return false
}

// ReleaseCICaches convert ReleasedConfigItem to ReleaseCICache.
func ReleaseCICaches(rs []*table.ReleasedConfigItem) []*ReleaseCICache {
	list := make([]*ReleaseCICache, len(rs))

	for index, one := range rs {
		list[index] = &ReleaseCICache{
			ID:           one.ID,
			ReleaseID:    one.ReleaseID,
			CommitID:     one.CommitID,
			ConfigItemID: one.ConfigItemID,
			CommitSpec: &CommitSpecCache{
				ContentID: one.CommitSpec.ContentID,
				Signature: one.CommitSpec.Content.Signature,
				ByteSize:  one.CommitSpec.Content.ByteSize,
			},
			ConfigItemSpec: &ConfigItemSpecCache{
				Name:     one.ConfigItemSpec.Name,
				Path:     one.ConfigItemSpec.Path,
				FileType: one.ConfigItemSpec.FileType,
				FileMode: one.ConfigItemSpec.FileMode,
				Permission: &FilePermissionCache{
					User:      one.ConfigItemSpec.Permission.User,
					UserGroup: one.ConfigItemSpec.Permission.UserGroup,
					Privilege: one.ConfigItemSpec.Permission.Privilege,
				},
			},
			Attachment: &table.ConfigItemAttachment{
				BizID: one.Attachment.BizID,
				AppID: one.Attachment.AppID,
			},
		}
	}

	return list
}

// ReleaseKvCaches convert ReleasedConfigItem to ReleaseKvCache.
func ReleaseKvCaches(rs []*table.ReleasedKv) []*ReleaseKvCache {
	list := make([]*ReleaseKvCache, len(rs))

	for index, one := range rs {
		list[index] = &ReleaseKvCache{
			ID:        one.ID,
			ReleaseID: one.ReleaseID,
			Key:       one.Spec.Key,
			KvType:    string(one.Spec.KvType),
			Reviser:   one.Revision.Reviser,
			UpdateAt:  one.Revision.UpdatedAt.Format(time.RFC3339),
			Attachment: &table.KvAttachment{
				BizID: one.Attachment.BizID,
				AppID: one.Attachment.AppID,
			},
		}
	}

	return list
}

// ReleaseKvValueCache is the release kv value info which will be stored in cache.
type ReleaseKvValueCache struct {
	ID        uint32 `json:"id"`
	ReleaseID uint32 `json:"reid"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	KvType    string `json:"kv_type"`
	Reviser   string `json:"reviser"`
	UpdateAt  string `json:"update_at"`
}
