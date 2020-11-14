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

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"io/ioutil"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-cc-agent/config"
)

type PropertyFilter struct {
	Condition string `json:"condition"`
	Rules     []Rule `json:"rules"`
}

type Rule struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type APIResponse struct {
	Result  bool     `json:"result"`
	Code    int      `json:"code"`
	Data    RespData `json:"data"`
	Message string   `json:"message"`
}

type RespData struct {
	Count int          `json:"count"`
	Info  []Properties `json:"info"`
}

type Properties struct {
	HostInnerIP string `json:"bk_host_innerip"`
	IdcId       int    `json:"idc_id"`
	IdcName     string `json:"idc_name"`
	IdcUnitId   int    `json:"idc_unit_id"`
	IdcUnitName string `json:"idc_unit_name"`
	SvrTypeName string `json:"svr_type_name"`
	Rack        string `json:"rack"`
}

// getInfoFromBkCmdb gets node info from bk-cmdb with the list_hosts_without_biz api
func getInfoFromBkCmdb(config *config.BcsCcAgentConfig, hostIp string) (*Properties, error) {
	payload := make(map[string]interface{})

	payload["app_code"] = config.AppCode
	payload["app_secret"] = config.AppSecret
	payload["bk_username"] = config.BkUsername
	payload["fields"] = []string{
		"bk_host_innerip",
		"idc_name",
		"idc_unit_name",
		"rack",
		"svr_type_name",
		"idc_unit_id",
		"idc_id",
	}
	payload["page"] = map[string]int{
		"start": 0,
		"limit": 1,
	}
	payload["host_property_filter"] = PropertyFilter{
		Condition: "AND",
		Rules: []Rule{
			{
				Field:    "bk_host_innerip",
				Operator: "equal",
				Value:    hostIp,
			},
		},
	}

	url := config.EsbUrl + "/api/c/compapi/v2/cc/list_hosts_without_biz"
	payloadBytes, _ := json.Marshal(payload)
	body := bytes.NewBuffer(payloadBytes)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request to privilege: %s", err.Error())
	}
	defer resp.Body.Close()

	// Parse body as JSON
	var result APIResponse
	respBody, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(respBody, &result)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code %d", resp.StatusCode)
	}
	if err != nil {
		return nil, fmt.Errorf("non-Json response: %s", err.Error())
	}
	if !result.Result {
		return nil, fmt.Errorf("failed to get host info from bk-cmdb, response code: %d, response message: %s", result.Code, result.Message)
	}

	if result.Data.Count != 1 {
		blog.Infof("%d", result.Data.Count)
		return nil, fmt.Errorf("the host count get from bk-cmdb is not 1")
	}

	return &result.Data.Info[0], nil
}
