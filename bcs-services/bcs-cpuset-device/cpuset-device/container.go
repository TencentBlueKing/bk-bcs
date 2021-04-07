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
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	docker "github.com/fsouza/go-dockerclient"
)

// CgroupCpusetRoot cgroup fs root path for cpuset
const CgroupCpusetRoot = "/sys/fs/cgroup/cpuset/docker"

func (s *CpusetDevicePlugin) loopUpdateCpusetNodes() {
	for {
		time.Sleep(time.Minute * 10)
		s.updateCpusetNodes()
	}
}

func (s *CpusetDevicePlugin) updateCpusetNodes() error {
	s.lockNodes()
	defer s.unlockNodes()

	containers, err := s.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		blog.Errorf("ListContainers failed: %s", err.Error())
		return err
	}

	for _, node := range s.nodes {
		node.AllocatedCpuset = make([]string, 0)
	}
	// traversal container for get allocated cpusets
	for _, container := range containers {
		info, err := s.client.InspectContainer(container.ID)
		if err != nil {
			blog.Errorf("inspect container %s failed %s, then continue", container.ID, err.Error())
			continue
		}
		if info.State.Status != "running" {
			blog.Infof("container %s status %s, and continue", container.ID, info.State.Status)
			continue
		}

		// example:
		// node=0
		// cpusets=0,1,2,3
		node, cpusets := s.getContainerCpusetInfo(info)
		if node == "" || cpusets == "" {
			continue
		}

		mNode, ok := s.nodes[node]
		if !ok {
			blog.Errorf("container %s invalid, node %s not found", info.ID, node)
			continue
		}
		// when a used cpuset is marked reserved after rebooting bcs-cpuset-device,
		// just append into node.AllocatedCpuset, the cpuset will be remove from AllocatedCpuset,
		// but it will never be appended into node.CpuSet. See implementation of CpusetNode.ReleaseCpuset
		// in bkbcs/bcs-services/bcs-cpuset-device/types/types.go.
		// Based on the above, don't need filter reserved cpuset
		mNode.AllocatedCpuset = append(mNode.AllocatedCpuset, strings.Split(cpusets, ",")...)
	}

	return nil
}

// start listen docker api event
func (s *CpusetDevicePlugin) listenerDockerEvent() {
	listener := make(chan *docker.APIEvents)
	err := s.client.AddEventListener(listener)
	if err != nil {
		blog.Errorf("listen docker event error %s", err.Error())
		os.Exit(1)
	}

	defer func() {
		err = s.client.RemoveEventListener(listener)
		if err != nil {
			blog.Errorf("remove docker event error  %s", err.Error())
		}
	}()

	for {
		var msg *docker.APIEvents
		select {
		case msg = <-listener:
			blog.Infof("receive docker event action %s container %s", msg.Action, msg.ID)
			c, err := s.client.InspectContainer(msg.ID)
			if err != nil {
				blog.Errorf("inspect container %s error %s", msg.ID, err.Error())
				break
			}

			switch msg.Action {
			//start container
			case "start":
				//set container group cpuset
				s.setContainerCpuset(c)

			// stop container
			case "stop", "die":
				//release container cpuset resources
				s.releaseCpuset(c)
			}
		}
	}
}

// get container cpuset info
func (s *CpusetDevicePlugin) getContainerCpusetInfo(c *docker.Container) (string, string) {
	// cpuset env, example: node:0;cpuset:0,1,2,3
	var envValue string
	// docker env format []string
	// example: []string{"k1=v1","k2=v2"...}
	for _, o := range c.Config.Env {
		// envs[0] is key, envs[1] is value
		envs := strings.Split(o, "=")
		if envs[0] == EnvBkbcsAllocateCpuset {
			envValue = envs[1]
			break
		}
	}
	// if container don't contain bkbcs_allocate_cpuset env, then continue
	if envValue == "" {
		blog.Infof("container %s don't contain bkbcs_allocate_cpuset env, then continue", c.ID)
		return "", ""
	}
	blog.Infof("container %s contains env(%s=%s)", c.ID, EnvBkbcsAllocateCpuset, envValue)

	// node:0;cpuset:0,1,2,3
	values := strings.Split(envValue, ";")
	// node:0, nv[0]=node, nv[1]=0
	nv := strings.Split(values[0], ":")
	// 0
	node := nv[1]
	// cpuset:0,1,2,3
	// cv[0]=cpuset, cv[1]=0,1,2,3
	cv := strings.Split(values[1], ":")
	// 0,1,2,3
	cpusets := cv[1]

	return node, cpusets
}

func (s *CpusetDevicePlugin) setContainerCpuset(c *docker.Container) {
	node, cpusets := s.getContainerCpusetInfo(c)
	if node == "" || cpusets == "" {
		return
	}
	// set container cgroup cpuset.cpus、cpuset.mems
	cpus := fmt.Sprintf("%s/%s/cpuset.cpus", CgroupCpusetRoot, c.ID)
	fcpus, err := os.Create(cpus)
	if err != nil {
		blog.Errorf("open file %s error %s", cpus, err.Error())
		return
	}
	defer fcpus.Close()
	_, err = fcpus.WriteString(cpusets)
	if err != nil {
		blog.Errorf("write file %s error %s", cpus, err.Error())
		return
	}

	mems := fmt.Sprintf("%s/%s/cpuset.mems", CgroupCpusetRoot, c.ID)
	fmems, err := os.Create(mems)
	if err != nil {
		blog.Errorf("open file %s error %s", cpus, err.Error())
		return
	}
	defer fmems.Close()
	_, err = fmems.WriteString(node)
	if err != nil {
		blog.Errorf("write file %s error %s", mems, err.Error())
	}
	blog.Infof("set container %s cpuset(%s) success", c.ID, cpusets)
}

func (s *CpusetDevicePlugin) releaseCpuset(c *docker.Container) {
	node, cpusets := s.getContainerCpusetInfo(c)
	if node == "" || cpusets == "" {
		return
	}

	mNode, ok := s.nodes[node]
	if !ok {
		blog.Errorf("container %s node %s not found", c.ID, node)
		return
	}
	mNode.ReleaseCpuset(strings.Split(cpusets, ","))
}

func (s *CpusetDevicePlugin) lockNodes() {
	for _, node := range s.nodes {
		node.Lock()
	}
}

func (s *CpusetDevicePlugin) unlockNodes() {
	for _, node := range s.nodes {
		node.Unlock()
	}
}
