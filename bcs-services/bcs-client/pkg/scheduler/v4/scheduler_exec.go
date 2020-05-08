package v4

import (
	"bk-bcs/bcs-common/common/codec"
	"bk-bcs/bcs-services/bcs-client/pkg/types"
	"context"
	"fmt"
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
		//fmt.Sprintf(bcsSchedulerCreateExecUri, bs.bcsAPIAddress, hostIp),
		fmt.Sprintf(bcsSchedulerCreateExecUri, bs.bcsAPIAddress),
		http.MethodPost,
		data,
		getClusterIDHeader(clusterId),
	)

	var execResp CreatExecResp
	err = codec.DecJson(resp, &execResp)
	return execResp.Id, err
}

func (bs *bcsScheduler) StartContainerExec(ctx context.Context, clusterId, execId, containerId, hostIp string) (types.HijackedResponse, error) {
	//uri := fmt.Sprintf(bcsSchedulerStartExecUri, bs.bcsAPIAddress, hostIp, containerId, execId)
	uri := fmt.Sprintf(bcsSchedulerStartExecUri, bs.bcsAPIAddress, containerId, execId)
	//return bs.requester.PostHijacked(ctx, uri, getClusterIDHeader(clusterId))

	return bs.requester.DoWebsocket(uri, getClusterIDHeader(clusterId))
}

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
		//fmt.Sprintf(bcsSchedulerResizeExecUri, bs.bcsAPIAddress, hostIp),
		fmt.Sprintf(bcsSchedulerResizeExecUri, bs.bcsAPIAddress),
		http.MethodPost,
		data,
		getClusterIDHeader(clusterId),
	)
	return err
}
