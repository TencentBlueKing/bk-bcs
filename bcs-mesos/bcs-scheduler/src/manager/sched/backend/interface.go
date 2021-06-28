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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

type Backend interface {
	// ClusterId get mesos cluster'id, example for BCS-TESTBCSTEST01-10001
	ClusterId() string

	// SaveApplication save application in db, example for zk
	SaveApplication(*types.Application) error

	// SaveVersion save application version in db
	// first parameter is application's namespace
	// second parameter is application's appid
	SaveVersion(string, string, *types.Version) error

	// GetVersion get a specific application version
	// first parameter is application's namespace
	// second parameter is application's appid
	GetVersion(string, string) (*types.Version, error)

	// CheckVersion check whether version is valid
	// if version is valid, return nil
	// else return error()
	CheckVersion(*types.Version) error

	// LaunchApplication launch application
	//the function is asynchronous
	//so you have to see the launch progress by getting the status
	//of the application
	LaunchApplication(*types.Version) error

	// DeleteApplication will delete all data associated with application.
	// including docker container. the function is asynchronous.
	// the application object is deleted until the actual deletion of docker
	// is successful.
	// first parameter is namespace, second parameter is appid
	// the third one is whether force to delete the application.
	// if true, then force to delete
	// if the slave lost(perhaps due to host down or network problem), then
	// need force to delete application and the admin must be sure of docker
	// container is down
	// the fourth param is kind, tell which type of application is going to be deleted, application or process. it
	// is no allowed to delete a different kind application.
	DeleteApplication(string, string, bool, commtypes.BcsDataType) error

	// ListApplications list all applications of a specific namespace
	// parameter is namespace
	ListApplications(string) ([]*types.Application, error)

	// FetchApplication fetch a specific application
	// first para is namespace, and second one is appid
	FetchApplication(string, string) (*types.Application, error)

	// ListApplicationTaskGroups list all taskgroups of a specific application
	// first para is namespace, and second one is appid
	ListApplicationTaskGroups(string, string) ([]*types.TaskGroup, error)

	// ListApplicationVersions list all versions of application
	// first para is namespace, and second one is appid
	ListApplicationVersions(string, string) ([]string, error)

	// FetchApplicationVersion fetch a specific application version
	// first para is namespace, second one is appid and the third is versionID
	FetchApplicationVersion(string, string, string) (*types.Version, error)

	// UpdateApplication update application
	// first para is namespace, second one is appid
	// third one is the number of updated docker containers
	// the last is the new version of application
	UpdateApplication(string, string, string, int, *types.Version) error

	// ScaleApplication scale up or down application
	// first para is namespace, second one is appid,
	// third one is the final container number of applications.
	// the number must be >= 1.
	// the fourth param is kind, tell which type of application is going to be scale, application or process. it
	// is no allowed to scale a different kind application.
	// the fifth param is bool, means if the caller is from API, it depends whether we should check the
	// request kind of scale or not. If true then do the check.
	ScaleApplication(string, string, uint64, commtypes.BcsDataType, bool) error

	// SendToApplication send message to a specific application
	// first para is namespace, second one is appid
	// third one is message
	SendToApplication(string, string, *types.BcsMessage) ([]*types.TaskGroupOpResult, []*types.TaskGroupOpResult, error)

	// SendToApplicationTaskGroup send message to a specific application
	// first para is namespace, second one is appid
	// third one is taskgroupid, the last one is message
	SendToApplicationTaskGroup(string, string, string, *types.BcsMessage) error

	// SaveConfigMap create configmap
	SaveConfigMap(configmap *commtypes.BcsConfigMap) error

	// FetchConfigMap fetch a specific configmap, ns is namespace, name is configmap's name
	FetchConfigMap(ns, name string) (*commtypes.BcsConfigMap, error)

	// DeleteConfigMap delete configmap, ns is namespace, name is configmap's name
	DeleteConfigMap(ns string, name string) error

	// SaveSecret create secret
	SaveSecret(secret *commtypes.BcsSecret) error

	// FetchSecret fetch secret, ns is namespace, name is secret's name
	FetchSecret(ns, name string) (*commtypes.BcsSecret, error)

	// DeleteSecret delete secret, ns is namespace, name is secret's name
	DeleteSecret(ns string, name string) error

	// SaveService create service
	SaveService(service *commtypes.BcsService) error

	// FetchService fetch service, ns is namespace, name is service's name
	FetchService(ns, name string) (*commtypes.BcsService, error)

	// DeleteService delete service, ns is namespace, name is service's name
	DeleteService(ns string, name string) error

	// GetRole get schedueler's role, 'master' or 'slave'
	GetRole() string

	// GetClusterResources get mesos cluster's resources, including cpu, port, dis, port
	GetClusterResources() (*commtypes.BcsClusterResource, error)

	// GetClusterEndpoints get cluster's endpoints, including scheduler, mesos master
	GetClusterEndpoints() *commtypes.ClusterEndpoints

	// CreateDeployment deployment controller providers declarative updates for applications.
	// you describe a desired state in a Deployment object,
	// and the Deployment controller changes the actual state to the desired state at a controlled rate.
	// you can define Deployments to create new application
	CreateDeployment(*types.DeploymentDef) (int, error)

	// GetDeployment get deployment
	GetDeployment(string, string) (*types.Deployment, error)

	// UpdateDeployment rolling update deployment's application
	UpdateDeployment(*types.DeploymentDef) (int, error)

	// UpdateDeploymentResource update deployment resource only
	UpdateDeploymentResource(*types.DeploymentDef) (int, error)

	// CancelUpdateDeployment cancel update deployment, and rollback the application
	// first para is namespace, second one is deployment's name
	CancelUpdateDeployment(string, string) error

	// PauseUpdateDeployment pause udpate deployment
	// first para is namespace, second one is deployment's name
	PauseUpdateDeployment(string, string) error

	// ResumeUpdateDeployment resume a paused updated deployment
	// first para is namespace, second one is deployment's name
	ResumeUpdateDeployment(string, string) error

	// DeleteDeployment delete deployment, include the associated application
	// first para is namespace, second one is deployment's name
	// third one is whether force to delete the deployment
	// the meaming is similar to application
	DeleteDeployment(string, string, bool) (int, error)

	// ScaleDeployment scale up or down deployment
	// first para is namespace, second one is appid,
	// third one is the final container number of deployment.
	// the number must be >= 1.
	ScaleDeployment(string, string, uint64) error

	// HealthyReport healthy report
	HealthyReport(*commtypes.HealthCheckResult)

	// RescheduleTaskgroup rescheduler taskgroup
	// will be deleting the old taskgroup, and launch new taskgroup
	// in any suitable physical machine.
	// para is taskgroupid
	RescheduleTaskgroup(string, int64) error

	// QueryAgentSetting query user custom mesos slave attributes
	// para is slave ip
	QueryAgentSetting(string) (*commtypes.BcsClusterAgentSetting, error)

	// DisableAgent disalbe mesos slave
	// if disable, then will be not launch new taskgroup in the slave,
	// but don't delete the already existing containers
	// para is slave ip
	DisableAgent(string) error

	// EnableAgent enable agent, para is slave ip
	EnableAgent(string) error

	// QueryAgentSettingList query user custom mesos slaves attributes
	// para is array of slave ip
	QueryAgentSettingList([]string) ([]*commtypes.BcsClusterAgentSetting, int, error)

	// delete user custom mesos slaves attributes
	// para is array of slave ip
	// DeleteAgentSettingList([]string) (int, error)

	// SetAgentSettingList set user custom mesos slaves attributes
	SetAgentSettingList([]*commtypes.BcsClusterAgentSetting) (int, error)

	// DisableAgentList disable mesos slaves
	// para is array of slave ip
	DisableAgentList(IPs []string) (int, error)

	// EnableAgentList enable mesos slaves
	// para is array of slave ip
	EnableAgentList(IPs []string) (int, error)

	// update user custom mesos slaves attributes
	// UpdateAgentSettingList(*commtypes.BcsClusterAgentSettingUpdate) (int, error)

	// TaintAgents taints agent
	TaintAgents([]*commtypes.BcsClusterAgentSetting) error

	// UpdateExtendedResources update agent extenedresources
	UpdateExtendedResources(ex *commtypes.ExtendedResource) error

	// RegisterCustomResource custom resource register
	RegisterCustomResource(*commtypes.Crr) error

	// UnregisterCustomResource custom resource unregister
	// para1: crr.spec.names.kind
	UnregisterCustomResource(string) error

	// CreateCustomResource create custom resource definition
	CreateCustomResource(*commtypes.Crd) error

	// UpdateCustomResource update custom resource definition
	UpdateCustomResource(*commtypes.Crd) error

	// DeleteCustomResource delete custom resource definition
	// para1: crd.kind
	// para2: namespace
	// para3: name
	DeleteCustomResource(string, string, string) error

	// ListCustomResourceDefinition fetch custom resource definition list
	// para1: crd.kind
	// para2: namespace
	ListCustomResourceDefinition(kind, ns string) ([]*commtypes.Crd, error)

	// ListAllCrds list all crds
	ListAllCrds(kind string) ([]*commtypes.Crd, error)

	// FetchCustomResourceDefinition fetch custom resource definition
	// para1: crd.kind
	// para2: namespace
	// para3: name
	FetchCustomResourceDefinition(kind, ns, name string) (*commtypes.Crd, error)

	// CommitImage commit task(taskgroup->image) to url
	CommitImage(string, string, string) (*types.BcsMessage, error)

	// GetCurrentOffers get current offers
	GetCurrentOffers() []*types.OfferWithDelta
	// RestartTaskGroup send restart taskGroup command, only for process.
	RestartTaskGroup(taskGroupID string) (*types.BcsMessage, error)

	// ReloadTaskGroup send reload taskGroup command, only for process.
	ReloadTaskGroup(taskGroupID string) (*types.BcsMessage, error)

	// GetCommand get command
	GetCommand(ID string) (*commtypes.BcsCommandInfo, error)
	// DeleteCommand delete command
	DeleteCommand(ID string) error
	// DoCommand do command
	DoCommand(command *commtypes.BcsCommandInfo) error

	/*=========AdmissionWebhook==========*/
	// SaveAdmissionWebhook save admission webhook to store
	SaveAdmissionWebhook(configmap *commtypes.AdmissionWebhookConfiguration) error
	// UpdateAdmissionWebhook update admission webhook to store
	UpdateAdmissionWebhook(admission *commtypes.AdmissionWebhookConfiguration) error
	// FetchAdmissionWebhook get admission webhook from store
	FetchAdmissionWebhook(ns, name string) (*commtypes.AdmissionWebhookConfiguration, error)
	// DeleteAdmissionWebhook delete admission webhook from store
	DeleteAdmissionWebhook(ns, name string) error
	// FetchAllAdmissionWebhooks get all admission webhook list
	FetchAllAdmissionWebhooks() ([]*commtypes.AdmissionWebhookConfiguration, error)
	/*=========AdmissionWebhook==========*/

	// LaunchDaemonset launch daemonset
	LaunchDaemonset(def *types.BcsDaemonsetDef) error
	// DeleteDaemonset delete daemonset
	DeleteDaemonset(namespace, name string, force bool) error

	/*==========Transaction==============*/
	// ListTransaction list all transactions in one namespace
	ListTransaction(ns string) ([]*types.Transaction, error)
	// DeleteTransaction delete transaction
	DeleteTransaction(transNs, transName string) error
	/*==========Transaction==============*/
}
