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

package output

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/app/config"
	"time"
)

// New create a new Output object
func New(ctx context.Context, cfg *config.Config) (Output, error) {

	httpcli := httpclient.NewHttpClient()
	httpcli.SetTlsNoVerity()
	httpcli.SetTimeOut(time.Duration(60) * time.Second)

	httpcli.SetHeader("Content-Type", "application/json")
	httpcli.SetHeader("Accept", "application/json")

	if cfg.ExporterClientCert.IsSSL {
		if err := httpcli.SetTlsVerity(cfg.ExporterClientCert.CAFile, cfg.ExporterClientCert.CertFile, cfg.ExporterClientCert.KeyFile, cfg.ExporterClientCert.CertPasswd); nil != err {
			blog.Error("failed to set tls ")
			return nil, err
		}
	}

	tmp := &output{
		input:  make(chan *InputMessage, 4096),
		cfg:    cfg,
		client: httpcli.GetClient(),
	}

	go tmp.run(ctx)

	return tmp, nil

}
