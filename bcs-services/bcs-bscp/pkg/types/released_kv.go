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

import "errors"

// CreateReleasedKvOption defines options to create released kv.
type CreateReleasedKvOption struct {
	BizID     uint32
	AppID     uint32
	ReleaseID uint32
	Key       string
	Value     string
	KvType    KvType
}

// Validate is used to validate the effectiveness of the CreateReleasedKvOption structure.
func (o *CreateReleasedKvOption) Validate() error {
	if o.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if o.AppID <= 0 {
		return errors.New("invalid app id, should >= 1")
	}

	if o.ReleaseID <= 0 {
		return errors.New("invalid revision id, should >= 1")
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

// GetRKvOption ..
type GetRKvOption struct {
	BizID      uint32
	AppID      uint32
	Key        string
	ReleasedID uint32
	Version    int
}

// Validate is used to validate the effectiveness of the GetKvByVersion structure.
func (o *GetRKvOption) Validate() error {
	if o.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if o.AppID <= 0 {
		return errors.New("invalid app id, should >= 1")
	}

	if o.Key == "" {
		return errors.New("kv key is required")
	}

	if o.ReleasedID <= 0 {
		return errors.New("invalid revision id, should >= 1")
	}

	if o.Version <= 0 {
		return errors.New("invalid version, should >= 1")
	}

	return nil
}

// ListRKvOption defines options to list released kv.
type ListRKvOption struct {
	BizID     uint32    `json:"biz_id"`
	AppID     uint32    `json:"app_id"`
	Key       string    `json:"key"`
	ReleaseID uint32    `json:"release_id"`
	SearchKey string    `json:"search_key"`
	All       bool      `json:"all"`
	Page      *BasePage `json:"page"`
}

// Validate is used to validate the effectiveness of the ListKvOption structure.
func (opt *ListRKvOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}
	if opt.AppID <= 0 {
		return errors.New("invalid app id, should >= 1")
	}
	if opt.ReleaseID <= 0 {
		return errors.New("invalid Release id, should >= 1")
	}

	if opt.Page == nil {
		return errors.New("page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}
