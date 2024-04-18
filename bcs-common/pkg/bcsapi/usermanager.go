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

package bcsapi

import (
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	restclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
)

const (
	usermanagerPrefix = types.BCS_MODULE_USERMANAGER + "/v1"
)

// UserManager http API SDK difinition
type UserManager interface {
	// ListAllClusters get all registered kubernetes api-server
	ListAllClusters() ([]*ClusterCredential, error)
}

// NewUserManager create UserManager SDK implementation
func NewUserManager(config *Config) UserManager {
	c := &UserManagerCli{
		Config: config,
	}
	if config.TLSConfig != nil {
		c.Client = restclient.NewRESTClientWithTLS(config.TLSConfig)
	} else {
		c.Client = restclient.NewRESTClient()
	}
	return c
}

// ClusterCredential holds one kubernetes api-server connection credential
type ClusterCredential struct {
	ClusterID     string `json:"clusterId,omitempty"`
	ClusterDomain string `json:"cluster_domain"`
	// kubernetes api-server addresses, splited by comma
	ServerAddresses string `json:"server_addresses"`
	UserToken       string `json:"user_token"`
}

// UserManagerCli client for bcs-user-manager
type UserManagerCli struct {
	Config *Config
	Client *restclient.RESTClient
}

func (cli *UserManagerCli) getRequestPath() string {
	if cli.Config.Gateway {
		// format bcs-api-gateway path
		return fmt.Sprintf("%s%s/", gatewayPrefix, usermanagerPrefix)
	}
	return fmt.Sprintf("/%s/", usermanagerPrefix)
}

// ListAllClusters get all registered kubernetes api-server
func (cli *UserManagerCli) ListAllClusters() ([]*ClusterCredential, error) {
	var response BasicResponse
	err := bkbcsSetting(cli.Client.Get(), cli.Config).
		WithEndpoints(cli.Config.Hosts).
		WithBasePath(cli.getRequestPath()).
		SubPathf("/clusters/credentials").
		Do().
		Into(&response)
	if err != nil {
		return nil, err
	}
	if !response.Result {
		return nil, fmt.Errorf(response.Message)
	}
	// decode specified cluster credentials
	clusterMap := make(map[string]*ClusterCredential)
	if err := json.Unmarshal(response.Data, &clusterMap); err != nil {
		return nil, fmt.Errorf("cluster data decode err: %s", err.Error())
	}
	if len(clusterMap) == 0 {
		// No data retrieve from bcs-user-manager
		return nil, nil
	}
	var clusters []*ClusterCredential
	for k, v := range clusterMap {
		v.ClusterID = k
		clusters = append(clusters, v)
	}
	return clusters, nil
}
