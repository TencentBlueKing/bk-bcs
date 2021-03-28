/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"fmt"
	"net/http"

	pbapiserver "bk-bscp/internal/protocol/apiserver"
	pbauthserver "bk-bscp/internal/protocol/authserver"
	pbcommon "bk-bscp/internal/protocol/common"
	pbconfigserver "bk-bscp/internal/protocol/configserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	pbgsecontroller "bk-bscp/internal/protocol/gse-controller"
	pbtemplateserver "bk-bscp/internal/protocol/templateserver"
	pbtunnelserver "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/json"
)

func (s *APIServer) healthzConfigServer(seq string) (*pbcommon.ModuleHealthzInfo, pbcommon.ErrCode, string) {
	r := &pbconfigserver.HealthzReq{Seq: seq}
	resp, err := s.configSvrCli.Healthz(context.Background(), r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_API_SYSTEM_UNKNOWN, err.Error()
	}
	return resp.Data, resp.Code, resp.Message
}

func (s *APIServer) healthzTemplateServer(seq string) (*pbcommon.ModuleHealthzInfo, pbcommon.ErrCode, string) {
	r := &pbtemplateserver.HealthzReq{Seq: seq}
	resp, err := s.templateSvrCli.Healthz(context.Background(), r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_API_SYSTEM_UNKNOWN, err.Error()
	}
	return resp.Data, resp.Code, resp.Message
}

func (s *APIServer) healthzAuthServer(seq string) (*pbcommon.ModuleHealthzInfo, pbcommon.ErrCode, string) {
	r := &pbauthserver.HealthzReq{Seq: seq}
	resp, err := s.authSvrCli.Healthz(context.Background(), r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_API_SYSTEM_UNKNOWN, err.Error()
	}
	return resp.Data, resp.Code, resp.Message
}

func (s *APIServer) healthzGSEController(seq string) (*pbcommon.ModuleHealthzInfo, pbcommon.ErrCode, string) {
	r := &pbgsecontroller.HealthzReq{Seq: seq}
	resp, err := s.gseControllerCli.Healthz(context.Background(), r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_API_SYSTEM_UNKNOWN, err.Error()
	}
	return resp.Data, resp.Code, resp.Message
}

func (s *APIServer) healthzTunnelServer(seq string) (*pbcommon.ModuleHealthzInfo, pbcommon.ErrCode, string) {
	r := &pbtunnelserver.HealthzReq{Seq: seq}
	resp, err := s.tunnelSvrCli.Healthz(context.Background(), r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_API_SYSTEM_UNKNOWN, err.Error()
	}
	return resp.Data, resp.Code, resp.Message
}

func (s *APIServer) healthzDataManager(seq string) (*pbcommon.ModuleHealthzInfo, pbcommon.ErrCode, string) {
	r := &pbdatamanager.HealthzReq{Seq: seq}
	resp, err := s.dataMgrCli.Healthz(context.Background(), r)
	if err != nil {
		return nil, pbcommon.ErrCode_E_API_SYSTEM_UNKNOWN, err.Error()
	}
	return resp.Data, resp.Code, resp.Message
}

// healthz returns healthz info include module version.
func (s *APIServer) healthz(w http.ResponseWriter, r *http.Request) error {
	// #lizard forgives
	seq := common.Sequence()

	response := &pbapiserver.HealthzResponse{Result: true, Code: pbcommon.ErrCode_E_OK, Message: "OK"}
	response.Data = &pbapiserver.HealthzResponse_RespData{IsHealthy: true, Modules: []*pbcommon.ModuleHealthzInfo{}}

	// configserver.
	healthInfo, errCode, errMsg := s.healthzConfigServer(seq)
	if errCode != pbcommon.ErrCode_E_OK {
		healthInfo = &pbcommon.ModuleHealthzInfo{
			Module:    "bk-bscp-configserver",
			IsHealthy: false,
			Message:   fmt.Sprintf("healthz config server failed, %s", errMsg),
		}
	}
	if !healthInfo.IsHealthy {
		response.Data.IsHealthy = false
	}
	response.Data.Modules = append(response.Data.Modules, healthInfo)

	// templateserver.
	healthInfo, errCode, errMsg = s.healthzTemplateServer(seq)
	if errCode != pbcommon.ErrCode_E_OK {
		healthInfo = &pbcommon.ModuleHealthzInfo{
			Module:    "bk-bscp-templateserver",
			IsHealthy: false,
			Message:   fmt.Sprintf("healthz template server failed, %s", errMsg),
		}
	}
	if !healthInfo.IsHealthy {
		response.Data.IsHealthy = false
	}
	response.Data.Modules = append(response.Data.Modules, healthInfo)

	// authserver.
	healthInfo, errCode, errMsg = s.healthzAuthServer(seq)
	if errCode != pbcommon.ErrCode_E_OK {
		healthInfo = &pbcommon.ModuleHealthzInfo{
			Module:    "bk-bscp-authserver",
			IsHealthy: false,
			Message:   fmt.Sprintf("healthz auth server failed, %s", errMsg),
		}
	}
	if !healthInfo.IsHealthy {
		response.Data.IsHealthy = false
	}
	response.Data.Modules = append(response.Data.Modules, healthInfo)

	// gse controller.
	healthInfo, errCode, errMsg = s.healthzGSEController(seq)
	if errCode != pbcommon.ErrCode_E_OK {
		healthInfo = &pbcommon.ModuleHealthzInfo{
			Module:    "bk-bscp-gse-controller",
			IsHealthy: false,
			Message:   fmt.Sprintf("healthz gsecontroller failed, %s", errMsg),
		}
	}
	if !healthInfo.IsHealthy {
		response.Data.IsHealthy = false
	}
	response.Data.Modules = append(response.Data.Modules, healthInfo)

	// tunnelserver.
	healthInfo, errCode, errMsg = s.healthzTunnelServer(seq)
	if errCode != pbcommon.ErrCode_E_OK {
		healthInfo = &pbcommon.ModuleHealthzInfo{
			Module:    "bk-bscp-tunnelserver",
			IsHealthy: false,
			Message:   fmt.Sprintf("healthz tunnel server failed, %s", errMsg),
		}
	}
	if !healthInfo.IsHealthy {
		response.Data.IsHealthy = false
	}
	response.Data.Modules = append(response.Data.Modules, healthInfo)

	// datamanager.
	healthInfo, errCode, errMsg = s.healthzDataManager(seq)
	if errCode != pbcommon.ErrCode_E_OK {
		healthInfo = &pbcommon.ModuleHealthzInfo{
			Module:    "bk-bscp-datamanager",
			IsHealthy: false,
			Message:   fmt.Sprintf("healthz data manager failed, %s", errMsg),
		}
	}
	if !healthInfo.IsHealthy {
		response.Data.IsHealthy = false
	}
	response.Data.Modules = append(response.Data.Modules, healthInfo)

	// build response.
	data, err := json.MarshalPB(response)
	if err != nil {
		return fmt.Errorf("marshal healthz response failed, %+v", err)
	}
	fmt.Fprintf(w, data)

	return nil
}
