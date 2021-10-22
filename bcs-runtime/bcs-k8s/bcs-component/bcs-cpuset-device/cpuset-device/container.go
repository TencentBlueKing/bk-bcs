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
	"path/filepath"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	docker "github.com/fsouza/go-dockerclient"
)

func (s *CpusetDevicePlugin) loopUpdateCpusetNodes() {
	for {
		time.Sleep(s.checkInterval)
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

	cpusetNodeMap := make(map[string]map[string]struct{})
	// traversal container for get allocated cpusets
	for _, container := range containers {
		info, err := s.client.InspectContainer(container.ID)
		if err != nil {
			blog.Errorf("inspect container %s failed %s, update interrupts", container.ID, err.Error())
			return err
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

		// update cpuset of container
		s.setContainerCpuset(info)

		// collect cpuset in running containers
		_, ok = cpusetNodeMap[node]
		if !ok {
			cpusetNodeMap[node] = make(map[string]struct{})
		}
		for _, cpuset := range strings.Split(cpusets, ",") {
			// collect cpuset in running containers
			cpusetNodeMap[node][cpuset] = struct{}{}
			// construct allocated cpuset
			found := false
			for _, aset := range mNode.AllocatedCpuset {
				if aset == cpuset {
					found = true
					break
				}
			}
			if !found {
				blog.Infof("recover allocated data node:%s,cpuset:%s", node, cpuset)
				mNode.AllocatedCpuset = append(mNode.AllocatedCpuset, cpuset)
				mNode.AllocatedCpusetTime[cpuset] = time.Now()
			}
		}
		blog.Infof("node %s AllocatedCpuset(%v) AllCpuset(%v)", mNode.Id, mNode.AllocatedCpuset, mNode.Cpuset)
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
	s.cgroupFileLock.Lock()
	defer s.cgroupFileLock.Unlock()
	node, cpusets := s.getContainerCpusetInfo(c)
	if node == "" || cpusets == "" {
		return
	}
	if c.HostConfig == nil {
		blog.Warnf("container %s hostconfig is empty, do nothing", c.ID)
		return
	}
	var cgroupParent string
	if len(c.HostConfig.CgroupParent) != 0 {
		cgroupParent = filepath.Join(s.conf.CgroupCpusetRoot, c.HostConfig.CgroupParent)
	} else {
		cgroupParent = s.conf.CgroupCpusetRoot
	}

	// set container cgroup cpuset.cpus、cpuset.mems
	cpus := fmt.Sprintf("%s/%s/cpuset.cpus", cgroupParent, c.ID)
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

	mems := fmt.Sprintf("%s/%s/cpuset.mems", cgroupParent, c.ID)
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
