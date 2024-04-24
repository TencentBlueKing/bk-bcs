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

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TencentBlueKing/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-external-privilege/dbm"
	"net/http"
)

type DBPrivEnv struct {
	AppName   string   `json:"appName"`
	TargetDb  string   `json:"targetDb"`
	CallUser  string   `json:"callUser"`
	Password  string   `json:"password,omitempty"`
	Grants    []string `json:"grants,omitempty"`
	DbName    string   `json:"dbName"`
	TableName string   `json:"tableName,omitempty"`
	CallType  string   `json:"callType"`
	Operator  string   `json:"operator"`
	UseCDP    bool     `json:"useCDP,omitempty"`
}

func (env *DBPrivEnv) InitClient(op *Option) (ExternalPrivilege, error) {
	if op == nil {
		return nil, fmt.Errorf("InitClient failed, empty options")
	}
	if len(op.ExternalSysType) == 0 || len(op.ExternalSysConfig) == 0 {
		return nil, fmt.Errorf("InitClient failed, empty ExternalSysType or ExternalSysConfig")
	}

	switch op.ExternalSysType {
	case ExternalSysTypeDBM:
		client, err := dbm.NewDBMClient(op)
		if err != nil {
			return nil, err
		}
		return client, nil
	default:
		return nil, fmt.Errorf("unknown ExternalSysType %s", op.ExternalSysType)
	}
}

func (env *DBPrivEnv) IsSCR() bool {
	if env.Password != "" && len(env.Grants) != 0 {
		return true
	}
	return false
}

type RequestEsb struct {
	AppCode     string `json:"bk_app_code"`
	AppSecret   string `json:"bk_app_secret"`
	Operator    string `json:"-"`
	AccessToken string `json:"access_token,omitempty"`
}

func (c *RequestEsb) PrepareRequest(method, url string, payload map[string]interface{}) *http.Request {
	var req *http.Request

	payload["app_code"] = c.AppCode
	payload["app_secret"] = c.AppSecret
	if op, ok := payload["operator"]; !ok || op == "" {
		payload["operator"] = c.Operator
	}
	payloadBytes, _ := json.Marshal(payload)
	body := bytes.NewBuffer(payloadBytes)
	req, _ = http.NewRequest(method, url, body)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bkapi-Accept-Code-Type", "int")
	fmt.Printf("header:\n")
	for k, v := range req.Header {
		if k == "X-Bkapi-Authorization" {
			fmt.Printf("\t%s: xxxxxxxx\n", k)
		} else {
			fmt.Printf("\t%s: %s\n", k, v)
		}
	}
	fmt.Printf("url: %s\n", url)
	fmt.Printf("body:\n")
	for k, v := range payload {
		if k == "app_code" || k == "app_secret" {
			fmt.Printf("\t%s: xxxxxxxx", k)
		} else {
			fmt.Printf("\t%s: %v\n", k, v)
		}
	}
	return req
}
