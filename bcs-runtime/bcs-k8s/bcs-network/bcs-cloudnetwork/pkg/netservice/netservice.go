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

package netservice

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	netsvc "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
)

// Interface interface for netservice
type Interface interface {
	Init() error
	CreateOrUpdatePool(pool *netsvc.NetPool) error
	UpdateIPInstance(ins *netsvc.IPInst) error
	DeletePool(clusterid, net string) error
}

// Client client for netservice
type Client struct {
	client    bcsapi.Netservice
	zookeeper string
	key       string
	cert      string
	ca        string
}

// New create a netservice client
func New(zookeeper, key, cert, ca string) *Client {
	return &Client{
		zookeeper: zookeeper,
		key:       key,
		cert:      cert,
		ca:        ca,
	}
}

// Init netservice client
func (c *Client) Init() error {
	if len(c.zookeeper) == 0 {
		return fmt.Errorf("netservice zookeeper config cannot be empty")
	}
	client := bcsapi.NewNetserviceCli()
	if len(c.key) != 0 || len(c.cert) != 0 || len(c.ca) != 0 {
		if err := client.SetCerts(c.ca, c.key, c.cert, static.ClientCertPwd); err != nil {
			return err
		}
	}

	//client get bcs-netservice info
	c.zookeeper = strings.Replace(c.zookeeper, ";", ",", -1)
	hosts := strings.Split(c.zookeeper, ",")

	if err := client.GetNetService(hosts); err != nil {
		return fmt.Errorf("get netservice failed, %s", err.Error())
	}

	c.client = client
	return nil
}

// CreateOrUpdatePool create or update pool
func (c *Client) CreateOrUpdatePool(pool *netsvc.NetPool) error {
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

// UpdateIPInstance update ip instance
func (c *Client) UpdateIPInstance(ins *netsvc.IPInst) error {
	err := c.client.UpdateIPInstance(ins)
	if err != nil {
		blog.Errorf("update ip instance with %v failed, err %s", ins, err.Error())
		return err
	}
	return nil
}

// DeletePool delete ip pool
func (c *Client) DeletePool(clusterid, net string) error {
	err := c.client.DeletePool(clusterid, net)
	if err != nil {
		blog.Errorf("net service client delete pool failed, err %s", err.Error())
		return fmt.Errorf("net service client delete pool failed, err %s", err.Error())
	}
	return nil
}
