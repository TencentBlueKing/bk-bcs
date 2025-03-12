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

// Package steps include all steps for federation manager
package steps

// StepNames step names
type StepNames struct {
	Name  string
	Alias string
}

// CallBackNames call back names
type CallBackNames struct {
	Name  string
	Alias string
}

const (
	// ProjectIdKey is used for task transfer params
	ProjectIdKey = "ProjectId"
	// ProjectCodeKey is used for task transfer params
	ProjectCodeKey = "ProjectCode"
	// HostProjectIdKey is projectid for host cluster, host cluster's project may be different with federation cluster
	HostProjectIdKey = "HostProjectId"
	// HostProjectCodeKey is project code for host cluster, host cluster's project may be different with federation cluster
	HostProjectCodeKey = "HostProjectCode"
	// SubProjectIdKey is used for task transfer params
	SubProjectIdKey = "SubProjectId"
	// SubProjectCodeKey is used for task transfer params
	SubProjectCodeKey = "SubProjectCode"
	// ClusterIdKey is used for task transfer params
	ClusterIdKey = "ClusterId"
	// UserTokenKey is used for task transfer params which used to install unified-apiserver
	UserTokenKey = "UserToken"
	// LoadBalancerIdKey is used for task transfer params
	LoadBalancerIdKey = "LbId"
	// BcsUnifiedApiserverAddressKey is used for task transfer params
	BcsUnifiedApiserverAddressKey = "BcsUnifiedApiserverAddress"
	// CreatorKey is used for task transfer params
	CreatorKey = "Creator"
	// UpdaterKey is used for task transfer params
	UpdaterKey = "Updater"
	// FedClusterIdKey is proxy cluster id for proxy cluster
	FedClusterIdKey = "FedClusterId"
	// HostClusterIdKey is clusterId for host cluster
	HostClusterIdKey = "HostClusterId"
	// SubClusterIdKey is used for task transfer params
	SubClusterIdKey = "SubClusterId"

	// FederationBusinessIdKey is the business id that federation cluster belong to
	FederationBusinessIdKey = "FederationBusinessId"
	// FederationProjectIdKey is the project id that federation cluster belong to
	FederationProjectIdKey = "FederationProjectId"
	// FederationProjectCodeKey is the project code that federation cluster belong to
	FederationProjectCodeKey = "FederationProjectCode"
	// FederationClusterNameKey is used for task transfer params
	FederationClusterNameKey = "FederationClusterName"
	// FederationClusterEnvKey is used for task transfer params
	FederationClusterEnvKey = "FederationEnv"
	// FederationClusterDescriptionKey is used for task transfer params
	FederationClusterDescriptionKey = "FederationClusterDescription"
	// FederationClusterLabelsStrKey is used for task transfer params
	FederationClusterLabelsStrKey = "FederationClusterLabels"

	// Success step result for success steps
	Success = "success"
	// Failed  step result for failed steps
	Failed = "failed"

	// BootstrapTokenNamespace is the namespace of bootstrap token
	BootstrapTokenNamespace = "kube-system"
	// BootstrapTokenIdKey is used for task transfer params
	BootstrapTokenIdKey = "token-id"
	// BootstrapTokenSecretKey is used for task transfer params
	BootstrapTokenSecretKey = "token-secret"

	// BcsGatewayAddressKey is bcs gateway address used for install estimator and clusternet agent
	BcsGatewayAddressKey = "BcsGatewayAddress"
	// BcsThirdpartyServiceDomain domain name for service
	BcsThirdpartyServiceDomain = "bcsthirdpartyservice.bkbcs.tencent.com"
	// CreateKey handle type
	CreateKey = "create"
	// UpdateKey handle type
	UpdateKey = "update"
	// DeleteKey handle type
	DeleteKey = "delete"
	// SubClusterForTaiji taiji cluster
	SubClusterForTaiji = "taiji"
	// SubClusterForSuanli suanli cluster
	SubClusterForSuanli = "suanli"
	// SubClusterForHunbu hunbu cluster
	SubClusterForHunbu = "hunbu"
	// SubClusterForNormal normal cluster
	SubClusterForNormal = "normal"
	// ClusterQuotaKey cluster quota
	ClusterQuotaKey = "quota"

	// DefaultAttemptTimes 尝试次数
	DefaultAttemptTimes = 5
	// DefaultRetryDelay 重试延迟
	DefaultRetryDelay = 1
	// DefaultMaxDelay 最大延迟
	DefaultMaxDelay = 10

	// NamespaceKey namespace
	NamespaceKey = "namespace"
	// ParameterKey parameter
	ParameterKey = "parameter"
	// HandleTypeKey handle type
	HandleTypeKey = "handleType"
)
