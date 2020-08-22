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
	switch r.(type) {
	case *pbsidecar.PingReq:
	case *pbsidecar.InjectReq:
		req := r.(*pbsidecar.InjectReq)

		length := len(req.BusinessName)
		if length == 0 {
			return errors.New("invalid params, businessName missing")
		}
		if length > database.BSCPNAMELENLIMIT {
			return errors.New("invalid params, businessName too long")
		}

		length = len(req.AppName)
		if length == 0 {
			return errors.New("invalid params, appName missing")
		}
		if length > database.BSCPNAMELENLIMIT {
			return errors.New("invalid params, appName too long")
		}

		length = len(req.Path)
		if length == 0 {
			return errors.New("invalid params, path missing")
		}
		if length > database.BSCPCFGSETFPATHLENLIMIT {
			return errors.New("invalid params, path too long")
		}

		if len(req.Labels) == 0 {
			return errors.New("invalid params, labels missing")
		}
	case *pbsidecar.WatchReloadReq:
		req := r.(*pbsidecar.WatchReloadReq)

		length := len(req.BusinessName)
		if length == 0 {
			return errors.New("invalid params, businessName missing")
		}
		if length > database.BSCPNAMELENLIMIT {
			return errors.New("invalid params, businessName too long")
		}

		length = len(req.AppName)
		if length == 0 {
			return errors.New("invalid params, appName missing")
		}
		if length > database.BSCPNAMELENLIMIT {
			return errors.New("invalid params, appName too long")
		}

		length = len(req.Path)
		if length == 0 {
			return errors.New("invalid params, path missing")
		}
		if length > database.BSCPCFGSETFPATHLENLIMIT {
			return errors.New("invalid params, path too long")
		}

	case *pbsidecar.ReportReloadReq:
		req := r.(*pbsidecar.ReportReloadReq)

		length := len(req.BusinessName)
		if length == 0 {
			return errors.New("invalid params, businessName missing")
		}
		if length > database.BSCPNAMELENLIMIT {
			return errors.New("invalid params, businessName too long")
		}

		length = len(req.AppName)
		if length == 0 {
			return errors.New("invalid params, appName missing")
		}
		if length > database.BSCPNAMELENLIMIT {
			return errors.New("invalid params, appName too long")
		}

		length = len(req.Path)
		if length == 0 {
			return errors.New("invalid params, path missing")
		}
		if length > database.BSCPCFGSETFPATHLENLIMIT {
			return errors.New("invalid params, path too long")
		}

		if len(req.Releaseid) > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, releaseid too long")
		}
		if len(req.MultiReleaseid) > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, multiReleaseid too long")
		}
		if len(req.Releaseid) == 0 && len(req.MultiReleaseid) == 0 {
			return errors.New("invalid params, releaseid and multiReleaseid missing")
		}

		length = len(req.ReloadTime)
		if length == 0 {
			return errors.New("invalid params, reloadTime missing")
		}
		if length > database.BSCPNORMALSTRLENLIMIT {
			return errors.New("invalid params, reloadTime too long, eg: 2006-01-02 15:04:05")
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
