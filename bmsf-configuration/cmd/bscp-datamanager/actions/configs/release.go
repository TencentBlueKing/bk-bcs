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

// ReleaseConfigsAction is release configs query action object.
type ReleaseConfigsAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	collector    *metrics.Collector
	commitCache  gcache.Cache
	configsCache gcache.Cache

	req  *pb.QueryReleaseConfigsReq
	resp *pb.QueryReleaseConfigsResp

	sd *dbsharding.ShardingDB

	isTemplateMode bool
}

// NewReleaseConfigsAction creates new ReleaseConfigsAction.
func NewReleaseConfigsAction(viper *viper.Viper, smgr *dbsharding.ShardingManager, collector *metrics.Collector,
	commitCache, configsCache gcache.Cache, req *pb.QueryReleaseConfigsReq, resp *pb.QueryReleaseConfigsResp) *ReleaseConfigsAction {

	action := &ReleaseConfigsAction{viper: viper, smgr: smgr, collector: collector, commitCache: commitCache,
		configsCache: configsCache, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *ReleaseConfigsAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *ReleaseConfigsAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *ReleaseConfigsAction) Output() error {
	// do nothing.
	return nil
}

func (act *ReleaseConfigsAction) verify() error {
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

	if len(act.req.Index) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, index too long")
	}

	length = len(act.req.Cfgsetid)
	if length == 0 {
		return errors.New("invalid params, cfgsetid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, cfgsetid too long")
	}

	length = len(act.req.Commitid)
	if length == 0 {
		return errors.New("invalid params, commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
	}
	return nil
}

func (act *ReleaseConfigsAction) queryCommit() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Commit{})

	// query commit in cache.
	if cache, err := act.commitCache.Get(act.req.Commitid); err == nil && cache != nil {
		act.collector.StatCommitCache(true)
		commit := cache.(*pbcommon.Commit)

		logger.V(3).Infof("QueryReleaseConfigs[%d]| query commit cache hit success[%s]", act.req.Seq, act.req.Commitid)

		if len(commit.Template) != 0 || len(commit.Templateid) != 0 {
			act.isTemplateMode = true
		}
	} else {
		// can't find commit in cache.
		act.collector.StatCommitCache(false)
		commit := database.Commit{}

		err := act.sd.DB().
			Where(&database.Commit{Bid: act.req.Bid, Commitid: act.req.Commitid}).
			Last(&commit).Error

		// not found.
		if err == dbsharding.RECORDNOTFOUND {
			return pbcommon.ErrCode_E_DM_NOT_FOUND, "commit non-exist."
		}
		if err != nil {
			return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
		}

		if commit.State != int32(pbcommon.CommitState_CS_CONFIRMED) {
			return pbcommon.ErrCode_E_DM_SYSTEM_UNKONW, "target commit is not confirmed"
		}

		newCache := &pbcommon.Commit{
			Bid:           commit.Bid,
			Commitid:      commit.Commitid,
			Appid:         commit.Appid,
			Cfgsetid:      commit.Cfgsetid,
			Op:            commit.Op,
			Operator:      commit.Operator,
			Templateid:    commit.Templateid,
			Template:      commit.Template,
			TemplateRule:  commit.TemplateRule,
			PrevConfigs:   commit.PrevConfigs,
			Configs:       commit.Configs,
			Changes:       commit.Changes,
			Releaseid:     commit.Releaseid,
			Memo:          commit.Memo,
			State:         commit.State,
			CreatedAt:     commit.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     commit.UpdatedAt.Format("2006-01-02 15:04:05"),
			MultiCommitid: commit.MultiCommitid,
		}

		if err := act.commitCache.Set(act.req.Commitid, newCache); err != nil {
			logger.Warnf("QueryReleaseConfigs[%d]| set commit[%s] cache failed, %+v", act.req.Seq, act.req.Commitid, err)
		}

		if len(commit.Template) != 0 || len(commit.Templateid) != 0 {
			act.isTemplateMode = true
		}
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *ReleaseConfigsAction) genConfigsCacheKey(bid, cfgsetid, commitid, appid, clusterid, zoneid, index string) string {
	return common.SHA256(bid + ":" + cfgsetid + ":" + commitid + ":" + appid + ":" + clusterid + ":" + zoneid + ":" + index)
}

func (act *ReleaseConfigsAction) queryIndexLevelConfigs() (pbcommon.ErrCode, string) {
	key := act.genConfigsCacheKey(act.req.Bid, act.req.Cfgsetid, act.req.Commitid,
		act.req.Appid, act.req.Clusterid, act.req.Zoneid, act.req.Index)

	if cache, err := act.configsCache.Get(key); err == nil && cache != nil {
		act.collector.StatConfigsCache(true)
		configs := cache.(*pbcommon.Configs)

		logger.V(3).Infof("QueryReleaseConfigs[%d]| query configs cache hit[%d] success(index-level)", act.req.Seq, len(configs.Content))
		finalConfigs := *configs
		if act.req.Abstract {
			finalConfigs.Content = []byte{}
		}
		act.resp.Configs = &finalConfigs
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatConfigsCache(false)

	var st database.Configs

	err := act.sd.DB().
		Where(&database.Configs{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Commitid: act.req.Commitid}).
		Where("Fappid = ?", act.req.Appid).
		Where("Fclusterid = ?", act.req.Clusterid).
		Where("Fzoneid = ?", act.req.Zoneid).
		Where("Findex = ?", act.req.Index).
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
	newCache := *configs

	if err := act.configsCache.Set(newCachekey, &newCache); err != nil {
		logger.Warn("QueryReleaseConfigs[%d]| update local configs cache, %+v", act.req.Seq, err)
	}
	if act.req.Abstract {
		configs.Content = []byte{}
	}
	act.resp.Configs = configs

	return pbcommon.ErrCode_E_OK, ""
}

func (act *ReleaseConfigsAction) queryZoneLevelConfigs() (pbcommon.ErrCode, string) {
	key := act.genConfigsCacheKey(act.req.Bid, act.req.Cfgsetid, act.req.Commitid,
		act.req.Appid, act.req.Clusterid, act.req.Zoneid, "")

	if cache, err := act.configsCache.Get(key); err == nil && cache != nil {
		act.collector.StatConfigsCache(true)
		configs := cache.(*pbcommon.Configs)

		logger.V(3).Infof("QueryReleaseConfigs[%d]| query configs cache hit[%d] success(zone-level)", act.req.Seq, len(configs.Content))
		finalConfigs := *configs
		if act.req.Abstract {
			finalConfigs.Content = []byte{}
		}
		act.resp.Configs = &finalConfigs
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatConfigsCache(false)

	var st database.Configs

	err := act.sd.DB().
		Where(&database.Configs{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Commitid: act.req.Commitid}).
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
	newCache := *configs

	if err := act.configsCache.Set(newCachekey, &newCache); err != nil {
		logger.Warn("QueryReleaseConfigs[%d]| update local configs cache, %+v", act.req.Seq, err)
	}
	if act.req.Abstract {
		configs.Content = []byte{}
	}
	act.resp.Configs = configs

	return pbcommon.ErrCode_E_OK, ""
}

func (act *ReleaseConfigsAction) queryClusterLevelConfigs() (pbcommon.ErrCode, string) {
	key := act.genConfigsCacheKey(act.req.Bid, act.req.Cfgsetid, act.req.Commitid,
		act.req.Appid, act.req.Clusterid, "", "")

	if cache, err := act.configsCache.Get(key); err == nil && cache != nil {
		act.collector.StatConfigsCache(true)
		configs := cache.(*pbcommon.Configs)

		logger.V(3).Infof("QueryReleaseConfigs[%d]| query configs cache hit[%d] success(cluster-level)", act.req.Seq, len(configs.Content))
		finalConfigs := *configs
		if act.req.Abstract {
			finalConfigs.Content = []byte{}
		}
		act.resp.Configs = &finalConfigs
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatConfigsCache(false)

	var st database.Configs

	err := act.sd.DB().
		Where(&database.Configs{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Commitid: act.req.Commitid}).
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
	newCache := *configs

	if err := act.configsCache.Set(newCachekey, &newCache); err != nil {
		logger.Warn("QueryReleaseConfigs[%d]| update local configs cache, %+v", act.req.Seq, err)
	}
	if act.req.Abstract {
		configs.Content = []byte{}
	}
	act.resp.Configs = configs

	return pbcommon.ErrCode_E_OK, ""
}

func (act *ReleaseConfigsAction) queryAppLevelConfigs() (pbcommon.ErrCode, string) {
	key := act.genConfigsCacheKey(act.req.Bid, act.req.Cfgsetid, act.req.Commitid, act.req.Appid, "", "", "")

	if cache, err := act.configsCache.Get(key); err == nil && cache != nil {
		act.collector.StatConfigsCache(true)
		configs := cache.(*pbcommon.Configs)

		logger.V(3).Infof("QueryReleaseConfigs[%d]| query configs cache hit[%d] success(app-level)", act.req.Seq, len(configs.Content))
		finalConfigs := *configs
		if act.req.Abstract {
			finalConfigs.Content = []byte{}
		}
		act.resp.Configs = &finalConfigs
		return pbcommon.ErrCode_E_OK, ""
	}
	act.collector.StatConfigsCache(false)

	var st database.Configs

	err := act.sd.DB().
		Where(&database.Configs{Bid: act.req.Bid, Cfgsetid: act.req.Cfgsetid, Commitid: act.req.Commitid}).
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
	newCache := *configs

	if err := act.configsCache.Set(newCachekey, &newCache); err != nil {
		logger.Warn("QueryReleaseConfigs[%d]| update local configs cache, %+v", act.req.Seq, err)
	}
	if act.req.Abstract {
		configs.Content = []byte{}
	}
	act.resp.Configs = configs

	return pbcommon.ErrCode_E_OK, ""
}

func (act *ReleaseConfigsAction) queryConfigs() (pbcommon.ErrCode, string) {
	if !act.isTemplateMode {
		// not template mode, just query app level configs.
		errCode, errMsg := act.queryAppLevelConfigs()
		if errCode == pbcommon.ErrCode_E_OK {
			return pbcommon.ErrCode_E_OK, ""
		}
		if errCode != pbcommon.ErrCode_E_DM_NOT_FOUND {
			return errCode, errMsg
		}
		return pbcommon.ErrCode_E_DM_RELEASE_CONFIGS_NOT_FOUND, "can't find any level content for this configset!"
	}

	// foreach every level configs, index(optional) > zone > cluster > app.
	if len(act.req.Index) != 0 {
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
	return pbcommon.ErrCode_E_DM_RELEASE_CONFIGS_NOT_FOUND, "can't find any level content for this template configset!"
}

// Do makes the workflows of this action base on input messages.
func (act *ReleaseConfigsAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd

	// query commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query configs.
	if errCode, errMsg := act.queryConfigs(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
