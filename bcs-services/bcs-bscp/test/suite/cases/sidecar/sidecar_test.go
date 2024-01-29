/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sidecar

import (
	"fmt"
	"io"
	"os/exec"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey" // import convey.

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/util"
)

func TestSidecar(t *testing.T) {
	SetDefaultFailureMode(FailureHalts)

	cli := suite.GetClient()
	gener := generator{
		cli: cli.ApiClient,
	}

	Convey("Prepare Job", t, func() {
		err := util.ClearDB(suite.DB)
		So(err, ShouldBeNil)
	})

	Convey("Sidecar Suite Test", t, func() {
		Convey("Init Data For Sidecar Test", func() {
			err := gener.InitData(kit.New())
			So(err, ShouldBeNil)
		})

		Convey("Start And Check Sidecar Workspace", func() {
			// start sidecar.
			err := startSidecar(suite.SidecarStartCmd)
			So(err, ShouldBeNil)

			sleepTime := 5 * time.Second
			retry := 0
			var hitErr error
			for {
				time.Sleep(sleepTime)
				if retry == 3 {
					So(fmt.Errorf("check sidecar release file timeout, err: %v", hitErr), ShouldBeNil)
				}

				hitErr = nil
				for appID, releaseMetas := range gener.data {
					rl := releaseMetas[len(releaseMetas)-1]
					if err = checkSidecarReleaseFile(testBizID, appID, rl.releaseID, rl.ciMeta); err != nil {
						hitErr = err
						break
					}
				}

				if hitErr != nil {
					retry++                   //nolint
					sleepTime = sleepTime * 4 //nolint
				}

				break
			}
			So(hitErr, ShouldBeNil)
		})

		Convey("Simulation App Publish Related Operation For Sidecar Test", func() {
			err := gener.SimulationData(kit.New())
			So(err, ShouldBeNil)
		})

		Convey("Check Sidecar Workspace After Simulation Publish", func() {
			sleepTime := 5 * time.Second
			retry := 0
			var hitErr error
			for {
				time.Sleep(sleepTime)
				if retry == 3 {
					So(fmt.Errorf("check sidecar release file timeout, err: %v", hitErr), ShouldBeNil)
				}

				hitErr = nil
				for appID, releaseMetas := range gener.data {
					rl := releaseMetas[len(releaseMetas)-1]
					if err := checkSidecarReleaseFile(testBizID, appID, rl.releaseID, rl.ciMeta); err != nil {
						hitErr = err
						break
					}
				}

				if hitErr != nil {
					retry++                   //nolint
					sleepTime = sleepTime * 4 //nolint
				}

				break
			}
			So(hitErr, ShouldBeNil)
		})
	})
}

// Exec used to perform command line operations.
// Params: cmdStr(exec's command string)
func startSidecar(cmdStr string) error {
	cmd := exec.Command("/bin/bash", "-c", cmdStr)

	// create cmd stdout pipe.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// exec cmd.
	if err = cmd.Start(); err != nil {
		return err
	}

	// read exec stdout.
	result, err := io.ReadAll(stdout)
	if err != nil {
		return err
	}

	if len(result) != 0 {
		return fmt.Errorf("sidecar start stdout not expect %s", result)
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
