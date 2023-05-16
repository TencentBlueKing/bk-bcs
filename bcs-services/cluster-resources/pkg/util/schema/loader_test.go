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
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

var schemaMap4LoaderTest = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"number_key": 1,
		"float_key":  1.2,
	},
}

var exceptLoadRet = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"number_key": json.Number("1"),
		"float_key":  json.Number("1.2"),
	},
}

func TestGoLoader(t *testing.T) {
	loader := NewGoLoader(schemaMap4LoaderTest)
	schema, err := loader.Load()
	assert.Nil(t, err)
	assert.Equal(t, exceptLoadRet, schema)
}

var yamlFileContent4LoaderTest = `
type: object
properties:
  number_key: 1
  float_key: 1.2
`

func TestYmlFileLoader(t *testing.T) {
	tmpFile, _ := ioutil.TempFile("", "*.yml")
	_, _ = tmpFile.Write([]byte(yamlFileContent4LoaderTest))

	loader := NewFileLoader(tmpFile.Name())
	schema, err := loader.Load()
	assert.Nil(t, err)
	assert.Equal(t, exceptLoadRet, schema)
}

func TestYamlFileLoader(t *testing.T) {
	tmpFile, _ := ioutil.TempFile("", "*.yaml")
	_, _ = tmpFile.Write([]byte(yamlFileContent4LoaderTest))

	loader := NewFileLoader(tmpFile.Name())
	schema, err := loader.Load()
	assert.Nil(t, err)
	assert.Equal(t, exceptLoadRet, schema)
}

var jsonFileContent4LoaderTest = `
{
  "type": "object",
  "properties": {
    "number_key": 1,
    "float_key": 1.2
  }
}
`

func TestJSONFileLoader(t *testing.T) {
	tmpFile, _ := ioutil.TempFile("", "*.json")
	_, _ = tmpFile.Write([]byte(jsonFileContent4LoaderTest))

	loader := NewFileLoader(tmpFile.Name())
	schema, err := loader.Load()
	assert.Nil(t, err)
	assert.Equal(t, exceptLoadRet, schema)
}
