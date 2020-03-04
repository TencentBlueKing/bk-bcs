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

/*
Package api provides scheduler http api routes implements.

Including application, deployment, service, configmap, secret.
Please see the api document for details.

	backend := Backend{}
	router := NewRouter(backend)
	actions := router.GetActions()

	//use the routing table information to register http client
	//...
*/
package api

import (
	"bk-bcs/bcs-common/common/http/httpserver"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/backend"
)

type Router struct {
	backend backend.Backend
	actions []*httpserver.Action
}

// NewRouter return api application router
func NewRouter(b backend.Backend) *Router {
	r := &Router{
		backend: b,
		actions: make([]*httpserver.Action, 0),
	}

	r.initRoutes()
	return r
}

//get http api routing table information, and use it to register http client
func (r *Router) GetActions() []*httpserver.Action {
	return r.actions
}

func (r *Router) initRoutes() {
	/*-------------- application ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/apps", nil, r.buildApplication))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/{runAs}/apps", nil, r.listApplications))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/apps/{runAs}/{appId}", nil, r.fetchApplication))
	r.actions = append(r.actions, httpserver.NewAction("DELETE", "/apps/{runAs}/{appId}", nil, r.deleteApplication))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/apps/{runAs}/{appId}/update", nil, r.updateApplication))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/apps/{runAs}/{appId}/scale", nil, r.scaleApplication))
	/*-------------- application ---------------*/

	/*-------------- taskgroup ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("GET", "/apps/{runAs}/{appId}/taskgroups", nil, r.listApplicationTaskGroups))
	r.actions = append(r.actions, httpserver.NewAction("PUT", "/taskgroup/{taskgroupId}/rescheduler", nil, r.reschedulerTaskgroup))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/taskgroup/{taskGroupID}/restart", nil, r.restartTaskGroup))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/taskgroup/{taskGroupID}/reload", nil, r.reloadTaskGroup))
	/*-------------- taskgroup ---------------*/

	/*-------------- task ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("GET", "/apps/{runAs}/{appId}/tasks", nil, r.listApplicationTasks))
	/*-------------- task ---------------*/

	/*-------------- message ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/apps/{runAs}/{appId}/message", nil, r.sendMessageApplication))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/apps/{runAs}/{appId}/message/{taskgroupId}", nil, r.sendMessageApplicationTaskGroup))
	/*-------------- message ---------------*/

	/*-------------- version ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("GET", "/apps/{runAs}/{appId}/versions", nil, r.listApplicationVersions))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/apps/{runAs}/{appId}/versions/{versionId}", nil, r.fetchApplicationVersion_r))
	/*-------------- version ---------------*/

	/*-------------- configmap ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/configmap", nil, r.createConfigMap))
	r.actions = append(r.actions, httpserver.NewAction("PUT", "/configmap", nil, r.updateConfigMap))
	r.actions = append(r.actions, httpserver.NewAction("DELETE", "/configmap/{namespace}/{name}", nil, r.deleteConfigMap))
	/*-------------- configmap ---------------*/

	/*-------------- secret ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/secret", nil, r.createSecret))
	r.actions = append(r.actions, httpserver.NewAction("PUT", "/secret", nil, r.updateSecret))
	r.actions = append(r.actions, httpserver.NewAction("DELETE", "/secret/{namespace}/{name}", nil, r.deleteSecret))
	/*-------------- secret ---------------*/

	/*-------------- service ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/service", nil, r.createService))
	r.actions = append(r.actions, httpserver.NewAction("PUT", "/service", nil, r.updateService))
	r.actions = append(r.actions, httpserver.NewAction("DELETE", "/service/{namespace}/{name}", nil, r.deleteService))
	/*-------------- service ---------------*/

	/*-------------- cluster ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("GET", "/cluster/resources", nil, r.getClusterResources))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/cluster/endpoints", nil, r.getClusterEndpoints))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/cluster/current/offers", nil, r.getCurrentOffers))
	/*-------------- cluster ---------------*/

	/*-------------- deployment ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/deployment/{namespace}/{name}", nil, r.createDeployment))
	r.actions = append(r.actions, httpserver.NewAction("PUT", "/deployment/{namespace}/{name}", nil, r.updateDeployment))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/deployment/{namespace}/{name}/cancelupdate", nil, r.cancelUpdateDeployment))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/deployment/{namespace}/{name}/pauseupdate", nil, r.pauseUpdateDeployment))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/deployment/{namespace}/{name}/resumeupdate", nil, r.resumeUpdateDeployment))
	r.actions = append(r.actions, httpserver.NewAction("DELETE", "/deployment/{namespace}/{name}", nil, r.deleteDeployment))
	r.actions = append(r.actions, httpserver.NewAction("PUT", "/deployment/{namespace}/{name}/scale/{instances}", nil, r.scaleDeployment_r))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/deployment/{namespace}/{name}", nil, r.getDeployment_r))
	/*-------------- deployment ---------------*/

	/*-------------- healthcheck ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/healthcheck", nil, r.healthCheckReport))
	/*-------------- healthcheck ---------------*/

	/*-------------- agent setting ---------------*/
	r.actions = append(r.actions, httpserver.NewAction("GET", "/agentsetting/{IP}", nil, r.queryAgentSetting))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/agentsetting/{IP}/enable", nil, r.enableAgent))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/agentsetting/{IP}/disable", nil, r.disableAgent))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/agentsettings", nil, r.queryAgentSettingList))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/agentsettings/delete", nil, r.deleteAgentSettingList))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/agentsettings", nil, r.setAgentSettingList))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/agentsettings/update", nil, r.updateAgentSettingList))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/agentsettings/enable", nil, r.enableAgentList))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/agentsettings/disable", nil, r.disableAgentList))
	/*-------------- agent setting ---------------*/

	/*-------------- custom resource -----------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/crr/register", nil, r.registerCustomResource))
	r.actions = append(r.actions, httpserver.NewAction("POST", "/crd/namespaces/{ns}/{kind}", nil, r.createCustomResource))
	r.actions = append(r.actions, httpserver.NewAction("PUT", "/crd/namespaces/{ns}/{kind}", nil, r.updateCustomResource))
	r.actions = append(r.actions, httpserver.NewAction("DELETE", "/crd/namespaces/{ns}/{kind}/{name}", nil, r.deleteCustomResource))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/crd/namespaces/{ns}/{kind}", nil, r.listCustomResource))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/crd/namespaces/{ns}/{kind}/{name}", nil, r.getCustomResource))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/crd/{kind}", nil, r.listAllCustomResource))
	/*-------------- custom resource -----------------*/

	/*-------------- image -----------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/image/commit/{taskgroup}", nil, r.commitImage))
	/*-------------- image -----------------*/

	/*------------- definition --------------------*/
	r.actions = append(r.actions, httpserver.NewAction("GET", "/definition/application/{ns}/{name}", nil, r.getApplicationDef))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/definition/deployment/{ns}/{name}", nil, r.getDeploymentDef))
	/*------------- definition --------------------*/

	/*------------- command ---------------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/command/application/{ns}/{name}", nil, r.sendApplicationCommand))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/command/application/{ns}/{name}", nil, r.getApplicationCommand))
	r.actions = append(r.actions, httpserver.NewAction("DELETE", "/command/application/{ns}/{name}", nil, r.deleteApplicationCommand))

	r.actions = append(r.actions, httpserver.NewAction("POST", "/command/deployment/{ns}/{name}", nil, r.sendDeploymentCommand))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/command/deployment/{ns}/{name}", nil, r.getDeploymentCommand))
	r.actions = append(r.actions, httpserver.NewAction("DELETE", "/command/deployment/{ns}/{name}", nil, r.deleteDeploymentCommand))
	/*--------------command ----------------------*/

	/*--------------admissionwebhook ----------------------*/
	r.actions = append(r.actions, httpserver.NewAction("POST", "/admissionwebhook", nil, r.createAdmissionwebhook))
	r.actions = append(r.actions, httpserver.NewAction("PUT", "/admissionwebhook", nil, r.updateAdmissionwebhook))
	r.actions = append(r.actions, httpserver.NewAction("DELETE", "/admissionwebhook/{namespace}/{name}", nil, r.deleteAdmissionwebhook))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/admissionwebhooks", nil, r.fetchAllAdmissionwebhooks))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/admissionwebhook/{namespace}/{name}", nil, r.fetchAdmissionwebhook))
	/*--------------admissionwebhook ----------------------*/
}
