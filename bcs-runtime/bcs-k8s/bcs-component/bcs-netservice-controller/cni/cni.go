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
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	cniSpecVersion "github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/cni/logging"
	cnitypes "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/cni/types"
)

// NetConf cni config
type NetConf struct {
	types.NetConf

	LogFile     string `json:"logFile"`
	LogLevel    string `json:"logLevel"`
	LogToStderr bool   `json:"logToStderr,omitempty"`
}

func init() {
	// This is to ensure that all the namespace operations are performed for
	// a single thread
	runtime.LockOSThread()
}

const (
	ethernetMTU = 1500
)

// nolint
func loadConf(args *skel.CmdArgs) (*NetConf, *cnitypes.K8SArgs, error) {
	conf := NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return nil, nil, errors.Wrap(err, "failed to loading config from args")
	}

	k8sArgs := cnitypes.K8SArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		return nil, nil, errors.Wrap(err, "failed to load k8s config from args")
	}

	// Logging
	logging.SetLogStderr(conf.LogToStderr)
	if conf.LogFile != "" {
		logging.SetLogFile(conf.LogFile)
	}
	if conf.LogLevel != "" {
		logging.SetLogLevel(conf.LogLevel)
	}
	return &conf, &k8sArgs, nil
}

func setupVeth(netns ns.NetNS, ifName string) (*current.Interface, *current.Interface, error) {
	contIface := &current.Interface{}
	hostIface := &current.Interface{}

	err := netns.Do(func(hostNS ns.NetNS) error {
		// create the veth pair in the container and move host end into host netns
		hostVeth, containerVeth, err := ip.SetupVeth(ifName, ethernetMTU, hostNS)
		if err != nil {
			return err
		}
		contIface.Name = containerVeth.Name
		contIface.Mac = containerVeth.HardwareAddr.String()
		contIface.Sandbox = netns.Path()
		hostIface.Name = hostVeth.Name
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// need to lookup hostVeth again as its index has changed during ns move
	hostVeth, err := netlink.LinkByName(hostIface.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup %q: %v", hostIface.Name, err)
	}
	hostIface.Mac = hostVeth.Attrs().HardwareAddr.String()

	return hostIface, contIface, nil
}

func configIface(netns ns.NetNS, hostIface *current.Interface, contIface *current.Interface, ipv4Addr net.IP) error {
	hostVeth, err := netlink.LinkByName(hostIface.Name)
	if err != nil {
		return errors.Wrapf(err, "failed get link %s", hostIface.Name)
	}
	if err = netlink.LinkSetUp(hostVeth); err != nil {
		return errors.Wrapf(err, "failed to set link %q up", hostIface.Name)
	}
	// Add host route
	addrHostAddr := &net.IPNet{
		IP:   ipv4Addr,
		Mask: net.CIDRMask(32, 32)}
	if err = netlink.RouteReplace(&netlink.Route{
		LinkIndex: hostVeth.Attrs().Index,
		Scope:     netlink.SCOPE_LINK,
		Dst:       addrHostAddr}); err != nil {
		return errors.Wrap(err, "failed to add host route")
	}

	return netns.Do(func(_ ns.NetNS) error {
		contVeth, err := netlink.LinkByName(contIface.Name)
		if err != nil {
			return err
		}

		if err = netlink.LinkSetUp(contVeth); err != nil {
			return errors.Wrapf(err, "setup NS network: failed to seti link %q up", contIface.Name)
		}

		if err = netlink.AddrAdd(contVeth, &netlink.Addr{IPNet: netlink.NewIPNet(ipv4Addr)}); err != nil {
			return errors.Wrapf(err, "setup NS network: failed to add IP addr to %q", contIface.Name)
		}

		// Add a connected route to a dummy next hop (169.254.1.1)
		// # ip route show
		// default via 169.254.1.1 dev eth0  src 10.0.32.140
		// 169.254.1.1 dev eth0  scope link
		gw := net.IPv4(169, 254, 1, 1)
		if err = netlink.RouteAdd(&netlink.Route{
			LinkIndex: contVeth.Attrs().Index,
			Scope:     netlink.SCOPE_LINK,
			Dst:       netlink.NewIPNet(gw),
		}); err != nil {
			return errors.Wrap(err, "setup NS network: failed to direct route")
		}

		defaultRoute := netlink.Route{
			LinkIndex: contVeth.Attrs().Index,
			Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
			Scope:     netlink.SCOPE_UNIVERSE,
			Gw:        gw,
			Src:       ipv4Addr,
		}

		if err = netlink.RouteReplace(&defaultRoute); err != nil {
			return errors.Wrap(err, "setup NS network: failed to add default gateway")
		}

		// add static ARP entry for default gateway
		// we are using routed mode on the host and container need this static ARP entry to resolve its default gateway.
		neigh := &netlink.Neigh{
			LinkIndex:    contVeth.Attrs().Index,
			State:        netlink.NUD_PERMANENT,
			IP:           gw,
			HardwareAddr: hostVeth.Attrs().HardwareAddr,
		}

		if err = netlink.NeighAdd(neigh); err != nil {
			return errors.Wrap(err, "setup NS network: failed to add static ARP")
		}
		return nil
	})
}

func cmdAdd(args *skel.CmdArgs) (retErr error) {
	conf, _, err := loadConf(args)
	if err != nil {
		return err
	}
	logging.Verbosef("received CNI add request: ContainerID(%s) Netns(%s) IfName(%s) Args(%s) Path(%s) argsStdinData(%s)",
		args.ContainerID, args.Netns, args.IfName, args.Args, args.Path, args.StdinData)

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return errors.Wrapf(err, "failed to open netns %q", args.Netns)
	}
	defer netns.Close() // nolint

	hostIface, contIface, err := setupVeth(netns, args.IfName)
	if err != nil {
		return err
	}

	ipamRet, err := ipam.ExecAdd(conf.IPAM.Type, args.StdinData)
	if err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			logging.Errorf("failed process cni add request: %v", retErr) // nolint
			if err = ipam.ExecDel(conf.IPAM.Type, args.StdinData); err != nil {
				logging.Errorf("failed to rollback result from ipam %v: %v", ipamRet, err) // nolint
			}
		}
	}()
	result, err := current.NewResultFromResult(ipamRet)
	if err != nil {
		return err
	}
	if len(result.IPs) == 0 {
		return errors.New("IPAM plugin returned missing IP config")
	}
	logging.Debugf("get result from ipam: %v", result)

	var ipv4Addr net.IP
	for _, ipc := range result.IPs {
		if ipc.Version == "4" {
			ipv4Addr = ipc.Address.IP
		}
		// All addresses belong to the ipvlan interface
		ipc.Interface = current.Int(0)
	}
	if ipv4Addr == nil {
		return errors.New("no ipv4 address from ipam")
	}

	if err := configIface(netns, hostIface, contIface, ipv4Addr); err != nil {
		return err
	}

	result.Interfaces = []*current.Interface{contIface}
	result.DNS = conf.DNS

	return types.PrintResult(result, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	conf, _, err := loadConf(args)
	if err != nil {
		return err
	}

	logging.Verbosef("received CNI del request: ContainerID(%s) Netns(%s) IfName(%s) Args(%s) Path(%s) argsStdinData(%s)",
		args.ContainerID, args.Netns, args.IfName, args.Args, args.Path, args.StdinData)

	if err := ipam.ExecDel(conf.IPAM.Type, args.StdinData); err != nil {
		return err
	}

	// see https://github.com/kubernetes/kubernetes/issues/20379#issuecomment-255272531
	if args.Netns == "" {
		return nil
	}

	return ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
		_, err := ip.DelLinkByNameAddr(args.IfName)
		if err != nil && err == ip.ErrLinkNotFound {
			return nil
		}
		return err
	})
}

func cmdCheck(args *skel.CmdArgs) error {
	conf, _, err := loadConf(args)
	if err != nil {
		return err
	}

	// run the IPAM plugin and get back the config to apply
	err = ipam.ExecCheck(conf.IPAM.Type, args.StdinData)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, cniSpecVersion.All, "bcs-underlay-cni")
}
