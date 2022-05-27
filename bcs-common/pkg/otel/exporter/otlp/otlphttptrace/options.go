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
	"crypto/tls"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

// WithInsecure disables client transport security for the exporter's gRPC connection
// just like grpc.WithInsecure() https://pkg.go.dev/google.golang.org/grpc#WithInsecure
// does. Note, by default, client security is required unless WithInsecure is used.
func WithInsecure() otlptracehttp.Option {
	return otlptracehttp.WithInsecure()
}

// WithEndpoint allows one to set the endpoint that the exporter will
// connect to the collector on. If unset, it will instead try to use
// connect to DefaultCollectorHost:DefaultCollectorPort.
func WithEndpoint(endpoint string) otlptracehttp.Option {
	return otlptracehttp.WithEndpoint(endpoint)
}

// WithCompressoion will set the compressor for the gRPC client to use when sending requests.
// It is the responsibility of the caller to ensure that the compressor set has been registered
// with google.golang.org/grpc/encoding. This can be done by encoding.RegisterCompressor. Some
// compressors auto-register on import, such as gzip, which can be registered by calling
// `import _ "google.golang.org/grpc/encoding/gzip"`.
func WithCompressoion(compression otlptracehttp.Compression) otlptracehttp.Option {
	return otlptracehttp.WithCompression(compression)
}

// WithURLPath allows one to override the default URL path used
// for sending traces. If unset, default ("/v1/traces") will be used.
func WithURLPath(urlPath string) otlptracehttp.Option {
	return otlptracehttp.WithURLPath(urlPath)
}

// WithTLSClientConfig can be used to set up a custom TLS
// configuration for the client used to send payloads to the
// collector. Use it if you want to use a custom certificate.
func WithTLSClientConfig(tlsCfg *tls.Config) otlptracehttp.Option {
	return otlptracehttp.WithTLSClientConfig(tlsCfg)
}

// WithHeaders will send the provided headers with gRPC requests.
func WithHeaders(headers map[string]string) otlptracehttp.Option {
	return otlptracehttp.WithHeaders(headers)
}

// WithTimeout tells the driver the max waiting time for the backend to process
// each spans batch. If unset, the default will be 10 seconds.
func WithTimeout(duration time.Duration) otlptracehttp.Option {
	return otlptracehttp.WithTimeout(duration)
}

// WithRetry configures the retry policy for transient errors that may occurs
// when exporting traces. An exponential back-off algorithm is used to ensure
// endpoints are not overwhelmed with retries. If unset, the default retry
// policy will retry after 5 seconds and increase exponentially after each
// error for a total of 1 minute.
func WithRetry(rc otlptracehttp.RetryConfig) otlptracehttp.Option {
	return otlptracehttp.WithRetry(rc)
}
