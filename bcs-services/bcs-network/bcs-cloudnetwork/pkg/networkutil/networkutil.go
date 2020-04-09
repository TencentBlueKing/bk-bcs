/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package networkutil

import (
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"reflect"
	"strings"

	"bk-bcs/bcs-common/common/blog"

	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/vishvananda/netlink"
)

// Interface defines ip rule operaions, netlink operations, iptables operations and so on
type Interface interface {
	// GetAvailableHostIP get available host ip by network interface names
	GetAvailableHostIP(ifnames []string) (string, error)
	// SetHostNetwork disable rp_filter, enable ip_forwarding
	SetHostNetwork()
	// SetUpNetworkInterface set elastic network interface up
	SetUpNetworkInterface(ip, cidrBlock, eniMac, eniName string, table int) error
	// GetHostName get hostname
	GetHostName() (string, error)
	// GetNetworkInterfaceMaxIndex get max index of network interfaces on hosts
	GetNetworkInterfaceMaxIndex() (int, error)
	// LinkByMac find link by mac
	LinkByMac(mac string) (netlink.Link, error)
}

// NetUtil network util
type NetUtil struct {
}

// GetAvailableHostIP get available host ip by network interface names
func (nc *NetUtil) GetAvailableHostIP(ifnames []string) (string, error) {
	linkList, err := netlink.LinkList()
	if err != nil {
		blog.Errorf("failed to list links, err %s", err.Error())
		return "", fmt.Errorf("failed to list links, err %s", err.Error())
	}
	netifMap := make(map[string]string)
	for _, l := range linkList {
		addrs, err := netlink.AddrList(l, unix.AF_INET)
		if err != nil {
			blog.Errorf("failed to list addrs for link with mac %s", l.Attrs().HardwareAddr)
			return "", fmt.Errorf("failed to list addrs for link with mac %s", l.Attrs().HardwareAddr)
		}
		if len(addrs) == 0 {
			blog.V(3).Infof("skip link with mac %s", l.Attrs().HardwareAddr)
			continue
		}
		netifMap[l.Attrs().Name] = addrs[0].IP.String()
	}
	for _, name := range ifnames {
		if addr, ok := netifMap[name]; ok {
			return addr, nil
		}
	}
	return "", fmt.Errorf("no available ip for network interfaces %+v", ifnames)
}

// SetHostNetwork disable rp_filter, enable ip_forwarding
func (nc *NetUtil) SetHostNetwork() {

}

func isRouteExisted(routes []netlink.Route, route netlink.Route) bool {
	for _, r := range routes {
		if r.LinkIndex == route.LinkIndex &&
			reflect.DeepEqual(r.Dst, route.Dst) &&
			r.Scope == route.Scope &&
			r.Table == route.Table {
			return true
		}
	}
	return false
}

func isRuleExisted(rules []netlink.Rule, rule netlink.Rule) bool {
	for _, r := range rules {
		if reflect.DeepEqual(r.Src, rule.Src) && reflect.DeepEqual(r.Dst, rule.Dst) {
			return true
		}
	}
	return false
}

// SetUpNetworkInterface set up network interface
func (nc *NetUtil) SetUpNetworkInterface(addr, cidrBlock, eniMac, eniName string, table int) error {
	eniLink, err := nc.LinkByMac(eniMac)
	if err != nil {
		return err
	}
	_, cidrNet, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		return fmt.Errorf("parse cidr block %s to IPNet failed, err %s", cidrBlock, err.Error())
	}

	// set netlink name if necessary
	if curName := eniLink.Attrs().Name; curName != eniName {
		blog.Infof("set netlink name %s to name %s", curName, eniName)
		if err := netlink.LinkSetName(eniLink, eniName); err != nil {
			return fmt.Errorf("set netlink with mac %s to name %s failed, err %s", eniMac, eniName, err.Error())
		}
	}

	// if netlink is not up, set up
	if eniLink.Attrs().Flags&net.FlagUp == 0 {
		blog.Infof("set netlink name %s up", eniName)
		if err := netlink.LinkSetUp(eniLink); err != nil {
			return fmt.Errorf("set up netlink %s failed, err %s", eniName, err.Error())
		}
	}

	// set ip addr
	ipNet := &net.IPNet{
		IP:   net.ParseIP(addr),
		Mask: cidrNet.Mask,
	}
	ipAddr := &netlink.Addr{
		IPNet: ipNet,
		Label: "",
	}
	addrs, err := netlink.AddrList(eniLink, unix.AF_INET)
	if err != nil {
		return fmt.Errorf("failed to list ip addresses for eni with mac %s", eniMac)
	}
	if len(addrs) == 0 {
		if err := netlink.AddrAdd(eniLink, ipAddr); err != nil {
			return fmt.Errorf("add addr %+v to link with mac %s failed, err %s", addr, eniMac, err.Error())
		}
	}

	// get the subnet cidr of applied ip address,
	// and get gateway according to subnet cidr
	cidrIP, _, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		blog.Errorf("parse cidr %s failed, err %s", cidrBlock, err.Error())
		os.Exit(1)
	}
	gw := ip.NextIP(cidrIP)
	// ensure default route for eni
	defaultRoute := netlink.Route{
		LinkIndex: eniLink.Attrs().Index,
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
		Scope:     netlink.SCOPE_UNIVERSE,
		Gw:        gw,
		Table:     table,
	}

	routes, err := netlink.RouteList(eniLink, unix.AF_INET)
	if err != nil {
		return fmt.Errorf("failed to list route list for netlink with mac %s, err %s", eniMac, err.Error())
	}

	if !isRouteExisted(routes, defaultRoute) {
		// add default route
		err = netlink.RouteAdd(&defaultRoute)
		if err != nil {
			return fmt.Errorf("add default route %+v for table %d failed, err %s", defaultRoute, table, err.Error())
		}
	}

	fromEniRule := netlink.NewRule()
	fromEniRule.Src = ipNet
	fromEniRule.Table = table

	toEniRule := netlink.NewRule()
	toEniRule.Dst = ipNet
	toEniRule.Table = table

	rules, err := netlink.RuleList(unix.AF_INET)
	if err != nil {
		return fmt.Errorf("failed to list rule list for netlink with mac %s, err %s", eniMac, err.Error())
	}

	if !isRuleExisted(rules, *fromEniRule) {
		// add from eni rule
		err = netlink.RuleAdd(fromEniRule)
		if err != nil {
			return fmt.Errorf("add from eni rule %+v for table %d failed, err %s", fromEniRule, table, err.Error())
		}
	}
	if !isRuleExisted(rules, *toEniRule) {
		// add to eni rule
		err = netlink.RuleAdd(toEniRule)
		if err != nil {
			return fmt.Errorf("add to eni rule %+v for table %d failed, err %s", toEniRule, table, err.Error())
		}
	}

	return nil
}

// GetHostName get hostname
func (nc *NetUtil) GetHostName() (string, error) {
	return os.Hostname()
}

// GetNetworkInterfaceMaxIndex get max index of network interfaces on hosts
func (nc *NetUtil) GetNetworkInterfaceMaxIndex() (int, error) {
	linkList, err := netlink.LinkList()
	if err != nil {
		blog.Errorf("failed to list links, err %s", err.Error())
		return -1, fmt.Errorf("failed to list links, err %s", err.Error())
	}
	maxIndex := 1
	for _, link := range linkList {
		curIndex := link.Attrs().Index
		if curIndex > maxIndex {
			maxIndex = curIndex
		}
	}
	return maxIndex, nil
}

// LinkByMac find link by mac
func (nc *NetUtil) LinkByMac(mac string) (netlink.Link, error) {
	linkList, err := netlink.LinkList()
	if err != nil {
		blog.Errorf("failed to list links, err %s", err.Error())
		return nil, err
	}
	for _, link := range linkList {
		macFound := link.Attrs().HardwareAddr.String()
		linkName := link.Attrs().Name
		blog.V(3).Infof("link with mac: %s, name: %s", macFound, linkName)
		if strings.ToLower(macFound) == strings.ToLower(mac) {
			blog.V(3).Infof("found eni with mac %s", mac)
			return link, nil
		}
	}
	return nil, fmt.Errorf("no found eni with mac %s", mac)
}
