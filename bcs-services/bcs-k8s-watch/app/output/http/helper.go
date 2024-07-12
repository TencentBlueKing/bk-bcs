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

package http

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
)

// parseSelectors parse selectors from string to map, use for storage api
func parseSelectors(selector string, prefix string) (map[string]string, error) {
	selectorMap := map[string]string{}
	if selector == "" {
		return selectorMap, nil
	}

	labelSet, err := labels.ConvertSelectorToLabelsMap(selector)
	if err != nil {
		return nil, err
	}
	labelMap := map[string]string(labelSet)

	for k, v := range labelMap {
		key := fmt.Sprintf(prefix + k)
		selectorMap[key] = v
	}
	return selectorMap, nil
}
