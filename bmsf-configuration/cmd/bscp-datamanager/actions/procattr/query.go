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

package procattr

import (
	"errors"
	"path/filepath"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// QueryAction is procattr query action object.
type QueryAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryHostProcAttrReq
	resp *pb.QueryHostProcAttrResp

	sd *dbsharding.ShardingDB

	procAttr database.ProcAttr
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryHostProcAttrReq, resp *pb.QueryHostProcAttrResp) *QueryAction {
	action := &QueryAction{viper: viper, smgr: smgr, req: req, resp: resp}

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
	procAttr := &pbcommon.ProcAttr{
		Cloudid:      act.procAttr.Cloudid,
		IP:           act.procAttr.IP,
		Bid:          act.procAttr.Bid,
		Appid:        act.procAttr.Appid,
		Clusterid:    act.procAttr.Clusterid,
		Zoneid:       act.procAttr.Zoneid,
		Dc:           act.procAttr.Dc,
		Path:         act.procAttr.Path,
		Labels:       act.procAttr.Labels,
		Creator:      act.procAttr.Creator,
		Memo:         act.procAttr.Memo,
		State:        act.procAttr.State,
		LastModifyBy: act.procAttr.LastModifyBy,
		CreatedAt:    act.procAttr.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    act.procAttr.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	act.resp.ProcAttr = procAttr
	return nil
}

func (act *QueryAction) verify() error {
	length := len(act.req.Cloudid)
	if length == 0 {
		return errors.New("invalid params, cloudid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cloudid too long")
	}

	length = len(act.req.IP)
	if length == 0 {
		return errors.New("invalid params, ip missing")
	}
	if length > database.BSCPNORMALSTRLENLIMIT {
		return errors.New("invalid params, ip too long")
	}

	length = len(act.req.Bid)
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

	if len(act.req.Path) == 0 {
		return errors.New("invalid params, path missing")
	}
	act.req.Path = filepath.Clean(act.req.Path)
	if len(act.req.Path) > database.BSCPCFGSETFPATHLENLIMIT {
		return errors.New("invalid params, path too long")
	}

	return nil
}

func (act *QueryAction) queryProcAttr() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ProcAttr{})

	err := act.sd.DB().
		Where(&database.ProcAttr{Cloudid: act.req.Cloudid, IP: act.req.IP, Bid: act.req.Bid, Appid: act.req.Appid, Path: act.req.Path}).
		Where("Fstate = ?", pbcommon.ProcAttrState_PS_CREATED).
		Last(&act.procAttr).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "procattr non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// BSCP sharding db.
	sd, err := act.smgr.ShardingDB(dbsharding.BSCPDBKEY)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query procattr.
	if errCode, errMsg := act.queryProcAttr(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
