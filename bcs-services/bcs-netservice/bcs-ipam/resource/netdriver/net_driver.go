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
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/bcs-ipam/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/bcs-ipam/resource"
)

var (
	defaultConfig = "/data/bcs/bcs-cni/bin/conf/bcs.conf"
)

// NewDriver create IPDriver for bcs-netservice
func NewDriver() (resource.IPDriver, error) {
	// check config for zookeeper list
	config, err := conf.LoadConfigFromFile(defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("load config bcs.conf failed, %s", err.Error())
	}
	client := bcsapi.NewNetserviceCli()
	if config.TLS != nil {
		config.TLS.Passwd = static.ClientCertPwd
		if err = client.SetCerts(config.TLS.CACert, config.TLS.Key, config.TLS.PubKey, config.TLS.Passwd); err != nil {
			return nil, err
		}
	}
	driver := &NetDriver{
		netClient: client,
	}

	if config.EtcdHost != "" {
		var etcdTls *tls.Config
		if config.EtcdCA != "" && config.EtcdCert != "" && config.EtcdKey != "" {
			etcdTls, err = ssl.ClientTslConfVerity(config.EtcdCA, config.EtcdCert, config.EtcdKey,
				static.ServerCertPwd)
			if err != nil {
				return nil, errors.Wrapf(err, "load etcd tls config failed")
			}
		}
		if err := client.GetNetServiceWithEtcd(strings.Split(config.EtcdHost, ","), etcdTls); err != nil {
			return nil, fmt.Errorf("get netservice failed, %s", err.Error())
		}
	} else {
		// client get bcs-netservice info
		config.ZkHost = strings.Replace(config.ZkHost, ",", ";", -1)
		hosts := strings.Split(config.ZkHost, ";")
		if err := client.GetNetService(hosts); err != nil {
			return nil, fmt.Errorf("get netservice failed, %s", err.Error())
		}
	}
	return driver, nil
}

// NetDriver driver for bcs-netservice
type NetDriver struct {
	netClient bcsapi.Netservice
}

// GetIPAddr get available ip resource for contaienr
func (driver *NetDriver) GetIPAddr(host, containerID, requestIP string) (*types.IPInfo, error) {
	// construct types.IPLean
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

// ReleaseIPAddr release ip address for container
func (driver *NetDriver) ReleaseIPAddr(host string, containerID string, ipInfo *types.IPInfo) error {
	if host == "" || containerID == "" {
		return fmt.Errorf("host/container info lost")
	}
	hostInfo, err := driver.netClient.GetHostInfo(host, 3)
	if err != nil {
		return fmt.Errorf("get host info for host %s failed, err %s", host, err.Error())
	}
	found := false
	for cID := range hostInfo.Containers {
		if cID == containerID {
			found = true
			break
		}
	}
	if !found {
		return nil
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

// GetHostInfo Get host info from driver
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
