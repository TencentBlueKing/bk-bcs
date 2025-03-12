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

// Package types xxx
package types

// FederationCharts federation charts
type FederationCharts struct {
	ClusternetHub   *Chart `json:"clusternethub"`
	Scheduler       *Chart `json:"scheduler"`
	Controller      *Chart `json:"controller"`
	Apiserver       *Chart `json:"apiserver"`
	ClusternetAgent *Chart `json:"clusternetagent"`
	EstimatorAgent  *Chart `json:"estimatoragent"`
}

// Chart helm chart
type Chart struct {
	ChartVersion     string `json:"chartVersion"`
	ChartName        string `json:"chartName"`
	ReleaseName      string `json:"releaseName"`
	ReleaseNamespace string `json:"releaseNamespace"`
	DefaultValues    string `json:"values"`
}
