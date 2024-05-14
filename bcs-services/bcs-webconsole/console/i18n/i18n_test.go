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

package i18n

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func TestT(t *testing.T) {

	type args struct {
		tags   string
		format string
		values interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "zh-hans_服务请求失败",
			args: args{tags: "zh-hans", format: "服务请求失败: %s", values: errors.New("失败")},
			want: "服务请求失败: 失败",
		},
		{
			name: "en_服务请求失败",
			args: args{tags: "en", format: "服务请求失败: %s", values: errors.New("失败")},
			want: "Service request failed: 失败",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &gin.Context{
				Request: &http.Request{
					URL: &url.URL{
						RawQuery: fmt.Sprintf("lang=%s", tt.args.tags),
					},
				},
			}

			if got := T(ctx, tt.args.format, tt.args.values); got != tt.want {
				t.Errorf("T() = %v, want %v", got, tt.want)
			}
		})
	}
}
