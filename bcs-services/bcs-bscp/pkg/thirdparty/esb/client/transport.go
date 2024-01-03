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

package client

import (
	"encoding/json"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/types"
)

// esbAuthTransport esb auth transport
type esbAuthTransport struct {
	Transport  http.RoundTripper
	commParams *types.CommParams
	authValue  string
}

// newEsbAuthTransport .
func newEsbAuthTransport(cfg *cc.Esb, transport http.RoundTripper) (http.RoundTripper, error) {
	params := types.GetCommParams(cfg)
	value, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	t := &esbAuthTransport{
		commParams: params,
		authValue:  string(value),
		Transport:  transport,
	}
	return t, nil
}

// RoundTrip esb auth Transport
func (t *esbAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Bkapi-Authorization", t.authValue)
	resp, err := t.transport(req).RoundTrip(req)
	return resp, err
}

func (t *esbAuthTransport) transport(req *http.Request) http.RoundTripper { //nolint:unparam
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}
