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

import ()

// Store Manager
type managerStore struct {
	Db Dbdrvier
}

// Create a store manager by a db driver
func NewManagerStore(dbDriver Dbdrvier) Store {
	return &managerStore{
		Db: dbDriver,
	}
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
