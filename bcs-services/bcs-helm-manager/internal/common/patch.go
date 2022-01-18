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
	"strings"
)

const (
	// patch templates const key format
	patchTemplateKeyPrefix = "__BCS_HELM_PATCH_"
	patchTemplateKeySuffix = "__"
)

// tk define the template key format
func tk(s string) string {
	return patchTemplateKeyPrefix + s + patchTemplateKeySuffix
}

// patch templates keys
var (
	TKProjectID    = tk("PROJECTID")
	TKClusterID    = tk("CLUSTERID")
	TKNamespace    = tk("NAMESPACE")
	TKCreator      = tk("CREATOR")
	TKUpdator      = tk("UPDATOR")
	TKVersion      = tk("VERSION")
	TKName         = tk("NAME")
	TKCustomLabels = tk("CUSTOM_LABELS")

	templateKeyRegex = regexp.MustCompile(patchTemplateKeyPrefix + ".+" + patchTemplateKeySuffix)
)

// IsTemplateKey check if the provided string is a template key
func IsTemplateKey(key string) bool {
	return strings.HasPrefix(key, patchTemplateKeyPrefix) && strings.HasSuffix(key, patchTemplateKeySuffix)
}

// EmptyAllTemplateKey find all template keys in source and replace them all with empty string
func EmptyAllTemplateKey(source []byte) []byte {
	return templateKeyRegex.ReplaceAll(source, nil)
}
