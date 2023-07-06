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
	"os"
	"testing"
	"time"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/kit"
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbcontent "bscp.io/pkg/protocol/core/content"
	sfs "bscp.io/pkg/sf-share"
	"bscp.io/test/unit"
)

// TestAppFactory app factory unit test.
// test AppFactory PushJob、CurrentRelease func.
// Note: run test, will delete testWorkspace(./workspace) !!!
func TestAppFactory(t *testing.T) {
	unit.InitTestLogOptions(unit.DefaultTestLogDir)
	testClearWorkspace(t)

	downloadAddr := os.Getenv(constant.EnvUnitTestDownloadAddr)
	if len(downloadAddr) == 0 {
		t.Error("downloader download address not set")
		return
	}

	setting := cc.SidecarSetting{
		Workspace: cc.SidecarWorkspace{
			RootDirectory: testWorkspace,
		},
		AppSpec: cc.SidecarAppSpec{
			BizID: testBizID,
			Applications: []cc.AppMetadata{
				{
					AppID:     testAppID,
					Namespace: "app-factory-1",
					Uid:       "app-factory-uid-1",
				},
			},
		},
		Upstream: cc.SidecarUpstream{
			Authentication: cc.SidecarAuthentication{
				User:  testDownloadUser,
				Token: testDownloadToken,
			},
		},
	}

	ws, err := NewWorkspace(setting.Workspace, setting.AppSpec)
	if err != nil {
		t.Errorf("new app workspace failed, err: %v", err)
		return
	}

	reloader, err := testPrepareReloader(ws)
	if err != nil {
		t.Errorf("prepare reloader failed, err: %v", err)
		return
	}

	factory, err := InitAppFactory(ws, reloader, setting, nil)
	if err != nil {
		t.Errorf("init app factory failed, err: %v", err)
		return
	}

	if !factory.Have(setting.AppSpec.Applications[0].AppID) {
		t.Errorf("app factory except has %d app, but not has", setting.AppSpec.Applications[0].AppID)
		return
	}

	if factory.Have(666) {
		t.Errorf("app factory except not has %d app, but has", 666)
		return
	}

	release, cursorID, exist := factory.CurrentRelease(setting.AppSpec.Applications[0].AppID)
	if exist || release != 0 || cursorID != 0 {
		t.Errorf("app factory except not has %d app' release, but has", setting.AppSpec.Applications[0].AppID)
		return
	}

	_, cancel := context.WithCancel(context.Background())
	jobCtx := &JobContext{
		Vas:     kit.NewVas(),
		Cancel:  cancel,
		JobType: PublishRelease,
		Descriptor: &sfs.ReleaseEventMetaV1{
			AppID:     setting.AppSpec.Applications[0].AppID,
			ReleaseID: 3,
			CIMetas: []*sfs.ConfigItemMetaV1{
				{
					ID: 4,
					ContentSpec: &pbcontent.ContentSpec{
						Signature: testSmallFileSha256,
						ByteSize:  1021,
					},
					ConfigItemSpec: &pbci.ConfigItemSpec{
						Name:     "server.yaml",
						Path:     "/etc",
						FileType: "yaml",
						FileMode: "unix",
						Memo:     "unit test",
						Permission: &pbci.FilePermission{
							User:      testCIUser,
							UserGroup: testCIUserGroup,
							Privilege: "755",
						},
					},
					RepositoryPath: &sfs.RepositorySpecV1{
						Path: testSmallFileUri,
					},
				},
			},
			Repository: &sfs.RepositoryV1{
				Root: downloadAddr,
				TLS:  nil,
			},
		},
		CursorID:    1,
		RetryPolicy: defaultRetryPolicy,
	}
	if err = factory.PushJob(setting.AppSpec.Applications[0].AppID, jobCtx); err != nil {
		t.Errorf("app factory push job failed, err: %v", err)
		return
	}

	sleepTime := 2 * time.Second
	retry := 0
	for {
		time.Sleep(sleepTime)
		if retry == 3 {
			t.Errorf("app factory except has %d app's release, but not has", setting.AppSpec.Applications[0].AppID)
			return
		}

		release, cursorID, exist = factory.CurrentRelease(setting.AppSpec.Applications[0].AppID)
		// TODO: cursorID has not been used yet.
		if !exist || release != jobCtx.Descriptor.ReleaseID || cursorID != 0 {
			retry++
			sleepTime = sleepTime * 2
			continue
		}

		break
	}
}
