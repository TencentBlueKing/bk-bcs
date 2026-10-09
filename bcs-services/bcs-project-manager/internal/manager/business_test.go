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

package manager

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/business"
)

type fakeBusinessStore struct {
	upserted     []*business.Business
	deletedNotIn []string
	upsertErr    error
	deleteErr    error
}

func (f *fakeBusinessStore) UpsertBusiness(_ context.Context, biz *business.Business) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	copied := *biz
	f.upserted = append(f.upserted, &copied)
	return nil
}

func (f *fakeBusinessStore) DeleteBusinessesNotIn(_ context.Context, keepIDs []string) (int64, error) {
	if f.deleteErr != nil {
		return 0, f.deleteErr
	}
	f.deletedNotIn = append([]string{}, keepIDs...)
	return int64(len(keepIDs)), nil
}

func TestTransferBusiness(t *testing.T) {
	got := transferBusiness(cmdb.BusinessData{
		BKBizID:         100,
		BKBizName:       "demo",
		Default:         0,
		BKBizMaintainer: "alice,bob",
		BS2NameID:       7,
		BkBizProductor:  "p",
		BkBizTester:     "t",
		BkBizDeveloper:  "d",
	}, "2026-01-02T03:04:05Z")
	require.NotNil(t, got)
	assert.Equal(t, "100", got.BusinessID)
	assert.Equal(t, "demo", got.Name)
	assert.Equal(t, 0, got.Default)
	assert.Equal(t, "alice,bob", got.Maintainer)
	assert.Equal(t, 7, got.Bs2NameID)
	assert.Equal(t, "p", got.Productor)
	assert.Equal(t, "t", got.Tester)
	assert.Equal(t, "d", got.Developer)
	assert.Equal(t, "2026-01-02T03:04:05Z", got.SyncTime)
	assert.Nil(t, transferBusiness(cmdb.BusinessData{BKBizID: 0}, "now"))
}

func TestApplyBusinessSnapshot(t *testing.T) {
	fake := &fakeBusinessStore{}
	upserted, deleted, err := applyBusinessSnapshot(context.Background(), fake, []cmdb.BusinessData{
		{BKBizID: 1, BKBizName: "a"},
		{BKBizID: 0, BKBizName: "invalid"},
		{BKBizID: 2, BKBizName: "b"},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, upserted)
	assert.Equal(t, int64(2), deleted)
	require.Len(t, fake.upserted, 2)
	assert.Equal(t, []string{"1", "2"}, fake.deletedNotIn)
}

func TestApplySnapshotUpsertFail(t *testing.T) {
	fake := &fakeBusinessStore{upsertErr: fmt.Errorf("mongo down")}
	_, _, err := applyBusinessSnapshot(context.Background(), fake, []cmdb.BusinessData{
		{BKBizID: 1, BKBizName: "a"},
	})
	require.Error(t, err)
	assert.Nil(t, fake.deletedNotIn)
}

func TestApplySnapshotEmptyDel(t *testing.T) {
	fake := &fakeBusinessStore{}
	upserted, _, err := applyBusinessSnapshot(context.Background(), fake, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, upserted)
	assert.NotNil(t, fake.deletedNotIn)
	assert.Empty(t, fake.deletedNotIn)
}
