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
 *
 */

package upgrader

import (
	"errors"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"

	"github.com/emicklei/go-restful"
)

// UpgradeResponse is the response of upgrade
type UpgradeResponse struct {
	Msg              string   `json:"msg"`
	PreVersion       string   `json:"pre_version"`
	CurrentVersion   string   `json:"current_version"`
	FinishedVersions []string `json:"finished_migrations"`
}

// Upgrade is to upgrade version
func (u *Upgrader) Upgrade(req *restful.Request, resp *restful.Response) {

	preVersion, finishedVersions, err := RunUpgrade(u.ctx, u.upgradeHelper)
	if err != nil {
		blog.Errorf("db upgrade failed, err: %+v", err)
		errResp := bhttp.InternalError(common.BcsErrApiInternalFail, err.Error())
		resp.Write([]byte(errResp.Error()))
		return
	}

	currentVersion := preVersion
	if len(finishedVersions) > 0 {
		currentVersion = finishedVersions[len(finishedVersions)-1]
	}

	data := UpgradeResponse{
		Msg:              "upgrade success",
		PreVersion:       preVersion,
		CurrentVersion:   currentVersion,
		FinishedVersions: finishedVersions,
	}
	result := errors.New(bhttp.GetRespone(common.BcsSuccess, common.BcsSuccessStr, data))

	resp.Write([]byte(result.Error()))
}
