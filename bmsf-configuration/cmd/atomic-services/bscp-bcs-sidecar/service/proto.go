/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"errors"
	"fmt"

	"bk-bscp/internal/database"
	pbsidecar "bk-bscp/internal/protocol/sidecar"
)

func verifyProto(r interface{}) error {
	// #lizard forgives
	switch r.(type) {
	case *pbsidecar.PingReq:
	case *pbsidecar.InjectReq:
		req := r.(*pbsidecar.InjectReq)

		length := len(req.BizId)
		if length == 0 {
			return errors.New("invalid params, biz_id missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, biz_id too long")
		}

		length = len(req.AppId)
		if length == 0 {
			return errors.New("invalid params, app_id missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, app_id too long")
		}

		length = len(req.Path)
		if length == 0 {
			return errors.New("invalid params, path missing")
		}
		if length > database.BSCPCFGFPATHLENLIMIT {
			return errors.New("invalid params, path too long")
		}

		if len(req.Labels) == 0 {
			return errors.New("invalid params, labels missing")
		}
	case *pbsidecar.WatchReloadReq:
		req := r.(*pbsidecar.WatchReloadReq)

		length := len(req.BizId)
		if length == 0 {
			return errors.New("invalid params, biz_id missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, biz_id too long")
		}

		length = len(req.AppId)
		if length == 0 {
			return errors.New("invalid params, app_id missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, app_id too long")
		}

		length = len(req.Path)
		if length == 0 {
			return errors.New("invalid params, path missing")
		}
		if length > database.BSCPCFGFPATHLENLIMIT {
			return errors.New("invalid params, path too long")
		}

	case *pbsidecar.ReportReloadReq:
		req := r.(*pbsidecar.ReportReloadReq)

		length := len(req.BizId)
		if length == 0 {
			return errors.New("invalid params, biz_id missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, biz_id too long")
		}

		length = len(req.AppId)
		if length == 0 {
			return errors.New("invalid params, app_id missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, app_id too long")
		}

		length = len(req.Path)
		if length == 0 {
			return errors.New("invalid params, path missing")
		}
		if length > database.BSCPCFGFPATHLENLIMIT {
			return errors.New("invalid params, path too long")
		}

		if len(req.ReleaseId) > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, release_id too long")
		}
		if len(req.MultiReleaseId) > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, multi_releasae_id too long")
		}
		if len(req.ReleaseId) == 0 && len(req.MultiReleaseId) == 0 {
			return errors.New("invalid params, release_id and multi_release_id missing")
		}

		length = len(req.ReloadTime)
		if length == 0 {
			return errors.New("invalid params, reload_time missing")
		}
		if length > database.BSCPNORMALSTRLENLIMIT {
			return errors.New("invalid params, reload_time too long, eg: 2006-01-02 15:04:05")
		}

		if req.ReloadCode == 0 {
			return errors.New("invalid params, reloadCode missing")
		}

		if len(req.ReloadMsg) > database.BSCPEFFECTRELOADERRLENLIMIT {
			return errors.New("invalid params, reloadMsg too long")
		}

	default:
		return fmt.Errorf("invalid request type[%+v]", r)
	}

	return nil
}
