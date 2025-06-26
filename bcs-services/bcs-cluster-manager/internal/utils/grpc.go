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

package utils

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// Authentication defines the common interface for the credentials which need to
// attach auth info to every RPC
type Authentication struct {
	InnerClientName string
	Insecure        bool
}

// GetRequestMetadata gets the current request metadata
func (a *Authentication) GetRequestMetadata(context.Context, ...string) (
	map[string]string, error,
) {
	return map[string]string{middleware.InnerClientHeaderKey: a.InnerClientName}, nil
}

// RequireTransportSecurity indicates whether the credentials requires
// transport security.
func (a *Authentication) RequireTransportSecurity() bool {
	return !a.Insecure
}

// NewTokenAuth implementations of grpc credentials interface
func NewTokenAuth(t string) *GrpcTokenAuth {
	return &GrpcTokenAuth{
		Token: t,
	}
}

// GrpcTokenAuth grpc token
type GrpcTokenAuth struct {
	Token string
}

// GetRequestMetadata convert http Authorization for grpc key
func (t GrpcTokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", t.Token),
	}, nil
}

// RequireTransportSecurity RequireTransportSecurity
func (t GrpcTokenAuth) RequireTransportSecurity() bool {
	return false
}
