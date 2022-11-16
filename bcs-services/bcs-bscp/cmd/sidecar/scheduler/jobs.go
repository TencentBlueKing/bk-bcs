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
	"errors"
	"fmt"
	"sync"

	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	sfs "bscp.io/pkg/sf-share"
	"bscp.io/pkg/tools"
)

var defaultRetryPolicy = tools.NewRetryPolicy(5, [2]uint{500, 15000})

// NewJobs new a jobs instances
func NewJobs() *Jobs {
	// Note: size should >=2
	size := 10
	return &Jobs{
		size:     size,
		notifier: make(chan string, size+5),
		jobs:     make([]*JobContext, 0),
	}
}

// Jobs defines all the job received from the upstream or generated
// by the local scheduler.
// All the job must be executed one by one within the same app to ensure
// the app's config is published or rollback with the extremely correct
// order.
type Jobs struct {
	lock sync.RWMutex

	// size is the size of this job's pool size.
	// size should be >2, otherwise, Push will panic.
	size int
	// notifier 's size should be a little bigger than the value of job's size.
	// so that more event's can be accepted to trigger the job's rolling.
	notifier chan string

	jobs []*JobContext
}

// Notifier return the channel to receive the jobs change event.
// the channel carried with the job's request id.
func (j *Jobs) Notifier() <-chan string {
	return j.notifier
}

// HaveMore return true if there is still have >=2 jobs waiting to execute.
// which means more jobs is waiting to execute except the "current" job.
func (j *Jobs) HaveMore() bool {
	j.lock.RLock()
	defer j.lock.RUnlock()

	if len(j.jobs) <= 1 {
		return false
	}

	// still have >=2 jobs to do
	return true
}

// Next return the job witch is waiting to be handled, as is the
// second jobs in the pool.
// Note: only when the second job is exists, then the returned job
// is no-nil, so check the returned exists' value at before using
// the returned job.
func (j *Jobs) Next() (job *JobContext, exist bool) {
	j.lock.Lock()
	defer j.lock.Unlock()

	if len(j.jobs) >= 2 {
		return j.jobs[1], true
	}

	return nil, false
}

// Current return the current job which is waiting to execute.
// Note: the returned exist must be checked, only when the returned
// exist value is true, then the returned job is not nil.
func (j *Jobs) Current() (job *JobContext, exist bool) {
	j.lock.Lock()
	defer j.lock.Unlock()

	if len(j.jobs) <= 0 {
		return nil, false
	}

	return j.jobs[0], true
}

// PopCurrent pop the current job
func (j *Jobs) PopCurrent() {
	j.lock.Lock()
	defer j.lock.Unlock()

	if len(j.jobs) <= 0 {
		return
	}

	// discard the current job
	j.jobs = j.jobs[1:]
	return
}

// Push a new job to the job list's end.
// if the job's pool is full, the second one will be removed automatically.
func (j *Jobs) Push(job *JobContext) error {
	if job == nil {
		return errors.New("empty job context")
	}

	if len(job.JobType) == 0 {
		return errors.New("empty job operator")
	}

	// set default retry policy
	if job.RetryPolicy == nil {
		job.RetryPolicy = defaultRetryPolicy
	}

	j.lock.Lock()
	defer j.lock.Unlock()

	if len(j.jobs) >= j.size {
		// job's pool is already full
		// pop the second-oldest job, cause the latest one is
		// being handled.
		rearrangedJobs := []*JobContext{j.jobs[0]}
		rearrangedJobs = append(rearrangedJobs, j.jobs[2:]...)
		j.jobs = rearrangedJobs
	}

	j.jobs = append(j.jobs, job)

	// notify a new job is arrived.
	select {
	case j.notifier <- job.Vas.Rid:
	default:
		// channel is full, drop it directly, because notifier length
		// is a littler bigger that the job pool's length.
	}

	logs.Infof("received new job, rid: %s", job.Vas.Rid)

	return nil
}

// JobType defines the job's basic operate type
type JobType string

const (
	// PublishRelease means this is a job which describe the app's current release configuration
	// has changed.
	PublishRelease JobType = "ReleaseChange"
)

// Validate check the job type is valid or not.
func (jt JobType) Validate() error {
	switch jt {
	case PublishRelease:
	default:
		return fmt.Errorf("unsupported job type: %s", jt)
	}

	return nil
}

// JobContext define a job's executive context.
type JobContext struct {
	Vas *kit.Vas
	// Cancel is used to cancel the context, which should be
	// generated form the Ctx under the Vas.
	Cancel context.CancelFunc

	// job's operate type
	JobType JobType

	// job's detail info
	Descriptor *sfs.ReleaseEventMetaV1
	// CursorID is the event's cursor which is bound to the Descriptor
	CursorID uint32

	// retry policy controls how to retry this job.
	RetryPolicy *tools.RetryPolicy
}

// Validate the job context is valid or not
func (jc *JobContext) Validate() error {
	if jc.Vas == nil {
		return errors.New("job context vas is nil")
	}

	if jc.Cancel == nil {
		return errors.New("job context's cancel function is nil")
	}

	if err := jc.JobType.Validate(); err != nil {
		return fmt.Errorf("invalid job context's job type, err: %v", err)
	}

	if jc.Descriptor == nil {
		return errors.New("job context's descriptor is nil")
	}

	if jc.RetryPolicy == nil {
		return errors.New("job context's retry policy is nil")
	}

	return nil
}
