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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	"github.com/avast/retry-go"
	"github.com/parnurzeal/gorequest"
)

const (
	retryCount    = 10
	retryInterval = 5 * time.Second
)

// 重试状态码
var retryStatusCode = []int{503, 504}

// ErrThirdPartyTimeout is timeout error
var ErrThirdPartyTimeout = fmt.Errorf("third party timeout")

// 是否可重试判断，根据 err 判断
var retryable = retry.RetryIf(func(err error) bool {
	return errors.Is(err, ErrThirdPartyTimeout)
})

// BCSAppClient bcs app client
type BCSAppClient struct {
	server     string
	appCode    string
	appSecret  string
	bkUsername string
	debug      bool
}

// NewBCSAppClient creates a new bcs app client
func NewBCSAppClient(server, appCode, appSecret, bkUsername string, debug bool) *BCSAppClient {
	return &BCSAppClient{
		server:     server,
		appCode:    appCode,
		appSecret:  appSecret,
		bkUsername: bkUsername,
		debug:      debug,
	}
}

func (c *BCSAppClient) generateGateWayAuth() (string, error) {
	if c == nil {
		return "", ErrServerNotInit
	}

	auth := &AuthInfo{
		BkAppCode:   c.appCode,
		BkAppSecret: c.appSecret,
		BkUserName:  c.bkUsername,
	}

	userAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(userAuth), nil
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

// ListCharts list charts
func (c *BCSAppClient) ListCharts(projectID string) (*ListChartsResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	url := fmt.Sprintf("%s/backend/apis/projects/%s/helm/charts", c.server, projectID)
	userAuth, err := c.generateGateWayAuth()
	if err != nil {
		blog.Errorf("bcs app client generateGateWayAuth failed: %v", err)
		return nil, err
	}

	resp := &ListChartsResponse{}
	err = retry.Do(func() error {
		_, _, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Retry(retryCount, retryInterval, retryStatusCode...).
			Get(url).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", userAuth).
			SetDebug(c.debug).
			EndStruct(resp)
		if len(errs) > 0 {
			blog.Errorf("call api ListCharts failed: %v", errs[0])
			return errs[0]
		}
		if resp.Code != 0 {
			blog.Warnf("call api ListCharts failed: %s", utils.ToJSONString(resp))
			return ErrThirdPartyTimeout
		}
		return nil
	}, retry.Attempts(retryCount), retry.Delay(retryInterval), retryable)
	if err != nil {
		return nil, fmt.Errorf("call api ListCharts failed: %v, resp: %s", err, utils.ToJSONString(resp))
	}
	return resp, nil
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

// GetApp get app
func (c *BCSAppClient) GetApp(projectID string, appID int) (*GetAppResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	url := fmt.Sprintf("%s/backend/apis/projects/%s/helm/apps/%d", c.server, projectID, appID)
	userAuth, err := c.generateGateWayAuth()
	if err != nil {
		blog.Errorf("bcs app client generateGateWayAuth failed: %v", err)
		return nil, err
	}

	resp := &GetAppResponse{}
	err = retry.Do(func() error {
		_, _, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Retry(retryCount, retryInterval, retryStatusCode...).
			Get(url).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", userAuth).
			SetDebug(c.debug).
			EndStruct(resp)
		if len(errs) > 0 {
			blog.Errorf("call api GetApp failed: %v", errs[0])
			return errs[0]
		}
		if resp.Code != 0 {
			blog.Warnf("call api GetApp failed: %s", utils.ToJSONString(resp))
			return ErrThirdPartyTimeout
		}
		return nil
	}, retry.Attempts(retryCount), retry.Delay(retryInterval))
	if err != nil {
		return nil, fmt.Errorf("call api GetApp failed: %v, resp: %s", err, utils.ToJSONString(resp))
	}
	return resp, nil
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

// ListNamespace list namespace
func (c *BCSAppClient) ListNamespace(projectID, clusterID string) (*ListNamespaceResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	url := fmt.Sprintf("%s/backend/apis/projects/%s/helm/namespaces?cluster_id=%s", c.server,
		projectID, clusterID)
	userAuth, err := c.generateGateWayAuth()
	if err != nil {
		blog.Errorf("bcs app client generateGateWayAuth failed: %v", err)
		return nil, err
	}

	resp := &ListNamespaceResponse{}
	err = retry.Do(func() error {
		_, _, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Retry(retryCount, retryInterval, retryStatusCode...).
			Get(url).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", userAuth).
			SetDebug(c.debug).
			EndStruct(resp)
		if len(errs) > 0 {
			blog.Errorf("call api ListNamespace failed: %v", errs[0])
			return errs[0]
		}
		// REQUEST_BACKEND_TIMEOUT, need retry
		if resp.Code != 0 {
			blog.Warnf("call api ListNamespace failed: %s", utils.ToJSONString(resp))
			return ErrThirdPartyTimeout
		}
		return nil
	}, retry.Attempts(retryCount), retry.Delay(retryInterval))
	if err != nil {
		return nil, fmt.Errorf("call api ListNamespace failed: %v, resp: %s", err, utils.ToJSONString(resp))
	}
	return resp, nil
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

// ListApps list apps
func (c *BCSAppClient) ListApps(projectID, clusterID, namespace string,
	page, limit, offset int) (*ListAppsResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	url := fmt.Sprintf("%s/backend/apis/projects/%s/helm/apps?cluster_id=%s&namespace=%s&page=%d&offset=%d&limit=%d",
		c.server, projectID, clusterID, namespace, page, offset, limit)
	userAuth, err := c.generateGateWayAuth()
	if err != nil {
		blog.Errorf("bcs app client generateGateWayAuth failed: %v", err)
		return nil, err
	}
	resp := &ListAppsResponse{}
	err = retry.Do(func() error {
		_, _, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Retry(retryCount, retryInterval, retryStatusCode...).
			Get(url).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", userAuth).
			SetDebug(c.debug).
			EndStruct(resp)
		if len(errs) > 0 {
			blog.Errorf("call api ListApps failed: %v", errs[0])
			return errs[0]
		}
		if resp.Code != 0 {
			blog.Warnf("call api ListApps failed: %s", utils.ToJSONString(resp))
			return ErrThirdPartyTimeout
		}
		return nil
	}, retry.Attempts(retryCount), retry.Delay(retryInterval))
	if err != nil {
		return nil, fmt.Errorf("call api ListApps failed: %v, resp: %s", err, utils.ToJSONString(resp))
	}
	return resp, nil
}

// CreateAppRequest create app request
type CreateAppRequest struct {
	ProjectID     string
	Answers       []string `json:"answers"`
	Name          string   `json:"name"`
	ClusterID     string   `json:"cluster_id"`
	ChartVersion  int      `json:"chart_version"`
	NamespaceInfo int      `json:"namespace_info"`
	ValueFile     string   `json:"valuefile"`
}

// CreateAppResponse create app response
type CreateAppResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// CreateApp create app
func (c *BCSAppClient) CreateApp(req *CreateAppRequest) (*CreateAppResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	url := fmt.Sprintf("%s/backend/apis/projects/%s/helm/apps", c.server, req.ProjectID)
	userAuth, err := c.generateGateWayAuth()
	if err != nil {
		blog.Errorf("bcs app client generateGateWayAuth failed: %v", err)
		return nil, err
	}

	resp := &CreateAppResponse{}
	err = retry.Do(func() error {
		_, _, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Retry(retryCount, retryInterval, retryStatusCode...).
			Post(url).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", userAuth).
			SetDebug(c.debug).
			Send(req).
			EndStruct(resp)
		if len(errs) > 0 {
			blog.Errorf("call api CreateApp failed: %v", errs[0])
			return errs[0]
		}
		if resp.Code != 0 {
			blog.Warnf("call api CreateApp failed: %s", utils.ToJSONString(resp))
			return ErrThirdPartyTimeout
		}
		return nil
	}, retry.Attempts(retryCount), retry.Delay(retryInterval))
	if err != nil {
		return nil, fmt.Errorf("call api CreateApp failed: %v, resp: %s", err, utils.ToJSONString(resp))
	}
	return resp, nil
}

// UpdateAppRequest update app request
type UpdateAppRequest struct {
	ProjectID      string
	AppID          int
	Answers        []string `json:"answers"`
	UpgradeVersion int      `json:"upgrade_verion"`
	ValueFile      string   `json:"valuefile"`
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

// UpdateApp update app
func (c *BCSAppClient) UpdateApp(req *UpdateAppRequest) (*UpdateAppResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	url := fmt.Sprintf("%s/backend/apis/projects/%s/helm/apps/%d", c.server, req.ProjectID, req.AppID)
	userAuth, err := c.generateGateWayAuth()
	if err != nil {
		blog.Errorf("bcs app client generateGateWayAuth failed: %v", err)
		return nil, err
	}

	resp := &UpdateAppResponse{}
	err = retry.Do(func() error {
		_, bt, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Retry(retryCount, retryInterval, retryStatusCode...).
			Put(url).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", userAuth).
			SetDebug(c.debug).
			Send(req).
			EndStruct(resp)
		if len(errs) > 0 {
			blog.Errorf("call api UpdateApp failed: %v, resp: %s", errs[0], string(bt))
			return errs[0]
		}
		if resp.Code != 0 {
			blog.Warnf("call api UpdateApp failed: %s", utils.ToJSONString(resp))
			return ErrThirdPartyTimeout
		}
		return nil
	}, retry.Attempts(retryCount), retry.Delay(retryInterval))
	if err != nil {
		return nil, fmt.Errorf("call api UpdateApp failed: %v, resp: %s", err, utils.ToJSONString(resp))
	}
	return resp, nil
}
