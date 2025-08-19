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

package tracing

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace"
	"go.opentelemetry.io/otel/attribute"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// ServiceName 不同模块的区分
const serviceName = "bcs-platform-manager"

// InitTracing init tracing
func InitTracing(op *config.TracingConf) (func(context.Context) error, error) {
	if !op.Enabled {
		return nil, nil
	}
	opts := []trace.Option{}

	if op.Endpoint != "" {
		opts = append(opts, trace.OTLPEndpoint(op.Endpoint))
	}
	attrs := make([]attribute.KeyValue, 0)

	if op.Token != "" {
		attrs = append(attrs, attribute.String("bk.data.token", op.Token))
	}

	if op.ResourceAttrs != nil {
		attrs = append(attrs, newResource(op.ResourceAttrs)...)
	}

	opts = append(opts, trace.ResourceAttrs(attrs))

	tracer, err := trace.InitTracingProvider(serviceName, opts...)
	if err != nil {
		return nil, err
	}

	return tracer, nil
}

func newResource(attrs map[string]string) []attribute.KeyValue {
	attrValues := make([]attribute.KeyValue, 0)
	for k, v := range attrs {
		attrValues = append(attrValues, attribute.String(k, v))
	}
	return attrValues
}
