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
	"io/ioutil"
	"path"

	"google.golang.org/grpc"

	global "bk-bscp/cmd/bscp-client/option"
	"bk-bscp/internal/protocol/accessserver"
	"bk-bscp/internal/protocol/common"
	"bk-bscp/internal/strategy"
	pkgcommon "bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

const (
	// enter operator is all
	AllOperator = "all"
)

// the current dir record file struct
type ConfigFile struct {
	State int32
}

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
//	appName: app name
//  cfgset: configSet
//return:
//	app: specified configset, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetConfigSet(cxt context.Context, appName string, cfgset *common.ConfigSet) (*common.ConfigSet, error) {
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	cfgset.Appid = app.Appid
	cfgset.Bid = business.Bid
	return operator.innergetConfigSet(cxt, cfgset)
}

func (operator *AccessOperator) GetConfigSetById(cxt context.Context, cfgset *common.ConfigSet) (*common.ConfigSet, error) {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return nil, err
	}
	cfgset.Bid = business.Bid
	return operator.innergetConfigSet(cxt, cfgset)
}

// Query configset by bid and appid
// Query conditions：cfgsetId、cfgsetName、cfgsetPath
func (operator *AccessOperator) QueryConfigSet(cxt context.Context, cfgset *common.ConfigSet) (*common.ConfigSet, error) {
	return operator.innergetConfigSet(cxt, cfgset)
}

func (operator *AccessOperator) innergetConfigSet(cxt context.Context, cfgset *common.ConfigSet) (*common.ConfigSet, error) {
	request := &accessserver.QueryConfigSetReq{
		Seq:      pkgcommon.Sequence(),
		Bid:      cfgset.Bid,
		Appid:    cfgset.Appid,
		Name:     cfgset.Name,
		Fpath:    cfgset.Fpath,
		Cfgsetid: cfgset.Cfgsetid,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryConfigSet(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetConfigSet %s failed, %s", request, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetConfigSet %s: resource Not Found.", request)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetConfigSet %s successfully, but response Err, %s", request, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.ConfigSet, nil
}

// Query Application by bid
// Query conditions：appid、appName
func (operator *AccessOperator) QueryApplication(cxt context.Context, app *common.App) (*common.App, error) {
	return operator.innergetApplication(cxt, app)
}

func (operator *AccessOperator) innergetApplication(cxt context.Context, app *common.App) (*common.App, error) {
	request := &accessserver.QueryAppReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   app.Bid,
		Appid: app.Appid,
		Name:  app.Name,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryApp(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetAPP %s failed, %s", request, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetAPP %s: resource Not Found.", request)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetAPP %s successfully, but response Err, %s", request, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.App, nil
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
type CreateMultiCommitOption struct {
	//AppName for app relation
	AppName string
	// Memo
	Memo string
	// RecordConfigFiles
	RecordConfigFiles map[string]ConfigFile
}

type CommitMetadataJson struct {
	Cfgset     string
	ConfigFile string
}

//CreateCommit multicreate commit for ConfigSet
//return:
//	appID: when creating successfully, system will response ID
//	error: any error if happened
func (operator *AccessOperator) CreateMultiCommit(cxt context.Context, option *CreateMultiCommitOption) (string, error) {
	if option == nil {
		return "", fmt.Errorf("Lost create Commit info")
	}

	// read commitMetadataJson json
	commitMetadatas := []*common.CommitMetadata{}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, option.AppName)
	if err != nil {
		return "", err
	}
	for recordMapKey, recordMapValue := range option.RecordConfigFiles {
		var cfgsetId string
		if recordMapValue.State != 1 {
			continue
		}
		// query cfgset
		cfgsetPath, cfgsetName := path.Split(recordMapKey)
		cfgsetPath = path.Clean("/" + cfgsetPath)
		query := &common.ConfigSet{
			Fpath: cfgsetPath,
			Name:  cfgsetName,
		}
		configSet, err := operator.GetConfigSet(cxt, option.AppName, query)
		if err != nil {
			return "", err
		}
		if configSet == nil {
			request := &accessserver.CreateConfigSetReq{
				Seq:     pkgcommon.Sequence(),
				Bid:     business.Bid,
				Appid:   app.Appid,
				Name:    cfgsetName,
				Creator: operator.User,
				Fpath:   cfgsetPath,
				Memo:    "",
			}
			createCfgsetId, err := operator.CreateConfigSet(cxt, request)
			if err != nil {
				return "", fmt.Errorf("create %s configset fail", recordMapKey)
			}
			if len(createCfgsetId) == 0 {
				return "", fmt.Errorf("create %s configset fail", recordMapKey)
			}
			cfgsetId = createCfgsetId
		} else {
			cfgsetId = configSet.Cfgsetid
		}

		cfgContent, _ := ioutil.ReadFile(recordMapKey)
		commitMetadatas = append(commitMetadatas, &common.CommitMetadata{
			Cfgsetid: cfgsetId,
			Configs:  cfgContent,
		})
	}
	request := &accessserver.CreateMultiCommitReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       business.Bid,
		Appid:     app.Appid,
		Operator:  operator.User,
		Memo:      option.Memo,
		Metadatas: commitMetadatas,
	}
	response, err := operator.Client.CreateMultiCommit(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CreateMultiCommit: post new MultiCommit for ConfigSet [%s] failed, %s", commitMetadatas, err.Error())
		return "", err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof(
			"CreateMultiCommit: post new new MultiCommit for ConfigSet [%s] successfully, but response failed: %s",
			commitMetadatas, response.ErrMsg,
		)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.MultiCommitid) == 0 {
		logger.V(3).Infof("CreateMultiConfigSet: BSCP system error, No MultiCommitID response")
		return "", fmt.Errorf("Lost MultiCommitID from configuraiotn platform")
	}
	return response.MultiCommitid, nil
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

func (operator *AccessOperator) GetEffectedInstanceList(cxt context.Context, releaseid string) ([]*common.AppInstance, error) {
	release, err := operator.GetRelease(cxt, releaseid)
	if err != nil {
		return nil, err
	}
	request := &accessserver.QueryEffectedAppInstancesReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       release.Bid,
		Cfgsetid:  release.Cfgsetid,
		Releaseid: releaseid,
		Index:     operator.index,
		Limit:     operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryEffectedAppInstances(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetEffectedInstanceList %s details failed, %s", releaseid, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetEffectedInstanceList %s: resource Not Found.", releaseid)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetEffectedInstanceList %s successfully, but response Err, %s", releaseid, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Instances, nil
}

//GetMultiCommit get specified Commit details information
//Args:
//	name: app name
//return:
//	app: specified configset, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetMultiCommit(cxt context.Context, multiCommitID string) (*common.MultiCommit, []*common.CommitMetadata, error) {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return nil, nil, err
	}
	if business == nil {
		logger.V(3).Infof("GetMultiCommit: found no relative Business %s", operator.Business)
		return nil, nil, fmt.Errorf("No relative Business %s", operator.Business)
	}

	request := &accessserver.QueryMultiCommitReq{
		Seq:           pkgcommon.Sequence(),
		Bid:           business.Bid,
		MultiCommitid: multiCommitID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryMultiCommit(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetMultiCommit %s details failed, %s", multiCommitID, err.Error())
		return nil, nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetMultiCommit %s: resource Not Found.", multiCommitID)
		return nil, nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetMultiCommit %s successfully, but response Err, %s", multiCommitID, response.ErrMsg)
		return nil, nil, fmt.Errorf("%s", response.ErrMsg)
	}

	return response.MultiCommit, response.Metadatas, nil
}

//CancelMultiCommit By multiCommitId
func (operator *AccessOperator) CancelMultiCommitById(cxt context.Context, multiCommitID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("CancelMultiCommit: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}
	request := &accessserver.CancelMultiCommitReq{
		Seq:           pkgcommon.Sequence(),
		Bid:           business.Bid,
		Operator:      operator.User,
		MultiCommitid: multiCommitID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.CancelMultiCommit(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CancelMultiCommit %s failed, %s", multiCommitID, err.Error())
		return err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("CancelMultiCommit %s: resource Not Found.", multiCommitID)
		return nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("CancelMultiCommit %s successfully, but response Err, %s", multiCommitID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//CancelMultiRelease By multiRelease
func (operator *AccessOperator) CancelMultiReleaseById(cxt context.Context, multiReleaseID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("CancelMultiReleaseById: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}
	request := &accessserver.CancelMultiReleaseReq{
		Seq:            pkgcommon.Sequence(),
		Bid:            business.Bid,
		Operator:       operator.User,
		MultiReleaseid: multiReleaseID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.CancelMultiRelease(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("CancelMultiReleaseById %s failed, %s", multiReleaseID, err.Error())
		return err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("CancelMultiReleaseById %s: resource Not Found.", multiReleaseID)
		return nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("CancelMultiReleaseById %s successfully, but response Err, %s", multiReleaseID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//RollbackMultiReleaseById By multiReleaseId
func (operator *AccessOperator) RollbackMultiReleaseById(cxt context.Context, multiReleaseID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("RollbackMultiReleaseById: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}
	request := &accessserver.RollbackMultiReleaseReq{
		Seq:            pkgcommon.Sequence(),
		Bid:            business.Bid,
		Operator:       operator.User,
		MultiReleaseid: multiReleaseID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.RollbackMultiRelease(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("RollbackMultiReleaseById %s failed, %s", multiReleaseID, err.Error())
		return err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("RollbackMultiReleaseById %s: resource Not Found.", multiReleaseID)
		return nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("RollbackMultiReleaseById %s successfully, but response Err, %s", multiReleaseID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//ReloadMultiReleaseById By multiReleaseId
func (operator *AccessOperator) ReloadMultiReleaseById(cxt context.Context, multiReleaseID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("ReloadMultiReleaseById: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}
	request := &accessserver.ReloadReq{
		Seq:            pkgcommon.Sequence(),
		Bid:            business.Bid,
		Operator:       operator.User,
		MultiReleaseid: multiReleaseID,
		Rollback:       true,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.Reload(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ReloadMultiReleaseById %s failed, %s", multiReleaseID, err.Error())
		return err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("ReloadMultiReleaseById %s: resource Not Found.", multiReleaseID)
		return nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ReloadMultiReleaseById %s successfully, but response Err, %s", multiReleaseID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}

//ListCommitsAllByID list all Commits of ConfigSet information
//return:
//	Commits: all Commit, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListMultiCommitsAllByAppID(cxt context.Context, bID, appID string) ([]*common.MultiCommit, error) {
	user := ""
	if operator.User != AllOperator {
		user = operator.User
	}
	request := &accessserver.QueryHistoryMultiCommitsReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       bID,
		Appid:     appID,
		QueryType: 0,
		Operator:  user,
		Index:     operator.index,
		Limit:     operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryHistoryMultiCommits(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListMultiCommits failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListMultiCommits all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.MultiCommits, nil
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
func (operator *AccessOperator) ConfirmMultiCommit(cxt context.Context, multiCommitID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("ConfirmMultiCommit: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}

	//setting cancel commit request
	request := &accessserver.ConfirmMultiCommitReq{
		Seq:           pkgcommon.Sequence(),
		Bid:           business.Bid,
		MultiCommitid: multiCommitID,
		Operator:      operator.User,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.ConfirmMultiCommit(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ConfirmMultiCommit %s failed, %s", multiCommitID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ConfirmMultiCommit %s successfully, but response Err, %s", multiCommitID, response.ErrMsg)
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
	//Memo strategy details, it's json data
	Memo string
}

// strategy is json file struct of create strategies
type CreateStrategy struct {
	App      string
	Clusters []string
	Zones    []string
	Dcs      []string
	IPs      []string
	// Labels is instance labels in strategy which control "OR".
	Labels map[string]string
	// LabelsAnd is instance labels in strategy which control "AND".
	LabelsAnd map[string]string
}

//CreateStrategy create strategy for release
//return:
//	strategyID: when creating successfully, system will response ID
//	error: any error if happened
func (operator *AccessOperator) CreateStrategy(cxt context.Context, option *StrategyOption) (string, error) {
	if option == nil {
		return "", fmt.Errorf("lost create Strategy info")
	}
	// query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, option.AppName)
	if err != nil {
		return "", err
	}

	// get strategy
	createStrategyJson := &CreateStrategy{}
	if err := json.Unmarshal([]byte(option.Content), createStrategyJson); err != nil {
		return "", err
	}
	strategies, err := operator.getStrategyFromNameToId(cxt, createStrategyJson)
	if err != nil {
		return "", err
	}

	// judge flag appName == create strategy json file appName
	if strategies.Appid != app.Appid {
		return "", fmt.Errorf("the application of the strategy does not match the application in the json file")
	}

	logger.V(3).Infof("CreateStrategy: post new Strategy %s for Business [%s] : %s", option.Name, operator.Business, strategies)
	request := &accessserver.CreateStrategyReq{
		Seq:        pkgcommon.Sequence(),
		Bid:        business.Bid,
		Appid:      app.Appid,
		Name:       option.Name,
		Memo:       option.Memo,
		Clusterids: strategies.Clusterids,
		Zoneids:    strategies.Zoneids,
		Dcs:        strategies.Dcs,
		IPs:        strategies.IPs,
		Labels:     strategies.Labels,
		LabelsAnd:  strategies.LabelsAnd,
		Creator:    operator.User,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
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

//getStrategyFromNameToId to to convert strategy content's resource id to name
//Args:
// 	createStrategy:  create strategy json file content
//return:
// 	strategy: get strategy struct from CreateStrategy
// 	error: error info if that happened
func (operator *AccessOperator) getStrategyFromNameToId(cxt context.Context, createStrategy *CreateStrategy) (*strategy.Strategy, error) {
	strategies := &strategy.Strategy{}

	// query app
	app, err := operator.GetApp(cxt, createStrategy.App)
	if err != nil {
		return nil, fmt.Errorf("json file error ( %s )", err)
	}
	if app == nil {
		return nil, fmt.Errorf("app [%s] not found from json file", createStrategy.App)
	}
	strategies.Appid = app.Appid

	// query cluster
	for _, clusterName := range createStrategy.Clusters {
		cluster, err := operator.GetLogicCluster(cxt, app.Name, clusterName)
		if err != nil {
			return nil, fmt.Errorf("json file error ( %s )", err)
		}
		if cluster == nil {
			return nil, fmt.Errorf("cluster [%s] not found from json file", clusterName)
		}
		strategies.Clusterids = append(strategies.Clusterids, cluster.Clusterid)
	}

	// query zone
	for _, zoneName := range createStrategy.Zones {
		zone, err := operator.GetZoneByName(cxt, app.Name, zoneName)
		if err != nil {
			return nil, fmt.Errorf("json file error ( %s )", err)
		}
		if zone == nil {
			return nil, fmt.Errorf("zone [%s] not found from json file", zoneName)
		}
		strategies.Zoneids = append(strategies.Zoneids, zone.Zoneid)
	}

	// labels labelsAnd
	strategies.Dcs = createStrategy.Dcs
	strategies.IPs = createStrategy.IPs
	strategies.Labels = createStrategy.Labels
	strategies.LabelsAnd = createStrategy.LabelsAnd

	return strategies, nil
}

//GetStrategyJsonFromIdToName to convert strategy content's resource name to id
//Args:
// 	strategyJson: query strategy struct's content
//return:
// 	createStrategy: struct that save convert resource name to id
// 	error: error info if that happened
func (operator *AccessOperator) GetStrategyFromIdToName(cxt context.Context, strategy *strategy.Strategy) ([]byte, error) {
	createStrategy := &CreateStrategy{}

	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return nil, err
	}
	if business == nil {
		return nil, fmt.Errorf("problem with the content of the strategy: business [%s] not found ", operator.Business)
	}

	// query app
	app, err := operator.GetAppByAppID(cxt, business.Bid, strategy.Appid)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, fmt.Errorf("problem with the content of the strategy: app [%s] not found", strategy.Appid)
	}
	createStrategy.App = app.Name

	// query cluster
	for _, clusterID := range strategy.Clusterids {
		cluster, err := operator.GetClusterAllByID(cxt, business.Bid, app.Appid, clusterID)
		if err != nil {
			return nil, err
		}
		if cluster == nil {
			return nil, fmt.Errorf("problem with the content of the strategy: cluster [%s] not found", clusterID)
		}
		createStrategy.Clusters = append(createStrategy.Clusters, cluster.Name)
	}

	// query zone
	for _, zoneID := range strategy.Zoneids {
		zone, err := operator.GetZoneAllByID(cxt, business.Bid, app.Appid, zoneID)
		if err != nil {
			return nil, err
		}
		if zone == nil {
			return nil, fmt.Errorf("problem with the content of the strategy: zone [%s] not found", zoneID)
		}
		createStrategy.Zones = append(createStrategy.Zones, zone.Name)
	}

	// labels labelsAnd
	createStrategy.Dcs = strategy.Dcs
	createStrategy.IPs = strategy.IPs
	createStrategy.Labels = strategy.Labels
	createStrategy.LabelsAnd = strategy.LabelsAnd

	strategyJson, err := json.Marshal(createStrategy)
	if err != nil {
		return nil, err
	}
	return strategyJson, nil
}

//GetStrategy get specified Strategy details information
//Args:
//	name: strategy name
//	ID: strategy ID
//return:
//	strategy: specified configset, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) GetStrategyByName(cxt context.Context, appName, name string) (*common.Strategy, error) {
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	return operator.innerGetStrategy(cxt, business.Bid, app.Appid, name, "")
}

func (operator *AccessOperator) innerGetStrategyByID(cxt context.Context, bID, appID, name string) (*common.Strategy, error) {
	strategy, err := operator.innerGetStrategy(cxt, bID, appID, name, "")
	return strategy, err
}

func (operator *AccessOperator) innerGetStrategy(cxt context.Context, bID, appID, name, strategyid string) (*common.Strategy, error) {
	request := &accessserver.QueryStrategyReq{
		Seq:        pkgcommon.Sequence(),
		Bid:        bID,
		Appid:      appID,
		Name:       name,
		Strategyid: strategyid,
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

func (operator *AccessOperator) GetStrategyById(cxt context.Context, strategyid string) (*common.Strategy, error) {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		logger.V(3).Infof("Query relative business %s failed, %s", operator.Business, err.Error())
		return nil, err
	}
	if business == nil {
		return nil, fmt.Errorf("No relative Business %s Resource", operator.Business)
	}
	return operator.innerGetStrategy(cxt, business.Bid, "", "", strategyid)
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
	MultiCommitID string
	//Memo release relative to Commit
	Memo string
}

//CreateMultiRelease create multi-release for specified Commit
//return:
//	multi-releaseID: when creating successfully, system will response ID
//	error: any error if happened
func (operator *AccessOperator) CreateMultiRelease(cxt context.Context, option *ReleaseOption) (string, error) {
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

	request := &accessserver.CreateMultiReleaseReq{
		Seq:           pkgcommon.Sequence(),
		Bid:           business.Bid,
		Appid:         app.Appid,
		Name:          option.Name,
		MultiCommitid: option.MultiCommitID,
		Creator:       operator.User,
		Memo:          option.Memo,
	}

	// check strategy for this release.
	if len(option.StrategyName) != 0 {
		strategy, styerr := operator.innerGetStrategyByID(cxt, business.Bid, app.Appid, option.StrategyName)
		if styerr != nil {
			return "", styerr
		}

		if strategy == nil {
			logger.V(3).Infof("CreateMultiRelease: No relative Strategy %s with Release.", option.StrategyName)
			return "", fmt.Errorf("No relative Strategy[%s] under application[%s]", option.StrategyName, app.Name)
		}
		request.Strategyid = strategy.Strategyid
	}

	response, err := operator.Client.CreateMultiRelease(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof(
			"CreateMultiRelease: post new Release %s for App[%s]/Cfgset[%s]/Commit %s failed, %s",
			option.Name, option.AppName, option.CfgSetName, option.MultiCommitID, err.Error(),
		)
		return "", err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof(
			"CreateMultiRelease: post new MultiRelease %s for App[%s]/Cfgset[%s]/MultiCommit %s successfully, but reponse failed: %s",
			option.Name, option.AppName, option.CfgSetName, option.MultiCommitID, response.ErrMsg,
		)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}
	if len(response.MultiReleaseid) == 0 {
		logger.V(3).Infof("CreateMultiRelease: BSCP system error, No ReleaseID response")
		return "", fmt.Errorf("Lost ReleaseID from configuraiotn platform")
	}
	return response.MultiReleaseid, nil
}

//GetMultiRelease get specified MultiRelease details information
//Args:
//	name: multi-releaseID
//return:
//	multi-release
func (operator *AccessOperator) GetMultiRelease(cxt context.Context, ID string) (*common.MultiRelease, []*common.ReleaseMetadata, error) {
	//do not implemented
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return nil, nil, err
	}
	if business == nil {
		logger.V(3).Infof("GetMultiRelease: found no relative Business %s", operator.Business)
		return nil, nil, fmt.Errorf("No relative Business %s", operator.Business)
	}

	return operator.innerGetMultiReleaseByID(cxt, business.Bid, ID)
}

func (operator *AccessOperator) innerGetMultiReleaseByID(cxt context.Context, bID, ID string) (*common.MultiRelease, []*common.ReleaseMetadata, error) {
	request := &accessserver.QueryMultiReleaseReq{
		Seq:            pkgcommon.Sequence(),
		Bid:            bID,
		MultiReleaseid: ID,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryMultiRelease(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("GetMultiRelease %s details failed, %s", ID, err.Error())
		return nil, nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetMultiRelease %s: resource Not Found.", ID)
		return nil, nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetMultiRelease %s successfully, but response Err, %s", ID, response.ErrMsg)
		return nil, nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.MultiRelease, response.Metadatas, nil
}

//ListMultiReleaseByApp list all release information about specified App
//return:
//	strategies: all strategies, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListMultiReleaseByApp(cxt context.Context, appName string, queryType int32) ([]*common.MultiRelease, error) {
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	user := ""
	if operator.User != AllOperator {
		user = operator.User
	}
	request := &accessserver.QueryHistoryMultiReleasesReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       business.Bid,
		Appid:     app.Appid,
		QueryType: queryType,
		Operator:  user,
		Index:     operator.index,
		Limit:     operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryHistoryMultiReleases(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListMultiReleaseByApp failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListMultiReleaseByApp all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.MultiReleases, nil
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

//ConfirmMultiRelease make Release affective and publish to specified endpoints
//return:
//	err: if any error happened
func (operator *AccessOperator) ConfirmMultiRelease(cxt context.Context, multiReleaseID string) error {
	//query business and app first
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		logger.V(3).Infof("ConfirmMultiRelease: found no relative Business %s", operator.Business)
		return fmt.Errorf("No relative Business %s", operator.Business)
	}

	//setting cancel commit request
	request := &accessserver.PublishMultiReleaseReq{
		Seq:            pkgcommon.Sequence(),
		Bid:            business.Bid,
		MultiReleaseid: multiReleaseID,
		Operator:       operator.User,
	}
	grpcOptions := []grpc.CallOption{}
	response, err := operator.Client.PublishMultiRelease(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ConfirmMultiRelease %s failed, %s", multiReleaseID, err.Error())
		return err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ConfirmMultiRelease %s successfully, but response Err, %s", multiReleaseID, response.ErrMsg)
		return fmt.Errorf("%s", response.ErrMsg)
	}
	return nil
}
