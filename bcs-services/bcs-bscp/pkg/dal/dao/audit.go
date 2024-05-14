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

package dao

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/orm"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/sharding"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// AuditDao supplies all the audit operations.
type AuditDao interface {
	// Decorator is used to handle the audit process as a pipeline
	// according CUD scenarios.
	Decorator(kit *kit.Kit, bizID uint32, res enumor.AuditResourceType) AuditDecorator
	DecoratorV2(kit *kit.Kit, bizID uint32) AuditPrepare
	// One insert one resource's audit.
	One(kit *kit.Kit, audit *table.Audit, opt *AuditOption) error
}

// AuditOption defines all the needed infos to audit a resource.
type AuditOption struct {
	// resource's transaction infos.
	Txn *sqlx.Tx
	// ResShardingUid is the resource's sharding instance.
	ResShardingUid string
	genQ           *gen.Query
}

var _ AuditDao = new(audit)

// NewAuditDao create the audit DAO
func NewAuditDao(db *gorm.DB, orm orm.Interface, sd *sharding.Sharding, idGen IDGenInterface) (AuditDao, error) {
	return &audit{
		db:         db,
		genQ:       gen.Use(db),
		orm:        orm,
		sd:         sd,
		adSharding: sd.Audit(),
		idGen:      idGen,
	}, nil
}

type audit struct {
	db   *gorm.DB
	genQ *gen.Query
	orm  orm.Interface
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

// DecoratorV2 return audit decorator for to record audit.
func (au *audit) DecoratorV2(kit *kit.Kit, bizID uint32) AuditPrepare {
	return initAuditBuilderV2(kit, bizID, au)
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

	var q gen.IAuditDo

	if opt.genQ != nil && au.db.Migrator().CurrentDatabase() == opt.genQ.CurrentDatabase() {
		// 使用同一个库，事务处理
		q = opt.genQ.Audit.WithContext(kit.Ctx)
	} else {
		// 使用独立的 DB
		q = au.genQ.Audit.WithContext(kit.Ctx)
	}

	if err := q.Create(audit); err != nil {
		return fmt.Errorf("insert audit failed, err: %v", err)
	}
	return nil
}
