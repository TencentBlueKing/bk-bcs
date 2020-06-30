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

package v4

import (
	"context"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"
	"net/http"
)

type CreateExecReq struct {
	ContainerId string   `json:"container_id"`
	Cmd         []string `json:"cmd"`
}

type ResizeExecReq struct {
	ExecId string `json:"exec_id"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type CreatExecResp struct {
	Id string `json:"Id"`
}

// call create_exec api of consoleproxy
func (bs *bcsScheduler) CreateContainerExec(clusterId, containerId, hostIp string, command []string) (string, error) {
	execReq := &CreateExecReq{
		ContainerId: containerId,
		Cmd:         command,
	}
	var data []byte
	err := codec.EncJson(execReq, &data)
	if err != nil {
		return "", err
	}

	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerCreateExecUri, bs.bcsAPIAddress, hostIp),
		http.MethodPost,
		data,
		getClusterIDHeader(clusterId),
	)

	var execResp CreatExecResp
	err = codec.DecJson(resp, &execResp)
	return execResp.Id, err
}

// call start_exec api of consoleproxy
func (bs *bcsScheduler) StartContainerExec(ctx context.Context, clusterId, execId, containerId, hostIp string) (types.HijackedResponse, error) {
	uri := fmt.Sprintf(bcsSchedulerStartExecUri, bs.bcsAPIAddress, hostIp, containerId, execId)

	//return bs.requester.PostHijacked(ctx, uri, getClusterIDHeader(clusterId))
	return bs.requester.DoWebsocket(uri, getClusterIDHeader(clusterId))
}

// call resize_exec api of consoleproxy
func (bs *bcsScheduler) ResizeContainerExec(clusterId, execId, hostIp string, height, width int) error {
	resizeReq := &ResizeExecReq{
		ExecId: execId,
		Height: height,
		Width:  width,
	}
	var data []byte
	err := codec.EncJson(resizeReq, &data)
	if err != nil {
		return err
	}

	_, err = bs.requester.Do(
		fmt.Sprintf(bcsSchedulerResizeExecUri, bs.bcsAPIAddress, hostIp),
		http.MethodPost,
		data,
		getClusterIDHeader(clusterId),
	)
	return err
}
