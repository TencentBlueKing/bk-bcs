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

package configs

import (
	"context"
	"errors"
	"fmt"

	"github.com/bluele/gcache"
	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-connserver/modules/metrics"
	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/logger"
)

// PullAction pull release configs.
type PullAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	configsCache gcache.Cache
	collector    *metrics.Collector

	req  *pb.PullReleaseConfigsReq
	resp *pb.PullReleaseConfigsResp

	content []byte
	release *pbcommon.Release
}

// NewPullAction creates new PullAction.
func NewPullAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient, configsCache gcache.Cache, collector *metrics.Collector,
	req *pb.PullReleaseConfigsReq, resp *pb.PullReleaseConfigsResp) *PullAction {
	action := &PullAction{viper: viper, dataMgrCli: dataMgrCli, configsCache: configsCache, collector: collector, req: req, resp: resp}

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
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CONNS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PullAction) Output() error {
	// do nothing.
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

	length = len(act.req.Cfgsetid)
	if length == 0 {
		return errors.New("invalid params, cfgsetid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	length = len(act.req.Releaseid)
	if length == 0 {
		return errors.New("invalid params, releaseid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, releaseid too long")
	}

	length = len(act.req.Cid)
	if length == 0 {
		return errors.New("invalid params, cid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cid too long")
	}
	return nil
}

func (act *PullAction) queryRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Releaseid: act.req.Releaseid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PullReleaseConfigs[%d]| request to datamanager QueryRelease, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CONNS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	act.release = resp.Release

	return resp.ErrCode, resp.ErrMsg
}

func (act *PullAction) queryConfigs() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseConfigsReq{
		Seq:       act.req.Seq,
		Bid:       act.req.Bid,
		Appid:     act.req.Appid,
		Clusterid: act.req.Clusterid,
		Zoneid:    act.req.Zoneid,
		Cfgsetid:  act.req.Cfgsetid,
		Commitid:  act.release.Commitid,
		Abstract:  (len(act.content) != 0),
		Index:     act.req.IP,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("PullReleaseConfigs[%d]| request to datamanager QueryReleaseConfigs, %+v", act.req.Seq, r)

	resp, err := act.dataMgrCli.QueryReleaseConfigs(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CONNS_SYSTEM_UNKONW, fmt.Sprintf("request to datamanager QueryReleaseConfigs, %+v", err)
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return resp.ErrCode, resp.ErrMsg
	}

	if len(act.content) != 0 {
		resp.Configs.Content = act.content
	} else {
		// update local release cache.
		if err := act.configsCache.Set(act.req.Cid, resp.Configs.Content); err != nil {
			logger.Warn("update local configs contet cache, %+v", err)
		}
	}

	act.resp.Cid = resp.Configs.Cid
	act.resp.CfgLink = resp.Configs.CfgLink
	act.resp.Content = resp.Configs.Content

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *PullAction) Do() error {
	// query configs content cache.
	if cache, err := act.configsCache.Get(act.req.Cid); err == nil && cache != nil {
		act.content = cache.([]byte)

		logger.V(3).Infof("PullReleaseConfigs[%d]| query configs cache(size:%d) hit success[%s]",
			act.req.Seq, len(act.content), act.req.Cid)

		act.collector.StatConfigsCache(true)
	} else {
		act.collector.StatConfigsCache(false)
	}

	// query target release.
	if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	if act.release.State != int32(pbcommon.ReleaseState_RS_PUBLISHED) {
		return act.Err(pbcommon.ErrCode_E_CONNS_SYSTEM_UNKONW, "target release is not published")
	}

	// query release configs.
	if errCode, errMsg := act.queryConfigs(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
