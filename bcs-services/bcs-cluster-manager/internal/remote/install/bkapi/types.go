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

package bkapi

import (
	"errors"
	"fmt"
	"time"

	"github.com/avast/retry-go"
)

const (
	retryCount    = 3
	retryInterval = 100 * time.Millisecond
)

// 重试状态码
var retryStatusCode = []int{503, 504}

// ErrThirdPartyTimeout is timeout error
var ErrThirdPartyTimeout = fmt.Errorf("third party timeout")

// 是否可重试判断，根据 err 判断
var retryable = retry.RetryIf(func(err error) bool {
	return errors.Is(err, ErrThirdPartyTimeout)
})

var (
	defaultTimeOut         = time.Second * 5
	defaultCheckAppTimeout = time.Minute * 10
	// ErrServerNotInit server notInit
	ErrServerNotInit = errors.New("server not inited")
)

// AuthInfo auth info
type AuthInfo struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	BkUserName  string `json:"bk_username"`
}

// ListChartsResponse list charts response
type ListChartsResponse struct {
	Code      int              `json:"code"`
	Message   string           `json:"message"`
	RequestID string           `json:"request_id"`
	Data      []ListChartsData `json:"data"`
}

// ListChartsData list charts data
type ListChartsData struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Repository struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"repository"`
	DefaultChartVersion struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"defaultChartVersion"`
}

// GetAppResponse get app response
type GetAppResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      *GetAppData `json:"data"`
}

// GetAppData get app data
type GetAppData struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	Namespace            string `json:"namespace"`
	ClusterID            string `json:"cluster_id"`
	TransitioningOn      bool   `json:"transitioning_on"`
	TransitioningMessage string `json:"transitioning_message"`
	TransitioningResult  bool   `json:"transitioning_result"`
	Release              struct {
		ID                   int `json:"id"`
		ChartVersionSnapshot struct {
			ID        int `json:"id"`
			VersionID int `json:"version_id"`
		} `json:"chartVersionSnapshot"`
	} `json:"release"`
}

// CreateNamespaceRequest create namespace request
type CreateNamespaceRequest struct {
	Name        string            `json:"name"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// CreateNamespaceResponse response
type CreateNamespaceResponse struct {
	Code      int           `json:"code"`
	Message   string        `json:"message"`
	RequestID string        `json:"request_id"`
	Data      NamespaceData `json:"data"`
}

// NamespaceData namespace data
type NamespaceData struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
}

// ListNamespaceResponse list namespace response
type ListNamespaceResponse struct {
	Code      int                 `json:"code"`
	Message   string              `json:"message"`
	RequestID string              `json:"request_id"`
	Data      []ListNamespaceData `json:"data"`
}

// ListNamespaceData list namespace data
type ListNamespaceData struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ClusterID string `json:"cluster_id"`
}

// ListAppsResponse list apps response
type ListAppsResponse struct {
	Code      int          `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"request_id"`
	Data      ListAppsData `json:"data"`
}

// ListAppsData list apps data
type ListAppsData struct {
	Count   int              `json:"count"`
	Results []ListAppsResult `json:"results"`
}

// ListAppsResult list apps result
type ListAppsResult struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	Namespace            string `json:"namespace"`
	ClusterID            string `json:"cluster_id"`
	ProjectID            string `json:"project_id"`
	TransitioningOn      bool   `json:"transitioning_on"`
	TransitioningMessage string `json:"transitioning_message"`
	TransitioningResult  bool   `json:"transitioning_result"`
}

// CreateAppRequest create app request
type CreateAppRequest struct {
	ProjectID     string
	Answers       []string                 `json:"answers"`
	Name          string                   `json:"name"`
	ClusterID     string                   `json:"cluster_id"`
	ChartVersion  int                      `json:"chart_version"`
	NamespaceInfo int                      `json:"namespace_info"`
	ValueFile     string                   `json:"valuefile"`
	CmdFlags      []map[string]interface{} `json:"cmd_flags"`
}

// CreateAppResponse create app response
type CreateAppResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// UpdateAppRequest update app request
type UpdateAppRequest struct {
	ProjectID      string
	AppID          int
	Answers        []string                 `json:"answers"`
	UpgradeVersion int                      `json:"upgrade_verion"`
	ValueFile      string                   `json:"valuefile"`
	CmdFlags       []map[string]interface{} `json:"cmd_flags"`
}

// UpdateAppResponse update app response
type UpdateAppResponse struct {
	Code      int           `json:"code"`
	Message   string        `json:"message"`
	RequestID string        `json:"request_id"`
	Data      UpdateAppData `json:"data"`
}

// UpdateAppData struct for update app
type UpdateAppData struct {
	Name                 string `json:"name"`
	Namespace            string `json:"namespace"`
	ClusterID            string `json:"cluster_id"`
	TransitioningOn      bool   `json:"transitioning_on"`
	TransitioningMessage string `json:"transitioning_message"`
	TransitioningResult  bool   `json:"transitioning_result"`
}

// DeleteAppRequest delete app request
type DeleteAppRequest struct {
	ProjectID string
	AppID     int
}

// DeleteAppResponse delete app response
type DeleteAppResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}
