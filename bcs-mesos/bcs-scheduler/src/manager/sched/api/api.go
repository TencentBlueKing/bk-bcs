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

package api

import (
	"bk-bcs/bcs-common/common"
	comm "bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	bhttp "bk-bcs/bcs-common/common/http"
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/scheduler"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"encoding/json"
	"errors"
	"github.com/emicklei/go-restful"
	"strconv"
	"strings"
	"time"
)

func (r *Router) queryAgentSettingList(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv query agentsettinglist request")

	var IPs []string
	if req.QueryParameter("ips") != "" {
		IPs = strings.Split(req.QueryParameter("ips"), ",")
	}
	settingList, errcode, err := r.backend.QueryAgentSettingList(IPs)
	if err != nil {
		blog.Error("fail to query agentsettinglist, err:%s", err.Error())
		data := createResponeDataV2(errcode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "", settingList)
	resp.Write([]byte(data))
	blog.Info("query agentsettinglist finish")

	return
}

func (r *Router) deleteAgentSettingList(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv delete agentsettinglist request")

	var IPs []string
	if req.QueryParameter("ips") != "" {
		IPs = strings.Split(req.QueryParameter("ips"), ",")
	}
	errcode, err := r.backend.DeleteAgentSettingList(IPs)
	if err != nil {
		blog.Error("fail to delete agentsettinglist, err:%s", err.Error())
		data := createResponeDataV2(errcode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("delete agentsettinglist finish")

	return
}

func (r *Router) setAgentSettingList(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv set agentsettinglist request")

	var agents []*commtypes.BcsClusterAgentSetting

	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&agents); err != nil {
		blog.Warn("fail to Decode json for agents, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	errcode, err := r.backend.SetAgentSettingList(agents)
	if err != nil {
		blog.Error("fail to set agentsettinglist, err:%s", err.Error())
		data := createResponeDataV2(errcode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("set agentsettinglist finish")

	return
}

func (r *Router) disableAgentList(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv disable agentlist request")

	var IPs []string
	if req.QueryParameter("ips") != "" {
		IPs = strings.Split(req.QueryParameter("ips"), ",")
	}
	errcode, err := r.backend.DisableAgentList(IPs)
	if err != nil {
		blog.Error("fail to disable agentlist, err:%s", err.Error())
		data := createResponeDataV2(errcode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("disable agentlist finish")

	return
}

func (r *Router) enableAgentList(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv enable agentlist request")

	var IPs []string
	if req.QueryParameter("ips") != "" {
		IPs = strings.Split(req.QueryParameter("ips"), ",")
	}
	errcode, err := r.backend.EnableAgentList(IPs)
	if err != nil {
		blog.Error("fail to enable agentlist, err:%s", err.Error())
		data := createResponeDataV2(errcode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("enable agentlist finish")

	return
}

func (r *Router) updateAgentSettingList(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv update agentsetting request")

	var update commtypes.BcsClusterAgentSettingUpdate

	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&update); err != nil {
		blog.Error("fail to Decode json for update, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	errcode, err := r.backend.UpdateAgentSettingList(&update)
	if err != nil {
		blog.Error("fail to update agentsetting, err:%s", err.Error())
		data := createResponeDataV2(errcode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("update agentsetting finish")

	return
}

func (r *Router) queryAgentSetting(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv query agent setting request")

	IP := req.PathParameter("IP")

	setting, err := r.backend.QueryAgentSetting(IP)
	if err != nil {
		blog.Error("fail to query agent(%s) setting, err:%s", IP, err.Error())
		data := createResponeDataV2(comm.BcsErrCommGetZkNodeFail, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "", setting)
	resp.Write([]byte(data))
	blog.Info("query agent(%s) setting finish", IP)

	return
}

func (r *Router) disableAgent(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv disable agent request")

	IP := req.PathParameter("IP")

	err := r.backend.DisableAgent(IP)
	if err != nil {
		blog.Error("fail to disable agent(%s), err:%s", IP, err.Error())
		data := createResponeDataV2(comm.BcsErrCommCreateZkNodeFail, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("disable agent(%s) finish", IP)

	return
}

func (r *Router) enableAgent(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv enable agent request")

	IP := req.PathParameter("IP")

	err := r.backend.EnableAgent(IP)
	if err != nil {
		blog.Error("fail to enable agent(%s), err:%s", IP, err.Error())
		data := createResponeDataV2(comm.BcsErrCommCreateZkNodeFail, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("enable agent(%s) finish", IP)

	return
}

func (r *Router) healthCheckReport(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	var healthCheck commtypes.HealthCheckResult
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&healthCheck); err != nil {
		blog.Error("fail to Decode json for healthCheck, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	//blog.Infof("recv HealthCheckResult: %+v", healthCheck)
	go r.backend.HealthyReport(&healthCheck)

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	return
}

func (r *Router) createDeployment(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv create deployment request")

	var deploymentDef types.DeploymentDef
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&deploymentDef); err != nil {
		blog.Error("fail to Decode json to create deployment, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request create deployment(%s.%s)",
		deploymentDef.ObjectMeta.NameSpace, deploymentDef.ObjectMeta.Name)
	if deploymentDef.RawJson == nil {
		blog.Warn("request create deployment(%s.%s) without raw json, please check driver version",
			deploymentDef.ObjectMeta.NameSpace, deploymentDef.ObjectMeta.Name)
	}

	if errcode, err := r.backend.CreateDeployment(&deploymentDef); err != nil {
		blog.Error("fail to create deployment(%s.%s), err:%s",
			deploymentDef.ObjectMeta.NameSpace, deploymentDef.ObjectMeta.Name, err.Error())
		data := createResponeDataV2(errcode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request create deployment(%s.%s) end",
		deploymentDef.ObjectMeta.NameSpace, deploymentDef.ObjectMeta.Name)
	return
}

func (r *Router) updateDeployment(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv update deployment request")

	var deploymentDef types.DeploymentDef
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&deploymentDef); err != nil {
		blog.Error("fail to Decode json to update deployment , err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request update deployment(%s.%s)",
		deploymentDef.ObjectMeta.NameSpace, deploymentDef.ObjectMeta.Name)
	if deploymentDef.RawJson == nil {
		blog.Warn("request update deployment(%s.%s) without raw json, please check driver version",
			deploymentDef.ObjectMeta.NameSpace, deploymentDef.ObjectMeta.Name)
	}

	if errCode, err := r.backend.UpdateDeployment(&deploymentDef); err != nil {
		blog.Error("fail to update deployment(%s.%s), err:%s",
			deploymentDef.ObjectMeta.NameSpace, deploymentDef.ObjectMeta.Name, err.Error())
		data := createResponeDataV2(errCode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request update deployment(%s.%s) end",
		deploymentDef.ObjectMeta.NameSpace, deploymentDef.ObjectMeta.Name)
	return
}

func (r *Router) cancelUpdateDeployment(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	ns := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.Infof("request cancelupdate depolyment(%s.%s)", ns, name)

	if err := r.backend.CancelUpdateDeployment(ns, name); err != nil {
		blog.Error("fail to cancelupdate deployment(%s.%s), err:%s", ns, name, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request cancelupdate deployment(%s.%s) end", ns, name)
	return
}

func (r *Router) pauseUpdateDeployment(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	ns := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.Infof("request pauseupdate depolyment(%s.%s)", ns, name)

	if err := r.backend.PauseUpdateDeployment(ns, name); err != nil {
		blog.Error("fail to pauseupdate deployment(%s.%s), err:%s", ns, name, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request pauseupdate deployment(%s.%s) end", ns, name)
	return
}

func (r *Router) resumeUpdateDeployment(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	ns := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.Infof("request resumeupdate depolyment(%s.%s)", ns, name)

	if err := r.backend.ResumeUpdateDeployment(ns, name); err != nil {
		blog.Error("fail to resumeupdate deployment(%s.%s), err:%s", ns, name, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request resumeupdate deployment(%s.%s) end", ns, name)
	return
}

func (r *Router) deleteDeployment(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	enforce := false
	enforcePara := req.QueryParameter("enforce")
	if enforcePara == "1" {
		enforce = true
	}

	ns := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.Infof("request delete deployment(%s.%s)", ns, name)

	var data string
	if errCode, err := r.backend.DeleteDeployment(ns, name, enforce); err != nil {
		blog.Error("fail to delete deployment(%s.%s), err:%s", ns, name, err.Error())
		//if strings.Contains(err.Error(),"node does not exist") {
		//	data = createResponeDataV2(common.BcsErrMesosSchedNotFound, err.Error(), nil)
		//}else{
		//	data = createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		//}
		data = createResponeDataV2(errCode, err.Error(), nil)

		resp.Write([]byte(data))
		return
	}

	data = createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request delete deployment(%s.%s) end", ns, name)
	return
}

func (r *Router) getClusterResources(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	blog.V(3).Info("request get cluster resource request")

	res, err := r.backend.GetClusterResources()
	if err != nil {
		blog.Error("request get cluster resource request err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
	} else {
		data := createResponeData(nil, "", res)
		resp.Write([]byte(data))
		blog.Info("request get cluster resource request finish")
	}
	return
}

func (r *Router) getCurrentOffers(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	blog.V(3).Info("request get current offers request")

	res := r.backend.GetCurrentOffers()
	data := createResponeData(nil, "", res)
	resp.Write([]byte(data))
	blog.Info("request get current offers request finish")
}

func (r *Router) getClusterEndpoints(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	blog.V(3).Info("request get endpoints request")

	endpoints := r.backend.GetClusterEndpoints()
	data := createResponeData(nil, "", endpoints)
	resp.Write([]byte(data))
	blog.Info("request get endpoints request finish")
	return
}

func (r *Router) createConfigMap(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	var configmap commtypes.BcsConfigMap
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&configmap); err != nil {
		blog.Error("fail to decode configmap json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request create configmap(%s.%s): %+v", configmap.ObjectMeta.NameSpace, configmap.ObjectMeta.Name, configmap)

	currData, _ := r.backend.FetchConfigMap(configmap.ObjectMeta.NameSpace, configmap.ObjectMeta.Name)
	if currData != nil {
		err := errors.New("configmap already exist")
		data := createResponeDataV2(comm.BcsErrMesosSchedResourceExist, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.SaveConfigMap(&configmap); err != nil {
		blog.Error("fail to save configmap, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request create configmap(%s.%s) end", configmap.ObjectMeta.NameSpace, configmap.ObjectMeta.Name)

	return
}

func (r *Router) updateConfigMap(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	var configmap commtypes.BcsConfigMap
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&configmap); err != nil {
		blog.Error("fail to decode configmap json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request update configmap(%s.%s): %+v",
		configmap.ObjectMeta.NameSpace, configmap.ObjectMeta.Name, configmap)
	currData, _ := r.backend.FetchConfigMap(configmap.ObjectMeta.NameSpace, configmap.ObjectMeta.Name)
	if currData == nil {
		err := errors.New("configmap not exist")
		data := createResponeDataV2(comm.BcsErrMesosSchedNotFound, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.SaveConfigMap(&configmap); err != nil {
		blog.Error("fail to save configmap, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("request update configmap(%s.%s) end", configmap.ObjectMeta.NameSpace, configmap.ObjectMeta.Name)
	return
}

func (r *Router) deleteConfigMap(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	ns := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.V(3).Infof("request delete configmap(%s.%s)", ns, name)

	var data string
	if err := r.backend.DeleteConfigMap(ns, name); err != nil {
		blog.Error("fail to delete configmap, err:%s", err.Error())
		if strings.Contains(err.Error(), "node does not exist") {
			data = createResponeDataV2(common.BcsErrMesosSchedNotFound, err.Error(), nil)
		} else {
			data = createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		}
		resp.Write([]byte(data))
		return
	}

	data = createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request delete configmap(%s.%s) end", ns, name)
	return
}

func (r *Router) createSecret(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	var secret commtypes.BcsSecret
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&secret); err != nil {
		blog.Error("fail to decode secret json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request create secret(%s.%s): %+v", secret.ObjectMeta.NameSpace, secret.ObjectMeta.Name, secret)

	currData, _ := r.backend.FetchSecret(secret.ObjectMeta.NameSpace, secret.ObjectMeta.Name)
	if currData != nil {
		err := errors.New("secret already exist")
		data := createResponeDataV2(comm.BcsErrMesosSchedResourceExist, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.SaveSecret(&secret); err != nil {
		blog.Error("fail to save secret, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request create secret(%s.%s) end", secret.ObjectMeta.NameSpace, secret.ObjectMeta.Name)

	return
}

func (r *Router) updateSecret(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	var secret commtypes.BcsSecret
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&secret); err != nil {
		blog.Error("fail to decode secret json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request update secret(%s.%s): %+v",
		secret.ObjectMeta.NameSpace, secret.ObjectMeta.Name, secret)
	currData, _ := r.backend.FetchSecret(secret.ObjectMeta.NameSpace, secret.ObjectMeta.Name)
	if currData == nil {
		err := errors.New("secret not exist")
		data := createResponeDataV2(comm.BcsErrMesosSchedNotFound, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.SaveSecret(&secret); err != nil {
		blog.Error("fail to save secret, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("request secret secret(%s.%s) end", secret.ObjectMeta.NameSpace, secret.ObjectMeta.Name)
	return
}

func (r *Router) deleteSecret(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	ns := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.V(3).Infof("request delete secret(%s.%s)", ns, name)

	var data string
	if err := r.backend.DeleteSecret(ns, name); err != nil {
		blog.Error("fail to delete secret, err:%s", err.Error())
		if strings.Contains(err.Error(), "node does not exist") {
			data = createResponeDataV2(common.BcsErrMesosSchedNotFound, err.Error(), nil)
		} else {
			data = createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		}
		resp.Write([]byte(data))
		return
	}

	data = createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request delete secret(%s.%s) end", ns, name)
	return
}

func (r *Router) createService(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	var service commtypes.BcsService
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&service); err != nil {
		blog.Error("fail to decode service json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request create service(%s.%s):%+v", service.ObjectMeta.NameSpace, service.ObjectMeta.Name, service)

	currData, _ := r.backend.FetchService(service.ObjectMeta.NameSpace, service.ObjectMeta.Name)
	if currData != nil {
		err := errors.New("service already exist")
		data := createResponeDataV2(comm.BcsErrMesosSchedResourceExist, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.SaveService(&service); err != nil {
		blog.Error("fail to save service, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request create service(%s.%s) end", service.ObjectMeta.NameSpace, service.ObjectMeta.Name)

	return
}

func (r *Router) updateService(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	var service commtypes.BcsService
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&service); err != nil {
		blog.Error("fail to decode service json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request update servie(%s.%s): %+v", service.ObjectMeta.NameSpace, service.ObjectMeta.Name, service)

	currData, _ := r.backend.FetchService(service.ObjectMeta.NameSpace, service.ObjectMeta.Name)
	if currData == nil {
		err := errors.New("service not exist")
		data := createResponeDataV2(comm.BcsErrMesosSchedNotFound, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	//only update service spec
	service.ObjectMeta.Name = currData.ObjectMeta.Name
	service.ObjectMeta.NameSpace = currData.ObjectMeta.NameSpace
	service.ObjectMeta.Labels = currData.ObjectMeta.Labels
	service.TypeMeta = currData.TypeMeta

	if err := r.backend.SaveService(&service); err != nil {
		blog.Error("fail to save service, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("request update service(%s.%s) end", service.ObjectMeta.NameSpace, service.ObjectMeta.Name)
	return
}

func (r *Router) deleteService(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	ns := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.V(3).Infof("request delete service(%s.%s)", ns, name)

	var data string
	if err := r.backend.DeleteService(ns, name); err != nil {
		blog.Error("fail to delete service, err:%s", err.Error())
		if strings.Contains(err.Error(), "node does not exist") {
			data = createResponeDataV2(common.BcsErrMesosSchedNotFound, err.Error(), nil)
		} else {
			data = createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		}
		resp.Write([]byte(data))
		return
	}

	data = createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request delete service(%s.%s) end", ns, name)
	return
}

// BuildApplication is used to build a new application.
func (r *Router) buildApplication(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv build application request")

	var version types.Version
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&version); err != nil {
		blog.Error("fail to decode json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if version.RawJson == nil {
		blog.Error("request create application(%s.%s) without raw json", version.RunAs, version.ID)
	}

	if version.Instances <= 0 {
		blog.Error("request build application(%s.%s) Instances(%d) err", version.RunAs, version.ID, version.Instances)
		err := errors.New("instances error")
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	versionErr := r.backend.CheckVersion(&version)
	if versionErr != nil {
		blog.Error("build application(%s.%s) version error: %s", version.RunAs, version.ID, versionErr.Error())
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, versionErr.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	//check the resource, if not, set default
	err := version.CheckAndDefaultResource()
	if err != nil {
		blog.Error("build application(%s.%s) version error: %s", version.RunAs, version.ID, err.Error())
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if version.CheckConstraints() == false {
		blog.Error("request build: check constraints failed")
		err := errors.New("constraints error")
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	app, err := r.backend.FetchApplication(version.RunAs, version.ID)
	if err != nil && err != store.ErrNoFound {
		blog.Error("request build: fail to fetch application, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if app != nil {
		err = errors.New("application already exist")
		blog.Warn("request build fail: app(%s.%s) is already exist", version.RunAs, version.ID)
		data := createResponeDataV2(comm.BcsErrMesosSchedResourceExist, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	application := types.Application{
		Kind:             version.Kind,
		ID:               version.ID,
		Name:             version.ID,
		DefineInstances:  uint64(version.Instances),
		Instances:        0,
		RunningInstances: 0,
		RunAs:            version.RunAs,
		ClusterId:        r.backend.ClusterId(),
		Status:           types.APP_STATUS_STAGING,
		Created:          time.Now().Unix(),
		UpdateTime:       time.Now().Unix(),
		ObjectMeta:       version.ObjectMeta,
	}

	blog.Info("request build: save application(RunAs:%s ID:%s)", application.RunAs, application.ID)
	if err := r.backend.SaveApplication(&application); err != nil {
		blog.Error("request build: fail to SaveApplication(%s.%s), err:%s", application.RunAs, application.ID, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.SaveVersion(version.RunAs, version.ID, &version); err != nil {
		blog.Error("request build: fail to SaveVersion(%s.%s), err:%s", version.RunAs, version.ID, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.LaunchApplication(&version); err != nil {
		blog.Error("request build application(%s.%s) failed with error: %s", version.RunAs, version.ID, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request build application(%s.%s) end", version.RunAs, version.ID)
	return
}

// ListApplications is used to list all applications.
func (r *Router) listApplications(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv list applications request")
	runAs := req.PathParameter("runAs")

	blog.Info("request list applications under namespace(%s)", runAs)

	apps, err := r.backend.ListApplications(runAs)
	if err != nil {
		blog.Error("request list application under namespace(%s) failed: %s", runAs, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "", apps)
	resp.Write([]byte(data))

	blog.Info("request list applications under namespcae(%s) end", runAs)
	return
}

// FetchApplication is used to fetch a application via applicaiton id.
func (r *Router) fetchApplication(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv fetch application request")
	runAs := req.PathParameter("runAs")
	appId := req.PathParameter("appId")

	blog.Info("request fetch application(%s %s)", runAs, appId)

	app, err := r.backend.FetchApplication(runAs, appId)
	if err != nil {
		blog.Error("request fetch application(%s %s) failed: %s", runAs, appId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "", app)
	resp.Write([]byte(data))

	blog.Info("request fetch application(%s %s) end", runAs, appId)
	return
}

func (r *Router) getApplicationDef(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv get application definition request")
	runAs := req.PathParameter("ns")
	appId := req.PathParameter("name")

	blog.Info("request definition of  application(%s::%s)", runAs, appId)

	version, err := r.backend.GetVersion(runAs, appId)
	if err != nil {
		blog.Error("request get definition of application(%s::%s) failed: %s", runAs, appId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if version == nil || version.RawJson == nil {
		blog.Error("request get definition of application(%s::%s) failed: rawJson is nil ", runAs, appId)
		err := errors.New("application's definition not exist, maybe the application was created by deployment")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "", version.RawJson)
	resp.Write([]byte(data))

	blog.Info("request get definition of application(%s::%s) end", runAs, appId)
	return

}

func (r *Router) getDeploymentDef(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv get deployment definition request")
	runAs := req.PathParameter("ns")
	deploymentId := req.PathParameter("name")

	blog.Info("request definition of deployment(%s::%s)", runAs, deploymentId)

	deployment, err := r.backend.GetDeployment(runAs, deploymentId)
	if err != nil {
		blog.Error("request get definition of deployment(%s::%s) failed: %s", runAs, deploymentId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if deployment == nil || deployment.RawJson == nil {
		blog.Error("request get definition of deployment(%s::%s) failed: rawJson is nil ", runAs, deploymentId)
		err := errors.New("deployment's definition not exist")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "", deployment.RawJson)
	resp.Write([]byte(data))

	blog.Info("request get definition of deployment(%s::%s) end", runAs, deploymentId)
	return
}

// DeleteApplication is used to delete a application from mesos and consul via application id.
func (r *Router) deleteApplication(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv delete application request")
	runAs := req.PathParameter("runAs")
	appId := req.PathParameter("appId")
	kind := commtypes.BcsDataType(req.QueryParameter("kind"))

	enforce := false
	enforcePara := req.QueryParameter("enforce")
	if enforcePara == "1" {
		enforce = true
	}

	blog.Info("request delete application(%s %s), enfore:%s", runAs, appId, enforcePara)

	var data string
	if err := r.backend.DeleteApplication(runAs, appId, enforce, kind); err != nil {
		blog.Warn("request delete application (%s %s) failed: %s", runAs, appId, err.Error())
		if strings.Contains(err.Error(), "node does not exist") {
			data = createResponeDataV2(common.BcsErrMesosSchedNotFound, common.BcsErrMesosSchedNotFoundStr, nil)
		} else {
			data = createResponeData(err, err.Error(), nil)
		}
		resp.Write([]byte(data))
		return
	}

	data = createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request delete application(%s %s) end", runAs, appId)
	return
}

//ListApplicationTasks is used to list all tasks belong to application via application id.
func (r *Router) listApplicationTasks(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv list application tasks request")
	runAs := req.PathParameter("runAs")
	appId := req.PathParameter("appId")

	blog.Info("request list application(%s %s) tasks", runAs, appId)

	tasks, err := r.backend.ListApplicationTasks(runAs, appId)
	if err != nil {
		blog.Error("request list application tasks (%s %s) failed: %s", runAs, appId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request list application(%s %s) tasks, return num(%d)", runAs, appId, len(tasks))

	data := createResponeData(nil, "", tasks)
	resp.Write([]byte(data))

	blog.Info("request list application(%s %s) tasks end", runAs, appId)

	return
}

func (r *Router) listApplicationTaskGroups(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv list application taskgroups request")
	runAs := req.PathParameter("runAs")
	appId := req.PathParameter("appId")

	blog.Info("request list application(%s %s) taskgroups", runAs, appId)

	taskGroups, err := r.backend.ListApplicationTaskGroups(runAs, appId)
	if err != nil {
		blog.Error("request list application taskgroups (%s %s) failed: %s", runAs, appId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request list application(%s %s) taskgroups, return num(%d)", runAs, appId, len(taskGroups))

	data := createResponeData(nil, "", taskGroups)
	resp.Write([]byte(data))
	blog.Info("request list application(%s %s) taskgroups end", runAs, appId)
	return
}

// DeleteApplicationTaskGroups is used to delete all tasks belong to application via applicaiton id.
func (r *Router) deleteApplicationTaskGroups_r(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.Error("receive delete taskgroups request")
}

// DeleteApplicationTaskGroup is used to delete specified task belong to application via application id and task id.
func (r *Router) deleteApplicationTaskGroup_r(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.Error("receive delete taskgroup request")
}

// ListApplicationVersions is used to list all versions for a application specified by applicationId.
func (r *Router) listApplicationVersions(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv list application versions request")
	runAs := req.PathParameter("runAs")
	appId := req.PathParameter("appId")

	blog.Info("request list application(%s %s) versions", runAs, appId)

	appVersions, err := r.backend.ListApplicationVersions(runAs, appId)
	if err != nil {
		blog.Error("request list application versions (%s %s) failed: %s", runAs, appId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "", appVersions)
	resp.Write([]byte(data))
	blog.Info("request list application(%s %s) versions end", runAs, appId)
	return
}

// FetchApplicationVersion is used to fetch specified version from consul by version id and application id.
func (r *Router) fetchApplicationVersion_r(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv fetch application versions request")
	runAs := req.PathParameter("runAs")
	appId := req.PathParameter("appId")
	versionId := req.PathParameter("versionId")

	blog.Info("request fetch application version(%s %s %s)", runAs, appId, versionId)

	version, err := r.backend.FetchApplicationVersion(runAs, appId, versionId)
	if err != nil {
		blog.Error("request fetch application version (%s %s %s) failed: %s", runAs, appId, versionId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "", version)
	resp.Write([]byte(data))
	blog.Info("request fetch application version(%s %s %s) end", runAs, appId, versionId)
	return
}

// UpdateApplication is used to update application version.
func (r *Router) updateApplication(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("recv update application request")

	runAs := req.PathParameter("runAs")
	appId := req.PathParameter("appId")
	instances := req.QueryParameter("instances")
	args := req.QueryParameter("args")
	blog.Info("request update application(%s.%s): instances(%s), args(%s)", runAs, appId, instances, args)

	var version types.Version
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&version); err != nil {
		blog.Error("request update application(%s.%s) fail to decode version. err:%s", err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}
	versionErr := r.backend.CheckVersion(&version)
	if versionErr != nil {
		blog.Error("update application(%s.%s) version error: %s", version.RunAs, version.ID, versionErr.Error())
		data := createResponeData(versionErr, versionErr.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if version.RunAs != runAs || version.ID != appId {
		blog.Error("request update application(%s.%s) version err: version(%s.%s)", runAs, appId, version.RunAs, version.ID)
		err := errors.New("version RunAs or ID not correct")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	//check the resource, if not, set default
	err := version.CheckAndDefaultResource()
	if err != nil {
		blog.Error("update application (%s.%s) version error: %s", version.RunAs, version.ID, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if version.CheckConstraints() == false {
		blog.Error("request update application(%s.%s) fail for version constraints error", runAs, appId)
		err = errors.New("Version Constraints error")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	currVersion, _ := r.backend.GetVersion(runAs, appId)
	if currVersion == nil {
		blog.Error("request update application(%s.%s) fail for cannot get curr version", runAs, appId)
		err := errors.New("cannot get old version data")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	// added  20181011, currVersion.Kind is a new field and the current existing version kind maybe empty.
	// if currVersion.Kind is empty and request kind is PROCESS, the update will not be allowed.
	// if currVersion.Kind is not empty and it is different from request kind, also the update will not be allowed.
	currentKind := currVersion.Kind
	if currentKind == "" {
		currentKind = commtypes.BcsDataType_APP
	}
	if currentKind != version.Kind {
		blog.Errorf("request update application(%s.%s) fail for different kind, current(%s) updated(%s)", runAs, appId, currentKind, version.Kind)
		err := errors.New("cannot update different kind application")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if currVersion.Instances != version.Instances {
		blog.Error("request update application(%s.%s) err for old version instances(%d) != new version instances(%d)",
			runAs, appId, currVersion.Instances, version.Instances)
		version.Instances = currVersion.Instances
		versionErr := r.backend.CheckVersion(&version)
		if versionErr != nil {
			blog.Error("update application (%s %s) version error: %s", version.RunAs, version.ID, versionErr.Error())
			data := createResponeData(versionErr, versionErr.Error(), nil)
			resp.Write([]byte(data))
			return
		}
	}

	var instanceNum uint64
	instanceNum, err = strconv.ParseUint(instances, 10, 64)
	if args == "resource" {
		blog.Infof("request update application(%s.%s) resource", runAs, appId)
	} else {
		if err != nil {
			blog.Error("request update application(%s.%s) parameter err instances(%s)", runAs, appId, instances)
			err = errors.New("instances must be specified")
			data := createResponeData(err, err.Error(), nil)
			resp.Write([]byte(data))
			return
		}
		blog.Infof("request update application(%s.%s), instances(%s)", runAs, appId, instances)
	}
	if instanceNum > uint64(version.Instances) {
		blog.Error("request update application(%s.%s) err: instances(%d) > version.Instances(%d)", instanceNum, version.Instances)
		err := errors.New("update instances num is more than version.Instances")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.SaveVersion(runAs, appId, &version); err != nil {
		blog.Error("request update application(%s.%s) fail to save version. err:%s", runAs, appId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.UpdateApplication(runAs, appId, args, int(instanceNum), &version); err != nil {
		blog.Error("request update application(%s.%s) err:%s", runAs, appId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("request update application(%s.%s) end", runAs, appId)
	return
}

// ScaleApplication is used to scale application instances.
func (r *Router) scaleApplication(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive scale application request")

	runAs := req.PathParameter("runAs")
	appId := req.PathParameter("appId")
	instances := req.QueryParameter("instances")
	kind := commtypes.BcsDataType(req.QueryParameter("kind"))
	blog.Info("request scale application(%s %s) to instances(%s)", runAs, appId, instances)

	// limit the target instances
	instanceNum, err := strconv.ParseUint(instances, 10, 64)
	if err != nil {
		blog.Error("request scale application(%s %s): parameter err instances(%s)", instances)
		err = errors.New("instances must be specified")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}
	if instanceNum <= 0 {
		blog.Error("request scale application(%s %s): parameter err instances(%s)", instances)
		err = errors.New("target instances can not be littler than 1, maybe you want to use delete command")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.ScaleApplication(runAs, appId, instanceNum, kind, true); err != nil {
		blog.Error("request scale application(%s %s) instances(%d) err(%s)", runAs, appId, instanceNum, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Infof("request scale application(%s %s) instances(%d) end", runAs, appId, instanceNum)
	return
}

func (r *Router) sendApplicationCommand(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	var command commtypes.BcsCommand
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&command); err != nil {
		blog.Error("fail to decode json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if command.Spec == nil {
		blog.Error("command has no spec")
		err := errors.New("command has no spec")
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	pathNs := req.PathParameter("ns")
	pathName := req.PathParameter("name")
	kind := command.Spec.CommandTargetRef.Kind
	ns := command.Spec.CommandTargetRef.Namespace
	name := command.Spec.CommandTargetRef.Name
	if kind != "Application" || ns != pathNs || name != pathName {
		blog.Warn("send application command, data not correct: kind(%s), namespace(%s:%s), name(%s:%s)",
			kind, pathNs, ns, pathName, name)
		err := errors.New("request data error")
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	commandID := kind + "-" + ns + "-" + name + "-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	blog.Info("send command(%s) to %s:%s.%s", commandID, kind, ns, name)
	commandInfo := commtypes.BcsCommandInfo{
		Id:         commandID,
		Spec:       command.Spec,
		Status:     new(commtypes.BcsCommandStatus),
		CreateTime: time.Now().Unix(),
	}

	//do command
	if err := r.backend.DoCommand(&commandInfo); err != nil {
		blog.Error("fail to do command(%s), err:%s", commandID, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", commandID)
	resp.Write([]byte(data))

	blog.Info("request send command(%s) end", commandID)
	return
}

func (r *Router) getApplicationCommand(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	id := req.QueryParameter("id")
	blog.Info("request get command(%s)", id)
	command, err := r.backend.GetCommand(id)
	if err != nil {
		blog.Error("request get command(%s) failed: %s", id, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	pathNs := req.PathParameter("ns")
	pathName := req.PathParameter("name")
	kind := command.Spec.CommandTargetRef.Kind
	ns := command.Spec.CommandTargetRef.Namespace
	name := command.Spec.CommandTargetRef.Name
	if kind != "Application" || ns != pathNs || name != pathName {
		blog.Warn("get application command, data not correct: kind(%s), namespace(%s:%s), name(%s:%s)",
			kind, pathNs, ns, pathName, name)
		err := errors.New("request data error")
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "", command)
	resp.Write([]byte(data))

	blog.Info("request get command(%s) end", id)
	return
}

func (r *Router) deleteApplicationCommand(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	id := req.QueryParameter("id")
	blog.Infof("request delete command(%s)", id)

	//should auth the path ns:name with the command data, todo
	if err := r.backend.DeleteCommand(id); err != nil {
		blog.Error("fail to delete command(%s), err:%s", id, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("request delete command(%s) end", id)
	return
}

func (r *Router) sendDeploymentCommand(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	var command commtypes.BcsCommand
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&command); err != nil {
		blog.Error("fail to decode json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if command.Spec == nil {
		blog.Error("command has no spec")
		err := errors.New("command has no spec")
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	pathNs := req.PathParameter("ns")
	pathName := req.PathParameter("name")
	kind := command.Spec.CommandTargetRef.Kind
	ns := command.Spec.CommandTargetRef.Namespace
	name := command.Spec.CommandTargetRef.Name
	if kind != "Deployment" || ns != pathNs || name != pathName {
		blog.Warn("send deployment command, data not correct: kind(%s), namespace(%s:%s), name(%s:%s)",
			kind, pathNs, ns, pathName, name)
		err := errors.New("request data error")
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	commandID := strings.ToLower(kind + "-" + ns + "-" + name + "-" + strconv.FormatInt(time.Now().UnixNano(), 10))
	blog.Info("send command(%s) to %s:%s.%s", commandID, kind, ns, name)
	commandInfo := commtypes.BcsCommandInfo{
		Id:         commandID,
		Spec:       command.Spec,
		Status:     new(commtypes.BcsCommandStatus),
		CreateTime: time.Now().Unix(),
	}

	//do command
	if err := r.backend.DoCommand(&commandInfo); err != nil {
		blog.Error("fail to do command(%s), err:%s", commandID, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", commandID)
	resp.Write([]byte(data))

	blog.Info("request send command(%s) end", commandID)
	return
}

func (r *Router) getDeploymentCommand(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	id := req.QueryParameter("id")
	blog.Info("request get command(%s)", id)

	command, err := r.backend.GetCommand(id)
	if err != nil {
		blog.Error("request get command(%s) failed: %s", id, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	pathNs := req.PathParameter("ns")
	pathName := req.PathParameter("name")
	kind := command.Spec.CommandTargetRef.Kind
	ns := command.Spec.CommandTargetRef.Namespace
	name := command.Spec.CommandTargetRef.Name
	if kind != "Deployment" || ns != pathNs || name != pathName {
		blog.Warn("get deployment command, data not correct: kind(%s), namespace(%s:%s), name(%s:%s)",
			kind, pathNs, ns, pathName, name)
		err := errors.New("request data error")
		data := createResponeDataV2(comm.BcsErrCommRequestDataErr, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "", command)
	resp.Write([]byte(data))

	blog.Info("request get command(%s) end", id)
	return
}

func (r *Router) deleteDeploymentCommand(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}

	id := req.QueryParameter("id")
	blog.Infof("request delete command(%s)", id)

	//should auth the path ns:name with the command data, todo

	if err := r.backend.DeleteCommand(id); err != nil {
		blog.Error("fail to delete command(%s), err:%s", id, err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("request delete command(%s) end", id)
	return
}

//SendMessageApplication send msg to application
func (r *Router) sendMessageApplication(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive send message to application request")

	runAs := req.PathParameter("runAs")
	appId := req.PathParameter("appId")

	blog.Info("request send message to application(%s %s)", runAs, appId)

	var msg types.BcsMessage
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&msg); err != nil {
		blog.Error("request send message to application(%s %s): fail to Decode message json, err:%s", runAs, appId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	_, fail, err := r.backend.SendToApplication(runAs, appId, &msg)
	if err != nil {
		blog.Error("request send message to application(%s %s): fail to send message, err:%s", runAs, appId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if len(fail) != 0 {
		blog.Error("request send message to application(%s %s): fail count: %d", runAs, appId, len(fail))
		data := createResponeData(nil, "success", fail)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("request send message to application(%s %s) end", runAs, appId)
	return
}

//SendMessageApplicationTaskGroup send msg to the specified taskgroup
func (r *Router) sendMessageApplicationTaskGroup(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive send message to taskgroup request")

	runAs := req.PathParameter("runAs")
	appId := req.PathParameter("appId")
	taskgroupId := req.PathParameter("taskgroupId")

	blog.Info("request send message to taskgroup(%s)", taskgroupId)

	var msg types.BcsMessage
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&msg); err != nil {
		blog.Error("request send message to taskgroup(%s): fail to Decode json, err:%s", taskgroupId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.SendToApplicationTaskGroup(runAs, appId, taskgroupId, &msg); err != nil {
		blog.Error("request send message to taskgroup(%s): fail to send message, err:%s", taskgroupId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("request send message to taskgroup(%s) end", taskgroupId)
	return
}

// TODO(jinrui), http server will improve, this function is just use for this http server framwork
func createResponeData(err error, msg string, data interface{}) string {
	var rpyErr error
	if err != nil {
		rpyErr = bhttp.InternalError(common.BcsErrMesosSchedCommon, msg)
	} else {
		rpyErr = errors.New(bhttp.GetRespone(common.BcsSuccess, common.BcsSuccessStr, data))
	}

	blog.V(3).Infof("createRespone: %s", rpyErr.Error())

	return rpyErr.Error()
}

func createResponeDataV2(errCode int, msg string, data interface{}) string {
	var rpyErr error
	if errCode != 0 {
		rpyErr = bhttp.InternalError(errCode, msg)
	} else {
		rpyErr = errors.New(bhttp.GetRespone(common.BcsSuccess, common.BcsSuccessStr, data))
	}

	blog.V(3).Infof("createRespone: %s", rpyErr.Error())

	return rpyErr.Error()
}

// RescheduleTaskgroup is used to rescheduler taskgroup.
func (r *Router) reschedulerTaskgroup(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive rescheduler taskgroup request")

	taskgroupId := req.PathParameter("taskgroupId")

	var hostRetainTime int64
	hostRetainTimeStr := req.QueryParameter("hostRetainTime")
	if hostRetainTimeStr != "" {
		hostRetainTime, _ = strconv.ParseInt(hostRetainTimeStr, 10, 64)
	}

	blog.Info("request rescheduler taskgroup(%s) hostRetainTime(%d)", taskgroupId, hostRetainTime)

	if err := r.backend.RescheduleTaskgroup(taskgroupId, hostRetainTime); err != nil {
		blog.Error("request rescheduler taskgroup(%s) err(%s)", taskgroupId, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Infof("request rescheduler taskgroup(%s) end", taskgroupId)
	return
}

// ScaleDeployment is used to scale deployment instances.
func (r *Router) scaleDeployment_r(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive scale deployment request")

	runAs := req.PathParameter("namespace")
	name := req.PathParameter("name")
	instances := req.PathParameter("instances")
	blog.Info("request scale deployment(%s %s) to instances(%s)", runAs, name, instances)

	// limit the target instances
	instanceNum, err := strconv.ParseUint(instances, 10, 64)
	if err != nil {
		blog.Error("request scale deployment(%s %s): parameter err instances(%s)", runAs, name, instances)
		err = errors.New("instances must be specified")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}
	if instanceNum <= 0 {
		blog.Error("request scale deployment(%s %s): parameter err instances(%s)", instances)
		err = errors.New("target instances can not be littler than 1, maybe you want to use delete command")
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.ScaleDeployment(runAs, name, instanceNum); err != nil {
		blog.Error("request scale deployment(%s %s) instances(%d) err(%s)", runAs, name, instanceNum, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Infof("request scale deployment(%s %s) instances(%d) end", runAs, name, instanceNum)
	return
}

// ScaleDeployment is used to scale deployment instances.
func (r *Router) getDeployment_r(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive scale deployment request")

	runAs := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.Info("request get deployment(%s %s) information", runAs, name)
	o, err := r.backend.GetDeployment(runAs, name)
	if err != nil {
		blog.Error("request get deployment(%s %s) err(%s)", runAs, name, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", o)
	resp.Write([]byte(data))
	blog.Infof("request get deployment(%s %s) end", runAs, name)
	return
}

func (r *Router) registerCustomResource(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive register custom resource request")

	var crr *commtypes.Crr
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&crr); err != nil {
		blog.Error("request register custom resource fail to decode crr. err:%s", err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.RegisterCustomResource(crr); err != nil {
		blog.Error("request register custom resource(%s) err(%s)", crr.Spec.Names.Kind, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Infof("request register custom resource(%s)end", crr.Spec.Names.Kind)
	return
}

func (r *Router) createCustomResource(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive create custom resource request")

	var crd *commtypes.Crd
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&crd); err != nil {
		blog.Error("request create custom resource fail to decode Crd. err:%s", err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.CreateCustomResource(crd); err != nil {
		blog.Error("request create custom resource(%s %s %s) err(%s)", crd.Kind, crd.NameSpace, crd.Name, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Infof("request create custom resource(%s %s %s)end", crd.Kind, crd.NameSpace, crd.Name)
	return
}

func (r *Router) updateCustomResource(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive update custom resource request")

	var crd *commtypes.Crd
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&crd); err != nil {
		blog.Error("request update custom resource fail to decode Crd. err:%s", err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.UpdateCustomResource(crd); err != nil {
		blog.Error("request update custom resource(%s %s %s) err(%s)", crd.Kind, crd.NameSpace, crd.Name, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Infof("request update custom resource(%s %s %s)end", crd.Kind, crd.NameSpace, crd.Name)
	return
}

func (r *Router) deleteCustomResource(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive delete custom resource request")

	kind := req.PathParameter("kind")
	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	var data string
	if err := r.backend.DeleteCustomResource(kind, ns, name); err != nil {
		blog.Error("request delete custom resource(%s %s %s) err(%s)", kind, ns, name, err.Error())
		if strings.Contains(err.Error(), "node does not exist") {
			data = createResponeDataV2(common.BcsErrMesosSchedNotFound, common.BcsErrMesosSchedNotFoundStr, nil)
		} else {
			data = createResponeData(err, err.Error(), nil)
		}
		resp.Write([]byte(data))
		return
	}

	data = createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Infof("request delete custom resource(%s %s %s)end", kind, ns, name)
	return
}

func (r *Router) listCustomResource(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	kind := req.PathParameter("kind")
	ns := req.PathParameter("ns")
	blog.V(3).Infof("receive list custom resource request kind %s namespace %s", kind, ns)

	crds, err := r.backend.ListCustomResourceDefinition(kind, ns)
	if err != nil {
		blog.Error("request list custom resource(%s %s) err(%s)", kind, ns, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", crds)
	resp.Write([]byte(data))
	blog.V(3).Infof("request list custom resource(%s %s)end", kind, ns)
	return
}

func (r *Router) listAllCustomResource(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	kind := req.PathParameter("kind")
	blog.V(3).Infof("receive list all custom resource request kind %s", kind)

	crds, err := r.backend.ListAllCrds(kind)
	if err != nil {
		blog.Error("request list custom resource(%s) err(%s)", kind, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", crds)
	resp.Write([]byte(data))
	blog.V(3).Infof("request list custom resource(%s)end", kind)
	return
}

func (r *Router) getCustomResource(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	kind := req.PathParameter("kind")
	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	blog.V(3).Infof("receive get custom resource request kind %s namespace %s name %s", kind, ns, name)

	crd, err := r.backend.FetchCustomResourceDefinition(kind, ns, name)
	if err != nil {
		blog.Error("request get custom resource(%s %s %s) err(%s)", kind, ns, name, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", crd)
	resp.Write([]byte(data))
	blog.V(3).Infof("request get custom resource(%s %s %s)end", kind, ns, name)
	return
}

func (r *Router) commitImage(req *restful.Request, resp *restful.Response) {

	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive commit image request")

	taskgroup := req.PathParameter("taskgroup")
	image := req.QueryParameter("image")
	url := req.QueryParameter("url")

	var msg *types.BcsMessage
	var err error
	if msg, err = r.backend.CommitImage(taskgroup, image, url); err != nil {
		blog.Error("request commit image(%s %s %s) err(%s)", taskgroup, image, url, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", msg)
	resp.Write([]byte(data))
	blog.Error("request commit image(%s %s %s)end", taskgroup, image, url)
	return
}

func (r *Router) restartTaskGroup(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive restart taskGroup request")

	taskGroupID := req.PathParameter("taskGroupID")

	blog.Info("request restart taskGroup(%s)", taskGroupID)

	msg, err := r.backend.RestartTaskGroup(taskGroupID)
	if err != nil {
		blog.Error("request restart taskGroup(%s) err(%s)", taskGroupID, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", msg)
	resp.Write([]byte(data))
	blog.Infof("request restart taskGroup(%s) end", taskGroupID)
	return
}

func (r *Router) reloadTaskGroup(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("receive reload taskGroup request")

	taskGroupID := req.PathParameter("taskGroupID")

	blog.Info("request reload taskGroup(%s)", taskGroupID)

	msg, err := r.backend.ReloadTaskGroup(taskGroupID)
	if err != nil {
		blog.Error("request reload taskGroup(%s) err(%s)", taskGroupID, err.Error())
		data := createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", msg)
	resp.Write([]byte(data))
	blog.Infof("request reload taskGroup(%s) end", taskGroupID)
	return
}

func (r *Router) createAdmissionwebhook(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	var admission commtypes.AdmissionWebhookConfiguration
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&admission); err != nil {
		blog.Error("fail to decode admission json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request create admission(%s.%s):%+v", admission.ObjectMeta.NameSpace, admission.ObjectMeta.Name, admission)

	currData, _ := r.backend.FetchAdmissionWebhook(admission.ObjectMeta.NameSpace, admission.ObjectMeta.Name)
	if currData != nil {
		err := errors.New("admission already exist")
		data := createResponeDataV2(comm.BcsErrMesosSchedResourceExist, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	if err := r.backend.SaveAdmissionWebhook(&admission); err != nil {
		blog.Error("fail to save admission, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request create admission(%s.%s) end", admission.ObjectMeta.NameSpace, admission.ObjectMeta.Name)

	return
}

func (r *Router) updateAdmissionwebhook(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	var admission commtypes.AdmissionWebhookConfiguration
	decoder := json.NewDecoder(req.Request.Body)
	if err := decoder.Decode(&admission); err != nil {
		blog.Error("fail to decode admission json, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrCommJsonDecode, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	blog.Info("request update admission(%s.%s): %+v", admission.ObjectMeta.NameSpace, admission.ObjectMeta.Name, admission)

	currData, _ := r.backend.FetchAdmissionWebhook(admission.ObjectMeta.NameSpace, admission.ObjectMeta.Name)
	if currData == nil {
		err := errors.New("admission not exist")
		data := createResponeDataV2(comm.BcsErrMesosSchedNotFound, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	//only update service spec
	admission.ObjectMeta = currData.ObjectMeta
	admission.TypeMeta = currData.TypeMeta

	if err := r.backend.UpdateAdmissionWebhook(&admission); err != nil {
		blog.Error("fail to save admission, err:%s", err.Error())
		data := createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data := createResponeData(nil, "success", nil)
	resp.Write([]byte(data))
	blog.Info("request update admission(%s.%s) end", admission.ObjectMeta.NameSpace, admission.ObjectMeta.Name)
	return
}

func (r *Router) deleteAdmissionwebhook(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	ns := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.V(3).Infof("request delete admission(%s.%s)", ns, name)

	var data string
	if err := r.backend.DeleteAdmissionWebhook(ns, name); err != nil {
		blog.Error("fail to delete admission, err:%s", err.Error())
		if strings.Contains(err.Error(), "node does not exist") {
			data = createResponeDataV2(common.BcsErrMesosSchedNotFound, err.Error(), nil)
		} else {
			data = createResponeDataV2(comm.BcsErrMesosSchedCommon, err.Error(), nil)
		}
		resp.Write([]byte(data))
		return
	}

	data = createResponeData(nil, "success", nil)
	resp.Write([]byte(data))

	blog.Info("request delete admission(%s.%s) end", ns, name)
	return
}

func (r *Router) fetchAllAdmissionwebhooks(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	blog.V(3).Infof("request fetch all admissions")

	var data string
	admissions, err := r.backend.FetchAllAdmissionWebhooks()
	if err != nil {
		blog.Error("request list all admissions err(%s)", err.Error())
		data = createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data = createResponeData(nil, "success", admissions)
	resp.Write([]byte(data))

	blog.V(3).Info("request list all admissions end")
	return
}

func (r *Router) fetchAdmissionwebhook(req *restful.Request, resp *restful.Response) {
	if r.backend.GetRole() != scheduler.SchedulerRoleMaster {
		blog.Warn("scheduler is not master, can not process cmd")
		return
	}
	ns := req.PathParameter("namespace")
	name := req.PathParameter("name")
	blog.V(3).Infof("request fetch admission(%s:%s)", ns, name)

	var data string
	admission, err := r.backend.FetchAdmissionWebhook(ns, name)
	if err != nil {
		blog.Error("request list all admissions err(%s)", err.Error())
		data = createResponeData(err, err.Error(), nil)
		resp.Write([]byte(data))
		return
	}

	data = createResponeData(nil, "success", admission)
	resp.Write([]byte(data))

	blog.V(3).Info("request fetch admissions end")
	return
}
