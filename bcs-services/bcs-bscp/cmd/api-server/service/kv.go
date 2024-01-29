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

package service

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gopkg.in/yaml.v3"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

type kvService struct {
	authorizer auth.Authorizer
	cfgClient  pbcs.ConfigClient
}

func newKvService(authorizer auth.Authorizer,
	cfgClient pbcs.ConfigClient) *kvService {
	s := &kvService{
		authorizer: authorizer,
		cfgClient:  cfgClient,
	}
	return s
}

// Import is used to handle file import requests.
func (m *kvService) Import(w http.ResponseWriter, r *http.Request) {

	kt := kit.MustGetKit(r.Context())

	appIdStr := chi.URLParam(r, "app_id")
	appId, _ := strconv.Atoi(appIdStr)
	if appId == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("validation parameter fail")))
		return
	}
	reader := bufio.NewReader(r.Body)

	bytes, err := io.ReadAll(reader)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	var kvMap map[string]interface{}
	switch {
	case readJSONFile(bytes, &kvMap):
	case readYAMLFile(bytes, &kvMap):
	default:
		_ = render.Render(w, r, rest.BadRequest(errors.New("unsupported file type")))
		return
	}

	kvs, err := handleKv(kvMap)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	req := &pbcs.BatchUpsertKvsReq{
		BizId: kt.BizID,
		AppId: uint32(appId),
		Kvs:   kvs,
	}

	resp, err := m.cfgClient.BatchUpsertKvs(kt.RpcCtx(), req)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	_ = render.Render(w, r, rest.OKRender(resp))
}

func handleKv(result map[string]interface{}) ([]*pbcs.BatchUpsertKvsReq_Kv, error) {
	kvMap := []*pbcs.BatchUpsertKvsReq_Kv{}
	for key, value := range result {
		KVType := ""
		entry, ok := value.(map[string]interface{})
		if !ok {
			// 判断是不是数值类型
			if isNumber(value) {
				kvMap = append(kvMap, &pbcs.BatchUpsertKvsReq_Kv{
					Key:    key,
					Value:  fmt.Sprintf("%v", value),
					KvType: string(table.KvNumber),
				})
			} else {
				KVType = determineType(value.(string))
				kvMap = append(kvMap, &pbcs.BatchUpsertKvsReq_Kv{
					Key:    key,
					Value:  fmt.Sprintf("%v", value),
					KvType: KVType,
				})
			}
		} else {
			kvType, okType := entry["kv_type"].(string)
			kvValue, okVal := entry["value"]
			if !okType && !okVal {
				return nil, fmt.Errorf("file format error, please check the key: %s", key)
			}
			if err := validateKvType(kvType); err != nil {
				return nil, fmt.Errorf("key: %s %s", key, err.Error())
			}
			var val string
			val = fmt.Sprintf("%v", kvValue)
			// json 和 yaml 都需要格式化
			if kvType == string(table.KvJson) {
				v, ok := kvValue.(string)
				if !ok {
					return nil, fmt.Errorf("key: %s format error", key)
				}
				mv, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("key: %s json marshal error", key)
				}
				val = string(mv)
			} else if kvType == string(table.KvYAML) {
				_, ok := kvValue.(string)
				if !ok {
					ys, err := yaml.Marshal(kvValue)
					if err != nil {
						return nil, fmt.Errorf("key: %s yaml marshal error", key)
					}
					val = string(ys)
				}
			}
			kvMap = append(kvMap, &pbcs.BatchUpsertKvsReq_Kv{
				Key:    key,
				Value:  val,
				KvType: kvType,
			})
		}
	}
	return kvMap, nil
}

func validateKvType(kvType string) error {
	switch kvType {
	case string(table.KvStr):
	case string(table.KvNumber):
	case string(table.KvText):
	case string(table.KvJson):
	case string(table.KvYAML):
	case string(table.KvXml):
	default:
		return errors.New("invalid data-type")
	}
	return nil
}

// 读取json文件
func readJSONFile(bytes []byte, result *map[string]interface{}) bool {
	return !json.Valid(bytes) && json.Unmarshal(bytes, &result) == nil
}

// 读取yaml文件
func readYAMLFile(bytes []byte, result *map[string]interface{}) bool {
	return yaml.Unmarshal(bytes, &result) == nil
}

// 根据值判断类型
func determineType(value string) string {
	var result string
	switch {
	case isJSON(value):
		result = string(table.KvJson)
	case isYAML(value):
		result = string(table.KvYAML)
	case isXML(value):
		result = string(table.KvXml)
	case isTEXT(value):
		result = string(table.KvText)
	case isNumber(value):
		result = string(table.KvNumber)
	default:
		result = string(table.KvStr)
	}
	return result
}

// 判断是否为 JSON
func isJSON(value string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(value), &js) == nil
}

// 判断是否为 YAML
func isYAML(value string) bool {
	var yml map[string]interface{}
	return yaml.Unmarshal([]byte(value), &yml) == nil
}

// 判断是否为 XML
func isXML(value string) bool {
	var xmlData interface{}
	return xml.Unmarshal([]byte(value), &xmlData) == nil
}

// 判断是否为 TEXT
func isTEXT(value string) bool {
	return strings.Contains(value, "\n")
}

// 判断是不是 Number
func isNumber(value interface{}) bool {
	// 获取值的类型
	valType := reflect.TypeOf(value)

	// 检查类型是否为数字
	switch valType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
