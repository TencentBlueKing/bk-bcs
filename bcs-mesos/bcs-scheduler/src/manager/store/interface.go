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

package store

import (
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
)

// The interface for object storage
type Store interface {

	// save framework id to db
	SaveFrameworkID(string) error
	// fetch framework id from db
	FetchFrameworkID() (string, error)
	// check if framework id is in db or not
	HasFrameworkID() (bool, error)

	// save application to db
	SaveApplication(*types.Application) error
	// fetch application from db
	FetchApplication(string, string) (*types.Application, error)
	// list all applications under a namepace
	ListApplications(string) ([]*types.Application, error)
	// delete application from db
	DeleteApplication(string, string) error
	// list namespaces
	ListRunAs() ([]string, error)
	// list application nodes
	ListApplicationNodes(runAs string) ([]string, error)
	// lock application
	LockApplication(appID string)
	// unLock application
	UnLockApplication(appID string)
	// init lock pool
	InitLockPool()
	// list all applications
	ListAllApplications() ([]*types.Application, error)

	// save task to db
	SaveTask(*types.Task) error
	// list all tasks belong to a application by namespace and application name
	ListTasks(string, string) ([]*types.Task, error)
	// fetch task from db
	FetchTask(string) (*types.Task, error)
	// delete task from db
	DeleteTask(string) error

	// save version to db
	SaveVersion(*types.Version) error
	// list all versions
	ListVersions(string, string) ([]string, error)
	// fetch version from db by version id
	FetchVersion(string, string, string) (*types.Version, error)
	// delete version from db
	DeleteVersion(string, string, string) error
	// delete version's root node
	DeleteVersionNode(runAs, versionId string) error
	// get version from db
	GetVersion(runAs, appId string) (*types.Version, error)
	//update version, current only migrate tool use it
	UpdateVersion(version *types.Version) error

	// save taskgroup to db
	SaveTaskGroup(*types.TaskGroup) error
	// list taskgroups under namespace,appid
	ListTaskGroups(string, string) ([]*types.TaskGroup, error)
	// fetch taskgroup
	FetchTaskGroup(string) (*types.TaskGroup, error)
	FetchDBTaskGroup(string) (*types.TaskGroup, error)
	// delete taskgroup by taskgroup id
	DeleteTaskGroup(string) error
	// delete taskgroup by appID
	//DeleteApplicationTaskGroups(string, string) error
	//FetchTaskgroupByIndex(string, string, int) (*types.TaskGroup, error)
	//GetApplicationRootPath() string

	// save agent
	SaveAgent(agent *types.Agent) error
	// fetch agent
	FetchAgent(Key string) (*types.Agent, error)
	// list agents
	ListAgentNodes() ([]string, error)
	// delete agent
	DeleteAgent(key string) error
	// list all agent
	ListAllAgents() ([]*types.Agent, error)

	// save agentsetting
	SaveAgentSetting(*commtypes.BcsClusterAgentSetting) error
	// fetch agentsetting
	FetchAgentSetting(string) (*commtypes.BcsClusterAgentSetting, error)
	// delete agentsetting
	DeleteAgentSetting(string) error
	// list agentsetting
	ListAgentSettingNodes() ([]string, error)

	// save agentschedinfo
	SaveAgentSchedInfo(*types.AgentSchedInfo) error
	// fetch agentschedinfo
	FetchAgentSchedInfo(string) (*types.AgentSchedInfo, error)
	// delete agentschedinfo
	DeleteAgentSchedInfo(string) error

	// save configmap
	SaveConfigMap(configmap *commtypes.BcsConfigMap) error
	// fetch configmap
	FetchConfigMap(ns, name string) (*commtypes.BcsConfigMap, error)
	// delete configmap
	DeleteConfigMap(ns, name string) error
	// list ns configmap
	ListConfigmaps(runAs string) ([]*commtypes.BcsConfigMap, error)
	// list all configmap
	ListAllConfigmaps() ([]*commtypes.BcsConfigMap, error)

	// save secret
	SaveSecret(secret *commtypes.BcsSecret) error
	// fetch secret
	FetchSecret(ns, name string) (*commtypes.BcsSecret, error)
	// delete secret
	DeleteSecret(ns, name string) error
	// list ns secret
	ListSecrets(runAs string) ([]*commtypes.BcsSecret, error)
	// list all secret
	ListAllSecrets() ([]*commtypes.BcsSecret, error)

	// save service
	SaveService(service *commtypes.BcsService) error
	// fetch service
	FetchService(ns, name string) (*commtypes.BcsService, error)
	// delete service
	DeleteService(ns, name string) error
	// list service by namespace
	ListServices(runAs string) ([]*commtypes.BcsService, error)
	// list all services
	ListAllServices() ([]*commtypes.BcsService, error)

	// save endpoint
	SaveEndpoint(endpoint *commtypes.BcsEndpoint) error
	// fetch endpoint
	FetchEndpoint(ns, name string) (*commtypes.BcsEndpoint, error)
	// delete endpoint
	DeleteEndpoint(ns, name string) error

	// save deployment
	SaveDeployment(deployment *types.Deployment) error
	// fetch deployment
	FetchDeployment(ns, name string) (*types.Deployment, error)
	// list deployments
	ListDeployments(ns string) ([]*types.Deployment, error)
	// delete deployment
	DeleteDeployment(ns, name string) error
	// list all deployment
	ListAllDeployments() ([]*types.Deployment, error)

	// init deployments lock pool
	InitDeploymentLockPool()
	// list namespaces for deployments
	ListDeploymentRunAs() ([]string, error)
	// list deployment nodes
	ListDeploymentNodes(runAs string) ([]string, error)
	// lock a deployment
	LockDeployment(deploymentName string)
	// unlock a deployment
	UnLockDeployment(deploymentName string)

	// init cache manager
	InitCacheMgr(bool) error
	// uninit cache manager
	UnInitCacheMgr() error

	//save custom resource register
	SaveCustomResourceRegister(*commtypes.Crr) error

	//delete custom resource register
	//para1: crr.spec.names.kind
	DeleteCustomResourceRegister(string) error

	//fetch custom resource register list
	ListCustomResourceRegister() ([]*commtypes.Crr, error)

	//list all crds
	ListAllCrds(kind string) ([]*commtypes.Crd, error)

	//save custom resource definition
	SaveCustomResourceDefinition(*commtypes.Crd) error

	//delete custom resource definition
	//para1: crd.kind
	//para2: namespace
	//para3: name
	DeleteCustomResourceDefinition(string, string, string) error

	// init command lock
	InitCmdLockPool()
	//lock command by command_id
	LockCommand(cmdId string)
	//unlock command by command_id
	UnLockCommand(cmdId string)
	// save command
	SaveCommand(command *commtypes.BcsCommandInfo) error
	// fetch command
	FetchCommand(ID string) (*commtypes.BcsCommandInfo, error)
	// delete command
	DeleteCommand(ID string) error
	//fetch custom resource definition list
	//para1: crd.kind
	//para2: namespace
	ListCustomResourceDefinition(kind, ns string) ([]*commtypes.Crd, error)

	//fetch custom resource definition
	//para1: crd.kind
	//para2: namespace
	//para3: name
	FetchCustomResourceDefinition(kind, ns, name string) (*commtypes.Crd, error)

	/*=========AdmissionWebhook==========*/
	SaveAdmissionWebhook(admission *commtypes.AdmissionWebhookConfiguration) error
	FetchAdmissionWebhook(ns, name string) (*commtypes.AdmissionWebhookConfiguration, error)
	DeleteAdmissionWebhook(ns, name string) error
	FetchAllAdmissionWebhooks() ([]*commtypes.AdmissionWebhookConfiguration, error)
	/*=========AdmissionWebhook==========*/

	//list object namespaces, object = applicationNode、versionNode...
	//ListObjectNamespaces(objectNode string) ([]string, error)

	//start metrics
	StartStoreObjectMetrics()
	//stop metrics
	StopStoreMetrics()
}

// The interface for db operations
type Dbdrvier interface {
	// Connect
	Connect() error
	// save data to db
	Insert(string, string) error
	// fetch data from db
	Fetch(string) ([]byte, error)
	// update data to db
	Update(string, string) error
	// delete data from db
	Delete(string) error
	// list the key of the data from db
	List(string) ([]string, error)
}
