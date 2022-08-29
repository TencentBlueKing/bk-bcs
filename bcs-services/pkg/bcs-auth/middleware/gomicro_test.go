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

package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/micro/go-micro/v2/codec"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"
)

func TestGoMicroAuth(t *testing.T) {
	jwtClient, err := jwt.NewJWTClient(jwt.JWTOptions{VerifyKeyFile: "./testdata/app.rsa.pub",
		SignKeyFile: "./testdata/app.rsa"})
	if err != nil {
		t.Fatal(err)
	}
	clientUser, err := jwtClient.JWTSign(&jwt.UserInfo{SubType: jwt.Client.String(), ClientName: "test",
		Issuer: jwt.JWTIssuer, ExpiredTime: int64(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	plainUser, err := jwtClient.JWTSign(&jwt.UserInfo{SubType: jwt.User.String(), UserName: "test",
		Issuer: jwt.JWTIssuer, ExpiredTime: int64(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name          string
		ctx           context.Context
		skipHandler   func(ctx context.Context, req server.Request) bool
		exemptClient  func(ctx context.Context, req server.Request, client string) bool
		checkUserPerm func(ctx context.Context, req server.Request, username string) (bool, error)
		allow         bool
		err           bool
	}{
		{
			name:  "no header",
			ctx:   metadata.NewContext(context.TODO(), map[string]string{}),
			allow: false,
			err:   true,
		},
		{
			name: "inner client",
			ctx: metadata.NewContext(context.TODO(), map[string]string{
				InnerClientHeaderKey: "test",
			}),
			allow: true,
		},
		{
			name: "skip client user",
			ctx: metadata.NewContext(context.TODO(), map[string]string{
				AuthorizationHeaderKey: "Bearer " + clientUser,
			}),
			exemptClient: func(ctx context.Context, req server.Request, client string) bool {
				return client == "test"
			},
			allow: true,
		},
		{
			name: "client and username, not allow",
			ctx: metadata.NewContext(context.TODO(), map[string]string{
				AuthorizationHeaderKey:  "Bearer " + clientUser,
				CustomUsernameHeaderKey: "user1",
			}),
			checkUserPerm: func(ctx context.Context, req server.Request, username string) (bool, error) {
				return false, nil
			},
			err:   true,
			allow: false,
		},
		{
			name: "client and username, allow",
			ctx: metadata.NewContext(context.TODO(), map[string]string{
				AuthorizationHeaderKey:  "Bearer " + clientUser,
				CustomUsernameHeaderKey: "user1",
			}),
			checkUserPerm: func(ctx context.Context, req server.Request, username string) (bool, error) {
				return username == "user1", nil
			},
			allow: true,
		},
		{
			name: "plain user, allow",
			ctx: metadata.NewContext(context.TODO(), map[string]string{
				AuthorizationHeaderKey: "Bearer " + plainUser,
			}),
			checkUserPerm: func(ctx context.Context, req server.Request, username string) (bool, error) {
				return username == "test", nil
			},
			allow: true,
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			srv := &MockServer{ctx: v.ctx}
			auth := NewGoMicroAuth(jwtClient)
			auth.EnableSkipHandler(v.skipHandler)
			auth.EnableSkipClient(v.exemptClient)
			auth.SetCheckUserPerm(v.checkUserPerm)
			srv.WrapHandler(auth.AuthenticationFunc)
			srv.WrapHandler(auth.AuthorizationFunc)
			err := srv.Do()
			if (err != nil) != v.err {
				t.Errorf("error: %v", err)
			}
			if srv.allow != v.allow {
				t.Errorf("expect %v, got %v", v.allow, srv.allow)
			}
		})
	}
}

// MockServer for mocking gomicro server
type MockServer struct {
	ctx      context.Context
	allow    bool
	wrappers []func(server.HandlerFunc) server.HandlerFunc
}

// Request grpc request
type Request struct {
	service     string
	method      string
	endpoint    string
	contentType string
	header      map[string]string
	body        []byte
	rawBody     interface{}
	stream      bool
	first       bool
}

func (r Request) Service() string {
	return r.service
}

func (r Request) Method() string {
	return r.method
}

func (r Request) Endpoint() string {
	return r.endpoint
}

func (r Request) ContentType() string {
	return r.contentType
}

func (r Request) Header() map[string]string {
	return r.header
}

func (r Request) Body() interface{} {
	return r.rawBody
}

func (r Request) Read() ([]byte, error) {
	return nil, nil
}

func (r Request) Codec() codec.Reader {
	return nil
}

func (r Request) Stream() bool {
	return r.stream
}

func (m *MockServer) WrapHandler(w ...server.HandlerWrapper) {
	for _, wrap := range w {
		m.wrappers = append(m.wrappers, wrap)
	}
}

func (m *MockServer) Do() error {
	handler := m.SampleFunc
	for i := len(m.wrappers); i > 0; i-- {
		handler = m.wrappers[i-1](handler)
	}
	return handler(m.ctx, Request{}, nil)
}

func (m *MockServer) SampleFunc(ctx context.Context, req server.Request, rsp interface{}) error {
	m.allow = true
	return nil
}
