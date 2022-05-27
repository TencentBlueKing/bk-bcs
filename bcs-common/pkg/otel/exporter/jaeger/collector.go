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
	"net/http"

	"go.opentelemetry.io/otel/exporters/jaeger"
)

// CollectorEndpoint configs jaeger agent endpoint
type CollectorEndpoint struct {
	Endpoint         string                           `json:"endpoint,omitempty" usage:"endpoint for sending spans directly to a collector"`
	Username         string                           `json:"username,omitempty" usage:"username to be used for authentication with the collector endpoint"`
	Password         string                           `json:"password,omitempty" usage:"password to be used for authentication with the collector endpoint"`
	CollectorOptions []jaeger.CollectorEndpointOption `json:"-"`
}

// WithCollectorEndpoint defines the full URL to the Jaeger HTTP Thrift collector. This will
// use the following environment variables for configuration if no explicit option is provided:
//
// - OTEL_EXPORTER_JAEGER_ENDPOINT is the HTTP endpoint for sending spans directly to a collector.
// - OTEL_EXPORTER_JAEGER_USER is the username to be sent as authentication to the collector endpoint.
// - OTEL_EXPORTER_JAEGER_PASSWORD is the password to be sent as authentication to the collector endpoint.
//
// The passed options will take precedence over any environment variables.
// If neither values are provided for the endpoint, the default value of "http://localhost:14268/api/traces" will be used.
// If neither values are provided for the username or the password, they will not be set since there is no default.
func WithCollectorEndpoint(options ...jaeger.CollectorEndpointOption) jaeger.EndpointOption {
	return jaeger.WithCollectorEndpoint(options...)
}

// WithEndpoint is the URL for the Jaeger collector that spans are sent to.
// This option overrides any value set for the
// OTEL_EXPORTER_JAEGER_ENDPOINT environment variable.
// If this option is not passed and the environment variable is not set,
// "http://localhost:14268/api/traces" will be used by default.
func WithEndpoint(endpoint string) jaeger.CollectorEndpointOption {
	return jaeger.WithEndpoint(endpoint)
}

// WithUsername sets the username to be used in the authorization header sent for all requests to the collector.
// This option overrides any value set for the
// OTEL_EXPORTER_JAEGER_USER environment variable.
// If this option is not passed and the environment variable is not set, no username will be set.
func WithUsername(username string) jaeger.CollectorEndpointOption {
	return jaeger.WithUsername(username)
}

// WithPassword sets the password to be used in the authorization header sent for all requests to the collector.
// This option overrides any value set for the
// OTEL_EXPORTER_JAEGER_PASSWORD environment variable.
// If this option is not passed and the environment variable is not set, no password will be set.
func WithPassword(password string) jaeger.CollectorEndpointOption {
	return jaeger.WithPassword(password)
}

// WithHTTPClient sets the http client to be used to make request to the collector endpoint.
func WithHTTPClient(client *http.Client) jaeger.CollectorEndpointOption {
	return jaeger.WithHTTPClient(client)
}
