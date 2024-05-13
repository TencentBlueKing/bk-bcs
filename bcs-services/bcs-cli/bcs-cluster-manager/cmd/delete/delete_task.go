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
 */

package delete

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	taskMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	deleteTaskExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager delete task --taskID xxx`))
)

func newDeleteTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "task",
		Short:   "delete task from bcs-cluster-manager",
		Example: deleteTaskExample,
		Run:     deleteTask,
	}

	cmd.Flags().StringVarP(&taskID, "taskID", "t", "", `task ID`)
	_ = cmd.MarkFlagRequired("taskID")

	return cmd
}

func deleteTask(cmd *cobra.Command, args []string) {
	err := taskMgr.New(context.Background()).Delete(types.DeleteTaskReq{
		TaskID: taskID,
	})
	if err != nil {
		klog.Fatalf("delete task failed: %v", err)
	}

	fmt.Println("delete task succeed")
}
