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

package tools

import (
	"encoding/json"

	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-mesos/bcs-container-executor/container"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
)

//if status!=running, return ""
func GetTaskgroupIp(t *types.TaskGroup) string {
	if t.Status != types.TASKGROUP_STATUS_RUNNING {
		return ""
	}
	bcsInfo := new(container.BcsContainerInfo)
	for _, oneTask := range t.Taskgroup {
		// process task do not have the statusData upload by executor, because process executor
		// do not have the hostIP and port information. So we make NodeIP, ContainerIP, HostIP directly with AgentIPAddress
		// which is got from offer
		// current running taskgroup kind maybe empty, regard them as APP.
		switch oneTask.Kind {
		case commtypes.BcsDataType_PROCESS:
			return oneTask.AgentIPAddress
		case commtypes.BcsDataType_APP, "":
			if len(oneTask.StatusData) == 0 {
				continue
			}
			if err := json.Unmarshal([]byte(oneTask.StatusData), &bcsInfo); err != nil {
				blog.Warn("task %s StatusData unmarshal err: %s, cannot add to backend", oneTask.ID, err.Error())
				continue
			}

			if bcsInfo.IPAddress != "" {
				return bcsInfo.IPAddress
			}
		}
	}

	return ""
}
