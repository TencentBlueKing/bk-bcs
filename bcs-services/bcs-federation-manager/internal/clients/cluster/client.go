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

// Package cluster xxx
package cluster

import (
	"context"
	"crypto/tls"
	"fmt"

	v1beta1 "github.com/clusternet/clusternet/pkg/apis/clusters/v1beta1"
	"github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/registry"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/requester"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/common"
	federationv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/kubeapi/federationquota/api/v1"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
)

const (
	// ListK8SNamespacePath original kubernetes path with bcs prefix
	ListK8SNamespacePath = "/clusters/%s/api/v1/namespaces"

	// FedNamespaceIsFederatedKey federation namespace annotation
	FedNamespaceIsFederatedKey = "federation.bkbcs.tencent.com/is-federated-namespace"
	// FedNamespaceClusterRangeKey federation namespace cluster-range key
	FedNamespaceClusterRangeKey = "federation.bkbcs.tencent.com/cluster-range"
	// FedNamespaceProjectCodeKey federation namespace project code key
	FedNamespaceProjectCodeKey = "io.tencent.bcs.projectcode"
	// FedNamespaceProjectCodeTest federation namespace project code val
	FedNamespaceProjectCodeTest = "test"
	// FedNamespaceBkbcsProjectCodeKey federation namespace project code key
	FedNamespaceBkbcsProjectCodeKey = "federation.bkbcs.tencent.com/projectcode"
	// ClusterAffinityMode clusterAffinity mode
	ClusterAffinityMode = "federation.bkbcs.tencent.com/cluster-affinity-mode"
	// ClusterAffinitySelector clusterAffinity selector
	ClusterAffinitySelector = "federation.bkbcs.tencent.com/cluster-affinity-selector"
	// NamespaceUpdateTimestamp clusterAffinity selector
	NamespaceUpdateTimestamp = "federation.bkbcs.tencent.com/update-time"
	// ManagedClusterTypeLabel 集群的ManagedCluster对象的label
	ManagedClusterTypeLabel = "subscription.bkbcs.tencent.com/clustertype"
	// HostClusterNamespaceStatus 联邦集群命名空间状态-annotations
	HostClusterNamespaceStatus = "federation.bkbcs.tencent.com/host-cluster-status"
	// SubClusterNamespaceStatus 子集群命名空间状态-annotations
	SubClusterNamespaceStatus = "federation.bkbcs.tencent.com/sub-cluster-status"
	// NamespaceCreating ns status Creating
	NamespaceCreating = "Creating"
	// NamespaceSuccess ns status Creating
	NamespaceSuccess = "Success"
	// NamespaceActive ns status Active
	NamespaceActive = "Active"
	// NamespaceTerminating ns status Terminating
	NamespaceTerminating = "Terminating"
	// CreateNamespaceTaskId task id
	CreateNamespaceTaskId = "federation.bkbcs.tencent.com/create-namespace-taskId"
	// FedNamespaceBkBizId taiji bk_biz_id
	FedNamespaceBkBizId = "federation.bkbcs.tencent.com/bk-biz-id"
	// FedNamespaceBkModuleId  taiji bk_module_id
	FedNamespaceBkModuleId = "federation.bkbcs.tencent.com/bk-module-id"
	// LabelsMixerClusterKey is the label key for the mixer cluster.
	LabelsMixerClusterKey = "subscription.bkbcs.tencent.com/mixercluster"
	// LabelsMixerClusterPriorityKey is the label key for the mixer cluster.
	LabelsMixerClusterPriorityKey = "subscription.bkbcs.tencent.com/mixercluster-low-priority"
	// LabelsMixerClusterTkeNetworksKey is the label key for the mixer cluster.
	LabelsMixerClusterTkeNetworksKey = "subscription.bkbcs.tencent.com/mixercluster-tke-networks"

	// AnnotationMixerClusterPreemptionPolicyKey is the annotation key for the mixer cluster.
	AnnotationMixerClusterPreemptionPolicyKey = "mixer.kubernetes.io/preemption-policy"
	// AnnotationMixerClusterPreemptionClassKey is the annotation key for the mixer cluster.
	AnnotationMixerClusterPreemptionClassKey = "mixer.kubernetes.io/priority-class"
	// AnnotationMixerClusterPreemptionValueKey is the annotation key for the mixer cluster.
	AnnotationMixerClusterPreemptionValueKey = "mixer.kubernetes.io/priority-value"
	// AnnotationMixerClusterMixerNamespaceKey is the annotation key for the mixer cluster.
	AnnotationMixerClusterMixerNamespaceKey = "mixer.kubernetes.io/is-mixer-namespace"
	// AnnotationMixerClusterNetworksKey is the annotation key for the mixer cluster.
	AnnotationMixerClusterNetworksKey = "tke.cloud.tencent.com/networks"
	// AnnotationSubClusterForTaiji is the annotation key for the taiji cluster.
	AnnotationSubClusterForTaiji = "bkbcs.tencent.com/taiji-location"
	// AnnotationKeyInstalledPlatform is the annotation key cluster.
	AnnotationKeyInstalledPlatform = "installed-platform"
	// AnnotationIsPrivateResourceKey  is the annotation isPrivateResource key for the taiji.
	AnnotationIsPrivateResourceKey = "isPrivateResource"
	// AnnotationScheduleAlgorithmKey  is the annotation scheduleAlgorithm key for the taiji.
	AnnotationScheduleAlgorithmKey = "scheduleAlgorithm"
	// AnnotationScheduleAlgorithmValue  is the annotation scheduleAlgorithm default value for the MEMA.
	AnnotationScheduleAlgorithmValue = "MEMA"

	// MixerClusterNetworksValue networks Value
	MixerClusterNetworksValue = "flannel"
	// ValueIsTrue default value is true
	ValueIsTrue = "true"

	// MixerClusterPreemptionValue preemption value
	MixerClusterPreemptionValue = "-100"
	// MixerClusterPreemptionClassValue preemption class value
	MixerClusterPreemptionClassValue = "offline-pod-priority"
	// MixerClusterPreemptionPolicyValue preemption policy value
	MixerClusterPreemptionPolicyValue = "Never"

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

	// CreateKey handle type
	CreateKey = "create"
	// UpdateKey handle type
	UpdateKey = "update"
	// DeleteKey handle type
	DeleteKey = "delete"
	// TaskGpuTypeKey attributes gpu type key
	TaskGpuTypeKey = "task.bkbcs.tencent.com/gpu-type"
	// TaijiGPUNameKey gpu name key
	TaijiGPUNameKey = "GPUName"
	// ResultSuccessKey SUCCESS
	ResultSuccessKey = "SUCCESS"
)

var clusterCli Client

// SetClusterClient set cluster client
func SetClusterClient(opts *ClientOptions) {
	cli := NewClient(opts)
	clusterCli = cli
}

// GetClusterClient get cluster client
func GetClusterClient() Client {
	return clusterCli
}

// Client client interface for request k8s cluster
type Client interface {
	// GetCluster get cluster from cluster manager
	GetCluster(context.Context, string) (*clustermanager.Cluster, error)
	// ListProjectCluster list project cluster
	ListProjectCluster(context.Context, string) ([]*clustermanager.Cluster, error)
	// GetSubnetId get subnet id for cluster
	GetSubnetId(context.Context, string) (string, error)
	// CreateFederationCluster register federation cluster
	CreateFederationCluster(context.Context, *FederationClusterCreateReq) (string, error)
	// UpdateFederationClusterCredentials update federation cluster credentials
	UpdateFederationClusterCredentials(context.Context, string, string) error
	// UpdateFederationClusterStatus update federation cluster status
	UpdateFederationClusterStatus(context.Context, string, string) error
	// DeleteFederationCluster delete federation cluster
	DeleteFederationCluster(context.Context, string, string) error

	// GetClusterLabels get cluster labels
	GetClusterLabels(context.Context, string) (map[string]string, error)
	// UpdateClusterLabels update cluster labels
	UpdateClusterLabels(context.Context, string, map[string]string) error
	// AddClusterLabels add cluster label
	AddClusterLabels(context.Context, string, map[string]string) error
	// DeleteClusterLabels delete cluster label
	DeleteClusterLabels(context.Context, string, []string) error
	// UpdateHostCluster add host cluster label
	UpdateHostClusterLabel(context.Context, string) error
	// DeleteHostClusterLabel delete host cluster label
	DeleteHostClusterLabel(context.Context, string) error
	// UpdateSubCluster add sub cluster label
	UpdateSubClusterLabel(context.Context, string) error
	// DeleteSubClusterLabel delete sub cluster label
	DeleteSubClusterLabel(context.Context, string) error

	// ListFederationNamespaces list federation namespaces
	ListFederationNamespaces(string) ([]*FederationNamespace, error)

	// CreateNamespace create namespace in target cluster
	CreateNamespace(clusterID, namespace string) error
	// GetLoadbalancerIp get external ip of service
	GetLoadbalancerIp(opt *ResourceGetOptions) (string, error)
	// CreateSecret create secret in target cluster
	CreateSecret(secret *corev1.Secret, opt *ResourceCreateOptions) error
	// ListSecrets list secrets in target cluster
	ListSecrets(opt *ResourceGetOptions) ([]corev1.Secret, error)
	// GetBootstrapSecret get bootstrap secret in target cluster
	GetBootstrapSecret(opt *ResourceGetOptions) (*corev1.Secret, error)

	// GetManagedCluster get managed cluster
	GetManagedCluster(hostClusterId, subClusterID string) (*v1beta1.ManagedCluster, error)
	// DeleteManagedCluster delete managed cluster
	DeleteManagedCluster(hostClusterId, subClusterID string) error
	// GetClusterRegistrationRequest create cluster registration request
	GetClusterRegistrationRequest(hostClusterId, subClusterID string) (*v1beta1.ClusterRegistrationRequest, error)
	// DeleteClusterRegistrationRequest delete cluster registration request
	DeleteClusterRegistrationRequest(hostClusterId, subClusterID string) error

	// GetNamespace get ns by clusterId nsName
	GetNamespace(clusterID, nsName string) (*corev1.Namespace, error)
	// ListNamespace get nss by clusterId
	ListNamespace(clusterID string) ([]corev1.Namespace, error)
	// DeleteNamespace delete namespace by clusterId nsName
	DeleteNamespace(clusterID, nsName string) error
	// UpdateNamespace update namespace
	UpdateNamespace(clusterId string, ns *corev1.Namespace) error
	// CreateClusterNamespace create namespace
	CreateClusterNamespace(clusterId, namespace string, annotations map[string]string) error

	// GetNamespaceQuota get quota by clusterId nsName quotaName
	GetNamespaceQuota(clusterID, nsName, quotaName string) (*federationv1.MultiClusterResourceQuota, error)
	// ListNamespaceQuota get quotas by clusterId nsName
	ListNamespaceQuota(clusterID, nsName string) (*federationv1.MultiClusterResourceQuotaList, error)
	// DeleteNamespaceQuota delete quota by clusterId nsName quotaName
	DeleteNamespaceQuota(clusterID, nsName, quotaName string) error
	// UpdateNamespaceQuota update quota
	UpdateNamespaceQuota(clusterId, namespace string, mcResourceQuota *federationv1.MultiClusterResourceQuota) error
	// CreateNamespaceQuota create namespace quota
	CreateNamespaceQuota(federationId string, req *federationmgr.CreateFederationClusterNamespaceQuotaRequest) error
	// CreateMultiClusterResourceQuota create multiClusterResourceQuota
	CreateMultiClusterResourceQuota(clusterId, namespace string, mc *federationv1.MultiClusterResourceQuota) error
	// GetMultiClusterResourceQuota get multiClusterResourceQuota
	GetMultiClusterResourceQuota(clusterId, namespace, quotaName string) (*federationv1.MultiClusterResourceQuota, error)
}

// ClientOptions options for create client
type ClientOptions struct {
	ClientTLS     *tls.Config
	EtcdEndpoints []string
	EtcdTLS       *tls.Config
	requester.BaseOptions
}

// NewClient create client with options
func NewClient(opts *ClientOptions) Client {
	header := make(map[string]string)
	header[common.HeaderAuthorizationKey] = fmt.Sprintf("Bearer %s", opts.Token)
	header[common.BcsHeaderClientKey] = common.InnerModuleName
	if opts.Sender == nil {
		opts.Sender = requester.NewRequester()
	}

	c := grpc.NewClient(
		client.Registry(etcd.NewRegistry(
			registry.Addrs(opts.EtcdEndpoints...),
			registry.TLSConfig(opts.ClientTLS)),
		),
		grpc.AuthTLS(opts.ClientTLS),
	)
	cli := clustermanager.NewClusterManagerService(common.ModuleClusterManager, c)

	return &clusterClient{
		opt:           opts,
		defaultHeader: header,
		clusterSvc:    cli,
	}
}

type clusterClient struct {
	opt           *ClientOptions
	defaultHeader map[string]string
	clusterSvc    clustermanager.ClusterManagerService
}
