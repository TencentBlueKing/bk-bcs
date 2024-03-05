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

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/cmd/bcs-eni-ipam/cloudagent"
)

var (
	//default directory for log output
	defaultLogDir = "./logs"
)

// BcsIPAMConfig represents the IP related network configuration.
type BcsIPAMConfig struct {
	Routes []types.Route       `json:"routes"`
	Args   *cloudagent.K8sArgs `json:"-"`
}

// NetConf network configuration from json config, reading from stdin
type NetConf struct {
	Name               string         `json:"name"`
	CNIVersion         string         `json:"cniVersion"`
	LogDir             string         `json:"logDir,omitempty"`
	CloudAgentEndpoint string         `json:"cloudAgentEndpoint"`
	IPAM               *BcsIPAMConfig `json:"ipam"`
}

// LoadBcsIPAMConfig creates a NetworkConfig from the given network name.
func LoadBcsIPAMConfig(bytes []byte, args string) (*NetConf, *BcsIPAMConfig, error) {
	n := &NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, nil, err
	}
	if len(n.LogDir) == 0 {
		blog.Errorf("log dir is empty, use default log dir './logs'")
		n.LogDir = defaultLogDir
	}
	//check ipam item from stdin json file
	if n.IPAM == nil {
		return nil, nil, fmt.Errorf("IPAM config missing 'ipam' key")
	}
	if args != "" {
		n.IPAM.Args = &cloudagent.K8sArgs{}
		err := types.LoadArgs(args, n.IPAM.Args)
		if err != nil {
			return nil, nil, err
		}
	}
	return n, n.IPAM, nil
}

func init() {
	util.InitFlags()
}

// cmdAdd handle add command, reply available ip info
func cmdAdd(args *skel.CmdArgs) error {
	//loading config from stdin
	netConf, ipamConf, err := LoadBcsIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	// init inner log tool
	// ! pay more attention, CNI command line can not output log
	// ! to stderr or stdout according to cni specification
	blog.InitLogs(conf.LogConfig{
		LogDir: netConf.LogDir,
		// never log to stderr
		StdErrThreshold: "6",
		LogMaxSize:      20,
		LogMaxNum:       100,
	})
	defer blog.CloseLogs()

	agentClient, err := cloudagent.NewClient(netConf.CloudAgentEndpoint)
	if err != nil {
		return err
	}

	info, getErr := agentClient.Alloc(args)
	if getErr != nil {
		return getErr
	}

	//create results
	result := &current.Result{}
	//mac address
	if info.MacAddr != "" {
		netInterface := &current.Interface{}
		netInterface.Name = args.IfName
		netInterface.Mac = info.MacAddr
		result.Interfaces = []*current.Interface{netInterface}
	}
	//ip info
	ip, ipAddr, _ := net.ParseCIDR(info.IPAddr + "/" + strconv.Itoa(info.Mask))
	iface := 0
	ipConf := &current.IPConfig{
		Version:   "4",
		Interface: &iface,
		Address:   net.IPNet{IP: ip, Mask: ipAddr.Mask},
		Gateway:   net.ParseIP(info.Gateway),
	}
	result.IPs = []*current.IPConfig{ipConf}
	//route info, if no gateway info ,use ipinfo.Gateway
	for _, configRoute := range ipamConf.Routes {
		if configRoute.GW == nil {
			route := &types.Route{
				Dst: configRoute.Dst,
				GW:  net.ParseIP(info.Gateway),
			}
			result.Routes = append(result.Routes, route)
		} else {
			result.Routes = append(result.Routes, &configRoute)
		}
	}

	return types.PrintResult(result, netConf.CNIVersion)
}

// cmdDel release ip address
func cmdDel(args *skel.CmdArgs) error {
	//loading config from stdin
	netConf, _, err := LoadBcsIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	// init inner log tool
	// ! pay more attention, CNI command line can not output log
	// ! to stderr or stdout according to cni specification
	blog.InitLogs(conf.LogConfig{
		LogDir: netConf.LogDir,
		// never log to stderr
		StdErrThreshold: "6",
		LogMaxSize:      20,
		LogMaxNum:       100,
	})
	defer blog.CloseLogs()

	agentClient, err := cloudagent.NewClient(netConf.CloudAgentEndpoint)
	if err != nil {
		return err
	}

	//release ip with IPInfo
	return agentClient.Release(args)
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.PluginSupports("0.3.0"))
}
