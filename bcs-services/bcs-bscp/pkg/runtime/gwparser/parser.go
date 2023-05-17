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

package gwparser

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/kit"
)

// Parser is request header parser.
type Parser interface {
	Parse(ctx context.Context, r http.Header) (kt *kit.Kit, err error)
}

// defaultParser used to parse requests api-service directly in the scenario.
type defaultParser struct{}

// Note: authorize in prod env may cause security problems, only use for dev/test
// NewDefaultParser
func NewDefaultParser() Parser {
	return &defaultParser{}
}

// Parse http request header to context kit and validate.
func (p *defaultParser) Parse(ctx context.Context, header http.Header) (*kit.Kit, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	kt := &kit.Kit{
		Ctx:     ctx,
		User:    header.Get(constant.UserKey),
		Rid:     header.Get(constant.RidKey),
		AppCode: header.Get(constant.AppCodeKey),
	}

	if err := kt.Validate(); err != nil {
		return nil, errf.New(errf.InvalidParameter, err.Error())
	}

	return kt, nil
}

// jwtParser used to parse requests from blueking api-gateway.
type jwtParser struct {
	// PublicKey used to parse jwt token from blueking api-gateway http request.
	PublicKey    string
	PublicKeyObj *rsa.PublicKey `yaml:"-"`
}

// NewJWTParser
func NewJWTParser(pubKey string) (Parser, error) {
	if pubKey == "" {
		return nil, errors.New("pubkey is required")
	}

	pubKeyObj, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pubKey))
	if err != nil {
		return nil, errors.Wrapf(err, "parse pubkey: %s", pubKey)
	}

	parser := &jwtParser{
		PublicKey:    pubKey,
		PublicKeyObj: pubKeyObj,
	}
	return parser, nil
}

// Parse api-gateway request header to context kit and validate.
func (p *jwtParser) Parse(ctx context.Context, header http.Header) (*kit.Kit, error) {
	jwtToken := header.Get(constant.BKGWJWTTokenKey)
	if len(jwtToken) == 0 {
		return nil, errf.New(errf.InvalidParameter, "jwt token is required")
	}

	token, err := p.parseToken(jwtToken, p.PublicKey)
	if err != nil {
		return nil, err
	}

	if err := token.validate(); err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	kt := &kit.Kit{
		Ctx:     ctx,
		User:    token.User.UserName,
		AppCode: token.App.AppCode,
		Rid:     header.Get(constant.RidKey),
	}

	if err := kt.Validate(); err != nil {
		return nil, errf.New(errf.InvalidParameter, err.Error())
	}

	return kt, nil
}

// app blueking application info.
type app struct {
	Version  int64  `json:"version"`
	AppCode  string `json:"app_code"`
	Verified bool   `json:"verified"`
}

// validate app.
func (a *app) validate() error {
	if !a.Verified {
		return errf.New(errf.InvalidParameter, "app not verified")
	}
	return nil
}

// user blueking user info.
type user struct {
	Version  int64  `json:"version"`
	UserName string `json:"username"`
	Verified bool   `json:"verified"`
}

// validate user.
func (u *user) validate() error {
	if !u.Verified {
		return errf.New(errf.InvalidParameter, "user not verified")
	}
	return nil
}

// claims blueking api gateway jwt struct.
type claims struct {
	App  *app  `json:"app"`
	User *user `json:"user"`
	jwt.RegisteredClaims
}

// validate claims.
func (c *claims) validate() error {
	if c.App == nil {
		return errf.New(errf.InvalidParameter, "app info is required")
	}

	if err := c.App.validate(); err != nil {
		return err
	}

	if c.User == nil {
		return errf.New(errf.InvalidParameter, "user info is required")
	}

	if err := c.User.validate(); err != nil {
		return err
	}

	return nil
}

// parseToken parse token by jwt token and secret.
func (p *jwtParser) parseToken(token, jwtSecret string) (*claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return p.PublicKeyObj, nil
	})
	if err != nil {
		return nil, err
	}

	if tokenClaims == nil {
		return nil, errors.New("can not get token from parse with claims")
	}

	claims, ok := tokenClaims.Claims.(*claims)
	if !ok {
		return nil, errors.New("token claims type error")
	}

	if !tokenClaims.Valid {
		return nil, errors.New("token claims valid failed")
	}

	return claims, nil
}
