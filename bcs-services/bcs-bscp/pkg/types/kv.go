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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// KvType is the type of kv
type KvType string

const (
	// KvStr is the type for string kv
	KvStr KvType = "string"
	// KvNumber is the type for number kv
	KvNumber KvType = "number"
	// KvText is the type for text kv
	KvText KvType = "text"
	// KvJson is the type for json kv
	KvJson KvType = "json"
	// KvYAML is the type for yaml kv
	KvYAML KvType = "yaml"
)

func (k KvType) Validate(value string) error {

	if value == "" {
		return errors.New("kv value is null")
	}

	switch k {
	case KvStr:
		return nil
	case KvNumber:
		if isStringConvertibleToNumber(value) {
			return fmt.Errorf("value is not a number")
		}
		return nil
	case KvText:
		return nil
	case KvJson:
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("value is not a json")
		}
		return nil
	case KvYAML:
		var data interface{}
		if err := yaml.Unmarshal([]byte(value), &data); err != nil {
			return fmt.Errorf("value is not a yaml, err: %v", err)
		}
		return nil
	default:
		return errors.New("revision not set")
	}
}

func isStringConvertibleToNumber(s string) bool {
	_, err := strconv.Atoi(s)
	if err == nil {
		return true
	}

	_, err = strconv.ParseFloat(s, 64)
	return err == nil

}

// UpsertKvOption ...
type UpsertKvOption struct {
	BizID  uint32
	AppID  uint32
	Key    string
	Value  string
	KvType KvType
}

// Validate ...
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

	if err := o.KvType.Validate(o.Value); err != nil {
		return err
	}

	return nil
}

// GetLastKvOpt ...
type GetLastKvOpt struct {
	BizID uint32
	AppID uint32
	Key   string
}

// Validate ...
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
