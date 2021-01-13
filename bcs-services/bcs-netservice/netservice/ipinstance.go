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

package netservice

import (
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"path/filepath"
	"time"
)

//UpdateAvailableIPInstance update ip instance info.
//only support: MacAddress, App
func (srv *NetService) UpdateAvailableIPInstance(ipinst *types.IPInst) error {
	started := time.Now()
	//check ip instance path first, only focus on available ip instance
	instpath := filepath.Join(defaultPoolInfoPath, ipinst.Cluster, ipinst.Pool, "available", ipinst.IPAddr)
	if exist, _ := srv.store.Exist(instpath); !exist {
		blog.Errorf("Update Available ip instance failed, no instance data: %s", instpath)
		reportMetrics("updateAvailableIPInstance", stateNonExistFailure, started)
		return fmt.Errorf("lost available ip instance %s", ipinst.IPAddr)
	}
	//get ip instance data
	data, instErr := srv.store.Get(instpath)
	if instErr != nil {
		blog.Errorf("Get no ip instance data from path %s failed, %s", instpath, instErr)
		reportMetrics("updateAvailableIPInstance", stateStorageFailure, started)
		return fmt.Errorf("got instance data failed, %s", instErr)
	}
	oldInst := &types.IPInst{}
	if err := json.Unmarshal(data, oldInst); err != nil {
		blog.Errorf("UpdateAvailableIPInstance decode ip instance %s failed, %s", instpath, err)
		reportMetrics("updateAvailableIPInstance", stateJSONFailure, started)
		return fmt.Errorf("ip instance data format err: %s", err)
	}
	//update ip instance data
	if ipinst.MacAddr != "" {
		oldInst.MacAddr = ipinst.MacAddr
	}
	if ipinst.App != "" {
		oldInst.App = ipinst.App
	}
	oldInst.Update = time.Now().Format("2006-01-02 15:04:05")
	//format json data
	newData, _ := json.Marshal(oldInst)
	if err := srv.store.Add(instpath, newData); err != nil {
		blog.Errorf("UpdatAvailable ip instance %s failed, %v", instpath, err)
		reportMetrics("updateAvailableIPInstance", stateStorageFailure, started)
		return fmt.Errorf("update ip instance failed: %s", err.Error())
	}
	blog.Info("UpdateAvailable ip instance %s success.", instpath)
	reportMetrics("updateAvailableIPInstance", stateSuccess, started)
	return nil
}

//TransferIPAttribute change IP status
func (srv *NetService) TransferIPAttribute(tranInput *types.TranIPAttrInput) (int, error) {
	started := time.Now()
	//try to lock
	lockpath := filepath.Join(defaultLockerPath, tranInput.Cluster, tranInput.Net)
	poolLocker, lErr := srv.store.GetLocker(lockpath)
	if lErr != nil {
		blog.Errorf("create locker %s failed:%s", lockpath, lErr)
		reportMetrics("transferIPAttribute", stateLogicFailure, started)
		return types.ALL_IP_FAILED, fmt.Errorf("create locker %s err, %s", lockpath, lErr.Error())
	}
	defer poolLocker.Unlock()
	if err := poolLocker.Lock(); err != nil {
		blog.Errorf("try to lock pool %s/%s err, %s", tranInput.Cluster, tranInput.Net, err.Error())
		reportMetrics("transferIPAttribute", stateLogicFailure, started)
		return types.ALL_IP_FAILED, fmt.Errorf("lock pool %s err, %s", tranInput.Net, err.Error())
	}
	srcSubPath := filepath.Join(defaultPoolInfoPath, tranInput.Cluster, tranInput.Net, tranInput.SrcStatus)
	destSubPath := filepath.Join(defaultPoolInfoPath, tranInput.Cluster, tranInput.Net, tranInput.DestStatus)
	failedCode := types.ALL_IP_FAILED
	//var failedIPList []string
	for i, IP := range tranInput.IPList {
		if i != 0 {
			failedCode = types.SOME_IP_FAILED
		}
		srcIPPath := filepath.Join(srcSubPath, IP)
		destIPPath := filepath.Join(destSubPath, IP)
		//delete old node
		ipData, dErr := srv.store.Delete(srcIPPath)
		if dErr != nil {
			blog.Errorf("delete src ip %s failed:%s", srcIPPath, dErr.Error())
			failedIPList := make([]string, len(tranInput.IPList)-i)
			copy(failedIPList, tranInput.IPList[i:])
			reportMetrics("transferIPAttribute", stateStorageFailure, started)
			return failedCode, fmt.Errorf("delete src ip %s,failedIPList %v failed:%s", srcIPPath, failedIPList, dErr.Error())
		}
		ipInst := &types.IPInst{}
		if err := json.Unmarshal(ipData, ipInst); err != nil {
			blog.Errorf("Fatal error in TransferIPAttribute, %s data decode json err, %v", srcIPPath, err)
			failedIPList := make([]string, len(tranInput.IPList)-i)
			copy(failedIPList, tranInput.IPList[i:])
			reportMetrics("transferIPAttribute", stateJSONFailure, started)
			return failedCode, fmt.Errorf("ip %s data json decode err:%s,failedIPList %v", srcIPPath, err.Error(), failedIPList)
		}
		//change ip status record in storage
		ipInst.LastStatus = tranInput.SrcStatus
		ipInst.Status = tranInput.DestStatus
		ipInst.Update = time.Now().Format("2006-01-02 15:04:05")
		destIPData, jsonErr := json.Marshal(ipInst)
		if jsonErr != nil {
			blog.Errorf("Marshal %v failed:%s", ipInst, jsonErr.Error())
			failedIPList := make([]string, len(tranInput.IPList)-i)
			copy(failedIPList, tranInput.IPList[i:])
			reportMetrics("transferIPAttribute", stateJSONFailure, started)
			return failedCode, fmt.Errorf("Marshal %v failed:%s,failedIPList %s", ipInst, jsonErr.Error(), failedIPList)
		}
		//add dest node
		if err := srv.store.Add(destIPPath, destIPData); err != nil {
			blog.Errorf("move ip %s to %s failed:%s", IP, destIPPath, err.Error())
			failedIPList := make([]string, len(tranInput.IPList)-i)
			copy(failedIPList, tranInput.IPList[i:])
			reportMetrics("transferIPAttribute", stateLogicFailure, started)
			return failedCode, fmt.Errorf("move ip %s to %s failed:%s,failedIPList %s", IP, destIPPath, err.Error(), failedIPList)
		}
		blog.Infof("move IP %s from %s to %s success", IP, srcIPPath, destIPPath)
	}
	reportMetrics("transferIPAttribute", stateSuccess, started)
	return 0, nil
}
