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

package blueking

import (
	"fmt"
)

var (
	cloudName = "blueking"
)

const (
	// deleteClusterNodesTaskTemplate bk-sops delete clusterNodes task template
	deleteClusterNodesTaskTemplate = "blueking-remove nodes: %s"
	// addClusterNodesTaskTemplate bk-sops add clusterNodes task template
	addClusterNodesTaskTemplate = "blueking-add nodes: %s"
	// deleteClusterTaskTemplate bk-sops delete cluster task template
	deleteClusterTaskTemplate = "blueking-delete cluster: %s"
	// createClusterTaskTemplate bk-sops delete cluster task template
	createClusterTaskTemplate = "blueking-create cluster: %s"
)

var (
	updateCreateClusterDBInfoTask = fmt.Sprintf("%s-UpdateCreateClusterDBInfoTask", cloudName)
	cleanClusterDBInfoTask        = fmt.Sprintf("%s-CleanClusterDBInfoTask", cloudName)
	updateAddNodeDBInfoTask       = fmt.Sprintf("%s-UpdateAddNodeDBInfoTask", cloudName)
	updateRemoveNodeDBInfoTask    = fmt.Sprintf("%s-UpdateRemoveNodeDBInfoTask", cloudName)
)
