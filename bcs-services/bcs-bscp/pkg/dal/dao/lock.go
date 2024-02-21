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

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// LockDao supplies all the lock operations.
// NOTICE: the lock must be in the same transaction and database with the operation to lock.
type LockDao interface {
	// IncreaseCount increase the lock resource count, and returns the previous count.
	// need to call DecreaseCount after the resource is deleted to ensure the lock count is correct.
	IncreaseCount(kit *kit.Kit, tx *gen.Query, lock *table.ResourceLock, num uint32) (uint32, error)
	// DecreaseCount decrease the lock resource count, if the lock count is zero, delete the lock.
	DecreaseCount(kit *kit.Kit, tx *gen.Query, lock *table.ResourceLock, num uint32) error
	// AddUnique validate if the resource is unique by adding a lock with unique index, returns true if it is unique.
	// need to call DeleteUnique after the resource is deleted to ensure the lock unique is correct.
	AddUnique(kit *kit.Kit, tx *gen.Query, lock *table.ResourceLock) (bool, error)
	// DeleteUnique delete the unique resource lock.
	DeleteUnique(kit *kit.Kit, tx *gen.Query, lock *table.ResourceLock) error
}

var _ LockDao = new(lockDao)

type lockDao struct {
	genQ  *gen.Query
	idGen IDGenInterface
}

// IncreaseCount increase the lock resource count, and returns current count.
// need to call DecreaseCount after the resource is deleted to ensure the lock count is correct.
func (dao *lockDao) IncreaseCount(kit *kit.Kit, tx *gen.Query, lock *table.ResourceLock, num uint32) (uint32, error) {
	if lock == nil {
		return 0, errf.New(errf.InvalidParameter, "lock is nil")
	}

	if err := lock.Validate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	m := tx.ResourceLock

	result, err := m.WithContext(kit.Ctx).
		Where(m.ResType.Eq(lock.ResType), m.ResKey.Eq(lock.ResKey), m.BizID.Eq(lock.BizID)).
		Update(m.ResCount, m.ResCount.Add(num))

	if err != nil {
		logs.Errorf("increase lock count failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return 0, fmt.Errorf("increase lock count failed, err: %v", err)
	}

	if result.RowsAffected > 1 {
		logs.Errorf("lock rows affected is %d, should be 0 or 1, lock: %v, rid: %s", result.RowsAffected, lock, kit.Rid)
		return 0, fmt.Errorf("lock rows affected is %d", result.RowsAffected)
	}

	// the lock exists, get the count from db and returns the count before the operation
	if result.RowsAffected == 1 {
		var count uint32
		count, err = dao.getLockCount(kit, tx, lock)
		if err != nil {
			return 0, err
		}
		return count, nil
	}

	// the lock key is not exist, set count = 1 and insert it, returns 0.
	lock.ResCount = num
	var id uint32
	id, err = dao.idGen.One(kit, table.ResourceLockTable)
	if err != nil {
		logs.Errorf("generate lock id failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return 0, fmt.Errorf("insert lock failed, err: %v", err)
	}
	lock.ID = id

	if e := m.WithContext(kit.Ctx).Create(lock); e != nil {
		// nolint
		// TODO: 压测看看并发插入时是否会导致死锁，再如果死锁是否需要重试事务
		logs.Errorf("insert lock failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return 0, fmt.Errorf("insert lock failed, err: %v", err)
	}
	return 0, nil
}

func (dao *lockDao) getLockCount(kit *kit.Kit, tx *gen.Query, lock *table.ResourceLock) (uint32, error) {
	m := tx.ResourceLock
	err := m.WithContext(kit.Ctx).
		Select(m.ResCount).
		Where(m.ResType.Eq(lock.ResType), m.ResKey.Eq(lock.ResKey), m.BizID.Eq(lock.BizID)).
		Scan(&lock.ResCount)
	if err != nil {
		logs.Errorf("query lock count failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return 0, fmt.Errorf("get lock failed, err: %v", err)
	}

	// validate if the lock count is valid or not.
	if lock.ResCount < 1 {
		logs.Errorf("get invalid lock count %d, lock: %v, rid: %s", lock.ResCount, lock, kit.Rid)
		return 0, fmt.Errorf("get invalid lock count %d", lock.ResCount)
	}

	return lock.ResCount, nil
}

// DecreaseCount decrease the lock resource count, if the lock count is zero, delete the lock.
func (dao *lockDao) DecreaseCount(kit *kit.Kit, tx *gen.Query, lock *table.ResourceLock, num uint32) error {
	if lock == nil {
		return errf.New(errf.InvalidParameter, "lock is nil")
	}

	if err := lock.Validate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	count, err := dao.getLockCount(kit, tx, lock)
	if err != nil {
		logs.Errorf("decrease lock count failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return fmt.Errorf("decrease lock count failed, err: %v", err)
	}

	m := tx.ResourceLock

	// the current lock is related to more than one resource, decrease the lock count.
	if count > 1 {
		if _, err := m.WithContext(kit.Ctx).
			Where(m.ResType.Eq(lock.ResType), m.ResKey.Eq(lock.ResKey), m.BizID.Eq(lock.BizID)).
			Update(m.ResCount, m.ResCount.Sub(num)); err != nil {
			return err
		}
		return nil
	}
	// the current lock is related to only one resource, delete the lock.
	if _, err := m.WithContext(kit.Ctx).
		Where(m.ResType.Eq(lock.ResType), m.ResKey.Eq(lock.ResKey), m.BizID.Eq(lock.BizID)).
		Delete(); err != nil {
		logs.Errorf("delete lock failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return fmt.Errorf("delete lock failed, err: %v", err)
	}
	return nil
}

// AddUnique validate if the resource is unique by adding a lock with unique index, returns true if it is unique.
// need to call DeleteUnique after the resource is deleted to ensure the lock unique is correct.
func (dao *lockDao) AddUnique(kit *kit.Kit, tx *gen.Query, lock *table.ResourceLock) (bool, error) {
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

	m := tx.ResourceLock
	if e := m.WithContext(kit.Ctx).Create(lock); e != nil {
		logs.Errorf("insert lock failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		if errors.Is(e, gorm.ErrDuplicatedKey) {
			return false, nil
		}
		return false, fmt.Errorf("add lock failed, err: %v", err)
	}

	return true, nil
}

// DeleteUnique one resource, decrease the lock count, if the lock count is zero, delete the lock.
func (dao *lockDao) DeleteUnique(kit *kit.Kit, tx *gen.Query, lock *table.ResourceLock) error {
	if lock == nil {
		return errf.New(errf.InvalidParameter, "lock is nil")
	}

	if err := lock.Validate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	count, err := dao.getLockCount(kit, tx, lock)
	if err != nil {
		return err
	}

	if count > 1 {
		logs.Errorf("unique lock count(%d) is more than one, lock: %v,  rid: %s", count, lock, kit.Rid)
		return fmt.Errorf("unique lock has %d count", count)
	}

	m := tx.ResourceLock
	if _, err := m.WithContext(kit.Ctx).
		Where(m.ResType.Eq(lock.ResType), m.ResKey.Eq(lock.ResKey), m.BizID.Eq(lock.BizID)).
		Delete(); err != nil {
		logs.Errorf("delete lock failed, lock: %v, err: %v, rid: %s", lock, err, kit.Rid)
		return fmt.Errorf("delete lock failed, err: %v", err)
	}
	return nil
}
