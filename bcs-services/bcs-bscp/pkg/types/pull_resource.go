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

package types

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
)

// ListInstancesOption list instance options.
type ListInstancesOption struct {
	ResourceType string    `json:"resource_type"`
	ParentType   string    `json:"parent_type"`
	ParentID     string    `json:"parent_id"`
	Page         *BasePage `json:"page"`
}

// Validate list instance options.
func (o *ListInstancesOption) Validate(po *PageOption) error {

	if o.ResourceType == "" {
		return errf.New(errf.InvalidParameter, "resource type is required")
	}

	if o.ParentType != "" && o.ParentID == "" {
		return errf.New(errf.InvalidParameter, "parent id is required")
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
	ID   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

// ListInstanceDetails defines the response details of requested ListInstancesOption.
type ListInstanceDetails struct {
	Count   uint32              `json:"count"`
	Details []*InstanceResource `json:"details"`
}

// FetchInstanceInfoOption fetch instance info options.
type FetchInstanceInfoOption struct {
	ResourceType string   `json:"resource_type"`
	IDs          []string `json:"ids"`
}

// Validate fetch instance info options.
func (o *FetchInstanceInfoOption) Validate() error {

	if o.ResourceType == "" {
		return errf.New(errf.InvalidParameter, "resource type is required")
	}

	if len(o.IDs) == 0 {
		return errf.New(errf.InvalidParameter, "ids is required")
	}

	return nil
}

// InstanceInfo define fetch instance info result.
type InstanceInfo struct {
	ID          string   `db:"id" json:"id"`
	DisplayName string   `db:"display_name" json:"display_name"`
	Approver    []string `db:"approver" json:"approver"`
	Path        []string `db:"path" json:"path"`
}

// FetchInstanceInfoDetails defines the response details of requested ListInstancesOption.
type FetchInstanceInfoDetails struct {
	Details []*InstanceInfo `json:"details"`
}
