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

package create

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
	createTaskExample = templates.Examples(i18n.T(`create task from json file. file template: 
	{"taskID":"feec6ed2-c3e3-481f-a58b-xxxxxx","taskType":"blueking-xxxxxxxxxxxx","status":"FAILED",
	"message":"step bksopsjob-xxxxxxxxxx running failed","start":"2022-11-11T18:23:32+08:00",
	"end":"2022-11-11T18:24:03+08:00","executionTime":31,"currentStep":"bksopsjob-xxxxxxxxxx",
	"stepSequence":["bksopsjob-xxxxxxxxxx","blueking-xxxxxxxxxxxxxxxxxxxxxx"],"steps":{"bksopsjob-xxxxxxxxxx":
	{"name":"bksopsjob-xxxxxxxxxx","system":"bksops","link":"","params":{"taskUrl":"http://apps.xxx.com"},
	"retry":0,"start":"2022-11-11T18:23:32+08:00","end":"2022-11-11T18:24:03+08:00","executionTime":31,
	"status":"FAILURE","message":"running fialed","lastUpdate":"2022-11-11T18:24:03+08:00","taskMethod":"xxx",
	"taskName":"标准运维任务","skipOnFailed":false},"blueking-xxxxxxxxxxxxxxxxxxxxxx":{"name":
	"blueking-xxxxxxxxxxxxxxxxxxxxxx","system":"api","link":"","params":null,"retry":0,"start":"","end":"",
	"executionTime":0,"status":"NOTSTARTED","message":"","lastUpdate":"","taskMethod":
	"blueking-xxxxxxxxxxxxxxxxxxxxxx","taskName":"更新任务状态","skipOnFailed":false}},"clusterID":"BCS-K8S-xxx",
	"projectID":"b363e23b1b354928a0f3exxxxxx","creator":"bcs","lastUpdate":"2022-11-11T18:24:03+08:00",
	"updater":"bcs","forceTerminate":false,"commonParams":{"jobType":"add-node","nodeIPs":"xxx.xxx.xxx.xxx",
	"operator":"bcs","taskName":"blueking-add nodes: BCS-K8S-xxxxxxxxx","user":"bcs"},
	"taskName":"集群添加节点任务","nodeIPList":["xxx.xxx.xxx.xxx"],"nodeGroupID":"xxx"}`))
)

func newCreateTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "task",
		Short:   "create task from bcs-cluster-manager",
		Example: createTaskExample,
		Run:     createTask,
	}

	return cmd
}

func createTask(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := types.CreateTaskReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	resp, err := taskMgr.New(context.Background()).Create(req)
	if err != nil {
		klog.Fatalf("create task failed: %v", err)
	}

	fmt.Printf("create task succeed: taskID: %v\n", resp.TaskID)
}
