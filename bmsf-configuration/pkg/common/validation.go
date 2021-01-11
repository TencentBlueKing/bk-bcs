/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"fmt"
)

// ValidateString validates string field base on min/max length limit.
func ValidateString(name, data string, minLen, maxLen int) error {
	length := len(data)

	if minLen > 0 && length == 0 {
		return fmt.Errorf("invalid input data, %s is required", name)
	}

	if length < minLen {
		return fmt.Errorf("invalid input data, %s is too short (min length: %d)", name, minLen)
	}

	if length > maxLen {
		return fmt.Errorf("invalid input data, %s is too long (max length: %d)", name, maxLen)
	}

	return nil
}

// ValidateInt validates int field base on min/max value limit.
func ValidateInt(name string, data, minValue, maxValue int) error {
	if data < minValue {
		return fmt.Errorf("invalid input data, %s is too little (min value: %d)", name, minValue)
	}

	if data > maxValue {
		return fmt.Errorf("invalid input data, %s is too large (max value: %d)", name, maxValue)
	}

	return nil
}

// ValidateInt32 validates int32 field base on min/max value limit.
func ValidateInt32(name string, data, minValue, maxValue int32) error {
	if data < minValue {
		return fmt.Errorf("invalid input data, %s is too little (min value: %d)", name, minValue)
	}

	if data > maxValue {
		return fmt.Errorf("invalid input data, %s is too large (max value: %d)", name, maxValue)
	}

	return nil
}

// ValidateInt64 validates int64 field base on min/max value limit.
func ValidateInt64(name string, data, minValue, maxValue int64) error {
	if data < minValue {
		return fmt.Errorf("invalid input data, %s is too little (min value: %d)", name, minValue)
	}

	if data > maxValue {
		return fmt.Errorf("invalid input data, %s is too large (max value: %d)", name, maxValue)
	}

	return nil
}

// ValidateUint validates uint field base on min/max value limit.
func ValidateUint(name string, data, minValue, maxValue uint) error {
	if data < minValue {
		return fmt.Errorf("invalid input data, %s is too little (min value: %d)", name, minValue)
	}

	if data > maxValue {
		return fmt.Errorf("invalid input data, %s is too large (max value: %d)", name, maxValue)
	}

	return nil
}

// ValidateUint32 validates uint32 field base on min/max value limit.
func ValidateUint32(name string, data, minValue, maxValue uint32) error {
	if data < minValue {
		return fmt.Errorf("invalid input data, %s is too little (min value: %d)", name, minValue)
	}

	if data > maxValue {
		return fmt.Errorf("invalid input data, %s is too large (max value: %d)", name, maxValue)
	}

	return nil
}

// ValidateUint64 validates uint64 field base on min/max value limit.
func ValidateUint64(name string, data, minValue, maxValue uint64) error {
	if data < minValue {
		return fmt.Errorf("invalid input data, %s is too little (min value: %d)", name, minValue)
	}

	if data > maxValue {
		return fmt.Errorf("invalid input data, %s is too large (max value: %d)", name, maxValue)
	}

	return nil
}

// ValidateStrings validates string field must in target slice base on values limit.
func ValidateStrings(name, data string, values ...string) error {
	for _, value := range values {
		if data == value {
			return nil
		}
	}
	return fmt.Errorf("invalid input data, %s is not supported (values: %s)", name, values)
}

// ValidateEnums validates enum field must in target slice base on values limit.
func ValidateEnums(name string, data interface{}, values ...interface{}) error {
	for _, value := range values {
		if data == value {
			return nil
		}
	}
	return fmt.Errorf("invalid input data, %s is not supported (values: %s)", name, values)
}
