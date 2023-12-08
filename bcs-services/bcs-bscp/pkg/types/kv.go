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
	"errors"

	"bscp.io/pkg/dal/table"
)

// UpsertKvOption is used to define options for inserting or updating key-value data.
type UpsertKvOption struct {
	BizID  uint32
	AppID  uint32
	Key    string
	Value  string
	KvType table.DataType
}

// Validate is used to validate the effectiveness of the UpsertKvOption structure.
func (o *UpsertKvOption) Validate() error {
	if o.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if o.AppID <= 0 {
		return errors.New("invalid app id, should >= 1")
	}

	if o.Key == "" {
		return errors.New("kv key is required")
	}

	if o.Value == "" {
		return errors.New("kv value is required")
	}

	if err := o.KvType.ValidateValue(o.Value); err != nil {
		return err
	}

	return nil
}

// GetLastKvOpt is used to define options for retrieving the last key-value data.
type GetLastKvOpt struct {
	BizID uint32
	AppID uint32
	Key   string
}

// Validate is used to validate the effectiveness of the GetLastKvOpt structure.
func (o *GetLastKvOpt) Validate() error {
	if o.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if o.AppID <= 0 {
		return errors.New("invalid app id, should >= 1")
	}

	if o.Key == "" {
		return errors.New("kv key is required")
	}
	return nil
}

// GetKvByVersion is used to define options for retrieving key-value data by version.
type GetKvByVersion struct {
	BizID   uint32
	AppID   uint32
	Key     string
	Version int
}

// Validate is used to validate the effectiveness of the GetKvByVersion structure.
func (o *GetKvByVersion) Validate() error {
	if o.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if o.AppID <= 0 {
		return errors.New("invalid app id, should >= 1")
	}

	if o.Key == "" {
		return errors.New("kv key is required")
	}

	if o.Version <= 0 {
		return errors.New("invalid version, should >= 1")
	}

	return nil
}

// DeleteKvOpt is used to define options for deleting key-value data.
type DeleteKvOpt struct {
	BizID uint32
	AppID uint32
	Key   string
}

// Validate is used to validate the effectiveness of the DeleteKvOpt structure.
func (o *DeleteKvOpt) Validate() error {
	if o.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if o.AppID <= 0 {
		return errors.New("invalid app id, should >= 1")
	}

	if o.Key == "" {
		return errors.New("kv key is required")
	}
	return nil
}

// ListKvOption defines options to list kv.
type ListKvOption struct {
	BizID     uint32    `json:"biz_id"`
	AppID     uint32    `json:"app_id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	SearchKey string    `json:"search_key"`
	All       bool      `json:"all"`
	Page      *BasePage `json:"page"`
	IDs       []uint32  `json:"ids"`
	KvType    bool      `json:"kv_type"`
}

// Validate is used to validate the effectiveness of the ListKvOption structure.
func (opt *ListKvOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}
	if opt.AppID <= 0 {
		return errors.New("invalid app id, should >= 1")
	}

	if opt.Page == nil {
		return errors.New("page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}
