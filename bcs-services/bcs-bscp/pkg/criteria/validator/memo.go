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

package validator

import (
	"errors"
	"regexp"
)

const (
	// qualifiedMemoFmt bscp resource's memo format.
	qualifiedMemoFmt string = "(" + qMemoFmt + qExtMemoFmt + "*)?" + qMemoFmt
	qMemoFmt         string = "[\u4E00-\u9FA5A-Za-z0-9]"
	qExtMemoFmt      string = "[\u4E00-\u9FA5A-Za-z0-9-_\\s]"
)

// qualifiedMemoRegexp bscp resource's memo regexp.
var qualifiedMemoRegexp = regexp.MustCompile("^" + qualifiedMemoFmt + "$")

// ValidateMemo validate bscp resource memo's length and format.
func ValidateMemo(memo string, required bool) error {
	// check data is nil and required.
	if required && len(memo) == 0 {
		return errors.New("memo is required, can not be empty")
	}

	if required {
		if len(memo) == 0 {
			return errors.New("memo is required, can not be empty")
		}
	} else {
		if len(memo) == 0 {
			return nil
		}
	}

	if len(memo) > 256 {
		return errors.New("invalid memo, length should <= 256")
	}

	if !qualifiedMemoRegexp.MatchString(memo) {
		return errors.New("invalid memo, only allows include chinese、english、numbers、underscore (_)" +
			"、hyphen (-)、space, and must start and end with an chinese、english、numbers")
	}

	return nil
}
