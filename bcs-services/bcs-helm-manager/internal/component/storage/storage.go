/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package storage

import (
	"crypto/tls"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
)

// Client xxx
type Client struct {
	ClientTLSConfig *tls.Config
}

var client *Client

// NewClient create storage service client
func NewClient(tlsConfig *tls.Config) error {
	client = &Client{
		ClientTLSConfig: tlsConfig,
	}
	return nil
}

// GetPods get cluster pods
func GetPods(clusterID, namespace string) ([]*storage.Pod, error) {
	u, err := url.Parse(options.GlobalOptions.Release.APIServer)
	if err != nil {
		return nil, err
	}
	cfg := bcsapi.Config{
		Gateway:   true,
		TLSConfig: client.ClientTLSConfig,
		Hosts:     []string{u.Host},
		AuthToken: options.GlobalOptions.Release.Token,
	}
	cli := bcsapi.NewStorage(&cfg)
	return cli.QueryK8SPod(clusterID, namespace)
}
