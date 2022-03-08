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

package bkrepo

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	bkRepoAuth "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo/bkrepo/auth"
)

const (
	respCodeOK = 0
)

var (
	errNotExist     = fmt.Errorf("not exist")
	errAlreadyExist = fmt.Errorf("already exist")
)

func (h *handler) get(ctx context.Context, uri string, header http.Header, data []byte) (
	*httpclient.HttpRespone, error) {
	return h.bkRepo.client.request(ctx, "GET", h.getUri(uri), h.auth, header, data)
}

func (h *handler) post(ctx context.Context, uri string, header http.Header, data []byte) (
	*httpclient.HttpRespone, error) {
	return h.bkRepo.client.request(ctx, "POST", h.getUri(uri), h.auth, header, data)
}

func (h *handler) put(ctx context.Context, uri string, header http.Header, data []byte) (
	*httpclient.HttpRespone, error) {
	return h.bkRepo.client.request(ctx, "PUT", h.getUri(uri), h.auth, header, data)
}

func (h *handler) delete(ctx context.Context, uri string, header http.Header, data []byte) (
	*httpclient.HttpRespone, error) {
	return h.bkRepo.client.request(ctx, "DELETE", h.getUri(uri), h.auth, header, data)
}

func (h *handler) getUri(uri string) string {
	return h.config.URL + uri
}

func newClient() *client {
	return &client{
		cli: httpclient.NewHttpClient(),
	}
}

type client struct {
	cli *httpclient.HttpClient
}

func (c *client) request(
	_ context.Context, method, uri string, auth *bkRepoAuth.Auth, header http.Header, data []byte) (
	*httpclient.HttpRespone, error) {

	// init header
	if header == nil {
		header = http.Header{}
	}
	header.Set("Content-Type", "application/json")
	if auth != nil {
		auth.SetHeader(header)
	}
	blog.V(5).Infof("request to bk-repo [%s] %s, header(%v), body(%s)", method, uri, header, data)

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

	beforeReq := time.Now().Local()
	r, err := request(uri, header, data)
	blog.V(5).Infof("request to bk-repo [%s] %s spent time %s",
		method, uri, time.Now().Local().Sub(beforeReq).String())

	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK && r.StatusCode != http.StatusBadRequest {
		return nil, fmt.Errorf("request to bk-repo failed, http(%d)%s: %s", r.StatusCode, r.Status, uri)
	}
	blog.V(5).Infof("request to bk-repo [%s] %s, get resp(%s)", method, uri, string(r.Reply))

	return r, nil
}
