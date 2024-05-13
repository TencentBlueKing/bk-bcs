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

package networkutil

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"

	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/vishvananda/netlink"
)

// Interface defines ip rule operaions, netlink operations, iptables operations and so on
type Interface interface {
	// GetAvailableHostIP get available host ip by network interface names
	GetAvailableHostIP(ifnames []string) (string, string, error)
	// SetHostNetwork disable rp_filter, enable ip_forwarding
	SetHostNetwork(instanceEth string, routeTableIDs map[string]string) error
	// SetUpNetworkInterface set elastic network interface up
	SetUpNetworkInterface(ip, cidrBlock, eniMac, eniName string, table, mtu int, rules []netlink.Rule) error
	// SetDownNetworkInterface set elastic network interface down
	SetDownNetworkInterface(ip, cidrBlock, eniMac, eniName string, table int, rules []netlink.Rule) error
	// GetHostName get hostname
	GetHostName() (string, error)
	// GetNetworkInterfaceMaxIndex get max index of network interfaces on hosts
	// GetNetworkInterfaceMaxIndex() (int, error)
	// LinkByMac find link by mac
	LinkByMac(mac string) (netlink.Link, error)
	// RuleList list route table rules
	RuleList() ([]netlink.Rule, error)
}

// NetUtil network util
type NetUtil struct{}

// GetAvailableHostIP get available host ip by network interface names
func (nc *NetUtil) GetAvailableHostIP(ifnames []string) (string, string, error) {
	linkList, err := netlink.LinkList()
	if err != nil {
		blog.Errorf("failed to list links, err %s", err.Error())
		return "", "", fmt.Errorf("failed to list links, err %s", err.Error())
	}
	netifMap := make(map[string]string)
	for _, l := range linkList {
		addrs, err := netlink.AddrList(l, unix.AF_INET)
		if err != nil {
			blog.Errorf("failed to list addrs for link with mac %s", l.Attrs().HardwareAddr)
			return "", "", fmt.Errorf("failed to list addrs for link with mac %s", l.Attrs().HardwareAddr)
		}
		if len(addrs) == 0 {
			blog.V(3).Infof("skip link with mac %s", l.Attrs().HardwareAddr)
			continue
		}
		netifMap[l.Attrs().Name] = addrs[0].IP.String()
	}
	for _, name := range ifnames {
		if addr, ok := netifMap[name]; ok {
			return addr, name, nil
		}
	}
	return "", "", fmt.Errorf("no available ip for network interfaces %+v", ifnames)
}

// SetHostNetwork disable rp_filter, enable ip_forwarding
func (nc *NetUtil) SetHostNetwork(instanceEth string, routeTableIDs map[string]string) error {

	// disable all rp_filter
	allRpFilterValue, err := getRpFilter("all")
	blog.Infof("get system all.rp_filter = %d", allRpFilterValue)
	if err != nil {
		return fmt.Errorf("get all.rp_filter failed, err %s", err.Error())
	}
	if allRpFilterValue != 0 {
		if ok := setRpFilter("all", false); !ok {
			blog.Warnf("set all.rp_filter to 0 failed")
		}
	}

	// disable rp_filter to deal with asymmetric route problem
	rpFilterValue, err := getRpFilter(instanceEth)
	blog.Infof("get system %s.rp_filter = %d", instanceEth, rpFilterValue)
	if err != nil {
		return fmt.Errorf("get rpFilter failed, err %s", err.Error())
	}
	if rpFilterValue != 0 {
		if ok := setRpFilter(instanceEth, false); !ok {
			blog.Warnf("set rp_filter %s to 0 failed", instanceEth)
		}
	}

	// set ip forward
	ipForwardValue, err := getIPForward()
	if err != nil {
		return err
	}
	blog.Infof("get system ipv4.ip_forward = %d", ipForwardValue)
	if ipForwardValue != 1 {
		blog.Infof("enable system ipv4.ip_forward ......")
		if ok := setIPForward(true); !ok {
			blog.Warnf("enable system ipv4.ip_forward failed")
		}
	}

	// create route tables
	blog.Infof("ensure route table IDs in /etc/iproute2/rt_tables......")
	if ok := ensureRouteTables(routeTableIDs); !ok {
		blog.Errorf("ensure route tables failed")
		return fmt.Errorf("ensure route tables failed")
	}
	blog.Infof("ensure route table IDs successfully")

	return nil
}

func isDefaultRouteExisted(routes []netlink.Route, route netlink.Route) bool {
	for _, r := range routes {
		if r.LinkIndex == route.LinkIndex &&
			reflect.DeepEqual(r.Gw, route.Gw) &&
			r.Scope == route.Scope &&
			r.Table == route.Table {
			return true
		}
	}
	return false
}

func findToTableRule(rules []netlink.Rule, rule *netlink.Rule) bool {
	for _, r := range rules {
		if r.Table == rule.Table {
			if r.Dst != nil && rule.Dst != nil {
				if r.Dst.String() == rule.Dst.String() {
					return true
				}
			}
		}
	}
	return false
}

func findFromTableRule(rules []netlink.Rule, rule *netlink.Rule) bool {
	for _, r := range rules {
		if r.Table == rule.Table {
			if r.Src != nil && rule.Src != nil {
				if r.Src.String() == rule.Src.String() {
					return true
				}
			}
		}
	}
	return false
}

func deleteEniDefaultMainRoute(eniLink netlink.Link, cidr *net.IPNet, ip net.IP) error {
	defaultRoute := netlink.Route{
		Dst:   cidr,
		Src:   ip,
		Table: constant.RouteTableMain,
		Scope: netlink.SCOPE_LINK,
	}
	routes, err := netlink.RouteListFiltered(unix.AF_INET, &defaultRoute,
		netlink.RT_FILTER_TABLE|netlink.RT_FILTER_SCOPE)
	if err != nil {
		return fmt.Errorf("failed to list route list with route %+v , err %s", defaultRoute, err.Error())
	}
	isExisted := false
	for _, r := range routes {
		if r.LinkIndex == eniLink.Attrs().Index &&
			r.Src != nil && r.Src.String() == defaultRoute.Src.String() &&
			r.Scope == defaultRoute.Scope &&
			r.Table == defaultRoute.Table {
			isExisted = true
			break
		}
	}
	if isExisted {
		blog.Infof("do delete eni default route %v ", defaultRoute)
		if err := netlink.RouteDel(&defaultRoute); err != nil {
			blog.Errorf("delete eni default route %v failed, err %s", defaultRoute, err.Error())
			return fmt.Errorf("delete eni default route %v failed, err %s", defaultRoute, err.Error())
		}
	}
	return nil
}

// SetUpNetworkInterface set up network interface
func (nc *NetUtil) SetUpNetworkInterface(
	addr, cidrBlock, eniMac, eniIfaceName string, table, mtu int,
	existedRules []netlink.Rule) error {

	blog.Infof("set up network interface, addr %s, cidrBlock %s, eniMac %s, name %s, table %d",
		addr, cidrBlock, eniMac, eniIfaceName, table)
	eniLink, err := nc.LinkByMac(eniMac)
	if err != nil {
		return err
	}
	_, cidrNet, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		return fmt.Errorf("parse cidr block %s to IPNet failed, err %s", cidrBlock, err.Error())
	}

	// set netlink name if necessary
	blog.Infof("checking netlink name ......")
	if curName := eniLink.Attrs().Name; curName != eniIfaceName {
		blog.Infof("set netlink name %s to name %s", curName, eniIfaceName)
		if err := netlink.LinkSetName(eniLink, eniIfaceName); err != nil {
			return fmt.Errorf("set netlink with mac %s to name %s failed, err %s", eniMac, eniIfaceName, err.Error())
		}
	}

	// if netlink is not up, set up
	blog.Infof("checking status of netlink ......")
	if eniLink.Attrs().Flags&net.FlagUp == 0 {
		blog.Infof("set netlink mtu %d", mtu)
		if err := netlink.LinkSetMTU(eniLink, mtu); err != nil {
			return fmt.Errorf("set netlink mtu %s failed, err %s", eniIfaceName, err.Error())
		}
		blog.Infof("set netlink name %s up", eniIfaceName)
		if err := netlink.LinkSetUp(eniLink); err != nil {
			return fmt.Errorf("set up netlink %s failed, err %s", eniIfaceName, err.Error())
		}
	}

	// set ip addr
	blog.Infof("checking ip addr of netlink ......")
	ipNet := &net.IPNet{
		IP:   net.ParseIP(addr),
		Mask: cidrNet.Mask,
	}
	ipAddr := netlink.Addr{
		IPNet: ipNet,
		Label: "",
	}
	addrs, err := netlink.AddrList(eniLink, unix.AF_INET)
	if err != nil {
		return fmt.Errorf("failed to list ip addresses for eni with mac %s", eniMac)
	}
	if len(addrs) == 0 {
		if err := netlink.AddrAdd(eniLink, &ipAddr); err != nil {
			return fmt.Errorf("add addr %+v to link with mac %s failed, err %s", addr, eniMac, err.Error())
		}
	}

	// get the subnet cidr of applied ip address,
	// and get gateway according to subnet cidr
	blog.Infof("checking default routing ......")
	cidrIP, _, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		blog.Errorf("parse cidr %s failed, err %s", cidrBlock, err.Error())
		os.Exit(1)
	}
	gw := ip.NextIP(cidrIP)
	// ensure default route for eni
	eniTableRoute := netlink.Route{
		LinkIndex: eniLink.Attrs().Index,
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
		Scope:     netlink.SCOPE_UNIVERSE,
		Gw:        gw,
		Table:     table,
	}

	routes, err := netlink.RouteListFiltered(unix.AF_INET, &eniTableRoute,
		netlink.RT_FILTER_TABLE|netlink.RT_FILTER_SCOPE|netlink.RT_FILTER_GW)
	if err != nil {
		return fmt.Errorf("failed to list route list with route %+v , err %s", eniTableRoute, err.Error())
	}
	blog.Infof("find routes %+v", routes)

	if !isDefaultRouteExisted(routes, eniTableRoute) {
		blog.Infof("add default route %+v", eniTableRoute)
		// add default route
		err = netlink.RouteAdd(&eniTableRoute)
		if err != nil {
			return fmt.Errorf("add default route %+v for table %d failed, err %s", eniTableRoute, table, err.Error())
		}
	}

	blog.Infof("checking ip rules ......")
	fromEniRule := netlink.NewRule()
	fromEniRule.Src = &net.IPNet{IP: net.ParseIP(addr), Mask: net.IPv4Mask(255, 255, 255, 255)}
	fromEniRule.Table = table
	fromEniRule.Priority = 2048

	if !findFromTableRule(existedRules, fromEniRule) {
		blog.Infof("add ip rule %+v", fromEniRule)
		// add from eni rule
		err = netlink.RuleAdd(fromEniRule)
		if err != nil {
			return fmt.Errorf("add from eni rule %+v for table %d failed, err %s", fromEniRule, table, err.Error())
		}
	}

	toEniRule := netlink.NewRule()
	toEniRule.Dst = &net.IPNet{IP: net.ParseIP(addr), Mask: net.IPv4Mask(255, 255, 255, 255)}
	toEniRule.Table = table
	toEniRule.Priority = 2048

	if !findToTableRule(existedRules, toEniRule) {
		blog.Infof("add ip rule %+v", toEniRule)
		// add to eni rule
		err = netlink.RuleAdd(toEniRule)
		if err != nil {
			return fmt.Errorf("add to eni rule %+v for table %d failed, err %s", toEniRule, table, err.Error())
		}
	}

	blog.Infof("checking rp filter ......")
	rpFilterValue, err := getRpFilter(eniIfaceName)
	blog.Infof("get system %s.rp_filter = %d", eniIfaceName, rpFilterValue)
	if err != nil {
		return fmt.Errorf("get rpFilter failed, err %s", err.Error())
	}
	if rpFilterValue != 0 {
		if ok := setRpFilter(eniIfaceName, false); !ok {
			blog.Warnf("set rp_filter %s to %s failed", eniIfaceName, "0")
		}
	}

	blog.Infof("checking eni default route in main table ......")
	if err := deleteEniDefaultMainRoute(eniLink, cidrNet, net.ParseIP(addr)); err != nil {
		return err
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

// SetDownNetworkInterface set down network interface
func (nc *NetUtil) SetDownNetworkInterface(
	addr, cidrBlock, eniMac, eniIfaceName string, table int,
	existedRules []netlink.Rule) error {

	blog.Infof("set down network interface, addr %s, cidrBlock %s, eniMac %s, name %s, table %s",
		addr, cidrBlock, eniMac, eniIfaceName, table)

	eniLink, err := nc.LinkByMac(eniMac)
	if err != nil {
		return err
	}

	blog.Infof("clean ip rules ......")
	fromEniRule := netlink.NewRule()
	fromEniRule.Src = &net.IPNet{IP: net.ParseIP(addr), Mask: net.IPv4Mask(255, 255, 255, 255)}
	fromEniRule.Table = table

	if findFromTableRule(existedRules, fromEniRule) {
		blog.Infof("del ip rule %+v", fromEniRule)
		// del from eni rule
		err = netlink.RuleDel(fromEniRule)
		if err != nil {
			blog.Warnf("del from eni rule %+v for table %d failed, err %s", fromEniRule, table, err.Error())
		}
	}

	toEniRule := netlink.NewRule()
	toEniRule.Dst = &net.IPNet{IP: net.ParseIP(addr), Mask: net.IPv4Mask(255, 255, 255, 255)}
	toEniRule.Table = table

	if findToTableRule(existedRules, toEniRule) {
		blog.Infof("del ip rule %+v", toEniRule)
		// add to eni rule
		err = netlink.RuleDel(toEniRule)
		if err != nil {
			blog.Warnf("add to eni rule %+v for table %d failed, err %s", toEniRule, table, err.Error())
		}
	}

	blog.Infof("checking default routing ......")
	cidrIP, _, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		blog.Errorf("parse cidr %s failed, err %s", cidrBlock, err.Error())
		os.Exit(1)
	}
	gw := ip.NextIP(cidrIP)
	// del default route for eni
	eniTableRoute := netlink.Route{
		LinkIndex: eniLink.Attrs().Index,
		Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
		Scope:     netlink.SCOPE_UNIVERSE,
		Gw:        gw,
		Table:     table,
	}
	routes, err := netlink.RouteListFiltered(unix.AF_INET, &eniTableRoute,
		netlink.RT_FILTER_TABLE|netlink.RT_FILTER_SCOPE|netlink.RT_FILTER_GW)
	if err != nil {
		return fmt.Errorf("failed to list route list with route %+v , err %s", eniTableRoute, err.Error())
	}
	blog.Infof("find routes %+v", routes)

	if isDefaultRouteExisted(routes, eniTableRoute) {
		blog.Infof("del default route %+v", eniTableRoute)
		// del default route
		err = netlink.RouteDel(&eniTableRoute)
		if err != nil {
			return fmt.Errorf("del default route %+v for table %d failed, err %s", eniTableRoute, table, err.Error())
		}
	}

	// if netlink is up, set down
	blog.Infof("checking status of netlink ......")
	if eniLink.Attrs().Flags&net.FlagUp != 0 {
		blog.Infof("set netlink name %s down", eniIfaceName)
		if err := netlink.LinkSetDown(eniLink); err != nil {
			return fmt.Errorf("set down netlink %s failed, err %s", eniIfaceName, err.Error())
		}
	}

	return nil
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

// RuleList list rule
func (nc *NetUtil) RuleList() ([]netlink.Rule, error) {
	rules, err := netlink.RuleList(unix.AF_INET)
	if err != nil {
		return nil, fmt.Errorf("failed to list rule list, err %s", err.Error())
	}
	return rules, nil
}
