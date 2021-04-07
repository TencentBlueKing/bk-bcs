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

package metrics

import (
	"context"
	"time"
)

// HTTPReqProperties for the metrics based on client request.
type HTTPReqProperties struct {
	// ID is the id of the request handler or handlerName.
	Handler string
	// Method is the method of the request.
	Method string
	// Code is the response of the request.
	Code string
	// ClusterID is the parameter of the request
	ClusterID string
	// ResourceType is the parameter of the request
	ResourceType string
}

// HTTPProperties for the global server metrics.
type HTTPProperties struct {
	// ID is the id of the request handler or handlerName.
	Handler string
	// Method is the method of the request.
	Method string
}

// Recorder record and measure the metrics.
// This Interface has the required methods to be used with the HTTP middleware.
type Recorder interface {
	// ObserveHTTPRequestCounterDuration measures the count and duration of an HTTP request.
	ObserveHTTPRequestCounterDuration(ctx context.Context, props HTTPReqProperties, duration time.Duration)
	// ObserveHTTPResponseSize measures the size of an HTTP response in bytes.
	ObserveHTTPResponseSize(ctx context.Context, props HTTPReqProperties, sizeBytes int64)
	// AddInflightRequests increments and decrements the number of inflight request being processed.
	AddInflightRequests(ctx context.Context, props HTTPProperties, quantity int)
}
