/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"context"
	"testing"
	"time"

	"bscp.io/pkg/kit"
	sfs "bscp.io/pkg/sf-share"
	"bscp.io/test/unit"
)

// TestJobs jobs unit test.
// test jobs support func[Push、HaveMore、Notifier、Current、Next、PopCurrent].
func TestJobs(t *testing.T) {
	unit.InitTestLogOptions(unit.DefaultTestLogDir)

	jobs := NewJobs()

	_, cancel := context.WithCancel(context.Background())
	jobCtx1 := &JobContext{
		Vas:     kit.NewVas(),
		Cancel:  cancel,
		JobType: PublishRelease,
		Descriptor: &sfs.ReleaseEventMetaV1{
			AppID:     1,
			ReleaseID: 1,
		},
		CursorID:    1,
		RetryPolicy: nil,
	}

	if err := jobs.Push(jobCtx1); err != nil {
		t.Errorf("jobs push failed, err: %v", err)
		return
	}

	if jobs.HaveMore() {
		t.Error("jobs HaveMore except false, but true")
		return
	}

	_, cancel = context.WithCancel(context.Background())
	jobCtx2 := &JobContext{
		Vas:     kit.NewVas(),
		Cancel:  cancel,
		JobType: PublishRelease,
		Descriptor: &sfs.ReleaseEventMetaV1{
			AppID:     1,
			ReleaseID: 2,
		},
		CursorID:    2,
		RetryPolicy: nil,
	}
	if err := jobs.Push(jobCtx2); err != nil {
		t.Errorf("jobs push failed, err: %v", err)
		return
	}

	if !jobs.HaveMore() {
		t.Error("jobs HaveMore except true, but false")
		return
	}

	notifier := jobs.Notifier()
	after := time.After(2 * time.Second)
	select {
	case <-after:
		t.Error("jobs Notifier except notify, but not has notify")
		return

	case <-notifier:
	}

	currentJob, exist := jobs.Current()
	if !exist {
		t.Error("jobs Current except exist job, but not exist")
		return
	}

	if currentJob.CursorID != jobCtx1.CursorID {
		t.Errorf("jobs Current job cursorID except %d, but %d", jobCtx1.CursorID, currentJob.CursorID)
		return
	}

	nextJob, exist := jobs.Next()
	if !exist {
		t.Error("jobs Next except exist job, but not exist")
		return
	}

	if nextJob.CursorID != jobCtx2.CursorID {
		t.Errorf("jobs Next job cursorID except %d, but %d", jobCtx2.CursorID, nextJob.CursorID)
		return
	}

	jobs.PopCurrent()
	jobs.Current()
	currentJob, exist = jobs.Current()
	if !exist {
		t.Error("jobs Current except exist job after PopCurrent, but not exist")
		return
	}

	if currentJob.CursorID != jobCtx2.CursorID {
		t.Errorf("jobs Current job cursorID except %d after PopCurrent, but %d", jobCtx2.CursorID, currentJob.CursorID)
		return
	}
}
