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
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/avast/retry-go"
	"github.com/parnurzeal/gorequest"
)

var (
	// BcsApp bcsPlatform install
	BcsApp install.InstallerType = "bcs_app"
)

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

// generateGateWayAuth create gateway auth
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

	// List charts
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

	// get app
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

// CreateNamespace create namespace wait bcs-gateway cluster normal
func (c *BCSAppClient) CreateNamespace(projectID, clusterID string, req CreateNamespaceRequest) (
	*CreateNamespaceResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	url := fmt.Sprintf("%s/apis/resources/projects/%s/clusters/%s/namespaces/", c.server, projectID, clusterID)
	userAuth, err := c.generateGateWayAuth()
	if err != nil {
		blog.Errorf("bcs app client generateGateWayAuth failed: %v", err)
		return nil, err
	}

	var (
		defaultRetryCnt  uint = 10
		defaultRetryTime      = 10 * time.Second
	)

	// create namespace
	resp := &CreateNamespaceResponse{}
	err = retry.Do(func() error {
		_, _, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Post(url).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", userAuth).
			SetDebug(c.debug).
			Send(req).
			EndStruct(resp)
		if len(errs) > 0 {
			blog.Errorf("call api CreateNamespace failed: %v", errs[0])
			return errs[0]
		}
		if resp.Code != 0 {
			blog.Warnf("call api CreateNamespace failed: %s", utils.ToJSONString(resp))
			return ErrThirdPartyTimeout
		}
		return nil
	}, retry.Attempts(defaultRetryCnt), retry.Delay(defaultRetryTime), retry.DelayType(retry.FixedDelay))
	if err != nil {
		return nil, fmt.Errorf("call api CreateNamespace failed: %v, resp: %s", err, utils.ToJSONString(resp))
	}

	blog.Infof("BCSAppClient CreateNamespace successful[%s:%s:%s]", projectID, clusterID, req.Name)

	return resp, nil
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

	// list namespace
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

	// list apps
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

	// create app
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

	// update app
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

// DeleteApp delete app
func (c *BCSAppClient) DeleteApp(req *DeleteAppRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	url := fmt.Sprintf("%s/backend/apis/projects/%s/helm/apps/%d", c.server, req.ProjectID, req.AppID)
	userAuth, err := c.generateGateWayAuth()
	if err != nil {
		blog.Errorf("bcs app client generateGateWayAuth failed: %v", err)
		return err
	}

	// delete app
	resp := &DeleteAppResponse{}
	err = retry.Do(func() error {
		_, bt, errs := gorequest.New().
			Timeout(defaultTimeOut).
			Retry(retryCount, retryInterval, retryStatusCode...).
			Delete(url).
			Set("Content-Type", "application/json").
			Set("Accept", "application/json").
			Set("X-Bkapi-Authorization", userAuth).
			SetDebug(c.debug).
			EndStruct(resp)
		if len(errs) > 0 {
			blog.Errorf("call api DeleteApp failed: %v, resp: %s", errs[0], string(bt))
			return errs[0]
		}
		if resp.Code != 0 {
			blog.Warnf("call api DeleteApp failed: %s", utils.ToJSONString(resp))
			return ErrThirdPartyTimeout
		}

		return nil
	}, retry.Attempts(retryCount), retry.Delay(retryInterval))
	if err != nil {
		return fmt.Errorf("call api DeleteApp failed: %v, resp: %s", err, utils.ToJSONString(resp))
	}

	blog.Infof("bcs app client DeleteApp successful[%s:%v]", req.ProjectID, req.AppID)
	return nil
}
