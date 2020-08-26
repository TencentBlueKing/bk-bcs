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

	"github.com/bluele/gcache"
	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-datamanager/modules/metrics"
	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// QueryReleaseAction is appinstance release query action object.
type QueryReleaseAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	collector    *metrics.Collector
	configsCache gcache.Cache

	req  *pb.QueryAppInstanceReleaseReq
	resp *pb.QueryAppInstanceReleaseResp

	sd *dbsharding.ShardingDB

	appInstance        database.AppInstance
	appInstanceRelease database.AppInstanceRelease
	release            database.Release
}

// NewQueryReleaseAction creates new QueryReleaseAction.
func NewQueryReleaseAction(viper *viper.Viper, smgr *dbsharding.ShardingManager, collector *metrics.Collector, configsCache gcache.Cache,
	req *pb.QueryAppInstanceReleaseReq, resp *pb.QueryAppInstanceReleaseResp) *QueryReleaseAction {
	action := &QueryReleaseAction{viper: viper, smgr: smgr, collector: collector, configsCache: configsCache, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryReleaseAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *QueryReleaseAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryReleaseAction) Output() error {
	// do nothing.
	return nil
}

func (act *QueryReleaseAction) verify() error {
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

	length = len(act.req.Cfgsetid)
	if length == 0 {
		return errors.New("invalid params, cfgsetid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}
	return nil
}

func (act *QueryReleaseAction) queryAppInstance() (pbcommon.ErrCode, string) {
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

func (act *QueryReleaseAction) queryAppInstanceRelease() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.AppInstanceRelease{})

	var sts []database.AppInstanceRelease

	err := act.sd.DB().
		Limit(1).
		Order("Feffect_time DESC, Fid DESC").
		Where(&database.AppInstanceRelease{Instanceid: act.appInstance.ID, Cfgsetid: act.req.Cfgsetid}).
		Find(&sts).Error

	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	if len(sts) == 0 {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "app instance release non-exist."
	}
	act.appInstanceRelease = sts[0]

	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) queryRelease() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Release{})

	err := act.sd.DB().
		Where(&database.Release{Bid: act.req.Bid, Releaseid: act.appInstanceRelease.Releaseid}).
		Last(&act.release).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "release non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}

	act.resp.Releaseid = act.appInstanceRelease.Releaseid
	act.resp.Commitid = act.release.Commitid

	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) genConfigsCacheKey(bid, cfgsetid, commitid, appid, clusterid, zoneid, index string) string {
	return common.SHA256(bid + ":" + cfgsetid + ":" + commitid + ":" + appid + ":" + clusterid + ":" + zoneid + ":" + index)
}

func (act *QueryReleaseAction) queryIndexLevelConfigs() (pbcommon.ErrCode, string) {
	key := act.genConfigsCacheKey(act.req.Bid, act.req.Cfgsetid, act.release.Commitid,
		act.req.Appid, act.req.Clusterid, act.req.Zoneid, act.req.IP)

	if cache, err := act.configsCache.Get(key); err == nil && cache != nil {
		act.collector.StatConfigsCache(true)
		configs := cache.(*pbcommon.Configs)

		logger.V(3).Infof("QueryAppInstanceRelease[%d]| query configs cache hit[%d] success(index-level)", act.req.Seq, len(configs.Content))
		act.resp.Cid = configs.Cid
		act.resp.CfgLink = configs.CfgLink
		act.resp.Content = configs.Content
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatConfigsCache(false)

	var st database.Configs

	err := act.sd.DB().
		Where(&database.Configs{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Commitid: act.release.Commitid}).
		Where("Fappid = ?", act.req.Appid).
		Where("Fclusterid = ?", act.req.Clusterid).
		Where("Fzoneid = ?", act.req.Zoneid).
		Where("Findex = ?", act.req.IP).
		Last(&st).Error

	if err != nil {
		if err != dbsharding.RECORDNOTFOUND {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "can't find index level content for this configset!"
	}

	configs := &pbcommon.Configs{
		Bid:          st.Bid,
		Cfgsetid:     st.Cfgsetid,
		Appid:        st.Appid,
		Clusterid:    st.Clusterid,
		Zoneid:       st.Zoneid,
		Index:        st.Index,
		Commitid:     st.Commitid,
		Cid:          st.Cid,
		CfgLink:      st.CfgLink,
		Content:      st.Content,
		Creator:      st.Creator,
		LastModifyBy: st.LastModifyBy,
		Memo:         st.Memo,
		State:        st.State,
		CreatedAt:    st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	newCachekey := act.genConfigsCacheKey(st.Bid, st.Cfgsetid, st.Commitid, st.Appid, st.Clusterid, st.Zoneid, st.Index)

	if err := act.configsCache.Set(newCachekey, configs); err != nil {
		logger.Warn("QueryAppInstanceRelease[%d]| update local configs cache, %+v", act.req.Seq, err)
	}
	act.resp.Cid = configs.Cid
	act.resp.CfgLink = configs.CfgLink
	act.resp.Content = configs.Content

	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) queryZoneLevelConfigs() (pbcommon.ErrCode, string) {
	key := act.genConfigsCacheKey(act.req.Bid, act.req.Cfgsetid, act.release.Commitid,
		act.req.Appid, act.req.Clusterid, act.req.Zoneid, "")

	if cache, err := act.configsCache.Get(key); err == nil && cache != nil {
		act.collector.StatConfigsCache(true)
		configs := cache.(*pbcommon.Configs)

		logger.V(3).Infof("QueryAppInstanceRelease[%d]| query configs cache hit[%d] success(zone-level)", act.req.Seq, len(configs.Content))
		act.resp.Cid = configs.Cid
		act.resp.CfgLink = configs.CfgLink
		act.resp.Content = configs.Content
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatConfigsCache(false)

	var st database.Configs

	err := act.sd.DB().
		Where(&database.Configs{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Commitid: act.release.Commitid}).
		Where("Fappid = ?", act.req.Appid).
		Where("Fclusterid = ?", act.req.Clusterid).
		Where("Fzoneid = ?", act.req.Zoneid).
		Where("Findex = ''").
		Last(&st).Error

	if err != nil {
		if err != dbsharding.RECORDNOTFOUND {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "can't find zone level content for this configset!"
	}

	configs := &pbcommon.Configs{
		Bid:          st.Bid,
		Cfgsetid:     st.Cfgsetid,
		Appid:        st.Appid,
		Clusterid:    st.Clusterid,
		Zoneid:       st.Zoneid,
		Index:        st.Index,
		Commitid:     st.Commitid,
		Cid:          st.Cid,
		CfgLink:      st.CfgLink,
		Content:      st.Content,
		Creator:      st.Creator,
		LastModifyBy: st.LastModifyBy,
		Memo:         st.Memo,
		State:        st.State,
		CreatedAt:    st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	newCachekey := act.genConfigsCacheKey(st.Bid, st.Cfgsetid, st.Commitid, st.Appid, st.Clusterid, st.Zoneid, st.Index)

	if err := act.configsCache.Set(newCachekey, configs); err != nil {
		logger.Warn("QueryAppInstanceRelease[%d]| update local configs cache, %+v", act.req.Seq, err)
	}
	act.resp.Cid = configs.Cid
	act.resp.CfgLink = configs.CfgLink
	act.resp.Content = configs.Content

	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) queryClusterLevelConfigs() (pbcommon.ErrCode, string) {
	key := act.genConfigsCacheKey(act.req.Bid, act.req.Cfgsetid, act.release.Commitid,
		act.req.Appid, act.req.Clusterid, "", "")

	if cache, err := act.configsCache.Get(key); err == nil && cache != nil {
		act.collector.StatConfigsCache(true)
		configs := cache.(*pbcommon.Configs)

		logger.V(3).Infof("QueryAppInstanceRelease[%d]| query configs cache hit[%d] success(cluster-level)", act.req.Seq, len(configs.Content))
		act.resp.Cid = configs.Cid
		act.resp.CfgLink = configs.CfgLink
		act.resp.Content = configs.Content
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatConfigsCache(false)

	var st database.Configs

	err := act.sd.DB().
		Where(&database.Configs{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Commitid: act.release.Commitid}).
		Where("Fappid = ?", act.req.Appid).
		Where("Fclusterid = ?", act.req.Clusterid).
		Where("Fzoneid = ''").
		Where("Findex = ''").
		Last(&st).Error

	if err != nil {
		if err != dbsharding.RECORDNOTFOUND {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "can't find cluster level content for this configset!"
	}

	configs := &pbcommon.Configs{
		Bid:          st.Bid,
		Cfgsetid:     st.Cfgsetid,
		Appid:        st.Appid,
		Clusterid:    st.Clusterid,
		Zoneid:       st.Zoneid,
		Index:        st.Index,
		Commitid:     st.Commitid,
		Cid:          st.Cid,
		CfgLink:      st.CfgLink,
		Content:      st.Content,
		Creator:      st.Creator,
		LastModifyBy: st.LastModifyBy,
		Memo:         st.Memo,
		State:        st.State,
		CreatedAt:    st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	newCachekey := act.genConfigsCacheKey(st.Bid, st.Cfgsetid, st.Commitid, st.Appid, st.Clusterid, st.Zoneid, st.Index)

	if err := act.configsCache.Set(newCachekey, configs); err != nil {
		logger.Warn("QueryAppInstanceRelease[%d]| update local configs cache, %+v", act.req.Seq, err)
	}
	act.resp.Cid = configs.Cid
	act.resp.CfgLink = configs.CfgLink
	act.resp.Content = configs.Content

	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) queryAppLevelConfigs() (pbcommon.ErrCode, string) {
	key := act.genConfigsCacheKey(act.req.Bid, act.req.Cfgsetid, act.release.Commitid, act.req.Appid, "", "", "")

	if cache, err := act.configsCache.Get(key); err == nil && cache != nil {
		act.collector.StatConfigsCache(true)
		configs := cache.(*pbcommon.Configs)

		logger.V(3).Infof("QueryAppInstanceRelease[%d]| query configs cache hit[%d] success(app-level)", act.req.Seq, len(configs.Content))
		act.resp.Cid = configs.Cid
		act.resp.CfgLink = configs.CfgLink
		act.resp.Content = configs.Content
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatConfigsCache(false)

	var st database.Configs

	err := act.sd.DB().
		Where(&database.Configs{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Commitid: act.release.Commitid}).
		Where("Fappid = ?", act.req.Appid).
		Where("Fclusterid = ''").
		Where("Fzoneid = ''").
		Where("Findex = ''").
		Last(&st).Error

	if err != nil {
		if err != dbsharding.RECORDNOTFOUND {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "can't find app level content for this configset!"
	}

	configs := &pbcommon.Configs{
		Bid:          st.Bid,
		Cfgsetid:     st.Cfgsetid,
		Appid:        st.Appid,
		Clusterid:    st.Clusterid,
		Zoneid:       st.Zoneid,
		Index:        st.Index,
		Commitid:     st.Commitid,
		Cid:          st.Cid,
		CfgLink:      st.CfgLink,
		Content:      st.Content,
		Creator:      st.Creator,
		LastModifyBy: st.LastModifyBy,
		Memo:         st.Memo,
		State:        st.State,
		CreatedAt:    st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	newCachekey := act.genConfigsCacheKey(st.Bid, st.Cfgsetid, st.Commitid, st.Appid, st.Clusterid, st.Zoneid, st.Index)

	if err := act.configsCache.Set(newCachekey, configs); err != nil {
		logger.Warn("QueryAppInstanceRelease[%d]| update local configs cache, %+v", act.req.Seq, err)
	}
	act.resp.Cid = configs.Cid
	act.resp.CfgLink = configs.CfgLink
	act.resp.Content = configs.Content

	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryReleaseAction) queryConfigs() (pbcommon.ErrCode, string) {
	if len(act.req.IP) != 0 {
		errCode, errMsg := act.queryIndexLevelConfigs()
		if errCode == pbcommon.ErrCode_E_OK {
			return pbcommon.ErrCode_E_OK, ""
		}
		if errCode != pbcommon.ErrCode_E_DM_NOT_FOUND {
			return errCode, errMsg
		}
	}

	errCode, errMsg := act.queryZoneLevelConfigs()
	if errCode == pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_OK, ""
	}
	if errCode != pbcommon.ErrCode_E_DM_NOT_FOUND {
		return errCode, errMsg
	}

	errCode, errMsg = act.queryClusterLevelConfigs()
	if errCode == pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_OK, ""
	}
	if errCode != pbcommon.ErrCode_E_DM_NOT_FOUND {
		return errCode, errMsg
	}

	errCode, errMsg = act.queryAppLevelConfigs()
	if errCode == pbcommon.ErrCode_E_OK {
		return pbcommon.ErrCode_E_OK, ""
	}
	if errCode != pbcommon.ErrCode_E_DM_NOT_FOUND {
		return errCode, errMsg
	}
	return pbcommon.ErrCode_E_DM_RELEASE_CONFIGS_NOT_FOUND, "can't find any configs on the configset for this app instance!"
}

// Do makes the workflows of this action base on input messages.
func (act *QueryReleaseAction) Do() error {
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

	// query appinstance release.
	if errCode, errMsg := act.queryAppInstanceRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query release.
	if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query configs.
	if errCode, errMsg := act.queryConfigs(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
