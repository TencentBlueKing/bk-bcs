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
	"fmt"
	"strconv"
	"strings"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// LockDao supplies all the lock operations.
// NOTICE: the lock must be in the same transaction and database with the operation to lock.
type LockDao interface {
	// IncreaseCount increase the lock resource count, and returns the previous count.
	// need to call DecreaseCount after the resource is deleted to ensure the lock count is correct.
	IncreaseCount(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) (uint32, error)
	// DecreaseCount decrease the lock resource count, if the lock count is zero, delete the lock.
	DecreaseCount(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) error
	// TruncateCount truncate the lock resource count to zero.
	TruncateCount(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) error
	// AddUnique validate if the resource is unique by adding a lock with unique index, returns true if it is unique.
	// need to call DeleteUnique after the resource is deleted to ensure the lock unique is correct.
	AddUnique(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) (bool, error)
	// DeleteUnique delete the unique resource lock.
	DeleteUnique(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) error
}

// LockOption defines all the needed infos to lock a resource.
type LockOption struct {
	// Txn resource's transaction infos.
	Txn *sqlx.Tx
}

var _ LockDao = new(lockDao)

type lockDao struct {
	orm   orm.Interface
	idGen IDGenInterface
}

// IncreaseCount increase the lock resource count, and returns the previous count.
// need to call DecreaseCount after the resource is deleted to ensure the lock count is correct.
func (dao *lockDao) IncreaseCount(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) (uint32, error) {
	if lock == nil {
		return 0, errf.New(errf.InvalidParameter, "lock is nil")
	}

	if err := lock.Validate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// update the lock and increase the count.
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.ResourceLockTable.Name(), " SET res_count = res_count + 1 WHERE biz_id = ",
		strconv.Itoa(int(lock.BizID)), " AND res_type = '", lock.ResType, "' AND res_key = '", lock.ResKey, "'")
	sql := filter.SqlJoint(sqlSentence)

	result, err := opt.Txn.ExecContext(kit.Ctx, sql)
	if err != nil {
		logs.Errorf("increase lock count failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return 0, fmt.Errorf("get lock failed, err: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logs.Errorf("get lock rows affected failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return 0, fmt.Errorf("get lock failed, err: %v", err)
	}

	if rowsAffected > 1 {
		logs.Errorf("lock rows affected is %d, should be 0 or 1, lock: %v, rid: %s", rowsAffected, lock, kit.Rid)
		return 0, fmt.Errorf("lock rows affected is %d", rowsAffected)
	}

	// the lock exists, get the count from db and returns the count before the operation
	if rowsAffected == 1 {
		count, err := dao.getLockCount(kit, lock, opt)
		if err != nil {
			return 0, err
		}
		return count - 1, nil
	}

	// the lock key is not exist, set count = 1 and insert it, returns 0.
	lock.ResCount = 1
	id, err := dao.idGen.One(kit, table.ResourceLockTable)
	if err != nil {
		logs.Errorf("generate lock id failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return 0, fmt.Errorf("get lock failed, err: %v", err)
	}
	lock.ID = id

	var sqlSentenceIn []string
	sqlSentenceIn = append(sqlSentenceIn, "INSERT INTO ", table.ResourceLockTable.Name(), " (", table.ResLockColumns.ColumnExpr(),
		") VALUES (", table.ResLockColumns.ColonNameExpr(), ")")
	sql = filter.SqlJoint(sqlSentenceIn)
	if err := dao.orm.Txn(opt.Txn).Insert(kit.Ctx, sql, lock); err != nil {
		logs.Errorf("insert lock failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		// mysql will add a gap lock if updated row is not exist, and gap lock allows multiple transaction to acquire,
		// and insert needs an exclusive lock. so in concurrent scenario, insert will wait for the gap lock to release
		// which will result in a deadlock. so we will retry the transaction later to avoid the conflict.
		if strings.Contains(err.Error(), orm.ErrDeadLock) {
			return 0, sharding.ErrRetryTransaction
		}
		return 0, fmt.Errorf("get lock failed, err: %v", err)
	}

	return 0, nil
}

func (dao *lockDao) getLockCount(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) (uint32, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT res_count from ", table.ResourceLockTable.Name(), " WHERE biz_id = ", strconv.Itoa(int(lock.BizID)),
		" AND res_type = '", lock.ResType, "' AND res_key = '", lock.ResKey, "'")
	queryExpr := filter.SqlJoint(sqlSentence)

	rows, err := opt.Txn.QueryContext(kit.Ctx, queryExpr)
	if err != nil {
		logs.Errorf("query lock count failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return 0, fmt.Errorf("get lock failed, err: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&lock.ResCount); err != nil {
			logs.Errorf("scan lock count failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
			return 0, fmt.Errorf("get lock failed, err: %v", err)
		}
		// only one row is queried, "biz_id + res_type + res_key" is a unique index key.
		break
	}

	// validate if the lock count is valid or not.
	if lock.ResCount < 1 {
		logs.Errorf("get invalid lock count %d, lock: %v, rid: %s", lock.ResCount, lock, kit.Rid)
		return 0, fmt.Errorf("get invalid lock count %d", lock.ResCount)
	}

	return lock.ResCount, nil
}

// DecreaseCount decrease the lock resource count, if the lock count is zero, delete the lock.
func (dao *lockDao) DecreaseCount(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) error {
	if lock == nil {
		return errf.New(errf.InvalidParameter, "lock is nil")
	}

	if err := lock.Validate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	count, err := dao.getLockCount(kit, lock, opt)
	if err != nil {
		return err
	}

	// the current lock is related to more than one resource, decrease the lock count.
	var sqlSentence []string
	if count > 1 {
		sqlSentence = append(sqlSentence, "UPDATE ", table.ResourceLockTable.Name(), " SET res_count = res_count - 1 WHERE biz_id = ", strconv.Itoa(int(lock.BizID)),
			" AND res_type = '", lock.ResType, "' AND res_key = '", lock.ResKey, "'")
		sql := filter.SqlJoint(sqlSentence)

		_, err := opt.Txn.ExecContext(kit.Ctx, sql)
		if err != nil {
			logs.Errorf("decrease lock count failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
			return fmt.Errorf("delete lock failed, err: %v", err)
		}

		return nil
	}

	// the current lock is related to only one resource, delete the lock.
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.ResourceLockTable.Name(), " WHERE biz_id = ", strconv.Itoa(int(lock.BizID)),
		" AND res_type = '", lock.ResType, "' AND res_key = '", lock.ResKey, "'")
	sql := filter.SqlJoint(sqlSentence)

	_, err = opt.Txn.ExecContext(kit.Ctx, sql)
	if err != nil {
		logs.Errorf("delete lock failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return fmt.Errorf("delete lock failed, err: %v", err)
	}

	return nil
}

// TruncateLock truncate the lock resource count to zero.
func (dao *lockDao) TruncateCount(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) error {
	if lock == nil {
		return errf.New(errf.InvalidParameter, "lock is nil")
	}

	if err := lock.Validate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	// set the lock count to zero.
	var sqlSentence []string
	// the current lock is related to only one resource, delete the lock.
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.ResourceLockTable.Name(),
		" WHERE biz_id = ", strconv.Itoa(int(lock.BizID)),
		" AND res_type = '", lock.ResType, "' AND res_key = '", lock.ResKey, "'")
	sql := filter.SqlJoint(sqlSentence)

	_, err := opt.Txn.ExecContext(kit.Ctx, sql)
	if err != nil {
		logs.Errorf("delete lock failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return fmt.Errorf("delete lock failed, err: %v", err)
	}

	return nil
}

// AddUnique validate if the resource is unique by adding a lock with unique index, returns true if it is unique.
// need to call DeleteUnique after the resource is deleted to ensure the lock unique is correct.
func (dao *lockDao) AddUnique(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) (bool, error) {
	if lock == nil {
		return false, errf.New(errf.InvalidParameter, "lock is nil")
	}

	if err := lock.Validate(); err != nil {
		return false, errf.New(errf.InvalidParameter, err.Error())
	}

	lock.ResCount = 1
	id, err := dao.idGen.One(kit, table.ResourceLockTable)
	if err != nil {
		logs.Errorf("generate lock id failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return false, fmt.Errorf("get lock failed, err: %v", err)
	}
	lock.ID = id

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.ResourceLockTable.Name(), " (", table.ResLockColumns.ColumnExpr(),
		") VALUES (", table.ResLockColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)
	if err := dao.orm.Txn(opt.Txn).Insert(kit.Ctx, sql, lock); err != nil {
		logs.Errorf("insert lock failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		if strings.Contains(err.Error(), "Error 1062: Duplicate entry") {
			return false, nil
		}
		return false, fmt.Errorf("add lock failed, err: %v", err)
	}

	return true, nil
}

// DeleteUnique one resource, decrease the lock count, if the lock count is zero, delete the lock.
func (dao *lockDao) DeleteUnique(kit *kit.Kit, lock *table.ResourceLock, opt *LockOption) error {
	if lock == nil {
		return errf.New(errf.InvalidParameter, "lock is nil")
	}

	if err := lock.Validate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	count, err := dao.getLockCount(kit, lock, opt)
	if err != nil {
		return err
	}

	if count > 1 {
		logs.Errorf("unique lock count(%d) is more than one, lock: %v,  rid: %s", count, lock, kit.Rid)
		return fmt.Errorf("unique lock has %d count", count)
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.ResourceLockTable.Name(), " WHERE biz_id = ", strconv.Itoa(int(lock.BizID)),
		" AND res_type = '", lock.ResType, "' AND res_key = '", lock.ResKey, "'")
	sql := filter.SqlJoint(sqlSentence)

	_, err = opt.Txn.ExecContext(kit.Ctx, sql)
	if err != nil {
		logs.Errorf("delete lock failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return fmt.Errorf("delete lock failed, err: %v", err)
	}

	return nil
}
