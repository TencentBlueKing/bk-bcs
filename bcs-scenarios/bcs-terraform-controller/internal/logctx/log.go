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

package logctx

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog/glog"
)

const (
	TraceKey  = "t-trace"
	ObjectKey = "t-object"
)

var (
	keys = []string{TraceKey, ObjectKey}
)

func getKeyValueFromCtx(ctx context.Context) string {
	result := make([]string, 0)
	for i := range keys {
		trace := ctx.Value(keys[i])
		if trace != nil {
			result = append(result, keys[i]+"="+trace.(string))
		}
	}
	if len(result) == 0 {
		return ""
	}
	return ", " + strings.Join(result, ", ")
}

// Infof Info 级别日志
func Infof(ctx context.Context, format string, args ...interface{}) {
	glog.InfoDepth(1, fmt.Sprintf(format+getKeyValueFromCtx(ctx), args...))
}

// Warnf Warn 级别日志
func Warnf(ctx context.Context, format string, args ...interface{}) {
	glog.WarningDepth(1, fmt.Sprintf(format+getKeyValueFromCtx(ctx), args...))
}

// Errorf Error 级别日志
func Errorf(ctx context.Context, format string, args ...interface{}) {
	glog.ErrorDepth(1, fmt.Sprintf(format+getKeyValueFromCtx(ctx), args...))
}

// Fatalf Fatal 级别日志
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	glog.FatalDepth(1, fmt.Sprintf(format+getKeyValueFromCtx(ctx), args...))
}
