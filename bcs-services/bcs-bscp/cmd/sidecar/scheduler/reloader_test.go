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
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	pbci "bscp.io/pkg/protocol/core/config-item"
	"bscp.io/pkg/runtime/jsoni"
	sfs "bscp.io/pkg/sf-share"
	"bscp.io/test/unit"
)

// TestReloader test reloader notify reload func.
// Note: run test, will delete testWorkspace(./workspace) !!!
func TestReloader(t *testing.T) {
	unit.InitTestLogOptions(unit.DefaultTestLogDir)
	testClearWorkspace(t)

	workspace, err := testPrepareWorkspace()
	if err != nil {
		t.Errorf("prepare workspace failed, err: %v", err)
		return
	}

	appReloads := map[uint32]*sfs.Reload{
		testAppID: {
			ReloadType: table.ReloadWithFile,
			FileReloadSpec: &sfs.FileReloadSpec{
				ReloadFilePath: testReloadFilePath,
			},
		},
	}

	reloader, err := NewReloader(workspace, appReloads)
	if err != nil {
		t.Errorf("new reloader failed, err: %v", err)
		return
	}

	meta := &sfs.ReleaseEventMetaV1{
		AppID:     testAppID,
		ReleaseID: testReleaseID,
		CIMetas: []*sfs.ConfigItemMetaV1{
			{
				ConfigItemSpec: &pbci.ConfigItemSpec{
					Name: "mysql.yaml",
					Path: "/mysql",
				},
			},
			{
				ConfigItemSpec: &pbci.ConfigItemSpec{
					Name: "redis.ini",
					Path: "/redis/conf",
				},
			},
		},
	}
	err = reloader.NotifyReload(kit.NewVas(), meta)
	if err != nil {
		t.Errorf("reloader notify reload failed, err: %v", err)
		return
	}

	file, err := os.Open(testReloadFilePath)
	if err != nil {
		t.Errorf("open reload file failed, err: %v", err)
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("read reload file failed, err: %v", err)
		return
	}

	r := new(reloadMetadataV1)
	if err = jsoni.Unmarshal(bytes, r); err != nil {
		t.Errorf("unmarshal reload json failed, err: %v", err)
		return
	}

	if r.Version != reloadMetadataVersion ||
		len(r.Timestamp) == 0 ||
		r.AppID != testAppID ||
		r.ReleaseID != testReleaseID ||
		!strings.Contains(r.RootDirectory, "/bk-bscp/fileReleaseV1/1/2/3/configItems") ||
		r.ConfigItem[0] != "/mysql/mysql.yaml" ||
		r.ConfigItem[1] != "/redis/conf/redis.ini" {

		t.Errorf("reload file content not except")
		return
	}

	return
}
