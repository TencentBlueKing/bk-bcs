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

package api

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Client struct of api client
type Client struct {
	App         *App
	Hook        *Hook
	ConfigItem  *ConfigItem
	Content     *Content
	Commit      *Commit
	Release     *Release
	Instance    *Instance
	StrategySet *StrategySet
	Strategy    *Strategy
	Publish     *Publish
}

// NewApiClient get a new api client
func NewApiClient(host string, c *tools.TLSConfig) (*Client, error) {
	httpCli, err := client.NewClient(c)
	if err != nil {
		return nil, err
	}

	capCli := &client.Capability{
		Client: httpCli,
		Discover: &discovery{
			server: host,
		},
		ToleranceLatencyTime: 2 * time.Second,
		MetricOpts:           client.MetricOption{Register: nil},
	}

	restCli := rest.NewClient(capCli, "/api/v1")

	return &Client{
		App:         NewAppClient(restCli),
		Hook:        NewHookClient(restCli),
		ConfigItem:  NewConfigItemClient(restCli),
		Content:     NewContentClient(restCli),
		Commit:      NewCommitClient(restCli),
		Release:     NewReleaseClient(restCli),
		Instance:    NewInstanceClient(restCli),
		StrategySet: NewStrategySetClient(restCli),
		Strategy:    NewStrategyClient(restCli),
		Publish:     NewPublishClient(restCli),
	}, nil
}

type discovery struct {
	server string
}

// GetServers get api severs
func (r *discovery) GetServers() ([]string, error) {
	if len(r.server) == 0 {
		return nil, errors.New("can not get api server")
	}
	return []string{r.server}, nil
}
