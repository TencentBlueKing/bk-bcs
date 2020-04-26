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

package utils

import (
	"bk-bcs/bcs-common/common/http"
	"bk-bcs/bcs-common/common/http/httpclient"
	"crypto/tls"
)

type ApiRequester interface {
	Do(uri, method string, data []byte, header ...*http.HeaderSet) ([]byte, error)
	DoForResponse(uri, method string, data []byte, header ...*http.HeaderSet) (*httpclient.HttpRespone, error)
}

//NewApiRequester api request
func NewApiRequester(clientSSL *tls.Config, bcsToken string) ApiRequester {
	return &bcsApiRequester{
		clientSSL: clientSSL,
		bcsToken:  bcsToken,
	}
}

// BcsApiRequester is the way to request to all bcs-api uri
type bcsApiRequester struct {
	clientSSL *tls.Config
	bcsToken  string
}

func (b *bcsApiRequester) Do(uri, method string, data []byte, header ...*http.HeaderSet) ([]byte, error) {
	httpCli := httpclient.NewHttpClient()
	httpCli.SetHeader("Content-Type", "application/json")
	httpCli.SetHeader("Accept", "application/json")
	httpCli.SetHeader("Authorization", "Bearer "+b.bcsToken)
	//httpCli.SetHeader("X-Bcs-User-Token", b.bcsToken)

	if header != nil {
		httpCli.SetBatchHeader(header)
	}

	if b.clientSSL != nil {
		httpCli.SetTlsVerityConfig(b.clientSSL)
	}

	return httpCli.Request(uri, method, nil, data)
}

func (b *bcsApiRequester) DoForResponse(uri, method string, data []byte, header ...*http.HeaderSet) (*httpclient.HttpRespone, error) {
	httpCli := httpclient.NewHttpClient()
	httpCli.SetHeader("Content-Type", "application/json")
	httpCli.SetHeader("Accept", "application/json")
	httpCli.SetHeader("X-Bcs-User-Token", b.bcsToken)

	if header != nil {
		httpCli.SetBatchHeader(header)
	}

	if b.clientSSL != nil {
		httpCli.SetTlsVerityConfig(b.clientSSL)
	}

	return httpCli.RequestEx(uri, method, nil, data)
}
