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

// Package values xxx
package values

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// MergeValues merge values
func MergeValues(values ...string) ([]string, error) {
	if len(values) == 0 {
		return []string{}, nil
	}

	// merge values, only last values will be shown at bcs frontend
	merged := make(map[interface{}]interface{})
	for _, value := range values {
		var m map[interface{}]interface{}
		err := yaml.Unmarshal([]byte(value), &m)
		if err != nil {
			return nil, fmt.Errorf("merge values failed: %v", err)
		}
		mergeMaps(merged, m)
	}

	resultValues, err := yaml.Marshal(&merged)
	if err != nil {
		return nil, fmt.Errorf("merge values failed when marshal yaml: %v", err)
	}

	return []string{string(resultValues)}, nil
}

// mergeMaps merge maps
func mergeMaps(dest, src map[interface{}]interface{}) {
	for k, v := range src {
		if mv, ok := v.(map[interface{}]interface{}); ok {
			if dv, ok := dest[k].(map[interface{}]interface{}); ok {
				mergeMaps(dv, mv)
				continue
			}
		}
		dest[k] = v
	}
}
