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

package tspider

import (
	"context"
	"errors"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// ErrNotImplemented is an error returned when a method is not implemented.
var ErrNotImplemented = errors.New("Not implemented by tspider store")

// ModelInterface struct implement all interfaces of store interface
type ModelInterface struct {
}

// NewModelInterface creates a new instance of ModelInterface.
func NewModelInterface() *ModelInterface {
	return &ModelInterface{}
}

// GetProjectList is not implemented.
func (s *ModelInterface) GetProjectList(ctx context.Context, req *datamanager.GetAllProjectListRequest) ([]*datamanager.Project, int64, error) {
	return nil, 0, ErrNotImplemented
}

// GetProjectInfo is not implemented.
func (s *ModelInterface) GetProjectInfo(ctx context.Context, req *datamanager.GetProjectInfoRequest) (*datamanager.Project, error) {
	return nil, ErrNotImplemented
}

// GetClusterInfoList is not implemented.
func (s *ModelInterface) GetClusterInfoList(ctx context.Context, req *datamanager.GetClusterListRequest) ([]*datamanager.Cluster, int64, error) {
	return nil, 0, ErrNotImplemented
}

// GetClusterInfo is not implemented.
func (s *ModelInterface) GetClusterInfo(ctx context.Context, req *datamanager.GetClusterInfoRequest) (*datamanager.Cluster, error) {
	return nil, ErrNotImplemented
}

// GetNamespaceInfoList is not implemented.
func (s *ModelInterface) GetNamespaceInfoList(ctx context.Context, req *datamanager.GetNamespaceInfoListRequest) ([]*datamanager.Namespace, int64, error) {
	return nil, 0, ErrNotImplemented
}

// GetNamespaceInfo is not implemented.
func (s *ModelInterface) GetNamespaceInfo(ctx context.Context, req *datamanager.GetNamespaceInfoRequest) (*datamanager.Namespace, error) {
	return nil, ErrNotImplemented
}

// GetWorkloadInfoList is not implemented.
func (s *ModelInterface) GetWorkloadInfoList(ctx context.Context, req *datamanager.GetWorkloadInfoListRequest) ([]*datamanager.Workload, int64, error) {
	return nil, 0, ErrNotImplemented
}

// GetWorkloadInfo is not implemented.
func (s *ModelInterface) GetWorkloadInfo(ctx context.Context, req *datamanager.GetWorkloadInfoRequest) (*datamanager.Workload, error) {
	return nil, ErrNotImplemented
}

// GetRawPublicInfo is not implemented.
func (s *ModelInterface) GetRawPublicInfo(ctx context.Context, opts *types.JobCommonOpts) ([]*types.PublicData, error) {
	return nil, ErrNotImplemented
}

// GetRawProjectInfo is not implemented.
func (s *ModelInterface) GetRawProjectInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.ProjectData, error) {
	return nil, ErrNotImplemented
}

// GetRawClusterInfo is not implemented.
func (s *ModelInterface) GetRawClusterInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.ClusterData, error) {
	return nil, ErrNotImplemented
}

// GetRawNamespaceInfo is not implemented.
func (s *ModelInterface) GetRawNamespaceInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.NamespaceData, error) {
	return nil, ErrNotImplemented
}

// GetRawWorkloadInfo is not implemented.
func (s *ModelInterface) GetRawWorkloadInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.WorkloadData, error) {
	return nil, ErrNotImplemented
}

// InsertProjectInfo is not implemented.
func (s *ModelInterface) InsertProjectInfo(ctx context.Context, metrics *types.ProjectMetrics, opts *types.JobCommonOpts) error {
	return ErrNotImplemented
}

// InsertClusterInfo is not implemented.
func (s *ModelInterface) InsertClusterInfo(ctx context.Context, metrics *types.ClusterMetrics, opts *types.JobCommonOpts) error {
	return ErrNotImplemented
}

// InsertNamespaceInfo is not implemented.
func (s *ModelInterface) InsertNamespaceInfo(ctx context.Context, metrics *types.NamespaceMetrics, opts *types.JobCommonOpts) error {
	return ErrNotImplemented
}

// InsertWorkloadInfo is not implemented.
func (s *ModelInterface) InsertWorkloadInfo(ctx context.Context, metrics *types.WorkloadMetrics, opts *types.JobCommonOpts) error {
	return ErrNotImplemented
}

// GetWorkloadCount is not implemented.
func (s *ModelInterface) GetWorkloadCount(ctx context.Context, opts *types.JobCommonOpts, bucket string, after time.Time) (int64, error) {
	return 0, ErrNotImplemented
}

// InsertPublicInfo is not implemented.
func (s *ModelInterface) InsertPublicInfo(ctx context.Context, metrics *types.PublicData, opts *types.JobCommonOpts) error {
	return ErrNotImplemented
}

// InsertPodAutoscalerInfo is not implemented.
func (s *ModelInterface) InsertPodAutoscalerInfo(ctx context.Context, metrics *types.PodAutoscalerMetrics, opts *types.JobCommonOpts) error {
	return ErrNotImplemented
}

// GetPodAutoscalerList is not implemented.
func (s *ModelInterface) GetPodAutoscalerList(ctx context.Context, request *datamanager.GetPodAutoscalerListRequest) ([]*datamanager.PodAutoscaler, int64, error) {
	return nil, 0, ErrNotImplemented
}

// GetPodAutoscalerInfo is not implemented.
func (s *ModelInterface) GetPodAutoscalerInfo(ctx context.Context, request *datamanager.GetPodAutoscalerRequest) (*datamanager.PodAutoscaler, error) {
	return nil, ErrNotImplemented
}

// GetRawPodAutoscalerInfo is not implemented.
func (s *ModelInterface) GetRawPodAutoscalerInfo(ctx context.Context, opts *types.JobCommonOpts, bucket string) ([]*types.PodAutoscalerData, error) {
	return nil, ErrNotImplemented
}

// GetLatestWorkloadRequest is not implemented.
func (s *ModelInterface) GetLatestWorkloadRequest(ctx context.Context,
	req *datamanager.GetWorkloadRequestRecommendResultReq) (*datamanager.GetWorkloadRequestRecommendResultRsp, error) {
	return nil, ErrNotImplemented
}

// CreateWorkloadOriginRequest is not implemented.
func (s *ModelInterface) CreateWorkloadOriginRequest(ctx context.Context, result *types.WorkloadOriginRequestResult) error {
	return ErrNotImplemented
}

// ListWorkloadOriginRequest is not implemented.
func (s *ModelInterface) ListWorkloadOriginRequest(ctx context.Context,
	req *datamanager.GetWorkloadOriginRequestResultReq) ([]*datamanager.WorkloadOriginRequestResult, error) {
	return nil, ErrNotImplemented
}
