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

package update

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	taskMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	updateTaskExample = templates.Examples(i18n.T(`update task from json file. file template: 
	{"cloudID":"","networkType":"","region":"","vpcName":"test","vpcID":""}`))
)

func newUpdateTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "task",
		Short:   "update task from bcs-cluster-manager",
		Example: updateTaskExample,
		Run:     updateTask,
	}

	return cmd
}

func updateTask(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := types.UpdateTaskReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = taskMgr.New(context.Background()).Update(req)
	if err != nil {
		klog.Fatalf("update task failed: %v", err)
	}

	fmt.Println("update task succeed")
}
