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

	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	"bscp.io/pkg/runtime/jsoni"
	"k8s.io/klog/v2"

	gprm "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

// UnaryServerInterceptorWithMetrics returns default grpc interceptor with metrics
func UnaryServerInterceptorWithMetrics(mc *gprm.ServerMetrics) grpc.UnaryServerInterceptor {
	// EnableHandlingTimeHistogram enables grpc server's histograms
	mc.EnableHandlingTimeHistogram(metrics.GrpcBuckets)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		resp interface{}, err error) {

		mi := mc.UnaryServerInterceptor()

		return mi(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {

			kt := kit.FromGrpcContext(ctx)
			defer func() {
				if fatalErr := recover(); fatalErr != nil {
					logs.Errorf("[bscp server panic], err: %v, rid: %s, debug strace: %s", fatalErr, kt.Rid,
						debug.Stack())
					logs.CloseLogs()
				}
			}()

			if logs.V(4) {
				js, _ := jsoni.Marshal(req)
				logs.Infof("request method: %v, app_code: %s, user: %s, req: %s, rid: %s",
					info.FullMethod, kt.AppCode, kt.User, string(js), kt.Rid)
			}

			resp, err := handler(ctx, req)

			if logs.V(5) {
				logs.Infof("resp: %v, err: %v, rid: %s", resp, err, kt.Rid)
			}

			return resp, err
		})
	}
}

// LogUnaryServerInterceptor 添加请求日志
func LogUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		st := time.Now()
		kt := kit.FromGrpcContext(ctx)
		service := path.Dir(info.FullMethod)[1:]
		method := path.Base(info.FullMethod)

		defer func() {
			klog.InfoS("grpc", "rid", kt.Rid, "system", "grpc", "span.kind", "grpc.service", "service", service, "method", method, "grpc.duration", time.Since(st))
		}()

		resp, err = handler(ctx, req)
		return resp, err
	}
}
