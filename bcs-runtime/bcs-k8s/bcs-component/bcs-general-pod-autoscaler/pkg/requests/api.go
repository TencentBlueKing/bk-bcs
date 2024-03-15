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

// Package requests xxx
package requests

import "k8s.io/apimachinery/pkg/types"

// AutoscaleRequest defines the request to webhook autoscaler endpoint
type AutoscaleRequest struct {
	// UID is used for tracing the request and response.
	UID types.UID `json:"uid"`
	// Name is the name of the workload(Squad, Statefulset...) being scaled
	Name string `json:"name"`
	// Namespace is the workload namespace
	Namespace string `json:"namespace"`
	// Parameters are the parameter that required by webhook
	Parameters map[string]string `json:"parameters"`
	// CurrentReplicas is the current replicas
	CurrentReplicas int32 `json:"currentReplicas"`
}

// AutoscaleResponse defines the response of webhook server
type AutoscaleResponse struct {
	// UID is used for tracing the request and response.
	// It should be same as it in the request.
	UID types.UID `json:"uid"`
	// Set to false if should not do scaling
	Scale bool `json:"scale"`
	// Replicas is targeted replica count from the webhookServer
	Replicas int32 `json:"replicas"`
}

// AutoscaleReview is passed to the webhook with a populated Request value,
// and then returned with a populated Response.
type AutoscaleReview struct {
	Request  *AutoscaleRequest  `json:"request"`
	Response *AutoscaleResponse `json:"response"`
}
