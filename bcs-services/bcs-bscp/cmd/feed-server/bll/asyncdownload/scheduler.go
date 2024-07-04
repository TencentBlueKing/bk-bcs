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

// Package asyncdownload NOTES
package asyncdownload

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	prm "github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/bcs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/gse"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/bedis"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/lock"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

var (
	// JobTimeoutSeconds is the timeout seconds for async download job
	JobTimeoutSeconds = 10 * 60
)

// Scheduler scheduled task to process download jobs
type Scheduler struct {
	ctx               context.Context
	cancel            context.CancelFunc
	bds               bedis.Client
	redLock           *lock.RedisLock
	fileLock          *lock.FileLock
	provider          repository.Provider
	serverAgentID     string
	serverContainerID string
	metric            *metric
}

// NewScheduler create a async download scheduler
func NewScheduler(mc *metric, redLock *lock.RedisLock) (*Scheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())
	bds, err := bedis.NewRedisCache(cc.FeedServer().RedisCluster)
	if err != nil {
		cancel()
		return nil, err
	}
	// set ttl to 60 seconds, cause the job include downloading which may cost a lot of time
	fileLock := lock.NewFileLock()
	provider, err := repository.NewProvider(cc.FeedServer().Repository)
	if err != nil {
		cancel()
		return nil, err
	}

	// bcs-watch report pod/container data may delay, so retry to get server agent id and container id
	retry := tools.NewRetryPolicy(5, [2]uint{3000, 5000})

	var serverAgentID, serverContainerID string
	var lastErr error
	for {
		select {
		case <-ctx.Done():
			cancel()
			return nil, fmt.Errorf("get server agent id and container id failed, err, %s", ctx.Err().Error())
		default:
		}

		if retry.RetryCount() == 5 {
			cancel()
			return nil, lastErr
		}

		serverAgentID, serverContainerID, lastErr = getAsyncDownloadServerInfo(ctx, cc.FeedServer().GSE)
		if lastErr != nil {
			retry.Sleep()
			continue
		}
		break
	}

	logs.Infof("server agent id: %s, server container id: %s", serverAgentID, serverContainerID)
	return &Scheduler{
		ctx:               ctx,
		cancel:            cancel,
		bds:               bds,
		redLock:           redLock,
		fileLock:          fileLock,
		provider:          provider,
		serverAgentID:     serverAgentID,
		serverContainerID: serverContainerID,
		metric:            mc,
	}, nil
}

// Run run a scheduled task
func (a *Scheduler) Run() {

	go func() {
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ticker.C:
				a.do()
			case <-a.ctx.Done():
				logs.Infof("async downloader stopped")
				return
			}
		}
	}()
}

// Stop stop scheduled task
func (a *Scheduler) Stop() {
	a.cancel()
}

func (a *Scheduler) do() {

	keys, err := a.bds.Keys(a.ctx, "AsyncDownloadJob:*")
	if err != nil {
		logs.Errorf("list async download job keys from redis failed, err: %s", err.Error())
		return
	}

	for _, key := range keys {
		if err := func() error {
			// lock by job to prevent from
			// 1. concurrency writing in api AsyncDownload
			// 2. concurrency writing in other feedserver instance cronjob
			if a.redLock.TryAcquire(key) {
				defer a.redLock.Release(key)
				data, err := a.bds.Get(a.ctx, key)
				if err != nil {
					return err
				}
				if data == "" {
					return nil
				}
				job := &types.AsyncDownloadJob{}
				if err := jsoni.Unmarshal([]byte(data), job); err != nil {
					return err
				}
				switch job.Status {
				case types.AsyncDownloadJobStatusPending:
					if time.Since(job.CreateTime) < 15*time.Second {
						// continue to collect target clients

						// TODO: optimize the logic to collect target clients
						// 1. create a new job set execute time as 5 seconds later
						// 2. if any target collected during pending, set the execute time as 5 seconds later from now
						// 3. if execute time expired, stop collect, start to download
						return nil
					}
					return a.handleDownload(job)
				case types.AsyncDownloadJobStatusRunning:
					return a.checkJobStatus(job)
				case types.AsyncDownloadJobStatusSuccess,
					types.AsyncDownloadJobStatusFailed,
					types.AsyncDownloadJobStatusTimeout:
					return nil
				default:
					logs.Errorf("invalid async download job status: %s", job.Status)
				}
			}
			return nil
		}(); err != nil {
			logs.Errorf("handle async download job %s failed, err: %s", key, err.Error())
		}
	}
}

func (a *Scheduler) handleDownload(job *types.AsyncDownloadJob) error {

	kt := kit.New()
	kt.BizID = job.BizID
	kt.AppID = job.AppID

	// 1. 更新任务状态
	job.Status = types.AsyncDownloadJobStatusRunning
	job.ExecuteTime = time.Now()
	if err := a.updateAsyncDownloadJobStatus(a.ctx, job); err != nil {
		return err
	}

	// 2. 下载文件到本地
	sourceDir := path.Join(cc.FeedServer().GSE.SourceDir, strconv.Itoa(int(job.BizID)))
	if err := os.MkdirAll(sourceDir, os.ModePerm); err != nil {
		return err
	}
	// filepath = source/{biz_id}/{sha256}
	signature := job.FileSignature
	serverFilePath := path.Join(sourceDir, signature)
	if err := a.checkAndDownloadFile(kt, serverFilePath, signature); err != nil {
		return err
	}

	// 3. 创建GSE文件传输任务
	targetAgents := make([]gse.TransferFileAgent, 0, len(job.Targets))
	for _, target := range job.Targets {
		targetAgents = append(targetAgents, gse.TransferFileAgent{
			BkAgentID:     target.AgentID,
			BkContainerID: target.ContainerID,
			User:          job.TargetUser,
		})
	}

	taskID, err := gse.CreateTransferFileTask(a.ctx, a.serverAgentID, a.serverContainerID, sourceDir,
		cc.FeedServer().GSE.AgentUser, signature, job.TargetFileDir, targetAgents)
	if err != nil {
		return fmt.Errorf("create gse transfer file task failed, %s", err.Error())
	}

	// 4. 更新任务状态
	job.GSETaskID = taskID

	if err := a.updateAsyncDownloadJobStatus(a.ctx, job); err != nil {
		return err
	}

	return nil
}

// updateAsyncDownloadJobStatus update async download job status to redis
// ! make sure job must be locked by upper caller to avoid concurrency update
func (a *Scheduler) updateAsyncDownloadJobStatus(ctx context.Context, job *types.AsyncDownloadJob) error {
	data, err := a.bds.Get(ctx, job.JobID)
	if err != nil {
		return err
	}
	if data == "" {
		logs.Errorf("update asyncdownload job %s status failed, not found in redis", job.JobID)
		return nil
	}

	old := new(types.AsyncDownloadJob)
	if err = jsoni.UnmarshalFromString(data, old); err != nil {
		return err
	}

	if old.Status != job.Status {
		var duration float64
		if old.Status == types.AsyncDownloadJobStatusPending {
			duration = time.Since(job.CreateTime).Seconds()
		} else {
			duration = time.Since(job.ExecuteTime).Seconds()
		}

		a.metric.jobDurationSeconds.With(prm.Labels{"biz": strconv.Itoa(int(job.BizID)),
			"app": strconv.Itoa(int(job.AppID)), "file": path.Join(job.FilePath, job.FileName),
			"targets": strconv.Itoa(len(job.Targets)), "status": old.Status}).
			Observe(duration)
		a.metric.jobCounter.With(prm.Labels{"biz": strconv.Itoa(int(job.BizID)),
			"app": strconv.Itoa(int(job.AppID)), "file": path.Join(job.FilePath, job.FileName),
			"targets": strconv.Itoa(len(job.Targets)), "status": job.Status}).Inc()
	}

	js, err := jsoni.Marshal(job)
	if err != nil {
		return err
	}
	return a.bds.Set(ctx, job.JobID, string(js), 30*60)
}

func (a *Scheduler) checkAndDownloadFile(kt *kit.Kit, filePath, signature string) error {
	// block until file download to avoid repeat download from another job
	a.fileLock.Acquire(filePath)
	defer a.fileLock.Release(filePath)
	if _, iErr := os.Stat(filePath); iErr != nil {
		if !os.IsNotExist(iErr) {
			return iErr
		}
		// not exists in feed server, download to local disk
		file, iErr := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if iErr != nil {
			return iErr
		}
		defer file.Close()

		reader, _, iErr := a.provider.Download(kt, signature)
		if iErr != nil {
			return iErr
		}
		defer reader.Close()
		if _, e := io.Copy(file, reader); e != nil {
			return e
		}
		if e := file.Sync(); e != nil {
			return e
		}
	}
	return nil
}

func getAsyncDownloadServerInfo(ctx context.Context, gseConf cc.GSE) (
	agentID string, containerID string, err error) {
	if gseConf.NodeAgentID != "" {
		// if serverAgentID configured, it measn feed server was deployed in binary mode, source is node
		agentID = gseConf.NodeAgentID
		return agentID, "", nil
	}
	// if serverAgentID not configured, it means feed server was deployed in container mode, source is container
	if gseConf.ClusterID == "" || gseConf.PodID == "" {
		return "", "", fmt.Errorf("server agent_id or (cluster_id and pod_id is required")
	}
	pod, qErr := bcs.QueryPod(ctx, gseConf.ClusterID, gseConf.PodID)
	if qErr != nil {
		return "", "", qErr
	}
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == gseConf.ContainerName {
			containerID = tools.SplitContainerID(container.ContainerID)
		}
	}
	if containerID == "" {
		return "", "", fmt.Errorf("server container %s not found in pod %s/%s",
			gseConf.ContainerName, gseConf.ClusterID, gseConf.PodID)
	}
	node, qErr := bcs.QueryNode(ctx, gseConf.ClusterID, pod.Spec.NodeName)
	if qErr != nil {
		return "", "", qErr
	}
	agentID = node.Labels[constant.LabelKeyAgentID]
	if agentID == "" {
		return "", "", fmt.Errorf("bk-agent-id not found in server node %s/%s", gseConf.ClusterID, pod.Spec.NodeName)
	}
	return agentID, containerID, nil
}

func (a *Scheduler) checkJobStatus(job *types.AsyncDownloadJob) error {

	if err := a.updateJobTargetsStatus(job); err != nil {
		return err
	}

	// if all targets are success, then the entire job is success
	if len(job.SuccessTargets) == len(job.Targets) {
		job.Status = types.AsyncDownloadJobStatusSuccess
		if err := a.updateAsyncDownloadJobStatus(a.ctx, job); err != nil {
			return err
		}
		return nil
	}
	// if all targets are in a final status and there is a failed target, then the entire job is failed
	if len(job.SuccessTargets)+len(job.FailedTargets) == len(job.Targets) {
		job.Status = types.AsyncDownloadJobStatusFailed
		if err := a.updateAsyncDownloadJobStatus(a.ctx, job); err != nil {
			return err
		}
		return nil
	}
	// if the entire job is not finished, check if timeout
	// if time out, set all the downloading status target to timeout
	if time.Since(job.ExecuteTime) > time.Duration(JobTimeoutSeconds)*time.Second {
		job.Status = types.AsyncDownloadJobStatusTimeout
		for k, v := range job.DownloadingTargets {
			job.TimeoutTargets[k] = v
		}
		for k := range job.DownloadingTargets {
			delete(job.DownloadingTargets, k)
		}
		if err := a.updateAsyncDownloadJobStatus(a.ctx, job); err != nil {
			return err
		}

		// TODO: need to check cancel gse task status ?
		timeoutTargets := make([]gse.TransferFileAgent, 0, len(job.TimeoutTargets))
		for _, content := range job.DownloadingTargets {
			timeoutTargets = append(timeoutTargets, gse.TransferFileAgent{
				BkAgentID:     content.DestAgentID,
				BkContainerID: content.DestContainerID,
			})
		}
		if _, err := gse.TerminateTransferFileTask(a.ctx, job.JobID, timeoutTargets); err != nil {
			logs.Errorf("cancel timeout transfer file task %s failed, err: %s", job.JobID, err.Error())
		}
		return nil
	}

	// if the entire job is not finished and not timeout, continue to update job status
	// so that downloading status target can be updated
	return a.updateAsyncDownloadJobStatus(a.ctx, job)
}

func (a Scheduler) updateJobTargetsStatus(job *types.AsyncDownloadJob) error {
	gseTaskResults, err := gse.TransferFileResult(a.ctx, job.GSETaskID)
	if err != nil {
		return err
	}

	// ! make sure that success + failed + downloading + timeout = all targets
	// success/failed/timeout is the final status, downloading is the intermediate status
	// so when set a target as success/failed/timeout, need to delete if from downloading list
	for _, result := range gseTaskResults {
		// upload result would not append to the targets list
		// if upload task failed, set all the task to failed
		// case in gse, if upload failed, all the download tasks must be failed
		if result.Content.Type == "upload" {
			if result.ErrorCode != 0 && result.ErrorCode != 115 {
				for k := range job.SuccessTargets {
					delete(job.SuccessTargets, k)
				}
				for k := range job.FailedTargets {
					delete(job.FailedTargets, k)
				}
				for k := range job.DownloadingTargets {
					delete(job.DownloadingTargets, k)
				}
				for k := range job.TimeoutTargets {
					delete(job.TimeoutTargets, k)
				}
				for _, target := range job.Targets {
					job.FailedTargets[fmt.Sprintf("%s:%s", target.AgentID, target.ContainerID)] = result.Content
				}
			}
		} else {
			// only download task would append to the targets list
			if result.ErrorCode == 0 {
				job.SuccessTargets[fmt.Sprintf("%s:%s", result.Content.DestAgentID, result.Content.DestContainerID)] =
					result.Content
				delete(job.DownloadingTargets,
					fmt.Sprintf("%s:%s", result.Content.DestAgentID, result.Content.DestContainerID))
			} else if result.ErrorCode == 115 {
				// If the result is 115 downloading state
				job.DownloadingTargets[fmt.Sprintf("%s:%s", result.Content.DestAgentID, result.Content.DestContainerID)] =
					result.Content
			} else {
				// other error code means failed
				job.FailedTargets[fmt.Sprintf("%s:%s", result.Content.DestAgentID, result.Content.DestContainerID)] =
					result.Content
				delete(job.DownloadingTargets,
					fmt.Sprintf("%s:%s", result.Content.DestAgentID, result.Content.DestContainerID))
			}
		}
	}
	return nil
}
