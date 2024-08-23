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

// Package appauth xxx
package appauth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	paasclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
)

// Config item for BlueKing Auth Center
type Config struct {
	// Hosts AuthCenter hosts, without http/https, default is http
	Hosts []string
	// config for https
	TLSConfig *tls.Config
	AppCode   string
	AppSecret string
}

// Client BlueKing Auth Center interface difinition
type Client interface {
	// GetAccessToken get specified environmental BK PaaS access token
	// by App identifier. Access token expires in 180 days
	GetAccessToken(env string) (string, error)
}

// NewAuthClient create authClient instance
func NewAuthClient(cfg *Config) (Client, error) {
	// validate config
	if len(cfg.Hosts) == 0 {
		return nil, fmt.Errorf("Lost hosts config item(required)")
	}
	if len(cfg.AppCode) == 0 || len(cfg.AppSecret) == 0 {
		return nil, fmt.Errorf("Lost Authorization information(required)")
	}
	var c *authClient
	if cfg.TLSConfig != nil {
		c = &authClient{
			config: cfg,
			client: paasclient.NewRESTClientWithTLS(cfg.TLSConfig).
				WithCredential(cfg.AppCode, cfg.AppSecret),
		}
	} else {
		c = &authClient{
			config: cfg,
			client: paasclient.NewRESTClient().
				WithCredential(cfg.AppCode, cfg.AppSecret),
		}
	}
	return c, nil
}

// authClient auth center sdk implementation
type authClient struct {
	config *Config
	client *paasclient.RESTClient
}

// GetAccessToken get specified environmental BK PaaS access token
func (c *authClient) GetAccessToken(env string) (string, error) {
	if !(env == PaasOAuthEnvProd || env == PaasOAuthEnvTest) {
		return "", fmt.Errorf("Error Environment")
	}
	request := map[string]interface{}{
		"env_name":   env,
		"grant_type": PaasGrantTypeClient,
	}
	auth := map[string]string{
		"app_code":   c.config.AppCode,
		"app_secret": c.config.AppSecret,
	}
	authBytes, _ := json.Marshal(auth)
	authHeader := http.Header{}
	authHeader.Add("X-Bkapi-Authorization", string(authBytes))
	var response OAuthResponse
	err := c.client.Post().
		WithEndpoints(c.config.Hosts).
		WithHeaders(authHeader).
		WithBasePath("/").
		SubPathf("/auth_api/token/").
		Body(request).
		Do().
		Into(&response)
	if err != nil {
		return "", err
	}
	if response.Code != 0 {
		return "", fmt.Errorf("GetAccessToken failed, %s", response.Message)
	}
	return response.Data.AccessToken, nil
}
