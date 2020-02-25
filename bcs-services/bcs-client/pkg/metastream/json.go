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

package metastream

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"

	mesostype "bk-bcs/bcs-common/common/types"
)

/*
multiple json format must like below:
---
{jsonObject}
---
{jsonObject}
---

multiple yaml format must like below:
---
{yamlObject}
---
{yamlObject}

*/

const (
	//JSONFormat json list detail for content
	JSONFormat = "json"
	//YAMLFormat yaml list detail for content
	YAMLFormat = "yaml"
)

type metaObject struct {
	mesostype.ObjectMeta `json:"metadata,omitempty"`
}

//NewJSONStream create stream implementation
func NewMetaStream(r io.Reader, ft string) Stream {
	allDatas, err := ioutil.ReadAll(r)
	if err != nil || len(allDatas) == 0 {
		return &jsonArray{}
	}
	rawList := strings.Split(string(allDatas), "---\n")
	//clean empty line
	var clearList []string
	for _, line := range rawList {
		newLine := strings.Trim(line, " \n")
		//line has apiVersion & kind inforamtion at least
		if len(newLine) > 20 {
			//convert format from yaml to json
			if YAMLFormat == ft {
				newJSON, err := yaml.YAMLToJSON([]byte(newLine))
				if err != nil {
					fmt.Printf("yaml convert err, %s\n", err.Error())
					continue
				}
				newLine = string(newJSON)
				//fmt.Printf("original yaml convert to json: %s\n", newLine)
			}
			clearList = append(clearList, newLine)
		}
	}
	js := &jsonArray{
		rawDatas: clearList,
		index:    0,
	}
	return js
}

//jsonArray implementation for Stream
type jsonArray struct {
	index        int
	rawDatas     []string
	indexRawJson string
}

//Length check if stream has Next JSON data
func (js *jsonArray) Length() int {
	return len(js.rawDatas)
}

//HasNext check if stream has Next JSON data
func (js *jsonArray) HasNext() bool {
	if js.index >= len(js.rawDatas) {
		return false
	}
	js.indexRawJson = js.rawDatas[js.index]
	js.index++
	return true
}

//GetResourceKind return apiVersion and Kind
func (js *jsonArray) GetResourceKind() (string, string, error) {
	meta := &mesostype.TypeMeta{}
	err := json.Unmarshal([]byte(js.indexRawJson), meta)
	if err != nil {
		return "", "", fmt.Errorf("json %d object err: %s", js.index-1, err.Error())
	}
	if len(meta.APIVersion) == 0 || len(meta.Kind) == 0 {
		return "", "", fmt.Errorf("json %d lost apiVersion or kind", js.index-1)
	}
	return meta.APIVersion, string(meta.Kind), nil
}

//GetResourceKey return JSON object index: namespace & name
func (js *jsonArray) GetResourceKey() (string, string, error) {
	objMeta := &metaObject{}
	err := json.Unmarshal([]byte(js.indexRawJson), objMeta)
	if err != nil {
		return "", "", fmt.Errorf("json %d object err: %s", js.index-1, err.Error())
	}
	// if len(objMeta.NameSpace) == 0 {
	// 	objMeta.NameSpace = "default"
	// }
	if len(objMeta.Name) == 0 {
		return "", "", fmt.Errorf("json %d lost meta.name", js.index-1)
	}
	return objMeta.NameSpace, objMeta.Name, nil
}

//GetRawJSON return  detail raw json string
func (js *jsonArray) GetRawJSON() []byte {
	if len(js.indexRawJson) == 0 {
		return nil
	}
	return []byte(js.indexRawJson)
}
