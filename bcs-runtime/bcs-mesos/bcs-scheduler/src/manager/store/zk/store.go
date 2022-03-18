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

package zk

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	typesplugin "github.com/Tencent/bk-bcs/bcs-common/common/plugin"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-scheduler/src/pluginManager"
)

// Store Manager
type managerStore struct {
	Db store.Dbdrvier

	ctx    context.Context
	cancel context.CancelFunc

	// plugin manager, ip-resources
	pm        *pluginManager.PluginManager
	clusterId string
}

// Create a store manager by a db driver
func NewManagerStore(dbDriver store.Dbdrvier, pm *pluginManager.PluginManager, clusterId string) store.Store {
	s := &managerStore{
		Db:        dbDriver,
		pm:        pm,
		clusterId: clusterId,
	}

	return s
}

func (s *managerStore) StopStoreMetrics() {
	if s.cancel == nil {
		return
	}
	s.cancel()
	time.Sleep(time.Second)
}

func (s *managerStore) StartStoreObjectMetrics() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	for {
		time.Sleep(time.Minute)
		if cacheMgr == nil {
			continue
		}
		blog.Infof("start produce metrics")
		store.ObjectResourceInfo.Reset()
		store.TaskgroupInfo.Reset()
		store.AgentCpuResourceRemain.Reset()
		store.AgentCpuResourceTotal.Reset()
		store.AgentMemoryResourceRemain.Reset()
		store.AgentMemoryResourceTotal.Reset()
		store.AgentIpResourceRemain.Reset()
		store.StorageOperatorFailedTotal.Reset()
		store.StorageOperatorLatencyMs.Reset()
		store.StorageOperatorTotal.Reset()
		store.ClusterMemoryResourceRemain.Reset()
		store.ClusterCpuResourceRemain.Reset()
		store.ClusterMemoryResourceTotal.Reset()
		store.ClusterCpuResourceTotal.Reset()
		store.ClusterCpuResourceAvailable.Reset()
		store.ClusterMemoryResourceAvailable.Reset()

		// handle service metrics
		services, err := s.ListAllServices()
		if err != nil {
			blog.Errorf("list all services error %s", err.Error())
		}
		for _, service := range services {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceService, service.NameSpace, service.Name, "")
		}

		// handle application metrics
		apps, err := s.ListAllApplications()
		if err != nil {
			blog.Errorf("list all applications error %s", err.Error())
		}
		for _, app := range apps {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceApplication, app.RunAs, app.Name, app.Status)

			// handle taskgroup metrics
			taskgroups, err := s.ListTaskGroups(app.RunAs, app.Name)
			if err != nil {
				blog.Errorf("list all services error %s", err.Error())
			}
			for _, taskgroup := range taskgroups {
				store.ReportTaskgroupInfoMetrics(taskgroup.RunAs, taskgroup.AppID, taskgroup.ID, taskgroup.Status)
			}
		}

		// handle deployment metrics
		deployments, err := s.ListAllDeployments()
		if err != nil {
			blog.Errorf("list all deployment error %s", err.Error())
		}
		for _, deployment := range deployments {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceDeployment, deployment.ObjectMeta.NameSpace, deployment.ObjectMeta.Name, "")
		}

		// handle configmap metrics
		configmaps, err := s.ListAllConfigmaps()
		if err != nil {
			blog.Errorf("list all configmap error %s", err.Error())
		}
		for _, configmap := range configmaps {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceConfigmap, configmap.NameSpace, configmap.Name, "")
		}

		// handle secrets metrics
		secrets, err := s.ListAllSecrets()
		if err != nil {
			blog.Errorf("list all secret error %s", err.Error())
		}
		for _, secret := range secrets {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceSecret, secret.NameSpace, secret.Name, "")
		}

		var (
			clusterCpu   float64
			clusterMem   float64
			remainCpu    float64
			remainMem    float64
			availableCpu float64
			availableMem float64
		)

		// handle agentSettings
		agentSettingsMap := make(map[string]bool)
		agentSettings, err := s.ListAgentsettings()
		if err != nil {
			blog.Error("list all agent settings error %s", err.Error())
		}
		for _, setting := range agentSettings {
			agentSettingsMap[setting.InnerIP] = setting.Disabled
		}

		// handle agents metrics
		agents, err := s.ListAllAgents()
		if err != nil {
			blog.Errorf("list all agent error %s", err.Error())
		}
		for _, agent := range agents {
			info := agent.GetAgentInfo()
			if info.IP == "" {
				blog.Errorf("agent %s don't have InnerIP attribute", agent.Key)
				continue
			}

			schedInfo, err := s.FetchAgentSchedInfo(info.HostName)
			if err != nil && !errors.Is(err, store.ErrNoFound) {
				blog.Infof("failed to to fetch agent sched info of host %s, err %s", info.HostName, err.Error())
				continue
			}

			var ipValue float64
			if s.pm != nil {
				// request netservice to node container ip
				para := &typesplugin.HostPluginParameter{
					Ips:       []string{info.IP},
					ClusterId: s.clusterId,
				}

				outerAttri, err := s.pm.GetHostAttributes(para)
				if err != nil {
					blog.Errorf("Get host(%s) ip-resources failed: %s", info.IP, err.Error())
					continue
				}
				attr, ok := outerAttri[info.IP]
				if !ok {
					blog.Errorf("host(%s) don't have ip-resources attributes", info.IP)
					continue
				}
				ipAttr := attr.Attributes[0]
				blog.Infof("Host(%s) %s Scalar(%f)", info.IP, ipAttr.Name, ipAttr.Scalar.Value)
				ipValue = ipAttr.Scalar.Value
			}

			// if ip-resources is zero, then ignore it
			if s.pm == nil || ipValue > 0 {
				agentDisabled, ok := agentSettingsMap[info.IP]
				if schedInfo != nil {
					remainCpu += float2Float(info.CpuTotal - info.CpuUsed - schedInfo.DeltaCPU)
					remainMem += float2Float(info.MemTotal - info.MemUsed - schedInfo.DeltaMem)
					// no need to add remain cpu if agent is disabled
					if ok && !agentDisabled {
						availableCpu += float2Float(info.CpuTotal - info.CpuUsed - schedInfo.DeltaCPU)
						availableMem += float2Float(info.MemTotal - info.MemUsed - schedInfo.DeltaMem)
					}
				} else {
					remainCpu += float2Float(info.CpuTotal - info.CpuUsed)
					remainMem += float2Float(info.MemTotal - info.MemUsed)
					if ok && !agentDisabled {
						availableCpu += float2Float(info.CpuTotal - info.CpuUsed)
						availableMem += float2Float(info.MemTotal - info.MemUsed)
					}
				}
			}
			clusterCpu += float2Float(info.CpuTotal)
			clusterMem += float2Float(info.MemTotal)

			store.ReportAgentInfoMetrics(info.IP, s.clusterId, info.CpuTotal, info.CpuTotal-info.CpuUsed,
				info.MemTotal, info.MemTotal-info.MemUsed, ipValue)
		}
		store.ReportClusterInfoMetrics(s.clusterId, remainCpu, availableCpu, clusterCpu, remainMem,
			availableMem, clusterMem)
	}
}

func float2Float(num float64) float64 {
	float_num, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", num), 64)
	return float_num
}

func (store *managerStore) ListObjectNamespaces(objectNode string) ([]string, error) {

	rootPath := "/" + bcsRootNode + "/" + objectNode

	runAses, err := store.Db.List(rootPath)
	if err != nil {
		return nil, err
	}

	if nil == runAses {
		blog.Error("no runAs in (%s)", rootPath)
		return nil, nil
	}

	return runAses, nil
}

const (
	// applicationNode is the zk node name of the application
	applicationNode string = "application"
	// versionNode is the zk node name of the version
	versionNode string = "version"
	// frameWorkNode is the zk node name of the framwork
	frameWorkNode string = "framework"
	// bcsRootNode is the root node name
	bcsRootNode string = "blueking"
	// Default namespace
	defaultRunAs string = "defaultGroup"
	// agentNode is the zk node: sync from mesos master
	agentNode string = "agent"
	// ZK node for agent settings: configured by users
	agentSettingNode string = "agentsetting"
	// ZK node for agent information: scheduler created information
	agentSchedInfoNode string = "agentschedinfo"
	// configMapNode is the zk node name of configmaps
	configMapNode string = "configmap"
	// secretNode is the zk node name of secrets
	secretNode string = "secret"
	// serviceNode is the zk node name of services
	serviceNode string = "service"
	// Endpoint zk node
	endpointNode string = "endpoint"
	// Deployment zk node
	deploymentNode string = "deployment"
	// crr zk node
	crrNode string = "crr"
	//crd zk node
	crdNode string = "crd"
	//command zk node
	commandNode string = "command"
	//admission webhook zk node
	AdmissionWebhookNode string = "admissionwebhook"
	// Transaction zk node
	transactionNode string = "transaction"
)
