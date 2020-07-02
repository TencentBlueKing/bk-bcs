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

package bkiam

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/auth"
)

type QueryParam struct {
	PrincipalType string      `json:"principal_type"`
	PrincipalID   string      `json:"principal_id"`
	ScopeType     string      `json:"scope_type"`
	ScopeID       string      `json:"scope_id"`
	ActionID      auth.Action `json:"action_id"`
	ResourceType  string      `json:"resource_type"`
	ResourceID    string      `json:"resource_id"`
}

func (qp *QueryParam) ParseResource(resource auth.Resource) {
	if resource.Namespace == "" {
		qp.ResourceType = "cluster"
		qp.ResourceID = fmt.Sprintf("cluster:%s", resource.ClusterID)
		return
	}

	qp.ResourceType = "namespace"
	qp.ResourceID = fmt.Sprintf("cluster:%s/namespace:%s", resource.ClusterID, resource.Namespace)
}

type QueryResp struct {
	RequestID  string    `json:"request_id"`
	Result     bool      `json:"result"`
	ErrCode    int       `json:"bk_error_code"`
	ErrMessage string    `json:"bk_error_msg"`
	Data       QueryData `json:"data"`
}

type QueryData struct {
	IsPass bool `json:"is_pass"`
}

type ApiGwData struct {
	ISS     string           `json:"iss"`
	App     ApiGwDataApp     `json:"app"`
	Project ApiGwDataProject `json:"project"`
	User    ApiGwDataUser    `json:"user"`
	Exp     float64          `json:"exp"`
	NBF     float64          `json:"nbf"`
}

type ApiGwDataApp struct {
	Version  float64 `json:"version"`
	Verified bool    `json:"verified"`
	AppCode  string  `json:"app_code"`
}

type ApiGwDataProject struct {
	ProjectID   string `json:"project_id"`
	ProjectCode string `json:"project_code"`
	Verified    bool   `json:"verified"`
}

type ApiGwDataUser struct {
	Username string  `json:"username"`
	Version  float64 `json:"version"`
	Verified bool    `json:"verified"`
}
