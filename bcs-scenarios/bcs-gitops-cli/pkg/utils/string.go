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

package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/json2yaml"
	"gopkg.in/yaml.v3"
)

// TrimLeadAndTrailQuotes trim the lead and trail quotes. Trim only execute when lead and trail all have quotes
func TrimLeadAndTrailQuotes(str string) string {
	if strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"") {
		str = strings.TrimPrefix(str, "\"")
		str = strings.TrimSuffix(str, "\"")
		return str
	}
	if strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'") {
		str = strings.TrimPrefix(str, "'")
		str = strings.TrimSuffix(str, "'")
		return str
	}
	return str
}

// CheckStringJsonOrYaml check the request body is json or yaml
func CheckStringJsonOrYaml(body []byte) []byte {
	var jsonData map[string]interface{}
	var yamlData map[string]interface{}
	jsonErr := json.Unmarshal(body, &jsonData)
	yamlErr := yaml.Unmarshal(body, &yamlData)
	if jsonErr != nil && yamlErr != nil {
		ExitError("request body not json or yaml type")
	}
	if yamlErr == nil {
		var err error
		if body, err = json.Marshal(yamlData); err != nil {
			ExitError(fmt.Sprintf("yaml to json failed: %s", err.Error()))
		}
	}
	return body
}

// JsonToYaml transfer json to yaml
func JsonToYaml(jsonData []byte) []byte {
	var output strings.Builder
	input := strings.NewReader(string(jsonData))
	if err := json2yaml.Convert(&output, input); err != nil {
		ExitError(fmt.Sprintf("json to yaml failed: %s", err.Error()))
	}
	return []byte(output.String())
}

// YamlToJson transfer yaml to json
func YamlToJson(body []byte) []byte {
	var yamlData map[string]interface{}
	err := yaml.Unmarshal(body, &yamlData)
	if err != nil {
		ExitError(fmt.Sprintf("unmarshal yaml '%s' failed: %s", string(body), err.Error()))
	}
	body, err = json.Marshal(yamlData)
	if err != nil {
		ExitError(fmt.Sprintf("marshal json data failed: %s", err.Error()))
	}
	return body
}
