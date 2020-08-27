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
	pb "bk-bscp/internal/protocol/accessserver"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/logger"
)

// CreateAction creates a multi commit object.
type CreateAction struct {
	viper    *viper.Viper
	buSvrCli pbbusinessserver.BusinessClient

	req  *pb.CreateMultiCommitReq
	resp *pb.CreateMultiCommitResp
}

// NewCreateAction creates new CreateAction.
func NewCreateAction(viper *viper.Viper, buSvrCli pbbusinessserver.BusinessClient,
	req *pb.CreateMultiCommitReq, resp *pb.CreateMultiCommitResp) *CreateAction {
	action := &CreateAction{viper: viper, buSvrCli: buSvrCli, req: req, resp: resp}

	action.resp.Seq = req.Seq
	action.resp.ErrCode = pbcommon.ErrCode_E_OK
	action.resp.ErrMsg = "OK"

	return action
}

// Err setup error code message in response and return the error.
func (act *CreateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	act.resp.ErrCode = errCode
	act.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *CreateAction) Input() error {
	if err := act.verify(); err != nil {
		return act.Err(pbcommon.ErrCode_E_AS_PARAMS_INVALID, err.Error())
	}
	return nil
}

// Output handles the output messages.
func (act *CreateAction) Output() error {
	// do nothing.
	return nil
}

func (act *CreateAction) verify() error {
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

	length = len(act.req.Operator)
	if length == 0 {
		return errors.New("invalid params, operator missing")
	}
	if length > database.BSCPNAMELENLIMIT {
		return errors.New("invalid params, operator too long")
	}

	if len(act.req.ReuseCommitid) > database.BSCPIDLENLIMIT {
		return errors.New("invalid params, reuse commitid too long")
	}

	if len(act.req.Metadatas) == 0 {
		return errors.New("invalid params, invalid metadatas")
	}

	for _, metadata := range act.req.Metadatas {
		length := len(metadata.Cfgsetid)
		if length == 0 {
			return errors.New("invalid params, cfgsetid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, cfgsetid too long")
		}

		if metadata.Configs == nil {
			metadata.Configs = []byte{}
		}

		if len(metadata.Configs) > database.BSCPCONFIGSSIZELIMIT {
			return errors.New("invalid params, configs content too big")
		}
		if len(metadata.Changes) > database.BSCPCHANGESSIZELIMIT {
			return errors.New("invalid params, configs changes too big")
		}

		if len(metadata.Templateid) > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, templateid too long")
		}
		if len(metadata.Template) > database.BSCPTPLSIZELIMIT {
			return errors.New("invalid params, template size too big")
		}
		if len(metadata.TemplateRule) > database.BSCPTPLRULESSIZELIMIT {
			return errors.New("invalid params, template rules too long")
		}

		if len(metadata.Configs) != 0 && len(metadata.Template) != 0 {
			return errors.New("invalid params, configs and template concurrence")
		}
		if len(metadata.Configs) != 0 && len(metadata.Templateid) != 0 {
			return errors.New("invalid params, configs and templateid concurrence")
		}
		if len(metadata.Template) != 0 && len(metadata.Templateid) != 0 {
			return errors.New("invalid params, template and templateid concurrence")
		}
		if len(metadata.Template) != 0 && len(metadata.TemplateRule) == 0 {
			return errors.New("invalid params, empty template rules")
		}
	}

	if len(act.req.Memo) > database.BSCPLONGSTRLENLIMIT {
		return errors.New("invalid params, memo too long")
	}
	return nil
}

func (act *CreateAction) create() (pbcommon.ErrCode, string) {
	r := &pbbusinessserver.CreateMultiCommitReq{
		Seq:           act.req.Seq,
		Bid:           act.req.Bid,
		Appid:         act.req.Appid,
		Operator:      act.req.Operator,
		Memo:          act.req.Memo,
		Metadatas:     act.req.Metadatas,
		ReuseCommitid: act.req.ReuseCommitid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("businessserver.calltimeoutLT"))
	defer cancel()

	logger.V(2).Infof("CreateMultiCommit[%d]| request to businessserver CreateMultiCommit, %+v", act.req.Seq, r)

	resp, err := act.buSvrCli.CreateMultiCommit(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_AS_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver CreateMultiCommit, %+v", err)
	}
	act.resp.MultiCommitid = resp.MultiCommitid
	act.resp.SuccCfgsets = resp.SuccCfgsets
	act.resp.FailCfgsets = resp.FailCfgsets

	return resp.ErrCode, resp.ErrMsg
}

// Do makes the workflows of this action base on input messages.
func (act *CreateAction) Do() error {
	// create commit.
	if errCode, errMsg := act.create(); errCode != pbcommon.ErrCode_E_OK {
		return act.Err(errCode, errMsg)
	}
	return nil
}
