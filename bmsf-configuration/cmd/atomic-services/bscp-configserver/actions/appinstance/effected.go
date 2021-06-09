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
	"fmt"
	"math"
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/orderedmap"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/types"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// EffectedAction query app instance list which effected target release.
type EffectedAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryEffectedAppInstancesReq
	resp *pb.QueryEffectedAppInstancesResp

	release *pbcommon.Release

	isEffectTimeout bool

	matchedInstances  []*pbcommon.AppInstanceRelease
	effectedInstances []*pbcommon.AppInstanceRelease
}

// NewEffectedAction creates new EffectedAction.
func NewEffectedAction(kit kit.Kit, viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryEffectedAppInstancesReq, resp *pb.QueryEffectedAppInstancesResp) *EffectedAction {
	action := &EffectedAction{kit: kit, viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Result = true
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *EffectedAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *EffectedAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *EffectedAction) Output() error {
	// do nothing.
	return nil
}

func (act *EffectedAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("app_id", act.req.AppId,
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

	if act.req.TimeoutSec == 0 {
		act.req.TimeoutSec = act.viper.GetInt32("server.effectTimeoutSec")
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

func (act *EffectedAction) queryRelease() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryReleaseReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		ReleaseId: act.req.ReleaseId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("QueryEffectedAppInstances[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryRelease, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.release = resp.Data

	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *EffectedAction) checkEffectTimeout() (pbcommon.ErrCode, string) {
	// TODO: add publish timestamp.
	baseTime, err := time.ParseInLocation("2006-01-02 15:04:05", act.release.CreatedAt, time.Local)
	if err != nil {
		baseTime = time.Now()
		logger.Warnf("QueryEffectedAppInstances[%s]| parse release[%s] effect timeout, local base timestamp, %+v",
			act.kit.Rid, act.req.ReleaseId, err)
	}

	if time.Now().Unix()-baseTime.Unix() > int64(act.req.TimeoutSec) {
		act.isEffectTimeout = true
	}
	return pbcommon.ErrCode_E_OK, "OK"
}

func (act *EffectedAction) queryMatchedAppInstances(index, limit int) ([]*pbcommon.AppInstanceRelease,
	pbcommon.ErrCode, string) {

	r := &pbdatamanager.QueryMatchedAppInstancesReq{
		Seq:        act.kit.Rid,
		BizId:      act.req.BizId,
		AppId:      act.req.AppId,
		StrategyId: act.release.StrategyId,
		Page:       &pbcommon.Page{Start: int32(index), Limit: int32(limit)},
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("QueryEffectedAppInstances[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryMatchedAppInstances(ctx, r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to datamanager QueryMatchedAppInstances, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return nil, resp.Code, resp.Message
	}

	appInstanceReleases := []*pbcommon.AppInstanceRelease{}

	for _, inst := range resp.Data.Info {
		appInstanceRelease := &pbcommon.AppInstanceRelease{
			InstanceId: inst.Id,
			BizId:      inst.BizId,
			AppId:      inst.AppId,
			CloudId:    inst.CloudId,
			Ip:         inst.Ip,
			Path:       inst.Path,
			Labels:     inst.Labels,
			CfgId:      act.req.CfgId,
			ReleaseId:  act.req.ReleaseId,
		}

		if inst.State == int32(pbcommon.AppInstanceState_INSS_OFFLINE) {
			// app instance offline from process app procattr.
			appInstanceRelease.EffectCode = types.EffectCodeOffline
			appInstanceRelease.EffectMsg = types.EffectMsgOffline
			appInstanceRelease.ReloadCode = types.ReloadCodeOffline
			appInstanceRelease.ReloadMsg = types.ReloadMsgOffline
		}
		appInstanceReleases = append(appInstanceReleases, appInstanceRelease)
	}

	return appInstanceReleases, pbcommon.ErrCode_E_OK, "OK"
}

func (act *EffectedAction) queryMatchedAppInstanceList() ([]*pbcommon.AppInstanceRelease, pbcommon.ErrCode, string) {
	appInstanceReleases := []*pbcommon.AppInstanceRelease{}

	index := 0
	limit := database.BSCPQUERYLIMITMB

	startTime := time.Now()

	for {
		instanceReleases, errCode, errMsg := act.queryMatchedAppInstances(index, limit)
		if errCode != pbcommon.ErrCode_E_OK {
			return nil, errCode, errMsg
		}
		appInstanceReleases = append(appInstanceReleases, instanceReleases...)

		if len(instanceReleases) < limit {
			break
		}
		index += len(instanceReleases)
	}

	logger.V(2).Infof("QueryEffectedAppInstances[%s]| query matched app instance list done, count[%d], cost: %+v",
		act.kit.Rid, len(appInstanceReleases), time.Since(startTime))

	return appInstanceReleases, pbcommon.ErrCode_E_OK, "OK"
}

func (act *EffectedAction) queryEffectedAppInstances(index, limit int) ([]*pbcommon.AppInstanceRelease,
	pbcommon.ErrCode, string) {

	r := &pbdatamanager.QueryEffectedAppInstancesReq{
		Seq:       act.kit.Rid,
		BizId:     act.req.BizId,
		CfgId:     act.req.CfgId,
		ReleaseId: act.req.ReleaseId,
		Page:      &pbcommon.Page{Start: int32(index), Limit: int32(limit)},
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("QueryEffectedAppInstances[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryEffectedAppInstances(ctx, r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN,
			fmt.Sprintf("request to datamanager QueryEffectedAppInstances, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return nil, resp.Code, resp.Message
	}
	return resp.Data.Info, pbcommon.ErrCode_E_OK, "OK"
}

func (act *EffectedAction) queryEffectedAppInstanceList() ([]*pbcommon.AppInstanceRelease, pbcommon.ErrCode, string) {
	appInstanceReleases := []*pbcommon.AppInstanceRelease{}

	index := 0
	limit := database.BSCPQUERYLIMITMB

	startTime := time.Now()

	for {
		instanceReleases, errCode, errMsg := act.queryEffectedAppInstances(index, limit)
		if errCode != pbcommon.ErrCode_E_OK {
			return nil, errCode, errMsg
		}
		appInstanceReleases = append(appInstanceReleases, instanceReleases...)

		if len(instanceReleases) < limit {
			break
		}
		index += len(instanceReleases)
	}

	for _, instanceRelease := range appInstanceReleases {
		// refix effect code and message for the app instance release
		// which has not do effect action success.
		if instanceRelease.EffectCode == types.EffectCodePending {
			instanceRelease.EffectMsg = types.EffectMsgPending
		}

		// refix reload code and message for the app instance release
		// which only do effect action not reload action.
		if instanceRelease.ReloadCode == types.ReloadCodePending {
			instanceRelease.ReloadMsg = types.ReloadMsgPending
		}
	}

	logger.V(2).Infof("QueryEffectedAppInstances[%s]| query effected app instance list done, count[%d], cost: %+v",
		act.kit.Rid, len(appInstanceReleases), time.Since(startTime))

	return appInstanceReleases, pbcommon.ErrCode_E_OK, "OK"
}

func (act *EffectedAction) instanceKey(cloudID, ip, path string) string {
	metadata := cloudID + ":" + ip + ":" + filepath.Clean(path)

	// gen sha1 as an instance key.
	key := common.SHA1(metadata)
	if len(key) != 0 {
		return key
	}
	return metadata
}

func (act *EffectedAction) effected() (pbcommon.ErrCode, string) {
	// effected instance map, instance key --> AppInstanceRelease.
	effectedInstancesMap := orderedmap.New()

	startTime := time.Now()

	for _, instance := range act.effectedInstances {
		key := act.instanceKey(instance.CloudId, instance.Ip, instance.Path)
		effectedInstancesMap.Set(key, instance)
	}

	// matched instance map, instance key --> AppInstanceRelease.
	matchedInstancesMap := orderedmap.New()

	for _, instance := range act.matchedInstances {
		key := act.instanceKey(instance.CloudId, instance.Ip, instance.Path)

		effectedInstance, isExist := effectedInstancesMap.Get(key)
		if isExist {
			matchedInstancesMap.Set(key, effectedInstance)
			continue
		}

		if instance.EffectCode == types.EffectCodeOffline ||
			instance.ReloadCode == types.ReloadCodeOffline {
			// offline app instance from procattr.
			matchedInstancesMap.Set(key, instance)
			continue
		}

		if act.isEffectTimeout {
			instance.EffectCode = types.EffectCodeTimeout
			instance.EffectMsg = types.EffectMsgTimeout
			instance.ReloadCode = types.ReloadCodeTimeout
			instance.ReloadMsg = types.ReloadMsgTimeout
		} else {
			instance.EffectCode = types.EffectCodePending
			instance.EffectMsg = types.EffectMsgPending
			instance.ReloadCode = types.ReloadCodePending
			instance.ReloadMsg = types.ReloadMsgPending
		}
		matchedInstancesMap.Set(key, instance)
	}

	// add new matched instances.
	for _, instance := range act.effectedInstances {
		key := act.instanceKey(instance.CloudId, instance.Ip, instance.Path)

		_, isExist := matchedInstancesMap.Get(key)
		if !isExist {
			matchedInstancesMap.Set(key, instance)
		}
	}

	// rebuild final app instance list and split pages in memory.
	finalEffectedAppInstances := []*pbcommon.AppInstanceRelease{}
	for el := matchedInstancesMap.Front(); el != nil; el = el.Next() {
		instanceRelease, ok := el.Value.(*pbcommon.AppInstanceRelease)
		if ok {
			finalEffectedAppInstances = append(finalEffectedAppInstances, instanceRelease)
		}
	}

	logger.V(2).Infof("QueryEffectedAppInstances[%s]| effected done, final effected instances count[%d], cost: %+v",
		act.kit.Rid, len(finalEffectedAppInstances), time.Since(startTime))

	start := act.req.Page.Start
	end := start + act.req.Page.Limit

	var totalCount int64
	if act.req.Page.ReturnTotal {
		totalCount = int64(len(finalEffectedAppInstances))
	}

	if int(start) >= len(finalEffectedAppInstances) {
		act.resp.Data = &pb.QueryEffectedAppInstancesResp_RespData{
			TotalCount: uint32(totalCount),
			Info:       []*pbcommon.AppInstanceRelease{},
		}
		return pbcommon.ErrCode_E_OK, ""
	}

	if int(end) >= len(finalEffectedAppInstances) {
		act.resp.Data = &pb.QueryEffectedAppInstancesResp_RespData{
			TotalCount: uint32(totalCount),
			Info:       finalEffectedAppInstances[start:],
		}
		return pbcommon.ErrCode_E_OK, ""
	}

	act.resp.Data = &pb.QueryEffectedAppInstancesResp_RespData{
		TotalCount: uint32(totalCount),
		Info:       finalEffectedAppInstances[start:end],
	}

	return pbcommon.ErrCode_E_OK, "OK"
}

// Do makes the workflows of this action base on input messages.
func (act *EffectedAction) Do() error {
	// query release info.
	if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// check release effect timeout.
	if errCode, errMsg := act.checkEffectTimeout(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query matched instance list.
	matchedInstances, errCode, errMsg := act.queryMatchedAppInstanceList()
	if errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	act.matchedInstances = matchedInstances

	// query effected instance list.
	effectedInstances, errCode, errMsg := act.queryEffectedAppInstanceList()
	if errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	act.effectedInstances = effectedInstances

	// build effected instance result.
	if errCode, errMsg := act.effected(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	return nil
}
