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
	"os"
	"testing"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/dal/table"
	sfs "bscp.io/pkg/sf-share"
)

// scheduler unit test use const defines.
const (
	testBizID           uint32 = 1
	testAppID           uint32 = 2
	testReleaseID       uint32 = 3
	testWorkspace              = "./workspace"
	testSmallFileSha256        = "bc610f99e0cbc17eddb2e8d15bca38b731fd3df15e325795c0ea2da69abc6ac8" // size: 1021 Byte
	testBigFileSha256          = "dd65dd17c4b8bdfb2f8135cbd3257034ec4a73f2c8b6baafca09058eb43bb21b" // size: 93965 Byte
	testSmallFileUri           = "/generic/bscp/bscp-download-test/file/1021"
	testBigFileUri             = "/generic/bscp/bscp-download-test/file/93965"
	testCIUser                 = "root"
	testCIUserGroup            = "root"
	testDownloadUser           = "downloader"
	testDownloadToken          = "downloader-token"
	testReloadFilePath         = "/data/bscp/reload/reload.json"
)

func testClearWorkspace(t *testing.T) {
	if err := os.RemoveAll(testWorkspace); err != nil {
		t.Fatalf("remove root dir failed, err: %v", err)
	}
}

func testPrepareWorkspace() (*RuntimeWorkspace, error) {
	setting := cc.SidecarSetting{
		Workspace: cc.SidecarWorkspace{
			RootDirectory: testWorkspace,
		},
		AppSpec: cc.SidecarAppSpec{
			BizID: testBizID,
			Applications: []cc.AppMetadata{
				{
					AppID:     testAppID,
					Namespace: "app-factory-2",
					Uid:       "app-factory-uid-2",
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

	return NewWorkspace(setting.Workspace, setting.AppSpec)
}

func testPrepareReloader(ws *RuntimeWorkspace) (Reloader, error) {
	appReloads := map[uint32]*sfs.Reload{
		testAppID: &sfs.Reload{
			ReloadType: table.ReloadWithFile,
			FileReloadSpec: &sfs.FileReloadSpec{
				ReloadFilePath: testReloadFilePath,
			},
		},
	}

	return NewReloader(ws, appReloads)
}
