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
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// PublishAction is release publish action object.
type PublishAction struct {
	viper *viper.Viper
	smgr  *dbsharding.ShardingManager

	releaseCache gcache.Cache
	configsCache gcache.Cache

	req  *pb.PublishReleaseReq
	resp *pb.PublishReleaseResp

	sd *dbsharding.ShardingDB
	tx *gorm.DB

	release database.Release
	commit  database.Commit
}

// NewPublishAction creates new PublishAction.
func NewPublishAction(viper *viper.Viper, smgr *dbsharding.ShardingManager, releaseCache, configsCache gcache.Cache,
	req *pb.PublishReleaseReq, resp *pb.PublishReleaseResp) *PublishAction {
	action := &PublishAction{viper: viper, smgr: smgr, releaseCache: releaseCache, configsCache: configsCache, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *PublishAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *PublishAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PublishAction) Output() error {
	// do nothing.
	return nil
}

func (act *PublishAction) verify() error {
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

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}
	return nil
}

func (act *PublishAction) queryRelease() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Release{})

	err := act.tx.Where(&database.Release{Bid: act.req.Bid, Releaseid: act.req.Releaseid}).
		Last(&act.release).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "release non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) queryCommit() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Commit{})

	err := act.tx.Where(&database.Commit{Bid: act.req.Bid, Commitid: act.release.Commitid, State: int32(pbcommon.CommitState_CS_CONFIRMED)}).
		Last(&act.commit).Error

	// not found.
	if err == dbsharding.RECORDNOTFOUND {
		return pbcommon.ErrCode_E_DM_NOT_FOUND, "target release confirmed commit non-exist."
	}
	if err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) publishRelease() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Release{})

	ups := map[string]interface{}{
		"State":        int32(pbcommon.ReleaseState_RS_PUBLISHED),
		"LastModifyBy": act.req.Operator,
	}

	exec := act.tx.Model(&database.Release{}).
		Where(&database.Release{Bid: act.req.Bid, Releaseid: act.req.Releaseid}).
		Where("Fstate IN (?, ?)", pbcommon.ReleaseState_RS_INIT, pbcommon.ReleaseState_RS_PUBLISHED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_PUBLISH_RELEASE_FAILED, "publish the release failed, there is no release that fit the conditions."
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) updateCommit() (pbcommon.ErrCode, string) {
	act.sd.AutoMigrate(&database.Commit{})

	ups := map[string]interface{}{
		"Releaseid":    act.req.Releaseid,
		"LastModifyBy": act.req.Operator,
	}

	exec := act.tx.Model(&database.Commit{}).
		Where(&database.Commit{Bid: act.req.Bid, Commitid: act.commit.Commitid}).
		Where("Fstate = ?", pbcommon.CommitState_CS_CONFIRMED).
		Updates(ups)

	if err := exec.Error; err != nil {
		return pbcommon.ErrCode_E_DM_DB_EXEC_ERR, err.Error()
	}
	if exec.RowsAffected == 0 {
		return pbcommon.ErrCode_E_DM_DB_UPDATE_ERR, "publish release and update the commit failed(commit no-exist or not in confirmed state)."
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) queryConfigsList(bid, cfgsetid, commitid string, index, limit int32) ([]database.Configs, error) {
	act.sd.AutoMigrate(&database.Configs{})

	configsList := []database.Configs{}

	err := act.sd.DB().
		Offset(index).Limit(limit).
		Order("Fupdate_time DESC, Fid DESC").
		Where(&database.Configs{Bid: bid, Cfgsetid: cfgsetid, Commitid: commitid}).
		Find(&configsList).Error

	if err != nil {
		return nil, err
	}
	return configsList, nil
}

func (act *PublishAction) genConfigsCacheKey(bid, cfgsetid, commitid, appid, clusterid, zoneid, index string) string {
	return common.SHA256(bid + ":" + cfgsetid + ":" + commitid + ":" + appid + ":" + clusterid + ":" + zoneid + ":" + index)
}

func (act *PublishAction) addConfigsCache() error {
	index := 0
	limit := 1

	for {
		list, err := act.queryConfigsList(act.release.Bid, act.release.Cfgsetid, act.release.Commitid, int32(index), int32(limit))
		if err != nil {
			return err
		}

		// query success.
		for _, cfg := range list {
			key := act.genConfigsCacheKey(cfg.Bid, cfg.Cfgsetid, cfg.Commitid, cfg.Appid, cfg.Clusterid, cfg.Zoneid, cfg.Index)

			configs := &pbcommon.Configs{
				Bid:          cfg.Bid,
				Cfgsetid:     cfg.Cfgsetid,
				Appid:        cfg.Appid,
				Clusterid:    cfg.Clusterid,
				Zoneid:       cfg.Zoneid,
				Index:        cfg.Index,
				Commitid:     cfg.Commitid,
				Cid:          cfg.Cid,
				CfgLink:      cfg.CfgLink,
				Content:      cfg.Content,
				Creator:      cfg.Creator,
				LastModifyBy: cfg.LastModifyBy,
				Memo:         cfg.Memo,
				State:        cfg.State,
				CreatedAt:    cfg.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:    cfg.UpdatedAt.Format("2006-01-02 15:04:05"),
			}

			act.configsCache.Set(key, configs)
			logger.Info("PublishRelease[%d]| add configs cache in publish event, len[%d], %s", act.req.Seq, len(configs.Content), key)
		}

		if len(list) < limit {
			break
		}
		index += len(list)
	}

	return nil
}

// Do makes the workflows of this action base on input messages.
func (act *PublishAction) Do() error {
	// business sharding db.
	sd, err := act.smgr.ShardingDB(act.req.Bid)
	if err != nil {
		return act.Err(pbcommon.ErrCode_E_DM_ERR_DBSHARDING, err.Error())
	}
	act.sd = sd
	act.tx = act.sd.DB().Begin()

	// query release.
	if errCode, errMsg := act.queryRelease(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// query commit.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// publish release.
	if errCode, errMsg := act.publishRelease(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}

	// update commit.
	if errCode, errMsg := act.updateCommit(); errCode != pbcommon.ErrCode_E_OK {
		act.tx.Rollback()
		return act.Err(errCode, errMsg)
	}
	act.tx.Commit()
	act.releaseCache.Remove(act.req.Releaseid)

	// add cache in prev mode.
	if err := act.addConfigsCache(); err != nil {
		logger.Warn("PublishRelease[%d]| add configs cache in publish event failed, %+v", act.req.Seq, err)
	}

	return nil
}
