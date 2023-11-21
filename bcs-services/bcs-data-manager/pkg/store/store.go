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

// Package store xxx
package store

import (
	"context"
	"time"

	any "github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// support store type
const (
	// TspiderStore tspider store
	TspiderStore = "tspider"
	// MongoStore mongoDB store
	MongoStore = "mongodb"
)

// Server store server interface
type Server interface {
	GetProjectList(ctx context.Context, req *datamanager.GetAllProjectListRequest) ([]*datamanager.Project, int64, error)
	GetProjectInfo(ctx context.Context, req *datamanager.GetProjectInfoRequest) (*datamanager.Project, error)
	GetClusterInfoList(ctx context.Context, req *datamanager.GetClusterListRequest) ([]*datamanager.Cluster,
		int64, error)
	GetClusterInfo(ctx context.Context, req *datamanager.GetClusterInfoRequest) (*datamanager.Cluster, error)
	GetNamespaceInfoList(ctx context.Context, req *datamanager.GetNamespaceInfoListRequest) ([]*datamanager.Namespace,
		int64, error)
	GetNamespaceInfo(ctx context.Context, req *datamanager.GetNamespaceInfoRequest) (*datamanager.Namespace, error)
	GetWorkloadInfoList(ctx context.Context, req *datamanager.GetWorkloadInfoListRequest) ([]*datamanager.Workload,
		int64, error)
	GetWorkloadInfo(ctx context.Context, req *datamanager.GetWorkloadInfoRequest) (*datamanager.Workload, error)
	GetRawPublicInfo(ctx context.Context, opts *types.JobCommonOpts) ([]*types.PublicData, error)
	GetRawProjectInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.ProjectData, error)
	GetRawClusterInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.ClusterData, error)
	GetRawNamespaceInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.NamespaceData, error)
	GetRawWorkloadInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.WorkloadData, error)
	InsertProjectInfo(ctx context.Context, metrics *types.ProjectMetrics, opts *types.JobCommonOpts) error
	InsertClusterInfo(ctx context.Context, metrics *types.ClusterMetrics, opts *types.JobCommonOpts) error
	InsertNamespaceInfo(ctx context.Context, metrics *types.NamespaceMetrics, opts *types.JobCommonOpts) error
	InsertWorkloadInfo(ctx context.Context, metrics *types.WorkloadMetrics, opts *types.JobCommonOpts) error
	GetWorkloadCount(ctx context.Context, opts *types.JobCommonOpts, bucket string,
		after time.Time) (int64, error)
	InsertPublicInfo(ctx context.Context, metrics *types.PublicData, opts *types.JobCommonOpts) error
	InsertPodAutoscalerInfo(ctx context.Context, metrics *types.PodAutoscalerMetrics,
		opts *types.JobCommonOpts) error
	GetPodAutoscalerList(ctx context.Context,
		request *datamanager.GetPodAutoscalerListRequest) ([]*datamanager.PodAutoscaler, int64, error)
	GetPodAutoscalerInfo(ctx context.Context,
		request *datamanager.GetPodAutoscalerRequest) (*datamanager.PodAutoscaler, error)
	GetRawPodAutoscalerInfo(ctx context.Context, opts *types.JobCommonOpts,
		bucket string) ([]*types.PodAutoscalerData, error)
	GetPowerTradingInfo(ctx context.Context,
		req *datamanager.GetPowerTradingDataRequest) ([]*any.Any, int64, error)
	GetCloudNativeWorkloadList(ctx context.Context,
		req *datamanager.GetCloudNativeWorkloadListRequest) (*datamanager.TEGMessage, error)
	GetUserOperationDataList(ctx context.Context,
		req *datamanager.GetUserOperationDataListRequest) ([]*structpb.Struct, int64, error)
	GetLatestWorkloadRequest(ctx context.Context,
		req *datamanager.GetWorkloadRequestRecommendResultReq) (*datamanager.GetWorkloadRequestRecommendResultRsp, error)

	CreateWorkloadOriginRequest(ctx context.Context, result *types.WorkloadOriginRequestResult) error
	ListWorkloadOriginRequest(ctx context.Context,
		req *datamanager.GetWorkloadOriginRequestResultReq) ([]*datamanager.WorkloadOriginRequestResult, error)
}
