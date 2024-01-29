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

package table

import (
	"errors"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
)

// ResLockColumns defines all the ResourceLock table's columns.
var ResLockColumns = mergeColumns(ResLockColumnDescriptor)

// ResLockColumnDescriptor is ResourceLock's column descriptors.
var ResLockColumnDescriptor = ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "res_type", NamedC: "res_type", Type: enumor.String},
	{Column: "res_key", NamedC: "res_key", Type: enumor.String},
	{Column: "res_count", NamedC: "res_count", Type: enumor.Numeric},
}

// ResourceLock is used to help test the uniqueness of a resource and the limitation of a kind of resources which
// can not be done within the resource's related database table with unique indexes, such as the unique limitation
// about the strategy's namespaces under a same app.
// This table has a combined unique index with BizID, ResType and ResKey which can limit only one 'ResKey' can
// exist under a ResType within the same business. The ResKey is a string, which can let you define different
// kind of compound resources under the different scenarios, which is very flexible and convenient.
// The ResCount is increased or decreased when an resource is created or deleted within the same database transaction,
// this can help us to limit the count of a kind of resources.
type ResourceLock struct {
	ID       uint32 `db:"id" gorm:"primaryKey" json:"id"`
	BizID    uint32 `db:"biz_id" gorm:"column:biz_id" json:"biz_id"`
	ResType  string `db:"res_type" gorm:"column:res_type" json:"resource_type"`
	ResKey   string `db:"res_key" gorm:"column:res_key" json:"resource_key"`
	ResCount uint32 `db:"res_count" gorm:"column:res_count" json:"resource_count"`
}

// TableName is the lock's database table name.
func (l ResourceLock) TableName() Name {
	return ResourceLockTable
}

// Validate the resource lock.
func (l ResourceLock) Validate() error {
	if l.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if len(l.ResType) == 0 {
		return errors.New("resource type should be set")
	}

	if len(l.ResKey) == 0 {
		return errors.New("resource key should be set")
	}

	if len(l.ResKey) > 256 {
		return errors.New("resource key exceeds maximum length")
	}

	return nil
}
