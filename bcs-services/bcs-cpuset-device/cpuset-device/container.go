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

package cpuset_device

import (
	"strings"

	"bk-bcs/bcs-common/common/blog"

	docker "github.com/fsouza/go-dockerclient"
)

func (c *CpusetDevicePlugin) updateCpusetNodes() error {
	//cpuset env, example: node:0;cpuset:0,1,2,3
	envs, err := c.listContainersCpusetEnvs()
	if err != nil {
		return err
	}

	for _, env := range envs {
		//node:0;cpuset:0,1,2,3
		values := strings.Split(env, ";")
		//node:0, nv[0]=node, nv[1]=0
		nv := strings.Split(values[0], ":")
		node, ok := c.nodes[nv[1]]
		if !ok {
			blog.Errorf("env %s invalid, node %s not found", env, nv[1])
			continue
		}
		//cpuset:0,1,2,3
		//cv[0]=cpuset, cv[1]=0,1,2,3
		cv := strings.Split(values[1], ":")
		sets := strings.Split(cv[1], ",")
		node.AllocatedCpuset = append(node.AllocatedCpuset, sets...)
	}
	blog.Infof("update cpuset nodes success")

	return nil
}

func (c *CpusetDevicePlugin) listContainersCpusetEnvs() ([]string, error) {
	containers, err := c.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		blog.Errorf("ListContainers failed: %s", err.Error())
		return nil, err
	}

	//traversal container for get allocated cpusets
	envs := make([]string, 0)
	for _, container := range containers {
		if container.Status != "running" {
			blog.Infof("container %s status %s, and continue", container.ID, container.Status)
			continue
		}
		info, err := c.client.InspectContainer(container.ID)
		if err != nil {
			blog.Errorf("inspect container %s failed %s, then continue", container.ID, err.Error())
			continue
		}

		//cpuset env, example: node:0;cpuset:0,1,2,3
		var envValue string
		//docker env format []string
		//example: []string{"k1=v1","k2=v2"...}
		for _, o := range info.Config.Env {
			//envs[0] is key, envs[1] is value
			envs := strings.Split(o, "=")
			if envs[0] == EnvBkbcsAllocateCpuset {
				envValue = envs[1]
				break
			}
		}
		//if container don't contain bkbcs_allocate_cpuset env, then continue
		if envValue == "" {
			blog.Infof("container %s don't contain bkbcs_allocate_cpuset env, then continue", container.ID)
			continue
		}

		blog.Infof("container %s contains env(%s=%s)", container.ID, EnvBkbcsAllocateCpuset, envValue)
		envs = append(envs, envValue)
	}

	return envs, nil
}
