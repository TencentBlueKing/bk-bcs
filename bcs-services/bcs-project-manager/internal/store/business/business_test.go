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

package business

import (
	"context"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestValidateBusinessID(t *testing.T) {
	assert.NoError(t, validateBusinessID("1"))
	assert.NoError(t, validateBusinessID("12345678901234567890"))
	assert.Error(t, validateBusinessID(""))
	assert.Error(t, validateBusinessID("12a"))
	assert.Error(t, validateBusinessID("-1"))
	assert.Error(t, validateBusinessID("123456789012345678901"))
}

func TestEnsureTableDropsLegacyIndex(t *testing.T) {
	ctx := context.Background()
	db := memory.NewDB("test")
	m := New(db)
	require.NoError(t, db.CreateTable(ctx, m.tableName))
	require.NoError(t, db.Table(m.tableName).CreateIndex(ctx, drivers.Index{
		Name:   legacyBusinessIDIndex,
		Key:    bson.D{bson.E{Key: FieldKeyBusinessID, Value: 1}},
		Unique: true,
	}))

	require.NoError(t, m.ensureTable(ctx))

	hasLegacy, err := db.Table(m.tableName).HasIndex(ctx, legacyBusinessIDIndex)
	require.NoError(t, err)
	assert.False(t, hasLegacy)
	hasNew, err := db.Table(m.tableName).HasIndex(ctx, businessIndexes[0].Name)
	require.NoError(t, err)
	assert.True(t, hasNew)
}

func TestTenantIDRequired(t *testing.T) {
	ctx := context.Background()
	m := New(memory.NewDB("test"))
	assert.Error(t, m.UpsertBusiness(ctx, &Business{BusinessID: "1"}))
	_, err := m.DeleteBusinessesNotIn(ctx, "", []string{"1"})
	assert.Error(t, err)
}
