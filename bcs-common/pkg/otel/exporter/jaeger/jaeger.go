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

package jaeger

import "go.opentelemetry.io/otel/exporters/jaeger"

type EndpointConfig struct {
	CollectorEndpoint *CollectorEndpoint `json:"collectorEndpointConfig,omitempty"`
	AgentEndpoint     *AgentEndpoint     `json:"AgentClientConfig,omitempty"`
}

// NewCollectorExporter returns an OTel Exporter implementation that exports the collected
// spans to Jaeger collector.
func NewCollectorExporter(option ...jaeger.CollectorEndpointOption) (*jaeger.Exporter, error) {
	return jaeger.New(jaeger.WithCollectorEndpoint(option...))
}

// NewAgentExporter returns an OTel Exporter implementation that exports the collected
// spans to Jaeger agent.
func NewAgentExporter(option ...jaeger.AgentEndpointOption) (*jaeger.Exporter, error) {
	return jaeger.New(jaeger.WithAgentEndpoint(option...))
}
