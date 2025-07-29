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

package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/types"
)

var (
	modelPushWhitelistIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: pushDomainKey, Value: 1},
				bson.E{Key: pushWhitelistUniqueKey, Value: 1},
			},
			Name:   pushWhitelistTableName + "_1",
			Unique: true,
		},
		{
			Key: bson.D{
				bson.E{Key: pushWhitelistUniqueKey, Value: 1},
			},
			Name:   pushWhitelistUniqueKey + "_1",
			Unique: true,
		},
	}
)

// ModelPushWhitelist is a MongoDB-based implementation of PushWhitelistStore.
type ModelPushWhitelist struct {
	Public
}

// NewModelPushWhitelist creates a new PushWhitelistStore instance.
func NewModelPushWhitelist(db drivers.DB) *ModelPushWhitelist {
	return &ModelPushWhitelist{
		Public: Public{
			TableName: tableNamePrefix + pushWhitelistTableName,
			Indexes:   modelPushWhitelistIndexes,
			DB:        db,
		}}
}

// CreatePushWhitelist inserts a new push whitelist into the database.
func (m *ModelPushWhitelist) CreatePushWhitelist(ctx context.Context, whitelist *types.PushWhitelist) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return fmt.Errorf("ensure table failed: %v", err)
	}
	if whitelist == nil {
		return fmt.Errorf("push whitelist is nil")
	}

	whitelist.ID = primitive.NewObjectID()
	whitelist.CreatedAt = time.Now()
	whitelist.UpdatedAt = time.Now()

	_, err := m.DB.Table(m.TableName).Insert(ctx, []interface{}{whitelist})
	if err != nil {
		return fmt.Errorf("create push whitelist failed: %v", err)
	}
	return nil
}

// DeletePushWhitelist soft-deletes a push whitelist from the database by its ID.
func (m *ModelPushWhitelist) DeletePushWhitelist(ctx context.Context, whitelistID string) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return fmt.Errorf("ensure table failed: %v", err)
	}
	if whitelistID == "" {
		return fmt.Errorf("whitelistID is empty")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		pushWhitelistUniqueKey: whitelistID,
	})

	now := time.Now()
	update := operator.M{
		"$set": operator.M{
			"deleted_at": &now,
			"updated_at": now,
		},
	}

	if err := m.DB.Table(m.TableName).Update(ctx, cond, update); err != nil {
		return fmt.Errorf("delete push whitelist failed: %v", err)
	}
	return nil
}

// GetPushWhitelist retrieves a single push whitelist from the database by its ID.
func (m *ModelPushWhitelist) GetPushWhitelist(ctx context.Context, whitelistID string) (*types.PushWhitelist, error) {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return nil, fmt.Errorf("ensure table failed: %v", err)
	}
	if whitelistID == "" {
		return nil, fmt.Errorf("whitelistID cannot be empty")
	}

	cond := operator.NewBranchCondition(operator.And,
		operator.NewLeafCondition(operator.Eq, operator.M{
			pushWhitelistUniqueKey: whitelistID,
		}),
		operator.NewBranchCondition(operator.Or,
			operator.NewLeafCondition(operator.Eq, operator.M{"deleted_at": nil}),
			operator.NewLeafCondition(operator.Eq, operator.M{"deleted_at": operator.M{"$exists": false}}),
		),
	)

	var whitelist types.PushWhitelist
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, &whitelist); err != nil {
		return nil, fmt.Errorf("get push whitelist failed: %v", err)
	}
	return &whitelist, nil
}

// ListPushWhitelists retrieves a list of push whitelists from the database with filtering and pagination.
func (m *ModelPushWhitelist) ListPushWhitelists(ctx context.Context, filter operator.M, page, pageSize int64) ([]*types.PushWhitelist, int64, error) {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return nil, 0, fmt.Errorf("ensure table failed: %v", err)
	}

	cond := operator.NewBranchCondition(operator.Or,
		operator.NewLeafCondition(operator.Eq, operator.M{"deleted_at": nil}),
		operator.NewLeafCondition(operator.Eq, operator.M{"deleted_at": operator.M{"$exists": false}}),
	)

	if filter != nil {
		cond = operator.NewBranchCondition(operator.And, cond, operator.NewLeafCondition(operator.Eq, filter))
	}

	var whitelists []*types.PushWhitelist
	finder := m.DB.Table(m.TableName).Find(cond)
	if page > 1 {
		finder = finder.WithStart((page - 1) * pageSize)
	}
	if pageSize > 0 {
		finder = finder.WithLimit(pageSize)
	}
	if err := finder.All(ctx, &whitelists); err != nil {
		return nil, 0, fmt.Errorf("list push whitelists failed: %v", err)
	}

	total, err := finder.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count push whitelists failed: %v", err)
	}

	return whitelists, total, nil
}

// UpdatePushWhitelist updates a push whitelist in the database.
func (m *ModelPushWhitelist) UpdatePushWhitelist(ctx context.Context, whitelistID string, update operator.M) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return fmt.Errorf("ensure table failed: %v", err)
	}
	if whitelistID == "" {
		return fmt.Errorf("whitelistID cannot be empty")
	}

	cond := operator.NewBranchCondition(operator.And,
		operator.NewLeafCondition(operator.Eq, operator.M{
			pushWhitelistUniqueKey: whitelistID,
		}),
		operator.NewBranchCondition(operator.Or,
			operator.NewLeafCondition(operator.Eq, operator.M{"deleted_at": nil}),
			operator.NewLeafCondition(operator.Eq, operator.M{"deleted_at": operator.M{"$exists": false}}),
		),
	)

	if update["$set"] == nil {
		update["$set"] = operator.M{}
	}
	update["$set"].(operator.M)["updated_at"] = time.Now()

	if err := m.DB.Table(m.TableName).Update(ctx, cond, update); err != nil {
		return fmt.Errorf("update push whitelist failed: %v", err)
	}
	return nil
}

// IsDimensionWhitelisted checks if a given domain and dimension are whitelisted, active, and approved.
func (m *ModelPushWhitelist) IsDimensionWhitelisted(ctx context.Context, domain string, dimension types.Dimension) (bool, error) {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return false, fmt.Errorf("ensure table failed: %v", err)
	}

	now := time.Now()
	cond := operator.NewBranchCondition(operator.And,
		operator.NewLeafCondition(operator.Eq, operator.M{
			"domain":           domain,
			"approval_status":  constant.ApprovalStatusApproved,
			"whitelist_status": constant.WhitelistStatusActive,
		}),
		operator.NewBranchCondition(operator.Or,
			operator.NewLeafCondition(operator.Eq, operator.M{"deleted_at": nil}),
			operator.NewLeafCondition(operator.Eq, operator.M{"deleted_at": operator.M{"$exists": false}}),
		),
	)

	var whitelists []*types.PushWhitelist
	if err := m.DB.Table(m.TableName).Find(cond).All(ctx, &whitelists); err != nil {
		return false, fmt.Errorf("find whitelist failed: %v", err)
	}

	for _, wl := range whitelists {
		match, _ := mapEqualsDetail(wl.Dimension.Fields, dimension.Fields)
		if !match {
			continue
		}
		if !wl.StartTime.IsZero() && wl.StartTime.After(now) {
			continue
		}
		if !wl.EndTime.IsZero() && wl.EndTime.Before(now) {
			update := operator.M{"$set": operator.M{
				"whitelist_status": constant.WhitelistStatusExpired,
				"updated_at":       now,
			}}
			cond := operator.NewLeafCondition(operator.Eq, operator.M{"_id": wl.ID})
			err := m.DB.Table(m.TableName).Update(ctx, cond, update)
			if err != nil {
				return false, fmt.Errorf("update failed: %v", err)
			}
			continue
		}
		return true, nil
	}
	return false, nil
}

// mapEqualsDetail xxx
func mapEqualsDetail(a, b map[string]string) (bool, string) {
	if len(a) != len(b) {
		return false, fmt.Sprintf("length not equal: a=%d, b=%d", len(a), len(b))
	}
	var missing, extra, diff []string
	for k, v := range a {
		if bv, ok := b[k]; !ok {
			missing = append(missing, k)
		} else if bv != v {
			diff = append(diff, fmt.Sprintf("key=%s, a=%s, b=%s", k, v, bv))
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			extra = append(extra, k)
		}
	}
	if len(missing) == 0 && len(extra) == 0 && len(diff) == 0 {
		return true, "maps equal"
	}
	return false, fmt.Sprintf("missing in b: %v, extra in b: %v, value diff: %v", missing, extra, diff)
}
