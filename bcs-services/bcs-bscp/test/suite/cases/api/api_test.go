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

package api

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey" // import convey.

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/suite/cases"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/test/util"
)

// ResourceManager instance
var rm *cases.ResourceManager

func TestApi(t *testing.T) {
	SetDefaultFailureMode(FailureHalts)

	Convey("Prepare Job", t, func() {
		rm = cases.NewResourceManager()

		err := util.ClearDB(suite.DB)
		So(err, ShouldBeNil)
	})

	TestApplication(t)
	TestHook(t)
	TestConfigItem(t)
	TestContent(t)
	TestCommit(t)
	TestRelease(t)
	TestStrategySet(t)
	// NOTE: strategy related test depends on group, add group test first
	//TestStrategy(t)
	//TestPublish(t)
	TestInstance(t)
}
