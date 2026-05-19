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

package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/middleware"
	"go-micro.dev/v4/codec"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"
)

// mockCodec implements codec.Reader for testing
type mockCodec struct{}

func (m *mockCodec) ReadHeader(msg *codec.Message, typ codec.MessageType) error { return nil }
func (m *mockCodec) ReadBody(msg interface{}) error                            { return nil }

// mockRequest implements server.Request for testing
type mockRequest struct {
	method string
}

func (m *mockRequest) Service() string            { return "test-service" }
func (m *mockRequest) Method() string             { return m.method }
func (m *mockRequest) Endpoint() string           { return m.method }
func (m *mockRequest) ContentType() string        { return "application/json" }
func (m *mockRequest) Header() map[string]string  { return nil }
func (m *mockRequest) Body() interface{}          { return nil }
func (m *mockRequest) Read() ([]byte, error)      { return nil, nil }
func (m *mockRequest) Stream() bool               { return false }
func (m *mockRequest) Codec() codec.Reader        { return &mockCodec{} }

func newTestJWTClient(t *testing.T) *jwt.JWTClient {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	cli, err := jwt.NewJWTClient(jwt.JWTOptions{
		SignKey:   privateKey,
		VerifyKey: &privateKey.PublicKey,
	})
	if err != nil {
		t.Fatalf("create JWT client: %v", err)
	}
	return cli
}

func signTestToken(t *testing.T, cli *jwt.JWTClient, userType jwt.UserType, username, clientName string) string {
	t.Helper()
	token, err := cli.JWTSign(&jwt.UserInfo{
		SubType:     userType.String(),
		UserName:    username,
		ClientName:  clientName,
		ExpiredTime: 3600,
	})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func TestAuthorizationFunc(t *testing.T) {
	cli := newTestJWTClient(t)
	jwtAuth := &JWTAuth{client: cli}

	tests := []struct {
		name           string
		setupMetadata  func(md metadata.Metadata)
		expectUsername string
		expectClient   string
		expectError    bool
	}{
		{
			name: "no_jwt_username_header_ignored",
			setupMetadata: func(md metadata.Metadata) {
				md.Set(middleware.CustomUsernameHeaderKey, "fake_admin")
			},
		},
		{
			name: "jwt_user_ignores_username_header",
			setupMetadata: func(md metadata.Metadata) {
				token := signTestToken(t, cli, jwt.User, "jwt_user", "")
				md.Set(middleware.AuthorizationHeaderKey, "Bearer "+token)
				md.Set(middleware.CustomUsernameHeaderKey, "fake_admin")
			},
			expectUsername: "jwt_user",
		},
		{
			name: "jwt_client_ignores_username_header",
			setupMetadata: func(md metadata.Metadata) {
				token := signTestToken(t, cli, jwt.Client, "", "bk-apigateway")
				md.Set(middleware.AuthorizationHeaderKey, "Bearer "+token)
				md.Set(middleware.CustomUsernameHeaderKey, "proxied_user")
			},
			expectClient: "bk-apigateway",
		},
		{
			name: "inner_client_only",
			setupMetadata: func(md metadata.Metadata) {
				md.Set(middleware.InnerClientHeaderKey, "inner-service")
			},
		},
		{
			name: "invalid_jwt",
			setupMetadata: func(md metadata.Metadata) {
				md.Set(middleware.AuthorizationHeaderKey, "Bearer invalid_token")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := make(metadata.Metadata)
			tt.setupMetadata(md)
			ctx := metadata.NewContext(context.Background(), md)

			var capturedAuthUser *middleware.AuthUser
			handler := jwtAuth.AuthorizationFunc(func(ctx context.Context, req server.Request, rsp interface{}) error {
				md, _ := metadata.FromContext(ctx)
				data, ok := md.Get(string(middleware.AuthUserKey))
				if ok && data != "" {
					authUser := &middleware.AuthUser{}
					json.Unmarshal([]byte(data), authUser)
					capturedAuthUser = authUser
				}
				return nil
			})

			err := handler(ctx, &mockRequest{method: "test.method"}, nil)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if capturedAuthUser == nil {
				if tt.expectUsername != "" || tt.expectClient != "" {
					t.Errorf("expected username=%q client=%q but authUser was nil", tt.expectUsername, tt.expectClient)
				}
				return
			}
			if capturedAuthUser.Username != tt.expectUsername {
				t.Errorf("username: want %q, got %q", tt.expectUsername, capturedAuthUser.Username)
			}
			if capturedAuthUser.ClientName != tt.expectClient {
				t.Errorf("client: want %q, got %q", tt.expectClient, capturedAuthUser.ClientName)
			}
		})
	}
}
