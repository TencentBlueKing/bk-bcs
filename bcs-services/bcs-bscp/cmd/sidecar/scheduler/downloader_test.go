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
	"fmt"
	"os"
	"testing"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/tools"
	"bscp.io/test/unit"
)

// TestDownloader downloader unit test.
// 1. test download small file by all download.
// 2. test download big file by range download.
// Note: run test, will delete testWorkspace(./workspace) !!!
func TestDownloader(t *testing.T) {
	unit.InitTestLogOptions(unit.DefaultTestLogDir)
	testClearWorkspace(t)

	downloadAddr := os.Getenv(constant.EnvUnitTestDownloadAddr)
	if len(downloadAddr) == 0 {
		t.Error("downloader download address not set")
		return
	}

	smallFileAddr := fmt.Sprintf("%s%s", downloadAddr, testSmallFileUri) // file size: 1021 Byte
	smallFileToFile := fmt.Sprintf("%s/small", testWorkspace)

	bigFileAddr := fmt.Sprintf("%s%s", downloadAddr, testBigFileUri) // file size: 93965 Byte
	bigFileToFile := fmt.Sprintf("%s/big", testWorkspace)

	if err := os.MkdirAll(testWorkspace, os.ModePerm); err != nil {
		t.Errorf("mkdir %s failed, err: %v", testWorkspace, err)
		return
	}

	auth := cc.SidecarAuthentication{
		User:  testDownloadUser,
		Token: testDownloadToken,
	}
	downloader, err := InitDownloader(auth, nil)
	if err != nil {
		t.Errorf("init downloader failed, err: %v", err)
		return
	}

	vas := kit.NewVas()
	// test all download.
	if err = downloader[cc.BkRepo].Download(vas, smallFileAddr, 1021, smallFileToFile); err != nil {
		t.Errorf("download file failed, err: %v", err)
		return
	}

	// test range download.
	if err = downloader[cc.BkRepo].Download(vas, bigFileAddr, 93965, bigFileToFile); err != nil {
		t.Errorf("download file failed, err: %v", err)
		return
	}

	sha256, err := tools.FileSHA256(smallFileToFile)
	if err != nil {
		t.Errorf("file sha256 failed, err: %v", err)
		return
	}

	if sha256 != testSmallFileSha256 {
		t.Errorf("download file sha256 except %s, but %s", testSmallFileSha256, sha256)
		return
	}

	sha256, err = tools.FileSHA256(bigFileToFile)
	if err != nil {
		t.Errorf("file sha256 failed, err: %v", err)
		return
	}

	if sha256 != testBigFileSha256 {
		t.Errorf("download file sha256 except %s, but %s", testBigFileSha256, sha256)
		return
	}
}
