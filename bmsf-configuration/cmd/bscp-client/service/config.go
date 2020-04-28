/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package service

import (
	"context"
	"encoding/json"
	"fmt"

	global "bk-bscp/cmd/bscp-client/option"
	"bk-bscp/internal/protocol/accessserver"
	"bk-bscp/internal/protocol/common"
	"bk-bscp/internal/strategy"
	pkgcommon "bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"

	"google.golang.org/grpc"
)

//CreateConfigSet create new application
//return:
//	appID: when creating successfully, system will response ID for application
//	error: any error if happened
func (operator *AccessOperator) CreateConfigSet(cxt context.Context, option *accessserver.CreateConfigSetReq) (string, error) {
	if option == nil {
		return "", fmt.Errorf("Lost create new ConfigSet info")
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	option.Seq = pkgcommon.Sequence()
	response, err := operator.Client.CreateConfigSet(cxt, option, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CreateConfigSet: post new ConfigSet [%s] failed, %s", option.Name, err.Error())
		return "", err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof(
			"CreateConfigSet: post new configset [%s] successfully, but response failed: %s",
			option.Name, response.ErrMsg,
		)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.Cfgsetid) == 0 {
		logger.V(3).Infof("CreateConfigSet: BSCP system error, No ConfigSetID response")
		return "", fmt.Errorf("Lost ConfigSetID from configuraiotn platform")
	}
	return response.Cfgsetid, nil
}

//GetConfigSet get specified configset information
//Args:
//	name: app name
//return:
//	app: specified configset, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetConfigSet(cxt context.Context, appName, cfgSetName string) (*common.ConfigSet, error) {
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	return operator.innergetConfigSet(cxt, business.Bid, app.Appid, cfgSetName)
}

func (operator *AccessOperator) innergetConfigSet(cxt context.Context, businessID, AppID, cfgSetName string) (*common.ConfigSet, error) {
	request := &accessserver.QueryConfigSetReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   businessID,
		Appid: AppID,
		Name:  cfgSetName,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryConfigSet(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetConfigSet %s failed, %s", cfgSetName, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetConfigSet %s: resource Not Found.", cfgSetName)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetConfigSet %s successfully, but response Err, %s", cfgSetName, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.ConfigSet, nil
}

//ListConfigSetByApp list all ConfigSet information under specified application
//return:
//	configsets: all configset, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListConfigSetByApp(cxt context.Context, appName string) ([]*common.ConfigSet, error) {
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	request := &accessserver.QueryConfigSetListReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   business.Bid,
		Appid: app.Appid,
		Index: operator.index,
		Limit: operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryConfigSetList(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListConfigSetByApp failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListConfigSetByApp all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.ConfigSets, nil
}

//UpdateConfigSet update specified ConfigSet information
//Args:
//	appID: specified AppID
//	option: updated information[now only name, DeployType]
//return:
//	error: error info if that happened
func (operator *AccessOperator) UpdateConfigSet(cxt context.Context, cfgSetName string, option *accessserver.UpdateConfigSetReq) error {
	//do not implemented
	return fmt.Errorf("Do not Implemented")
}

//DeleteConfigSetOption option for configset deletion
type DeleteConfigSetOption struct {
	AppName  string
	CfgSetID string
	Operator string
}

//DeleteConfigSet delete specified ConfigSet
//
//return:
//	err: if any error happened
func (operator *AccessOperator) DeleteConfigSet(cxt context.Context, option *DeleteConfigSetOption) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("DeleteConfigSet: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}

	//todo(DeveloperJim): confirm delete operation
	request := &accessserver.DeleteConfigSetReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      business.Bid,
		Cfgsetid: option.CfgSetID,
		Operator: option.Operator,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.DeleteConfigSet(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("DeleteConfigSet %s failed, %s", option.CfgSetID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("DeleteConfigSet %s successfully, but response Err, %s", option.CfgSetID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//LockConfigSet lock specified ConfigSet
//Args:
//	cfgSetID: configset string ID
//return:
//	err: if any error happened
func (operator *AccessOperator) LockConfigSet(cxt context.Context, cfgSetID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("LockConfigSet: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}
	//todo(DeveloperJim): confirm delete operation
	request := &accessserver.LockConfigSetReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      business.Bid,
		Cfgsetid: cfgSetID,
		Operator: global.GlobalOptions.Operator,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.LockConfigSet(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("LockConfigSet %s failed, %s", cfgSetID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("LockConfigSet %s successfully, but response Err, %s", cfgSetID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//UnLockConfigSet unlock specified ConfigSet
//Args:
//	cfgSetID: configset string ID
//return:
//	err: if any error happened
func (operator *AccessOperator) UnLockConfigSet(cxt context.Context, cfgSetID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("UnLockConfigSet: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}
	//todo(DeveloperJim): confirm delete operation
	request := &accessserver.UnlockConfigSetReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      business.Bid,
		Cfgsetid: cfgSetID,
		Operator: global.GlobalOptions.Operator,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.UnlockConfigSet(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("unLockConfigSet %s failed, %s", cfgSetID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("unLockConfigSet %s successfully, but response Err, %s", cfgSetID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//CreateCommitOption option for create Commit
type CreateCommitOption struct {
	//AppName for app relation
	AppName string
	//ConfigSetName for cfgset relation
	ConfigSetName string
	//Content config details
	Content []byte
	//Changes for diff with last commit
	Changes string
	//Template for future use
	Template string
	//TemplateID index from other system
	TemplateID string
}

//CreateCommit create commit for ConfigSet
//return:
//	appID: when creating successfully, system will response ID
//	error: any error if happened
func (operator *AccessOperator) CreateCommit(cxt context.Context, option *CreateCommitOption) (string, error) {
	if option == nil {
		return "", fmt.Errorf("Lost create Commit info")
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, option.AppName)
	if err != nil {
		return "", err
	}

	//query ConfigSet information for specified ID
	cfgSet, err := operator.innergetConfigSet(cxt, business.Bid, app.Appid, option.ConfigSetName)
	if err != nil {
		return "", err
	}
	if cfgSet == nil {
		return "", fmt.Errorf("Found no ConfigSet info")
	}

	request := &accessserver.CreateCommitReq{
		Seq:        pkgcommon.Sequence(),
		Bid:        business.Bid,
		Appid:      app.Appid,
		Cfgsetid:   cfgSet.Cfgsetid,
		Op:         0,
		Operator:   operator.User,
		Templateid: "",
		Template:   "",
		Configs:    option.Content,
		Changes:    "",
	}
	response, err := operator.Client.CreateCommit(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CreateCommit: post new Commit for ConfigSet [%s] failed, %s", option.ConfigSetName, err.Error())
		return "", err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof(
			"CreateCommit: post new new Commit for ConfigSet [%s] successfully, but response failed: %s",
			option.ConfigSetName, response.ErrMsg,
		)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.Commitid) == 0 {
		logger.V(3).Infof("CreateConfigSet: BSCP system error, No CommitID response")
		return "", fmt.Errorf("Lost CommitID from configuraiotn platform")
	}
	return response.Commitid, nil
}

//GetCommit get specified Commit details information
//Args:
//	name: app name
//return:
//	app: specified configset, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetCommit(cxt context.Context, commitID string) (*common.Commit, error) {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return nil, err
	}
	if business == nil {
		logger.V(3).Infof("GetCommit: found no relative Business %s", operator.Business)
		return nil, fmt.Errorf("No relative Business %s", operator.Business)
	}

	request := &accessserver.QueryCommitReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      business.Bid,
		Commitid: commitID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryCommit(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetCommit %s details failed, %s", commitID, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetCommit %s: resource Not Found.", commitID)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetCommit %s successfully, but response Err, %s", commitID, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Commit, nil
}

//ListCommitsByConfigSet list all Commits of ConfigSet information
//return:
//	Commits: all Commit, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListCommitsByConfigSet(cxt context.Context, appName, cfgSetName string) ([]*common.Commit, error) {
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	return operator.ListCommitsAllByID(cxt, business.Bid, app.Appid, cfgSetName)
}

//ListCommitsAllByID list all Commits of ConfigSet information
//return:
//	Commits: all Commit, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListCommitsAllByID(cxt context.Context, bID, appID, cfgSetName string) ([]*common.Commit, error) {
	configSet, err := operator.innergetConfigSet(cxt, bID, appID, cfgSetName)
	if err != nil {
		return nil, err
	}
	if configSet == nil {
		return nil, nil
	}

	request := &accessserver.QueryHistoryCommitsReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      bID,
		Appid:    appID,
		Cfgsetid: configSet.Cfgsetid,
		Index:    operator.index,
		Limit:    operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryHistoryCommits(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListCommitsByConfigSet failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListCommitsByConfigSet all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Commits, nil
}

//UpdateCommitOption for commit update
type UpdateCommitOption struct {
	CommitID   string
	TemplateID string
	Template   string
	Configs    []byte
	Changes    string
}

//UpdateCommit update specified Commit information.
//Args:
//	option: updated information
//return:
//	error: error info if that happened
func (operator *AccessOperator) UpdateCommit(cxt context.Context, option *UpdateCommitOption) error {
	//business first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("UpdateCommit: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}

	request := &accessserver.UpdateCommitReq{
		Seq:        pkgcommon.Sequence(),
		Bid:        business.Bid,
		Commitid:   option.CommitID,
		Templateid: option.TemplateID,
		Template:   option.Template,
		Configs:    option.Configs,
		Changes:    option.Changes,
		Operator:   operator.User,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.UpdateCommit(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("UpdateCommit %s failed, %s", option.CommitID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("UpdateCommit %s successfully, but response Err, %s", option.CommitID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil

}

//DeleteCommit delete specified Commit, only setting unvisible
//return:
//	err: if any error happened
func (operator *AccessOperator) DeleteCommit(cxt context.Context, commitID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("DeleteCommit: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}

	//setting cancel commit request
	request := &accessserver.CancelCommitReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      business.Bid,
		Commitid: commitID,
		Operator: operator.User,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.CancelCommit(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("DeleteCommit %s failed, %s", commitID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("DeleteCommit %s successfully, but response Err, %s", commitID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//ConfirmCommit make commit affective or generate from templates
//return:
//	err: if any error happened
func (operator *AccessOperator) ConfirmCommit(cxt context.Context, commitID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("ConfirmCommit: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}

	//setting cancel commit request
	request := &accessserver.ConfirmCommitReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      business.Bid,
		Commitid: commitID,
		Operator: operator.User,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.ConfirmCommit(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ConfirmCommit %s failed, %s", commitID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ConfirmCommit %s successfully, but response Err, %s", commitID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//StrategyOption use for Create and Update
type StrategyOption struct {
	//StrategyID id for strategy, response by system
	StrategyID string
	//Name for strategy
	Name string
	//AppName relative app name
	AppName string
	//Content strategy details, it's json data
	Content string
}

//CreateStrategy create strategy for release
//return:
//	strategyID: when creating successfully, system will response ID
//	error: any error if happened
func (operator *AccessOperator) CreateStrategy(cxt context.Context, option *StrategyOption) (string, error) {
	if option == nil {
		return "", fmt.Errorf("Lost create Commit info")
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, option.AppName)
	if err != nil {
		return "", err
	}

	strategies := &strategy.Strategy{}
	if err := json.Unmarshal([]byte(option.Content), strategies); err != nil {
		return "", err
	}

	request := &accessserver.CreateStrategyReq{
		Seq: pkgcommon.Sequence(),
		Bid: business.Bid,
		//todo(DeveloperJim): check AppID necessity
		Appid:      app.Appid,
		Name:       option.Name,
		Clusterids: strategies.Clusterids,
		Zoneids:    strategies.Zoneids,
		Dcs:        strategies.Dcs,
		IPs:        strategies.IPs,
		Labels:     strategies.Labels,
		Creator:    operator.User,
	}
	response, err := operator.Client.CreateStrategy(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CreateStrategy: post new Strategy %s for Business [%s] failed, %s", option.Name, operator.Business, err.Error())
		return "", err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof(
			"CreateStrategy: post new new Stratgy %s for Business [%s] successfully, but response failed: %s",
			option.Name, operator.Business, response.ErrMsg,
		)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.Strategyid) == 0 {
		logger.V(3).Infof("CreateStrategy: BSCP system error, No StrategyID response")
		return "", fmt.Errorf("Lost StrategyID from configuraiotn platform")
	}
	return response.Strategyid, nil
}

//GetStrategy get specified Strategy details information
//Args:
//	name: strategy name
//	ID: strategy ID
//return:
//	strategy: specified configset, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetStrategy(cxt context.Context, appName, name string) (*common.Strategy, error) {
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	return operator.innerGetStrategyByID(cxt, business.Bid, app.Appid, name)
}

func (operator *AccessOperator) innerGetStrategyByID(cxt context.Context, bID, appID, name string) (*common.Strategy, error) {
	request := &accessserver.QueryStrategyReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   bID,
		Appid: appID,
		Name:  name,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryStrategy(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetStrategy %s details failed, %s", name, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetStrategy %s: resource Not Found.", name)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetStrategy %s successfully, but response Err, %s", name, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Strategy, nil
}

//ListStrategyByApp list all strategy information.
//return:
//	strategies: all strategies, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListStrategyByApp(cxt context.Context, appName string) ([]*common.Strategy, error) {
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	request := &accessserver.QueryStrategyListReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   business.Bid,
		Appid: app.Appid,
		Index: operator.index,
		Limit: operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryStrategyList(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListStrategyByApp failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListStrategyByApp all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Strategies, nil
}

//DeleteStrategy delete specified Strategy
//return:
//	err: if any error happened
func (operator *AccessOperator) DeleteStrategy(cxt context.Context, ID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("DeleteStrategy: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}

	//setting cancel commit request
	request := &accessserver.DeleteStrategyReq{
		Seq:        pkgcommon.Sequence(),
		Bid:        business.Bid,
		Strategyid: ID,
		Operator:   operator.User,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.DeleteStrategy(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("DeleteCommit %s failed, %s", ID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("DeleteCommit %s successfully, but response Err, %s", ID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//ReleaseOption for CRUD
type ReleaseOption struct {
	//ReleaseID for update
	ReleaseID string
	//Name for release
	Name string
	//AppName release for specified App
	AppName string
	//CfgSetName release for specified ConfigSet
	CfgSetName string
	// StrategyName release relative to strategy.
	StrategyName string
	//CommitID release relative to Commit
	CommitID string
}

//CreateRelease create release for specified Commit
//return:
//	releaseID: when creating successfully, system will response ID
//	error: any error if happened
func (operator *AccessOperator) CreateRelease(cxt context.Context, option *ReleaseOption) (string, error) {
	if option == nil {
		return "", fmt.Errorf("Lost create Commit info")
	}

	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}

	// query business first.
	business, app, err := getBusinessAndApp(operator, operator.Business, option.AppName)
	if err != nil {
		return "", err
	}

	request := &accessserver.CreateReleaseReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      business.Bid,
		Name:     option.Name,
		Commitid: option.CommitID,
		Creator:  operator.User,
	}

	// check strategy for this release.
	if len(option.StrategyName) != 0 {
		strategy, styerr := operator.innerGetStrategyByID(cxt, business.Bid, app.Appid, option.StrategyName)
		if styerr != nil {
			return "", styerr
		}

		if strategy == nil {
			logger.V(3).Infof("CreateRelease: No relative Strategy %s with Release.", option.StrategyName)
			return "", fmt.Errorf("No relative Strategy %s", option.StrategyName)
		}
		request.Strategyid = strategy.Strategyid
	}

	response, err := operator.Client.CreateRelease(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof(
			"CreateRelease: post new Release %s for App[%s]/Cfgset[%s]/Commit %s failed, %s",
			option.Name, option.AppName, option.CfgSetName, option.CommitID, err.Error(),
		)
		return "", err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof(
			"CreateStrategy: post new Release %s for App[%s]/Cfgset[%s]/Commit %s successfully, but reponse failed: %s",
			option.Name, option.AppName, option.CfgSetName, option.CommitID, response.ErrMsg,
		)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.Releaseid) == 0 {
		logger.V(3).Infof("CreateStrategy: BSCP system error, No ReleaseID response")
		return "", fmt.Errorf("Lost ReleaseID from configuraiotn platform")
	}
	return response.Releaseid, nil
}

//GetRelease get specified Release details information
//Args:
//	name: strategy name
//	ID: strategy ID
//return:
//	strategy: specified configset, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetRelease(cxt context.Context, ID string) (*common.Release, error) {
	//do not implemented
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return nil, err
	}
	if business == nil {
		logger.V(3).Infof("GetRelease: found no relative Business %s", operator.Business)
		return nil, fmt.Errorf("No relative Business %s", operator.Business)
	}

	return operator.innerGetReleaseByID(cxt, business.Bid, ID)
}

func (operator *AccessOperator) innerGetReleaseByID(cxt context.Context, bID, ID string) (*common.Release, error) {
	request := &accessserver.QueryReleaseReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       bID,
		Releaseid: ID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryRelease(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetRelease %s details failed, %s", ID, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetRelease %s: resource Not Found.", ID)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetRelease %s successfully, but response Err, %s", ID, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Release, nil
}

//ListReleaseByApp list all release information about specified App & configset
//return:
//	strategies: all strategies, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListReleaseByApp(cxt context.Context, appName, cfgsetName string) ([]*common.Release, error) {
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}

	configSet, err := operator.innergetConfigSet(cxt, business.Bid, app.Appid, cfgsetName)
	if err != nil {
		return nil, err
	}
	if configSet == nil {
		return nil, nil
	}

	request := &accessserver.QueryHistoryReleasesReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      business.Bid,
		Cfgsetid: configSet.Cfgsetid,
		//fix: list all release
		//Operator: operator.User,
		Index: operator.index,
		Limit: operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryHistoryReleases(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListReleaseByApp failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListReleaseByApp all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Releases, nil
}

//UpdateRelease update specified Release.
//Args:
//	option: updated information
//return:
//	error: error info if that happened
func (operator *AccessOperator) UpdateRelease(cxt context.Context, option *ReleaseOption) error {
	//business first
	business, _, err := getBusinessAndApp(operator, operator.Business, option.AppName)
	if err != nil {
		return err
	}
	request := &accessserver.UpdateReleaseReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       business.Bid,
		Releaseid: option.ReleaseID,
		Name:      option.Name,
		Operator:  operator.User,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.UpdateRelease(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("UpdateRelease %s failed, %s", option.Name, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("UpdateRelease %s successfully, but response Err, %s", option.ReleaseID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil

}

//ConfirmRelease make Release affective and publish to specified endpoints
//return:
//	err: if any error happened
func (operator *AccessOperator) ConfirmRelease(cxt context.Context, releaseID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("ConfirmRelease: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}

	//setting cancel commit request
	request := &accessserver.PublishReleaseReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       business.Bid,
		Releaseid: releaseID,
		Operator:  operator.User,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.PublishRelease(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ConfirmRelease %s failed, %s", releaseID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ConfirmRelease %s successfully, but response Err, %s", releaseID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}
