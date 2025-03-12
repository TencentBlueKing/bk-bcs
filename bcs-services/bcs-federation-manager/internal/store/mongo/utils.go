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

// Package mongo sub cluster store
package mongo

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
)

func ensureTable(ctx context.Context, public *Public) error {
	public.IsTableEnsuredMutex.RLock()
	if public.IsTableEnsured {
		public.IsTableEnsuredMutex.RUnlock()
		return nil
	}
	if err := ensure(ctx, public.DB, public.TableName, public.Indexes); err != nil {
		public.IsTableEnsuredMutex.RUnlock()
		return err
	}
	public.IsTableEnsuredMutex.RUnlock()

	public.IsTableEnsuredMutex.Lock()
	public.IsTableEnsured = true
	public.IsTableEnsuredMutex.Unlock()
	return nil
}

// ensure xxx
// EnsureTable ensure object database table and table indexes
func ensure(ctx context.Context, db drivers.DB, tableName string, indexes []drivers.Index) error {
	hasTable, err := db.HasTable(ctx, tableName)
	if err != nil {
		return err
	}
	if !hasTable {
		tErr := db.CreateTable(ctx, tableName)
		if tErr != nil {
			return tErr
		}
	}
	// only ensure index when index name is not empty
	for _, idx := range indexes {
		hasIndex, iErr := db.Table(tableName).HasIndex(ctx, idx.Name)
		if iErr != nil {
			return iErr
		}
		if !hasIndex {
			if iErr = db.Table(tableName).CreateIndex(ctx, idx); iErr != nil {
				return iErr
			}
		}
	}
	return nil
}

func formatSubclusterUID(fedClusterId, subClusterId string) string {
	return fmt.Sprintf("%s/%s", fedClusterId, subClusterId)
}
