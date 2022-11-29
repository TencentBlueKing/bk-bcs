/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package schema

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// JSONLoader jsonSchema 加载器
type JSONLoader interface {
	Source() interface{}
	Load() (interface{}, error)
}

type jsonGoLoader struct {
	source interface{}
}

func (l *jsonGoLoader) Source() interface{} {
	return l.source
}

func (l *jsonGoLoader) Load() (interface{}, error) {
	jsonBytes, err := json.Marshal(l.Source())
	if err != nil {
		return nil, err
	}
	return decodeJSONUsingNumber(bytes.NewReader(jsonBytes))
}

// NewGoLoader 根据提供的 Go 结构体生成 JSONLoader
func NewGoLoader(source interface{}) JSONLoader {
	return &jsonGoLoader{source: source}
}

const (
	// jsonFileLoader 支持的文件扩展名
	ymlFileNameExt  = ".yml"
	yamlFileNameExt = ".yaml"
	jsonFileNameExt = ".json"
)

// 受支持的文件扩展名
var supportedFileNameExts = []string{ymlFileNameExt, yamlFileNameExt, jsonFileNameExt}

// 不受支持的文件格式
var errFormatUnsupported = errors.New("file format is unsupported")

type jsonFileLoader struct {
	source string
}

func (l *jsonFileLoader) Source() interface{} {
	return l.source
}

func (l *jsonFileLoader) Load() (interface{}, error) {
	fileNameExt := filepath.Ext(l.source)
	if !slice.StringInSlice(fileNameExt, supportedFileNameExts) {
		return nil, errFormatUnsupported
	}

	content, err := ioutil.ReadFile(l.source)
	if err != nil {
		return nil, err
	}

	jsonBytes := content
	// yaml 类型文件转 json 格式
	if fileNameExt == ymlFileNameExt || fileNameExt == yamlFileNameExt {
		goMap := map[string]interface{}{}
		if err = yaml.Unmarshal(content, &goMap); err != nil {
			return nil, err
		}
		if jsonBytes, err = json.Marshal(goMap); err != nil {
			return nil, err
		}
	}

	return decodeJSONUsingNumber(bytes.NewReader(jsonBytes))
}

// NewFileLoader 根据提供的文件路径生成 JSONLoader
func NewFileLoader(source string) JSONLoader {
	return &jsonFileLoader{source: source}
}

func decodeJSONUsingNumber(r io.Reader) (interface{}, error) {
	var document interface{}
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	return document, nil
}
