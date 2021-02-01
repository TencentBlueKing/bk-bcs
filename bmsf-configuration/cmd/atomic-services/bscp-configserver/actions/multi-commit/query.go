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

package multicommit

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/kit"
	"bk-bscp/pkg/logger"
)

// QueryAction query target multi commit object.
type QueryAction struct {
	kit        kit.Kit
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	req  *pb.QueryMultiCommitReq
	resp *pb.QueryMultiCommitResp

	multiCommit *pbcommon.MultiCommit
	commitIDs   []string
	metadatas   []*pbcommon.CommitMetadata
}

// NewQueryAction creates new QueryAction.
func NewQueryAction(kit kit.Kit, viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	req *pb.QueryMultiCommitReq, resp *pb.QueryMultiCommitResp) *QueryAction {
	action := &QueryAction{kit: kit, viper: viper, dataMgrCli: dataMgrCli, req: req, resp: resp}

	action.resp.Result = true
	action.resp.Code = pbcommon.ErrCode_E_OK
	action.resp.Message = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *QueryAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	if errCode != pbcommon.ErrCode_E_OK {
		act.resp.Result = false
	}
	act.resp.Code = errCode
	act.resp.Message = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *QueryAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_CS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *QueryAction) Output() error {
	act.resp.Data = &pb.QueryMultiCommitResp_RespData{MultiCommit: act.multiCommit, Metadatas: act.metadatas}
	return nil
}

func (act *QueryAction) verify() error {
	var err error

	if err = common.ValidateString("biz_id", act.req.BizId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	if err = common.ValidateString("multi_commit_id", act.req.MultiCommitId,
		database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
		return err
	}
	return nil
}

func (act *QueryAction) queryMultiCommit() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiCommitReq{
		Seq:           act.kit.Rid,
		BizId:         act.req.BizId,
		MultiCommitId: act.req.MultiCommitId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("QueryMultiCommit[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryMultiCommit, %+v", err)
	}
	act.multiCommit = resp.Data

	return resp.Code, resp.Message
}

func (act *QueryAction) querySubCommitList() (pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryMultiCommitSubListReq{
		Seq:           act.kit.Rid,
		BizId:         act.req.BizId,
		MultiCommitId: act.req.MultiCommitId,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("QueryMultiCommit[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryMultiCommitSubList(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryMultiCommitSubList, %+v", err)
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return resp.Code, resp.Message
	}
	act.commitIDs = resp.Data.CommitIds

	return pbcommon.ErrCode_E_OK, ""
}

func (act *QueryAction) queryCommit(commitID string) (*pbcommon.Commit, pbcommon.ErrCode, string) {
	r := &pbdatamanager.QueryCommitReq{
		Seq:      act.kit.Rid,
		BizId:    act.req.BizId,
		CommitId: commitID,
	}

	ctx, cancel := context.WithTimeout(act.kit.Ctx, act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("QueryMultiCommit[%s]| request to datamanager, %+v", r.Seq, r)

	resp, err := act.dataMgrCli.QueryCommit(ctx, r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_CS_SYSTEM_UNKNOWN, fmt.Sprintf("request to datamanager QueryCommit, %+v", err)
	}
	return resp.Data, resp.Code, resp.Message
}

func (act *QueryAction) queryMetadatas() (pbcommon.ErrCode, string) {
	for _, commitID := range act.commitIDs {
		commit, errCode, errMsg := act.queryCommit(commitID)
		if errCode != pbcommon.ErrCode_E_OK {
			return errCode, errMsg
		}

		act.metadatas = append(act.metadatas, &pbcommon.CommitMetadata{
			CfgId:      commit.CfgId,
			CommitMode: commit.CommitMode,
			CommitId:   commit.CommitId,
		})
	}
	return pbcommon.ErrCode_E_OK, ""
}

// Do makes the workflows of this action base on input messages.
func (act *QueryAction) Do() error {
	// query multi commit.
	if errCode, errMsg := act.queryMultiCommit(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query multi commit sub commit list.
	if errCode, errMsg := act.querySubCommitList(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}

	// query sub commit metadatas.
	if errCode, errMsg := act.queryMetadatas(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
