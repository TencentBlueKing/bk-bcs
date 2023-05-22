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

package create

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	createLong = templates.LongDesc(i18n.T(`
	create a resource from a file or from stdin.`))

	createExample = templates.Examples(i18n.T(`
	# create a cluster variable
	kubectl-bcs-cluster-manager create`))

	virtual  bool
	filename string
)

// NewCreateCmd 创建create子命令实例
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create -f FILENAME",
		Short:   i18n.T("Create a resource from a file or from stdin"),
		Long:    createLong,
		Example: createExample,
	}

	cmd.PersistentFlags().StringVarP(&filename, "filename", "f", "", "File address Support json file")

	// create subcommands
	cmd.AddCommand(newCreateCloudAccountCmd())
	cmd.AddCommand(newCreateCloudVPCCmd())
	cmd.AddCommand(newCreateClusterCmd())
	cmd.AddCommand(newCreateNodeGroupCmd())
	cmd.AddCommand(newCreateTaskCmd())

	return cmd
}
