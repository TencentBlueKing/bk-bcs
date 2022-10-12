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

package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config describe the options Client need
type Config struct {
	// APIServer for bcs-api-gateway address
	APIServer string
	// AuthToken for bcs permission token
	AuthToken string
	// Operator for the bk-repo operations
	Operator string
}

// UserManagerClient defines the client for bcs-user-manager
type UserManagerClient struct {
	cfg *Config
	ctx context.Context
}

// NewClientWithConfiguration new client with config
func NewClientWithConfiguration(ctx context.Context) *UserManagerClient {
	return &UserManagerClient{
		ctx: ctx,
		cfg: &Config{
			APIServer: viper.GetString("config.apiserver"),
			AuthToken: viper.GetString("config.bcs_token"),
			Operator:  viper.GetString("config.operator"),
		},
	}
}

func (c *UserManagerClient) do(url string, httpType string, query map[string]string, body interface{}) ([]byte, error) {
	url = c.cfg.APIServer + url
	var bodyReader *bytes.Reader
	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal body failed")
		}
		bodyReader = bytes.NewReader(bs)
	}
	req, err := http.NewRequestWithContext(c.ctx, httpType, url, bodyReader)
	if err != nil {
		return nil, errors.Wrapf(err, "create request failed")
	}
	if len(query) != 0 {
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "http do request failed")
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read response body failed")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrapf(errors.Errorf(string(bs)), "http response status not 200 but %d",
			resp.StatusCode)
	}
	return bs, nil
}
