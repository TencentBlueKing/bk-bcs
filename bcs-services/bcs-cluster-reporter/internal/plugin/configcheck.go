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

// Package plugin xxx
package plugin

import (
	"fmt"
	"strconv"
	"strings"
)

// FloatFlag xxx
type FloatFlag struct {
	Name string
	// ge, le, eq, none
	CompareType string
	Value       float64
	Needed      bool
}

// CheckFlag check flags
func CheckFlag(flagList []string, floatFlag FloatFlag) string {
	checked := false
	for _, flagStr := range flagList {
		if strings.HasPrefix(flagStr, floatFlag.Name) {
			checked = true
			strList := strings.Split(flagStr, "=")
			if len(strList) == 2 {
				value, err := strconv.ParseFloat(strList[1], 64)
				if err != nil {
					return fmt.Sprintf("%s value is %s, not float64", floatFlag.Name, strList[1])
				}

				if floatFlag.CompareType == "ge" {
					if floatFlag.Value > value {
						return fmt.Sprintf(StringMap[CheckFlagLeDetailFormat], floatFlag.Name, strList[1], floatFlag.Value)
					}
				}

			} else {
				return fmt.Sprintf("%s value is %s, not float64", floatFlag.Name, flagStr)
			}

		}
	}

	if !checked && floatFlag.Needed {
		return fmt.Sprintf(StringMap[CheckFlagNotSetDetailFormat], floatFlag.Name, floatFlag.Value)
	}

	return ""
}
