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

package eip

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	netsvc "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/qcloud-eip/conf"
)

// NetSvcClient client to operate netservice
type NetSvcClient struct {
	client bcsapi.Netservice
}

// NewNetSvcClient create client for
func NewNetSvcClient(conf *conf.NetArgs) (*NetSvcClient, error) {
	if conf == nil {
		return nil, fmt.Errorf("netservice config cannot be empty")
	}
	if len(conf.Zookeeper) == 0 {
		return nil, fmt.Errorf("netservice zookeeper config cannot be empty")
	}
	client := bcsapi.NewNetserviceCliWithTimeout(10)
	if len(conf.PubKey) != 0 || len(conf.Key) != 0 || len(conf.Ca) != 0 {
		if err := client.SetCerts(conf.Ca, conf.Key, conf.PubKey, static.ClientCertPwd); err != nil {
			return nil, err
		}
	}
	//client get bcs-netservice info
	conf.Zookeeper = strings.Replace(conf.Zookeeper, ",", ";", -1)
	hosts := strings.Split(conf.Zookeeper, ";")
	if err := client.GetNetService(hosts); err != nil {
		return nil, fmt.Errorf("get netservice failed, %s", err.Error())
	}
	return &NetSvcClient{
		client: client,
	}, nil
}

// CreateOrUpdatePool create or update net pool
func (c *NetSvcClient) CreateOrUpdatePool(pool *netsvc.NetPool) error {
	p, err := c.client.GetPool(pool.Cluster, pool.Net)
	if err != nil {
		if strings.Contains(err.Error(), "zk: node does not exist") {
			//no pool before, create new one
			return c.client.RegisterPool(pool)
		}
		return err
	}
	if p == nil {
		//no pool before, create new one
		return c.client.RegisterPool(pool)
	}
	return c.client.UpdatePool(pool)
}

// GetPool get net pool
func (c *NetSvcClient) GetPool(pool *netsvc.NetPool) (*netsvc.NetPool, error) {
	p, err := c.client.GetPool(pool.Cluster, pool.Net)
	if err != nil {
		return nil, err
	}
	return p[0], nil
}

// UpdateIPInstance update ip instance
func (c *NetSvcClient) UpdateIPInstance(ins *netsvc.IPInst) error {
	err := c.client.UpdateIPInstance(ins)
	if err != nil {
		blog.Errorf("update ip instance with %v failed, err %s", ins, err.Error())
		return err
	}
	return nil
}

// DeletePool delete ip pool
func (c *NetSvcClient) DeletePool(clusterid, net string) error {
	err := c.client.DeletePool(clusterid, net)
	if err != nil {
		blog.Errorf("net service client delete pool failed, err %s", err.Error())
		return fmt.Errorf("net service client delete pool failed, err %s", err.Error())
	}
	return nil
}
