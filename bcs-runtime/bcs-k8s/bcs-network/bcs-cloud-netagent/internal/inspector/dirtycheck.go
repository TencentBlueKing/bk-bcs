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

package inspector

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"

	dockerclient "github.com/fsouza/go-dockerclient"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DirtyCheck check dirty ips on the node
func (nni *NodeNetworkInspector) DirtyCheck(nodeNetwork *cloudv1.NodeNetwork) error {
	if nodeNetwork == nil {
		blog.Warnf("nodeNetwork is empty")
		return nil
	}
	containerIDMap := make(map[string]*pbcommon.IPObject)
	for _, eniObj := range nodeNetwork.Status.Enis {
		if eniObj.Status == constant.NodeNetworkEniStatusReady {
			resp, err := nni.cloudNetClient.ListIP(context.Background(), &pbcloudnet.ListIPsReq{
				EniID:  eniObj.EniID,
				Status: constant.IPStatusActive,
			})
			if err != nil {
				return fmt.Errorf("list active ips for eni %s failed, err %s", eniObj.EniName, err.Error())
			}
			if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
				return fmt.Errorf("list active ips for eni %s failed, errCode %d, errMsg %s",
					eniObj.EniName, resp.ErrCode, resp.ErrMsg)
			}
			for _, ip := range resp.Ips {
				containerIDMap[ip.ContainerID] = ip
			}
		}
	}

	// create docker client
	dockerCli, err := dockerclient.NewClient(nni.option.DockerSock)
	if err != nil {
		return fmt.Errorf("failed to create docker client, err %s", err.Error())
	}
	// list all running container from docker runtime
	containers, err := dockerCli.ListContainers(dockerclient.ListContainersOptions{All: false})
	if err != nil {
		return fmt.Errorf("list all running docker container failed, err %s", err.Error())
	}
	runningContainerIDMap := make(map[string]struct{})
	for _, con := range containers {
		runningContainerIDMap[con.ID] = struct{}{}
	}

	// do clean
	for cid, ipObj := range containerIDMap {
		if _, ok := runningContainerIDMap[cid]; !ok {
			blog.Infof("container %s for active ip %s does not found on node, do clean", cid, ipObj.Address)
			err = nni.client.CloudIPs(ipObj.Namespace).Delete(context.Background(),
				ipObj.PodName, metav1.DeleteOptions{})
			if err != nil {
				if !k8serrors.IsNotFound(err) {
					blog.Warnf("delete ip %s/%s from local cluster failed, err %s",
						ipObj.PodName, ipObj.Namespace, err.Error())
					continue
				}
				blog.Warnf("ip %v not found in local cluster, ", ipObj)
			}
			resp, err := nni.cloudNetClient.ReleaseIP(context.Background(), &pbcloudnet.ReleaseIPReq{
				Seq:          common.TimeSequence(),
				Cluster:      ipObj.Cluster,
				PodName:      ipObj.PodName,
				PodNamespace: ipObj.Namespace,
				ContainerID:  ipObj.ContainerID,
			})
			if err != nil {
				blog.Warnf("release ip %v failed, err %s", ipObj, err.Error())
				continue
			}
			if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
				blog.Warnf("release ip %v failed, errCode %s, errMsg %s", ipObj, resp.ErrCode, resp.ErrMsg)
			}
			nni.GetIPCache().DeleteEniIPbyContainerID(ipObj.ContainerID)
		}
	}
	return nil
}
