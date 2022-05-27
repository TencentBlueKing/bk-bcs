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

package otlpgrpctrace

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
)

// GRPCConfig sets the OTLP collector GRPC endpoint
type GRPCConfig struct {
	GRPCEndpoint string                 `json:"grpcEndpoint,omitempty" usage:"grpcEndpoint sets GRPC client endpoint"`
	GRPCURLPath  string                 `json:"grpcURLPath,omitempty" usage:"grpcURLPath sets GRPC client endpoint"`
	GRPCInsecure bool                   `json:"grpcInsecure,omitempty" usage:"grpcInsecure disables GRPC client transport security"`
	GRPCOptions  []otlptracegrpc.Option `json:"-"`
}

// New constructs a new Exporter and starts it.
func New(ctx context.Context, client otlptrace.Client) (*otlptrace.Exporter, error) {
	return otlptrace.New(ctx, client)
}

// NewUnstarted constructs a new Exporter and does not start it.
func NewUnstarted(client otlptrace.Client) *otlptrace.Exporter {
	return otlptrace.NewUnstarted(client)
}
