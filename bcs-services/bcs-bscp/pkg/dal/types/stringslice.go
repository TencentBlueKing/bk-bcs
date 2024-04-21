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

// Package types is for dal types
package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringSlice is []string type for dal
type StringSlice []string

// Value implements the driver.Valuer interface
// See gorm document about customizing data types: https://gorm.io/docs/data_types.html
func (u StringSlice) Value() (driver.Value, error) {
	// Convert the StringSlice to a JSON-encoded string
	data, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan implements the sql.Scanner interface
// See gorm document about customizing data types: https://gorm.io/docs/data_types.html
func (u *StringSlice) Scan(value interface{}) error {
	// Check if the value is nil
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// The value is of type []byte (MySQL driver representation for JSON columns)
		// Unmarshal the JSON-encoded value to StringSlice
		err := json.Unmarshal(v, u)
		if err != nil {
			return err
		}
	case string:
		// The value is of type string (fallback for older versions of MySQL driver)
		// Unmarshal the JSON-encoded value to StringSlice
		err := json.Unmarshal([]byte(v), u)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported Scan type for StringSlice")
	}

	return nil
}
