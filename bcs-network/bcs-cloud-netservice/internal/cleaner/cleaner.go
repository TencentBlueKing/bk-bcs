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

package cleaner

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-network/pkg/leaderelection"
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
}

// NewIPCleaner create ip cleaner
func NewIPCleaner(maxIdleTime time.Duration,
	checkInterval time.Duration,
	storeIf store.Interface,
	cloudIf cloud.Interface,
	elector *leaderelection.Client) *IPCleaner {
	return &IPCleaner{
		storeIf:       storeIf,
		cloudIf:       cloudIf,
		elector:       elector,
		maxIdleTime:   maxIdleTime,
		checkInterval: checkInterval,
	}
}

// Run run cleaner
func (i *IPCleaner) Run(ctx context.Context) error {
	blog.Infof("run ip cleaner")
	timer := time.NewTicker(i.checkInterval)
	for {
		select {
		case <-timer.C:
			if i.elector.IsMaster() {
				blog.Infof("do search and clean")
				i.searchAndClean()
			}
		case <-ctx.Done():
			blog.Infof("ip cleaner context done")
			return nil
		}
	}
}

func (i *IPCleaner) searchAndClean() {
	ipObjs, err := i.storeIf.ListIPObject(context.Background(), map[string]string{
		kube.CrdNameLabelsStatus:   types.IP_STATUS_AVAILABLE,
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
				blog.Warnf("do clean %+v failed, err %s", ipObj, err.Error())
				continue
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	// clean dirty data
	ipObjsDeleting, err := i.storeIf.ListIPObject(context.Background(), map[string]string{
		kube.CrdNameLabelsStatus:   types.IP_STATUS_DELETING,
		kube.CrdNameLabelsIsFixed: strconv.FormatBool(false),
	})
	if err != nil {
		blog.Warnf("list deleting non-fixed ip objects failed, err %s", err.Error())
		return
	}
	for _, ipObj := range ipObjsDeleting {
		if err := i.transStatus(ipObj); err != nil {
			blog.Warnf("transStatus deleting ip %+v failed, err %s", ipObj, err.Error())
		}
		if err := i.releaseFromCloud(ipObj); err != nil {
			blog.Warnf("releaseFromCloud deleting ip %+v failed, err %s", ipObj, err.Error())
		}
		if err := i.deleteFromStore(ipObj); err != nil {
			blog.Warnf("deleteFromStore deleting ip %+v failed, err %s", ipObj, err.Error())
		}
	}
}

func (i *IPCleaner) doClean(ipObj *types.IPObject) error {
	if err := i.transStatus(ipObj); err != nil {
		return err
	}
	if err := i.releaseFromCloud(ipObj); err != nil {
		return err
	}
	if err := i.deleteFromStore(ipObj); err != nil {
		return err
	}
	return nil
}

func (i *IPCleaner) transStatus(ipObj *types.IPObject) error {
	ipObj.Status = types.IP_STATUS_DELETING
	err := i.storeIf.UpdateIPObject(context.Background(), ipObj)
	if err != nil {
		blog.Errorf("change ip object %+v to deleting status failed, err %s", ipObj, err.Error())
		return fmt.Errorf("change ip object %+v to deleting status failed, err %s", ipObj, err.Error())
	}
	return nil
}

func (i *IPCleaner) releaseFromCloud(ipObj *types.IPObject) error {
	err := i.cloudIf.UnassignIPFromEni(ipObj.Address, ipObj.EniID)
	if err != nil {
		blog.Errorf("unassign ip %s from eni %s failed, err %s", ipObj.Address, ipObj.EniID, err.Error())
		return fmt.Errorf("unassign ip %s from eni %s failed, err %s", ipObj.Address, ipObj.EniID, err.Error())
	}
	return nil
}

func (i *IPCleaner) deleteFromStore(ipObj *types.IPObject) error {
	err := i.storeIf.DeleteIPObject(context.Background(), ipObj.Address)
	if err != nil {
		blog.Errorf("delete ip %s from store failed, err %s", ipObj.Address, err.Error())
		return fmt.Errorf("delete ip %s from store failed, err %s", ipObj.Address, err.Error())
	}
	return nil
}
