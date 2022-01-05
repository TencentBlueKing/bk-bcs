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

package service

import (
	"encoding/json"
	"errors"
	"time"

	sidecar "bk-bscp/cmd/atomic-services/bscp-bcs-sidecar/service"
	"bk-bscp/cmd/atomic-services/bscp-gse-plugin/modules/tunnel"
	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pbtunnelserver "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/internal/safeviper"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

const (
	// defaultPullProcattrListInterval is default pull procattr list interval.
	defaultPullProcattrListInterval = 60 * time.Second

	// defaultFirstPullProcattrListInterval is default first pull procattr list interval.
	defaultFirstPullProcattrListInterval = 5 * time.Second
)

// Sidecar is bscp bcs sidecar in local gse plugin.
type Sidecar struct {
	// configs handler.
	viper *safeviper.SafeViper

	gseTunnel *tunnel.Tunnel

	// reloader.
	reloader *sidecar.Reloader

	// app mods manager.
	appModMgr *sidecar.AppModManager

	// instance server.
	insServer *sidecar.InstanceServer
}

// NewSidecar creates new gse plugin local sidecar instance.
func NewSidecar(viper *safeviper.SafeViper, gseTunnel *tunnel.Tunnel) *Sidecar {
	return &Sidecar{viper: viper, gseTunnel: gseTunnel}
}

func (s *Sidecar) envName(bindKey string) string {
	return ENVPREFIX + "_" + bindKey
}

func (s *Sidecar) queryHostProcAttrList() ([]*pbcommon.ProcAttr, error) {
	procAttrs := []*pbcommon.ProcAttr{}

	index := 0

	for {
		req := &pbtunnelserver.GTCMDQueryHostProcAttrListReq{
			Seq:  common.Sequence(),
			Page: &pbcommon.Page{Start: int32(index), Limit: int32(database.BSCPQUERYLIMITLB)},
		}

		messageID := common.SequenceNum()

		resp, err := s.gseTunnel.QueryHostProcAttrList(messageID, req)
		if err != nil {
			return nil, err
		}
		if resp.Code != pbcommon.ErrCode_E_OK {
			return nil, errors.New(resp.Message)
		}
		procAttrs = append(procAttrs, resp.ProcAttrs...)

		if len(resp.ProcAttrs) < database.BSCPQUERYLIMITLB {
			break
		}
		index += len(resp.ProcAttrs)
	}

	return procAttrs, nil
}

// keep checking app mod in dynamic mode.
func (s *Sidecar) dynamicLoad() {
	isFirstPullSuccess := false

	for {
		if isFirstPullSuccess {
			time.Sleep(defaultPullProcattrListInterval)
		} else {
			time.Sleep(defaultFirstPullProcattrListInterval)
		}

		// query host proc attr list.
		procAttrs, err := s.queryHostProcAttrList()
		if err != nil {
			logger.Errorf("query host proc attrs list, %+v", err)
			continue
		}

		isSuccess := true
		appModInfos := []sidecar.AppModInfo{}

		for _, procAttr := range procAttrs {
			newMod := sidecar.AppModInfo{
				BizID:   procAttr.BizId,
				AppID:   procAttr.AppId,
				CloudID: procAttr.CloudId,
				Path:    procAttr.Path,
				Labels:  make(map[string]string),
			}

			if err := json.Unmarshal([]byte(procAttr.Labels), &newMod.Labels); err != nil {
				logger.Errorf("procattr config check, invalid appinfo labels, %+v", err)

				isSuccess = false
				break
			}
			appModInfos = append(appModInfos, newMod)
		}

		if !isSuccess {
			logger.Errorf("update host procattrs failed, try again later!")
			continue
		}

		appInfoModCfgVal, err := json.Marshal(&appModInfos)
		if err != nil {
			logger.Errorf("procattr config check, can't marshal appmods, %+v", err)
			continue
		}

		// update host procattrs.
		s.viper.Set("appmods", string(appInfoModCfgVal))
		isFirstPullSuccess = true
		logger.V(2).Infof("update host procattrs success, %+v", string(appInfoModCfgVal))
	}
}

// init configs reloader.
func (s *Sidecar) initReloader() {
	s.reloader = sidecar.NewReloader(s.viper)
	s.reloader.Init()
	logger.Info("Sidecar| init reloader success.")
}

// init app mods.
func (s *Sidecar) initAppMods() {
	// init dynamic app mod check.
	go s.dynamicLoad()

	// init app mod manager.
	s.appModMgr = sidecar.NewAppModManager(s.viper, s.reloader)
	s.appModMgr.Setup()
	logger.Info("Sidecar| init app mod manager setup success.")
}

// init instance server.
func (s *Sidecar) initInstanceServer() {
	if !s.viper.GetBool("instance.open") {
		return
	}

	// create instance server.
	s.insServer = sidecar.NewInstanceServer(s.viper,
		common.Endpoint(s.viper.GetString("instance.httpEndpoint.ip"), s.viper.GetInt("instance.httpEndpoint.port")),
		common.Endpoint(s.viper.GetString("instance.grpcEndpoint.ip"), s.viper.GetInt("instance.grpcEndpoint.port")),
		s.appModMgr, s.reloader)

	// init and run.
	if err := s.insServer.Init(); err != nil {
		logger.Warn("Sidecar| init instance server, %+v", err)
	} else {
		go s.insServer.Run()
	}
}

// Run runs the sidecar in local gse plugin.
func (s *Sidecar) Run() {
	// initialize reloader.
	s.initReloader()

	// initialize app mods.
	s.initAppMods()

	// initialize instance server.
	s.initInstanceServer()

	// run success.
	logger.Info("Sidecar running now.")
}
