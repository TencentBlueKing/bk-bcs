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
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
)

// AuditColumns defines all the audit table's columns.
var AuditColumns = mergeColumns(AuditColumnDescriptor)

// AuditColumnDescriptor is Audit's column descriptors.
var AuditColumnDescriptor = ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric},
	{Column: "res_type", NamedC: "res_type", Type: enumor.String},
	{Column: "res_id", NamedC: "res_id", Type: enumor.Numeric},
	{Column: "action", NamedC: "action", Type: enumor.String},
	{Column: "rid", NamedC: "rid", Type: enumor.String},
	{Column: "app_code", NamedC: "app_code", Type: enumor.String},
	{Column: "operator", NamedC: "operator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "detail", NamedC: "detail", Type: enumor.String}}

// Audit is used to save resource's audit information.
type Audit struct {
	ID           uint32                   `db:"id" json:"id" gorm:"primaryKey"`
	BizID        uint32                   `db:"biz_id" json:"biz_id" gorm:"column:biz_id"`
	AppID        uint32                   `db:"app_id" json:"app_id" gorm:"column:app_id"`
	ResourceType enumor.AuditResourceType `db:"res_type" json:"resource_type" gorm:"column:res_type"`
	ResourceID   uint32                   `db:"res_id" json:"resource_id" gorm:"column:res_id"`
	Action       enumor.AuditAction       `db:"action" json:"action" gorm:"column:action"`
	Rid          string                   `db:"rid" json:"rid" gorm:"column:rid"`
	AppCode      string                   `db:"app_code" json:"app_code" gorm:"column:app_code"`
	Operator     string                   `db:"operator" json:"operator" gorm:"column:operator"`
	CreatedAt    time.Time                `db:"created_at" json:"created_at" gorm:"column:created_at"`
	Detail       string                   `db:"detail" json:"detail" gorm:"column:detail"` // Detail is a json raw string
	ResInstance  string                   `db:"res_instance" json:"res_instance" gorm:"column:res_instance"`
	OperateWay   string                   `db:"operate_way" json:"operate_way" gorm:"column:operate_way"`
	Status       enumor.AuditStatus       `db:"status" json:"status" gorm:"column:status"`
	StrategyId   uint32                   `db:"strategy_id" json:"strategy_id" gorm:"column:strategy_id"`
	IsCompare    bool                     `db:"is_compare" json:"is_compare" gorm:"column:is_compare"`
}

// TableName is the audit's database table name.
func (a *Audit) TableName() Name {
	return "audits"
}

// AuditBasicDetail defines the audit's basic details.
type AuditBasicDetail struct {
	Prev    interface{} `json:"prev"`
	Changed interface{} `json:"changed"`
}

// AuditField defines the audit's basic field
type AuditField struct {
	OperateWay       string
	Action           enumor.AuditAction
	ResourceInstance string
	Status           enumor.AuditStatus
	AppId            uint32
	StrategyId       uint32
	IsCompare        bool
}
