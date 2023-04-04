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

package retry

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	retryLong = templates.LongDesc(i18n.T(`
	retry create a resource from stdin.`))

	retryExample = templates.Examples(i18n.T(`
	# retry a cluster variable
	kubectl-bcs-cluster-manager retry`))

	taskID    string
	clusterID string
)

// NewRetryCmd 创建retry子命令实例
func NewRetryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retry",
		Short:   i18n.T("retry create a resource from stdin."),
		Long:    retryLong,
		Example: retryExample,
	}

	// retry subcommands
	cmd.AddCommand(newRetryTaskCmd())
	cmd.AddCommand(newRetryCreateClusterTaskCmd())

	return cmd
}
