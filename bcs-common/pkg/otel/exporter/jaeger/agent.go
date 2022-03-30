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

import (
	"log"
	"time"

	"go.opentelemetry.io/otel/exporters/jaeger"
)

// AgentEndpoint configs jaeger agent endpoint
type AgentEndpoint struct {
	Host         string                       `json:"jaegerAgentHost,omitempty" value:"" usage:"host to be used in the agent client endpoint"`
	Port         string                       `json:"JaegerAgentPort,omitempty" value:"" usage:"port to be used in the agent client endpoint"`
	AgentOptions []jaeger.AgentEndpointOption `json:"-"`
}

// WithAgentEndpoint configures the Jaeger exporter to send spans to a Jaeger agent
// over compact thrift protocol. This will use the following environment variables for
// configuration if no explicit option is provided:
//
// - OTEL_EXPORTER_JAEGER_AGENT_HOST is used for the agent address host
// - OTEL_EXPORTER_JAEGER_AGENT_PORT is used for the agent address port
//
// The passed options will take precedence over any environment variables and default values
// will be used if neither are provided.
func WithAgentEndpoint(options ...jaeger.AgentEndpointOption) jaeger.EndpointOption {
	return jaeger.WithAgentEndpoint(options...)
}

// WithAgentHost sets a host to be used in the agent client endpoint.
// This option overrides any value set for the
// OTEL_EXPORTER_JAEGER_AGENT_HOST environment variable.
// If this option is not passed and the env var is not set, "localhost" will be used by default.
func WithAgentHost(host string) jaeger.AgentEndpointOption {
	return jaeger.WithAgentHost(host)
}

// WithAgentPort sets a port to be used in the agent client endpoint.
// This option overrides any value set for the
// OTEL_EXPORTER_JAEGER_AGENT_PORT environment variable.
// If this option is not passed and the env var is not set, "6831" will be used by default.
func WithAgentPort(port string) jaeger.AgentEndpointOption {
	return jaeger.WithAgentPort(port)
}

// WithLogger sets a logger to be used by agent client.
func WithLogger(logger *log.Logger) jaeger.AgentEndpointOption {
	return jaeger.WithLogger(logger)
}

// WithDisableAttemptReconnecting sets option to disable reconnecting udp client.
func WithDisableAttemptReconnecting() jaeger.AgentEndpointOption {
	return jaeger.WithDisableAttemptReconnecting()
}

// WithAttemptReconnectingInterval sets the interval between attempts to re resolve agent endpoint.
func WithAttemptReconnectingInterval(interval time.Duration) jaeger.AgentEndpointOption {
	return jaeger.WithAttemptReconnectingInterval(interval)
}

// WithMaxPacketSize sets the maximum UDP packet size for transport to the Jaeger agent.
func WithMaxPacketSize(size int) jaeger.AgentEndpointOption {
	return jaeger.WithMaxPacketSize(size)
}
