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

// Package client NOTES
package client

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/rest/client"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/bklogin"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/cmdb"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
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

	// esb 鉴权中间件
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
		cc:         cmdb.NewClient(restCli),
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
