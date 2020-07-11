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

package rdiscover

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/metric"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/health"

	"golang.org/x/net/context"
)

// Server define struct include all server information
type Server struct {
	metricServiceInfo []*types.MetricServiceInfo
	metricServiceLock sync.RWMutex

	apiInfo []*types.APIServInfo
	apiLock sync.RWMutex

	storageInfo []*types.BcsStorageInfo
	storageLock sync.RWMutex
}

// RDiscover register and discover
type RDiscover struct {
	ip         string
	port       uint
	metricPort uint
	isSSL      bool
	rd         *RegisterDiscover.RegDiscover
	cancel     context.CancelFunc
	server     *Server

	// elect concerns
	serviceLock       sync.RWMutex
	role              metric.RoleType
	roleEvent         chan metric.RoleType
	rawServerInfoData []byte
}

// NewRegDiscover create a RegDiscover object
func NewRDiscover(op *config.Config) *RDiscover {
	return &RDiscover{
		ip:         op.Address,
		port:       op.Port,
		metricPort: op.MetricPort,
		isSSL:      op.ServCert.IsSSL,
		rd:         RegisterDiscover.NewRegDiscoverEx(op.BCSZk, 10*time.Second),
		cancel:     nil,
		server:     &Server{},
	}
}

// Start the register and discover
func (r *RDiscover) Start() (roleEvent chan metric.RoleType, err error) {
	// create root context
	rootCtx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.role = metric.UnknownRole
	r.roleEvent = make(chan metric.RoleType, 10)

	// start regDiscover
	if err = r.rd.Start(); err != nil {
		blog.Error("fail to start register and discover server. err: %v", err)
		return
	}

	// register self
	if err = r.registerAndDiscover(); err != nil {
		blog.Error("fail to register metric service(%s). err: %v", r.ip, err)
		return
	}

	// discover other bcs service
	// self for elect
	metricServicePath := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_METRICSERVICE
	metricServiceEvent, err := r.rd.DiscoverService(metricServicePath)
	if err != nil {
		blog.Error("fail to register discover for metric service server. err: %v", err)
		return
	}

	// api
	apiPath := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_APISERVER
	apiEvent, err := r.rd.DiscoverService(apiPath)
	if err != nil {
		blog.Error("fail to register discover for api server. err: %v", err)
		return
	}

	// storage
	storagePath := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_STORAGE
	storageEvent, err := r.rd.DiscoverService(storagePath)
	if err != nil {
		blog.Error("fail to register discover for storage server. err: %v", err)
		return
	}

	go func() {
		for {
			select {
			case metricServiceEnv := <-metricServiceEvent:
				r.elect(metricServiceEnv.Server)
				r.discoverMetricServiceServer(metricServiceEnv.Server)
			case apiEnv := <-apiEvent:
				r.discoverApiServer(apiEnv.Server)
			case storageEnv := <-storageEvent:
				r.discoverStorageServer(storageEnv.Server)
			case <-rootCtx.Done():
				blog.Warn("register and discover done")
				return
			}
		}
	}()

	roleEvent = r.roleEvent
	return
}

// Stop the register and discover
func (r *RDiscover) Stop() error {
	r.cancel()
	r.rd.Stop()

	return nil
}

// GetApiServer fetch api server
func (r *RDiscover) GetApiServer() (string, error) {
	r.server.apiLock.RLock()
	defer r.server.apiLock.RUnlock()

	if len(r.server.apiInfo) <= 0 {
		return "", fmt.Errorf("there is no api server")
	}

	//rand
	rand.Seed(int64(time.Now().Nanosecond()))
	num := len(r.server.apiInfo)
	serverInfo := r.server.apiInfo[rand.Intn(num)]
	host := serverInfo.Scheme + "://" + serverInfo.IP + ":" + strconv.Itoa(int(serverInfo.Port))
	return host, nil
}

// GetStorageServer fetch storage server
func (r *RDiscover) GetStorageServer() (string, error) {
	r.server.storageLock.RLock()
	defer r.server.storageLock.RUnlock()

	if len(r.server.storageInfo) <= 0 {
		return "", fmt.Errorf("there is no storage server")
	}

	//rand
	rand.Seed(int64(time.Now().Nanosecond()))
	num := len(r.server.storageInfo)
	serverInfo := r.server.storageInfo[rand.Intn(num)]
	host := serverInfo.Scheme + "://" + serverInfo.IP + ":" + strconv.Itoa(int(serverInfo.Port))
	return host, nil
}

func (r *RDiscover) registerAndDiscover() error {
	metricServiceInfo := new(types.MetricServiceInfo)

	metricServiceInfo.IP = r.ip
	metricServiceInfo.Port = r.port
	metricServiceInfo.MetricPort = r.metricPort
	metricServiceInfo.Scheme = "http"
	if r.isSSL {
		metricServiceInfo.Scheme = "https"
	}
	metricServiceInfo.Version = version.GetVersion()
	metricServiceInfo.Pid = os.Getpid()

	data, err := json.Marshal(metricServiceInfo)
	if err != nil {
		blog.Errorf("fail to marshal metric service server info to json. err:%s", err.Error())
		return err
	}

	path := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_METRICSERVICE + "/" + r.ip

	r.rawServerInfoData = data
	return r.rd.RegisterAndWatchService(path, data)
}

func (r *RDiscover) discoverMetricServiceServer(serverInfo []string) {
	blog.Infof("discover metricService(%v)", serverInfo)

	metricServiceList := make([]*types.MetricServiceInfo, 0)

	for _, server := range serverInfo {
		metricService := new(types.MetricServiceInfo)
		if err := codec.DecJson([]byte(server), metricService); err != nil {
			blog.Warnf("fail to do json decode(%s), err: %v", server, err)
			continue
		}

		metricServiceList = append(metricServiceList, metricService)
	}

	r.server.metricServiceLock.Lock()
	defer r.server.metricServiceLock.Unlock()
	r.server.metricServiceInfo = metricServiceList
}

func (r *RDiscover) elect(serverInfo []string) {
	if len(serverInfo) == 0 {
		blog.Warnf("there is no metric service on zk")
		return
	}

	isMaster := false
	if serverInfo[0] == string(r.rawServerInfoData) {
		isMaster = true
	}

	blog.Infof("metric service check role change begin")
	r.serviceLock.Lock()
	defer func() {
		blog.Infof("metric service check role change end")
		r.serviceLock.Unlock()
	}()

	if isMaster && r.role == metric.MasterRole || !isMaster && r.role == metric.SlaveRole {
		blog.Infof("metric service role not changed: %s", r.role)
		return
	}

	if isMaster {
		blog.Infof("metric service change role: %s ---> %s", r.role, metric.MasterRole)
		r.role = metric.MasterRole
		r.roleEvent <- metric.MasterRole
		health.SetMaster()
		return
	}

	blog.Infof("metric service change role: %s ---> %s", r.role, metric.SlaveRole)
	r.role = metric.SlaveRole
	r.roleEvent <- metric.SlaveRole
	health.SetSlave()
}

func (r *RDiscover) discoverApiServer(serverInfo []string) {
	blog.Infof("discover api(%v)", serverInfo)

	apiList := make([]*types.APIServInfo, 0)

	for _, server := range serverInfo {
		api := new(types.APIServInfo)
		if err := codec.DecJson([]byte(server), api); err != nil {
			blog.Warnf("fail to do json decode(%s), err: %v", server, err)
			continue
		}

		apiList = append(apiList, api)
	}

	r.server.apiLock.Lock()
	defer r.server.apiLock.Unlock()
	r.server.apiInfo = apiList

	if len(apiList) == 0 {
		health.SetUnhealthy(health.ApiKey, "there is no api server")
	} else {
		health.SetHealth(health.ApiKey)
	}
}

func (r *RDiscover) discoverStorageServer(serverInfo []string) {
	blog.Infof("discover storage(%v)", serverInfo)

	storageList := make([]*types.BcsStorageInfo, 0)

	for _, server := range serverInfo {
		storage := new(types.BcsStorageInfo)
		if err := codec.DecJson([]byte(server), storage); err != nil {
			blog.Warnf("fail to do json decode(%s), err: %v", server, err)
			continue
		}

		storageList = append(storageList, storage)
	}

	r.server.storageLock.Lock()
	defer r.server.storageLock.Unlock()
	r.server.storageInfo = storageList

	if len(storageList) == 0 {
		health.SetUnhealthy(health.StorageKey, "there is no storage server")
	} else {
		health.SetHealth(health.StorageKey)
	}
}
