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

package jwt

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	// JWTIssuer issuer
	JWTIssuer = "BCS"
)

// BCSJWTAuthentication interface for jwt sign/decode
type BCSJWTAuthentication interface {
	JWTSign(user *UserInfo) (string, error)
	JWTDecode(jwtToken string) (*UserClaimsInfo, error)
}

// BCSJWTSigningMethod default sign method
var BCSJWTSigningMethod = jwt.SigningMethodRS256

type UserType string

// String to string
func (ut UserType) String() string {
	return string(ut)
}

var (
	User   UserType = "user"
	Client UserType = "client"
)

// UserInfo userInfo
type UserInfo struct {
	SubType      string
	UserName     string
	ClientName   string
	ClientSecret string
	Issuer       string
	// ExpiredTime second (0 for permanent jwtToken, > 0 for duration)
	ExpiredTime int64
}

func (user *UserInfo) validate() error {
	if user.SubType != User.String() && user.SubType != Client.String() {
		return ErrJWtSubType
	}

	if user.SubType == User.String() && user.UserName == "" {
		return ErrJWtUserNameEmpty
	}
	if user.SubType == Client.String() && user.ClientName == "" {
		return ErrJWtClientNameEmpty
	}

	if user.Issuer == "" {
		user.Issuer = JWTIssuer
	}
	// token may be empty
	if user.ExpiredTime <= 0 {
		user.ExpiredTime = 0
	}

	return nil
}

// UserClaimsInfo custom jwt claims
type UserClaimsInfo struct {
	SubType      string `json:"sub_type"`
	UserName     string `json:"username"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	// https://tools.ietf.org/html/rfc7519#section-4.1
	// aud: 接收jwt一方; exp: jwt过期时间; jti: jwt唯一身份认证; IssuedAt: 签发时间; Issuer: jwt签发者
	*jwt.StandardClaims
}

// JWTOptions jwt auth keyInfo
type JWTOptions struct {
	VerifyKey     *rsa.PublicKey
	VerifyKeyFile string
	SignKey       *rsa.PrivateKey
	SignKeyFile   string
}

func (opts JWTOptions) validate() error {
	if opts.SignKey == nil && opts.SignKeyFile == "" {
		return ErrJWtSignKeyEmpty
	}

	if opts.VerifyKey == nil && opts.VerifyKeyFile == "" {
		return ErrJWtSignKeyEmpty
	}

	return nil
}

// JWTClient client
type JWTClient struct {
	Options   JWTOptions
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
}

// NewJWTClient init jwt client
func NewJWTClient(opt JWTOptions) (*JWTClient, error) {
	err := opt.validate()
	if err != nil {
		return nil, err
	}
	jwtCli := &JWTClient{Options: opt}

	// parse verifyKey
	if opt.VerifyKey != nil {
		jwtCli.verifyKey = opt.VerifyKey
	} else {
		verifyBytes, err := ioutil.ReadFile(opt.VerifyKeyFile)
		if err != nil {
			return nil, err
		}
		jwtCli.verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
		if err != nil {
			return nil, err
		}
	}

	// parse SignKey
	if opt.SignKey != nil {
		jwtCli.signKey = opt.SignKey
	} else if opt.SignKeyFile != "" {
		signBytes, err := ioutil.ReadFile(opt.SignKeyFile)
		if err != nil {
			return nil, err
		}
		jwtCli.signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
		if err != nil {
			return nil, err
		}
	}

	return jwtCli, nil
}

// JWTDecode decode jwt token, when token expired or decode err return UserClaimsInfo == nil & error
func (jc *JWTClient) JWTDecode(jwtToken string) (*UserClaimsInfo, error) {
	if jc == nil {
		return nil, ErrServerNotInited
	}

	token, err := jwt.ParseWithClaims(jwtToken, &UserClaimsInfo{}, func(token *jwt.Token) (interface{}, error) {
		return jc.verifyKey, nil
	})

	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, ErrTokenIsNil
	}

	if claims, ok := token.Claims.(*UserClaimsInfo); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

// JWTSign sign jwtToken
func (jc *JWTClient) JWTSign(user *UserInfo) (string, error) {
	if jc == nil {
		return "", ErrServerNotInited
	}

	err := user.validate()
	if err != nil {
		return "", err
	}

	// generate uer claims
	claimsPayload := &UserClaimsInfo{
		SubType:      user.SubType,
		UserName:     user.UserName,
		ClientID:     user.ClientName,
		ClientSecret: user.ClientSecret,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: func() int64 {
				if user.ExpiredTime == 0 {
					return 0
				}
				return time.Now().Add(time.Second * time.Duration(user.ExpiredTime)).Unix()
			}(),
			Issuer: user.Issuer,
		},
	}

	token := jwt.NewWithClaims(BCSJWTSigningMethod, claimsPayload)
	jwtToken, err := token.SignedString(jc.signKey)
	if err != nil {
		errMsg := fmt.Errorf("JWTSign jwtToken with claim (%v) failed: %v", claimsPayload, err)
		return "", errMsg
	}

	return jwtToken, nil
}
