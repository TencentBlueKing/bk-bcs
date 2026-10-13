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

package httpx

import (
	"errors"
	"reflect"
	"testing"

	bcsJwt "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/agiledragon/gomonkey/v2"
	jwtGo "github.com/golang-jwt/jwt/v4"
)

// 由CodeBuddy（内网版）生成于2026.06.04 10:39:24
func Test_jwtDecode(t *testing.T) {
	type args struct {
		jwtToken string
	}
	tests := []struct {
		name    string
		args    args
		want    *bcsJwt.UserClaimsInfo
		wantErr bool
		prepare func(patches *gomonkey.Patches)
	}{
		{
			name:    "JWTPubKeyObj is nil",
			args:    args{jwtToken: "test"},
			want:    nil,
			wantErr: true,
			prepare: func(patches *gomonkey.Patches) {
				config.G.Auth.JWTPubKeyObj = nil
			},
		},
		{
			name:    "ParseWithClaims error",
			args:    args{jwtToken: "invalid_token"},
			want:    nil,
			wantErr: true,
			prepare: func(patches *gomonkey.Patches) {
				patches.ApplyFunc(jwtGo.ParseWithClaims, func(tokenString string, claims jwtGo.Claims, keyFunc jwtGo.Keyfunc) (*jwtGo.Token, error) {
					return nil, errors.New("parse error")
				})
			},
		},
		{
			name:    "token invalid",
			args:    args{jwtToken: "invalid_token"},
			want:    nil,
			wantErr: true,
			prepare: func(patches *gomonkey.Patches) {
				patches.ApplyFunc(jwtGo.ParseWithClaims, func(tokenString string, claims jwtGo.Claims, keyFunc jwtGo.Keyfunc) (*jwtGo.Token, error) {
					return &jwtGo.Token{Valid: false}, nil
				})
			},
		},
		{
			name:    "claims type error",
			args:    args{jwtToken: "invalid_token"},
			want:    nil,
			wantErr: true,
			prepare: func(patches *gomonkey.Patches) {
				patches.ApplyFunc(jwtGo.ParseWithClaims, func(tokenString string, claims jwtGo.Claims, keyFunc jwtGo.Keyfunc) (*jwtGo.Token, error) {
					return &jwtGo.Token{Valid: true, Claims: &jwtGo.StandardClaims{}}, nil
				})
			},
		},
		{
			name:    "success",
			args:    args{jwtToken: "valid_token"},
			want:    &bcsJwt.UserClaimsInfo{},
			wantErr: false,
			prepare: func(patches *gomonkey.Patches) {
				mockClaims := &bcsJwt.UserClaimsInfo{}
				patches.ApplyFunc(jwtGo.ParseWithClaims, func(tokenString string, claims jwtGo.Claims, keyFunc jwtGo.Keyfunc) (*jwtGo.Token, error) {
					return &jwtGo.Token{Valid: true, Claims: mockClaims}, nil
				})
			},
		},
	}

	origPubKey := config.G.Auth.JWTPubKeyObj
	defer func() {
		config.G.Auth.JWTPubKeyObj = origPubKey
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			config.G.Auth.JWTPubKeyObj = nil

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			if tt.prepare != nil {
				tt.prepare(patches)
			}

			got, err := jwtDecode(tt.args.jwtToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("%q. jwtDecode() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%q. jwtDecode() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
