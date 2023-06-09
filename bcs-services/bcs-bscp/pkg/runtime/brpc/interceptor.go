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

package brpc

import (
	"context"
	"path"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"k8s.io/klog/v2"
)

// RecoveryHandlerFuncContext 异常日志输出
func RecoveryHandlerFuncContext(ctx context.Context, p interface{}) (err error) {
	kt := kit.FromGrpcContext(ctx)
	logs.Errorf("[bscp server panic], err: %v, rid: %s, debug strace: %s", p, kt.Rid,
		debug.Stack())
	logs.CloseLogs()

	return status.Errorf(codes.Internal, "%v", p)
}

// LogUnaryServerInterceptor 添加请求日志
func LogUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		st := time.Now()
		kt := kit.FromGrpcContext(ctx)
		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)

		defer func() {
			klog.InfoS("grpc", "rid", kt.Rid, "system", "grpc", "span.kind", "grpc.service", "service", service, "method", method, "grpc.duration", time.Since(st), "error", err)
		}()

		resp, err = handler(ctx, req)
		return resp, err
	}
}
