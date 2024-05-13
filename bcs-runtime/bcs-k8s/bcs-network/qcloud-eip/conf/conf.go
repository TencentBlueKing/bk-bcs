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

package conf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"

	"github.com/containernetworking/cni/pkg/types"
)

const (
	// DefaultRouteRulePriority default priority of route rule
	DefaultRouteRulePriority = 2048
)

// CNIArgs args from Env CNI_ARGS
type CNIArgs struct {
	//for unknown config item
	types.CommonArgs
	IP      net.IP `json:"ip,omitempty"`      //IP address if designated
	Gateway net.IP `json:"gateway,omitempty"` //gateway if designated
}

// NetArgs cni net args
type NetArgs struct {
	Zookeeper string `json:"zookeeper"`
	Key       string `json:"key"`
	PubKey    string `json:"pubkey"`
	Ca        string `json:"cacert"`
}

// NetConf net config
type NetConf struct {
	types.NetConf
	TencentCloudCVMDomain string   `json:"tencentcloudCVMDomain"`
	TencentCloudVPCDomain string   `json:"tencentcloudVPCDomain"`
	Master                string   `json:"master"`
	ENIPrefix             string   `json:"eniPrefix"`
	ClusterID             string   `json:"clusterId"`
	Region                string   `json:"region"`
	Secret                string   `json:"secret"`
	UUID                  string   `json:"uuid"`
	SubnetID              string   `json:"subnetId,omitempty"`
	MTU                   int      `json:"mtu,omitempty"`
	RouteRulePriority     int      `json:"routeRulePriority,omitempty"`
	NetService            *NetArgs `json:"netservice,omitempty"`
	Args                  *CNIArgs
}

// LoadConf load config
func LoadConf(bytes []byte, args string) (*NetConf, string, error) {
	n := &NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}
	if n.Master == "" {
		return nil, "", fmt.Errorf(`"master" filed is required`)
	}
	if n.Secret == "" {
		return nil, "", fmt.Errorf(`"secret" is required`)
	}
	if n.Region == "" {
		return nil, "", fmt.Errorf("Lost region information")
	}
	if n.UUID == "" {
		return nil, "", fmt.Errorf("Lost Encrypted UUID")
	}
	if n.RouteRulePriority < 256 && n.RouteRulePriority != 0 {
		return nil, "", fmt.Errorf("invalid route rule priority %d", n.RouteRulePriority)
	}
	if n.RouteRulePriority == 0 {
		n.RouteRulePriority = DefaultRouteRulePriority
	}
	if args != "" {
		n.Args = &CNIArgs{}
		err := types.LoadArgs(args, n.Args)
		if err != nil {
			return nil, "", err
		}
	}
	return n, n.CNIVersion, nil
}

// LoadConfFromFile load config from file
func LoadConfFromFile(config string) (*NetConf, error) {
	//open file
	f, err := os.Open(config)
	if err != nil {
		blog.Errorf("Open configuration file %s failed, %s", config, err)
		return nil, err
	}
	blog.Infof("Open Configuration file %s success.", config)
	//reading contents
	allBytes, err := ioutil.ReadAll(f)
	if err != nil {
		blog.Errorf("Reading Configuration content failed, %s", err)
		return nil, err
	}
	blog.Infof("Reading Configuration json content success.")
	conf, _, err := LoadConf(allBytes, "")
	if err != nil {
		blog.Infof("Parsing Configuration %s failed, %s", config, err)
		return nil, err
	}
	blog.Infof("Parse json Configuration success")
	//decode encrypted uuid
	key, err := encrypt.DesDecryptFromBase([]byte(conf.UUID))
	if err != nil {
		blog.Infof("Loading UUDI failed, %s", err)
		return nil, err
	}
	conf.UUID = string(key)
	return conf, nil
}
