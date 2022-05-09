/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package store

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// Server store server interface
type Server interface {
	GetProjectInfo(ctx context.Context, req *datamanager.GetProjectInfoRequest) (*datamanager.Project, error)
	GetClusterInfoList(ctx context.Context, req *datamanager.GetClusterInfoListRequest) ([]*datamanager.Cluster,
		int64, error)
	GetClusterInfo(ctx context.Context, req *datamanager.GetClusterInfoRequest) (*datamanager.Cluster, error)
	GetNamespaceInfoList(ctx context.Context, req *datamanager.GetNamespaceInfoListRequest) ([]*datamanager.Namespace,
		int64, error)
	GetNamespaceInfo(ctx context.Context, req *datamanager.GetNamespaceInfoRequest) (*datamanager.Namespace, error)
	GetWorkloadInfoList(ctx context.Context, req *datamanager.GetWorkloadInfoListRequest) ([]*datamanager.Workload,
		int64, error)
	GetWorkloadInfo(ctx context.Context, req *datamanager.GetWorkloadInfoRequest) (*datamanager.Workload, error)
	GetRawPublicInfo(ctx context.Context, opts *common.JobCommonOpts) ([]*common.PublicData, error)
	GetRawProjectInfo(ctx context.Context, opts *common.JobCommonOpts, bucket string) ([]*common.ProjectData, error)
	GetRawClusterInfo(ctx context.Context, opts *common.JobCommonOpts, bucket string) ([]*common.ClusterData, error)
	GetRawNamespaceInfo(ctx context.Context, opts *common.JobCommonOpts, bucket string) ([]*common.NamespaceData, error)
	GetRawWorkloadInfo(ctx context.Context, opts *common.JobCommonOpts, bucket string) ([]*common.WorkloadData, error)
	InsertProjectInfo(ctx context.Context, metrics *common.ProjectMetrics, opts *common.JobCommonOpts) error
	InsertClusterInfo(ctx context.Context, metrics *common.ClusterMetrics, opts *common.JobCommonOpts) error
	InsertNamespaceInfo(ctx context.Context, metrics *common.NamespaceMetrics, opts *common.JobCommonOpts) error
	InsertWorkloadInfo(ctx context.Context, metrics *common.WorkloadMetrics, opts *common.JobCommonOpts) error
	GetWorkloadCount(ctx context.Context, opts *common.JobCommonOpts, bucket string,
		after time.Time) (int64, error)
	InsertPublicInfo(ctx context.Context, metrics *common.PublicData, opts *common.JobCommonOpts) error
}

type server struct {
	*ModelCluster
	*ModelNamespace
	*ModelProject
	*ModelWorkload
	*ModelPublic
}

// NewServer new db server
func NewServer(db drivers.DB) Server {
	return &server{
		ModelCluster:   NewModelCluster(db),
		ModelNamespace: NewModelNamespace(db),
		ModelWorkload:  NewModelWorkload(db),
		ModelProject:   NewModelProject(db),
		ModelPublic:    NewModelPublic(db),
	}
}
