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

package audit

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/viper"

	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

var (
	// audit handler.
	auditHandler *AuditHandler
)

// AuditHandler audit handler.
type AuditHandler struct {
	// viper object for audit handler.
	viper *viper.Viper

	// datamanager client for audit handler.
	dataMgrCli pbdatamanager.DataManagerClient

	// audit info chan for async create audit.
	auditInfoCh chan interface{}
}

// InitAuditHandler inits audit handler stuff.
func InitAuditHandler(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient) {
	auditHandler = &AuditHandler{
		viper:       viper,
		dataMgrCli:  dataMgrCli,
		auditInfoCh: make(chan interface{}, viper.GetDuration("audit.infoChanSize")),
	}
	go auditHandler.loop()
}

func (h *AuditHandler) loop() {
	for {
		info := <-h.auditInfoCh

		// TODO batch mode.

		switch info.(type) {
		case *pbdatamanager.CreateAuditReq:
			request := info.(*pbdatamanager.CreateAuditReq)

			if err := h.CreateAudit(request); err != nil {
				logger.Warn("async create audit record failed, %+v, %+v", request, err)
				continue
			}

		default:
			logger.Error("unknow audit info type, %+v", info)
			continue
		}
	}
}

func (h *AuditHandler) addAudit(info interface{}) {
	select {
	case h.auditInfoCh <- info:
	case <-time.After(h.viper.GetDuration("audit.infoChanTimeout")):
		logger.Warn("add audit info to chan timeout, %+v", info)
	}
}

func (h *AuditHandler) CreateAudit(request *pbdatamanager.CreateAuditReq) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	resp, err := h.dataMgrCli.CreateAudit(ctx, request)
	if err != nil {
		return err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return errors.New(resp.ErrMsg)
	}
	return nil
}

// Audit create audit record on target event.
func Audit(sourceType, opType int32, bid, sourceid, operator, memo string) {
	audit := &pbdatamanager.CreateAuditReq{
		Seq:        common.Sequence(),
		SourceType: sourceType,
		OpType:     opType,
		Bid:        bid,
		Sourceid:   sourceid,
		Operator:   operator,
		Memo:       memo,
	}

	// add audit to async chan.
	go auditHandler.addAudit(audit)
}
