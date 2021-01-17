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

package manager

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	nettypes "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"

	dockerclient "github.com/fsouza/go-dockerclient"
)

var (
	defaultContainerSock = "unix:///var/run/docker.sock"
)

//DirtyCheck check container existence, if container lost in record,
//releasing ip address from database for other container reuse.
func DirtyCheck() {
	driver, err := GetIPDriver()
	if err != nil {
		blog.Errorf("Get driver error in check mode, %s", err.Error())
		return
	}
	hostIP := util.GetIPAddress()
	hostInfo, err := driver.GetHostInfo(hostIP)
	if err != nil {
		blog.Errorf("Get host info in check mode failed, %s", err.Error())
		return
	}
	if hostInfo.Containers == nil || len(hostInfo.Containers) == 0 {
		blog.Infof("No active container & ip address in host %s, normal exit", hostIP)
		return
	}
	client, err := dockerclient.NewClient(defaultContainerSock)
	if err != nil {
		blog.Errorf("Create docker container client err, %s", err.Error())
		return
	}
	//Get all running container info from Container Runtime
	containers, err := client.ListContainers(dockerclient.ListContainersOptions{All: false})
	if err != nil {
		blog.Errorf("List all docker container err, %s", err.Error())
		return
	}
	//ready to clean map in HostInfo
	for _, container := range containers {
		if con, ok := hostInfo.Containers[container.ID]; ok {
			blog.Infof("Container %s is running with ip %s in host %s, skip.", container.ID, con.IPAddr, hostInfo.IPAddr)
			delete(hostInfo.Containers, container.ID)
		}
	}
	if len(hostInfo.Containers) == 0 {
		blog.Infof("No dirty Container data in storage, bcs-ipam check mode process finish.")
		return
	}
	//Now all left in HostInfo.Containers is dirty data
	//in database, ready to release ip address with Driver
	for containerID, ipInst := range hostInfo.Containers {
		if ipInst.Container != containerID {
			blog.Warnf("##container info mismatch warnning, host: %s, ip inst: %s", containerID, ipInst.Container)
		}
		ipInfo := &nettypes.IPInfo{}
		blog.Errorf("Host %s release dirty ip %s in container %s", hostIP, ipInst.IPAddr, containerID)
		err := driver.ReleaseIPAddr(hostIP, containerID, ipInfo)
		if err != nil {
			blog.Errorf("Host %s release container %s/%s err, %s", hostIP, containerID, ipInst.IPAddr, err.Error())
			continue
		}
	}
	blog.Info("bcs-ipam check mode process finish.")
}
