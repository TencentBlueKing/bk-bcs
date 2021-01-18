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

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	nettypes "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/bcs-ipam/manager"
	bcsconf "github.com/Tencent/bk-bcs/bcs-services/bcs-netservice/config"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/spf13/pflag"
)

var (
	checkMode = false
)

//BcsIPAMConfig represents the IP related network configuration.
type BcsIPAMConfig struct {
	Name   string
	Type   string           `json:"type"` //bcs-ipam
	Host   string           `json:"host"` //local host ip
	Routes []types.Route    `json:"routes"`
	Args   *bcsconf.CNIArgs `json:"-"`
}

//NetConf network configuration from json config, reading from stdin
type NetConf struct {
	Name       string         `json:"name"`
	CNIVersion string         `json:"cniVersion"`
	IPAM       *BcsIPAMConfig `json:"ipam"`
}

//LoadBcsIPAMConfig creates a NetworkConfig from the given network name.
func LoadBcsIPAMConfig(bytes []byte, args string) (*BcsIPAMConfig, string, error) {
	n := NetConf{}
	if err := json.Unmarshal(bytes, &n); err != nil {
		return nil, "", err
	}
	//check ipam item from stdin json file
	if n.IPAM == nil {
		return nil, "", fmt.Errorf("IPAM config missing 'ipam' key")
	}
	if args != "" {
		n.IPAM.Args = &bcsconf.CNIArgs{}
		err := types.LoadArgs(args, n.IPAM.Args)
		if err != nil {
			return nil, "", err
		}
	}
	// Copy net name into IPAM so not to drag Net struct around
	n.IPAM.Name = n.Name
	return n.IPAM, n.CNIVersion, nil
}

func init() {
	pflag.CommandLine.BoolVar(&checkMode, "check-mode", checkMode, "check mode, for releasing dirty ip data in storage.")
	util.InitFlags()
}

//cmdAdd handle add command, reply available ip info
func cmdAdd(args *skel.CmdArgs) error {
	//loading config from stdin
	ipamConf, cniVersion, err := LoadBcsIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	//check host ip if configuration designated
	if len(ipamConf.Host) == 0 {
		//get one available host ip from system
		ipamConf.Host = util.GetIPAddress()
	}
	if len(ipamConf.Host) == 0 {
		return fmt.Errorf("Get no available host ip for request")
	}
	//Get available ip address from resource
	ipDriver, driverErr := manager.GetIPDriver()
	if driverErr != nil {
		return driverErr
	}
	requestIP := ""
	if ipamConf.Args != nil && ipamConf.Args.IP != nil {
		requestIP = ipamConf.Args.IP.String()
	}
	info, getErr := ipDriver.GetIPAddr(ipamConf.Host, args.ContainerID, requestIP)
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

	return types.PrintResult(result, cniVersion)
}

//cmdDel release ip address
func cmdDel(args *skel.CmdArgs) error {
	//loading config from stdin
	ipamConf, _, err := LoadBcsIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	//check host ip if configuration designated
	if len(ipamConf.Host) == 0 {
		//get one available host ip from system
		ipamConf.Host = util.GetIPAddress()
	}
	if len(ipamConf.Host) == 0 {
		return fmt.Errorf("Get no available host ip for request")
	}
	ipDriver, driverErr := manager.GetIPDriver()
	if driverErr != nil {
		return driverErr
	}
	ipInfo := &nettypes.IPInfo{}
	if ipamConf.Args != nil && ipamConf.Args.IP != nil {
		ipInfo.IPAddr = ipamConf.Args.IP.String()
	}
	//release ip with IPInfo
	return ipDriver.ReleaseIPAddr(ipamConf.Host, args.ContainerID, ipInfo)
}

func main() {
	blog.InitLogs(conf.LogConfig{ToStdErr: true})
	defer blog.CloseLogs()
	if checkMode {
		manager.DirtyCheck()
		return
	}
	skel.PluginMain(cmdAdd, cmdDel, version.PluginSupports("0.3.0"))
}
