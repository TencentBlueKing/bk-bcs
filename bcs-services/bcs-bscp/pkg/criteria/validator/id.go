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

package validator

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	qualifiedUidFmt   string = "(" + qnameExtUidFmt + qnameExtUidSuffix + "*)?" + qnameExtUidFmt // nolint goconst
	qnameExtUidFmt    string = "[A-Za-z0-9]"
	qnameExtUidSuffix string = "[A-Za-z0-9-_:]"
)

var qualifiedUidRegexp = regexp.MustCompile("^" + qualifiedUidFmt + "$")

// ValidateUid is to validate the instance's unique id is valid or not
func ValidateUid(uid string) error {
	if len(uid) == 0 {
		return errors.New("invalid uid, length should >= 1")
	}

	if len(uid) > 64 {
		return errors.New("invalid uid, length should <= 64")
	}

	if !qualifiedUidRegexp.MatchString(uid) {
		return fmt.Errorf("invalid uid format, should match regexp: %s", qualifiedUidFmt)
	}

	return nil
}

// ValidateUidLength is to validate the length of app instance's uid is valid or not
func ValidateUidLength(uid string) error {
	if len(uid) == 0 {
		return errors.New("uid is empty")
	}

	if len(uid) > 64 {
		return errors.New("invalid uid, length should <= 64")
	}

	return nil
}
