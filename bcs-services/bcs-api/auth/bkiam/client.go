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
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
)

const (
	queryAuthURI = "%s/bkiam/api/v1/perm/systems/%s/resources-perms/verify"
	appCodeKey   = "X-BK-APP-CODE"
	appSecretKey = "X-BK-APP-SECRET"
)

func NewClient(conf *config.ApiServConfig) (*Client, error) {
	c := &Client{
		conf: conf,

		appCode:   conf.BKIamAuth.BKIamAuthAppCode,
		appSecret: conf.BKIamAuth.BKIamAuthAppSecret,

		queryAuthHost: conf.BKIamAuth.BKIamAuthHost,

		systemID: conf.BKIamAuth.BKIamAuthSystemID,
		scopeID:  conf.BKIamAuth.BKIamAuthScopeID,
	}

	c.queryAuthURL = fmt.Sprintf(queryAuthURI, c.queryAuthHost, c.systemID)
	header := http.Header{}
	header.Add(appCodeKey, c.appCode)
	header.Add(appSecretKey, c.appSecret)
	c.queryAuthHeader = header

	client := httpclient.NewHttpClient()
	client.SetTimeOut(30 * time.Second)
	c.client = client

	return c, nil
}

type Client struct {
	conf *config.ApiServConfig

	queryAuthHost string
	appCode       string
	appSecret     string

	queryAuthURL    string
	queryAuthHeader http.Header

	systemID     string
	scopeID      string
	actionID     string
	resourceType string
	resourceID   string

	client *httpclient.HttpClient
}

func (c *Client) Query(username string, action auth.Action, resource auth.Resource) (bool, error) {
	param := &QueryParam{
		PrincipalType: "user",
		PrincipalID:   username,
		ScopeType:     "project",
		ScopeID:       c.scopeID,
		ActionID:      action,
	}
	param.ParseResource(resource)

	resp, err := c.query(param)
	if err != nil {
		blog.Errorf("bkiam auth query failed: %v", err)
		return false, err
	}

	return resp.Data.IsPass, nil
}

func (c *Client) query(param *QueryParam) (*QueryResp, error) {
	var data []byte
	if err := codec.EncJson(param, &data); err != nil {
		blog.Errorf("bkiam auth query encode param failed: %v, url(%s)", err, c.queryAuthURL)
		return nil, err
	}

	raw, err := c.client.Request(c.queryAuthURL, "POST", c.queryAuthHeader, data)
	if err != nil {
		blog.Errorf("bkiam auth query request failed: %v, url(%s), body(%s), resp(%s)", err, c.queryAuthURL, string(data), string(raw))
		return nil, err
	}

	var resp QueryResp
	if err := codec.DecJson(raw, &resp); err != nil {
		blog.Errorf("bkiam auth query decode resp failed: %v, url(%s), body(%s), resp(%s)", err, c.queryAuthURL, string(data), string(raw))
		return nil, err
	}

	return &resp, nil
}
