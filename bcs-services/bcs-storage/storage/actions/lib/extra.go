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
 *
 */

package lib

import (
	"encoding/base64"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
)

// NewExtra create extra fields
func NewExtra(raw string) *ExtraField {
	return &ExtraField{raw: raw}
}

// ExtraField extra field
type ExtraField struct {
	raw string
	str []byte
}

func (ef *ExtraField) decode() error {
	r, err := base64.StdEncoding.DecodeString(ef.raw)
	if err != nil {
		return err
	}
	ef.str = r
	return nil
}

// GetStr get string value
func (ef *ExtraField) GetStr() (string, error) {
	if ef.str != nil {
		return string(ef.str), nil
	}
	err := ef.decode()
	return string(ef.str), err
}

// Unmarshal unmarshal extra field to struct
func (ef *ExtraField) Unmarshal(r interface{}) (err error) {
	if err = ef.decode(); err != nil {
		return
	}
	return codec.DecJson(ef.str, r)
}
