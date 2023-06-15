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

package types

import (
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/runtime/filter"
)

// ListInstancesOption list instance options.
type ListInstancesOption struct {
	BizID     uint32             `json:"biz_id"`
	TableName table.Name         `json:"table_name"`
	Filter    *filter.Expression `json:"filter"`
	Page      *BasePage          `json:"page"`
}

// Validate list instance options.
func (o *ListInstancesOption) Validate(po *PageOption) error {
	if o.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if len(o.TableName) == 0 {
		return errf.New(errf.InvalidParameter, "table name is required")
	}

	if o.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is required")
	}

	if o.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}

	if err := o.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// InstanceResource define list instances result.
type InstanceResource struct {
	ID          uint32 `db:"id" json:"id"`
	DisplayName string `db:"name" json:"display_name"`
}

// ListInstanceDetails defines the response details of requested ListInstancesOption.
type ListInstanceDetails struct {
	Count   uint32              `json:"count"`
	Details []*InstanceResource `json:"details"`
}
