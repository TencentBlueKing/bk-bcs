/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package esb

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"time"

	"bk-bscp/pkg/ssl"
)

const (
	// defaultTLSHandshakeTimeout is default tls handshake timeout.
	defaultTLSHandshakeTimeout = 5 * time.Second

	// defaultDialerTimeout is default dialer timeout.
	defaultDialerTimeout = 10 * time.Second

	// defaultMaxConnsPerHost is default max connections limit for per host.
	defaultMaxConnsPerHost = 500

	// defaultMaxIdleConnsPerHost is default max idle connections limit for per host.
	defaultMaxIdleConnsPerHost = 100

	// defaultIdleConnTimeout is default idle connection timeout.
	defaultIdleConnTimeout = time.Minute
)

// Option is esb comm option.
type Option struct {
	// Endpoints is esb target endpoints used for discovery and load balance.
	Endpoints []string

	// CertFile authentication certificate file.
	CertFile string

	// KeyFile authentication key file.
	KeyFile string

	// CAFile authentication root certificate file.
	CAFile string

	// Password authentication key file password.
	Password string
}

// Validate validates option.
func (opt *Option) Validate() error {
	if opt == nil {
		return errors.New("empty option")
	}
	if len(opt.Endpoints) == 0 {
		return errors.New("empty endpoints")
	}
	return nil
}

// Client is esb client.
type Client struct {
	// endpoints communication and load balance.
	endpoints []string

	// endpoints index.
	index uint64

	// http client.
	client *http.Client

	// url schema.
	schema string
}

// NewClient creates a new Client.
func NewClient(opt *Option) (*Client, error) {
	// validates option.
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	// new esb client.
	newClient := &Client{endpoints: opt.Endpoints, schema: "http"}

	// tls config.
	var tlsConf *tls.Config
	var err error

	// build tls.
	if len(opt.CAFile) != 0 || len(opt.CertFile) != 0 || len(opt.KeyFile) != 0 {
		if tlsConf, err = ssl.ClientTLSConfVerify(opt.CAFile, opt.CertFile, opt.KeyFile, opt.Password); err != nil {
			return nil, err
		}
		newClient.schema = "https"
	}

	// create http client.
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			Dial:                (&net.Dialer{Timeout: defaultDialerTimeout}).Dial,
			TLSHandshakeTimeout: defaultTLSHandshakeTimeout,
			TLSClientConfig:     tlsConf,
			MaxConnsPerHost:     defaultMaxConnsPerHost,
			MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost,
			IdleConnTimeout:     defaultIdleConnTimeout,
		},
	}
	newClient.client = httpClient

	return newClient, nil
}

// endpoint returns one target endpoint in RR mode.
func (cli *Client) endpoint() string {
	index := atomic.AddUint64(&cli.index, 1)

	num := uint64(len(cli.endpoints))
	if num == 0 {
		return ""
	}
	return cli.endpoints[index%num]
}

// Do handles the esb request base on http client.
func (cli *Client) Do(method, uri string, data []byte) ([]byte, error) {
	// pick one endpoint.
	target := cli.endpoint()
	if len(target) == 0 {
		return nil, errors.New("empty target endpoint")
	}

	// build request.
	uri = filepath.Clean("/" + uri)
	url := fmt.Sprintf("%s://%s%s", cli.schema, target, uri)

	request, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	response, err := cli.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status[%+v]", response.StatusCode)
	}
	return ioutil.ReadAll(response.Body)
}
