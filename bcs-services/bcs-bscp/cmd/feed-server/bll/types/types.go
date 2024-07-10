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

// Package types NOTES
package types

import (
	"net/http"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/gse"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbcommit "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/commit"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbcontent "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	pbhook "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/hook"
	pbkv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/kv"
)

var (
	// AsyncDownloadJobStatusPending in pending status, job can collect the target(task) clients.
	AsyncDownloadJobStatusPending = "Pending"
	// AsyncDownloadJobStatusRunning running status to precess downloading task and gse task.
	AsyncDownloadJobStatusRunning = "Running"
	// AsyncDownloadJobStatusSuccess means all the targets(tasks) in job were successd.
	AsyncDownloadJobStatusSuccess = "Success"
	// AsyncDownloadJobStatusFailed means there are failed targets(tasks) in job.
	AsyncDownloadJobStatusFailed = "Failed"
	// AsyncDownloadJobStatusTimeout means there are targets(tasks) still in process until job time out.
	AsyncDownloadJobStatusTimeout = "Timeout"
)

// AppInstanceMeta defines an app instance's metadata information.
type AppInstanceMeta struct {
	BizID     uint32            `json:"bizID"`
	AppID     uint32            `json:"appID"`
	App       string            `json:"app"`
	Namespace string            `json:"namespace"`
	Uid       string            `json:"uid"`
	Labels    map[string]string `json:"labels"`
}

// ListFileAppLatestReleaseMetaReq defines options to list a file type app's latest release metadata.
type ListFileAppLatestReleaseMetaReq struct {
	BizId     uint32            `json:"biz_id,omitempty"`
	AppId     uint32            `json:"app_id,omitempty"`
	Uid       string            `json:"uid,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// Bind go-chi/render Binder 接口实现
func (op *ListFileAppLatestReleaseMetaReq) Bind(r *http.Request) error {
	return op.Validate()
}

// Validate options is valid or not.
func (op *ListFileAppLatestReleaseMetaReq) Validate() error {
	if op.BizId <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should be > 0")
	}

	if op.AppId <= 0 {
		return errf.New(errf.InvalidParameter, "invalid app id, should be > 0")
	}

	if err := validator.ValidateUidLength(op.Uid); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := validator.ValidateLabel(op.Labels); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	return nil
}

// ReleasedCIMeta defines a release's released config item metadata
type ReleasedCIMeta struct {
	RciId                uint32                     `json:"rci_id,omitempty"`
	CommitID             uint32                     `json:"commit_id,omitempty"`
	CommitSpec           *pbcommit.CommitSpec       `json:"commit_spec,omitempty"`
	ConfigItemSpec       *pbci.ConfigItemSpec       `json:"config_item_spec,omitempty"`
	ConfigItemAttachment *pbci.ConfigItemAttachment `json:"config_item_attachment,omitempty"`
	ConfigItemRevision   *pbbase.Revision           `json:"config_item_revision,omitempty"`
	RepositorySpec       *RepositorySpec            `json:"repository_spec,omitempty"`
}

// RepositorySpec repository spec.
type RepositorySpec struct {
	// Path to pull the config file's sub uri.
	Path string `json:"path,omitempty"`
}

// AppLatestReleaseMeta an app's latest release metadata.
type AppLatestReleaseMeta struct {
	// ReleaseId the app's latest release's id.
	ReleaseId   uint32            `json:"release_id,omitempty"`
	ReleaseName string            `json:"release_name,omitempty"`
	Repository  *Repository       `json:"repository,omitempty"`
	ConfigItems []*ReleasedCIMeta `json:"config_items,omitempty"`
	PreHook     *pbhook.HookSpec  `json:"pre_hook,omitempty"`
	PostHook    *pbhook.HookSpec  `json:"post_hook,omitempty"`
}

// Repository data.
type Repository struct {
	Root string `json:"root,omitempty"`
}

// ListFileAppLatestReleaseMetaResp list a file type app's latest release metadata response.
type ListFileAppLatestReleaseMetaResp struct {
	Code    int32                 `json:"code,omitempty"`
	Message string                `json:"message,omitempty"`
	Data    *AppLatestReleaseMeta `json:"data,omitempty"`
}

// AppLatestReleaseKvMeta an app's latest release metadata.
type AppLatestReleaseKvMeta struct {
	// ReleaseId the app's latest release's id.
	ReleaseId uint32            `json:"release_id,omitempty"`
	Kvs       []*ReleasedKvMeta `json:"kvs,omitempty"`
	PreHook   *pbhook.HookSpec  `json:"pre_hook,omitempty"`
	PostHook  *pbhook.HookSpec  `json:"post_hook,omitempty"`
}

// ReleasedKvMeta defines a release's released kv metadata
type ReleasedKvMeta struct {
	Key          string                 `json:"key,omitempty"`
	KvType       string                 `json:"kv_type,omitempty"`
	Revision     *pbbase.Revision       `json:"revision,omitempty"`
	KvAttachment *pbkv.KvAttachment     `json:"kv_attachment,omitempty"`
	ContentSpec  *pbcontent.ContentSpec `json:"content_spec,omitempty"`
}

// AsyncDownloadJob defines async download job.
type AsyncDownloadJob struct {
	BizID              uint32                                             `json:"biz_id"`
	AppID              uint32                                             `json:"app_id"`
	JobID              string                                             `json:"job_id"`
	FilePath           string                                             `json:"file_path"`
	FileName           string                                             `json:"file_name"`
	FileSignature      string                                             `json:"file_signature"`
	TargetFileDir      string                                             `json:"target_file_dir"`
	TargetUser         string                                             `json:"target_user"`
	Targets            []*AsyncDownloadTarget                             `json:"targets"`
	GSETaskID          string                                             `json:"gse_task_id"`
	ExecuteTime        time.Time                                          `json:"execute_time"`
	Status             string                                             `json:"status"`
	CreateTime         time.Time                                          `json:"create_time"`
	SuccessTargets     map[string]gse.TransferFileResultDataResultContent `json:"success_targets"`
	FailedTargets      map[string]gse.TransferFileResultDataResultContent `json:"failed_targets"`
	DownloadingTargets map[string]gse.TransferFileResultDataResultContent `json:"downloading_targets"`
	TimeoutTargets     map[string]gse.TransferFileResultDataResultContent `json:"timeout_targets"`
}

// AsyncDownloadTarget defines async download target.
type AsyncDownloadTarget struct {
	AgentID     string `json:"agent_id"`
	ContainerID string `json:"container_id"`
}

// AsyncDownloadTask defines async download task.
type AsyncDownloadTask struct {
	BizID             uint32    `json:"biz_id"`
	AppID             uint32    `json:"app_id"`
	JobID             string    `json:"job_id"`
	TargetAgentID     string    `json:"target_agent_id"`
	TargetContainerID string    `json:"target_container_id"`
	FilePath          string    `json:"file_path"`
	FileName          string    `json:"file_name"`
	FileSignature     string    `json:"file_signature"`
	Status            string    `json:"status"`
	CreateTime        time.Time `json:"create_time"`
}
