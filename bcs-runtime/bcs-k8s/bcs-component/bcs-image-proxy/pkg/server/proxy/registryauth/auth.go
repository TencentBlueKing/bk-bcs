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

// Package registryauth xxx
package registryauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils"
)

// AuthToken defines the authed token after registry-auth
type AuthToken struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

// AuthRequest defines the auth request
type AuthRequest struct {
	Realm   string `json:"realm"`
	Service string `json:"service"`
	Scope   string `json:"scope"`
}

var (
	scopeRegexp = regexp.MustCompile(`^repository:(.*):.*`)
)

// ParseRepoFromScope parse the scope. e.g.: repository:library/centos:pull => library/centos
func ParseRepoFromScope(scope string) string {
	res := scopeRegexp.FindStringSubmatch(scope)
	if len(res) != 2 {
		return ""
	}
	return res[1]
}

// ParseAuthRequest parse the auth request header
func ParseAuthRequest(authHeader string) (*AuthRequest, error) {
	authenticate := strings.TrimSpace(authHeader)
	if authenticate == "" || !strings.HasPrefix(authenticate, "Bearer realm") {
		return nil, errors.Errorf("response header not have authenticate header")
	}
	realm, service, scope := parseAuthenticateHeader(authenticate)
	return &AuthRequest{
		Realm:   realm,
		Service: service,
		Scope:   scope,
	}, nil
}

var (
	registryAuthLock sync.RWMutex
)

// HandleRegistryUnauthorized upstream registry will responseCode 401 if not have bearerToken.
// We should initiative auth to registry and get the bearerToken.
// e.g: Www-Authenticate: Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:samalba/my-app:pull,push"
// refer to: https://docker-docs.uclv.cu/registry/spec/auth/token/#how-to-authenticate
func HandleRegistryUnauthorized(ctx context.Context, authReq *AuthRequest, registry *options.RegistryMapping) (
	*AuthToken, error) {
	authSlice := make([][]string, 0)
	if registry.CorrectUser != "" && registry.CorrectPass != "" {
		authSlice = append(authSlice, []string{registry.CorrectUser, registry.CorrectPass})
	}
	if registry.Username != "" && registry.Password != "" {
		authSlice = append(authSlice, []string{registry.Username, registry.Password})
	}
	for _, user := range registry.Users {
		authSlice = append(authSlice, []string{user.Username, user.Password})
	}
	if len(authSlice) == 0 {
		logctx.Warnf(ctx, "registry '%s' not have auth-user, but we still do auth", registry.OriginalHost)
		return handleRegistryAuth(ctx, authReq, "", "")
	}

	var result *AuthToken
	var err error
	for _, item := range authSlice {
		result, err = handleRegistryAuth(ctx, authReq, item[0], item[1])
		if err == nil {
			registryAuthLock.Lock()
			registry.CorrectUser = item[0]
			registry.CorrectPass = item[1]
			registryAuthLock.Unlock()
			return result, nil
		}
		logctx.Warnf(ctx, "handle registry auth with user '%s' and pass '%s' failed: %s",
			item[0], item[1], err.Error())
	}
	if result == nil {
		return nil, errors.Errorf("registry '%s' auth failed after %d times", registry.OriginalHost,
			len(authSlice))
	}
	return result, nil
}

func handleRegistryAuth(ctx context.Context, authReq *AuthRequest, user, passwd string) (*AuthToken, error) {
	queryParams := make(map[string]string)
	if authReq.Service != "" {
		queryParams["service"] = authReq.Service
	}
	if authReq.Scope != "" {
		queryParams["scope"] = authReq.Scope
	}
	header := make(map[string]string)
	if user != "" && passwd != "" {
		header["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
			user, passwd)))
	}
	authRespBody, err := utils.SendHTTPRequest(ctx, &utils.HTTPRequest{
		Url:         authReq.Realm,
		Method:      http.MethodGet,
		QueryParams: queryParams,
		Header:      header,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "handle unauth request failed")
	}
	result := new(AuthToken)
	if err = json.Unmarshal(authRespBody, result); err != nil {
		return nil, errors.Wrapf(err, "handle unauth request unmarshal response body failed")
	}
	if result.Token == "" {
		return nil, errors.Errorf("handle auth request response token is empty")
	}
	return result, nil
}

var (
	realmRegex   = regexp.MustCompile(`realm="(.*?)"`)
	serviceRegex = regexp.MustCompile(`service="(.*?)"`)
	scopeRegex   = regexp.MustCompile(`scope="(.*?)"`)
)

func parseAuthenticateHeader(header string) (string, string, string) {
	realm := realmRegex.FindStringSubmatch(header)
	service := serviceRegex.FindStringSubmatch(header)
	scope := scopeRegex.FindStringSubmatch(header)

	var realmValue, serviceValue, scopeValue string
	if len(realm) > 1 {
		realmValue = realm[1]
	}
	if len(service) > 1 {
		serviceValue = service[1]
	}
	if len(scope) > 1 {
		scopeValue = scope[1]
	}
	return realmValue, serviceValue, scopeValue
}
