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

// Package proxy xxx
package proxy

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	grpcproto "google.golang.org/grpc/encoding/proto"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

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
	IsTencent bool
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

const (
	poTencentUserPrefix = "potencent_"
)

// GetJWTInfo from request
func GetJWTInfo(req *http.Request, client *jwt.JWTClient) (*UserInfo, error) {
	raw := req.Header.Get("Authorization")
	user, err := GetJWTInfoWithAuthorization(raw, client)
	if err != nil {
		return nil, errors.Wrapf(err, "get authorization user failed")
	}
	if common.IsAdminUser(user.ClientID) {
		userName := req.Header.Get(common.HeaderBKUserName)
		user.UserName = userName
	}
	user.IsTencent = true
	if strings.HasPrefix(user.GetUser(), poTencentUserPrefix) {
		user.UserName = strings.TrimPrefix(user.UserName, poTencentUserPrefix)
		user.IsTencent = false
	}
	return user, nil
}

// GetJWTInfoWithAuthorization 根据 token 获取用户信息
func GetJWTInfoWithAuthorization(authorization string, client *jwt.JWTClient) (*UserInfo, error) {
	if len(authorization) == 0 {
		return nil, fmt.Errorf("lost 'Authorization' header")
	}
	if !strings.HasPrefix(authorization, "Bearer ") {
		return nil, fmt.Errorf("hader 'Authorization' malform")
	}
	token := strings.TrimPrefix(authorization, "Bearer ")
	claim, err := client.JWTDecode(token)
	if err != nil {
		return nil, err
	}
	u := &UserInfo{UserClaimsInfo: claim}
	if u.GetUser() == "" {
		return nil, fmt.Errorf("lost user information")
	}
	return u, nil
}

// IsGitOpsClient check if request comes from bcs-gitops client,
// only use for gitops command line
func IsGitOpsClient(req *http.Request) bool {
	token := req.Header.Get(common.HeaderBCSClient)
	return token == common.ServiceNameShort
}

// JSONResponse convenient tool for response
func JSONResponse(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	content, _ := json.Marshal(obj)
	fmt.Fprintln(w, string(content))
}

// DirectlyResponse 对象本身就是个字符串，直接写入返回
func DirectlyResponse(w http.ResponseWriter, obj interface{}) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, obj)
}

var (
	httpStatusCode = map[int]codes.Code{
		http.StatusOK:                  codes.OK,
		http.StatusBadRequest:          codes.InvalidArgument,
		http.StatusNotFound:            codes.NotFound,
		http.StatusUnauthorized:        codes.Unauthenticated,
		http.StatusInternalServerError: codes.Internal,
		http.StatusServiceUnavailable:  codes.Unavailable,
	}
)

func returnGrpcCode(statusCode int) codes.Code {
	v, ok := httpStatusCode[statusCode]
	if !ok {
		return codes.Unknown
	}
	return v
}

// GRPCErrorResponse 返回 gRPC 错误给客户端
func GRPCErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	grpcCode := returnGrpcCode(statusCode)
	w.Header().Set("Content-Type", "application/grpc+proto")
	w.Header().Set("grpc-status", strconv.Itoa(int(grpcCode)))
	w.Header().Set("grpc-message", err.Error())
	w.WriteHeader(http.StatusOK)

	// Write the error message as the response body
	_, _ = w.Write(nil)
}

var (
	// nolint
	grpcSuffixBytes = []byte{128, 0, 0, 0, 54, 99, 111, 110, 116, 101, 110, 116, 45, 116, 121, 112, 101, 58, 32, 97,
		112, 112, 108, 105, 99, 97, 116, 105, 111, 110, 47, 103, 114, 112, 99, 43, 112, 114, 111, 116, 111, 13, 10,
		103, 114, 112, 99, 45, 115, 116, 97, 116, 117, 115, 58, 32, 48, 13, 10}
)

// GRPCResponse 将对象通过 grpc 的数据格式返回
// grpc 返回的 proto 数据格式规定如下：
//   - 第 1 个 byte 表明是否是 compressed, 参见: google.golang.org/grpc/rpc_util.go 中的 compressed
//   - 第 2-5 个 byte 表明 body 的长度，参见: google.golang.org/grpc/rpc_util.go 的 recvMsg 方法
//     长度需要用到大端转换来获取实际值
//   - 后续的 byte 位是 body + content-type
//
// 在获取到 body 字节后，可以通过 grpc.encoding 来反序列化
func GRPCResponse(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/grpc+proto")
	w.Header().Set("grpc-status", "0")
	w.WriteHeader(http.StatusOK)
	bs, err := encoding.GetCodec(grpcproto.Name).Marshal(obj)
	if err != nil {
		blog.Errorf("grpc proto encoding marshal failed: %s", err.Error())
		_, _ = w.Write([]byte{})
		return
	}
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(bs)))
	result := make([]byte, 0, 5+len(bs)+len(grpcSuffixBytes))
	// grpc 返回的第一个 byte 表明是否是 compressed
	// 参见: google.golang.org/grpc/rpc_util.go 中的 compressionNone
	result = append(result, 0)
	result = append(result, header...)
	result = append(result, bs...)
	result = append(result, grpcSuffixBytes...)
	_, _ = w.Write(result)
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
