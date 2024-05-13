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

package cloudagent

import (
	"context"
	"fmt"

	"github.com/containernetworking/cni/pkg/skel"
	"google.golang.org/grpc"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	netservicetype "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	pbnetagent "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetagent"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// Client client for bcs-cloud-agent
type Client struct {
	agentClient pbnetagent.CloudNetagentClient
}

// NewClient create CloudAgentClient
func NewClient(endpoint string) (*Client, error) {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("grpc dial failed, err %s", err.Error())
	}
	agentClient := pbnetagent.NewCloudNetagentClient(conn)
	return &Client{
		agentClient: agentClient,
	}, nil
}

// Alloc alloc ip
func (cac *Client) Alloc(args *skel.CmdArgs) (*netservicetype.IPInfo, error) {
	k8sConf, err := LoadK8sArgs(args)
	if err != nil {
		blog.Errorf("failed to LoadK8sArgs, err %s", err.Error())
		return nil, fmt.Errorf("failed to LoadK8sArgs, err %s", err.Error())
	}

	resp, err := cac.agentClient.AllocIP(context.Background(), &pbnetagent.AllocIPReq{
		Seq:          common.TimeSequence(),
		ContainerID:  args.ContainerID,
		PodName:      string(k8sConf.K8S_POD_NAME),
		PodNamespace: string(k8sConf.K8S_POD_NAMESPACE),
		IpAddr:       k8sConf.IP.String(),
	})
	if err != nil {
		blog.Errorf("AllocIP failed, err %s", err.Error())
		return nil, fmt.Errorf("AllocIP failed, err %s", err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
		blog.Errorf("AllocIP not ok, resp %+v", resp)
		return nil, fmt.Errorf("AllocIP not ok, resp %+v", resp)
	}

	return &netservicetype.IPInfo{
		IPAddr:  resp.IpAddr,
		MacAddr: resp.MacAddr,
		Gateway: resp.Gateway,
		Mask:    int(resp.Mask),
	}, nil
}

// Release release ip
func (cac *Client) Release(args *skel.CmdArgs) error {
	k8sConf, err := LoadK8sArgs(args)
	if err != nil {
		blog.Errorf("failed to LoadK8sArgs, err %s", err.Error())
		return fmt.Errorf("failed to LoadK8sArgs, err %s", err.Error())
	}

	resp, err := cac.agentClient.ReleaseIP(context.Background(), &pbnetagent.ReleaseIPReq{
		Seq:          common.TimeSequence(),
		ContainerID:  args.ContainerID,
		PodName:      string(k8sConf.K8S_POD_NAME),
		PodNamespace: string(k8sConf.K8S_POD_NAMESPACE),
	})
	if err != nil {
		blog.Errorf("ReleaseIP failed, err %s", err.Error())
		return fmt.Errorf("ReleaseIP failed, err %s", err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
		blog.Errorf("ReleaseIP not ok, resp %+v", resp)
		return fmt.Errorf("ReleaseIP not ok, resp %+v", resp)
	}

	return nil
}
