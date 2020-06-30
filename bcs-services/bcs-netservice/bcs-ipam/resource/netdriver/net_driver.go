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

package netdriver

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/bcs-ipam/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/bcs-ipam/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/pkg/netservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/pkg/netservice/types"
	"strings"
)

var (
	defaultConfig = "/data/bcs/bcs-cni/bin/conf/bcs.conf"
)

// NewDriver create IPDriver for bcs-netservice
func NewDriver() (resource.IPDriver, error) {
	// check config for zookeeper list
	conf, err := conf.LoadConfigFromFile(defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("load config bcs.conf failed, %s", err.Error())
	}
	var client netservice.Client
	var clientErr error
	if conf.TLS == nil {
		client, clientErr = netservice.NewClient()
	} else {
		conf.TLS.Passwd = static.ClientCertPwd
		client, clientErr = netservice.NewTLSClient(conf.TLS.CACert, conf.TLS.Key, conf.TLS.PubKey, conf.TLS.Passwd)
	}
	if clientErr != nil {
		return nil, clientErr
	}
	driver := &NetDriver{
		netClient: client,
	}
	//client get bcs-netservice info
	conf.ZkHost = strings.Replace(conf.ZkHost, ",", ";", -1)
	hosts := strings.Split(conf.ZkHost, ";")
	if err := client.GetNetService(hosts); err != nil {
		return nil, fmt.Errorf("get netservice failed, %s", err.Error())
	}
	return driver, nil
}

//NetDriver driver for bcs-netservice
type NetDriver struct {
	netClient netservice.Client
}

//GetIPAddr get available ip resource for contaienr
func (driver *NetDriver) GetIPAddr(host, containerID, requestIP string) (*types.IPInfo, error) {
	//construct types.IPLean
	if host == "" || containerID == "" {
		return nil, fmt.Errorf("host/container info lost")
	}
	lease := &types.IPLease{
		Host:      host,
		Container: containerID,
		IPAddr:    requestIP,
	}
	ipInfo, err := driver.netClient.LeaseIPAddr(lease, 3)
	if err != nil {
		return nil, fmt.Errorf("lease ipaddr from bcs-netservice failed, %s", err.Error())
	}
	return ipInfo, nil
}

//ReleaseIPAddr release ip address for container
func (driver *NetDriver) ReleaseIPAddr(host string, containerID string, ipInfo *types.IPInfo) error {
	if host == "" || containerID == "" {
		return fmt.Errorf("host/container info lost")
	}
	release := &types.IPRelease{
		Host:      host,
		Container: containerID,
	}
	if err := driver.netClient.ReleaseIPAddr(release, ipInfo, 3); err != nil {
		return fmt.Errorf("release ipaddr from bcs-netservice failed, %s", err.Error())
	}
	return nil
}

//GetHostInfo Get host info from driver
func (driver *NetDriver) GetHostInfo(host string) (*types.HostInfo, error) {
	if len(host) == 0 {
		return nil, fmt.Errorf("host ip address lost")
	}
	hostInfo, err := driver.netClient.GetHostInfo(host, 3)
	if err != nil {
		return nil, err
	}
	return hostInfo, nil
}
