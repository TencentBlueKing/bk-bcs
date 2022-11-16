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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"bscp.io/pkg/cc"
	"bscp.io/test/unit"
)

// TestWorkspace workspace unit test.
// 1. test workspace dir whether right.
// 2. test app workspace dir whether right.
// 3. test app's release workspace dir whether right.
// 4. test config-item's path dir whether right.
// Note: run test, will delete testWorkspace(./workspace) !!!
func TestWorkspace(t *testing.T) {
	unit.InitTestLogOptions(unit.DefaultTestLogDir)
	testClearWorkspace(t)

	ciPath := "etc"
	resultPath := fmt.Sprintf("%s/%s/fileReleaseV1/%d/%d/%d/configItems/%s", testWorkspace, bscpWorkspaceDir,
		testBizID, testAppID, testReleaseID, ciPath)

	sw := cc.SidecarWorkspace{
		RootDirectory: testWorkspace,
		PurgePolicy: &cc.SidecarPurgePolicy{
			EnableAutoClean:      false,
			MaxSizeMB:            0,
			AutoCleanIntervalMin: 0,
		},
	}

	spec := cc.SidecarAppSpec{
		BizID: testBizID,
		Applications: []cc.AppMetadata{
			{
				AppID:     testAppID,
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
		t.Errorf("new workspace failed, err: %v", err)
		return
	}

	appWorkspace, err := workspace.AppFileReleaseWorkspace(testAppID)
	if err != nil {
		t.Errorf("new app workspace failed, err: %v", err)
		return
	}

	if err = appWorkspace.PrepareReleaseDirectory(testReleaseID); err != nil {
		t.Errorf("prapare release dir failed, err: %v", err)
		return
	}

	if err = appWorkspace.PrepareCIDirectory(testReleaseID, ciPath); err != nil {
		t.Errorf("prepare ci dir failed, err: %v", err)
		return
	}

	_, err = os.Stat(resultPath)
	if err != nil {
		t.Errorf("check ci file failed, err: %v", err)
		return
	}

	metadataFilePath, err := filepath.Abs(fmt.Sprintf("%s/%s/fileReleaseV1/%d/%d/%d/metadata.json", testWorkspace,
		bscpWorkspaceDir, testBizID, testAppID, testReleaseID))
	if err != nil {
		t.Errorf("file path abs failed, err: %v", err)
		return
	}

	if appWorkspace.MetadataFile(testReleaseID) != metadataFilePath {
		t.Errorf("app workspace metadata file path except %s, but %s", metadataFilePath,
			appWorkspace.MetadataFile(testReleaseID))
		return
	}

	lockFilePath, err := filepath.Abs(fmt.Sprintf("%s/%s/fileReleaseV1/%d/%d/%d/file.lock", testWorkspace,
		bscpWorkspaceDir, testBizID, testAppID, testReleaseID))
	if err != nil {
		t.Errorf("file path abs failed, err: %v", err)
		return
	}

	if appWorkspace.LockFile(testReleaseID) != lockFilePath {
		t.Errorf("app workspace lock file path except %s, but %s", lockFilePath, appWorkspace.LockFile(testReleaseID))
		return
	}
}
