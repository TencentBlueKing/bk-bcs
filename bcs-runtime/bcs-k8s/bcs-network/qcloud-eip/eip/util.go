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
	"net"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/vishvananda/netlink"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

func getBridgeRoutes(tableID int, bridgeName string) ([]netlink.Route, error) {
	link, err := netlink.LinkByName(bridgeName)
	if err != nil {
		blog.Errorf("there is no bridge with name %s", bridgeName)
		return nil, fmt.Errorf("there is no bridge with name %s", bridgeName)
	}
	addr, _, _ := getIPAddrByName(bridgeName)
	if len(addr) == 0 {
		blog.Errorf("get bridge %s addr failed", bridgeName)
		return nil, fmt.Errorf("get bridge %s addr failed", bridgeName)
	}
	ipObj := net.ParseIP(addr)
	if ipObj == nil {
		blog.Errorf("parse ip %s return nil", addr)
		return nil, fmt.Errorf("parse ip %s return nil", addr)
	}
	bridgeRouteRule := netlink.Route{
		LinkIndex: link.Attrs().Index,
		Scope:     netlink.SCOPE_LINK,
		Src:       ipObj,
		Table:     tableID,
	}
	routes, err := netlink.RouteListFiltered(unix.AF_INET, &bridgeRouteRule,
		netlink.RT_FILTER_TABLE|netlink.RT_FILTER_SCOPE|netlink.RT_FILTER_OIF|netlink.RT_FILTER_SRC)
	if err != nil {
		blog.Errorf("failed to list route list with route %+v , err %s", bridgeRouteRule, err.Error())
		return nil, err
	}
	return routes, nil
}

// GetIPAddrByName return eni ip address, mask, mac address
func getIPAddrByName(name string) (string, int, string) {
	//get ip address
	netFace, err := net.InterfaceByName(name)
	if err != nil {
		blog.Warnf("Got network interface %s ip address failed, %s", name, err)
		return "", 0, ""
	}
	netAddrs, err := netFace.Addrs()
	if err != nil {
		blog.Warnf("Got no network ip address under interface %s, %s", name, err)
		return "", 0, ""
	}
	netAddr := netAddrs[0]
	ipnet, ok := netAddr.(*net.IPNet)
	if !ok {
		blog.Warnf("Got no ip address from network interface %s", name)
		return "", 0, ""
	}
	str := ipnet.IP.String()
	mask, _ := ipnet.Mask.Size()
	return str, mask, netFace.HardwareAddr.String()
}

func parseCIDR(cidr string) (string, int, error) {
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		blog.Errorf("parse cidr %s addr failed, err %s", cidr, err.Error())
		return "", 0, err
	}
	strs := strings.Split(cidr, "/")
	if len(strs) != 2 {
		blog.Errorf("cidr %s format error", cidr)
		return "", 0, fmt.Errorf("cidr %s format error", cidr)
	}
	mask, _ := strconv.Atoi(strs[1])
	return strs[0], mask, nil
}

func getMaxPrivateIPNumPerENI(cores, mem int) int {
	if cores == 1 && mem == 1 {
		return 2
	}
	if cores == 1 && mem > 1 {
		return 6
	}
	if cores == 2 {
		return 10
	}
	if cores == 4 && mem <= 16 {
		return 10
	}
	if cores == 4 && mem > 16 {
		return 20
	}
	if cores >= 8 && cores <= 12 {
		return 20
	}
	if cores > 12 {
		return 30
	}
	blog.Warnf("get max private ip num per eni failed with cores %d and mem %d, use 2 default", cores, mem)
	return 2
}

func getMaxENINumPerCVM(cores, mem int) int {
	if cores == 1 && mem == 1 {
		return 2
	}
	if cores == 1 && mem > 1 {
		return 2
	}
	if cores == 2 {
		return 2
	}
	if cores == 4 && mem <= 16 {
		return 4
	}
	if cores == 4 && mem > 16 {
		return 4
	}
	if cores >= 8 && cores <= 12 {
		return 6
	}
	if cores > 12 {
		return 8
	}
	blog.Warnf("get max eni num per cvm failed with cores %d and mem %d, use 2 default", cores, mem)
	return 2
}
