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

package dao

import (
	"fmt"
	"strconv"
	"time"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"

	"github.com/jmoiron/sqlx"
)

// CRInstance supplies all the current released instance related operations.
type CRInstance interface {
	// Create one current released instance.
	Create(kit *kit.Kit, cri *table.CurrentReleasedInstance) (uint32, error)
	// List current released instance with options.
	List(kit *kit.Kit, opts *types.ListCRInstancesOption) (*types.ListCRInstanceDetails, error)
	// ListAppCRIMeta list an app's all the released instance meta info.
	ListAppCRIMeta(kit *kit.Kit, bizID uint32, appID uint32) ([]*types.AppCRIMeta, error)
	// GetAppCRIMeta get the released instance meta info by uid.
	GetAppCRIMeta(kit *kit.Kit, bizID uint32, appID uint32, uid string) (*types.AppCRIMeta, error)
	// Delete one current released instance.
	Delete(kit *kit.Kit, cri *table.CurrentReleasedInstance) error
}

var _ CRInstance = new(crInstanceDao)

type crInstanceDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	event    Event
	lock     LockDao
}

// Create one current released instance.
func (dao *crInstanceDao) Create(kit *kit.Kit, cri *table.CurrentReleasedInstance) (uint32, error) {

	if cri == nil {
		return 0, errf.New(errf.InvalidParameter, "current released instance is nil")
	}

	if err := cri.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentResExist(kit, cri.Attachment); err != nil {
		return 0, err
	}

	// validate instance binding release exist.
	if err := dao.validateReleaseExist(kit, cri.Attachment.BizID, cri.Spec.ReleaseID); err != nil {
		return 0, err
	}

	// generate a current released instance id and update to current released instance.
	id, err := dao.idGen.One(kit, table.CurrentReleasedInstanceTable)
	if err != nil {
		return 0, err
	}

	cri.ID = id

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", string(table.CurrentReleasedInstanceTable),
		" (", table.CurrentReleasedInstanceColumns.ColumnExpr(), ")  VALUES(", table.CurrentReleasedInstanceColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)

	eDecorator := dao.event.Eventf(kit)
	err = dao.sd.ShardingOne(cri.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err = dao.validateAppCRINumber(kit, cri.Attachment, &LockOption{Txn: txn}); err != nil {
				return err
			}

			if err = dao.orm.Txn(txn).Insert(kit.Ctx, sql, cri); err != nil {
				return err
			}

			// audit publish this to be published current released instance details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, cri.Attachment.BizID, enumor.CRInstance).
				AuditPublish(cri, au); err != nil {
				return fmt.Errorf("audit publish instance failed, err: %v", err)
			}

			// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
			one := types.Event{
				Spec: &table.EventSpec{
					Resource: table.PublishInstance,
					// use the published instance id, which represent a real publish operation.
					ResourceUid: cri.Spec.Uid,
					OpType:      table.InsertOp,
				},
				Attachment: &table.EventAttachment{BizID: cri.Attachment.BizID, AppID: cri.Attachment.AppID},
				Revision:   &table.CreatedRevision{Creator: kit.User, CreatedAt: time.Now()},
			}
			if err = eDecorator.Fire(one); err != nil {
				logs.Errorf("fire publish instance: %s event failed, err: %v, rid: %s", cri.Spec.Uid,
					err, kit.Rid)
				return errf.New(errf.DBOpFailed, "fire event failed, "+err.Error())
			}

			return nil
		})

	eDecorator.Finalizer(err)

	if err != nil {
		logs.Errorf("create current released instance, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create current released instance, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// List current released instance with options.
func (dao *crInstanceDao) List(kit *kit.Kit, opts *types.ListCRInstancesOption) (*types.ListCRInstanceDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list current released instance options null")
	}

	if err := opts.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "uid", "biz_id"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "biz_id",
					Op:    filter.Equal.Factory(),
					Value: opts.BizID,
				},
			},
		},
	}
	whereExpr, arg, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sql string
	var sqlSentence []string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sqlSentence = append(sqlSentence, "SELECT COUNT(*) FROM ", string(table.CurrentReleasedInstanceTable), whereExpr)
		sql = filter.SqlJoint(sqlSentence)
		var count uint32
		count, err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql, arg)
		if err != nil {
			return nil, err
		}

		return &types.ListCRInstanceDetails{Count: count, Details: make([]*table.CurrentReleasedInstance, 0)}, nil
	}

	// query current released instance list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sqlSentence = append(sqlSentence, "SELECT ", table.CurrentReleasedInstanceColumns.NamedExpr(), " FROM ",
		string(table.CurrentReleasedInstanceTable), whereExpr, pageExpr)
	sql = filter.SqlJoint(sqlSentence)

	list := make([]*table.CurrentReleasedInstance, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, arg)
	if err != nil {
		return nil, err
	}

	return &types.ListCRInstanceDetails{Count: 0, Details: list}, nil
}

// CRIMeta defines an app's current released instance meta info
type CRIMeta struct {
	ID        uint32 `db:"id"`
	Uid       string `db:"uid"`
	ReleaseID uint32 `db:"release_id"`
}

// ListAppCRIMeta list an app's all the released instance meta info.
func (dao *crInstanceDao) ListAppCRIMeta(kit *kit.Kit, bizID uint32, appID uint32) ([]*types.AppCRIMeta, error) {

	var step uint = 200
	page := &types.BasePage{
		Count: false,
		Start: 0,
		Limit: step,
	}
	pageExpr, err := page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	result := make([]*types.AppCRIMeta, 0)
	var id uint32 = 0
	for start := uint32(0); ; start += uint32(step) {
		var sqlSentence []string
		sqlSentence = append(sqlSentence, "SELECT id, uid, release_id FROM ", string(table.CurrentReleasedInstanceTable),
			" WHERE biz_id = ", strconv.Itoa(int(bizID)), " AND app_id = ", strconv.Itoa(int(appID)), " AND id > ", strconv.Itoa(int(id)), pageExpr)
		sql := filter.SqlJoint(sqlSentence)

		list := make([]*CRIMeta, 0)
		if err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Select(kit.Ctx, &list, sql); err != nil {
			return nil, err
		}

		if len(list) == 0 {
			break
		}

		for _, one := range list {
			result = append(result, &types.AppCRIMeta{
				Uid:       one.Uid,
				ReleaseID: one.ReleaseID,
			})
		}

		if len(list) < int(step) {
			break
		}

		id = list[len(list)-1].ID
	}

	return result, nil
}

// GetAppCRIMeta get the released instance meta info by uid.
func (dao *crInstanceDao) GetAppCRIMeta(kit *kit.Kit, bizID uint32, appID uint32, uid string) (*types.AppCRIMeta,
	error) {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT uid, release_id FROM ", string(table.CurrentReleasedInstanceTable),
		" WHERE uid = '", uid, "' AND app_id = ", strconv.Itoa(int(appID)))
	sql := filter.SqlJoint(sqlSentence)

	result := new(types.AppCRIMeta)
	if err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Get(kit.Ctx, result, sql); err != nil {
		if err == orm.ErrRecordNotFound {
			return new(types.AppCRIMeta), nil
		}

		return nil, err
	}

	return result, nil
}

// Delete one current released instance.
func (dao *crInstanceDao) Delete(kit *kit.Kit, cri *table.CurrentReleasedInstance) error {

	if cri == nil {
		return errf.New(errf.InvalidParameter, "current released instance is nil")
	}

	if err := cri.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, cri.Attachment); err != nil {
		return err
	}

	spec, err := dao.queryInstanceSpec(kit, cri.Attachment.BizID, cri.Attachment.AppID, cri.ID)
	if err != nil {
		return err
	}
	cri.Spec = spec

	ab := dao.auditDao.Decorator(kit, cri.Attachment.BizID, enumor.CRInstance).PrepareDelete(cri.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", string(table.CurrentReleasedInstanceTable), " WHERE id = ", strconv.Itoa(int(cri.ID)),
		" AND biz_id = ", strconv.Itoa(int(cri.Attachment.BizID)))
	expr := filter.SqlJoint(sqlSentence)

	eDecorator := dao.event.Eventf(kit)
	err = dao.sd.ShardingOne(cri.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		// delete the current released instance at first.
		err := dao.orm.Txn(txn).Delete(kit.Ctx, expr)
		if err != nil {
			return err
		}

		// audit this delete current released instance details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete current released instance failed, err: %v", err)
		}

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource:    table.PublishInstance,
				ResourceUid: cri.Spec.Uid,
				OpType:      table.DeleteOp,
			},
			Attachment: &table.EventAttachment{BizID: cri.Attachment.BizID, AppID: cri.Attachment.AppID},
			Revision:   &table.CreatedRevision{Creator: kit.User, CreatedAt: time.Now()},
		}
		if err := eDecorator.Fire(one); err != nil {
			logs.Errorf("fire delete instance: %s publish event failed, err: %v, rid: %s", cri.Spec.Uid,
				err, kit.Rid)
			return errf.New(errf.DBOpFailed, "fire event failed, "+err.Error())
		}

		// decrease the current released instance lock count after the deletion
		lock := lockKey.CurReleasedInst(cri.Attachment.BizID, cri.Attachment.AppID)
		if err := dao.lock.DecreaseCount(kit, lock, &LockOption{Txn: txn}); err != nil {
			return err
		}

		return nil
	})

	eDecorator.Finalizer(err)

	if err != nil {
		logs.Errorf("delete current released instance: %d failed, err: %v, rid: %v", cri.ID, err, kit.Rid)
		return fmt.Errorf("delete current released instance, but run txn failed, err: %v", err)
	}

	return nil
}

// queryInstanceSpec query instance spec info.
func (dao *crInstanceDao) queryInstanceSpec(kt *kit.Kit, bizID, appID, id uint32) (*table.ReleasedInstanceSpec,
	error) {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.RISpecColumns.NamedExpr(), " FROM ", string(table.CurrentReleasedInstanceTable),
		" WHERE id = ", strconv.Itoa(int(id)), " And biz_id = ", strconv.Itoa(int(bizID)), " And app_id = ", strconv.Itoa(int(appID)))
	sql := filter.SqlJoint(sqlSentence)

	spec := new(table.ReleasedInstanceSpec)
	err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Get(kt.Ctx, spec, sql)
	if err != nil {
		return nil, errf.New(errf.DBOpFailed, err.Error())
	}

	return spec, nil
}

// validateReleaseExist validate if instance's release exists before creating.
func (dao *crInstanceDao) validateReleaseExist(kt *kit.Kit, bizID, releaseID uint32) error {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE id = ", strconv.Itoa(int(releaseID)), " AND biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kt, dao.orm, dao.sd.ShardingOne(bizID), table.ReleaseTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RecordNotFound, fmt.Sprintf("release %d is not exist", releaseID))
	}

	return nil
}

// validateAttachmentResExist validate if attachment resource exists before creating current release instance.
func (dao *crInstanceDao) validateAttachmentResExist(kit *kit.Kit, am *table.ReleaseAttachment) error {
	return dao.validateAttachmentAppExist(kit, am)
}

// validateAttachmentAppExist validate if attachment app exists before creating current release instance.
func (dao *crInstanceDao) validateAttachmentAppExist(kit *kit.Kit, am *table.ReleaseAttachment) error {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE id = ", strconv.Itoa(int(am.AppID)), " AND biz_id = ", strconv.Itoa(int(am.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("current released instance attached app %d "+
			"is not exist", am.AppID))
	}

	return nil
}

// validateAppCRINumber verify whether the current number of app current released
// instance has reached the maximum.
func (dao *crInstanceDao) validateAppCRINumber(kt *kit.Kit, at *table.ReleaseAttachment, lo *LockOption) error {
	// try lock current released instance to ensure the number is limited when creating concurrently
	lock := lockKey.CurReleasedInst(at.BizID, at.AppID)
	count, err := dao.lock.IncreaseCount(kt, lock, lo)
	if err != nil {
		return err
	}

	if err := table.ValidateAppCRINumber(count); err != nil {
		return err
	}

	return nil
}
