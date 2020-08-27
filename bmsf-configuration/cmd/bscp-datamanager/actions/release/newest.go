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

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// NewestAction is newest release query action object.
type NewestAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryNewestReleasesReq
	resp *pb.QueryNewestReleasesResp

	sd *dbsharding.ShardingDB
}

// NewNewestAction creates new NewestAction.
func NewNewestAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryNewestReleasesReq, resp *pb.QueryNewestReleasesResp) *NewestAction {
	action := &NewestAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *NewestAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *NewestAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *NewestAction) Output() error {
	// do nothing.
	return nil
}

func (act *NewestAction) verify() error {
	length := len(act.req.Bid)
	if length == 0 {
		return errors.New("invalid params, bid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, bid too long")
	}

	length = len(act.req.Cfgsetid)
	if length == 0 {
		return errors.New("invalid params, cfgsetid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too big")
	}
	return nil
}

func (act *NewestAction) queryNewestReleases() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Release{})

	// query newest releases.
	var sts []database.Release

	// release tree must be sequential according to the ID order, not the update time.
	err := act.sd.DB().
		Offset(act.req.Index).Limit(act.req.Limit).
		Order("Fid DESC").
		Where(&database.Release{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, State: int32(pbcommon.ReleaseState_RS_PUBLISHED)}).
		Find(&sts).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	var releases []*pbcommon.Release
	for _, st := range sts {
		release := &pbcommon.Release{
			ID:             st.ID,
			Bid:            st.Bid,
			Releaseid:      st.Releaseid,
			Name:           st.Name,
			Appid:          st.Appid,
			Cfgsetid:       st.Cfgsetid,
			CfgsetName:     st.CfgsetName,
			CfgsetFpath:    st.CfgsetFpath,
			Commitid:       st.Commitid,
			MultiReleaseid: st.MultiReleaseid,
			Strategyid:     st.Strategyid,
			Strategies:     st.Strategies,
			Creator:        st.Creator,
			Memo:           st.Memo,
			State:          st.State,
			LastModifyBy:   st.LastModifyBy,
			CreatedAt:      st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		releases = append(releases, release)
	}
	act.resp.Releases = releases

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *NewestAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query newest releases.
	if errCode, errMsg := act.queryNewestReleases(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
