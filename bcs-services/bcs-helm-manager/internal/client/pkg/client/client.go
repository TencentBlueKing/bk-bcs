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

package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg"

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/protobuf/runtime/protoiface"
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

const (
	resultCodeSuccess = 0

	urlPrefix = "/bcsapi/v4"
)

// New return a new Client instance
func New(c Config) pkg.HelmClient {
	return &Client{
		conf: &c,
		cli:  httpclient.NewHttpClient(),
	}
}

// Client provide http actions to server
type Client struct {
	conf *Config
	cli  *httpclient.HttpClient
}

func (c *Client) get(ctx context.Context, uri string, header http.Header, data []byte) (
	*httpclient.HttpRespone, error) {
	return c.request(ctx, "GET", uri, header, data)
}

func (c *Client) post(ctx context.Context, uri string, header http.Header, data []byte) (
	*httpclient.HttpRespone, error) {
	return c.request(ctx, "POST", uri, header, data)
}

func (c *Client) put(ctx context.Context, uri string, header http.Header, data []byte) (
	*httpclient.HttpRespone, error) {
	return c.request(ctx, "PUT", uri, header, data)
}

func (c *Client) delete(ctx context.Context, uri string, header http.Header, data []byte) (
	*httpclient.HttpRespone, error) {
	return c.request(ctx, "DELETE", uri, header, data)
}

func (c *Client) request(_ context.Context, method, uri string, header http.Header, data []byte) (
	*httpclient.HttpRespone, error) {

	if header == nil {
		header = http.Header{}
	}
	header.Set("Content-Type", "application/json")
	header.Set("Authorization", "Bearer "+c.conf.AuthToken)

	var request func(string, http.Header, []byte) (*httpclient.HttpRespone, error)
	switch strings.ToUpper(method) {
	case "GET":
		request = c.cli.Get
	case "POST":
		request = c.cli.Post
	case "PUT":
		request = c.cli.Put
	case "DELETE":
		request = c.cli.Delete
	default:
		return nil, fmt.Errorf("unknown method %s", method)
	}

	fmt.Printf("request to %s\n", c.conf.APIServer+uri)
	r, err := request(c.conf.APIServer+uri, header, data)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to server failed, http(%d)%s: %s", r.StatusCode, r.Status, uri)
	}

	return r, nil
}

func unmarshalPB(data []byte, m protoiface.MessageV1) error {
	unmarshaler := &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
	return unmarshaler.Unmarshal(bytes.NewReader(data), m)
}
