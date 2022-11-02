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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

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

type ProjectManagerClient struct {
	cfg *Config
	ctx context.Context
}

// NewClientWithConfiguration new client with config
func NewClientWithConfiguration(ctx context.Context) *ProjectManagerClient {
	return &ProjectManagerClient{
		ctx: ctx,
		cfg: &Config{
			APIServer: viper.GetString("bcs.apiserver"),
			AuthToken: viper.GetString("bcs.token"),
			Operator:  viper.GetString("bcs.operator"),
		},
	}
}

func (p *ProjectManagerClient) do(urls string, httpType string, query url.Values, body interface{}) ([]byte, error) {
	urls = p.cfg.APIServer + urls
	var req *http.Request
	var err error
	_, err = url.Parse(p.cfg.APIServer)
	if err != nil {
		return nil, fmt.Errorf("url failed %v", err)
	}
	if body != nil {
		var bs []byte
		bs, err = json.Marshal(body)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal body failed")
		}
		req, err = http.NewRequestWithContext(p.ctx, httpType, urls, bytes.NewReader(bs))
	} else {
		req, err = http.NewRequestWithContext(p.ctx, httpType, urls, nil)
	}
	// 添加鉴权
	if len(p.cfg.AuthToken) != 0 {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.cfg.AuthToken))
	}
	if err != nil {
		return nil, errors.Wrapf(err, "create request failed")
	}
	if query != nil {
		req.URL.RawQuery = query.Encode()
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
