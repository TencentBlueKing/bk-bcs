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

package bkiam

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/codec"
	"bk-bcs/bcs-common/common/encrypt"
	"bk-bcs/bcs-services/bcs-api/auth"
	"bk-bcs/bcs-services/bcs-api/config"

	"github.com/dgrijalva/jwt-go"
)

const (
	ApiGwJWTKey     = "X-Bkapi-JWT"
	BcsUserTokenKey = "X-Bcs-User-Token"
)

func NewAuth(conf *config.ApiServConfig) (auth.BcsAuth, error) {
	if !conf.BKIamAuth.Auth {
		return &Auth{disabled: true}, nil
	}

	cache, err := NewTokenCache(conf)
	if err != nil {
		blog.Errorf("NewAuth get new token cache failed: %v", err)
		return nil, err
	}

	client, err := NewClient(conf)
	if err != nil {
		blog.Errorf("NewAuth get new client failed: %v", err)
		return nil, err
	}

	rsaCert, err := parseRSAFromFile(conf.BKIamAuth.ApiGwRsaFile)
	if err != nil {
		blog.Errorf("NewAuth parse RSA cert from file failed: %v", err)
		return nil, err
	}

	whitelist := make(map[string]bool)
	for _, raw := range conf.BKIamAuth.BKIamTokenWhiteList {
		wl, err := encrypt.DesDecryptFromBase([]byte(raw))
		if err != nil {
			blog.Errorf("decode from token whitelist(%s) failed: %v", raw, err)
			continue
		}

		whitelist[string(wl)] = true
	}

	return &Auth{
		cache:     cache,
		client:    client,
		cert:      rsaCert,
		whitelist: whitelist,
	}, nil
}

// Auth manage the authority check with bk-iam,
type Auth struct {
	disabled bool

	cert *rsa.PublicKey

	client *Client
	cache  *TokenCache

	whitelist map[string]bool
}

func (a *Auth) GetToken(header http.Header) (*auth.Token, error) {
	if a.disabled {
		return &auth.Token{}, nil
	}

	// userToken specified
	userToken := header.Get(BcsUserTokenKey)
	if userToken != "" {
		// whitelist for token
		if _, ok := a.whitelist[userToken]; ok {
			return &auth.Token{Token: userToken}, nil
		}

		token, err := a.cache.Get(userToken)
		if err != nil {
			blog.Errorf("GetToken get from cache failed: %v, userToken: %s", err, userToken)
			return nil, err
		}

		return token, nil
	}

	// username in jwt from api gateway
	jwtRaw := header.Get(ApiGwJWTKey)
	if jwtRaw == "" {
		blog.Errorf("GetToken user token and api gateway jwt are both empty")
		return nil, fmt.Errorf("user token and api gateway jwt are both empty")
	}
	data, err := parseJWT(jwtRaw, a.cert)
	if err != nil {
		blog.Errorf("GetToken parse jwt failed: %v, jwt: %s", err, jwtRaw)
		return nil, err
	}

	return &auth.Token{
		Username: data.User.Username,
	}, nil
}

func (a *Auth) Allow(token *auth.Token, action auth.Action, resource auth.Resource) (bool, error) {
	if a.disabled {
		return true, nil
	}

	// whitelist for token
	if _, ok := a.whitelist[token.Token]; ok {
		return true, nil
	}

	if token.Username == "" {
		blog.Errorf("bkiam auth get a empty username")
		return false, fmt.Errorf("get a empty username")
	}

	// update the token
	go a.cache.Update(token)

	ok, err := a.client.Query(token.Username, action, resource)
	if err != nil {
		blog.Errorf("bkiam auth use client query failed: %v, username: %s", err, token.Username)
		return false, err
	}
	return ok, nil
}

func parseJWT(myToken string, myKey *rsa.PublicKey) (*ApiGwData, error) {
	token, err := jwt.Parse(myToken, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})

	if err != nil || !token.Valid {
		if err == nil {
			err = fmt.Errorf("token is invalid")
		}
		return nil, err
	}

	var data []byte
	if err = codec.EncJson(token.Claims, &data); err != nil {
		return nil, err
	}

	fmt.Printf("%s\n", string(data))
	var respData ApiGwData
	if err = codec.DecJson(data, &respData); err != nil {
		return nil, err
	}

	return &respData, nil
}

func parseRSAFromFile(filePath string) (*rsa.PublicKey, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	var rsaRaw []byte
	if rsaRaw, err = ioutil.ReadAll(f); err != nil {
		return nil, err
	}

	block, errByte := pem.Decode(rsaRaw)
	if block == nil {
		return nil, fmt.Errorf("pem data no found: %s", string(errByte))
	}

	pri, _ := x509.ParsePKIXPublicKey(block.Bytes)
	key, ok := pri.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("data parsed is no rsa public key")
	}

	return key, nil
}
