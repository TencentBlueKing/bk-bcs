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
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestWithoutTLSClient(t *testing.T) {
	type args struct {
		header http.Header
		url    string
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "without tls client normal",
			args: args{
				header: make(http.Header),
				url:    "127.0.0.1",
			},
			want: &Client{
				baseURL:    "127.0.0.1",
				header:     make(http.Header),
				HttpClient: &http.Client{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithoutTLSClient(tt.args.header, tt.args.url); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithoutTLSClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Get(t *testing.T) {
	type fields struct {
		Request    *http.Request
		baseURL    string
		header     http.Header
		HttpClient *http.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   *Client
	}{
		{
			name: "test get normal",
			fields: fields{
				baseURL: "127.0.0.1",
			},
			want: &Client{
				baseURL: "127.0.0.1",
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					return tmp
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Request:    tt.fields.Request,
				baseURL:    tt.fields.baseURL,
				header:     tt.fields.header,
				HttpClient: tt.fields.HttpClient,
			}
			if got := c.Get(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_PUT(t *testing.T) {
	type fields struct {
		Request    *http.Request
		baseURL    string
		header     http.Header
		HttpClient *http.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   *Client
	}{
		{
			name: "test put normal",
			fields: fields{
				baseURL: "127.0.0.1",
			},
			want: &Client{
				baseURL: "127.0.0.1",
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("PUT", "127.0.0.1", nil)
					return tmp
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Request:    tt.fields.Request,
				baseURL:    tt.fields.baseURL,
				header:     tt.fields.header,
				HttpClient: tt.fields.HttpClient,
			}
			if got := c.PUT(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.PUT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_POST(t *testing.T) {
	type fields struct {
		Request    *http.Request
		baseURL    string
		header     http.Header
		HttpClient *http.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   *Client
	}{
		{
			name: "test post normal",
			fields: fields{
				baseURL: "127.0.0.1",
			},
			want: &Client{
				baseURL: "127.0.0.1",
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("POST", "127.0.0.1", nil)
					return tmp
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Request:    tt.fields.Request,
				baseURL:    tt.fields.baseURL,
				header:     tt.fields.header,
				HttpClient: tt.fields.HttpClient,
			}
			if got := c.POST(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.POST() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DELETE(t *testing.T) {
	type fields struct {
		Request    *http.Request
		baseURL    string
		header     http.Header
		HttpClient *http.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   *Client
	}{
		{
			name: "test delete normal",
			fields: fields{
				baseURL: "127.0.0.1",
			},
			want: &Client{
				baseURL: "127.0.0.1",
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("DELETE", "127.0.0.1", nil)
					return tmp
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Request:    tt.fields.Request,
				baseURL:    tt.fields.baseURL,
				header:     tt.fields.header,
				HttpClient: tt.fields.HttpClient,
			}
			if got := c.DELETE(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.DELETE() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_AddHeader(t *testing.T) {
	type fields struct {
		Request    *http.Request
		baseURL    string
		header     http.Header
		HttpClient *http.Client
	}
	type args struct {
		header http.Header
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Client
	}{
		{
			name: "test add header normal",
			fields: fields{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					return tmp
				}(),
			},
			args: args{
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			want: &Client{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					tmp.Header.Add("Accept", "application/json")
					return tmp
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Request:    tt.fields.Request,
				baseURL:    tt.fields.baseURL,
				header:     tt.fields.header,
				HttpClient: tt.fields.HttpClient,
			}
			if got := c.AddHeader(tt.args.header); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.AddHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Resource(t *testing.T) {
	type fields struct {
		Request    *http.Request
		baseURL    string
		header     http.Header
		HttpClient *http.Client
	}
	type args struct {
		resource string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Client
	}{
		{
			name: "test resource normal",
			fields: fields{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					tmp.URL, _ = url.Parse("127.0.0.1")
					return tmp
				}(),
			},
			args: args{
				resource: "node",
			},
			want: &Client{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					tmp.URL, _ = url.Parse("127.0.0.1/node")
					return tmp
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Request:    tt.fields.Request,
				baseURL:    tt.fields.baseURL,
				header:     tt.fields.header,
				HttpClient: tt.fields.HttpClient,
			}
			if got := c.Resource(tt.args.resource); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Resource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Name(t *testing.T) {
	type fields struct {
		Request    *http.Request
		baseURL    string
		header     http.Header
		HttpClient *http.Client
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Client
	}{
		{
			name: "test name normal",
			fields: fields{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					tmp.URL, _ = url.Parse("127.0.0.1/node")
					return tmp
				}(),
			},
			args: args{
				name: "n1",
			},
			want: &Client{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					tmp.URL, _ = url.Parse("127.0.0.1/node/n1")
					return tmp
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Request:    tt.fields.Request,
				baseURL:    tt.fields.baseURL,
				header:     tt.fields.header,
				HttpClient: tt.fields.HttpClient,
			}
			if got := c.Name(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Filter(t *testing.T) {
	type fields struct {
		Request    *http.Request
		baseURL    string
		header     http.Header
		HttpClient *http.Client
	}
	type args struct {
		parameters map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Client
	}{
		{
			name: "test filter normal",
			fields: fields{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					tmp.URL, _ = url.Parse("127.0.0.1")
					return tmp
				}(),
			},
			args: args{
				parameters: map[string]string{
					"k1": "v1",
				},
			},
			want: &Client{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					tmp.Form = make(url.Values)
					tmp.Form.Add("k1", "v1")
					return tmp
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Request:    tt.fields.Request,
				baseURL:    tt.fields.baseURL,
				header:     tt.fields.header,
				HttpClient: tt.fields.HttpClient,
			}
			if got := c.Filter(tt.args.parameters); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Base(t *testing.T) {
	type fields struct {
		Request    *http.Request
		baseURL    string
		header     http.Header
		HttpClient *http.Client
	}
	type args struct {
		basePath string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Client
	}{
		{
			name: "test base normal",
			fields: fields{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					return tmp
				}(),
			},
			args: args{
				basePath: "127.0.0.1",
			},
			want: &Client{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("GET", "127.0.0.1", nil)
					tmp.URL, _ = url.Parse("127.0.0.1")
					return tmp
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Request:    tt.fields.Request,
				baseURL:    tt.fields.baseURL,
				header:     tt.fields.header,
				HttpClient: tt.fields.HttpClient,
			}
			if got := c.Base(tt.args.basePath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Base() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Body(t *testing.T) {
	content := &UpdateGroupDesiredSizeRequest{
		DesiredSize: 5,
		Operator:    "bcs",
	}
	byteContent, _ := json.Marshal(&content)
	bodyReader := bytes.NewReader(byteContent)
	bodyBuffer := bytes.NewBuffer(byteContent)
	bodyStringReader := strings.NewReader(string(byteContent))
	type fields struct {
		Request    *http.Request
		baseURL    string
		header     http.Header
		HttpClient *http.Client
	}
	type args struct {
		body io.Reader
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Client
	}{
		{
			name: "test body bytes.Buffer",
			fields: fields{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("POST", "127.0.0.1", nil)
					return tmp
				}(),
			},
			args: args{
				body: bodyBuffer,
			},
			want: &Client{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("POST", "127.0.0.1", nil)
					tmp.Header = make(http.Header)
					tmp.Header.Add("Content-Type", "application/json")
					tmp.Body = ioutil.NopCloser(bodyBuffer)
					tmp.ContentLength = int64(bodyBuffer.Len())
					tmp.GetBody = func() (io.ReadCloser, error) {
						r := bytes.NewReader(bodyBuffer.Bytes())
						return ioutil.NopCloser(r), nil
					}
					return tmp
				}(),
			},
		},
		{
			name: "test body bytes.Reader",
			fields: fields{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("POST", "127.0.0.1", nil)
					return tmp
				}(),
			},
			args: args{
				body: bodyReader,
			},
			want: &Client{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("POST", "127.0.0.1", nil)
					tmp.Header = make(http.Header)
					tmp.Header.Add("Content-Type", "application/json")
					tmp.Body = ioutil.NopCloser(bodyReader)
					tmp.ContentLength = int64(bodyReader.Len())
					tmp.GetBody = func() (io.ReadCloser, error) {
						return ioutil.NopCloser(bodyReader), nil
					}
					return tmp
				}(),
			},
		},
		{
			name: "test body strings.Reader",
			fields: fields{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("POST", "127.0.0.1", nil)
					return tmp
				}(),
			},
			args: args{
				body: bodyStringReader,
			},
			want: &Client{
				Request: func() *http.Request {
					tmp, _ := http.NewRequest("POST", "127.0.0.1", nil)
					tmp.Header = make(http.Header)
					tmp.Header.Add("Content-Type", "application/json")
					tmp.Body = ioutil.NopCloser(bodyStringReader)
					tmp.ContentLength = int64(bodyStringReader.Len())
					tmp.GetBody = func() (io.ReadCloser, error) {
						return ioutil.NopCloser(bodyStringReader), nil
					}
					return tmp
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Request:    tt.fields.Request,
				baseURL:    tt.fields.baseURL,
				header:     tt.fields.header,
				HttpClient: tt.fields.HttpClient,
			}
			if got := c.Body(tt.args.body); !reflect.DeepEqual(got, tt.want) {
				got.Request.GetBody = nil
				tt.want.GetBody = nil
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Client.Body() = %+v, want %+v", got.Request, tt.want.Request)
				}
			}
		})
	}
}
