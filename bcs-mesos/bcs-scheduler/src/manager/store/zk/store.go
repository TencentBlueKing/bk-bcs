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
	"sync"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
)

// Store Manager
type managerStore struct {
	Db store.Dbdrvier

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// Create a store manager by a db driver
func NewManagerStore(dbDriver store.Dbdrvier) store.Store {
	s := &managerStore{
		Db: dbDriver,
	}

	return s
}

func (s *managerStore) StopStoreMetrics() {
	if s.cancel == nil {
		return
	}
	s.cancel()

	time.Sleep(time.Second)
	s.wg.Wait()
}

func (s *managerStore) StartStoreObjectMetrics() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	for {
		time.Sleep(time.Minute)

		select {
		case <-s.ctx.Done():
			blog.Infof("stop scheduler store metrics")
			return

		default:
			s.wg.Add(1)
			store.ObjectResourceInfo.Reset()
		}

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

			store.ReportAgentInfoMetrics(info.IP, info.CpuTotal, info.CpuTotal-info.CpuUsed,
				info.MemTotal, info.MemTotal-info.MemUsed)
		}

		s.wg.Done()
	}
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
)
