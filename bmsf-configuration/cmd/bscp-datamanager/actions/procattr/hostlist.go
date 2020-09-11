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

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
)

// HostListAction query app procattr list on host.
type HostListAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryHostProcAttrListReq
	resp *pb.QueryHostProcAttrListResp

	sd *dbsharding.ShardingDB

	procAttrs []database.ProcAttr
}

// NewHostListAction creates new HostListAction.
func NewHostListAction(viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryHostProcAttrListReq, resp *pb.QueryHostProcAttrListResp) *HostListAction {
	action := &HostListAction{viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *HostListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *HostListAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *HostListAction) Output() error {
	procAttrs := []*pbcommon.ProcAttr{}
	for _, st := range act.procAttrs {
		procAttr := &pbcommon.ProcAttr{
			Cloudid:      st.Cloudid,
			IP:           st.IP,
			Bid:          st.Bid,
			Appid:        st.Appid,
			Clusterid:    st.Clusterid,
			Zoneid:       st.Zoneid,
			Dc:           st.Dc,
			Path:         st.Path,
			Labels:       st.Labels,
			Creator:      st.Creator,
			Memo:         st.Memo,
			State:        st.State,
			LastModifyBy: st.LastModifyBy,
			CreatedAt:    st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		procAttrs = append(procAttrs, procAttr)
	}
	act.resp.ProcAttrs = procAttrs
	return nil
}

func (act *HostListAction) verify() error {
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

	if act.req.Limit == 0 {
		return errors.New("invalid params, limit missing")
	}
	if act.req.Limit > database.BSCPQUERYLIMIT {
		return errors.New("invalid params, limit too long")
	}
	return nil
}

func (act *HostListAction) queryProcAttrList() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.ProcAttr{})

	err := act.sd.DB().
		Offset(act.req.Index).Limit(act.req.Limit).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.ProcAttr{Cloudid: act.req.Cloudid, IP: act.req.IP}).
		Where("Fstate = ?", pbcommon.ProcAttrState_PS_CREATED).
		Find(&act.procAttrs).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *HostListAction) Do() error {
	// BSCP sharding db.
	sd, err := act.smgr.ShardingDB(dbsharding.BSCPDBKEY)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query procattr list.
	if errCode, errMsg := act.queryProcAttrList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
