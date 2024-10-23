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

// Package types NOTES
package types

import (
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/filter"
)

// ListAuditsOption defines options to list audit.
type ListAuditsOption struct {
	BizID  uint32             `json:"biz_id"`
	Filter *filter.Expression `json:"filter"`
	Page   *BasePage          `json:"page"`
}

// Validate the list audit options
func (lao *ListAuditsOption) Validate(po *PageOption) error {
	if lao.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if lao.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is nil")
	}

	exprOpt := &filter.ExprOption{
		// remove biz_id because it's a required field in the option.
		RuleFields: table.AuditColumns.WithoutColumn("biz_id"),
	}
	if err := lao.Filter.Validate(exprOpt); err != nil {
		return err
	}

	if lao.Page == nil {
		return errf.New(errf.InvalidParameter, "page is null")
	}

	if err := lao.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// ListAuditDetails defines the response details of requested ListAuditsOption
type ListAuditDetails struct {
	Count   uint32         `json:"count"`
	Details []*table.Audit `json:"details"`
}

// ListAuditsAppStrategy defines the model of audits app strategy
type ListAuditsAppStrategy struct {
	// Name is application's name
	App      AppPart      `json:"app" gorm:"embedded"`
	Audit    AuditPart    `json:"audit" gorm:"embedded"`
	Strategy StrategyPart `json:"strategy" gorm:"embedded"`
}

// AppPart app field
type AppPart struct {
	Name    string `json:"name" gorm:"column:name"`
	Creator string `json:"creator" gorm:"column:creator"`
}

// AuditPart audit field
type AuditPart struct {
	// Audit is used to save resource's audit information.
	ID           uint32    `db:"id" json:"id" gorm:"primaryKey"`
	BizID        uint32    `db:"biz_id" json:"biz_id" gorm:"column:biz_id"`
	AppID        uint32    `db:"app_id" json:"app_id" gorm:"column:app_id"`
	ResourceType string    `db:"res_type" json:"resource_type" gorm:"column:res_type"`
	ResourceID   uint32    `db:"res_id" json:"resource_id" gorm:"column:res_id"`
	Action       string    `db:"action" json:"action" gorm:"column:action"`
	Operator     string    `db:"operator" json:"operator" gorm:"column:operator"`
	CreatedAt    time.Time `db:"created_at" json:"created_at" gorm:"column:created_at"`
	ResInstance  string    `db:"res_instance" json:"res_instance" gorm:"column:res_instance"`
	OperateWay   string    `db:"operate_way" json:"operate_way" gorm:"column:operate_way"`
	Status       string    `db:"status" json:"status" gorm:"column:status"`
	IsCompare    bool      `db:"is_compare" json:"is_compare" gorm:"column:is_compare"`
}

// StrategyPart defines strategy fields
type StrategyPart struct {
	PublishType      string      `db:"publish_type" json:"publish_type" gorm:"column:publish_type"`
	PublishTime      string      `db:"publish_time" json:"publish_time" gorm:"column:publish_time"`
	PublishStatus    string      `db:"publish_status" json:"publish_status" gorm:"column:publish_status"`
	RejectReason     string      `db:"reject_reason" json:"reject_reason" gorm:"column:reject_reason"`
	Approver         string      `db:"approver" json:"approver" approver:"column:approver"`
	ApproverProgress string      `db:"approver_progress" json:"approver_progress" gorm:"column:approver_progress"`
	UpdatedAt        time.Time   `db:"updated_at" json:"updated_at" gorm:"column:updated_at"`
	Reviser          string      `db:"reviser" json:"reviser" gorm:"column:reviser"`
	Creator          string      `db:"creator" json:"creator" gorm:"column:creator"`
	ReleaseId        uint32      `db:"reviser" json:"release_id" gorm:"column:release_id"`
	Scope            table.Scope `db:"scope" json:"scope" gorm:"column:scope"`
}
