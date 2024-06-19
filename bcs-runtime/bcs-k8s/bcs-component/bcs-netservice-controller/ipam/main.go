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

// Package main xxx
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/jackpal/gateway"

	ipamtypes "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/ipam/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/pkg/client"
)

var (
	// default directory for log output
	defaultLogDir = "./logs"
)

// BcsIPAMConfig represents the IP related network configuration.
type BcsIPAMConfig struct {
	Routes             []types.Route      `json:"routes"`
	LogDir             string             `json:"logDir,omitempty"`
	Host               string             `json:"host"`
	NetserviceEndpoint string             `json:"netserviceEndpoint"`
	Args               *ipamtypes.K8sArgs `json:"-"`
}

// NetConf network configuration from json config, reading from stdin
type NetConf struct {
	Name       string         `json:"name"`
	CNIVersion string         `json:"cniVersion"`
	IPAM       *BcsIPAMConfig `json:"ipam"`
}

// LoadBcsIPAMConfig creates a NetworkConfig from the given network name.
func LoadBcsIPAMConfig(bytes []byte, args string) (*NetConf, *BcsIPAMConfig, error) {
	n := &NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, nil, err
	}
	// check ipam item from stdin json file
	if n.IPAM == nil {
		return nil, nil, fmt.Errorf("IPAM config missing 'ipam' key")
	}
	if len(n.IPAM.LogDir) == 0 {
		blog.Errorf("log dir is empty, use default log dir './logs'")
		n.IPAM.LogDir = defaultLogDir
	}
	if args != "" {
		n.IPAM.Args = &ipamtypes.K8sArgs{}
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
	// loading config from stdin
	netConf, ipamConf, err := LoadBcsIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	// init inner log tool
	// ! pay more attention, CNI command line can not output log
	// ! to stderr or stdout according to cni specification
	blog.InitLogs(conf.LogConfig{
		LogDir: ipamConf.LogDir,
		// never log to stderr
		StdErrThreshold: "6",
		LogMaxSize:      20,
		LogMaxNum:       100,
	})
	defer blog.CloseLogs()

	k8sConf, err := ipamtypes.LoadK8sArgs(args)
	if err != nil {
		blog.Errorf("failed to LoadK8sArgs, err %s", err.Error())
		return fmt.Errorf("failed to LoadK8sArgs, err %s", err.Error())
	}

	nsClient, err := client.New(ipamConf.NetserviceEndpoint)
	if err != nil {
		return err
	}

	// get host default gateway
	gw, gwerr := gateway.DiscoverGateway()
	if gwerr != nil {
		return gwerr
	}

	req := &client.AllocateReq{
		ContainerID:  args.ContainerID,
		PodName:      string(k8sConf.K8S_POD_NAME),
		PodNamespace: string(k8sConf.K8S_POD_NAMESPACE),
		Host:         ipamConf.Host,
		HostGateway:  gw.String(),
	}
	resp, getErr := nsClient.Allocate(req)
	if getErr != nil {
		return getErr
	}
	if resp.Code != 0 || !resp.Result {
		blog.Errorf("allocate ip for %s/%s failed, resp %v", req.PodName, req.PodNamespace, resp)
		return fmt.Errorf("allocate ip for %s/%s failed, resp %v", req.PodName, req.PodNamespace, resp)
	}

	// doto: to deal with mac address

	// create results
	result := &current.Result{}
	// ip info
	ip, ipAddr, _ := net.ParseCIDR(resp.Data.IPAddr + "/" + strconv.Itoa(resp.Data.Mask))
	iface := 0
	ipConf := &current.IPConfig{
		Version:   "4",
		Interface: &iface,
		Address:   net.IPNet{IP: ip, Mask: ipAddr.Mask},
		Gateway:   net.ParseIP(resp.Data.Gateway),
	}
	result.IPs = []*current.IPConfig{ipConf}
	// route info, if no gateway info ,use ipinfo.Gateway
	for _, configRoute := range ipamConf.Routes {
		if configRoute.GW == nil {
			route := &types.Route{
				Dst: configRoute.Dst,
				GW:  net.ParseIP(resp.Data.Gateway),
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
	// loading config from stdin
	_, ipamConf, err := LoadBcsIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}

	// init inner log tool
	// ! pay more attention, CNI command line can not output log
	// ! to stderr or stdout according to cni specification
	blog.InitLogs(conf.LogConfig{
		LogDir: ipamConf.LogDir,
		// never log to stderr
		StdErrThreshold: "6",
		LogMaxSize:      20,
		LogMaxNum:       100,
	})
	defer blog.CloseLogs()

	k8sConf, err := ipamtypes.LoadK8sArgs(args)
	if err != nil {
		blog.Errorf("failed to LoadK8sArgs, err %s", err.Error())
		return fmt.Errorf("failed to LoadK8sArgs, err %s", err.Error())
	}

	nsClient, err := client.New(ipamConf.NetserviceEndpoint)
	if err != nil {
		return err
	}

	// release ip with IPInfo
	resp, rErr := nsClient.Release(&client.ReleaseReq{
		ContainerID:  args.ContainerID,
		PodName:      string(k8sConf.K8S_POD_NAME),
		PodNamespace: string(k8sConf.K8S_POD_NAMESPACE),
		Host:         ipamConf.Host,
	})
	if rErr != nil {
		blog.Errorf("release failed, err %s", rErr.Error())
		return fmt.Errorf("release failed, err %s", rErr.Error())
	}
	if resp.Code != 0 || !resp.Result {
		blog.Errorf("release failed, resp %v", resp)
		return fmt.Errorf("release failed, resp %v", resp)
	}
	blog.Infof("release ip successfully, resp %v", resp)
	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.PluginSupports("0.3.0"))
}
