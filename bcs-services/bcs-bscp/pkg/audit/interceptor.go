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

package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"k8s.io/klog/v2"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

// UnaryServerInterceptor is a grpc interceptor that add audit.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		resp interface{}, err error) {
		fn, ok := auditGrpcMap[info.FullMethod]
		if !ok {
			klog.Warningf("no audit for grpc method:%s", info.FullMethod)
			resp, err = handler(ctx, req)
			return resp, err
		}
		res, act := fn()
		st := time.Now()

		defer func() {
			e := rest.GRPCErr(err)
			msg := "Success"
			if e.HTTPStatusCode != http.StatusOK {
				msg = e.Error.Message
			}
			var input map[string]any
			js, _ := json.Marshal(req)
			_ = json.Unmarshal(js, &input)
			res.ResourceData = input
			kt := kit.FromGrpcContext(ctx)

			p := auditParam{
				Username:     kt.User,
				SourceIP:     getClientIP(ctx),
				UserAgent:    getUserAgent(ctx),
				Rid:          kt.Rid,
				Resource:     res,
				Action:       act,
				StartTime:    st,
				ResultStatus: e.HTTPStatusCode,
				ResultMsg:    msg,
			}
			addAudit(p)
		}()

		resp, err = handler(ctx, req)
		return resp, err
	}
}

func getClientIP(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	xff := md.Get("x-forwarded-for")
	if len(xff) > 0 {
		return xff[0]
	}

	xri := md.Get("x-real-ip")
	if len(xri) > 0 {
		return xri[0]
	}

	return ""
}

func getUserAgent(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	gua := md.Get("grpcgateway-user-agent")
	if len(gua) > 0 {
		return gua[0]
	}

	ua := md.Get("user-agent")
	if len(ua) > 0 {
		return ua[0]
	}

	return ""
}
