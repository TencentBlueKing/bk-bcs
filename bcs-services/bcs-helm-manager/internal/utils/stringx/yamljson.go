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

package stringx

import (
	"gopkg.in/yaml.v2"
)

// Yaml2Json convert yaml to json
func Yaml2Json(y string) (map[interface{}]interface{}, error) {
	var j map[interface{}]interface{}
	if err := yaml.Unmarshal([]byte(y), &j); err != nil {
		return nil, err
	}
	return j, nil
}

// Json2Yaml convert json to yaml
func Json2Yaml(j map[interface{}]interface{}) ([]byte, error) {
	y, err := yaml.Marshal(j)
	if err != nil {
		return nil, err
	}
	return y, nil
}
