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

package app

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-cc-agent/config"
)

var mesosCacheInfo *NodeInfo

func synchronizeMesosNodeInfo(config *config.BcsCcAgentConfig) error {
	hostIp := os.Getenv("BCS_NODE_IP")
	if hostIp == "" {
		return fmt.Errorf("env [BCS_NODE_IP] is empty")
	}

	// init info cache
	mesosCacheInfo = &NodeInfo{}

	// sync info from bk-cmdb periodically
	go func() {
		ticker := time.NewTicker(time.Duration(1) * time.Minute)
		defer ticker.Stop()
		for {
			blog.Info("starting to synchronize node info...")

			nodeProperties, err := getInfoFromBkCmdb(config, hostIp)
			if err != nil {
				blog.Errorf("error synchronizing node info: %s", err.Error())
				continue
			}

			// currentNodeInfo represents the current node info
			currentNodeInfo := &NodeInfo{
				Properties: nodeProperties,
			}

			// if nodeInfo updated, then update to file and node label
			if !reflect.DeepEqual(*mesosCacheInfo, *currentNodeInfo) {
				mesosCacheInfo = currentNodeInfo
				err := updateMesosNodeInfo(k8sCacheInfo)
				if err != nil {
					blog.Errorf("error updating node info to file and node label: %s", err.Error())
					continue
				}
			}

			select {
			case <-ticker.C:
			}
		}
	}()

	return nil
}
