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
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-datamanager/modules/metrics"
	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/logger"
)

// CreateReleaseAction is appinstance release create action object.
type CreateReleaseAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	collector *metrics.Collector

	req  *pb.CreateAppInstanceReleaseReq
	resp *pb.CreateAppInstanceReleaseResp

	sd *dbsharding.ShardingDB

	appInstance database.AppInstance
}

// NewCreateReleaseAction creates new CreateReleaseAction.
func NewCreateReleaseAction(viper *viper.Viper, smgr *dbsharding.ShardingManager, collector *metrics.Collector,
	req *pb.CreateAppInstanceReleaseReq, resp *pb.CreateAppInstanceReleaseResp) *CreateReleaseAction {
	action := &CreateReleaseAction{viper: viper, smgr: smgr, collector: collector, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateReleaseAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateReleaseAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateReleaseAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateReleaseAction) verify() error {
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

	if len(act.req.Labels) == 0 {
		act.req.Labels = strategy.EmptySidecarLabels
	}
	if len(act.req.Labels) > database.BSCPLABELSSIZELIMIT {
		return errors.New("invalid params, labels too long")
	}

	if len(act.req.Infos) == 0 {
		return errors.New("invalid params, infos missing")
	}
	return nil
}

func (act *CreateReleaseAction) queryAppInstance() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.AppInstance{})

	err := act.sd.DB().
		Where(&database.AppInstance{Bid: act.req.Bid, Appid: act.req.Appid, Clusterid: act.req.Clusterid,
			Zoneid: act.req.Zoneid, Dc: act.req.Dc, IP: act.req.IP}).
		Last(&act.appInstance).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "appinstance non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateReleaseAction) createAppInstanceReleaseEffect(info *pbcommon.ReportInfo) (pbcommon.ErrCode, string) {
	effectTime, err := time.ParseInLocation("2006-01-02 15:04:05", info.EffectTime, time.Local)
	if err != nil {
		logger.Warn("CreateAppInstanceRelease[%d]| invalid EffectTime format, %+v", act.req.Seq, err)
		act.collector.StatAppInstanceRelease(false)
		return pbcommon.ErrCode_E_DM_SYSTEM_UNKONW, "invalid EffectTime format"
	}

	appInstanceRelease := database.AppInstanceRelease{
		Instanceid: act.appInstance.ID,
		Bid:        act.req.Bid,
		Appid:      act.req.Appid,
		Clusterid:  act.req.Clusterid,
		Zoneid:     act.req.Zoneid,
		Dc:         act.req.Dc,
		Labels:     act.req.Labels,
		IP:         act.req.IP,
		Cfgsetid:   info.Cfgsetid,
		Releaseid:  info.Releaseid,
		EffectTime: &effectTime,
		EffectCode: info.EffectCode,
		EffectMsg:  info.EffectMsg,
	}
	if len(info.EffectMsg) > database.BSCPEFFECTRELOADERRLENLIMIT {
		// maybe a large error message.
		appInstanceRelease.EffectMsg = info.EffectMsg[:database.BSCPEFFECTRELOADERRLENLIMIT]
	}

	err = act.sd.DB().
		Where(database.AppInstanceRelease{Instanceid: act.appInstance.ID, Cfgsetid: info.Cfgsetid, Releaseid: info.Releaseid}).
		Assign(appInstanceRelease).
		FirstOrCreate(&appInstanceRelease).Error

	if err != nil {
		e, ok := err.(*mysql.MySQLError)
		if !ok {
			logger.Warn("CreateAppInstanceRelease[%d]| create app instance release effect record, %+v", act.req.Seq, err)
			act.collector.StatAppInstanceRelease(false)
			return pbcommon.ErrCode_E_DM_SYSTEM_UNKONW, err.Error()
		}

		if e.Number != 1062 {
			logger.Warn("CreateAppInstanceRelease[%d]| create app instance release effect record, %+v", act.req.Seq, err)
			act.collector.StatAppInstanceRelease(false)
			return pbcommon.ErrCode_E_DM_SYSTEM_UNKONW, err.Error()
		}

		tryErr := act.sd.DB().
			Where(database.AppInstanceRelease{Instanceid: act.appInstance.ID, Cfgsetid: info.Cfgsetid, Releaseid: info.Releaseid}).
			Assign(appInstanceRelease).
			FirstOrCreate(&appInstanceRelease).Error

		if tryErr != nil {
			logger.Warn("CreateAppInstanceRelease[%d]| create app instance release effect record, try again, %+v", act.req.Seq, tryErr)
			act.collector.StatAppInstanceRelease(false)
			return pbcommon.ErrCode_E_DM_SYSTEM_UNKONW, err.Error()
		}
	}

	logger.V(2).Infof("CreateAppInstanceRelease[%d]| create app instance release effect record success, %+v",
		act.req.Seq, appInstanceRelease)
	act.collector.StatAppInstanceRelease(true)

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateReleaseAction) queryCfgsetidByReleae(releaseid string) (string, error) {
	act.sd.AutoMigrate(&database.Release{})

	var st database.Release
	err := act.sd.DB().
		Where(&database.Release{Bid: act.req.Bid, Releaseid: releaseid}).
		Last(&st).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return "", errors.New("release non-exist.")
	}
	if err != nil {
		return "", err
	}
	return st.Cfgsetid, nil
}

func (act *CreateReleaseAction) queryIDsByMultiReleae(multiReleaseid string) ([]string, []string, error) {
	act.sd.AutoMigrate(&database.Release{})

	var sts []database.Release
	err := act.sd.DB().
		Where(&database.Release{Bid: act.req.Bid, MultiReleaseid: multiReleaseid}).
		Find(&sts).Error

	if err != nil {
		return nil, nil, err
	}

	cfgsetids := []string{}
	releaseids := []string{}

	for _, st := range sts {
		cfgsetids = append(cfgsetids, st.Cfgsetid)
		releaseids = append(releaseids, st.Releaseid)
	}

	if len(cfgsetids) != len(releaseids) {
		return nil, nil, errors.New("invalid cfgsetids and releaseids num in multi release reload report")
	}
	return cfgsetids, releaseids, nil
}

func (act *CreateReleaseAction) createAppInstanceReleaseReload(info *pbcommon.ReportInfo) (pbcommon.ErrCode, string) {
	reloadTime, err := time.ParseInLocation("2006-01-02 15:04:05", info.ReloadTime, time.Local)
	if err != nil {
		logger.Warn("CreateAppInstanceRelease[%d]| invalid ReloadTime format, %+v", act.req.Seq, err)
		act.collector.StatAppInstanceRelease(false)
		return pbcommon.ErrCode_E_DM_SYSTEM_UNKONW, "invalid ReloadTime format"
	}

	finalCfgsetids := []string{}
	finalReleaseids := []string{}

	if len(info.Releaseid) != 0 {
		cfgsetid, err := act.queryCfgsetidByReleae(info.Releaseid)
		if err != nil {
			act.collector.StatAppInstanceRelease(false)
			return pbcommon.ErrCode_E_DM_SYSTEM_UNKONW, err.Error()
		}

		finalCfgsetids = append(finalCfgsetids, cfgsetid)
		finalReleaseids = append(finalReleaseids, info.Releaseid)
	} else if len(info.MultiReleaseid) != 0 {
		cfgsetids, releaseids, err := act.queryIDsByMultiReleae(info.MultiReleaseid)
		if err != nil {
			act.collector.StatAppInstanceRelease(false)
			return pbcommon.ErrCode_E_DM_SYSTEM_UNKONW, err.Error()
		}

		finalCfgsetids = append(finalCfgsetids, cfgsetids...)
		finalReleaseids = append(finalReleaseids, releaseids...)
	} else {
		act.collector.StatAppInstanceRelease(false)
		return pbcommon.ErrCode_E_DM_SYSTEM_UNKONW, "invalid report type, releaseid and multiReleaseid all missing"
	}

	for i, cfgsetid := range finalCfgsetids {
		appInstanceRelease := database.AppInstanceRelease{
			Instanceid: act.appInstance.ID,
			Bid:        act.req.Bid,
			Appid:      act.req.Appid,
			Clusterid:  act.req.Clusterid,
			Zoneid:     act.req.Zoneid,
			Dc:         act.req.Dc,
			Labels:     act.req.Labels,
			IP:         act.req.IP,
			Cfgsetid:   cfgsetid,
			Releaseid:  finalReleaseids[i],
			ReloadTime: &reloadTime,
			ReloadCode: info.ReloadCode,
			ReloadMsg:  info.ReloadMsg,
		}
		if len(info.ReloadMsg) > database.BSCPEFFECTRELOADERRLENLIMIT {
			// maybe a large error message.
			appInstanceRelease.ReloadMsg = info.ReloadMsg[:database.BSCPEFFECTRELOADERRLENLIMIT]
		}

		err := act.sd.DB().
			Where(database.AppInstanceRelease{Instanceid: act.appInstance.ID, Cfgsetid: cfgsetid, Releaseid: finalReleaseids[i]}).
			Assign(appInstanceRelease).
			FirstOrCreate(&appInstanceRelease).Error

		if err != nil {
			e, ok := err.(*mysql.MySQLError)
			if !ok {
				logger.Warn("CreateAppInstanceRelease[%d]| create app instance release reload record, %+v", act.req.Seq, err)
				act.collector.StatAppInstanceRelease(false)
				continue
			}

			if e.Number != 1062 {
				logger.Warn("CreateAppInstanceRelease[%d]| create app instance release reload record, %+v", act.req.Seq, err)
				act.collector.StatAppInstanceRelease(false)
				continue
			}

			tryErr := act.sd.DB().
				Where(database.AppInstanceRelease{Instanceid: act.appInstance.ID, Cfgsetid: cfgsetid, Releaseid: finalReleaseids[i]}).
				Assign(appInstanceRelease).
				FirstOrCreate(&appInstanceRelease).Error

			if tryErr != nil {
				logger.Warn("CreateAppInstanceRelease[%d]| create app instance release reload record, try again, %+v", act.req.Seq, tryErr)
				act.collector.StatAppInstanceRelease(false)
				continue
			}
		}

		logger.V(2).Infof("CreateAppInstanceRelease[%d]| create app instance release reload record success, %+v",
			act.req.Seq, appInstanceRelease)
		act.collector.StatAppInstanceRelease(true)
	}

	return pbcommon.ErrCode_E_OK, ""
}

func (act *CreateReleaseAction) createAppInstanceRelease() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.AppInstanceRelease{})

	for _, info := range act.req.Infos {
		if len(info.Cfgsetid) != 0 {
			// normal sidecar effect configs report.
			act.createAppInstanceReleaseEffect(info)
		} else {
			// instance api reload report.
			act.createAppInstanceReleaseReload(info)
		}
	}

	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *CreateReleaseAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query appinstance.
	if errCode, errMsg := act.queryAppInstance(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create/update appinstance release.
	if errCode, errMsg := act.createAppInstanceRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
