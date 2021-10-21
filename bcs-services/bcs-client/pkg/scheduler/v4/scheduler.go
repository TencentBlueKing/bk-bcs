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

package v4

import (
	"context"
	"net/url"

	commonTypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	schetypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/utils"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

//Scheduler mesos scheduler interface for bcs-api
type Scheduler interface {
	CreateApplication(clusterID, namespace string, data []byte) error
	CreateProcess(clusterID, namespace string, data []byte) error
	CreateConfigMap(clusterID, namespace string, data []byte) error
	CreateSecret(clusterID, namespace string, data []byte) error
	CreateService(clusterID, namespace string, data []byte) error
	CreateDeployment(clusterID, namespace string, data []byte) error
	CreateDaemonset(clusterID, namespace string, data []byte) error

	UpdateApplication(clusterID, namespace string, data []byte, extraValue url.Values) error
	UpdateProcess(clusterID, namespace string, data []byte, extraValue url.Values) error
	UpdateConfigMap(clusterID, namespace string, data []byte, extraValue url.Values) error
	UpdateSecret(clusterID, namespace string, data []byte, extraValue url.Values) error
	UpdateService(clusterID, namespace string, data []byte, extraValue url.Values) error
	UpdateDeployment(clusterID, namespace string, data []byte, extraValue url.Values) error

	DeleteApplication(clusterID, namespace, name string, enforce bool) error
	DeleteProcess(clusterID, namespace, name string, enforce bool) error
	DeleteConfigMap(clusterID, namespace, name string, enforce bool) error
	DeleteSecret(clusterID, namespace, name string, enforce bool) error
	DeleteService(clusterID, namespace, name string, enforce bool) error
	DeleteDeployment(clusterID, namespace, name string, enforce bool) error
	DeleteDaemonset(clusterID, namespace, name string, enforce bool) error

	ScaleApplication(clusterID, namespace, name string, instance int) error
	ScaleProcess(clusterID, namespace, name string, instance int) error

	RollBackApplication(clusterID, namespace string, data []byte) error
	RollBackProcess(clusterID, namespace string, data []byte) error

	RescheduleTaskGroup(clusterID, namespace, applicationName, taskGroupName string) error

	ResumeDeployment(clusterID, namespace, name string) error
	CancelDeployment(clusterID, namespace, name string) error
	PauseDeployment(clusterID, namespace, name string) error

	ListAgentInfo(clusterID string, ipList []string) ([]*commonTypes.BcsClusterAgentInfo, error)
	ListAgentSetting(clusterID string, ipList []string) ([]*commonTypes.BcsClusterAgentSetting, error)
	UpdateStringAgentSetting(clusterID string, ipList []string, key, value string) error
	UpdateScalarAgentSetting(clusterID string, ipList []string, key string, value float64) error
	UpdateAgentSetting(clusterID string, data []byte) error
	SetAgentSetting(clusterID string, data []byte) error
	DeleteAgentSetting(clusterID string, ipList []string) error
	EnableAgent(clusterID string, ipList []string) error
	DisableAgent(clusterID string, ipList []string) error

	GetApplicationDefinition(clusterID, namespace, name string) (*commonTypes.ReplicaController, error)
	GetProcessDefinition(clusterID, namespace, name string) (*commonTypes.ReplicaController, error)
	GetDeploymentDefinition(clusterID, namespace, name string) (*commonTypes.BcsDeployment, error)
	//GetOffer get specified mesos cluster resources list by agents
	GetOffer(clusterID string) ([]*schetypes.OfferWithDelta, error)

	/*
		CustomResourceDefinition section
	*/
	//CreateResourceDefinition create CRD by definition file
	CreateCustomResourceDefinition(clusterID string, data []byte) error
	//UpdateResourceDefinition replace specified CRD
	UpdateCustomResourceDefinition(clusterID, name string, data []byte) error
	//ListCustomResourceDefinition list all created CRD
	ListCustomResourceDefinition(clusterID string) (*v1beta1.CustomResourceDefinitionList, error)
	//GetCustomResourceDefinition get specified CRD
	GetCustomResourceDefinition(clusterID string, name string) (*v1beta1.CustomResourceDefinition, error)
	//DeleteCustomResourceDefinition delete specified CRD
	DeleteCustomResourceDefinition(clusterID, name string) error
	/*
		CustomResource section, depend on ListCustomResourceDefinition for validation
	*/
	//CreateResource create CRD by definition file
	CreateCustomResource(clusterID, apiVersion, plural, namespace string, data []byte) error
	//UpdateResource replace specified CRD
	UpdateCustomResource(clusterID, apiVersion, plural, namespace, name string, data []byte) error
	//ListCustomResource list all created CRD
	ListCustomResource(clusterID, apiVersion, plural, namespace string) ([]byte, error)
	//GetCustomResource get specified CRD
	GetCustomResource(clusterID, apiVersion, plural, namespace, name string) ([]byte, error)
	//DeleteCustomResource delete specified CRD
	DeleteCustomResource(clusterID, apiVersion, plural, namespace, name string) error

	CreateContainerExec(clusterId, containerId, hostIp string, command []string) (string, error)
	StartContainerExec(ctx context.Context, clusterId, execId, containerId, hostIp string) (types.HijackedResponse, error)
	ResizeContainerExec(clusterId, execId, hostIp string, height, width int) error

	ListTransaction(clusterID, objKind, objNs, objName string) ([]*schetypes.Transaction, error)
	DeleteTransaction(clusterID, ns, name string) error
}

const (
	bcsSchedulerResourceURI                 = "%s/bcsapi/v4/scheduler/mesos/namespaces/%s/%s?%s"
	bcsSchedulerDeleteResourceURI           = "%s/bcsapi/v4/scheduler/mesos/namespaces/%s/%s/%s?enforce=%d"
	bcsSchedulerScaleResourceURI            = "%s/bcsapi/v4/scheduler/mesos/namespaces/%s/%s/%s/scale/%d"
	bcsSchedulerRollBackResourceURI         = "%s/bcsapi/v4/scheduler/mesos/namespaces/%s/%s/rollback"
	bcsSchedulerResumeDeploymentURI         = "%s/bcsapi/v4/scheduler/mesos/namespaces/%s/deployments/%s/resumeupdate"
	bcsSchedulerCancelDeploymentURI         = "%s/bcsapi/v4/scheduler/mesos/namespaces/%s/deployments/%s/cancelupdate"
	bcsSchedulerPauseDeploymentURI          = "%s/bcsapi/v4/scheduler/mesos/namespaces/%s/deployments/%s/pauseupdate"
	bcsSchedulerClusterResourceURI          = "%s/bcsapi/v4/scheduler/mesos/cluster/resources"
	bcsSchedulerAgentSettingURI             = "%s/bcsapi/v4/scheduler/mesos/agentsettings/?ips=%s"
	bcsSchedulerUpdateAgentSettingURI       = "%s/bcsapi/v4/scheduler/mesos/agentsettings/update"
	bcsSchedulerSetAgentSettingURI          = "%s/bcsapi/v4/scheduler/mesos/agentsettings"
	bcsSchedulerEnableAgentURI              = "%s/bcsapi/v4/scheduler/mesos/agentsettings/enable?ips=%s"
	bcsSchedulerDisableAgentURI             = "%s/bcsapi/v4/scheduler/mesos/agentsettings/disable?ips=%s"
	bcsSchedulerRescheduleURI               = "%s/bcsapi/v4/scheduler/mesos/namespaces/%s/applications/%s/taskgroups/%s/rescheduler"
	bcsSchedulerOfferURI                    = "%s/bcsapi/v4/scheduler/mesos/cluster/current/offers"
	bcsSchedulerAppDefinitionURI            = "%s/bcsapi/v4/scheduler/mesos/definition/application/%s/%s"
	bcsSchedulerDeployDefinitionURI         = "%s/bcsapi/v4/scheduler/mesos/definition/deployment/%s/%s"
	bcsSchedulerCustomResourceURL           = "%s/bcsapi/v4/scheduler/mesos/customresources"
	bcsScheudlerCustomResourceDefinitionURL = "%s/bcsapi/v4/scheduler/mesos/customresourcedefinitions"
	bcsSchedulerCreateExecUri               = "%s/bcsapi/v4/scheduler/mesos/webconsole/create_exec?host_ip=%s"
	bcsSchedulerStartExecUri                = "%s/bcsapi/v4/scheduler/mesos/webconsole/start_exec?host_ip=%s&container_id=%s&exec_id=%s"
	bcsSchedulerResizeExecUri               = "%s/bcsapi/v4/scheduler/mesos/webconsole/resize_exec?host_ip=%s"
	bcsSchedulerTransactionListUri          = "%s/bcsapi/v4/scheduler/mesos/transactions/%s?objKind=%s&objName=%s"
	bcsSchedulerTransactionDeleteUri        = "%s/bcsapi/v4/scheduler/mesos/transactions/%s/%s"
)

type bcsScheduler struct {
	bcsAPIAddress string
	requester     utils.ApiRequester
}

//NewBcsScheduler create mesos scheduler api implemenation
func NewBcsScheduler(options types.ClientOptions) Scheduler {
	return &bcsScheduler{
		bcsAPIAddress: options.BcsApiAddress,
		requester:     utils.NewApiRequester(options.ClientSSL, options.BcsToken),
	}
}
