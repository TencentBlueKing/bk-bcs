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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Config represents the configuration options available for the http.Handler
// and http.Transport types.
type Config struct {
	TracerProvider trace.TracerProvider
}

// Option interface used for setting optional config properties.
type Option interface {
	apply(*Config)
}

// NewConfig creates a new config struct and applies opts to it.
func NewConfig(opts ...Option) *Config {
	c := &Config{
		TracerProvider: otel.GetTracerProvider(),
	}
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}
