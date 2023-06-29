/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/shutdown"
	sfs "bscp.io/pkg/sf-share"
	"bscp.io/pkg/tools"

	"golang.org/x/sync/semaphore"
)

// NewAppRuntime create a new app runtime instance.
func NewAppRuntime(appID uint32, workspace *AppFileWorkspace, downloader map[cc.StorageMode]Downloader, reloader Reloader) *AppRuntime {

	ar := &AppRuntime{
		appID:          appID,
		jobs:           NewJobs(),
		workspace:      workspace,
		repository:     downloader,
		currentRelease: new(currentRelease),
		reloader:       reloader,
	}

	go ar.waitForJobs()

	return ar
}

// AppRuntime handles all the app's runtime related tasks.
type AppRuntime struct {
	appID uint32

	jobs *Jobs

	currentRelease *currentRelease

	workspace  *AppFileWorkspace
	repository Downloader
	reloader   Reloader
}

// Push a new job to this app's runtime
func (ar *AppRuntime) Push(job *JobContext) error {
	return ar.jobs.Push(job)
}

func (ar *AppRuntime) waitForJobs() {
	for rid := range ar.jobs.Notifier() {
		logs.Infof("app %d receives new job, rid: %s", ar.appID, rid)

		// check if there are more jobs waiting to be handled.
		if ar.jobs.HaveMore() {
			// there are more jobs waiting to be handled,
			// then the current job need to be canceled, and pop the current job.
			current, exist := ar.jobs.Current()
			if !exist {
				continue
			}

			next, exist := ar.jobs.Next()
			if !exist {
				continue
			}

			switch next.JobType {
			case PublishRelease:
			default:
				logs.Errorf("unsupported app: %d job type: %s, skip, rid: %s", ar.appID, next.JobType, next.Vas.Rid)
				continue
			}

			// the second's job is valid, then cancel the current job and pop it.
			current.Cancel()
			ar.jobs.PopCurrent()
		}

		// run the new job
		current, exist := ar.jobs.Current()
		if !exist {
			logs.Errorf("received new job signal, but do not find it, this should not happen, rid: %s", rid)
			continue
		}

		switch current.JobType {
		case PublishRelease:
			go ar.loopAppPublishReleaseJob(current)

		default:
			logs.Errorf("unsupported app: %d job type: %s, skip, rid: %s", ar.appID, current.JobType, current.Vas.Rid)
			continue
		}
	}
}

func (ar *AppRuntime) loopAppPublishReleaseJob(job *JobContext) {
	logs.Infof("start handle app: %d publish release: %d job, rid: %s", ar.appID, job.Descriptor.ReleaseID, job.Vas.Rid)

	notifier := shutdown.AddNotifier()
	for {
		select {
		case <-notifier.Signal:
			logs.Infof("received shutdown signal, app runtime stop loop app: %d release change job, rid: %s",
				ar.appID, job.Vas.Rid)

			notifier.Done()
			return

		case <-job.Vas.Ctx.Done():
			logs.Warnf("received cancel job request, app runtime stop loop app: %d release: %d change job, rid: %s",
				ar.appID, job.Descriptor.ReleaseID, job.Vas.Rid)
			return

		default:
		}

		if ar.currentRelease.ReleaseID() == job.Descriptor.ReleaseID {
			logs.Infof("app: %d, job's release is same with the current release, skip, rid: %s", ar.appID, job.Vas.Rid)
			return
		}

		if err := ar.workspace.PrepareReleaseDirectory(job.Descriptor.ReleaseID); err != nil {
			logs.Errorf("prepare app: %d, job's release: %d directory failed, err: %s, rid: %s", ar.appID,
				job.Descriptor.ReleaseID, err, job.Vas.Rid)
			job.RetryPolicy.Sleep()
			continue
		}

		vas := &kit.Vas{
			Rid: fmt.Sprintf("%s-retry-%d", job.Vas.Rid, job.RetryPolicy.RetryCount()),
			Ctx: job.Vas.Ctx,
		}

		logs.Infof("start handle app: %d, release: %d job for %d times, rid: %s", ar.appID, job.Descriptor.ReleaseID,
			job.RetryPolicy.RetryCount(), vas.Rid)

		if ar.loopAppReleaseForOnce(vas, job.Descriptor) {
			logs.Warnf("loop app[%d] current published release[%d] job failed, retry later, rid: %s", ar.appID,
				job.Descriptor.ReleaseID, vas.Rid)
			job.RetryPolicy.Sleep()
			continue
		}

		ar.currentRelease.Set(job.Descriptor)

		logs.Infof("handle app: %d release: %d job success, rid: %s", ar.appID, job.Descriptor.ReleaseID, vas.Rid)

		return
	}

}

// loopAppReleaseForOnce try download the app's current release's configuration items for once, and
// save the related metadata to the local file under the app's workspace.
func (ar *AppRuntime) loopAppReleaseForOnce(vas *kit.Vas, desc *sfs.ReleaseEventMetaV1) (needRetry bool) {
	// lock the job with file lock
	lock, err := LockFile(ar.workspace.LockFile(desc.ReleaseID), true)
	if err != nil {
		logs.Errorf("lock app release file lock failed, err: %v, rid: %s", err, vas.Rid)
		return true
	}

	defer func() {
		if err = UnlockFile(lock); err != nil {
			logs.Errorf("unlock app release file lock failed, err: %v, rid: %s", err, vas.Rid)
		}
	}()

	// check the readiness of the release's configuration items, maybe it is already ready,
	// such as rollback app's release version, or it has already been downloaded successfully
	// in the previous loop.
	readiness, err := ar.getReadinessConfigItems(vas, desc.ReleaseID, desc.CIMetas)
	if err != nil {
		logs.Errorf("get readiness of CI failed, err: %v, rid: %s", err, vas.Rid)
		return true
	}

	// Note: check if the CI has already been downloaded before in the other release,
	// if yes, copy it to this release without download it from repository.

	start := time.Now()
	// download the configuration items from repository one by one or batch.
	if err := ar.downloadReleasedCI(vas, readiness, desc); err != nil {
		logs.Errorf("download app: %d, release: %d CI failed, err: %v, rid: %s", ar.appID, desc.ReleaseID, err, vas.Rid)
		return true
	}

	// write release meta
	meta := &appReleaseMetadata{
		DownloadedAt: time.Now().Format(constant.TimeStdFormat),
		CostTime:     time.Since(start).String(),
		Release:      desc,
	}
	if err := saveAppReleaseMetadata(meta, ar.workspace.MetadataFile(desc.ReleaseID)); err != nil {
		logs.Errorf("write app: %d, release: %d metadata to file failed, err: %v, rid: %s", ar.appID, desc.ReleaseID,
			err, vas.Rid)
		return true
	}

	if err := ar.reloader.NotifyReload(vas, desc); err != nil {
		logs.Errorf("notify app: %d to reload release: %d config file failed, err: %v, rid: %s", ar.appID,
			desc.ReleaseID, err, vas.Rid)
		return true
	}

	return false

}

func (ar *AppRuntime) downloadReleasedCI(vas *kit.Vas, readiness map[uint32]bool, desc *sfs.ReleaseEventMetaV1) error {

	start := time.Now()
	sem := semaphore.NewWeighted(3)
	wg := sync.WaitGroup{}
	var hitErr error
	for idx, one := range desc.CIMetas {
		if hitErr != nil {
			return hitErr
		}

		if ready, exist := readiness[one.ID]; exist && ready {
			logs.Infof("app: %d, release: %d, CI: %s/%s is already downloaded, skip download, rid: %s", ar.appID,
				desc.ReleaseID, one.ConfigItemSpec.Path, one.ConfigItemSpec.Name, vas.Rid)

			spec := one.ConfigItemSpec
			filePath := ar.workspace.ConfigItemFile(desc.ReleaseID, spec.Path, spec.Name)
			if err := ar.workspace.SetFilePermission(filePath, spec.Permission); err != nil {
				logs.Errorf("app: %d, release: %d, CI: %s/%s already downloaded file try again set file "+
					"permission failed, err: %v, rid: %s", ar.appID, one.ID, spec.Path, spec.Name, err, vas.Rid)
			}
			continue
		}

		if err := sem.Acquire(vas.Ctx, 1); err != nil {
			return fmt.Errorf("app: %d, CI: %d acquire lock failed, err: %v", ar.appID, one.ID, err)
		}

		wg.Add(1)

		subVas := &kit.Vas{
			Rid: fmt.Sprintf("%s-ci-%d", vas.Rid, one.ID),
			// Note: do not create new context.
			Ctx: vas.Ctx,
		}

		go func(vas *kit.Vas, ciMeta *sfs.ConfigItemMetaV1) {
			defer func() {
				wg.Done()
				sem.Release(1)
			}()

			spec := ciMeta.ConfigItemSpec
			filePath := ar.workspace.ConfigItemFile(desc.ReleaseID, spec.Path, spec.Name)
			if err := ar.workspace.PrepareCIDirectory(desc.ReleaseID, spec.Path); err != nil {
				hitErr = err
				logs.Errorf("download app: %d CI[%d, %s/%s], but prepare CI directory failed, err: %v, rid: %s",
					ar.appID, ciMeta.ID, spec.Path, spec.Name, err, vas.Rid)
				return
			}

			if err := ar.repository.Download(vas, ciMeta.RepositoryPath, ciMeta.ContentSpec.ByteSize, filePath); err != nil {
				hitErr = err
				logs.Errorf("download app: %d, CI[%d, %s/%s] failed, uri: %s, err: %v, rid: %s", ar.appID, ciMeta.ID,
					spec.Path, spec.Name, ciMeta.RepositoryPath, err, vas.Rid)
				return
			}

			if err := ar.workspace.SetFilePermission(filePath, spec.Permission); err != nil {
				logs.Errorf("download app: %d, CI[%d, %s/%s], but set file permission failed, err: %v, rid: %s",
					ar.appID, ciMeta.ID, spec.Path, spec.Name, err, vas.Rid)
				return
			}

			logs.Infof("download app: %d, CI[%d] %s success, rid: %s", ar.appID, ciMeta.ID, spec.Name, vas.Rid)

		}(subVas, desc.CIMetas[idx])
	}

	wg.Wait()

	if hitErr != nil {
		return hitErr
	}

	logs.Infof("download app: %d all release[%d] configure items success, cost: %s rid: %s", ar.appID, desc.ReleaseID,
		time.Since(start).String(), vas.Rid)

	return nil
}

// getReadinessConfigItems get each configuration item's file is already download and verified status.
func (ar *AppRuntime) getReadinessConfigItems(vas *kit.Vas, releaseID uint32, ciList []*sfs.ConfigItemMetaV1) (
	map[uint32]bool, error) {

	lock := sync.Mutex{}
	readiness := make(map[uint32]bool)
	sem := semaphore.NewWeighted(5)
	wg := sync.WaitGroup{}
	var hitError error

	for _, one := range ciList {
		if hitError != nil {
			return nil, hitError
		}

		_ = sem.Acquire(context.TODO(), 1)
		wg.Add(1)

		go func(release uint32, ci *sfs.ConfigItemMetaV1) {

			defer func() {
				sem.Release(1)
				wg.Done()
			}()

			ready, err := ar.isConfigItemReady(release, ci)
			if err != nil {
				hitError = err
				logs.Errorf("check app[%d] configure item[%s] ready failed, err: %v, rid: %s", ar.appID,
					ci.ConfigItemSpec.Name, err, vas.Rid)
				return
			}

			lock.Lock()
			readiness[ci.ID] = ready
			lock.Unlock()

		}(releaseID, one)
	}

	wg.Wait()

	if hitError != nil {
		return nil, hitError
	}

	return readiness, nil
}

// isConfigItemReady return true if it is already exist and its SHA256 is same with expected at the same time.
func (ar *AppRuntime) isConfigItemReady(releaseID uint32, ci *sfs.ConfigItemMetaV1) (bool, error) {
	// verify the config content is exist or not in the local.
	ciFile := ar.workspace.ConfigItemFile(releaseID, ci.ConfigItemSpec.Path, ci.ConfigItemSpec.Name)
	_, err := os.Stat(ciFile)
	if err != nil {
		if os.IsNotExist(err) {
			// content is not exist
			return false, nil
		}

		return false, err
	}

	sha, err := tools.FileSHA256(ciFile)
	if err != nil {
		return false, fmt.Errorf("check configuration item's SHA256 failed, err: %v", err)
	}

	if sha != ci.ContentSpec.Signature {
		return false, nil
	}

	return true, nil

}
