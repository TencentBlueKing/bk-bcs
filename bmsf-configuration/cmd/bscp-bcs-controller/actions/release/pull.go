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

package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pb "bk-bscp/internal/protocol/bcs-controller"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/logger"
)

// PullAction pulls release back for sidecar.
type PullAction struct {
	viper           *viper.Viper
	dataMgrCli      pbdatamanager.DataManagerClient
	strategyHandler *strategy.Handler

	req  *pb.PullReleaseReq
	resp *pb.PullReleaseResp

	release *pbcommon.Release
}

// NewPullAction creates new PullAction.
func NewPullAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient, strategyHandler *strategy.Handler,
	req *pb.PullReleaseReq, resp *pb.PullReleaseResp) *PullAction {
	action := &PullAction{viper: viper, dataMgrCli: dataMgrCli, strategyHandler: strategyHandler, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PullAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PullAction) Input() error {
	if len(act.req.Labels) == 0 {
		act.req.Labels = strategy.EmptySidecarLabels
	}

	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_BCS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PullAction) Output() error {
	act.resp.Release = act.release
	return nil
}

func (act *PullAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Appid)
	if length == 0 {
		return errors.New("invalid params, appid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, appid too long")
	}

	length = len(act.req.Clusterid)
	if length == 0 {
		return errors.New("invalid params, clusterid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, clusterid too long")
	}

	length = len(act.req.Zoneid)
	if length == 0 {
		return errors.New("invalid params, zoneid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, zoneid too long")
	}

	length = len(act.req.Dc)
	if length == 0 {
		return errors.New("invalid params, dc missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, dc too long")
	}

	length = len(act.req.IP)
	if length == 0 {
		return errors.New("invalid params, ip missing")
	}
	if length > database.BSCPNORMALSTRLENLIMIT {
		return errors.New("invalid params, ip too long")
	}

	length = len(act.req.Cfgsetid)
	if length == 0 {
		return errors.New("invalid params, cfgsetid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}
	return nil
}

func (act *PullAction) targetRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.req.Releaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PullRelease[%d]| request to datamanager QueryRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	release := resp.Release
	if release.Bid != act.req.Bid {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, "can't pull target release, inconsistent bid"
	}
	if release.Appid != act.req.Appid {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, "can't pull target release, inconsistent appid"
	}
	if release.Cfgsetid != act.req.Cfgsetid {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, "can't pull target release, inconsistent cfgsetid"
	}
	if release.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, "can't pull target release, not published"
	}

	// do not check rollback release here, the rollback event is handled
	// in newest releases logic.

	if release.Strategies == strategy.EmptyStrategy {
		// empty strategy, all sidecars would be accepted.
		act.release = release
		return pbcommon.ErrCode_E_OK, ""
	}

	// match the release strategy.
	strategies := strategy.Strategy{}
	if err := json.Unmarshal([]byte(release.Strategies), &strategies); err != nil {
		return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, err.Error()
	}

	ins := &pbcommon.AppInstance{
		Appid:     act.req.Appid,
		Clusterid: act.req.Clusterid,
		Zoneid:    act.req.Zoneid,
		Dc:        act.req.Dc,
		Labels:    act.req.Labels,
		IP:        act.req.IP,
	}
	matcher := act.strategyHandler.Matcher()
	if !matcher(&strategies, ins) {
		return pbcommon.ErrCode_E_OK, ""
	}

	// matched.
	act.release = release
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PullAction) newestRelease() (pbcommon.ErrCode, string) {
	ins := &pbcommon.AppInstance{
		Appid:     act.req.Appid,
		Clusterid: act.req.Clusterid,
		Zoneid:    act.req.Zoneid,
		Dc:        act.req.Dc,
		Labels:    act.req.Labels,
		IP:        act.req.IP,
	}

	// query the releases in batch mode, and match the strategy.
	var newest *pbcommon.Release

	var index int32
	for {
		r := &pbdatamanager.QueryNewestReleasesReq{
			Seq:            act.req.Seq,
			Bid:            act.req.Bid,
			Cfgsetid:       act.req.Cfgsetid,
			LocalReleaseid: act.req.LocalReleaseid,
			Index:          index,
			Limit:          act.viper.GetInt32("server.queryNewestLimit"),
		}

		ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
		defer cancel()

		logger.V(2).Infof("request to datamanager QueryNewestReleases[%d], %+v", index, r)

		resp, err := act.dataMgrCli.QueryNewestReleases(ctx, r)
		if err != nil {
			return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryNewestReleases[%d], %+v", index, err)
		}
		if resp.ErrCode != pbcommon.ErrCode_E_OK {
			return resp.ErrCode, resp.ErrMsg
		}
		logger.V(2).Infof("request to datamanager QueryNewestReleases response[%d], %+v", index, resp)

		if len(resp.Releases) == 0 {
			logger.V(2).Infof("finally, no more release for this sidecar now[%d]", index)
			return pbcommon.ErrCode_E_OK, ""
		}

		for _, release := range resp.Releases {
			if release.Bid != act.req.Bid {
				return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, "can't pull newest release, inconsistent bid"
			}
			if release.Appid != act.req.Appid {
				return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, "can't pull newest release, inconsistent appid"
			}
			if release.Cfgsetid != act.req.Cfgsetid {
				return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, "can't pull newest release, inconsistent cfgsetid"
			}
			if release.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) {
				return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, "can't pull newest release, not published"
			}

			if release.Strategies == strategy.EmptyStrategy {
				newest = release
				break
			}

			strategies := strategy.Strategy{}
			if err := json.Unmarshal([]byte(release.Strategies), &strategies); err != nil {
				return pbcommon.ErrCode_E_BCS_SYSTEM_UNKONW, err.Error()
			}

			matcher := act.strategyHandler.Matcher()
			if !matcher(&strategies, ins) {
				continue
			}

			newest = release
			break
		}

		if newest != nil {
			break
		}

		if len(resp.Releases) < act.viper.GetInt("server.queryNewestLimit") {
			break
		}
		index += int32(len(resp.Releases))
	}

	if newest == nil {
		logger.V(2).Infof("finally query, no more release for this sidecar now")
		return pbcommon.ErrCode_E_OK, ""
	}

	// no more newest release, no rollbacked.
	if newest.Releaseid == act.req.LocalReleaseid && len(act.req.LocalReleaseid) != 0 {
		logger.V(2).Infof("local release just the newest releaseid[%+v]", newest.Releaseid)
		return pbcommon.ErrCode_E_OK, ""
	}
	act.release = newest

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PullAction) Do() error {
	if len(act.req.Releaseid) != 0 {
		if errCode, errMsg := act.targetRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	} else {
		if errCode, errMsg := act.newestRelease(); errCode != pbcommon.ErrCode_E_OK {
			return act.Err(errCode, errMsg)
		}
	}
	return nil
}
