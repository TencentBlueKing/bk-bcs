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
	"fmt"
	"net/http"

	restclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/parnurzeal/gorequest"
)

func configureRequest(r *gorequest.SuperAgent, config *Config) *gorequest.SuperAgent { // nolint
	// setting insecureSkipVerify
	if config.TLSConfig != nil {
		config.TLSConfig.InsecureSkipVerify = true
		r.TLSClientConfig(config.TLSConfig)
	}
	if config.AuthToken != "" {
		r.Set("Authorization", fmt.Sprintf("Bearer %s", config.AuthToken))
	}
	if config.ClusterID != "" {
		r.Set(clusterIDHeader, config.ClusterID)
	}
	return r
}

func newGet(config *Config, address string) *gorequest.SuperAgent { // nolint
	r := gorequest.New().Get(address)
	return configureRequest(r, config)
}

func newPost(config *Config, address string) *gorequest.SuperAgent { // nolint
	r := gorequest.New().Post(address)
	return configureRequest(r, config)
}

func newDelete(config *Config, address string) *gorequest.SuperAgent { // nolint
	r := gorequest.New().Delete(address)
	return configureRequest(r, config)
}

func bkbcsSetting(req *restclient.Request, config *Config) *restclient.Request {
	header := make(http.Header)
	if config.AuthToken != "" {
		header.Add("Authorization", fmt.Sprintf("Bearer %s", config.AuthToken))
	}
	if config.ClusterID != "" {
		header.Add(clusterIDHeader, config.ClusterID)
	}
	return req.WithHeaders(header)
}
