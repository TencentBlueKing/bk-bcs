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

// Package gwparser NOTES
package gwparser

import (
	"context"
	"crypto/md5" //nolint
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// Parser is request header parser.
type Parser interface {
	Parse(ctx context.Context, r http.Header) (kt *kit.Kit, err error)
	Fingerprint() string
}

// defaultParser used to parse requests api-service directly in the scenario.
type defaultParser struct{}

// NewDefaultParser init default Parser
// Note: authorize in prod env may cause security problems, only use for dev/test
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
		return nil, errors.Wrapf(err, "validate kit")
	}

	return kt, nil
}

// Fingerprint 默认不带指纹
func (p *defaultParser) Fingerprint() string {
	return ""
}

// jwtParser used to parse requests from blueking api-gateway.
type jwtParser struct {
	// PublicKey used to parse jwt token from blueking api-gateway http request.
	publicKey    string
	publicKeyObj *rsa.PublicKey
}

// NewJWTParser init JWT Parser
func NewJWTParser(pubKey string) (Parser, error) {
	if pubKey == "" {
		return nil, errors.New("pubkey is required")
	}

	pubKeyObj, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pubKey))
	if err != nil {
		return nil, errors.Wrapf(err, "parse pubkey: %s", pubKey)
	}

	parser := &jwtParser{
		publicKey:    pubKey,
		publicKeyObj: pubKeyObj,
	}
	return parser, nil
}

// Fingerprint golang 指纹实现 https://github.com/golang/go/issues/12292
func (p *jwtParser) Fingerprint() string {
	hash := md5.Sum([]byte(strings.TrimSpace(p.publicKey))) //nolint
	out := ""
	for i := 0; i < 16; i++ {
		if i > 0 {
			out += ":"
		}
		out += fmt.Sprintf("%02x", hash[i]) // don't forget the leading zeroes
	}
	return out
}

// Parse api-gateway request header to context kit and validate.
func (p *jwtParser) Parse(ctx context.Context, header http.Header) (*kit.Kit, error) {
	jwtToken := header.Get(constant.BKGWJWTTokenKey)
	if len(jwtToken) == 0 {
		return nil, errors.Errorf("jwt token header %s is required", constant.BKGWJWTTokenKey)
	}

	token, err := p.parseToken(jwtToken)
	if err != nil {
		return nil, errors.Wrapf(err, "parse jwt token %s", jwtToken)
	}

	if err := token.validate(); err != nil {
		return nil, errors.Wrapf(err, "validate jwt token %s", jwtToken)
	}

	username := token.User.UserName
	if err := token.User.validate(); err != nil {
		username = header.Get(constant.UserKey)
	}

	kt := &kit.Kit{
		Ctx:     ctx,
		User:    username,
		AppCode: token.App.AppCode,
		Rid:     header.Get(constant.RidKey),
	}

	if err := kt.Validate(); err != nil {
		return nil, errors.Wrapf(err, "validate kit")
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
		return errors.New("app not verified")
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
		return errors.New("user not verified")
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
		return errors.New("app info is required")
	}

	if err := c.App.validate(); err != nil {
		return err
	}

	if c.User == nil {
		return errors.New("user info is required")
	}

	return nil
}

// parseToken parse token by jwt token and secret.
func (p *jwtParser) parseToken(token string) (*claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return p.publicKeyObj, nil
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
