/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"regexp"
)

const (
	varTemplateKeyPrefix = "__BCS_HELM_VAR_"
	varTemplateKeySuffix = "__"
)

// Vtk define the var-template-key format
func Vtk(s string) string {
	return varTemplateKeyPrefix + s + varTemplateKeySuffix
}

// var templates keys
var (
	varTemplateKeyRegex = regexp.MustCompile(varTemplateKeyPrefix + ".+" + varTemplateKeySuffix)
)

// EmptyAllVarTemplateKey find all var template keys in source and replace them all with empty string
func EmptyAllVarTemplateKey(source []byte) []byte {
	return varTemplateKeyRegex.ReplaceAll(source, nil)
}
