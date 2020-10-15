/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"context"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netcontroller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-network/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-network/pkg/common"
)

// IPCleaner ip cleaner
type IPCleaner struct {
	kubeClient client.Client

	cleanInterval time.Duration

	maxReservedTime time.Duration

	option *option.ControllerOption

	cloudNetClient pbcloudnet.CloudNetserviceClient

	isDoing bool
}

// NewIPCleaner create new ip cleaner
func NewIPCleaner(r client.Client,
	option *option.ControllerOption,
	cloudNetClient pbcloudnet.CloudNetserviceClient) *IPCleaner {

	return &IPCleaner{
		kubeClient:      r,
		cloudNetClient:  cloudNetClient,
		cleanInterval:   time.Duration(option.IPCleanCheckMinute) * time.Minute,
		maxReservedTime: time.Duration(option.IPCleanMaxReservedMinute) * time.Minute,
	}
}

// Run run clean routine
func (ic *IPCleaner) Run(ctx context.Context) {
	blog.Infof("run ip cleaner")
	ticker := time.NewTicker(ic.cleanInterval)
	for {
		select {
		case <-ticker.C:
			ic.handle()
		case <-ctx.Done():
			blog.Infof("ip cleaner context done, wait for cleanr done")
			for {
				if ic.isDoing {
					time.Sleep(time.Second)
					continue
				}
				break
			}
		}
	}
}

func (ic *IPCleaner) handle() {
	ic.isDoing = true
	defer func() {
		ic.isDoing = false
	}()
	ic.transIPStatus()
	ic.doClean()
}

func (ic *IPCleaner) transIPStatus() {
	cloudIPList := &cloudv1.CloudIPList{}
	if err := ic.kubeClient.List(context.TODO(), cloudIPList, &client.MatchingLabels{
		constant.IP_LABEL_KEY_FOR_IS_FIXED:         strconv.FormatBool(true),
		constant.IP_LABEL_KEY_FOR_STATUS:           constant.IP_STATUS_AVAILABLE,
		constant.IP_LABEL_KEY_FOR_IS_CLUSTER_LAYER: strconv.FormatBool(true),
	}); err != nil {
		blog.Errorf("unable list available fixed ips, err %s", err.Error())
		return
	}

	for _, cloudIP := range cloudIPList.Items {
		switch strings.ToLower(cloudIP.Spec.WorkloadKind) {
		case strings.ToLower("statefulset"):
			var sts appsv1.StatefulSet
			stsNamespacedName := k8stypes.NamespacedName{
				Namespace: cloudIP.Spec.Namespace,
				Name:      cloudIP.Spec.WorkloadName,
			}
			err := ic.kubeClient.Get(context.TODO(), stsNamespacedName, &sts)
			if err != nil {
				// unexpected errors
				if !k8serrors.IsNotFound(err) {
					blog.Warnf("get statefulset %s failed, err %s", stsNamespacedName.String(), err.Error())
					continue
				}

				blog.V(2).Infof("trans cloud %v to deleting status", cloudIP)
				// set status to deleting when workload is deleted
				timeNow := time.Now()
				cloudIP.Labels[constant.IP_LABEL_KEY_FOR_STATUS] = constant.IP_STATUS_DELETING
				cloudIP.Status.Status = constant.IP_STATUS_DELETING
				cloudIP.Status.UpdateTime = common.FormatTime(timeNow)
				err = ic.kubeClient.Update(context.TODO(), &cloudIP)
				if err != nil {
					blog.Warnf("update ip %s/%s to deleting failed, err %s",
						cloudIP.GetName(), cloudIP.GetNamespace(), err.Error())
					continue
				}
			}
		case strings.ToLower("gamestatefulset"):
			continue
		}
	}
}

func (ic *IPCleaner) doClean() {
	deletingCloudIPList := &cloudv1.CloudIPList{}
	if err := ic.kubeClient.List(context.TODO(), deletingCloudIPList, &client.MatchingLabels{
		constant.IP_LABEL_KEY_FOR_IS_FIXED:         strconv.FormatBool(true),
		constant.IP_LABEL_KEY_FOR_STATUS:           constant.IP_STATUS_DELETING,
		constant.IP_LABEL_KEY_FOR_IS_CLUSTER_LAYER: strconv.FormatBool(true),
	}); err != nil {
		blog.Errorf("unable list deleted fixed ips, err %s", err.Error())
		return
	}
	for _, cloudIP := range deletingCloudIPList.Items {
		timeNow := time.Now()
		updateTime, err := common.ParseTimeString(cloudIP.Status.UpdateTime)
		if err != nil {
			blog.Warnf("failed to parse deleted cloud ip %s/%s update time %s, err %s",
				cloudIP.GetName(), cloudIP.GetNamespace(), cloudIP.Status.UpdateTime, err.Error())
			continue
		}
		if timeNow.Sub(updateTime) > (ic.maxReservedTime) {
			// clean fixed ip
			blog.V(2).Infof("clean out-of-fashion fixed ip %+v", cloudIP)
			resp, err := ic.cloudNetClient.CleanFixedIP(context.Background(), &pbcloudnet.CleanFixedIPReq{
				Seq:          common.TimeSequence(),
				Region:       cloudIP.Spec.Region,
				Cluster:      cloudIP.Spec.Cluster,
				Namespace:    cloudIP.Spec.Namespace,
				WorkloadName: cloudIP.Spec.WorkloadName,
				WorkloadKind: cloudIP.Spec.WorkloadKind,
				Address:      cloudIP.Spec.Address,
			})
			if err != nil {
				blog.Warnf("failed to clean fixed ip to cloud netservice, err %s", err.Error())
				continue
			}
			if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
				blog.Warnf("failed to clean fixed ip, resp %+v", resp)
				continue
			}

			// delete fixed ip from apiserver
			err = ic.kubeClient.Delete(context.TODO(), &cloudIP)
			if err != nil {
				blog.Warnf("failed to delete ip %s/%s from apiserver, err %s",
					cloudIP.GetName(), cloudIP.GetNamespace(), err.Error())
			}
		}
	}
}
