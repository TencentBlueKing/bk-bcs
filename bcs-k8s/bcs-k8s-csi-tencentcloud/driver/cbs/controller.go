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
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	GB = 1 << (10 * 3)

	// cbs disk type
	DiskTypeAttr = "diskType"

	// cbs disk name
	DiskNameAttr = "diskName"
	// cbs disk tags
	DiskTagsAttr = "diskTags"

	// tencentcloud cbs types, now support CLOUD_BASIC, CLOUD_PREMIUM, CLOUD_SSD
	DiskTypeCloudBasic   = "CLOUD_BASIC"
	DiskTypeCloudPremium = "CLOUD_PREMIUM"
	DiskTypeCloudSsd     = "CLOUD_SSD"

	DiskTypeDefault = DiskTypeCloudBasic

	// cbs disk charge type
	DiskChargeTypeAttr           = "diskChargeType"
	DiskChargeTypePrePaid        = "PREPAID"
	DiskChargeTypePostPaidByHour = "POSTPAID_BY_HOUR"

	DiskChargeTypeDefault = DiskChargeTypePostPaidByHour

	// cbs disk charge prepaid options
	DiskChargePrepaidPeriodAttr = "diskChargeTypePrepaidPeriod"

	DiskChargePrepaidPeriodValidValues = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36}
	DiskChargePrepaidPeriodDefault     = 1

	DiskChargePrepaidRenewFlagAttr = "diskChargePrepaidRenewFlag"

	DiskChargePrepaidRenewFlagNotifyAndAutoRenew          = "NOTIFY_AND_AUTO_RENEW"
	DiskChargePrepaidRenewFlagNotifyAndManualRenewd       = "NOTIFY_AND_MANUAL_RENEW"
	DiskChargePrepaidRenewFlagDisableNotifyAndManualRenew = "DISABLE_NOTIFY_AND_MANUAL_RENEW"

	DiskChargePrepaidRenewFlagDefault = DiskChargePrepaidRenewFlagNotifyAndManualRenewd

	// cbs disk encrypt
	EncryptAttr   = "encrypt"
	EncryptEnable = "ENCRYPT"

	//cbs disk zone
	DiskZone = "diskZone"

	//cbs disk zones
	DiskZones = "diskZones"

	//cbs disk asp Id
	AspId = "aspId"
	// cbs status
	StatusUnattached = "UNATTACHED"
	StatusAttached   = "ATTACHED"
)

type cbsController struct {
	cbsClient *cbs.Client
	zone      string
}

//newCbsController create cbsController object
func newCbsController(secretId, secretKey, region, zone, cbsUrl string) (*cbsController, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = cbsUrl
	client, err := cbs.NewClient(common.NewCredential(secretId, secretKey), region, cpf)
	if err != nil {
		return nil, err
	}

	return &cbsController{
		cbsClient: client,
		zone:      zone,
	}, nil
}

//CreateVolume implements the csi grpc interface to create volume with tencentcloud api
func (ctrl *cbsController) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "volume name is empty")
	}

	volumeIdempotencyName := req.Name
	volumeCapacity := req.CapacityRange.RequiredBytes

	if len(req.VolumeCapabilities) <= 0 {
		return nil, status.Error(codes.InvalidArgument, "volume has no capabilities")
	}

	for _, c := range req.VolumeCapabilities {
		if c.GetBlock() != nil {
			return nil, status.Error(codes.InvalidArgument, "block volume is not supported")
		}
		if c.AccessMode.Mode != csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
			return nil, status.Error(codes.InvalidArgument, "block access mode only support singer node writer")
		}
	}

	// get volume type from storageclass parameters
	volumeType, ok := req.Parameters[DiskTypeAttr]
	if !ok {
		volumeType = DiskTypeDefault
	}

	// get volume name from storageclass parameters
	volumeName, ok := req.Parameters[DiskNameAttr]
	if !ok {
		volumeName = ""
	}

	// get volume tags from storageclass parameters
	volumeTags, ok := req.Parameters[DiskTagsAttr]
	if !ok {
		volumeTags = ""
	}

	// validate volume tags
	var cbsTags []*cbs.Tag
	if volumeTags != "" {
		volumeTagArray := strings.Split(volumeTags, ",")
		for _, volumeTag := range volumeTagArray {
			tag := strings.Split(volumeTag, ":")
			if len(tag) != 2 {
				return nil, status.Errorf(codes.InvalidArgument, "cbs tag not valid")
			}
			cbsTag := new(cbs.Tag)
			cbsTag.Key = &tag[0]
			cbsTag.Value = &tag[1]
			cbsTags = append(cbsTags, cbsTag)
		}
	}

	// validate disk type, now only support 3 disk types
	if volumeType != DiskTypeCloudBasic && volumeType != DiskTypeCloudPremium && volumeType != DiskTypeCloudSsd {
		return nil, status.Error(codes.InvalidArgument, "cbs type not supported")
	}

	volumeChargeType, ok := req.Parameters[DiskChargeTypeAttr]
	if !ok {
		volumeChargeType = DiskChargeTypeDefault
	}

	var volumeChargePrepaidPeriod int
	var volumeChargePrepaidRenewFlag string

	if volumeChargeType == DiskChargeTypePrePaid {
		var err error
		var ok bool
		volumeChargePrepaidPeriodStr, ok := req.Parameters[DiskChargePrepaidPeriodAttr]
		if !ok {
			volumeChargePrepaidPeriodStr = strconv.Itoa(DiskChargePrepaidPeriodDefault)
		}

		volumeChargePrepaidPeriod, err = strconv.Atoi(volumeChargePrepaidPeriodStr)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "prepaid period not valid")
		}

		found := false

		for _, p := range DiskChargePrepaidPeriodValidValues {
			if p == volumeChargePrepaidPeriod {
				found = true
			}
		}

		if !found {
			return nil, status.Error(codes.InvalidArgument, "can not found valid prepaid period")
		}

		volumeChargePrepaidRenewFlag, ok = req.Parameters[DiskChargePrepaidRenewFlagAttr]
		if !ok {
			volumeChargePrepaidRenewFlag = DiskChargePrepaidRenewFlagDefault
		}
		if volumeChargePrepaidRenewFlag != DiskChargePrepaidRenewFlagDisableNotifyAndManualRenew && volumeChargePrepaidRenewFlag != DiskChargePrepaidRenewFlagNotifyAndAutoRenew && volumeChargePrepaidRenewFlag != DiskChargePrepaidRenewFlagNotifyAndManualRenewd { // nolint
			return nil, status.Error(codes.InvalidArgument, "invalid renew flag")
		}

	}

	volumeEncrypt, ok := req.Parameters[EncryptAttr]
	if !ok {
		volumeEncrypt = ""
	}

	if volumeEncrypt != "" && volumeEncrypt != EncryptEnable {
		return nil, status.Error(codes.InvalidArgument, "volume encrypt not valid")
	}

	// build tencentcloud csi api request
	createCbsReq := cbs.NewCreateDisksRequest()

	createCbsReq.ClientToken = &volumeIdempotencyName
	createCbsReq.DiskType = &volumeType
	createCbsReq.DiskChargeType = &volumeChargeType

	// set volume name to request
	if volumeName != "" {
		glog.Infof("set volume name %s", volumeName)
		createCbsReq.DiskName = &volumeName
	}

	// set volume tags to request
	if len(cbsTags) > 0 {
		for _, cbsTagForLog := range cbsTags {
			glog.Infof("set volume tags for volume %s, key: %s, value: %s", volumeName, *cbsTagForLog.Key, *cbsTagForLog.Value)
		}

		createCbsReq.Tags = cbsTags
	}

	if volumeChargeType == DiskChargeTypePrePaid {
		period := uint64(volumeChargePrepaidPeriod)
		createCbsReq.DiskChargePrepaid = &cbs.DiskChargePrepaid{
			Period:    &period,
			RenewFlag: &volumeChargePrepaidRenewFlag,
		}
	}

	gb := uint64(volumeCapacity / int64(GB))

	createCbsReq.DiskSize = &gb

	if volumeEncrypt == EncryptEnable {
		createCbsReq.Encrypt = &EncryptEnable
	}

	//zone parameters
	volumeZone, ok1 := req.Parameters[DiskZone]
	volumeZones, ok2 := req.Parameters[DiskZones]
	if ok1 && ok2 {
		return nil, status.Error(codes.InvalidArgument, "both zone and zones StorageClass parameters must not be used at the same time")
	}
	if !ok1 && !ok2 {
		volumeZone = ctrl.zone
	}

	if !ok1 && ok2 {
		zonesSlice := strings.Split(volumeZones, ",")
		hash := rand.Uint32()
		volumeZone = zonesSlice[hash%uint32(len(zonesSlice))]
	}

	createCbsReq.Placement = &cbs.Placement{
		Zone: &volumeZone,
	}

	//aspId parameters
	//zone parameters
	aspId, ok1 := req.Parameters[AspId]
	if !ok {
		aspId = ""
	}

	createCbsResponse, err := ctrl.cbsClient.CreateDisks(createCbsReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(createCbsResponse.Response.DiskIdSet) <= 0 {
		return nil, status.Errorf(codes.Internal, "create disk failed, no disk id found in create disk response, request id %s", *createCbsResponse.Response.RequestId)
	}

	diskId := *createCbsResponse.Response.DiskIdSet[0]

	disk := new(cbs.Disk)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	for {
		select {
		case <-ticker.C:
			listCbsRequest := cbs.NewDescribeDisksRequest()
			listCbsRequest.DiskIds = []*string{&diskId}

			listCbsResponse, err := ctrl.cbsClient.DescribeDisks(listCbsRequest)
			if err != nil {
				continue
			}
			if len(listCbsResponse.Response.DiskSet) >= 1 {
				for _, d := range listCbsResponse.Response.DiskSet {
					if *d.DiskId == diskId && d.DiskState != nil {
						if *d.DiskState == StatusAttached || *d.DiskState == StatusUnattached {
							disk = d
							if aspId != "" {
								bindReq := cbs.NewBindAutoSnapshotPolicyRequest()
								bindReq.AutoSnapshotPolicyId = &aspId
								bindReq.DiskIds = []*string{disk.DiskId}
								_, err := ctrl.cbsClient.BindAutoSnapshotPolicy(bindReq)
								if err != nil {

								}
							}
							return &csi.CreateVolumeResponse{
								Volume: &csi.Volume{
									Id:            *disk.DiskId,
									CapacityBytes: int64(int(*disk.DiskSize) * GB),
								},
							}, nil
						}
					}
				}
			}
		case <-ctx.Done():
			return nil, status.Error(codes.DeadlineExceeded, "cbs disk is not ready before deadline exceeded")
		}
	}
}

//DeleteVolume implements the csi grpc interface to delete volume with tencentcloud api
func (ctrl *cbsController) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is empty")
	}

	describeDiskRequest := cbs.NewDescribeDisksRequest()
	describeDiskRequest.DiskIds = []*string{&req.VolumeId}
	describeDiskResponse, err := ctrl.cbsClient.DescribeDisks(describeDiskRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(describeDiskResponse.Response.DiskSet) <= 0 {
		return &csi.DeleteVolumeResponse{}, nil
	}

	terminateCbsRequest := cbs.NewTerminateDisksRequest()
	terminateCbsRequest.DiskIds = []*string{&req.VolumeId}

	_, err = ctrl.cbsClient.TerminateDisks(terminateCbsRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.DeleteVolumeResponse{}, nil
}

//ControllerPublishVolume implements the csi grpc interface to attach volume with tencentcloud api
func (ctrl *cbsController) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is empty")
	}
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node id is empty")
	}

	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "volume has no capabilities")
	}

	diskId := req.VolumeId
	instanceId := req.NodeId

	listCbsRequest := cbs.NewDescribeDisksRequest()
	listCbsRequest.DiskIds = []*string{&diskId}

	listCbsResponse, err := ctrl.cbsClient.DescribeDisks(listCbsRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(listCbsResponse.Response.DiskSet) <= 0 {
		return nil, status.Error(codes.NotFound, "disk not found")
	}

	for _, disk := range listCbsResponse.Response.DiskSet {
		if *disk.DiskId == diskId {
			if *disk.DiskState == StatusAttached && *disk.InstanceId == instanceId {
				return &csi.ControllerPublishVolumeResponse{}, nil
			}
			if *disk.DiskState == StatusAttached && *disk.InstanceId != instanceId {
				return nil, status.Error(codes.FailedPrecondition, "disk is attach to another instance already")
			}
		}
	}

	attachDiskRequest := cbs.NewAttachDisksRequest()
	attachDiskRequest.DiskIds = []*string{&diskId}
	attachDiskRequest.InstanceId = &instanceId

	_, err = ctrl.cbsClient.AttachDisks(attachDiskRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ticker := time.NewTicker(time.Second * 5)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	for {
		select {
		case <-ticker.C:
			listCbsRequest := cbs.NewDescribeDisksRequest()
			listCbsRequest.DiskIds = []*string{&diskId}

			listCbsResponse, err := ctrl.cbsClient.DescribeDisks(listCbsRequest)
			if err != nil {
				continue
			}
			if len(listCbsResponse.Response.DiskSet) >= 1 {
				for _, d := range listCbsResponse.Response.DiskSet {
					if *d.DiskId == diskId && d.DiskState != nil {
						if *d.DiskState == StatusAttached {
							return &csi.ControllerPublishVolumeResponse{}, nil
						}
					}
				}
			}
		case <-ctx.Done():
			return nil, status.Error(codes.Internal, "cbs disk is not attached before deadline exceeded")
		}
	}
}

//ControllerUnpublishVolume implements the csi grpc interface to detach volume with tencentcloud api
func (ctrl *cbsController) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "volume id is empty")
	}
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node id is empty")
	}

	diskId := req.VolumeId

	listCbsRequest := cbs.NewDescribeDisksRequest()
	listCbsRequest.DiskIds = []*string{&diskId}

	listCbsResponse, err := ctrl.cbsClient.DescribeDisks(listCbsRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(listCbsResponse.Response.DiskSet) <= 0 {
		return nil, status.Error(codes.NotFound, "disk not found")
	}

	for _, disk := range listCbsResponse.Response.DiskSet {
		if *disk.DiskId == diskId {
			if *disk.DiskState == StatusUnattached {
				return &csi.ControllerUnpublishVolumeResponse{}, nil
			}
		}
	}

	detachDiskRequest := cbs.NewDetachDisksRequest()
	detachDiskRequest.DiskIds = []*string{&diskId}

	_, err = ctrl.cbsClient.DetachDisks(detachDiskRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ticker := time.NewTicker(time.Second * 5)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	for {
		select {
		case <-ticker.C:
			listCbsRequest := cbs.NewDescribeDisksRequest()
			listCbsRequest.DiskIds = []*string{&diskId}

			listCbsResponse, err := ctrl.cbsClient.DescribeDisks(listCbsRequest)
			if err != nil {
				continue
			}
			if len(listCbsResponse.Response.DiskSet) >= 1 {
				for _, d := range listCbsResponse.Response.DiskSet {
					if *d.DiskId == diskId && d.DiskState != nil {
						if *d.DiskState == StatusUnattached {
							return &csi.ControllerUnpublishVolumeResponse{}, nil
						}
					}
				}
			}
		case <-ctx.Done():
			return nil, status.Error(codes.Internal, "cbs disk is not unattached before deadline exceeded")
		}
	}
}

func (ctrl *cbsController) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (ctrl *cbsController) ValidateVolumeCapabilities(context.Context, *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ctrl *cbsController) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ctrl *cbsController) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ctrl *cbsController) CreateSnapshot(context.Context, *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ctrl *cbsController) DeleteSnapshot(context.Context, *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (ctrl *cbsController) ListSnapshots(context.Context, *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
