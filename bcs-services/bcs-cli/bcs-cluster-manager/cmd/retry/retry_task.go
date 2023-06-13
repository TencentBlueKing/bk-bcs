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
	"context"
	"fmt"

	taskMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	retryTaskExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager retry task --taskID xxx`))
)

func newRetryTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "task",
		Short:   "retry task from bcs-cluster-manager",
		Example: retryTaskExample,
		Run:     retryTask,
	}

	cmd.Flags().StringVarP(&taskID, "taskID", "t", "", `task ID`)
	cmd.MarkFlagRequired("taskID")

	return cmd
}

func retryTask(cmd *cobra.Command, args []string) {
	resp, err := taskMgr.New(context.Background()).Retry(types.RetryTaskReq{
		TaskID: taskID,
	})
	if err != nil {
		klog.Fatalf("retry task failed: %v", err)
	}

	fmt.Printf("retry task succeed: taskID: %v\n", resp.TaskID)
}
