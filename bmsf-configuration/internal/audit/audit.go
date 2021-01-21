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

	pbcommon "bk-bscp/internal/protocol/common"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

var (
	// audit handler.
	auditHandler *Handler
)

// Handler audit handler.
type Handler struct {
	// audit info channel size.
	infoChanSize int

	// audit info channel timeout.
	infoChanTimeout time.Duration

	// datamanager client for audit handler.
	dataMgrCli pbdatamanager.DataManagerClient

	// datamanager client call timeout.
	dataMgrCliTimeout time.Duration

	// audit info chan for async create audit.
	auditInfoCh chan interface{}
}

// InitAuditHandler inits audit handler stuff.
func InitAuditHandler(infoChanSize int, infoChanTimeout time.Duration,
	dataMgrCli pbdatamanager.DataManagerClient, dataMgrCliTimeout time.Duration) {

	auditHandler = &Handler{
		infoChanSize:      infoChanSize,
		infoChanTimeout:   infoChanTimeout,
		dataMgrCli:        dataMgrCli,
		dataMgrCliTimeout: dataMgrCliTimeout,
		auditInfoCh:       make(chan interface{}, infoChanSize),
	}
	go auditHandler.loop()
}

func (h *Handler) loop() {
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
			logger.Error("unknown audit info type, %+v", info)
			continue
		}
	}
}

func (h *Handler) addAudit(info interface{}) {
	select {
	case h.auditInfoCh <- info:
	case <-time.After(h.infoChanTimeout):
		logger.Warn("add audit info to chan timeout, %+v", info)
	}
}

// CreateAudit creates a new audit log.
func (h *Handler) CreateAudit(request *pbdatamanager.CreateAuditReq) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.dataMgrCliTimeout)
	defer cancel()

	resp, err := h.dataMgrCli.CreateAudit(ctx, request)
	if err != nil {
		return err
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		return errors.New(resp.Message)
	}
	return nil
}

// Audit create audit record on target event.
func Audit(sourceType, opType int32, bizID, sourceID, operator, memo string) {
	audit := &pbdatamanager.CreateAuditReq{
		Seq:        common.Sequence(),
		SourceType: sourceType,
		OpType:     opType,
		BizId:      bizID,
		SourceId:   sourceID,
		Operator:   operator,
		Memo:       memo,
	}

	// add audit to async chan.
	go auditHandler.addAudit(audit)
}
