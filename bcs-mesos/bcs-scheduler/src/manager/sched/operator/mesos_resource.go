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

package operator

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	master "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos/master"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/client"

	"github.com/golang/protobuf/proto"
)

//GetMesosResource Get cluster current resource information from mesos master
func GetMesosResource(mesosClient *client.Client) (*commtypes.BcsClusterResource, error) {
	clusterRes := new(commtypes.BcsClusterResource)
	blog.Info("get cluster resource from mesos master")

	if mesosClient == nil {
		blog.Error("get cluster resource error: mesos Client is nil")
		return nil, fmt.Errorf("system error: mesos client is nil")
	}

	call := &master.Call{
		Type: master.Call_GET_AGENTS.Enum(),
	}
	req, err := proto.Marshal(call)
	if err != nil {
		blog.Error("get cluster resource: query agentInfo proto.Marshal err: %s", err.Error())
		return nil, fmt.Errorf("system error: proto marshal error")
	}
	resp, err := mesosClient.Send(req)
	if err != nil {
		blog.Error("get cluster resource: query agentInfo Send err: %s", err.Error())
		return nil, fmt.Errorf("send request to mesos error: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		blog.Error("get cluster resource: query agentInfo unexpected response statusCode: %d", resp.StatusCode)
		return nil, fmt.Errorf("mesos response statuscode: %d", resp.StatusCode)
	}

	var response master.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		blog.Error("get cluster resource: Decode response failed: %s", err.Error())
		return nil, fmt.Errorf("mesos response decode err: %s", err.Error())
	}
	blog.V(3).Infof("get cluster resource: response msg type(%d)", response.GetType())
	agentInfo := response.GetGetAgents()
	if agentInfo == nil {
		blog.Warn("get cluster resource: response Agents == nil")
	}

	cpuTotal := 0.0
	cpuUsed := 0.0
	memTotal := 0.0
	memUsed := 0.0
	diskTotal := 0.0
	diskUsed := 0.0
	agent := new(commtypes.BcsClusterAgentInfo)
	for _, oneAgent := range agentInfo.Agents {
		//blog.V(3).Infof("get agents: ===>agent[%d]: %+v", index, oneAgent)
		agent.HostName = oneAgent.GetAgentInfo().GetHostname()
		agent.IP = oneAgent.GetPid()
		totalRes := oneAgent.GetTotalResources()
		for _, resource := range totalRes {
			if resource.GetName() == "cpus" {
				agent.CpuTotal = resource.GetScalar().GetValue()
				cpuTotal += agent.CpuTotal
			}
			if resource.GetName() == "mem" {
				agent.MemTotal = resource.GetScalar().GetValue()
				memTotal += agent.MemTotal
			}
			if resource.GetName() == "disk" {
				agent.DiskTotal = resource.GetScalar().GetValue()
				diskTotal += agent.DiskTotal
			}
		}
		usedRes := oneAgent.GetAllocatedResources()
		for _, resource := range usedRes {
			if resource.GetName() == "cpus" {
				agent.CpuUsed = resource.GetScalar().GetValue()
				cpuUsed += agent.CpuUsed
			}
			if resource.GetName() == "mem" {
				agent.MemUsed = resource.GetScalar().GetValue()
				memUsed += agent.MemUsed
			}
			if resource.GetName() == "disk" {
				agent.DiskUsed = resource.GetScalar().GetValue()
				diskUsed += agent.DiskUsed
			}
		}
		clusterRes.Agents = append(clusterRes.Agents, *agent)
	}

	clusterRes.CpuTotal = cpuTotal
	clusterRes.MemTotal = memTotal
	clusterRes.DiskTotal = diskTotal
	clusterRes.CpuUsed = cpuUsed
	clusterRes.MemUsed = memUsed
	clusterRes.DiskUsed = diskUsed

	blog.Info("get cluster resource: cpu %f/%f  || mem %f/%f || disk %f/%f",
		cpuUsed, cpuTotal, memUsed, memTotal, diskUsed, diskTotal)

	blog.V(3).Infof("get cluster resource: %+v", clusterRes)

	return clusterRes, nil
}
