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

package bkrepo

type basicResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"traceId"`
}

type basicRecord struct {
	PageNumber   int64 `json:"pageNumber"`
	PageSize     int64 `json:"pageSize"`
	TotalRecords int64 `json:"totalRecords"`
	TotalPages   int64 `json:"totalPages"`
}

type createRepo struct {
	// project id in bk-repo
	ProjectID             string             `json:"projectId"`
	Name                  string             `json:"name"`
	Type                  string             `json:"type"`
	Category              string             `json:"category"`
	Public                bool               `json:"public"`
	Description           string             `json:"description"`
	Configuration         *repoConfiguration `json:"configuration"`
	StorageCredentialsKey string             `json:"storageCredentialsKey"`
	Quota                 int64              `json:"quota"`
}

type repoConfiguration struct {
	Type     string      `json:"type"`
	Settings interface{} `json:"settings"`
}

type repoConfiguration4Local struct {
}

type repoConfiguration4Remote struct {
	URL         string                                    `json:"url"`
	Credentials *repoConfiguration4RemoteAboutCredentials `json:"credentials"`
	Network     *repoConfiguration4RemoteAboutNetwork     `json:"network"`
	Cache       *repoConfiguration4RemoteAboutCache       `json:"cache"`
}

type repoConfiguration4RemoteAboutCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type repoConfiguration4RemoteAboutNetwork struct {
	Proxy          *repoConfiguration4RemoteAboutNetworkProxy `json:"proxy"`
	ConnectTimeout int64                                      `json:"connectTimeout"`
	ReadTimeout    int64                                      `json:"readTimeout"`
}

type repoConfiguration4RemoteAboutNetworkProxy struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type repoConfiguration4RemoteAboutCache struct {
	Enabled    bool  `json:"enabled"`
	Expiration int64 `json:"expiration"`
}
