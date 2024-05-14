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

package sharding

import (
	"github.com/jmoiron/sqlx"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// Tx db affairs.
type Tx struct {
	// this sharding db instance's unique id, which is
	// generated when the process is launched.
	shardingUid string
	tx          *sqlx.Tx
}

// ShardingUid return sharding uid
func (t *Tx) ShardingUid() string {
	return t.shardingUid
}

// Tx return tx.
func (t *Tx) Tx() *sqlx.Tx {
	return t.tx
}

// Commit commit tx.
func (t *Tx) Commit(kit *kit.Kit) error {
	if err := t.tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Rollback rollback tx.
func (t *Tx) Rollback(kit *kit.Kit) error {
	if err := t.tx.Rollback(); err != nil {
		logs.ErrorDepthf(1, "db transaction rollback failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}
	return nil
}
