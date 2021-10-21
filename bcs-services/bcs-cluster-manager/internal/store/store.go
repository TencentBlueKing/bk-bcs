/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package store

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/clustercredential"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/namespacequota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/tke"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// ClusterManagerModel database operation for
type ClusterManagerModel interface {
	CreateCluster(ctx context.Context, cluster *types.Cluster) error
	UpdateCluster(ctx context.Context, cluster *types.Cluster) error
	DeleteCluster(ctx context.Context, clusterID string) error
	GetCluster(ctx context.Context, clusterID string) (*types.Cluster, error)
	ListCluster(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.Cluster, error)

	CreateNamespace(ctx context.Context, ns *types.Namespace) error
	UpdateNamespace(ctx context.Context, ns *types.Namespace) error
	DeleteNamespace(ctx context.Context, name, federationClusterID string) error
	GetNamespace(ctx context.Context, name, federationClusterID string) (*types.Namespace, error)
	ListNamespace(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.Namespace, error)

	CreateQuota(ctx context.Context, quota *types.NamespaceQuota) error
	UpdateQuota(ctx context.Context, quota *types.NamespaceQuota) error
	DeleteQuota(ctx context.Context, namespace, federationClusterID, clusterID string) error
	GetQuota(ctx context.Context, namespace, federationClusterID, clusterID string) (*types.NamespaceQuota, error)
	ListQuota(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.NamespaceQuota, error)
	BatchDeleteQuotaByCluster(ctx context.Context, clusterID string) error

	PutClusterCredential(ctx context.Context, clusterCredential *types.ClusterCredential) error
	GetClusterCredential(ctx context.Context, serverKey string) (*types.ClusterCredential, bool, error)
	DeleteClusterCredential(ctx context.Context, serverKey string) error
	ListClusterCredential(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
		[]types.ClusterCredential, error)

	CreateTkeCidr(ctx context.Context, cidr *types.TkeCidr) error
	UpdateTkeCidr(ctx context.Context, cidr *types.TkeCidr) error
	DeleteTkeCidr(ctx context.Context, vpc string, cidr string) error
	GetTkeCidr(ctx context.Context, vpc string, cidr string) (*types.TkeCidr, error)
	ListTkeCidr(ctx context.Context, cond *operator.Condition, opt *options.ListOption) ([]types.TkeCidr, error)
	ListTkeCidrCount(ctx context.Context, opt *options.ListOption) ([]types.TkeCidrCount, error)
}

// ModelSet a set of client
type ModelSet struct {
	*cluster.ModelCluster
	*clustercredential.ModelClusterCredential
	*namespace.ModelNamespace
	*namespacequota.ModelNamespaceQuota
	*tke.ModelTkeCidr
}

// NewModelSet create model set
func NewModelSet(db drivers.DB) *ModelSet {
	return &ModelSet{
		ModelCluster:           cluster.New(db),
		ModelClusterCredential: clustercredential.New(db),
		ModelNamespace:         namespace.New(db),
		ModelNamespaceQuota:    namespacequota.New(db),
		ModelTkeCidr:           tke.New(db),
	}
}
