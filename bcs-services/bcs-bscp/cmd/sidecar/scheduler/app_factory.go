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
	"errors"
	"fmt"

	"bscp.io/pkg/cc"
	sfs "bscp.io/pkg/sf-share"
)

// InitAppFactory initialize the application factory.
func InitAppFactory(ws *RuntimeWorkspace, reloader Reloader, settings cc.SidecarSetting,
	repoTLS *sfs.TLSBytes) (*AppFactory, error) {

	repository, err := InitDownloader(settings.Upstream.Authentication, repoTLS)
	if err != nil {
		return nil, fmt.Errorf("init repository downloader failed, err: %v", err)
	}

	af := &AppFactory{
		pool:      make(map[uint32]*AppRuntime),
		workspace: ws,
	}

	for _, one := range settings.AppSpec.Applications {
		appWorkspace, err := ws.AppFileReleaseWorkspace(one.AppID)
		if err != nil {
			return nil, fmt.Errorf("prepare app workspace failed, err: %v", err)
		}

		af.pool[one.AppID] = NewAppRuntime(one.AppID, appWorkspace, repository, reloader)
	}

	return af, nil
}

// AppFactory is the application's factory to handle application's jobs.
type AppFactory struct {
	// map[appID]*Jobs
	pool      map[uint32]*AppRuntime
	workspace *RuntimeWorkspace
}

// Have check if the specific app is exists in the factory or not.
func (af *AppFactory) Have(appID uint32) bool {
	_, exist := af.pool[appID]
	return exist
}

// PushJob push the job to the app factory if the app is exists under
// this app factory.
// Note: PushJob should be called after Have function has been called and
// the app is already exist under this app factory.
func (af *AppFactory) PushJob(appID uint32, jc *JobContext) error {
	if err := jc.Validate(); err != nil {
		return err
	}

	appRuntime, exist := af.pool[appID]
	if !exist {
		return errors.New("app not exist under the app factory")
	}

	if err := appRuntime.Push(jc); err != nil {
		return err
	}

	return nil
}

// CurrentRelease returns the current metadata if it exists.
// if it not exists, then the returned meta is nil.
func (af *AppFactory) CurrentRelease(appID uint32) (releaseID uint32, cursorID uint32, exist bool) {

	appRuntime, exist := af.pool[appID]
	if !exist {
		return 0, 0, false
	}

	return appRuntime.currentRelease.Release()
}
