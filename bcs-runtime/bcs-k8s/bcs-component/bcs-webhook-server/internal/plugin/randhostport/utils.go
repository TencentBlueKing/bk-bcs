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

package randhostport

import (
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

func getHostPortByLabels(labels map[string]string) []uint64 {
	var retValues []uint64
	for k, v := range labels {
		if strings.HasSuffix(k, podHostportLabelSuffix) {
			hostportValue, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				blog.Warnf("cannot parse hostport from label %s/%s", k, v)
				continue
			}
			retValues = append(retValues, hostportValue)
		}
	}
	return retValues
}

func getPortStringsFromPodAnnotations(annotation map[string]string) []string {
	portStrs, ok := annotation[pluginPortsAnnotationKey]
	if !ok {
		return nil
	}
	var retStrs []string
	strs := strings.Split(portStrs, ",")
	for _, str := range strs {
		rawStr := strings.TrimSpace(str)
		if len(rawStr) == 0 {
			continue
		}
		retStrs = append(retStrs, rawStr)
	}
	return retStrs
}
