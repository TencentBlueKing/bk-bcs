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

// Package scheduler NOTES
package scheduler

import (
	"context"
	"fmt"

	"bscp.io/cmd/sidecar/stream"
	"bscp.io/cmd/sidecar/stream/types"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/jsoni"
	sfs "bscp.io/pkg/sf-share"
)

// Interface defines all the supported operation by the scheduler.
type Interface interface {
	OnAppReleaseChange(event *types.ReleaseChangeEvent)
	CurrentRelease(appID uint32) (releaseID uint32, cursorID uint32, exist bool)
}

// InitScheduler initialize the scheduler
func InitScheduler(opt *SchOptions) (Interface, error) {
	ws, err := NewWorkspace(opt.Settings.Workspace, opt.Settings.AppSpec)
	if err != nil {
		return nil, fmt.Errorf("new app workspace failed, err: %v", err)
	}

	reloader, err := NewReloader(ws, opt.AppReloads)
	if err != nil {
		return nil, fmt.Errorf("new reloader failed, err: %v", err)
	}

	factory, err := InitAppFactory(ws, reloader, opt.Settings, opt.RepositoryTLS)
	if err != nil {
		return nil, fmt.Errorf("initialize app factory failed, err: %v", err)
	}

	return &Scheduler{
		stream:  opt.Stream,
		factory: factory,
	}, nil
}

// Scheduler works to schedule the holds applications' release pulling jobs.
type Scheduler struct {
	stream  stream.Interface
	factory *AppFactory
}

// OnAppReleaseChange is used to receive app release change event from the upstream.
func (sch *Scheduler) OnAppReleaseChange(event *types.ReleaseChangeEvent) {

	// parse payload according the api version.
	pl := new(sfs.ReleaseChangePayload)
	if err := jsoni.Unmarshal(event.Payload, pl); err != nil {
		// TODO: sch.stream.FireEvent()
		logs.Errorf("decode release change event payload failed, skip the event, err: %v, rid: %s", err, event.Rid)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	job := &JobContext{
		Vas: &kit.Vas{
			Rid: event.Rid,
			Ctx: ctx,
		},
		Cancel:      cancel,
		JobType:     PublishRelease,
		Descriptor:  pl.ReleaseMeta,
		CursorID:    pl.CursorID,
		RetryPolicy: defaultRetryPolicy,
	}

	if err := sch.factory.PushJob(pl.ReleaseMeta.AppID, job); err != nil {
		logs.Errorf("push app: %d, release: %d job to app factory failed, detail: %s, err: %v, rid: %s",
			pl.ReleaseMeta.AppID, pl.ReleaseMeta.ReleaseID, event.Payload, err, event.Rid)
		return
	}

	logs.Infof("received app: %d release: %d change job, push it to app factory success, rid: %s", pl.ReleaseMeta.AppID,
		pl.ReleaseMeta.ReleaseID, event.Rid)

	logs.V(1).Infof("app release change event payload: %s, rid: %s", event.Payload, event.Rid)

	return
}

// CurrentRelease returns the current release metadata if it exists for an app.
// if it not exists, then the returned meta is nil.
func (sch *Scheduler) CurrentRelease(appID uint32) (releaseID uint32, cursorID uint32, exist bool) {
	if !sch.factory.Have(appID) {
		return 0, 0, false
	}

	return sch.factory.CurrentRelease(appID)
}
