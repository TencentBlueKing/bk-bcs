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
	"context"
	"errors"
	"math"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
)

// HostListAction query app procattr list on host.
type HostListAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryHostProcAttrListReq
	resp *pb.QueryHostProcAttrListResp

	sd *dbsharding.ShardingDB

	totalCount int64
	procAttrs  []database.ProcAttr
}

// NewHostListAction creates new HostListAction.
func NewHostListAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryHostProcAttrListReq, resp *pb.QueryHostProcAttrListResp) *HostListAction {
	action := &HostListAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *HostListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
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
			CloudId:      st.CloudID,
			Ip:           st.IP,
			BizId:        st.BizID,
			AppId:        st.AppID,
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
	act.resp.Data = &pb.QueryHostProcAttrListResp_RespData{TotalCount: uint32(act.totalCount), Info: procAttrs}
	return nil
}

func (act *HostListAction) verify() error {
	var err error

	if err = common.ValidateString("cloud_id", act.req.CloudId,
		database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("ip", act.req.Ip,
		database.BSCPNOTEMPTY, database.BSCPNORMALSTRLENLIMIT); err != nil {
		return err
	}

	if act.req.Page == nil {
		return errors.New("invalid input data, page is required")
	}
	if err = common.ValidateInt32("page.start", act.req.Page.Start,
		database.BSCPEMPTY, math.MaxInt32); err != nil {
		return err
	}
	if err = common.ValidateInt32("page.limit", act.req.Page.Limit,
		database.BSCPNOTEMPTY, database.BSCPQUERYLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *HostListAction) queryProcAttrCount() (pbcommon.ErrCode, string) {
	if !act.req.Page.ReturnTotal {
		return pbcommon.ErrCode_E_OK, ""
	}

	err := act.sd.DB().
		Model(&database.ProcAttr{}).
		Where(&database.ProcAttr{CloudID: act.req.CloudId, IP: act.req.Ip}).
		Count(&act.totalCount).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *HostListAction) queryProcAttrList() (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Offset(int(act.req.Page.Start)).Limit(int(act.req.Page.Limit)).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.ProcAttr{CloudID: act.req.CloudId, IP: act.req.Ip}).
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

	// query procattr count.
	if errCode, errMsg := act.queryProcAttrCount(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query procattr list.
	if errCode, errMsg := act.queryProcAttrList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
