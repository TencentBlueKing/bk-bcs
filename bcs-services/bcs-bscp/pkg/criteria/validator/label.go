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

package validator

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// MaxLabelKeyCount is the max number of labels.
	MaxLabelKeyCount = 5
	// MaxLabelKeyLength defines the maximum value of a label key's length
	MaxLabelKeyLength = 128
	// MaxLabelValueLength defines the maximum value of a label value's length
	MaxLabelValueLength = 128
)

// ReservedLabelKeyPrefix label key reserved prefix for future system expansion, not allow user create.
var ReservedLabelKeyPrefix = []string{"bk_bscp", "bscp"}

// ValidateLabel validate a label is valid or not
func ValidateLabel(label map[string]string) error {
	if len(label) == 0 {
		return nil
	}

	length := len(label)
	if length > MaxLabelKeyCount {
		return fmt.Errorf("label's max key number should be <= %d", MaxLabelKeyCount)
	}

	var err error
	for k, v := range label {
		err = ValidateLabelKey(k)
		if err != nil {
			return err
		}

		err = ValidateLabelValue(v)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateLabelKey validate if a label's key is valid or not
func ValidateLabelKey(key string) error {
	length := len(key)
	if length == 0 {
		return errors.New("label's key can not be empty")
	}

	if length > MaxLabelKeyLength {
		return fmt.Errorf("label's key length should <= %d", MaxLabelKeyLength)
	}

	for _, prefix := range ReservedLabelKeyPrefix {
		if strings.HasPrefix(strings.ToLower(key), prefix) {
			return fmt.Errorf("'%s' prefix is system reserved label, do not allow to use", prefix)
		}
	}

	return nil
}

// ValidateLabelValue validate if a label's value is valid or not when the value is an string.
func ValidateLabelValue(value string) error {
	length := len(value)
	if length == 0 {
		return errors.New("label's value can not be empty")
	}

	if length > MaxLabelValueLength {
		return fmt.Errorf("label's value length should <= %d", MaxLabelValueLength)
	}

	return nil
}
