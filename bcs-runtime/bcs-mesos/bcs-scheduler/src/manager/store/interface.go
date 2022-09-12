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
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// Store xxx
// The interface for object storage
type Store interface {

	// SaveFrameworkID xxx
	// save framework id to db
	SaveFrameworkID(string) error
	// FetchFrameworkID xxx
	// fetch framework id from db
	FetchFrameworkID() (string, error)
	// HasFrameworkID xxx
	// check if framework id is in db or not
	HasFrameworkID() (bool, error)

	// SaveApplication xxx
	// save application to db
	SaveApplication(*types.Application) error
	// FetchApplication xxx
	// fetch application from db
	FetchApplication(string, string) (*types.Application, error)
	// ListApplications xxx
	// list all applications under a namepace
	ListApplications(string) ([]*types.Application, error)
	// DeleteApplication xxx
	// delete application from db
	DeleteApplication(string, string) error
	// ListRunAs xxx
	// list namespaces
	ListRunAs() ([]string, error)
	// ListApplicationNodes xxx
	// list application nodes
	ListApplicationNodes(runAs string) ([]string, error)
	// LockApplication xxx
	// lock application
	LockApplication(appID string)
	// UnLockApplication xxx
	// unLock application
	UnLockApplication(appID string)
	// InitLockPool xxx
	// init lock pool
	InitLockPool()
	// ListAllApplications xxx
	// list all applications
	ListAllApplications() ([]*types.Application, error)

	// SaveTask xxx
	// save task to db
	SaveTask(*types.Task) error
	// FetchTask xxx
	// list all tasks belong to a application by namespace and application name
	// ListTasks(string, string) ([]*types.Task, error)
	// fetch task from db
	FetchTask(string) (*types.Task, error)
	// DeleteTask xxx
	// delete task from db
	DeleteTask(string) error

	// SaveVersion xxx
	// save version to db
	SaveVersion(*types.Version) error
	// ListVersions xxx
	// list all versions
	ListVersions(string, string) ([]string, error)
	// FetchVersion xxx
	// fetch version from db by version id
	FetchVersion(string, string, string) (*types.Version, error)
	// DeleteVersion xxx
	// delete version from db
	DeleteVersion(string, string, string) error
	// DeleteVersionNode xxx
	// delete version's root node
	DeleteVersionNode(runAs, versionId string) error
	// GetVersion xxx
	// get version from db
	GetVersion(runAs, appId string) (*types.Version, error)
	// UpdateVersion xxx
	// update version, current only migrate tool use it
	UpdateVersion(version *types.Version) error

	// SaveTaskGroup xxx
	// save taskgroup to db
	SaveTaskGroup(*types.TaskGroup) error
	// ListTaskGroups xxx
	// list taskgroups under namespace,appid
	ListTaskGroups(string, string) ([]*types.TaskGroup, error)
	// FetchTaskGroup xxx
	// fetch taskgroup
	FetchTaskGroup(string) (*types.TaskGroup, error)
	FetchDBTaskGroup(string) (*types.TaskGroup, error)
	// DeleteTaskGroup xxx
	// delete taskgroup by taskgroup id
	DeleteTaskGroup(string) error
	// ListClusterTaskgroups xxx
	// list mesos cluster taskgroups, include: application、deployment、daemonset...
	ListClusterTaskgroups() ([]*types.TaskGroup, error)

	// SaveAgent xxx
	// save agent
	SaveAgent(agent *types.Agent) error
	// FetchAgent xxx
	// fetch agent
	FetchAgent(Key string) (*types.Agent, error)
	// ListAgentNodes xxx
	// list agents
	ListAgentNodes() ([]string, error)
	// DeleteAgent xxx
	// delete agent
	DeleteAgent(key string) error
	// ListAllAgents xxx
	// list all agent
	ListAllAgents() ([]*types.Agent, error)

	// SaveAgentSetting xxx
	// save agentsetting
	SaveAgentSetting(*commtypes.BcsClusterAgentSetting) error
	// FetchAgentSetting xxx
	// fetch agentsetting
	FetchAgentSetting(string) (*commtypes.BcsClusterAgentSetting, error)
	// DeleteAgentSetting xxx
	// delete agentsetting
	DeleteAgentSetting(string) error
	// ListAgentSettingNodes xxx
	// list agentsetting
	ListAgentSettingNodes() ([]string, error)
	ListAgentsettings() ([]*commtypes.BcsClusterAgentSetting, error)
	// SaveAgentSchedInfo xxx
	// save agentschedinfo
	SaveAgentSchedInfo(*types.AgentSchedInfo) error
	// FetchAgentSchedInfo xxx
	// fetch agentschedinfo
	FetchAgentSchedInfo(string) (*types.AgentSchedInfo, error)
	// DeleteAgentSchedInfo xxx
	// delete agentschedinfo
	DeleteAgentSchedInfo(string) error
	// ListAgentSchedInfoNodes xxx
	// list agentschedinfo node
	ListAgentSchedInfoNodes() ([]string, error)
	// ListAgentSchedInfo xxx
	// list agentschedinfo
	ListAgentSchedInfo() ([]*types.AgentSchedInfo, error)

	// SaveConfigMap xxx
	// save configmap
	SaveConfigMap(configmap *commtypes.BcsConfigMap) error
	// FetchConfigMap xxx
	// fetch configmap
	FetchConfigMap(ns, name string) (*commtypes.BcsConfigMap, error)
	// DeleteConfigMap xxx
	// delete configmap
	DeleteConfigMap(ns, name string) error
	// ListAllConfigmaps xxx
	// list ns configmap
	// ListConfigmaps(runAs string) ([]*commtypes.BcsConfigMap, error)
	// list all configmap
	ListAllConfigmaps() ([]*commtypes.BcsConfigMap, error)

	// SaveSecret xxx
	// save secret
	SaveSecret(secret *commtypes.BcsSecret) error
	// FetchSecret xxx
	// fetch secret
	FetchSecret(ns, name string) (*commtypes.BcsSecret, error)
	// DeleteSecret xxx
	// delete secret
	DeleteSecret(ns, name string) error
	// ListAllSecrets xxx
	// list ns secret
	// ListSecrets(runAs string) ([]*commtypes.BcsSecret, error)
	// list all secret
	ListAllSecrets() ([]*commtypes.BcsSecret, error)

	// SaveService xxx
	// save service
	SaveService(service *commtypes.BcsService) error
	// FetchService xxx
	// fetch service
	FetchService(ns, name string) (*commtypes.BcsService, error)
	// DeleteService xxx
	// delete service
	DeleteService(ns, name string) error
	// ListAllServices xxx
	// list service by namespace
	// ListServices(runAs string) ([]*commtypes.BcsService, error)
	// list all services
	ListAllServices() ([]*commtypes.BcsService, error)

	// SaveEndpoint xxx
	// save endpoint
	SaveEndpoint(endpoint *commtypes.BcsEndpoint) error
	// FetchEndpoint xxx
	// fetch endpoint
	FetchEndpoint(ns, name string) (*commtypes.BcsEndpoint, error)
	// DeleteEndpoint xxx
	// delete endpoint
	DeleteEndpoint(ns, name string) error

	// SaveDeployment xxx
	// save deployment
	SaveDeployment(deployment *types.Deployment) error
	// FetchDeployment xxx
	// fetch deployment
	FetchDeployment(ns, name string) (*types.Deployment, error)
	// ListDeployments xxx
	// list deployments
	ListDeployments(ns string) ([]*types.Deployment, error)
	// DeleteDeployment xxx
	// delete deployment
	DeleteDeployment(ns, name string) error
	// ListAllDeployments xxx
	// list all deployment
	ListAllDeployments() ([]*types.Deployment, error)

	// InitDeploymentLockPool xxx
	// init deployments lock pool
	InitDeploymentLockPool()
	// ListDeploymentRunAs xxx
	// list namespaces for deployments
	ListDeploymentRunAs() ([]string, error)
	// ListDeploymentNodes xxx
	// list deployment nodes
	ListDeploymentNodes(runAs string) ([]string, error)
	// LockDeployment xxx
	// lock a deployment
	LockDeployment(deploymentName string)
	// UnLockDeployment xxx
	// unlock a deployment
	UnLockDeployment(deploymentName string)

	// InitCacheMgr xxx
	// init cache manager
	InitCacheMgr(bool) error
	// UnInitCacheMgr xxx
	// uninit cache manager
	UnInitCacheMgr() error

	// SaveCustomResourceRegister xxx
	// save custom resource register
	SaveCustomResourceRegister(*commtypes.Crr) error

	// DeleteCustomResourceRegister xxx
	// delete custom resource register
	// para1: crr.spec.names.kind
	DeleteCustomResourceRegister(string) error

	// ListCustomResourceRegister xxx
	// fetch custom resource register list
	ListCustomResourceRegister() ([]*commtypes.Crr, error)

	// ListAllCrds xxx
	// list all crds
	ListAllCrds(kind string) ([]*commtypes.Crd, error)

	// SaveCustomResourceDefinition xxx
	// save custom resource definition
	SaveCustomResourceDefinition(*commtypes.Crd) error

	// DeleteCustomResourceDefinition xxx
	// delete custom resource definition
	// para1: crd.kind
	// para2: namespace
	// para3: name
	DeleteCustomResourceDefinition(string, string, string) error

	// InitCmdLockPool xxx
	// init command lock
	InitCmdLockPool()
	// LockCommand xxx
	// lock command by command_id
	LockCommand(cmdId string)
	// UnLockCommand xxx
	// unlock command by command_id
	UnLockCommand(cmdId string)
	// SaveCommand xxx
	// save command
	SaveCommand(command *commtypes.BcsCommandInfo) error
	// FetchCommand xxx
	// fetch command
	FetchCommand(ID string) (*commtypes.BcsCommandInfo, error)
	// DeleteCommand xxx
	// delete command
	DeleteCommand(ID string) error
	// ListCustomResourceDefinition xxx
	// fetch custom resource definition list
	// para1: crd.kind
	// para2: namespace
	ListCustomResourceDefinition(kind, ns string) ([]*commtypes.Crd, error)

	// FetchCustomResourceDefinition xxx
	// fetch custom resource definition
	// para1: crd.kind
	// para2: namespace
	// para3: name
	FetchCustomResourceDefinition(kind, ns, name string) (*commtypes.Crd, error)

	// SaveAdmissionWebhook xxx
	/*=========AdmissionWebhook==========*/
	SaveAdmissionWebhook(admission *commtypes.AdmissionWebhookConfiguration) error
	FetchAdmissionWebhook(ns, name string) (*commtypes.AdmissionWebhookConfiguration, error)
	DeleteAdmissionWebhook(ns, name string) error
	FetchAllAdmissionWebhooks() ([]*commtypes.AdmissionWebhookConfiguration, error)
	/*=========AdmissionWebhook==========*/

	// list object namespaces, object = applicationNode、versionNode...
	// ListObjectNamespaces(objectNode string) ([]string, error)

	// StartStoreObjectMetrics xxx
	// start metrics
	StartStoreObjectMetrics()
	// StopStoreMetrics xxx
	// stop metrics
	StopStoreMetrics()
	// FetchDaemonset xxx
	// fetch daemonset
	FetchDaemonset(namespace, name string) (*types.BcsDaemonset, error)
	// SaveDaemonset xxx
	// save daemonset
	SaveDaemonset(daemon *types.BcsDaemonset) error
	// ListAllDaemonset xxx
	// List all daemonsets
	ListAllDaemonset() ([]*types.BcsDaemonset, error)
	// DeleteDaemonset xxx
	// delete daemonset
	DeleteDaemonset(namespace, name string) error
	// ListDaemonsetTaskGroups xxx
	// list daemonset't taskgroup
	ListDaemonsetTaskGroups(namespace, name string) ([]*types.TaskGroup, error)

	// FetchTransaction fetch transaction
	FetchTransaction(namespace, name string) (*types.Transaction, error)
	// SaveTransaction save transaction
	SaveTransaction(transaction *types.Transaction) error
	// ListTransaction list transaction by namespace
	ListTransaction(ns string) ([]*types.Transaction, error)
	// ListAllTransaction list all transaction
	ListAllTransaction() ([]*types.Transaction, error)
	// DeleteTransaction delete transaction
	DeleteTransaction(namespace, name string) error
}

// Dbdrvier xxx
// The interface for db operations
type Dbdrvier interface {
	// Connect xxx
	Connect() error
	// Insert xxx
	// save data to db
	Insert(string, string) error
	// Fetch data from db
	Fetch(string) ([]byte, error)
	// Update data to db
	Update(string, string) error
	// Delete data from db
	Delete(string) error
	// List the key of the data from db
	List(string) ([]string, error)
}
