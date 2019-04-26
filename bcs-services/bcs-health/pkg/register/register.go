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

package register

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	regd "bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/codec"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"
)

func Register(localIP string, scheme string, port uint, metricPort uint, zkAddr string, clusterid string) error {
	disc := regd.NewRegDiscoverEx(zkAddr, time.Duration(5*time.Second))
	if err := disc.Start(); nil != err {
		return fmt.Errorf("start discover service failed. Error:%v", err)
	}

	// watch storage and refresh
	storagePath := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_STORAGE
	storageEvent, err := disc.DiscoverService(storagePath)
	if err != nil {
		blog.Error("fail to register discover for storage server. err: %v", err)
		return err
	}
	go func() {
		for {
			select {
			case storageEnv := <-storageEvent:
				discoverStorageServer(storageEnv.Server)
			}
		}
	}()

	// register itself
	hostname, err := os.Hostname()
	if nil != err {
		blog.Errorf("get hostname failed. err: %v", err)
	}

	var id string
	if clusterid == "master" {
		id = ""
	} else {
		id = clusterid
	}

	info := types.BcsHealthInfo{
		ServerInfo: types.ServerInfo{
			IP:         localIP,
			Port:       port,
			MetricPort: metricPort,
			HostName:   hostname,
			Scheme:     scheme,
			Version:    version.GetVersion(),
			Cluster:    id,
			Pid:        os.Getpid(),
		},
	}

	js, err := json.Marshal(info)
	if nil != err {
		return fmt.Errorf("marshal health info failed. err: %v", err)
	}

	path := fmt.Sprintf("%s/%s/%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_HEALTH, clusterid, localIP)
	if err := disc.RegisterAndWatchService(path, js); nil != err {
		return fmt.Errorf("register service failed. Error:%v", err)
	}
	return nil
}

type RequestInfo struct {
	Module string `json:"module"`
	IP     string `json:"ip"`
}

type RespInfo struct {
	RequestInfo
	ClusterID string `json:"clusterid"`
	Extension string `json:"extendedinfo"`
}

type Response struct {
	ErrorCode int      `json:"errcode"`
	ErrorMsg  string   `json:"errmsg"`
	Resp      RespInfo `json:"data"`
}

var (
	storageLock sync.RWMutex
	storageInfo []types.BcsStorageInfo
)

func GetStorageServer() (string, error) {
	storageLock.RLock()
	defer storageLock.RUnlock()

	if len(storageInfo) <= 0 {
		return "", fmt.Errorf("there is no storage server")
	}

	//rand
	rand.Seed(int64(time.Now().Nanosecond()))
	num := len(storageInfo)
	serverInfo := storageInfo[rand.Intn(num)]
	host := serverInfo.Scheme + "://" + serverInfo.IP + ":" + strconv.Itoa(int(serverInfo.Port))
	return host, nil
}

func discoverStorageServer(serverInfo []string) {
	blog.Infof("discover storage(%v)", serverInfo)

	storageList := make([]types.BcsStorageInfo, 0)

	for _, server := range serverInfo {
		storage := new(types.BcsStorageInfo)
		if err := codec.DecJson([]byte(server), storage); err != nil {
			blog.Warnf("fail to do json decode(%s), err: %v", server, err)
			continue
		}

		storageList = append(storageList, *storage)
	}

	storageLock.Lock()
	defer storageLock.Unlock()
	storageInfo = storageList
}
