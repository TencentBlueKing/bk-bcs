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

// Package helm xxx
package helm

import (
	"time"
)

const (
	// PubicRepo public repo
	PubicRepo = "public-repo"
	// DefaultRetryCount
	DefaultRetryCount = 5
	// DefaultTimeout timeout
	DefaultTimeout = time.Second * 10

	// DefaultArgsFlagInsecure insecure flag
	DefaultArgsFlagInsecure = "--insecure-skip-tls-verify"
	// DefaultArgsFlagWait wait flag
	DefaultArgsFlagWait = "--wait"
)

var (
	// DefaultArgsFlag xxx
	DefaultArgsFlag = []string{DefaultArgsFlagInsecure, DefaultArgsFlagWait}
)

// HelmOptions xxx
type HelmOptions struct {
	ProjectID    string
	ClusterID    string
	Namespace    string
	ReleaseName  string
	ChartName    string
	ChartVersion string
	IsPublic     bool
}

// ReleaseBaseOptions options for federation release
type ReleaseBaseOptions struct {
	ProjectID       string
	ClusterID       string
	SkipWhenExisted bool
}

// BcsUnifiedApiserverOptions options for bcs-unified-apiserver
type BcsUnifiedApiserverOptions struct {
	UserToken      string
	LoadBalancerId string
	ReleaseBaseOptions
}

// BcsClusternetAgentOptions options for bcs-clusternet-agent
type BcsClusternetAgentOptions struct {
	SubClusterId      string
	UserToken         string
	RegistrationToken string
	BcsGateWayAddress string
	ReleaseBaseOptions
}

// BcsEstimatorAgentOptions options for bcs-estimator-agent
type BcsEstimatorAgentOptions struct {
	SubClusterId      string
	UserToken         string
	BcsGateWayAddress string
	ReleaseBaseOptions
}
