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

package otlphttptrace

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

// HTTPConfig sets the OTLP collector HTTP endpoint
type HTTPConfig struct {
	HTTPEndpoint string                 `json:"httpEndpoint,omitempty" usage:"httpEndpoint sets HTTP client endpoint"`
	HTTPURLPath  string                 `json:"httpURLPath,omitempty" usage:"httpURLPath sets HTTP client endpoint"`
	HTTPInsecure bool                   `json:"httpInsecure,omitempty" usage:"httpInsecure disables HTTP client transport security"`
	HTTPOptions  []otlptracehttp.Option `json:"-"`
}

// New constructs a new Exporter and starts it.
func New(ctx context.Context, opts ...otlptracehttp.Option) (*otlptrace.Exporter, error) {
	return otlptracehttp.New(ctx, opts...)
}

// NewUnstarted constructs a new Exporter and does not start it.
func NewUnstarted(client otlptrace.Client) *otlptrace.Exporter {
	return otlptrace.NewUnstarted(client)
}
