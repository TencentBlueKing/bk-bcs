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

package appinstance

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

// EffectedAction is appinstance effected list action object.
type EffectedAction struct {
	ctx   context.Context
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	req  *pb.QueryEffectedAppInstancesReq
	resp *pb.QueryEffectedAppInstancesResp

	sd *dbsharding.ShardingDB

	appInstanceReleases []database.AppInstanceRelease

	totalCount int64
	instances  []*pbcommon.AppInstanceRelease
}

// NewEffectedAction creates new EffectedAction.
func NewEffectedAction(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager,
	req *pb.QueryEffectedAppInstancesReq, resp *pb.QueryEffectedAppInstancesResp) *EffectedAction {

	action := &EffectedAction{ctx: ctx, viper: viper, smgr: smgr, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *EffectedAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *EffectedAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *EffectedAction) Output() error {
	act.resp.Data = &pb.QueryEffectedAppInstancesResp_RespData{TotalCount: uint32(act.totalCount), Info: act.instances}
	return nil
}

func (act *EffectedAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("cfg_id", act.req.CfgId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("release_id", act.req.ReleaseId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
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
		database.BSCPNOTEMPTY, database.BSCPQUERYLIMITMB); err != nil {
		return err
	}
	return nil
}

func (act *EffectedAction) queryEffectCount() (pbcommon.ErrCode, string) {
	if !act.req.Page.ReturnTotal {
		return pbcommon.ErrCode_E_OK, "OK"
	}

	err := act.sd.DB().
		Model(&database.AppInstanceRelease{}).
		Where(&database.AppInstanceRelease{
			BizID:     act.req.BizId,
			CfgID:     act.req.CfgId,
			ReleaseID: act.req.ReleaseId,
		}).
		Count(&act.totalCount).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *EffectedAction) queryAppInstanceReleases() (pbcommon.ErrCode, string) {
	err := act.sd.DB().
		Offset(int(act.req.Page.Start)).Limit(int(act.req.Page.Limit)).
		Order("Fcreate_time DESC, Fid DESC").
		Where(&database.AppInstanceRelease{
			BizID:     act.req.BizId,
			CfgID:     act.req.CfgId,
			ReleaseID: act.req.ReleaseId,
		}).
		Find(&act.appInstanceReleases).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *EffectedAction) queryAppInstance(instanceID uint64) (*database.AppInstance, pbcommon.ErrCode, string) {
	appInstance := database.AppInstance{}

	err := act.sd.DB().
		Where(&database.AppInstance{BizID: act.req.BizId, ID: instanceID}).
		Last(&appInstance).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return nil, pbcommon.ErrCode_E_DM_NOT_FOUND, "appinstance non-exist."
	}
	if err != nil {
		return nil, pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return &appInstance, pbcommon.ErrCode_E_OK, ""
}

func (act *EffectedAction) queryAppInstances() (pbcommon.ErrCode, string) {
	for _, st := range act.appInstanceReleases {
		instance, errCode, errMsg := act.queryAppInstance(st.InstanceID)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}

		ins := &pbcommon.AppInstanceRelease{
			InstanceId: instance.ID,
			BizId:      instance.BizID,
			AppId:      instance.AppID,
			CloudId:    instance.CloudID,
			Ip:         instance.IP,
			Path:       instance.Path,
			Labels:     instance.Labels,
			CfgId:      st.CfgID,
			ReleaseId:  st.ReleaseID,
			EffectCode: st.EffectCode,
			EffectMsg:  st.EffectMsg,
			ReloadCode: st.ReloadCode,
			ReloadMsg:  st.ReloadMsg,
			CreatedAt:  st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if st.EffectTime != nil {
			ins.EffectTime = st.EffectTime.Format("2006-01-02 15:04:05")
		}
		if st.ReloadTime != nil {
			ins.ReloadTime = st.ReloadTime.Format("2006-01-02 15:04:05")
		}
		act.instances = append(act.instances, ins)
	}

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *EffectedAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.BizId)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query appinstance release count.
	if errCode, errMsg := act.queryEffectCount(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query appinstance release list.
	if errCode, errMsg := act.queryAppInstanceReleases(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query appinstance list.
	if errCode, errMsg := act.queryAppInstances(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	return nil
}
