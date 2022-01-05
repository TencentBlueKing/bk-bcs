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
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pb "bk-bscp/internal/protocol/gse-controller"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// PullAction pulls release back for app instance.
type PullAction struct {
	ctx             context.Context
	viper           *viper.Viper
	dataMgrCli      pbdatamanager.DataManagerClient
	strategyHandler *strategy.Handler

	req  *pb.PullReleaseReq
	resp *pb.PullReleaseResp

	release *pbcommon.Release
}

// NewPullAction creates new PullAction.
func NewPullAction(ctx context.Context, viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	strategyHandler *strategy.Handler,
	req *pb.PullReleaseReq, resp *pb.PullReleaseResp) *PullAction {

	action := &PullAction{ctx: ctx, viper: viper, dataMgrCli: dataMgrCli,
		strategyHandler: strategyHandler, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PullAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PullAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_GSE_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PullAction) Output() error {
	act.resp.Release = act.release
	return nil
}

func (act *PullAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cloud_id", act.req.CloudId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("ip", act.req.Ip,
		database.BSCPNOTEMPTY, database.BSCPNORMALSTRLENLIMIT); err != nil {
		return err
	}
	act.req.Path = filepath.Clean(act.req.Path)
	if err = common.ValidateString("path", act.req.Path,
		database.BSCPNOTEMPTY, database.BSCPCFGFPATHLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cfg_id", act.req.CfgId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *PullAction) targetRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.req.Seq,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("PullRelease[%s]| request to datamanager, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}

	release := resp.Data
	if release.BizId != act.req.BizId {
		return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, "can't pull target release, inconsistent bizid"
	}
	if release.AppId != act.req.AppId {
		return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, "can't pull target release, inconsistent appid"
	}
	if release.CfgId != act.req.CfgId {
		return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, "can't pull target release, inconsistent cfgid"
	}
	if release.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) {
		return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, "can't pull target release, not published"
	}

	// do not check rollback release here, the rollback event is handled
	// in newest releases logic.

	if release.Strategies == strategy.EmptyStrategy {
		// empty strategy, all app instance would be accepted.
		act.release = release
		return pbcommon.ErrCode_E_OK, ""
	}

	// match the release strategy.
	strategies := strategy.Strategy{}
	if err := json.Unmarshal([]byte(release.Strategies), &strategies); err != nil {
		return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, err.Error()
	}

	ins := &pbcommon.AppInstance{
		AppId:   act.req.AppId,
		CloudId: act.req.CloudId,
		Ip:      act.req.Ip,
		Path:    act.req.Path,
		Labels:  act.req.Labels,
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
		AppId:   act.req.AppId,
		CloudId: act.req.CloudId,
		Ip:      act.req.Ip,
		Path:    act.req.Path,
		Labels:  act.req.Labels,
	}

	// query the releases in batch mode, and match the strategy.
	var newest *pbcommon.Release

	var index int32
	limit := act.viper.GetInt("server.queryNewestLimit")

	for {
		r := &pbdatamanager.QueryNewestReleasesReq{
			Seq:            act.req.Seq,
			BizId:          act.req.BizId,
			CfgId:          act.req.CfgId,
			LocalReleaseId: act.req.LocalReleaseId,
			Page:           &pbcommon.Page{Start: index, Limit: int32(limit)},
		}

		ctx, cancel := context.WithTimeout(act.ctx, act.viper.GetDuration("datamanager.callTimeout"))
		defer cancel()

		logger.V(4).Infof("PullRelease[%s]| request to datamanager[%d], %+v", act.req.Seq, index, r)

		resp, err := act.dataMgrCli.QueryNewestReleases(ctx, r)
		if err != nil {
			return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN,
				fmt.Sprintf("request to datamanager QueryNewestReleases[%d], %+v", index, err)
		}
		if resp.Code != pbcommon.ErrCode_E_OK {
			return resp.Code, resp.Message
		}
		logger.V(4).Infof("PullRelease[%s]| request to datamanager response[%d], %+v", act.req.Seq, index, resp)

		if len(resp.Data.Info) == 0 {
			logger.V(2).Infof("PullRelease[%s]| finally, no release for this app instance now[%d]", act.req.Seq, index)
			return pbcommon.ErrCode_E_OK, ""
		}

		startTime := time.Now()

		// index of matched release, 0 means that
		// can't find the matched release in this round.
		matchedIdx := 0

		for idx, release := range resp.Data.Info {
			if release.BizId != act.req.BizId {
				return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, "can't pull newest release, inconsistent bizid"
			}
			if release.AppId != act.req.AppId {
				return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, "can't pull newest release, inconsistent appid"
			}
			if release.CfgId != act.req.CfgId {
				return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, "can't pull newest release, inconsistent cfgid"
			}
			if release.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) {
				return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, "can't pull newest release, not published"
			}

			if release.Strategies == strategy.EmptyStrategy {
				newest = release
				matchedIdx = idx
				break
			}

			strategies := strategy.Strategy{}
			if err := json.Unmarshal([]byte(release.Strategies), &strategies); err != nil {
				return pbcommon.ErrCode_E_GSE_SYSTEM_UNKNOWN, err.Error()
			}

			matcher := act.strategyHandler.Matcher()
			if !matcher(&strategies, ins) {
				continue
			}

			newest = release
			matchedIdx = idx
			break
		}

		logger.V(2).Infof("PullRelease[%s]| match releases in round[%d], count[%d] matchedIdx[%d], cost: %+v",
			act.req.Seq, index, len(resp.Data.Info), matchedIdx, time.Since(startTime))

		if newest != nil {
			break
		}

		if len(resp.Data.Info) < limit {
			break
		}
		index += int32(len(resp.Data.Info))
	}

	if newest == nil {
		logger.V(2).Infof("PullRelease[%s]| finally, no more release for this app instance", act.req.Seq)
		return pbcommon.ErrCode_E_OK, ""
	}

	// no more newest release, no rollbacked.
	if newest.ReleaseId == act.req.LocalReleaseId && len(act.req.LocalReleaseId) != 0 {
		logger.V(2).Infof("PullRelease[%s]| local release just the newest release[%+v]", act.req.Seq, newest.ReleaseId)
		return pbcommon.ErrCode_E_OK, ""
	}
	act.release = newest

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PullAction) Do() error {
	if len(act.req.ReleaseId) != 0 {
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
