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
Package backend provides api route's backend function implements.

	//NewBackend return interface Backend object
	backend := NewBackend(sched,store)

	router := NewRouter(backend)
	actions := router.GetActions()

	//use the routing table information to register http client
	//...
*/
package backend

import (
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"
)

type Backend interface {
	// get mesos cluster'id, example for BCS-TESTBCSTEST01-10001
	ClusterId() string

	//save application in db, example for zk
	SaveApplication(*types.Application) error

	//save application version in db
	//first parameter is application's namespace
	//second parameter is application's appid
	SaveVersion(string, string, *types.Version) error

	//get a specific application version
	//first parameter is application's namespace
	//second parameter is application's appid
	GetVersion(string, string) (*types.Version, error)

	//check whether version is valid
	//if version is valid, return nil
	//else return error()
	CheckVersion(*types.Version) error

	//launch application
	//the function is asynchronous
	//so you have to see the launch progress by getting the status
	//of the application
	LaunchApplication(*types.Version) error

	//DeleteApplication will delete all data associated with application.
	//including docker container. the function is asynchronous.
	//the application object is deleted until the actual deletion of docker
	//is successful.
	//first parameter is namespace, second parameter is appid
	//the third one is whether force to delete the application.
	//if true, then force to delete
	//if the slave lost(perhaps due to host down or network problem), then
	//need force to delete application and the admin must be sure of docker
	//container is down
	//the fourth param is kind, tell which type of application is going to be deleted, application or process. it
	//is no allowed to delete a different kind application.
	DeleteApplication(string, string, bool, commtypes.BcsDataType) error

	//list all applications of a specific namespace
	//parameter is namespace
	ListApplications(string) ([]*types.Application, error)

	//fetch a specific application
	//first para is namespace, and second one is appid
	FetchApplication(string, string) (*types.Application, error)

	//list all taskgroups of a specific application
	//first para is namespace, and second one is appid
	ListApplicationTaskGroups(string, string) ([]*types.TaskGroup, error)

	//list all tasks of a specific application
	//first para is namespace, and second one is appid
	//ListApplicationTasks(string, string) ([]*types.Task, error)

	//list all versions of application
	//first para is namespace, and second one is appid
	ListApplicationVersions(string, string) ([]string, error)

	//fetch a specific application version
	//first para is namespace, second one is appid and the third is versionID
	FetchApplicationVersion(string, string, string) (*types.Version, error)

	//update application
	//first para is namespace, second one is appid
	//third one is the number of updated docker containers
	//the last is the new version of application
	UpdateApplication(string, string, string, int, *types.Version) error

	//scale up or down application
	//first para is namespace, second one is appid,
	//third one is the final container number of applications.
	//the number must be >= 1.
	//the fourth param is kind, tell which type of application is going to be scale, application or process. it
	//is no allowed to scale a different kind application.
	//the fifth param is bool, means if the caller is from API, it depends whether we should check the
	//request kind of scale or not. If true then do the check.
	ScaleApplication(string, string, uint64, commtypes.BcsDataType, bool) error

	//send message to a specific application
	//first para is namespace, second one is appid
	//third one is message
	SendToApplication(string, string, *types.BcsMessage) ([]*types.TaskGroupOpResult, []*types.TaskGroupOpResult, error)

	//send message to a specific application
	//first para is namespace, second one is appid
	//third one is taskgroupid, the last one is message
	SendToApplicationTaskGroup(string, string, string, *types.BcsMessage) error

	//create configmap
	SaveConfigMap(configmap *commtypes.BcsConfigMap) error

	//fetch a specific configmap, ns is namespace, name is configmap's name
	FetchConfigMap(ns, name string) (*commtypes.BcsConfigMap, error)

	//delete configmap, ns is namespace, name is configmap's name
	DeleteConfigMap(ns string, name string) error

	//create secret
	SaveSecret(secret *commtypes.BcsSecret) error

	//fetch secret, ns is namespace, name is secret's name
	FetchSecret(ns, name string) (*commtypes.BcsSecret, error)

	//delete secret, ns is namespace, name is secret's name
	DeleteSecret(ns string, name string) error

	//create service
	SaveService(service *commtypes.BcsService) error

	//fetch service, ns is namespace, name is service's name
	FetchService(ns, name string) (*commtypes.BcsService, error)

	//delete service, ns is namespace, name is service's name
	DeleteService(ns string, name string) error

	//get schedueler's role, 'master' or 'slave'
	GetRole() string

	//get mesos cluster's resources, including cpu, port, dis, port
	GetClusterResources() (*commtypes.BcsClusterResource, error)

	//get cluster's endpoints, including scheduler, mesos master
	GetClusterEndpoints() *commtypes.ClusterEndpoints

	//deployment controller providers declarative updates for applications.
	//you describe a desired state in a Deployment object,
	//and the Deployment controller changes the actual state to the desired state at a controlled rate.
	//you can define Deployments to create new application
	CreateDeployment(*types.DeploymentDef) (int, error)

	//get deployment
	GetDeployment(string, string) (*types.Deployment, error)

	//rolling update deployment's application
	UpdateDeployment(*types.DeploymentDef) (int, error)

	//cancel update deployment, and rollback the application
	//first para is namespace, second one is deployment's name
	CancelUpdateDeployment(string, string) error

	//pause udpate deployment
	//first para is namespace, second one is deployment's name
	PauseUpdateDeployment(string, string) error

	//resume a paused updated deployment
	//first para is namespace, second one is deployment's name
	ResumeUpdateDeployment(string, string) error

	//delete deployment, include the associated application
	//first para is namespace, second one is deployment's name
	//third one is whether force to delete the deployment
	//the meaming is similar to application
	DeleteDeployment(string, string, bool) (int, error)

	//scale up or down deployment
	//first para is namespace, second one is appid,
	//third one is the final container number of deployment.
	//the number must be >= 1.
	ScaleDeployment(string, string, uint64) error

	//healthy report
	HealthyReport(*commtypes.HealthCheckResult)

	//rescheduler taskgroup
	//will be deleting the old taskgroup, and launch new taskgroup
	//in any suitable physical machine.
	//para is taskgroupid
	RescheduleTaskgroup(string, int64) error

	//query user custom mesos slave attributes
	//para is slave ip
	QueryAgentSetting(string) (*commtypes.BcsClusterAgentSetting, error)

	//disalbe mesos slave
	//if disable, then will be not launch new taskgroup in the slave,
	//but don't delete the already existing containers
	//para is slave ip
	DisableAgent(string) error

	//enable agent, para is slave ip
	EnableAgent(string) error

	//query user custom mesos slaves attributes
	//para is array of slave ip
	QueryAgentSettingList([]string) ([]*commtypes.BcsClusterAgentSetting, int, error)

	//delete user custom mesos slaves attributes
	//para is array of slave ip
	//DeleteAgentSettingList([]string) (int, error)

	//set user custom mesos slaves attributes
	SetAgentSettingList([]*commtypes.BcsClusterAgentSetting) (int, error)

	//disable mesos slaves
	//para is array of slave ip
	DisableAgentList(IPs []string) (int, error)

	//enable mesos slaves
	//para is array of slave ip
	EnableAgentList(IPs []string) (int, error)

	//update user custom mesos slaves attributes
	//UpdateAgentSettingList(*commtypes.BcsClusterAgentSettingUpdate) (int, error)

	//taints agent
	TaintAgents([]*commtypes.BcsClusterAgentSetting) error

	//update agent extenedresources
	UpdateExtendedResources(ex *commtypes.ExtendedResource) error

	//custom resource register
	RegisterCustomResource(*commtypes.Crr) error

	//custom resource unregister
	//para1: crr.spec.names.kind
	UnregisterCustomResource(string) error

	//create custom resource definition
	CreateCustomResource(*commtypes.Crd) error

	//update custom resource definition
	UpdateCustomResource(*commtypes.Crd) error

	//delete custom resource definition
	//para1: crd.kind
	//para2: namespace
	//para3: name
	DeleteCustomResource(string, string, string) error

	//fetch custom resource definition list
	//para1: crd.kind
	//para2: namespace
	ListCustomResourceDefinition(kind, ns string) ([]*commtypes.Crd, error)

	// list all crds
	ListAllCrds(kind string) ([]*commtypes.Crd, error)

	//fetch custom resource definition
	//para1: crd.kind
	//para2: namespace
	//para3: name
	FetchCustomResourceDefinition(kind, ns, name string) (*commtypes.Crd, error)

	//commit task(taskgroup->image) to url
	CommitImage(string, string, string) (*types.BcsMessage, error)

	//get current offers
	GetCurrentOffers() []*mesos.Offer
	// send restart taskGroup command, only for process.
	RestartTaskGroup(taskGroupID string) (*types.BcsMessage, error)

	// send reload taskGroup command, only for process.
	ReloadTaskGroup(taskGroupID string) (*types.BcsMessage, error)

	// get command
	GetCommand(ID string) (*commtypes.BcsCommandInfo, error)
	// delete command
	DeleteCommand(ID string) error
	// do command
	DoCommand(command *commtypes.BcsCommandInfo) error

	/*=========AdmissionWebhook==========*/
	SaveAdmissionWebhook(configmap *commtypes.AdmissionWebhookConfiguration) error
	UpdateAdmissionWebhook(admission *commtypes.AdmissionWebhookConfiguration) error
	FetchAdmissionWebhook(ns, name string) (*commtypes.AdmissionWebhookConfiguration, error)
	DeleteAdmissionWebhook(ns, name string) error
	FetchAllAdmissionWebhooks() ([]*commtypes.AdmissionWebhookConfiguration, error)
	/*=========AdmissionWebhook==========*/
	//launch daemonset
	LaunchDaemonset(def *types.BcsDaemonsetDef) error
	//delete daemonset
	DeleteDaemonset(namespace, name string, force bool) error
}
