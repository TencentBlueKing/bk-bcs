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

package v4http

import (
	"io/ioutil"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	commonTypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/backend/webconsole"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/config"

	restful "github.com/emicklei/go-restful"
)

//Scheduler is data struct of mesos scheduler
type Scheduler struct {
	config *config.MesosDriverConfig
	client *httpclient.HttpClient
	acts   []*httpserver.Action
	hosts  []string
	rwHost *sync.RWMutex
	//reversProxy from 1.15.x for CustomResource
	localProxy *kubeProxy
	//proxy from 1.17.x for mesos webconsole
	consoleProxy *webconsole.WebconsoleProxy
}

//NewScheduler create a scheduler
func NewScheduler() *Scheduler {
	s := &Scheduler{
		client: httpclient.NewHttpClient(),
		rwHost: new(sync.RWMutex),
	}
	return s
}

//InitConfig scheduler init configuration
func (s *Scheduler) InitConfig(conf *config.MesosDriverConfig) {
	s.config = conf

	//s.SetHost([]string{s.config.MesosHost})

	//client
	if s.config.ClientCert.IsSSL {
		blog.Info("mesos driver scheduler API is SSL: CA:%s, Cert:%s, Key:%s",
			s.config.ClientCert.CAFile, s.config.ClientCert.CertFile, s.config.ClientCert.KeyFile)
		s.client.SetTlsVerity(s.config.ClientCert.CAFile, s.config.ClientCert.CertFile, s.config.ClientCert.KeyFile, s.config.ClientCert.CertPasswd)
	}

	s.client.SetHeader("Content-Type", "application/json")
	s.client.SetHeader("Accept", "application/json")
	//init kube client for CRD
	s.initKube()
	//init webconsole proxy
	s.initMesosWebconsole()
	s.initActions()
}

//Actions all http action implementation
func (s *Scheduler) Actions() []*httpserver.Action {
	return s.acts
}

//GetHttpClient get scheudler specified http client implementation
func (s *Scheduler) GetHttpClient() *httpclient.HttpClient {
	return s.client
}

func (s *Scheduler) initActions() {
	s.acts = []*httpserver.Action{
		/*================= application ====================*/
		httpserver.NewAction("POST", "/namespaces/{ns}/applications", nil, s.CreateApplicationHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/applications", nil, s.UpdateApplicationHandler),
		httpserver.NewAction("DELETE", "/namespaces/{ns}/applications/{name}", nil, s.DeleteApplicationHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/applications/rollback", nil, s.RollbackApplicationHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/applications/{name}/scale/{instances}", nil, s.ScaleApplicationHandler),
		httpserver.NewAction("GET", "/namespaces/{ns}/applications", nil, s.ListApplicationsHandler),
		httpserver.NewAction("GET", "/namespaces/{ns}/applications/{name}", nil, s.FetchApplicationHandler),
		/*================= application ====================*/

		/*================= process =======================*/
		httpserver.NewAction("POST", "/namespaces/{ns}/processes", nil, s.CreateProcessHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/processes", nil, s.UpdateProcessHandler),
		httpserver.NewAction("DELETE", "/namespaces/{ns}/processes/{name}", nil, s.DeleteProcessHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/processes/rollback", nil, s.RollbackProcessHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/processes/{name}/scale/{instances}", nil, s.ScaleProcessHandler),
		httpserver.NewAction("GET", "/namespaces/{ns}/processes", nil, s.ListProcessesHandler),
		httpserver.NewAction("GET", "/namespaces/{ns}/processes/{name}", nil, s.FetchProcessHandler),
		/*================= process =======================*/

		/*================= message ====================*/
		httpserver.NewAction("POST", "/namespaces/{ns}/applications/{name}/message", nil, s.SendMessageApplicationHandler),
		httpserver.NewAction("POST", "/namespaces/{ns}/applications/{name}/taskgroups/{taskgroup-name}/message", nil, s.SendMessageTaskgroupHandler),
		/*================= message ====================*/

		/*================= task ====================*/
		httpserver.NewAction("GET", "/namespaces/{ns}/applications/{name}/tasks", nil, s.ListApplicationTasksHandler),
		/*================= task ====================*/

		/*================= taskgroup ====================*/
		httpserver.NewAction("GET", "/namespaces/{ns}/applications/{name}/taskgroups", nil, s.ListApplicationTaskGroupsHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/applications/{appid}/taskgroups/{taskgroupId}/rescheduler", nil, s.reschedulerTaskgroupHandler),
		httpserver.NewAction("POST", "/namespaces/{ns}/applications/{appid}/taskgroups/{taskGroupID}/restart", nil, s.restartTaskGroupHandler),
		httpserver.NewAction("POST", "/namespaces/{ns}/applications/{appid}/taskgroups/{taskGroupID}/reload", nil, s.reloadTaskGroupHandler),
		/*================= taskgroup ====================*/

		/*================= version ====================*/
		httpserver.NewAction("GET", "/namespaces/{ns}/applications/{name}/versions", nil, s.ListApplicationVersionsHandler),
		httpserver.NewAction("GET", "/namespaces/{ns}/applications/{name}/versions/{versionid}", nil, s.FetchApplicationVersionHandler),
		/*================= version ====================*/

		/*================= configmap ====================*/
		httpserver.NewAction("POST", "/namespaces/{ns}/configmaps", nil, s.CreateConfigMapHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/configmaps", nil, s.UpdateConfigMapHandler),
		httpserver.NewAction("DELETE", "/namespaces/{ns}/configmaps/{name}", nil, s.DeleteConfigMapHandler),
		/*================= configmap ====================*/

		/*================= secret ====================*/
		httpserver.NewAction("POST", "/namespaces/{ns}/secrets", nil, s.CreateSecretHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/secrets", nil, s.UpdateSecretHandler),
		httpserver.NewAction("DELETE", "/namespaces/{ns}/secrets/{name}", nil, s.DeleteSecretHandler),
		/*================= secret ====================*/

		/*================= service ====================*/
		httpserver.NewAction("POST", "/namespaces/{ns}/services", nil, s.CreateServiceHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/services", nil, s.UpdateServiceHandler),
		httpserver.NewAction("DELETE", "/namespaces/{ns}/services/{name}", nil, s.DeleteServiceHandler),
		/*================= service ====================*/

		/*================= cluster ====================*/
		httpserver.NewAction("GET", "/cluster/resources", nil, s.GetClusterResourcesHandler),
		httpserver.NewAction("GET", "/cluster/endpoints", nil, s.GetClusterEndpointsHandler),
		httpserver.NewAction("GET", "/cluster/current/offers", nil, s.GetClusterCurrentOffersHandler),
		/*================= cluster ====================*/

		/*================= deployment ====================*/
		httpserver.NewAction("POST", "/namespaces/{ns}/deployments", nil, s.createDeploymentHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/deployments", nil, s.udpateDeploymentHandler),
		httpserver.NewAction("DELETE", "/namespaces/{ns}/deployments/{name}", nil, s.deleteDeploymentHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/deployments/{name}/cancelupdate", nil, s.cancelupdateDeploymentHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/deployments/{name}/pauseupdate", nil, s.pauseupdateDeploymentHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/deployments/{name}/resumeupdate", nil, s.resumeupdateDeploymentHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/deployments/{name}/scale/{instances}", nil, s.scaleDeploymentHandler),
		/*================= deployment ====================*/

		/*================= daemonset ====================*/
		httpserver.NewAction("POST", "/namespaces/{ns}/daemonset", nil, s.createDaemonsetHandler),
		httpserver.NewAction("DELETE", "/namespaces/{ns}/daemonset/{name}", nil, s.deleteDaemonsetHandler),
		/*================= daemonset ====================*/

		/*================= transaction ====================*/
		httpserver.NewAction("GET", "/transactions/{ns}", nil, s.listTransactionHandler),
		httpserver.NewAction("DELETE", "/transactions/{ns}/{name}", nil, s.deleteTransactionHandler),
		/*================= transaction ====================*/

		/*================= agentsetting ====================*/
		//	httpserver.NewAction("POST","/agentsetting/{IP}/disable",nil,s.disableAgentHandler),
		//	httpserver.NewAction("POST","/agentsetting/{IP}/enable",nil,s.enableAgentHandler),
		//	httpserver.NewAction("GET","/agentsetting/{IP}",nil,s.getAgentSettingHandler),

		httpserver.NewAction("GET", "/agentsettings", nil, s.getAgentSettingListHandler),
		httpserver.NewAction("DELETE", "/agentsettings", nil, s.deleteAgentSettingListHandler),
		httpserver.NewAction("POST", "/agentsettings", nil, s.setAgentSettingListHandler),
		//httpserver.NewAction("POST", "/agentsettings/update", nil, s.updateAgentSettingListHandler),
		httpserver.NewAction("POST", "/agentsettings/enable", nil, s.enableAgentListHandler),
		httpserver.NewAction("POST", "/agentsettings/disable", nil, s.disableAgentListHandler),
		httpserver.NewAction("PUT", "/agentsettings/taint", nil, s.taintAgentsHandler),
		httpserver.NewAction("PUT", "/agentsettings/extendedresource", nil, s.updateExtendedResourcesHandler),
		/*================= agentsetting ====================*/

		/*-------------- custom resource deprecated from 1.15.x -----------------*/
		httpserver.NewAction("POST", "/crr/register", nil, s.registerCustomResourceHander),
		httpserver.NewAction("POST", "/crd/namespaces/{ns}/{kind}", nil, s.createCustomResourceHander),
		httpserver.NewAction("PUT", "/crd/namespaces/{ns}/{kind}", nil, s.updateCustomResourceHander),
		httpserver.NewAction("DELETE", "/crd/namespaces/{ns}/{kind}/{name}", nil, s.deleteCustomResourceHander),
		httpserver.NewAction("GET", "/crd/namespaces/{ns}/{kind}/{name}", nil, s.getCustomResourceHander),
		httpserver.NewAction("GET", "/crd/namespaces/{ns}/{kind}", nil, s.listCustomResourceHander),
		httpserver.NewAction("GET", "/crd/{kind}", nil, s.listAllCustomResourceHander),

		/*-------------- image -----------------*/
		httpserver.NewAction("POST", "/image/commit/{taskgroup}", nil, s.commitImageHander),
		/*-------------- image -----------------*/

		/*------------- definition --------------------*/
		httpserver.NewAction("GET", "/definition/application/{ns}/{name}", nil, s.getApplicationDefHander),
		httpserver.NewAction("GET", "/definition/deployment/{ns}/{name}", nil, s.getDeploymentDefHander),
		/*------------- definition --------------------*/

		/*================= command ====================*/
		httpserver.NewAction("POST", "/command/application/{ns}/{name}", nil, s.sendApplicationCommandHandler),
		httpserver.NewAction("GET", "/command/application/{ns}/{name}", nil, s.getApplicationCommandHandler),
		httpserver.NewAction("DELETE", "/command/application/{ns}/{name}", nil, s.deleteApplicationCommandHandler),
		httpserver.NewAction("POST", "/command/deployment/{ns}/{name}", nil, s.sendDeploymentCommandHandler),
		httpserver.NewAction("GET", "/command/deployment/{ns}/{name}", nil, s.getDeploymentCommandHandler),
		httpserver.NewAction("DELETE", "/command/deployment/{ns}/{name}", nil, s.deleteDeploymentCommandHandler),
		/*================= command ====================*/

		/*================= admissionwebhook ====================*/
		httpserver.NewAction("POST", "/namespaces/{ns}/admissionwebhook", nil, s.CreateAdmissionwebhookHandler),
		httpserver.NewAction("PUT", "/namespaces/{ns}/admissionwebhook", nil, s.UpdateAdmissionwebhookHandler),
		httpserver.NewAction("DELETE", "/namespaces/{ns}/admissionwebhook/{name}", nil, s.DeleteAdmissionwebhookHandler),
		httpserver.NewAction("GET", "/namespaces/{ns}/admissionwebhook/{name}", nil, s.FetchAdmissionwebhookHandler),
		httpserver.NewAction("GET", "/admissionwebhooks", nil, s.FetchAllAdmissionwebhooksHandler),
		/*================= admissionwebhook ====================*/

		//*-------------- mesos webconsole proxy-----------------*//
		httpserver.NewAction("GET", "/webconsole/{uri:*}", nil, s.webconsoleForwarding),
		//httpserver.NewAction("DELETE", "/webconsole/{uri:*}", nil, s.webconsoleForwarding),
		//httpserver.NewAction("PUT", "/webconsole/{uri:*}", nil, s.webconsoleForwarding),
		httpserver.NewAction("POST", "/webconsole/{uri:*}", nil, s.webconsoleForwarding),
		//*-------------- mesos webconsole proxy-----------------*//
	}
	//custom resource solution that compatible with k8s & mesos
	if s.config.KubeConfig != "" {
		crd := []*httpserver.Action{
			/*-------------- custom resource definition implementation-----------------*/
			httpserver.NewAction("GET", "/customresourcedefinitions/{name}", nil, s.customResourceDefinitionForwarding),
			httpserver.NewAction("DELETE", "/customresourcedefinitions/{name}", nil, s.customResourceDefinitionForwarding),
			httpserver.NewAction("PUT", "/customresourcedefinitions/{name}", nil, s.customResourceDefinitionForwarding),
			httpserver.NewAction("GET", "/customresourcedefinitions", nil, s.customResourceDefinitionForwarding),
			httpserver.NewAction("POST", "/customresourcedefinitions", nil, s.customResourceDefinitionForwarding),

			/*-------------- custom resource implementation-----------------*/
			httpserver.NewAction("GET", "/customresources/{uri:*}", nil, s.customResourceForwarding),
			httpserver.NewAction("DELETE", "/customresources/{uri:*}", nil, s.customResourceForwarding),
			httpserver.NewAction("PUT", "/customresources/{uri:*}", nil, s.customResourceForwarding),
			httpserver.NewAction("POST", "/customresources/{uri:*}", nil, s.customResourceForwarding),
			//TODO(DeveloperJim): support custom resource watch if needed
			//httpserver.NewAction("PATCH", "/customresources/{uri:*}", nil, s.customResourceForwarding),
		}
		s.acts = append(s.acts, crd...)
	}
}

//GetHost scheduler implementation
func (s *Scheduler) GetHost() string {
	s.rwHost.RLock()
	defer s.rwHost.RUnlock()

	if len(s.hosts) <= 0 {
		return ""
	}

	return s.hosts[0]
}

//SetHost scheduler implementation
func (s *Scheduler) SetHost(hosts []string) {
	s.rwHost.Lock()
	defer s.rwHost.Unlock()

	s.hosts = hosts
}

func (s *Scheduler) getApplicationDefHander(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	blog.V(3).Infof("get definition for application(%s:%s) ", ns, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	url := s.GetHost() + "/v1/definition/application/" + ns + "/" + name
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
	return
}

func (s *Scheduler) getDeploymentDefHander(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	blog.V(3).Infof("get definition for deployment(%s:%s) ", ns, name)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	url := s.GetHost() + "/v1/definition/deployment/" + ns + "/" + name
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) disableAgentHandler(req *restful.Request, resp *restful.Response) {

	IP := req.PathParameter("IP")
	blog.V(3).Infof("disable agent %s", IP)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	url := s.GetHost() + "/v1/agentsetting/" + IP + "/disable"
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.POST(url, nil, nil)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
	return
}

func (s *Scheduler) enableAgentHandler(req *restful.Request, resp *restful.Response) {

	IP := req.PathParameter("IP")
	blog.V(3).Infof("enable agent %s", IP)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	url := s.GetHost() + "/v1/agentsetting/" + IP + "/enable"
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.POST(url, nil, nil)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) getAgentSettingHandler(req *restful.Request, resp *restful.Response) {

	IP := req.PathParameter("IP")
	blog.V(3).Infof("get agent %s setting", IP)

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	url := s.GetHost() + "/v1/agentsetting/" + IP
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) getAgentSettingListHandler(req *restful.Request, resp *restful.Response) {

	blog.V(3).Infof("get agentsettings")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	ips := req.QueryParameter("ips")

	url := s.GetHost() + "/v1/agentsettings?ips=" + ips
	blog.V(3).Infof("get a request to url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) deleteAgentSettingListHandler(req *restful.Request, resp *restful.Response) {

	blog.V(3).Infof("delete agentsettings")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	ips := req.QueryParameter("ips")
	url := s.GetHost() + "/v1/agentsettings/delete?ips=" + ips
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.POST(url, nil, nil)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) setAgentSettingListHandler(req *restful.Request, resp *restful.Response) {

	blog.V(3).Infof("set agentsettings")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	body, _ := s.getRequestInfo(req)

	url := s.GetHost() + "/v1/agentsettings"
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.POST(url, nil, body)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) updateAgentSettingListHandler(req *restful.Request, resp *restful.Response) {

	blog.V(3).Infof("update agentsettings")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	body, _ := s.getRequestInfo(req)

	url := s.GetHost() + "/v1/agentsettings/update"
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.POST(url, nil, body)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) enableAgentListHandler(req *restful.Request, resp *restful.Response) {

	blog.V(3).Infof("enable agentlist")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	ips := req.QueryParameter("ips")
	url := s.GetHost() + "/v1/agentsettings/enable?ips=" + ips
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.POST(url, nil, nil)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) disableAgentListHandler(req *restful.Request, resp *restful.Response) {

	blog.V(3).Infof("disable agentlist")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	ips := req.QueryParameter("ips")

	url := s.GetHost() + "/v1/agentsettings/disable?ips=" + ips
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.POST(url, nil, nil)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) taintAgentsHandler(req *restful.Request, resp *restful.Response) {
	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	body, _ := s.getRequestInfo(req)
	url := s.GetHost() + "/v1/agentsettings/taint"
	blog.Infof("put url(%s) body(%s)", url, string(body))

	reply, err := s.client.PUT(url, nil, body)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) updateExtendedResourcesHandler(req *restful.Request, resp *restful.Response) {
	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	body, _ := s.getRequestInfo(req)
	url := s.GetHost() + "/v1/agentsettings/extendedresource"
	blog.Infof("put url(%s) body(%s)", url, string(body))

	reply, err := s.client.PUT(url, nil, body)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) GetClusterResourcesHandler(req *restful.Request, resp *restful.Response) {

	blog.V(3).Infof("get cluster resources")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	url := s.GetHost() + "/v1/cluster/resources"
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("get request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) GetClusterEndpointsHandler(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("get cluster endpoints")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	url := s.GetHost() + "/v1/cluster/endpoints"
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("get request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) GetClusterCurrentOffersHandler(req *restful.Request, resp *restful.Response) {
	blog.V(3).Infof("get cluster current offers")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	url := s.GetHost() + "/v1/cluster/current/offers"
	blog.V(3).Infof("post a request to url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("get request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) CreateConfigMapHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_CONFIGMAP, body)
	if err != nil {
		blog.Error("fail to create configmap(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.CreateConfigMap(body)
	if err != nil {
		blog.Error("fail to create configmap(%s). reply(%s), err(%s)", string(body), reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) UpdateConfigMapHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_CONFIGMAP, body)
	if err != nil {
		blog.Error("fail to update configmap(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.UpdateConfigMap(body)
	if err != nil {
		blog.Error("fail to update configmap(%s). reply(%s), err(%s)", string(body), reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) DeleteConfigMapHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	reply, err := s.DeleteConfigMap(ns, name)
	if err != nil {
		blog.Error("fail to delete configmap(%s, %s). reply(%s), err(%s)", ns, name, reply, err.Error())
	}
	resp.Write([]byte(reply))
}

func (s *Scheduler) CreateSecretHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_SECRET, body)
	if err != nil {
		blog.Error("fail to create secret(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.CreateSecret(body)
	if err != nil {
		blog.Error("fail to create secret(%s). reply(%s), err(%s)", string(body), reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) UpdateSecretHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_SECRET, body)
	if err != nil {
		blog.Error("fail to update secret(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.UpdateSecret(body)
	if err != nil {
		blog.Error("fail to update secret(%s). reply(%s), err(%s)", string(body), reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) DeleteSecretHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	reply, err := s.DeleteSecret(ns, name)
	if err != nil {
		blog.Error("fail to delete secret(%s, %s). reply(%s), err(%s)", ns, name, reply, err.Error())
	}
	resp.Write([]byte(reply))
}

func (s *Scheduler) CreateServiceHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_SERVICE, body)
	if err != nil {
		blog.Error("fail to create service(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.CreateService(body)
	if err != nil {
		blog.Error("fail to create service(%s). reply(%s), err(%s)", string(body), reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) UpdateServiceHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_SERVICE, body)
	if err != nil {
		blog.Error("fail to update service(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.UpdateService(body)
	if err != nil {
		blog.Error("fail to update service(%s). reply(%s), err(%s)", string(body), reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) DeleteServiceHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	reply, err := s.DeleteService(ns, name)
	if err != nil {
		blog.Error("fail to delete service(%s, %s). reply(%s), err(%s)", ns, name, reply, err.Error())
	}
	resp.Write([]byte(reply))
}

func (s *Scheduler) CreateApplicationHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_APP, body)
	if err != nil {
		blog.Error("fail to create application(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.CreateApplication(body)
	if err != nil {
		blog.Error("fail to create application. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) CreateProcessHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_PROCESS, body)
	if err != nil {
		blog.Error("fail to create process(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.CreateApplication(body)
	if err != nil {
		blog.Error("fail to create process. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) UpdateApplicationHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_APP, body)
	if err != nil {
		blog.Error("fail to update application(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	instances := req.QueryParameter("instances")
	args := req.QueryParameter("args")

	reply, err := s.UpdateApplication(body, instances, args)
	if err != nil {
		blog.Error("fail to update application for instances(%d). reply(%s), err(%s)", instances, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) UpdateProcessHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_PROCESS, body)
	if err != nil {
		blog.Error("fail to update process(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	instances := req.QueryParameter("instances")
	args := req.QueryParameter("args")

	reply, err := s.UpdateApplication(body, instances, args)
	if err != nil {
		blog.Error("fail to update process for instances(%d). reply(%s), err(%s)", instances, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) DeleteApplicationHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	enforce := req.QueryParameter("enforce")

	reply, err := s.DeleteApplication(ns, name, enforce, commonTypes.BcsDataType_APP)
	if err != nil {
		blog.Error("fail to delete application. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) DeleteProcessHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	enforce := req.QueryParameter("enforce")

	reply, err := s.DeleteApplication(ns, name, enforce, commonTypes.BcsDataType_PROCESS)
	if err != nil {
		blog.Error("fail to delete process. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) DeleteApplicationTaskGroupsHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.DeleteApplicationTaskGroups(body)
	if err != nil {
		blog.Error("fail to delete application taskgroups. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) DeleteApplicationTaskGroupHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.DeleteApplicationTaskGroup(body)
	if err != nil {
		blog.Error("fail to delete application taskgroups. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) RollbackApplicationHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_APP, body)
	if err != nil {
		blog.Error("fail to rollback application(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.RollbackApplication(body, types.BcsDataType_APP)
	if err != nil {
		blog.Error("fail to rollback application. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) RollbackProcessHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_PROCESS, body)
	if err != nil {
		blog.Error("fail to rollback process(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.RollbackApplication(body, types.BcsDataType_PROCESS)
	if err != nil {
		blog.Error("fail to rollback process. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) ScaleApplicationHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	instances := req.PathParameter("instances")

	reply, err := s.ScaleApplication(ns, name, instances, types.BcsDataType_APP)
	if err != nil {
		blog.Error("fail to scale application. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) ScaleProcessHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	instances := req.PathParameter("instances")

	reply, err := s.ScaleApplication(ns, name, instances, types.BcsDataType_PROCESS)
	if err != nil {
		blog.Error("fail to scale process. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) SendMessageApplicationHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	reply, err := s.SendMessageApplication(ns, name, "", body)
	if err != nil {
		blog.Error("fail to send message application. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) SendMessageTaskgroupHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	taskgroupId := req.PathParameter("taskgroup-name")

	reply, err := s.SendMessageApplication(ns, name, taskgroupId, body)
	if err != nil {
		blog.Error("fail to send message application. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) ListApplicationsHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")

	reply, err := s.ListApplications(ns, types.BcsDataType_APP)
	if err != nil {
		blog.Error("fail to list applications. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) ListProcessesHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")

	reply, err := s.ListApplications(ns, types.BcsDataType_PROCESS)
	if err != nil {
		blog.Error("fail to list processes. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) ListApplicationTasksHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	reply, err := s.ListApplicationTasks(ns, name)
	if err != nil {
		blog.Error("fail to list application tasks. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) ListApplicationTaskGroupsHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	reply, err := s.ListApplicationTaskGroups(ns, name)
	if err != nil {
		blog.Error("fail to list application taskgroups. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) ListApplicationVersionsHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	reply, err := s.ListApplicationVersions(ns, name)
	if err != nil {
		blog.Error("fail to list application versions. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) FetchApplicationHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	reply, err := s.FetchApplication(ns, name, types.BcsDataType_APP)
	if err != nil {
		blog.Error("fail to list application versions. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) FetchProcessHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	reply, err := s.FetchApplication(ns, name, types.BcsDataType_PROCESS)
	if err != nil {
		blog.Error("fail to list application versions. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

//FetchApplicationVersionHandler get Application information
func (s *Scheduler) FetchApplicationVersionHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	versionid := req.PathParameter("versionid")

	reply, err := s.FetchApplicationVersion(ns, name, versionid)
	if err != nil {
		blog.Error("fail to list application versions. reply(%s), err(%s)", reply, err.Error())
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) getRequestInfo(req *restful.Request) ([]byte, error) {
	appid := req.PathParameter("appid")
	blog.V(3).Infof("recv a request from app(%s), url(%s)", appid, req.Request.RequestURI)
	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		blog.Error("fail to read request body. err:%s", err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpReadReqBody, common.BcsErrCommHttpReadReqBodyStr)
	}

	return body, err
}

func (s *Scheduler) createDeploymentHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_DEPLOYMENT, body)
	if err != nil {
		blog.Error("fail to create deployment(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.CreateDeployment(body)
	if err != nil {
		blog.Error("fail to create deployment. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) udpateDeploymentHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_DEPLOYMENT, body)
	if err != nil {
		blog.Error("fail to udpate deployment(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	args := req.QueryParameter("args")
	reply, err := s.UpdateDeployment(body, args)
	if err != nil {
		blog.Error("fail to create deployment. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) deleteDeploymentHandler(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	enforce := req.QueryParameter("enforce")
	reply, err := s.deleteDeployment(ns, name, enforce)
	if err != nil {
		blog.Error("fail to delete deployment namespace %s name %s. reply(%s), err(%s)", ns, name, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) cancelupdateDeploymentHandler(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	reply, err := s.cancelupdateDeployment(ns, name)
	if err != nil {
		blog.Error("fail to cancelupdate deployment namespace %s name %s. reply(%s), err(%s)", ns, name, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) pauseupdateDeploymentHandler(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	reply, err := s.pauseupdateDeployment(ns, name)
	if err != nil {
		blog.Error("fail to pauseupdate deployment namespace %s name %s. reply(%s), err(%s)", ns, name, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) resumeupdateDeploymentHandler(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	reply, err := s.resumeupdateDeployment(ns, name)
	if err != nil {
		blog.Error("fail to resumeupdate deployment namespace %s name %s. reply(%s), err(%s)", ns, name, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) scaleDeploymentHandler(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	ins := req.PathParameter("instances")

	instances, err := strconv.Atoi(ins)
	if err != nil {
		blog.Error("fail to scale deployment namespace %s name %s instances %s is invalid", ns, name, ins)
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.scaleDeployment(ns, name, instances)
	if err != nil {
		blog.Error("fail to resumeupdate deployment namespace %s name %s. reply(%s), err(%s)", ns, name, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) reschedulerTaskgroupHandler(req *restful.Request, resp *restful.Response) {
	taskgroupId := req.PathParameter("taskgroupId")
	hostRetainTime := req.QueryParameter("hostRetainTime")

	reply, err := s.RescheduleTaskgroup(taskgroupId, hostRetainTime)
	if err != nil {
		blog.Error("fail to rescheduler taskgroup (%s) reply(%s), err(%s)", taskgroupId, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) restartTaskGroupHandler(req *restful.Request, resp *restful.Response) {
	taskGroupID := req.PathParameter("taskGroupID")

	reply, err := s.RestartTaskGroup(taskGroupID)
	if err != nil {
		blog.Error("fail to restart taskGroup (%s) reply(%s), err(%s)", taskGroupID, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) reloadTaskGroupHandler(req *restful.Request, resp *restful.Response) {
	taskGroupID := req.PathParameter("taskGroupID")

	reply, err := s.ReloadTaskGroup(taskGroupID)
	if err != nil {
		blog.Error("fail to reload taskGroup (%s) reply(%s), err(%s)", taskGroupID, reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) registerCustomResourceHander(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.RegisterCustomResource(body)
	if err != nil {
		blog.Error("fail to register custom resource. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) createCustomResourceHander(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	ns := req.PathParameter("ns")
	kind := req.PathParameter("kind")

	reply, err := s.CreateCustomResource(ns, kind, body)
	if err != nil {
		blog.Error("fail to create custom resource. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) updateCustomResourceHander(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	ns := req.PathParameter("ns")
	kind := req.PathParameter("kind")

	reply, err := s.UpdateCustomResource(ns, kind, body)
	if err != nil {
		blog.Error("fail to update custom resource. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) deleteCustomResourceHander(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("ns")
	kind := req.PathParameter("kind")
	name := req.PathParameter("name")

	reply, err := s.DeleteCustomResource(ns, kind, name)
	if err != nil {
		blog.Error("fail to delete custom resource. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) listCustomResourceHander(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("ns")
	kind := req.PathParameter("kind")

	reply, err := s.ListCustomResource(ns, kind)
	if err != nil {
		blog.Error("fail to list custom resource. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) listAllCustomResourceHander(req *restful.Request, resp *restful.Response) {
	kind := req.PathParameter("kind")

	reply, err := s.ListAllCustomResource(kind)
	if err != nil {
		blog.Error("fail to list custom resource. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) getCustomResourceHander(req *restful.Request, resp *restful.Response) {
	ns := req.PathParameter("ns")
	kind := req.PathParameter("kind")
	name := req.PathParameter("name")

	reply, err := s.GetCustomResource(ns, kind, name)
	if err != nil {
		blog.Error("fail to get custom resource. reply(%s), err(%s)", reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) commitImageHander(req *restful.Request, resp *restful.Response) {

	blog.V(3).Infof("receive commit image")

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	taskgroup := req.PathParameter("taskgroup")
	image := req.QueryParameter("image")
	commit_url := req.QueryParameter("url")

	url := s.GetHost() + "/v1/image/commit/" + taskgroup + "?image=" + image + "&url=" + commit_url
	blog.Infof("post a request to url(%s)", url)

	reply, err := s.client.POST(url, nil, nil)
	if err != nil {
		blog.Error("request to url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) sendApplicationCommandHandler(req *restful.Request, resp *restful.Response) {

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}
	url := s.GetHost() + "/v1/command/application/" + ns + "/" + name

	blog.Info("post url(%s), request(%s)", url, string(body))

	rpyPost, rpyError := s.client.POST(url, nil, body)
	if rpyError != nil {
		blog.Error("post url(%s) failed! err(%s)", url, rpyError.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+rpyError.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(rpyPost))
}

func (s *Scheduler) getApplicationCommandHandler(req *restful.Request, resp *restful.Response) {

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	id := req.QueryParameter("id")
	url := s.GetHost() + "/v1/command/application/" + ns + "/" + name + "?id=" + id
	blog.V(3).Infof("get url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("get url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) deleteApplicationCommandHandler(req *restful.Request, resp *restful.Response) {
	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")

	id := req.QueryParameter("id")
	url := s.GetHost() + "/v1/command/application/" + ns + "/" + name + "?id=" + id
	blog.V(3).Infof("delete url(%s)", url)

	reply, err := s.client.DELETE(url, nil, nil)
	if err != nil {
		blog.Error("delete url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) sendDeploymentCommandHandler(req *restful.Request, resp *restful.Response) {

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}
	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}
	url := s.GetHost() + "/v1/command/deployment/" + ns + "/" + name

	blog.Info("post url(%s), request(%s)", url, string(body))

	rpyPost, rpyError := s.client.POST(url, nil, body)
	if rpyError != nil {
		blog.Error("post url(%s) failed! err(%s)", url, rpyError.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+rpyError.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(rpyPost))
}

func (s *Scheduler) getDeploymentCommandHandler(req *restful.Request, resp *restful.Response) {

	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	id := req.QueryParameter("id")
	url := s.GetHost() + "/v1/command/deployment/" + ns + "/" + name + "?id=" + id
	blog.V(3).Infof("get url(%s)", url)

	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Error("get url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

func (s *Scheduler) deleteDeploymentCommandHandler(req *restful.Request, resp *restful.Response) {
	if s.GetHost() == "" {
		blog.Error("no scheduler is connected by driver")
		err := bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+"scheduler not exist")
		resp.Write([]byte(err.Error()))
		return
	}
	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	id := req.QueryParameter("id")
	url := s.GetHost() + "/v1/command/deployment/" + ns + "/" + name + "?id=" + id
	blog.V(3).Infof("delete url(%s)", url)

	reply, err := s.client.DELETE(url, nil, nil)
	if err != nil {
		blog.Error("delete url(%s) failed! err(%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

//CreateAdmissionwebhookHandler create Admissionwebhook implementation
func (s *Scheduler) CreateAdmissionwebhookHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_Admissionwebhook, body)
	if err != nil {
		blog.Error("fail to create admissionwebhook(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.CreateAdmissionWebhook(body)
	if err != nil {
		blog.Error("fail to create admissionwebhook(%s). reply(%s), err(%s)", string(body), reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

//UpdateAdmissionwebhookHandler update Admissionwebhook
func (s *Scheduler) UpdateAdmissionwebhookHandler(req *restful.Request, resp *restful.Response) {
	body, err := s.getRequestInfo(req)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	err = util.CheckKind(types.BcsDataType_Admissionwebhook, body)
	if err != nil {
		blog.Error("fail to update admissionwebhook(%s). err(%s)", string(body), err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	reply, err := s.UpdateAdmissionWebhook(body)
	if err != nil {
		blog.Error("fail to update admissionwebhook(%s). reply(%s), err(%s)", string(body), reply, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}

	resp.Write([]byte(reply))
}

//DeleteAdmissionwebhookHandler delete Admissionwebhook implementation
func (s *Scheduler) DeleteAdmissionwebhookHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	reply, err := s.DeleteAdmissionWebhook(ns, name)
	if err != nil {
		blog.Error("fail to delete admissionwebhook(%s, %s). reply(%s), err(%s)", ns, name, reply, err.Error())
	}
	resp.Write([]byte(reply))
}

//FetchAdmissionwebhookHandler get specified admission webhook
func (s *Scheduler) FetchAdmissionwebhookHandler(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("ns")
	name := req.PathParameter("name")
	reply, err := s.FetchAdmissionWebhook(ns, name)
	if err != nil {
		blog.Error("fail to fetch admissionwebhook(%s, %s). reply(%s), err(%s)", ns, name, reply, err.Error())
	}
	resp.Write([]byte(reply))
}

//FetchAllAdmissionwebhooksHandler get all admission webhook request
func (s *Scheduler) FetchAllAdmissionwebhooksHandler(req *restful.Request, resp *restful.Response) {
	reply, err := s.FetchAllAdmissionWebhooks()
	if err != nil {
		blog.Error("fail to fetch all admissionwebhooks. reply(%s), err(%s)", reply, err.Error())
	}
	resp.Write([]byte(reply))
}
