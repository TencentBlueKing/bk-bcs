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

package authcenter

import (
	"crypto/tls"
	"fmt"

	paasclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
)

// Config item for BlueKing Auth Center
// AuthCenter requires AccessToken for authorization
type Config struct {
	//Hosts AuthCenter hosts, default https
	Hosts []string
	//config for https, when setting tls, use https instead
	TLSConfig   *tls.Config
	AccessToken string
}

// Client BlueKing Auth Center interface difinition
type Client interface {
	//QueryProjectUsers list all project users
	QueryProjectUsers(projectID string) ([]string, error)
}

// NewAuthClient create authClient instance
func NewAuthClient(cfg *Config) (Client, error) {
	//validate config
	if len(cfg.Hosts) == 0 {
		return nil, fmt.Errorf("Lost hosts config item(required)")
	}
	if len(cfg.AccessToken) == 0 {
		return nil, fmt.Errorf("lost AccessToken(required)")
	}
	var c *authClient
	if cfg.TLSConfig != nil {
		c = &authClient{
			config: cfg,
			client: paasclient.NewRESTClientWithTLS(cfg.TLSConfig),
		}
	} else {
		c = &authClient{
			config: cfg,
			client: paasclient.NewRESTClient(),
		}
	}
	return c, nil
}

// authClient auth center sdk implementation
type authClient struct {
	config *Config
	client *paasclient.RESTClient
}

//QueryProjectUsers list all project users
func (c *authClient) QueryProjectUsers(projectID string) ([]string, error) {
	response := &UsersQueryResponse{}
	err := c.client.Get().
		WithEndpoints(c.config.Hosts).
		WithBasePath("/").
		SubPathf("/projects/%s/users", projectID).
		WithParam("access_token", c.config.AccessToken).
		Do().
		Into(response)
	if err != nil {
		return nil, err
	}
	if response.Code != 0 {
		return nil, fmt.Errorf("query project user list failed, %s", response.Message)
	}
	return response.Data, nil
}
