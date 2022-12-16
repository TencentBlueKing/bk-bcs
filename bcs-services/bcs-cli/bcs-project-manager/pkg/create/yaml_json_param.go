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
 *
 */

package create

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

// GenerateStruct 把模板文件生成结构体
// filename 文件地址
func GenerateStruct(filename string) (interface{}, error) {
	var requestParam GenerateRequestParam
	var p interface{}
	// 读取文件
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("file open failed: %v", err)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		requestParam = &jsonToRequestParam{}
	case ".yaml":
		requestParam = &yamlToRequestParam{}
	default:
		return nil, fmt.Errorf("wrong format")
	}
	generate, err := requestParam.generate(file, p)
	if err != nil {
		return nil, err
	}
	return generate, nil
}

// GenerateRequestParam 生成请求参数接口
type GenerateRequestParam interface {
	generate(data []byte, param interface{}) (interface{}, error)
}

type jsonToRequestParam struct{}

func (j *jsonToRequestParam) generate(data []byte, param interface{}) (interface{}, error) {
	err := json.Unmarshal(data, &param)
	if err != nil {
		return nil, fmt.Errorf("[jsonToRequestParam] deserialize failed: %v", err)
	}
	return param, nil
}

type yamlToRequestParam struct{}

func (y *yamlToRequestParam) generate(data []byte, param interface{}) (interface{}, error) {
	createdJson, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("json to yaml failed: %v", err)
	}
	err = json.Unmarshal(createdJson, &param)
	if err != nil {
		return nil, fmt.Errorf("[yamlToRequestParam] deserialize failed: %v", err)
	}
	return param, nil
}
