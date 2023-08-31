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

package helm

import (
	"crypto/tls"
)

const (
	// ModuleHelmManager default discovery helmmanager module
	ModuleHelmManager = "helmmanager.bkbcs.tencent.com"

	// PubicRepo public repo
	PubicRepo = "public-repo"
)

var (
	// operator default operator
	operator = "bcs-cluster-manager"

	// FailedState failed
	FailedState = "failed"

	// Install status
	// PendingInstall xxx
	PendingInstall = "pending-install"
	// FailedInstall xxx
	FailedInstall = "failed-install"
	// DeployedInstall xxx
	DeployedInstall = "deployed"

	// upgrade status
	// PendingUpgrade xxx
	PendingUpgrade = "pending-upgrade"
	// FailedUpgrade xxx
	FailedUpgrade = "failed-upgrade"
	// DeployedUpgrade xxx
	DeployedUpgrade = "deployed"

	// rollback status
	// PendingRollback xxx
	PendingRollback = "pending-rollback"
	// FailedRollback xxx
	FailedRollback = "failed-rollback"
	// DeployedRollback xxx
	DeployedRollback = "deployed"

	// UnInstall status
	// UnInstalling xxx
	UnInstalling = "uninstalling"
	// FailedUninstall xxx
	FailedUninstall = "failed-uninstall"
)

// Config for bcsapi
type Config struct {
	// bcsapi host, available like 127.0.0.1:8080
	Hosts []string
	// tls configuratio
	TLSConfig *tls.Config
	// AuthToken for permission verification
	AuthToken string
	// clusterID for Kubernetes/Mesos operation
	ClusterID string
	// Header for request header
	Header map[string]string
	// InnerClientName for bcs inner auth, like bcs-cluster-manager
	InnerClientName string
}
