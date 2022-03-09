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

package auth

import (
	"encoding/base64"
	"net/http"
)

const (
	headerUID = "X-BKREPO-UID"
)

// New return a new Auth instance
//   authType: Basic or Platform
//   uid: user id
//   username: auth username
//   password: auth password
func New(authType, uid, username, password string) *Auth {
	if authType == "" {
		authType = "Basic"
	}

	return &Auth{
		authType: authType,
		uid:      uid,
		username: username,
		password: password,
	}
}

// Auth handle the auth mechanism of bk-repo
type Auth struct {
	authType string
	uid      string
	upwd     string
	username string
	password string

	// cache for token
	token string
}

// GetAuthToken return the Authorization token for bk-repo
// format:
// authType base64(username:password)
func (a *Auth) GetAuthToken() string {
	if a.token == "" {
		a.token = a.authType + " " + base64.StdEncoding.EncodeToString([]byte(a.username+":"+a.password))
	}

	return a.token
}

// SetHeader set the auth-relative http headers before request to bk-repo
// by default we need:
//   Authorization: token
//   X-BKREPO-UID: uid
func (a *Auth) SetHeader(header http.Header) {
	header.Set("Authorization", a.GetAuthToken())
	header.Set(headerUID, a.uid)
}
