/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
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

	"bscp.io/pkg/cc"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/rest/client"
	"bscp.io/pkg/thirdparty/esb/bklogin"
	"bscp.io/pkg/thirdparty/esb/cmdb"
	"bscp.io/pkg/thirdparty/esb/types"
	"bscp.io/pkg/tools"

	"github.com/prometheus/client_golang/prometheus"
)

// Client NOTES
type Client interface {
	Cmdb() cmdb.Client
	BKLogin() bklogin.Client
}

// NewClient new esb client.
func NewClient(cfg *cc.Esb, reg prometheus.Registerer) (Client, error) {
	tls := &tools.TLSConfig{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		CertFile:           cfg.TLS.CertFile,
		KeyFile:            cfg.TLS.KeyFile,
		CAFile:             cfg.TLS.CAFile,
		Password:           cfg.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}

	authTransport, err := newEsbAuthTransport(cfg, tools.NewCurlLogTransport(cli.Transport))
	if err != nil {
		return nil, err
	}

	cli.Transport = authTransport
	c := &client.Capability{
		Client: cli,
		Discover: &esbDiscovery{
			servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	restCli := rest.NewClient(c, "/api/c/compapi/v2")

	return &esbCli{
		cc:         cmdb.NewClient(restCli, cfg),
		bkloginCli: bklogin.NewClient(restCli),
	}, nil
}

type esbCli struct {
	cc         cmdb.Client
	bkloginCli bklogin.Client
}

// Cmdb NOTES
func (e *esbCli) Cmdb() cmdb.Client {
	return e.cc
}

// BKLogin NOTES
func (e *esbCli) BKLogin() bklogin.Client {
	return e.bkloginCli
}

// curlLogTransport print curl log transport
type esbAuthTransport struct {
	Transport  http.RoundTripper
	commParams *types.CommParams
	authValue  string
}

func newEsbAuthTransport(cfg *cc.Esb, Transport http.RoundTripper) (http.RoundTripper, error) {
	params := types.GetCommParams(cfg)
	value, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	t := &esbAuthTransport{
		commParams: params,
		authValue:  string(value),
		Transport:  Transport,
	}
	return t, nil
}

// RoundTrip curlLog Transport
func (t *esbAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Bkapi-Authorization", t.authValue)
	resp, err := t.transport(req).RoundTrip(req)
	return resp, err
}

func (t *esbAuthTransport) transport(req *http.Request) http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}
