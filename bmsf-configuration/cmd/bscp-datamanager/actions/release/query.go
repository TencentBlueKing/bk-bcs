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
	"errors"

	"github.com/bluele/gcache"
	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-datamanager/modules/metrics"
	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/logger"
)

// QueryAction is release query action object.
type QueryAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	collector    *metrics.Collector
	releaseCache gcache.Cache

	req  *pb.QueryReleaseReq
	resp *pb.QueryReleaseResp

	sd *dbsharding.ShardingDB
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(viper *viper.Viper, smgr *dbsharding.ShardingManager, collector *metrics.Collector, releaseCache gcache.Cache,
	req *pb.QueryReleaseReq, resp *pb.QueryReleaseResp) *QueryAction {
	action := &QueryAction{viper: viper, smgr: smgr, collector: collector, releaseCache: releaseCache, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *QueryAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryAction) Output() error {
	// do nothing.
	return nil
}

func (act *QueryAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Releaseid)
	if length == 0 {
		return errors.New("invalid params, releaseid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, releaseid too long")
	}
	return nil
}

func (act *QueryAction) queryRelease() (pbcommon.ErrCode, string) {
	// query release from cache.
	if cache, err := act.releaseCache.Get(act.req.Releaseid); err == nil && cache != nil {
		act.collector.StatReleaseCache(true)
		release := cache.(*pbcommon.Release)

		logger.V(3).Infof("QueryRelease[%d]| query release cache hit success[%s]", act.req.Seq, act.req.Releaseid)
		act.resp.Release = release
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatReleaseCache(false)

	// query release from db.
	act.sd.AutoMigrate(&database.Release{})

	var st database.Release
	err := act.sd.DB().
		Where(&database.Release{Bid: act.req.Bid, Releaseid: act.req.Releaseid}).
		Last(&st).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "release non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	release := &pbcommon.Release{
		ID:             st.ID,
		Bid:            st.Bid,
		Releaseid:      st.Releaseid,
		Name:           st.Name,
		Appid:          st.Appid,
		Cfgsetid:       st.Cfgsetid,
		CfgsetName:     st.CfgsetName,
		CfgsetFpath:    st.CfgsetFpath,
		MultiReleaseid: st.MultiReleaseid,
		Commitid:       st.Commitid,
		Strategyid:     st.Strategyid,
		Strategies:     st.Strategies,
		Creator:        st.Creator,
		Memo:           st.Memo,
		State:          st.State,
		LastModifyBy:   st.LastModifyBy,
		CreatedAt:      st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	act.resp.Release = release

	if release.State == int32(pbcommon.ReleaseState_RS_ROLLBACKED) ||
		release.State == int32(pbcommon.ReleaseState_RS_CANCELED) {
		if err := act.releaseCache.Set(act.req.Releaseid, release); err != nil {
			logger.Warn("QueryRelease[%d]| update local release cache, %+v", act.req.Seq, err)
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query release.
	if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
