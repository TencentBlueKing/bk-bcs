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

			config.G.Auth.JWTPubKeyObj = 1

			patches := gomonkey.NewPatches()
			defer patches.Reset()

			if tt.prepare != nil {
				tt.prepare(patches)
			}

			got, err := jwtDecode(tt.args.jwtToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("%q. jwtDecode() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				continue
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%q. jwtDecode() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
