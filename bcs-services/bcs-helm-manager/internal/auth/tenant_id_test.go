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

package auth

import (
	"context"
	"testing"

	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"go-micro.dev/v4/metadata"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
)

func TestTenantIDFromCtx(t *testing.T) {
	t.Parallel()

	withUser := func(ctx context.Context, tenantID string) context.Context {
		return context.WithValue(ctx, middleauth.AuthUserKey, middleauth.AuthUser{TenantId: tenantID})
	}

	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "header wins over context and jwt",
			ctx: func() context.Context {
				ctx := context.WithValue(context.Background(), contextx.TenantIDContextKey, "from-ctx")
				ctx = withUser(ctx, "from-jwt")
				md := metadata.Metadata{}
				md.Set(contextx.TenantIDHeaderKey, "from-header")
				return metadata.NewContext(ctx, md)
			}(),
			want: "from-header",
		},
		{
			name: "context wins over jwt",
			ctx:  withUser(context.WithValue(context.Background(), contextx.TenantIDContextKey, "from-ctx"), "from-jwt"),
			want: "from-ctx",
		},
		{
			name: "jwt when context empty",
			ctx:  withUser(context.Background(), "from-jwt"),
			want: "from-jwt",
		},
		{
			name: "default when nothing set",
			ctx:  context.Background(),
			want: contextx.DefaultTenantId,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := TenantIDFromCtx(tt.ctx); got != tt.want {
				t.Fatalf("TenantIDFromCtx() = %q, want %q", got, tt.want)
			}
		})
	}
}
