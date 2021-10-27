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

package cbs

import (
	"net/http"
	"os"
	"path"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/dbdd4us/qcloudapi-sdk-go/metadata"
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"
)

var (
	DiskByIdDevicePath       = "/dev/disk/by-id"
	DiskByIdDeviceNamePrefix = "virtio-"
)

type cbsNode struct {
	metadataClient *metadata.MetaData
	cbsClient      *cbs.Client
	mounter        mount.SafeFormatAndMount
}

func newCbsNode(secretId, secretKey, region string) (*cbsNode, error) {
	client, err := cbs.NewClient(common.NewCredential(secretId, secretKey), region, profile.NewClientProfile())
	if err != nil {
		return nil, err
	}

	node := cbsNode{
		metadataClient: metadata.NewMetaData(http.DefaultClient),
		cbsClient:      client,
		mounter: mount.SafeFormatAndMount{
			Interface: mount.New(""),
			Exec:      mount.NewOsExec(),
		},
	}
	return &node, nil
}

func (node *cbsNode) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is empty")
	}
	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "volume staging target path is empty")
	}
	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "volume has no capabilities")
	}
	if req.VolumeCapability.GetMount() == nil {
		return nil, status.Error(codes.InvalidArgument, "volume access type is not mount")
	}

	diskId := req.VolumeId

	stagingTargetPath := req.StagingTargetPath

	mountFlags := req.VolumeCapability.GetMount().MountFlags
	mountFsType := req.VolumeCapability.GetMount().FsType

	diskDevicePath := path.Join(DiskByIdDevicePath, DiskByIdDeviceNamePrefix+diskId)

	if _, err := os.Stat(stagingTargetPath); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(stagingTargetPath, 0750)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if err := node.mounter.FormatAndMount(diskDevicePath, stagingTargetPath, mountFsType, mountFlags); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (node *cbsNode) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "volume staging target path is empty")
	}

	stagingTargetPath := req.StagingTargetPath

	if err := node.mounter.Unmount(stagingTargetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (node *cbsNode) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is empty")
	}
	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "volume staging target path is empty")
	}
	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "volume target path is empty")
	}
	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "volume has no capabilities")
	}

	if req.VolumeCapability.GetMount() == nil {
		return nil, status.Error(codes.InvalidArgument, "volume access type is not mount")
	}

	source := req.StagingTargetPath
	target := req.TargetPath

	mountFlags := req.VolumeCapability.GetMount().MountFlags
	mountFlags = append(mountFlags, "bind")

	if req.Readonly {
		mountFlags = append(mountFlags, "ro")
	}

	mountFsType := req.VolumeCapability.GetMount().FsType

	if mountFsType == "" {
		mountFsType = "ext4"
	}

	if _, err := os.Stat(target); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(target, 0750)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if err := node.mounter.Mount(source, target, mountFsType, mountFlags); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (node *cbsNode) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "volume target path is empty")
	}

	targetPath := req.TargetPath

	if err := node.mounter.Unmount(targetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (node *cbsNode) NodeGetId(context.Context, *csi.NodeGetIdRequest) (*csi.NodeGetIdResponse, error) {
	nodeId, err := node.metadataClient.InstanceID()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodeGetIdResponse{
		NodeId: nodeId,
	}, nil
}

func (node *cbsNode) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{Capabilities: []*csi.NodeServiceCapability{{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
			},
		},
	}}}, nil
}

func (node *cbsNode) NodeGetInfo(context.Context, *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
