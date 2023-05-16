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

package passcc

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/auth"

	"github.com/parnurzeal/gorequest"
	"github.com/patrickmn/go-cache"
)

const (
	cacheCCProjectNamespaceKeyPrefix = "cached_passcc_project_namespace"
)

var (
	defaultTimeOut   = time.Second * 60
	errServerNotInit = errors.New("server not inited")
)

// CClient xxx
var CClient *ClientConfig

// SetCCClient set pass-cc client
func SetCCClient(options Options) error {
	if !options.Enable {
		CClient = nil
		return nil
	}
	cli := NewCCClient(options)

	CClient = cli
	return nil
}

// GetCCClient get pass-cc client
func GetCCClient() *ClientConfig {
	return CClient
}

// NewCCClient init client
func NewCCClient(opt Options) *ClientConfig {
	cli := &ClientConfig{
		server:    opt.Server,
		appCode:   opt.AppCode,
		appSecret: opt.AppSecret,
		debug:     opt.Debug,
		// Create a cache with a default expiration time of 5 minutes, and which
		// purges expired items every 1 hour
		cache: cache.New(time.Minute*5, time.Minute*60),
	}
	return cli
}

// Options opts parameter
type Options struct {
	// Server auth address
	Server string
	// AppCode app code
	AppCode string
	// AppSecret app secret
	AppSecret string
	// Enable enable feature
	Enable bool
	// Debug http debug
	Debug bool
}

// ClientConfig pass-cc client
type ClientConfig struct {
	server string

	appCode   string
	appSecret string
	cache     *cache.Cache
	debug     bool
}

// GetProjectSharedNamespaces get namespaces in pass-cc
func (cc *ClientConfig) GetProjectSharedNamespaces(projectID, clusterID string) ([]string, error) {
	if cc == nil {
		return nil, errServerNotInit
	}

	cacheName := func(projectID, clusterID string) string {
		return fmt.Sprintf("%s_%v_%s", cacheCCProjectNamespaceKeyPrefix, projectID, clusterID)
	}
	val, ok := cc.cache.Get(cacheName(projectID, clusterID))
	if ok && val != nil {
		if namespaces, ok1 := val.([]string); ok1 {
			blog.V(3).Infof("%s %s namespaces[%+v]", projectID, clusterID, namespaces)
			return namespaces, nil
		}
	}
	blog.V(3).Infof("GetProjectSharedNamespaces miss key cache")

	var (
		_    = "GetProjectSharedNamespaces"
		path = fmt.Sprintf("/projects/%s/clusters/%s/namespaces", projectID, clusterID)
	)

	// get access_token
	token, err := cc.getAccessToken(nil)
	if err != nil {
		blog.Errorf("GetProjectSharedNamespaces call getAccessToken failed: %v", err)
		return nil, err
	}

	var (
		url  = cc.server + path
		req  = &GetProjectsNamespaces{DesireAllData: 1}
		resp = &GetProjectsNamespacesResp{}
	)

	// desire_all_data=1 get all namespaces
	result, body, errs := gorequest.New().Timeout(defaultTimeOut).Get(url).
		Query(fmt.Sprintf("access_token=%s", token)).
		Query(fmt.Sprintf("desire_all_data=%s", "1")).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(req).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api GetProjectSharedNamespaces failed: %v", errs[0])
		return nil, errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call GetProjectSharedNamespaces API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return nil, errMsg
	}

	namespaceList := make([]string, 0)
	for _, ns := range resp.Data.Results {
		namespaceList = append(namespaceList, ns.Name)
	}

	err = cc.cache.Add(cacheName(projectID, clusterID), namespaceList, cache.DefaultExpiration)
	if err != nil {
		blog.Errorf("GetProjectSharedNamespaces set cache by cacheName[%s] failed: %v",
			cacheName(projectID, clusterID), err)
	}

	blog.Infof("GetProjectSharedNamespaces[%s:%s] count[%v] successful: %+v", projectID, clusterID,
		len(namespaceList), namespaceList)
	return namespaceList, nil
}

func (cc *ClientConfig) getAccessToken(clientAuth *auth.ClientAuth) (string, error) {
	if cc == nil {
		return "", errServerNotInit
	}

	if clientAuth != nil {
		return clientAuth.GetAccessToken()
	}

	return auth.GetAuthClient().GetAccessToken()
}
