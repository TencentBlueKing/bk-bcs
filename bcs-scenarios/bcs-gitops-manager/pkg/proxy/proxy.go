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

package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// GitOpsOptions for revese proxy
type GitOpsOptions struct {
	// backend gitops kubernetes service and port
	Service string
	// URL prefix like /gitopsmanager/proxy/
	PathPrefix string
	// storage interface for access gitops data
	Storage store.Store
	// JWTClient for authentication
	JWTDecoder *jwt.JWTClient
	// IAMClient is basic client
	IAMClient iam.PermClient
}

// Validate options
func (opt *GitOpsOptions) Validate() error {
	if len(opt.Service) == 0 {
		return fmt.Errorf("lost gitops system information")
	}
	if opt.Storage == nil {
		return fmt.Errorf("lost gitops storage access")
	}
	return nil
}

// GitOpsProxy definition for all kinds of
// gitops solution
type GitOpsProxy interface {
	http.Handler
	// Init proxy
	Init() error
}

// UserInfo for token validate
type UserInfo struct {
	*jwt.UserClaimsInfo
}

// GetUser string
func (user *UserInfo) GetUser() string {
	if len(user.UserName) != 0 {
		return user.UserName
	}
	if len(user.ClientID) != 0 {
		return user.ClientID
	}
	return ""
}

// GetJWTInfo from request
func GetJWTInfo(req *http.Request, client *jwt.JWTClient) (*UserInfo, error) {
	raw := req.Header.Get("Authorization")
	if len(raw) == 0 {
		blog.Errorf("request %s header %+v", req.URL.Path, req.Header)
		return nil, fmt.Errorf("lost Authorization")
	}
	if !strings.HasPrefix(raw, "Bearer ") {
		return nil, fmt.Errorf("Authorization malform")
	}
	token := strings.TrimPrefix(raw, "Bearer ")
	claim, err := client.JWTDecode(token)
	if err != nil {
		return nil, err
	}
	u := &UserInfo{claim}
	if u.GetUser() == "" {
		return nil, fmt.Errorf("lost user information")
	}
	return u, nil
}

// IsAdmin check if request comes from admin,
// only use for gitops command line
func IsAdmin(req *http.Request) bool {
	token := req.Header.Get(common.HeaderBCSClient)
	// todo(DeveloperJim): fix me
	return token == common.HeaderBCSClient
}

// JSONResponse convenient tool for response
func JSONResponse(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	content, _ := json.Marshal(obj)
	fmt.Fprintln(w, string(content))
}

// BUG21955Workaround ! copy from argocd
type BUG21955Workaround struct {
	Handler http.Handler
}

// Workaround for https://github.com/golang/go/issues/21955 to support escaped URLs in URL path.
var pathPatters = []*regexp.Regexp{
	regexp.MustCompile(`/api/v1/clusters/[^/]+`),
	regexp.MustCompile(`/api/v1/repositories/[^/]+`),
	regexp.MustCompile(`/api/v1/repocreds/[^/]+`),
	regexp.MustCompile(`/api/v1/repositories/[^/]+/apps`),
	regexp.MustCompile(`/api/v1/repositories/[^/]+/apps/[^/]+`),
	regexp.MustCompile(`/settings/clusters/[^/]+`),
}

// ServeHTTP implementation
func (work *BUG21955Workaround) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	blog.Infof("proxy %s RequestURI %s, header: %+v", r.Method, r.URL.RequestURI(), r.Header)
	for _, pattern := range pathPatters {
		if pattern.MatchString(r.URL.RawPath) {
			r.URL.Path = r.URL.RawPath
			blog.Warnf("proxy URL RawPath fix %s", r.URL.RawPath)
			break
		}
	}
	work.Handler.ServeHTTP(w, r)
}
