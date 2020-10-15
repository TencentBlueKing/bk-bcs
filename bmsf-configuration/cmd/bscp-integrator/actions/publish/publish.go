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

package publish

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/integrator"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/logger"
)

// PublishAction handles publishing logic actions.
type PublishAction struct {
	viper          *viper.Viper
	businessSvrCli pbbusinessserver.BusinessClient
	md             *structs.IntegrationMetadata

	req  *pb.IntegrateReq
	resp *pb.IntegrateResp

	business *pbcommon.Business
	app      *pbcommon.App
	commit   *pbcommon.Commit

	strategyid string
	releaseid  string
}

// NewPublishAction creates new PublishAction.
func NewPublishAction(viper *viper.Viper, businessSvrCli pbbusinessserver.BusinessClient, md *structs.IntegrationMetadata,
	req *pb.IntegrateReq, resp *pb.IntegrateResp) *PublishAction {
	action := &PublishAction{viper: viper, businessSvrCli: businessSvrCli, md: md, req: req, resp: resp}

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
		return act.Err(pbcommon.ErrCode_E_ITG_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *PublishAction) Output() error {
	// do nothing.
	return nil
}

func (act *PublishAction) verify() error {
	length := len(act.md.Spec.BusinessName)
	if length == 0 {
		return errors.New("invalid params, businessName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, businessName too long")
	}

	length = len(act.md.Spec.AppName)
	if length == 0 {
		return errors.New("invalid params, appName missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, appName too long")
	}

	length = len(act.md.Release.Commitid)
	if length == 0 {
		return errors.New("invalid params, commitid missing")
	}
	if length > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, commitid too long")
	}

	length = len(act.md.Release.Name)
	if length == 0 {
		return errors.New("invalid params, release name missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, release name too long")
	}

	if len(act.md.Release.StrategyName) > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, strategyName too long")
	}

	if len(act.md.Spec.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *PublishAction) queryBusiness() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.QueryBusinessReq{
		Seq:  act.req.Seq,
		Name: act.md.Spec.BusinessName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Publish[%d]| request to businessserver QueryBusiness, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.QueryBusiness(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver QueryBusiness, %+v", err)
	}
	act.business = resp.Business

	return resp.ErrCode, resp.ErrMsg
}

func (act *PublishAction) queryApp() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.QueryAppReq{
		Seq:  act.req.Seq,
		Bid:  act.business.Bid,
		Name: act.md.Spec.AppName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Publish[%d]| request to businessserver QueryApp, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.QueryApp(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver QueryApp, %+v", err)
	}
	act.app = resp.App

	return resp.ErrCode, resp.ErrMsg
}

func (act *PublishAction) queryCommit() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.QueryCommitReq{
		Seq:      act.req.Seq,
		Bid:      act.business.Bid,
		Commitid: act.md.Release.Commitid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Publish[%d]| request to businessserver QueryCommit, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.QueryCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver QueryCommit, %+v", err)
	}
	act.commit = resp.Commit

	return resp.ErrCode, resp.ErrMsg
}

func (act *PublishAction) queryClusterid(bid, appid, name string) (string, error) {
	// TODO not support cluster labels(from bking pipeline).
	r := &pbbusinessserver.QueryClusterReq{
		Seq:   act.req.Seq,
		Bid:   bid,
		Appid: appid,
		Name:  name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Publish[%d]| request to businessserver QueryCluster, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.QueryCluster(ctx, r)
	if err != nil {
		return "", err
	}
	if resp.ErrCode == pbcommon.ErrCode_E_DM_NOT_FOUND {
		return "", nil
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return "", errors.New(resp.ErrMsg)
	}
	return resp.Cluster.Clusterid, nil
}

func (act *PublishAction) queryZoneid(bid, appid, name string) (string, error) {
	r := &pbbusinessserver.QueryZoneReq{
		Seq:   act.req.Seq,
		Bid:   bid,
		Appid: appid,
		Name:  name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Publish[%d]| request to businessserver QueryZone, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.QueryZone(ctx, r)
	if err != nil {
		return "", err
	}
	if resp.ErrCode == pbcommon.ErrCode_E_DM_NOT_FOUND {
		return "", nil
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return "", errors.New(resp.ErrMsg)
	}
	return resp.Zone.Zoneid, nil
}

func (act *PublishAction) createStrategy() error {
	clusterids := []string{}
	for _, clusterName := range act.md.Release.Strategy.ClusterNames {
		clusterid, err := act.queryClusterid(act.business.Bid, act.app.Appid, clusterName)
		if err != nil {
			return fmt.Errorf("can't query clusterid[%+v] to create strategy for publishing, %+v", clusterName, err)
		}
		if len(clusterid) != 0 {
			clusterids = append(clusterids, clusterid)
		}
	}

	zoneids := []string{}
	for _, zoneName := range act.md.Release.Strategy.ZoneNames {
		zoneid, err := act.queryZoneid(act.business.Bid, act.app.Appid, zoneName)
		if err != nil {
			return fmt.Errorf("can't query zoneid[%+v] to create strategy for publishing, %+v", zoneName, err)
		}
		if len(zoneid) != 0 {
			zoneids = append(zoneids, zoneid)
		}
	}

	r := &pbbusinessserver.CreateStrategyReq{
		Seq:        act.req.Seq,
		Bid:        act.business.Bid,
		Appid:      act.app.Appid,
		Name:       act.md.Release.StrategyName,
		Clusterids: clusterids,
		Zoneids:    zoneids,
		Dcs:        act.md.Release.Strategy.Dcs,
		IPs:        act.md.Release.Strategy.IPs,
		Labels:     act.md.Release.Strategy.Labels,
		Memo:       act.md.Spec.Memo,
		Creator:    act.req.Operator,
		LabelsAnd:  act.md.Release.Strategy.LabelsAnd,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Publish[%d]| request to businessserver CreateStrategy, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.CreateStrategy(ctx, r)
	if err != nil {
		return fmt.Errorf("request to businessserver CreateStrategy, %+v", err)
	}

	if resp.ErrCode != pbcommon.ErrCode_E_OK &&
		resp.ErrCode != pbcommon.ErrCode_E_BS_ALREADY_EXISTS &&
		resp.ErrCode != pbcommon.ErrCode_E_DM_ALREADY_EXISTS {
		return errors.New(resp.ErrMsg)
	}
	act.resp.Strategyid = resp.Strategyid
	act.strategyid = resp.Strategyid

	return nil
}

func (act *PublishAction) queryStrategy(name string) (*pbcommon.Strategy, error) {
	r := &pbbusinessserver.QueryStrategyReq{
		Seq:   act.req.Seq,
		Bid:   act.business.Bid,
		Appid: act.app.Appid,
		Name:  name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Publish[%d]| request to businessserver QueryStrategy, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.QueryStrategy(ctx, r)
	if err != nil {
		return nil, err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return nil, errors.New(resp.ErrMsg)
	}
	return resp.Strategy, nil
}

func (act *PublishAction) handleStrategy() (pbcommon.ErrCode, string) {
	if len(act.md.Release.StrategyName) == 0 {
		return pbcommon.ErrCode_E_OK, ""
	}

	strategy, err := act.queryStrategy(act.md.Release.StrategyName)
	if err != nil {
		logger.Warn("Publish[%d]| query strategy[%+v] information failed, %+v", act.req.Seq, act.md.Release.StrategyName, err)
		if err := act.createStrategy(); err != nil {
			return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("can't handle strategy, %+v", err)
		}
	} else {
		act.strategyid = strategy.Strategyid
	}

	if len(act.strategyid) == 0 {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, "create release with strategy, but can't get strategyid finally"
	}
	return pbcommon.ErrCode_E_OK, ""
}

func (act *PublishAction) createRelease() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.CreateReleaseReq{
		Seq:        act.req.Seq,
		Bid:        act.business.Bid,
		Name:       act.md.Release.Name,
		Commitid:   act.md.Release.Commitid,
		Strategyid: act.strategyid,
		Memo:       act.md.Spec.Memo,
		Creator:    act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Publish[%d]| request to businessserver CreateRelease, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.CreateRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver CreateRelease, %+v", err)
	}
	act.resp.Releaseid = resp.Releaseid
	act.releaseid = resp.Releaseid

	return resp.ErrCode, resp.ErrMsg
}

func (act *PublishAction) publish() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.PublishReleaseReq{
		Seq:       act.req.Seq,
		Bid:       act.business.Bid,
		Releaseid: act.releaseid,
		Operator:  act.req.Operator,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Publish[%d]| request to businessserver PublishRelease, %+v", act.req.Seq, r)

	resp, err := act.businessSvrCli.PublishRelease(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_ITG_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver PublishRelease, %+v", err)
	}
	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *PublishAction) Do() error {
	// query busienss information used for publishing.
	if errCode, errMsg := act.queryBusiness(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query app information used for publishing.
	if errCode, errMsg := act.queryApp(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query commit information used for publishing.
	if errCode, errMsg := act.queryCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// handle publishing strategies.
	if errCode, errMsg := act.handleStrategy(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// create new release.
	if errCode, errMsg := act.createRelease(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// publish the release now.
	if errCode, errMsg := act.publish(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
