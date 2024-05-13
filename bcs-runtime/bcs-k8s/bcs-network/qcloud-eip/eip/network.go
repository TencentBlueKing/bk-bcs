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
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/vishvananda/netlink"
)

func appnedFile(filepath, value string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(value); err != nil {
		return err
	}
	return nil
}

func addRouteTables(ids []int) bool {
	var err error
	for index, id := range ids {
		err = appnedFile("/etc/iproute2/rt_tables", fmt.Sprintf("%d eni%d\n", id, index))
		if err != nil {
			blog.Errorf("error to add '%d eni%d' into /etc/iproute2/rt_tables, err %s", id, index, err.Error())
			return false
		}
		blog.Infof("add '%d eni%d' successfully", id, index)
	}
	return true
}

func cleanEniRtTable(content string) string {
	if !strings.Contains(content, "eni") {
		return content
	}
	strs := strings.Split(content, "\n")
	if len(strs) > 1 && strs[len(strs)-1] == "" {
		strs = strs[0 : len(strs)-1]
	}
	ret := ""
	for _, str := range strs {
		if strings.Contains(str, "eni") {
			continue
		}
		ret = ret + str + "\n"
	}
	return ret
}

func delRouteTables() bool {
	bytes, err := ioutil.ReadFile("/etc/iproute2/rt_tables")
	if err != nil {
		blog.Errorf("read file /etc/iproute2/rt_tables failed, err %s", err.Error())
		return false
	}
	str := string(bytes)
	retStr := cleanEniRtTable(str)
	if retStr == str {
		blog.Infof("no eni record in rt_table, no need to clean")
		return true
	}
	err = ioutil.WriteFile("/etc/iproute2/rt_tables", []byte(retStr), 0644)
	if err != nil {
		blog.Errorf("failed to clean /etc/iproute2/rt_tables")
		return false
	}
	return true
}

func setIPForward(value bool) bool {
	var err error
	v := []byte{'1', '\n'}
	if !value {
		v = []byte{'0', '\n'}
	}
	if err = ioutil.WriteFile("/proc/sys/net/ipv4/ip_forward", v, 0644); err != nil {
		blog.Errorf("failed to set ipv4.ip_forward to %s, err %s", string(v), err.Error())
		return false
	}
	blog.Infof("success to set ipv4.ip_forward to %s", string(v))
	return true
}

func setRpFilter(ifname string, value bool) bool {
	var err error
	v := []byte{'1', '\n'}
	if !value {
		v = []byte{'0', '\n'}
	}
	procPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/rp_filter", ifname)
	if err = ioutil.WriteFile(procPath, v, 0644); err != nil {
		blog.Errorf("failed to set ipv4.conf.%s.rp_filter to %s, err %s", ifname, string(v), err.Error())
		return false
	}
	blog.Infof("success to set ipv4.conf.%s.rp_filter to %s", ifname, string(v))
	return true
}

func setupNetworkInterface(ip, cidrBlock, eniName, eniMac string) error {
	linkList, err := netlink.LinkList()
	if err != nil {
		blog.Errorf("failed to get link list ,err %s", err.Error())
		return fmt.Errorf("failed to get link list ,err %s", err.Error())
	}
	for _, link := range linkList {
		mac := link.Attrs().HardwareAddr.String()
		ifname := link.Attrs().Name
		blog.Infof("get link with mac: %s, name: %s", mac, ifname)
		if strings.ToLower(mac) == strings.ToLower(eniMac) {
			blog.Infof("eni found")
			_, cidrNet, err := net.ParseCIDR(cidrBlock)
			if err != nil {
				blog.Infof("parse cidr block %s to IPNet failed, err %s", cidrBlock, err.Error())
				return fmt.Errorf("parse cidr block %s to IPNet failed, err %s", cidrBlock, err.Error())
			}
			if err := netlink.LinkSetName(link, eniName); err != nil {
				blog.Infof("set %s to name %s failed, err %s", ifname, eniName, err.Error())
				return fmt.Errorf("set %s to name %s failed, err %s", ifname, eniName, err.Error())
			}
			if err := netlink.LinkSetUp(link); err != nil {
				blog.Infof("set up ifname %s with mac %s failed, err %s", ifname, mac, err.Error())
				return fmt.Errorf("set up ifname %s with mac %s failed, err %s", ifname, mac, err.Error())
			}
			ipNet := &net.IPNet{
				IP:   net.ParseIP(ip),
				Mask: cidrNet.Mask,
			}
			addr := &netlink.Addr{
				IPNet: ipNet,
				Label: "",
			}
			if err := netlink.AddrAdd(link, addr); err != nil {
				blog.Infof("add Addr %v to %s failed, err %s", addr, ifname, err.Error())
				return fmt.Errorf("add Addr %v to %s failed, err %s", addr, ifname, err.Error())
			}
			blog.Infof("set up link %s successfully", ifname)
			return nil
		}
		blog.Infof("skip")
		continue
	}
	blog.Errorf("found no eni named %s", eniName)
	return fmt.Errorf("found no eni named %s", eniName)
}
