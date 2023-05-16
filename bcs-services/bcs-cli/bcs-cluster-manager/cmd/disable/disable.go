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

package disable

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	disableLong = templates.LongDesc(i18n.T(`
	disable auto-scale a deployment, replica set, stateful set, or replication controller.`))

	disableExample = templates.Examples(i18n.T(`
	# disable a cluster variable
	kubectl-bcs-cluster-manager disable`))

	nodeGroupID string
)

// NewDisableCmd 创建disable子命令实例
func NewDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "disable",
		Short:   i18n.T("disable auto-scale a deployment, replica set, stateful set, or replication controller"),
		Long:    disableLong,
		Example: disableExample,
	}

	// disable subcommands
	cmd.AddCommand(newDisableNodeGroupAutoScaleCmd())

	return cmd
}
