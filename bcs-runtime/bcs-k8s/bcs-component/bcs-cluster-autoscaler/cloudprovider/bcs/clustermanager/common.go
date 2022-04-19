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

package clustermanager

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/klog"
)

// Client is common sdk client
type Client struct {
	*http.Request
	baseURL    string
	header     http.Header
	HttpClient *http.Client
}

// WithoutTLSClient init a non-tls client
func WithoutTLSClient(header http.Header, url string) *Client {
	c := &Client{
		header:     nil,
		HttpClient: &http.Client{},
	}
	c.baseURL = url
	c.header = header
	return c
}

// Get set the http Method `GET`
func (c *Client) Get() *Client {
	request, _ := http.NewRequest("GET", c.baseURL, nil)
	c.Request = request
	if c.header != nil {
		c.AddHeader(c.header)
	}
	return c
}

// PUT set the http Method `PUT`
func (c *Client) PUT() *Client {
	request, _ := http.NewRequest("PUT", c.baseURL, nil)
	c.Request = request
	if c.header != nil {
		c.AddHeader(c.header)
	}
	return c
}

// POST set the http Method `POST`
func (c *Client) POST() *Client {
	request, _ := http.NewRequest("POST", c.baseURL, nil)
	c.Request = request
	if c.header != nil {
		c.AddHeader(c.header)
	}
	return c
}

// DELETE set the http Method `DELETE`
func (c *Client) DELETE() *Client {
	request, _ := http.NewRequest("DELETE", c.baseURL, nil)
	c.Request = request
	if c.header != nil {
		c.AddHeader(c.header)
	}
	return c
}

// AddHeader adds header to http header
func (c *Client) AddHeader(header http.Header) *Client {
	if c.Request != nil {
		for k, values := range header {
			for _, v := range values {
				c.Request.Header.Add(k, v)

			}
		}
		return c
	}
	return nil
}

// Resource set the resource to format url, e.g. nodepools, nodes
func (c *Client) Resource(resource string) *Client {
	if len(resource) == 0 {
		return c
	}
	if c.Request != nil {
		urlPath := c.Request.URL
		if urlPath == nil {
			return nil
		}
		url, err := url.Parse(strings.Join([]string{urlPath.String(), resource}, "/"))
		if err != nil {
			klog.Errorf("resourc: %v, %v", resource, err)
			return nil
		}
		c.URL = url
	}
	return c
}

// Name set the required resource name to the resource
func (c *Client) Name(name string) *Client {
	if len(name) == 0 {
		return c
	}
	if c.Request != nil {
		urlPath := c.Request.URL
		if urlPath == nil {
			return nil
		}
		url, err := url.Parse(strings.Join([]string{urlPath.String(), name}, "/"))
		if err != nil {
			klog.Error(err)
			return nil
		}
		c.URL = url
	}
	return c
}

// Filter set the required resource name to the resource
func (c *Client) Filter(parameters map[string]string) *Client {
	if len(parameters) == 0 {
		return c
	}
	if c.Request != nil {
		if c.Form == nil {
			c.Form = make(map[string][]string, 0)
		}
		for k, v := range parameters {
			c.Form.Add(k, v)
		}
	}
	return c
}

// Base set the base url of client
func (c *Client) Base(basePath string) *Client {
	if c.Request != nil {
		url, err := url.Parse(basePath)
		if err != nil {
			return nil
		}
		c.URL = url
		return c
	}
	return nil
}

// Body converts the body, it receives
// *bytes.Buffer, *strings.Buffer and *bytes.Buffer
func (c *Client) Body(body io.Reader) *Client {
	if c == nil {
		return nil
	}
	if c.Request == nil {
		return nil
	}
	if body == nil {
		return nil
	}
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	c.Request.Body = rc
	c.Request.Header.Add("Content-Type", "application/json")
	switch v := body.(type) {
	case *bytes.Buffer:
		c.Request.ContentLength = int64(v.Len())
		buf := v.Bytes()
		c.Request.GetBody = func() (io.ReadCloser, error) {
			r := bytes.NewReader(buf)
			return ioutil.NopCloser(r), nil
		}
	case *bytes.Reader:
		c.Request.ContentLength = int64(v.Len())
		snapshot := *v
		c.Request.GetBody = func() (io.ReadCloser, error) {
			r := snapshot
			return ioutil.NopCloser(&r), nil
		}
	case *strings.Reader:
		c.Request.ContentLength = int64(v.Len())
		snapshot := *v
		c.Request.GetBody = func() (io.ReadCloser, error) {
			r := snapshot
			return ioutil.NopCloser(&r), nil
		}
	default:
		// This is where we'd set it to -1 (at least
		// if body != NoBody) to mean unknown, but
		// that broke people during the Go 1.8 testing
		// period. People depend on it being 0 I
		// guess. Maybe retry later. See Issue 18117.
	}
	// For client requests, Request.ContentLength of 0
	// means either actually 0, or unknown. The only way
	// to explicitly say that the ContentLength is zero is
	// to set the Body to nil. But turns out too much code
	// depends on NewRequest returning a non-nil Body,
	// so we use a well-known ReadCloser variable instead
	// and have the http package also treat that sentinel
	// variable to mean explicitly zero.
	if c.Request.GetBody != nil && c.Request.ContentLength == 0 {
		c.Request.Body = http.NoBody
		c.Request.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
	}
	return c
}

// WithContext set the context
func (c *Client) WithContext(ctx context.Context) *Client {
	c.Request.WithContext(ctx)
	return c
}

// Do finishes the http request
func (c *Client) Do() ([]byte, error) {
	klog.V(4).Infof("Query %v, header: %+v, body: %+v", c.URL.String(), c.Request.Header, c.Request.Body)
	resp, err := c.HttpClient.Do(c.Request)
	if err != nil {
		return nil, fmt.Errorf("failed to finish this request: %v", err)
	}
	defer resp.Body.Close()
	contentsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %+v err: %v", resp, err)
	}
	if resp.StatusCode/100 > 2 {
		return nil, fmt.Errorf("failed to finish this request: %v, body: %v", resp.StatusCode, string(contentsBytes))
	}
	return contentsBytes, nil
}
