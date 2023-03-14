/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"errors"
	"fmt"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// AuditDao supplies all the audit operations.
type AuditDao interface {
	// Decorator is used to handle the audit process as a pipeline
	// according CUD scenarios.
	Decorator(kit *kit.Kit, bizID uint32, res enumor.AuditResourceType) AuditDecorator
	// One insert one resource's audit.
	One(kit *kit.Kit, audit *table.Audit, opt *AuditOption) error
}

// AuditOption defines all the needed infos to audit a resource.
type AuditOption struct {
	// resource's transaction infos.
	Txn *sqlx.Tx
	// ResShardingUid is the resource's sharding instance.
	ResShardingUid string
}

var _ AuditDao = new(audit)

// NewAuditDao create the audit DAO
func NewAuditDao(orm orm.Interface, sd *sharding.Sharding, idGen IDGenInterface) (AuditDao, error) {
	return &audit{
		orm:        orm,
		sd:         sd,
		adSharding: sd.Audit(),
		idGen:      idGen,
	}, nil
}

type audit struct {
	orm orm.Interface
	// sd is the common resource's sharding manager.
	sd *sharding.Sharding
	// adSharding is the audit's sharding instance
	adSharding *sharding.One
	idGen      IDGenInterface
}

// Decorator return audit decorator for to record audit.
func (au *audit) Decorator(kit *kit.Kit, bizID uint32, res enumor.AuditResourceType) AuditDecorator {
	return initAuditBuilder(kit, bizID, res, au)
}

// One audit one resource's operation.
func (au *audit) One(kit *kit.Kit, audit *table.Audit, opt *AuditOption) error {
	if audit == nil || opt == nil {
		return errors.New("invalid input audit or opt")
	}

	// generate an audit id and update to audit.
	id, err := au.idGen.One(kit, table.AuditTable)
	if err != nil {
		return err
	}

	audit.ID = id

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", string(table.AuditTable), " (", table.AuditColumns.ColumnExpr(), ") VALUES (", table.AuditColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)

	if au.adSharding.ShardingUid() != opt.ResShardingUid {
		// audit db is different with the resource's db, then do without transaction
		if err := au.orm.Do(au.adSharding.DB()).Insert(kit.Ctx, sql, audit); err != nil {
			logs.Errorf("audit %s resource: %s, id: %s failed, err: %v, rid: %s",
				audit.Action, audit.ResourceType, audit.ResourceID, err, kit.Rid)
			// skip return this error to ensue the resource's transaction can be executed successfully.
			// this may miss the audit log, it's acceptable.
		}

		return nil
	}

	// do with the same transaction with the resource, this transaction
	// is launched by resource's owner.
	if err := au.orm.Txn(opt.Txn).Insert(kit.Ctx, sql, audit); err != nil {
		return fmt.Errorf("insert audit failed, err: %v", err)
	}

	return nil
}
