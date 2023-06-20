/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Uint32Slice []uint32

// Value implements the driver.Valuer interface
func (u Uint32Slice) Value() (driver.Value, error) {
	// Convert the []uint32 to a JSON-encoded string
	data, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan implements the sql.Scanner interface
func (u *Uint32Slice) Scan(value interface{}) error {
	// Check if the value is nil
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []uint8:
		// The value is of type []uint8 (MySQL driver representation for JSON columns)
		// Unmarshal the JSON-encoded value to []uint32
		err := json.Unmarshal(v, u)
		if err != nil {
			return err
		}
	case string:
		// The value is of type string (fallback for older versions of MySQL driver)
		// Unmarshal the JSON-encoded value to []uint32
		err := json.Unmarshal([]byte(v), u)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported Scan type for Uint32Slice")
	}

	return nil
}
