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

// Package cleaner is cleaner for cloud ip
package cleaner

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/leaderelection"
)

// IPCleaner ip cleaner
type IPCleaner struct {
	// client for store ip object and subnet
	storeIf store.Interface

	// cloud interface for operate eni ip
	cloudIf cloud.Interface

	// elector elector for leader election
	elector *leaderelection.Client

	// maxIdleTime max idle time for ip object
	maxIdleTime time.Duration

	// checkInterval interval for check idle ip
	checkInterval time.Duration

	// fixIPCheckInterval interval for check idle fixed ip
	fixIPCheckInterval time.Duration
}

// NewIPCleaner create ip cleaner
func NewIPCleaner(maxIdleTime time.Duration,
	checkInterval time.Duration,
	fixIPCheckInterval time.Duration,
	storeIf store.Interface,
	cloudIf cloud.Interface,
	elector *leaderelection.Client) *IPCleaner {
	return &IPCleaner{
		storeIf:            storeIf,
		cloudIf:            cloudIf,
		elector:            elector,
		maxIdleTime:        maxIdleTime,
		checkInterval:      checkInterval,
		fixIPCheckInterval: fixIPCheckInterval,
	}
}

// Run run cleaner
func (i *IPCleaner) Run(ctx context.Context) error {
	blog.Infof("run ip cleaner")
	timer := time.NewTicker(i.checkInterval)
	fixedIPTimer := time.NewTicker(i.fixIPCheckInterval)
	for {
		select {
		case <-timer.C:
			if i.elector.IsMaster() {
				blog.Infof("do search and clean")
				i.searchAndClean()
			}
		case <-fixedIPTimer.C:
			if i.elector.IsMaster() {
				blog.Infof("do search and clean fixed ip")
				i.searchAndCleanFixedIP()
			}
		case <-ctx.Done():
			blog.Infof("ip cleaner context done")
			return nil
		}
	}
}

func (i *IPCleaner) searchAndClean() {
	ipObjs, err := i.storeIf.ListIPObject(context.Background(), map[string]string{
		kube.CrdNameLabelsStatus:  types.IPStatusAvailable,
		kube.CrdNameLabelsIsFixed: strconv.FormatBool(false),
	})
	if err != nil {
		blog.Warnf("list available non-fixed ip objects failed, err %s", err.Error())
		return
	}

	for _, ipObj := range ipObjs {
		now := time.Now()
		if now.Sub(ipObj.UpdateTime) > (i.maxIdleTime) {
			if err := i.doClean(ipObj); err != nil {
				blog.Warnf("do clean ip %+v failed, err %s", ipObj, err.Error())
				continue
			}
			blog.Infof("cleaned ip %s due to exceed max idle time %f minute", ipObj.Address, i.maxIdleTime.Minutes())
		}
	}
}

func (i *IPCleaner) searchAndCleanFixedIP() {
	fixedIPObjs, err := i.storeIf.ListIPObject(context.Background(), map[string]string{
		kube.CrdNameLabelsStatus:  types.IPStatusAvailable,
		kube.CrdNameLabelsIsFixed: strconv.FormatBool(true),
	})
	if err != nil {
		blog.Warnf("list available fixed ip objects failed, err %s", err.Error())
		return
	}

	for _, ipObj := range fixedIPObjs {
		duration, err := time.ParseDuration(ipObj.KeepDuration)
		if err != nil {
			blog.Errorf("ip %s has invalid keep duration %s", ipObj.Address, ipObj.KeepDuration)
			continue
		}
		now := time.Now()
		if now.Sub(ipObj.UpdateTime) > duration {
			if err := i.freeFixedIPFromStore(ipObj); err != nil {
				blog.Warnf("do free fixed ip %v failed, err %s", ipObj, err.Error())
				continue
			}
			blog.Infof("freed fixed ip %s for pod %s/%s due to exceed keep duration %s",
				ipObj.Address, ipObj.PodName, ipObj.Namespace, ipObj.KeepDuration)
		}
	}
}

func (i *IPCleaner) doClean(ipObj *types.IPObject) error {
	deletingIPObj, err := i.transStatus(ipObj)
	if err != nil {
		return err
	}
	if err := i.releaseFromCloud(deletingIPObj); err != nil {
		return err
	}
	if err := i.freeFromStore(deletingIPObj); err != nil {
		return err
	}
	return nil
}

func (i *IPCleaner) transStatus(ipObj *types.IPObject) (*types.IPObject, error) {
	ipObj.Status = types.IPStatusDeleting
	newIPObj, err := i.storeIf.UpdateIPObject(context.Background(), ipObj)
	if err != nil {
		blog.Errorf("change ip object %+v to deleting status failed, err %s", ipObj, err.Error())
		return nil, fmt.Errorf("change ip object %+v to deleting status failed, err %s", ipObj, err.Error())
	}
	return newIPObj, nil
}

func (i *IPCleaner) releaseFromCloud(ipObj *types.IPObject) error {
	err := i.cloudIf.UnassignIPFromEni([]string{ipObj.Address}, ipObj.EniID)
	if err != nil {
		blog.Errorf("unassign ip %s from eni %s failed, err %s", ipObj.Address, ipObj.EniID, err.Error())
		return fmt.Errorf("unassign ip %s from eni %s failed, err %s", ipObj.Address, ipObj.EniID, err.Error())
	}
	return nil
}

func (i *IPCleaner) freeFromStore(ipObj *types.IPObject) error {
	ipObj.Status = types.IPStatusFree
	ipObj.EniID = ""
	ipObj.Host = ""
	ipObj.ContainerID = ""
	ipObj.Cluster = ""
	_, err := i.storeIf.UpdateIPObject(context.Background(), ipObj)
	if err != nil {
		blog.Errorf("set ip %s free to store failed, err %s", ipObj.Address, err.Error())
		return fmt.Errorf("set ip %s free to store failed, err %s", ipObj.Address, err.Error())
	}
	return nil
}

func (i *IPCleaner) freeFixedIPFromStore(ipObj *types.IPObject) error {
	ipObj.Status = types.IPStatusFree
	ipObj.EniID = ""
	ipObj.Host = ""
	ipObj.ContainerID = ""
	ipObj.IsFixed = false
	ipObj.KeepDuration = ""
	ipObj.Cluster = ""
	_, err := i.storeIf.UpdateIPObject(context.Background(), ipObj)
	if err != nil {
		blog.Errorf("set fixed ip %s free to store failed, err %s", ipObj.Address, err.Error())
		return fmt.Errorf("set fixed ip %s free to store failed, err %s", ipObj.Address, err.Error())
	}
	return nil
}
