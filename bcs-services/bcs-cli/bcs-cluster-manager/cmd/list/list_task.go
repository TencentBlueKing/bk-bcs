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

package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/printer"
	taskMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	listTaskExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager list task --clusterID xxx --projectID xxx`))
)

func newListTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "task",
		Short:   "list task from bcs-cluster-manager",
		Example: listTaskExample,
		Run:     listTask,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", `cluster ID`)
	cmd.Flags().StringVarP(&projectID, "projectID", "p", "", `project ID`)

	return cmd
}

func listTask(cmd *cobra.Command, args []string) {
	resp, err := taskMgr.New(context.Background()).List(types.ListTaskReq{
		ClusterID: clusterID,
		ProjectID: projectID,
	})
	if err != nil {
		klog.Fatalf("list task failed: %v", err)
	}

	header := []string{"TASK_ID", "TASK_NAME", "TASK_TYPE", "STATUS", "MESSAGE", "START", "END",
		"EXECUTION_TIME", "CURRENT_STEP", "CLUSTER_ID", "PROJECT_ID", "NODE_GROUP_ID", "CREATOR", "UPDATER"}
	data := make([][]string, len(resp.Data))
	for key, item := range resp.Data {
		data[key] = []string{
			item.TaskID,
			item.TaskName,
			item.TaskType,
			item.Status,
			item.Message,
			item.Start,
			item.End,
			fmt.Sprintf("%d", item.ExecutionTime),
			item.CurrentStep,
			item.ClusterID,
			item.ProjectID,
			item.NodeGroupID,
			item.Creator,
			item.Updater,
		}
	}

	err = printer.PrintList(flagOutput, resp.Data, header, data)
	if err != nil {
		klog.Fatalf("list cloud account to perm failed: %v", err)
	}
}
