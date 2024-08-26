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

// Package cmdbv1 xxx
package cmdbv1

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	paasclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
)

// ClientInterface client interface for cmdb v1
type ClientInterface interface {
	// ESBTransHostModule trans hosts modules
	ESBTransHostModule(username string, assetIDs []string, appID, moduleID int64) (*ESBTransHostModuleResult, error)
}

// NewClientInterface create client interface
func NewClientInterface(host string, tlsConf *tls.Config) *Client {
	var cli *paasclient.RESTClient
	if tlsConf != nil {
		cli = paasclient.NewRESTClientWithTLS(tlsConf).
			WithRateLimiter(throttle.NewTokenBucket(100, 100))
	} else {
		cli = paasclient.NewRESTClient().
			WithRateLimiter(throttle.NewTokenBucket(100, 100))
	}
	return &Client{
		host:          host,
		client:        cli,
		defaultHeader: http.Header{},
	}
}

// Client paas cmdb v1 client
type Client struct {
	host          string
	defaultHeader http.Header
	client        *paasclient.RESTClient
	credential    Credential
}

// Credential credential to be filled in post body
type Credential struct {
	BKAppCode   string `json:"bk_app_code"`
	BKAppSecret string `json:"bk_app_secret"`
	BKUsername  string `json:"bk_username,omitempty"`
}

// SetDefaultHeader set default headers
func (c *Client) SetDefaultHeader(h http.Header) {
	c.defaultHeader = h
}

// GetHeader get headers
func (c *Client) GetHeader() http.Header {
	authBytes, _ := json.Marshal(c.credential)
	c.defaultHeader.Add("X-Bkapi-Authorization", string(authBytes))
	return c.defaultHeader
}

// WithCredential set credential
func (c *Client) WithCredential(appCode, appSecret, username string) {
	c.credential = Credential{
		BKAppCode:   appCode,
		BKAppSecret: appSecret,
		BKUsername:  username,
	}
}

// ESBTransHostModule trans hosts modules
func (c *Client) ESBTransHostModule(username string, assetIDs []string, appID, moduleID int64) (
	*ESBTransHostModuleResult, error) {

	if len(assetIDs) == 0 {
		return nil, fmt.Errorf("asset ids cannot be empty")
	}
	hostConditions := []map[string]string{}

	for _, assetID := range assetIDs {
		hostConditions = append(hostConditions, map[string]string{
			"host_assetId":  assetID,
			"module_id":     strconv.FormatInt(moduleID, 10),
			"ApplicationID": strconv.FormatInt(appID, 10),
		})
	}

	req := map[string]interface{}{
		"operator":              username,
		"app_id":                strconv.FormatInt(appID, 10),
		"host_module_condition": hostConditions,
	}
	result := new(ESBTransHostModuleResult)
	err := c.client.Post().
		WithEndpoints([]string{c.host}).
		WithBasePath("/component/compapi/cc/").
		SubPathf("host_module").
		WithHeaders(c.GetHeader()).
		Body(req).
		Do().
		Into(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
