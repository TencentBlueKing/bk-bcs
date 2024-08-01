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
	"unicode/utf8"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// ValidateMemo validate bscp resource memo's length and format.
func ValidateMemo(kit *kit.Kit, memo string, required bool) error {
	// check data is nil and required.
	if required && len(memo) == 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "memo is required, can not be empty"))
	}

	if required {
		if len(memo) == 0 {
			return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "memo is required, can not be empty"))
		}
	} else {
		if len(memo) == 0 {
			return nil
		}
	}

	charLength := utf8.RuneCountInString(memo)
	if charLength > 200 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid memo, length should <= 200"))
	}

	return nil
}
