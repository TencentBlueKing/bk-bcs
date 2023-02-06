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

package auth

import (
	"errors"
	"net/http"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/components/clientset"
	pbas "bscp.io/pkg/protocol/auth-server"
)

type auth struct {
	client *clientset.ClientSet
}

func NewAuth(client *clientset.ClientSet) *auth {
	return &auth{client: client}
}

func getUserCredentialFromCookies(r *http.Request) (*pbas.UserCredentialReq, error) {
	provider := cc.AuthServer().LoginAuth.Provider
	req := &pbas.UserCredentialReq{Provider: provider}

	switch provider {
	case "BK_PAAS":
		token, err := r.Cookie("bk_token")
		if err != nil {
			return nil, err
		}
		req.Token = token.Value
	case "BK_LOGIN":
		uid, err := r.Cookie("bk_uid")
		if err != nil {
			return nil, err
		}
		token, err := r.Cookie("bk_ticket")
		if err != nil {
			return nil, err
		}
		req.Uid = uid.Value
		req.Token = token.Value
	default:
		return nil, errors.New("provider not supported")
	}
	return req, nil
}

// LoginAuthentication 登入鉴权中间件
func (a *auth) LoginAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := getUserCredentialFromCookies(r)
		if err != nil || req.Token == "" {
			http.Redirect(w, r, "", http.StatusFound)
			return
		}

		user, err := a.client.AS.GetUserInfo(r.Context(), req)
		if err != nil || user.GetUsername() == "" {
			http.Redirect(w, r, "", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
