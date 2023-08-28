/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"context"
	"net/http"
	"strconv"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/rest"
)

// Content content client
type Content struct {
	client rest.ClientInterface
}

// NewContentClient new a content client
func NewContentClient(client rest.ClientInterface) *Content {
	return &Content{
		client: client,
	}
}

// Upload is to upload content.
func (c *Content) Upload(ctx context.Context, header http.Header, bizID, appID, tmplSpaceID uint32, sign string,
	data string) (*rest.Response, error) {
	if appID > 0 {
		header.Set(constant.AppIDHeaderKey, strconv.FormatUint(uint64(appID), 10))
	} else {
		header.Set(constant.TmplSpaceIDHeaderKey, strconv.FormatUint(uint64(tmplSpaceID), 10))
	}
	header.Set(constant.ContentIDHeaderKey, sign)

	resp := c.client.Put().
		WithContext(ctx).
		SubResourcef("/bizs/%d/content/upload", bizID).
		WithHeaders(header).
		Body(data).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	r := new(rest.Response)
	if err := resp.Into(r); err != nil {
		return nil, err
	}

	return r, nil
}

// Download is to upload content.
func (c *Content) Download(ctx context.Context, header http.Header, bizID, appID, tmplSpaceID uint32, sign string) (
	[]byte, error) {
	if appID > 0 {
		header.Set(constant.AppIDHeaderKey, strconv.FormatUint(uint64(appID), 10))
	} else {
		header.Set(constant.TmplSpaceIDHeaderKey, strconv.FormatUint(uint64(tmplSpaceID), 10))
	}
	header.Set(constant.ContentIDHeaderKey, sign)

	resp := c.client.Get().
		WithContext(ctx).
		SubResourcef("/bizs/%d/content/download", bizID).
		WithHeaders(header).
		Do()

	if resp.Err != nil {
		return nil, resp.Err
	}

	return resp.Body, nil
}
