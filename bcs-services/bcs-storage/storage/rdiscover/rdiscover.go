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
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/app/options"

	"golang.org/x/net/context"
)

// RegDiscover register and discover
type RegDiscover struct {
	ip           string
	port         uint
	externalIp   string
	externalPort uint
	metricPort   uint
	isSSL        bool
	rd           *RegisterDiscover.RegDiscover
	rootCtx      context.Context
	cancel       context.CancelFunc
}

// NewRegDiscover create a RegDiscover object
func NewRegDiscover(conf *options.StorageOptions) *RegDiscover {
	return &RegDiscover{
		ip:           conf.Address,
		port:         conf.Port,
		externalIp:   conf.ExternalIp,
		externalPort: conf.ExternalPort,
		isSSL:        conf.ServerCert.IsSSL,
		metricPort:   conf.MetricPort,
		rd:           RegisterDiscover.NewRegDiscoverEx(conf.BCSZk, 10*time.Second),
	}
}

// Start the register and discover
func (r *RegDiscover) Start() error {
	//create root context
	r.rootCtx, r.cancel = context.WithCancel(context.Background())

	//start regdiscover
	if err := r.rd.Start(); err != nil {
		blog.Error("fail to start register and discover serv. err:%s", err.Error())
		return err
	}

	// register storage
	if err := r.registerStorage(); err != nil {
		blog.Error("fail to register storage(%s), err:%s", r.ip, err.Error())
		return err
	}

	//here: discover other bcs services

	for {
		select {
		case <-r.rootCtx.Done():
			blog.Warn("register and discover serv done")
			return nil
		}
	}
}

// Stop the register and discover
func (r *RegDiscover) Stop() error {
	r.cancel()

	r.rd.Stop()

	return nil
}

func (r *RegDiscover) registerStorage() error {
	storageServerInfo := new(types.BcsStorageInfo)

	storageServerInfo.IP = r.ip
	storageServerInfo.Port = r.port
	storageServerInfo.ExternalIp = r.externalIp
	storageServerInfo.ExternalPort = r.externalPort
	storageServerInfo.Scheme = "http"
	storageServerInfo.MetricPort = r.metricPort
	if r.isSSL {
		storageServerInfo.Scheme = "https"
	}

	storageServerInfo.Version = version.GetVersion()
	storageServerInfo.Pid = os.Getpid()

	data, err := json.Marshal(storageServerInfo)
	if err != nil {
		blog.Error("fail to marshal storage server info to json. err:%s", err.Error())
		return err
	}

	path := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_STORAGE + "/" + r.ip

	return r.rd.RegisterAndWatchService(path, data)
}
