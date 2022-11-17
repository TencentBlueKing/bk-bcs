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
	"fmt"
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

// TestAppRuntime app runtime unit test.
// test app runtime push func.
// Note: run test, will delete testWorkspace(./workspace) !!!
func TestAppRuntime(t *testing.T) {
	unit.InitTestLogOptions(unit.DefaultTestLogDir)
	testClearWorkspace(t)

	downloadAddr := os.Getenv(constant.EnvUnitTestDownloadAddr)
	if len(downloadAddr) == 0 {
		t.Error("downloader download address not set")
		return
	}

	appWS, dl, reloader, err := testAppRuntimePrepare(testWorkspace, testBizID, testAppID)
	if err != nil {
		t.Errorf("app runtime prepare failed, err: %v", err)
		return
	}

	appRuntime := NewAppRuntime(testAppID, appWS, dl, reloader)

	_, cancel := context.WithCancel(context.Background())
	jobCtx := &JobContext{
		Vas:     kit.NewVas(),
		Cancel:  cancel,
		JobType: PublishRelease,
		Descriptor: &sfs.ReleaseEventMetaV1{
			AppID:     testAppID,
			ReleaseID: testReleaseID,
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
		RetryPolicy: nil,
	}
	if err = appRuntime.Push(jobCtx); err != nil {
		t.Errorf("app runtime push job failed, err: %v", err)
		return
	}

	sleepTime := 2 * time.Second
	retry := 0
	for {
		time.Sleep(sleepTime)
		if retry == 3 {
			t.Errorf("wait app runtime handle job timeout")
			return
		}

		if appRuntime.currentRelease.ReleaseID() == 0 {
			retry++
			sleepTime = sleepTime * 2
			continue
		}

		break
	}

	ciSpec := jobCtx.Descriptor.CIMetas[0].ConfigItemSpec
	ciPath := fmt.Sprintf("%s/%s/fileReleaseV1/%d/%d/%d/configItems/%s/%s", testWorkspace, bscpWorkspaceDir,
		testBizID, testAppID, testReleaseID, ciSpec.Path, ciSpec.Name)

	if _, err = os.Stat(ciPath); err != nil {
		t.Errorf("check config item file failed, err: %v", err)
		return
	}

	lockFilePath := fmt.Sprintf("%s/%s/fileReleaseV1/%d/%d/%d/file.lock", testWorkspace, bscpWorkspaceDir,
		testBizID, testAppID, testReleaseID)

	if _, err = os.Stat(lockFilePath); err != nil {
		t.Errorf("check lock file failed, err: %v", err)
		return
	}

	metaFilePath := fmt.Sprintf("%s/%s/fileReleaseV1/%d/%d/%d/metadata.json", testWorkspace,
		bscpWorkspaceDir, testBizID, testAppID, testReleaseID)

	if _, err = os.Stat(metaFilePath); err != nil {
		t.Errorf("check metadata file failed, err: %v", err)
		return
	}
}

func testAppRuntimePrepare(rootDir string, bizID, appID uint32) (*AppFileWorkspace, Downloader, Reloader, error) {
	sw := cc.SidecarWorkspace{
		RootDirectory: rootDir,
		PurgePolicy: &cc.SidecarPurgePolicy{
			EnableAutoClean:      false,
			MaxSizeMB:            0,
			AutoCleanIntervalMin: 0,
		},
	}

	spec := cc.SidecarAppSpec{
		BizID: bizID,
		Applications: []cc.AppMetadata{
			{
				AppID:     appID,
				Namespace: "workspace-test",
				Uid:       "workspace-test",
				Labels: map[string]string{
					"biz": "1",
				},
			},
		},
	}

	workspace, err := NewWorkspace(sw, spec)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("new workspace failed, err: %v", err)
	}

	reloader, err := testPrepareReloader(workspace)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("prepare reloader failed, err: %v", err)
	}

	appWorkspace, err := workspace.AppFileReleaseWorkspace(appID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("new app workspace failed, err: %v", err)
	}

	auth := cc.SidecarAuthentication{
		User:  testDownloadUser,
		Token: testDownloadToken,
	}
	downloader, err := InitDownloader(auth, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("init downloader failed, err: %v", err)
	}

	return appWorkspace, downloader, reloader, nil
}
