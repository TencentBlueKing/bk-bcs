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
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	netsvc "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/qcloud-eip/conf"
)

const (
	// EniRouteTableStartID start eni route table id
	EniRouteTableStartID = 100
	// MainRouteTableID main route table id
	MainRouteTableID = 254
)

// EIP object for use tencent network interface
type EIP struct{}

// Init implements eip interface
func (eip *EIP) Init(file string, eniNum int, ipNum int) {
	// load config file
	blog.Infof("load config %s ......", file)
	netConf, err := conf.LoadConfFromFile(file)
	if err != nil {
		blog.Errorf("load net config from file %s failed, err %s", file, err.Error())
		os.Exit(1)
	}
	// get the master interface address in conf file
	// use this address as cvm address
	hostIPAddr, _, _ := getIPAddrByName(netConf.Master)
	if len(hostIPAddr) == 0 {
		blog.Errorf("get cvm ip addr by name %s failed", netConf.Master)
		os.Exit(1)
	}

	// describe local cvm info
	blog.Infof("describe local cvm info ......")
	instanceClient := newInstanceClient(netConf)
	if instanceClient == nil {
		blog.Errorf("create cvm instance client failed, err %s", err.Error())
		os.Exit(1)
	}
	// get cvm instance info by master network interface address
	instance, err := instanceClient.describeInstanceByIP(hostIPAddr)
	if err != nil {
		blog.Errorf("describe instance by ip %s failed, err %s", hostIPAddr, err.Error())
		os.Exit(1)
	}

	// describe subnet info
	// use cvm subnet id if there is no subnet id in config
	blog.Infof("describe eni subnet %s info ......", netConf.SubnetID)
	if len(netConf.SubnetID) == 0 {
		blog.Infof("miss subnetID in config, use cvm subnetID %s", *instance.VirtualPrivateCloud.SubnetId)
		netConf.SubnetID = *instance.VirtualPrivateCloud.SubnetId
	}
	vpcClient := newVPCClient(netConf, *instance.VirtualPrivateCloud.VpcId)
	subnet, err := vpcClient.querySubnet(netConf.SubnetID)
	if err != nil {
		blog.Errorf("query subnet with subnet id %s failed, err %s", netConf.SubnetID, err.Error())
		os.Exit(1)
	}

	// validate network interface num
	blog.Infof("going to create enis, total %d ......", eniNum)
	// calculate the network interface per cvm according to tencent cloud api doc
	eniLimit := getMaxENINumPerCVM(int(*instance.CPU), int(*instance.Memory))
	if eniNum < 1 {
		blog.Errorf("eni num can not less than 1")
		os.Exit(1)
	}
	// the number of network interface to be applied should be less than max limit - 1
	// because master network interface is involved when talking about network interface limitation
	if eniNum > eniLimit-1 {
		blog.Errorf("eni num %d bigger than max eni num %d for %d cpu %d mem", eniNum, eniLimit, int(*instance.CPU), int(*instance.Memory))
		os.Exit(1)
	}

	// record route tables ids
	var routeTableIDs []int
	// map for link name to route table id
	routeTableIDMap := make(map[string]int)
	// record all available ips
	secondaryIPMap := make(map[string][]string)
	// according to the demanded network interface number and ip address number,
	// apply certain ip addresses for each newly applied network interface
	for i := 0; i < eniNum; i++ {

		// for each eni, there is a IP num limitation
		// apply (limitation - 1) ip, because the default primary ip
		ipLimit := getMaxPrivateIPNumPerENI(int(*instance.CPU), int(*instance.Memory))
		if ipNum < 0 {
			blog.Errorf("invalid ip num %d", ipNum)
			os.Exit(1)
		}
		if ipNum > ipLimit-1 || ipNum == 0 {
			blog.Errorf("get ip num %d, set ip num to %d", ipNum, ipLimit-1)
			ipNum = ipLimit - 1
		}
		// generate new network interface name according to its index
		newENIName := fmt.Sprintf("eni-%s-%d", *instance.InstanceId, i)
		blog.Infof("take over %s ", newENIName)
		newENI, err := vpcClient.TakeOverENI(*instance.InstanceId, uint64(ipNum), newENIName)
		if err != nil {
			blog.Errorf("take over eni %s for instance %s with %d ips failed, err %s", newENIName, *instance.InstanceId, uint64(ipLimit-1), err.Error())
			os.Exit(1)
		}
		blog.Infof("take over done")
		// generate new network interface local name according to its index
		eniLocalInterfaceName := fmt.Sprintf("%s%d", netConf.ENIPrefix, i)
		// set up eni with primary private ip address
		// record other applied ip addresses
		blog.Infof("set up eni %s ......", eniLocalInterfaceName)
		var secondaryIPs []string
		var primaryIP string
		for _, privateIPObj := range newENI.PrivateIpAddressSet {
			if *privateIPObj.Primary {
				primaryIP = *privateIPObj.PrivateIpAddress
			} else {
				blog.Infof("get secondary ip %s", *privateIPObj.PrivateIpAddress)
				secondaryIPs = append(secondaryIPs, *privateIPObj.PrivateIpAddress)
			}
		}
		secondaryIPMap[strings.ToLower(*newENI.MacAddress)] = secondaryIPs

		// if there is already some network interface name starting with netConf.ENIPrefix,
		// it means someone has already initialized the network of this cvm,
		// so stop the init action
		linkName := fmt.Sprintf("%s%d", netConf.ENIPrefix, i)
		blog.Infof("check eni %s......", linkName)
		eniIPAddr, _, _ := getIPAddrByName(linkName)
		if len(eniIPAddr) != 0 {
			blog.Errorf("%s is existed, no need to set up", linkName)
			continue
		}
		// set up eni
		err = setupNetworkInterface(primaryIP, *subnet.CidrBlock, eniLocalInterfaceName, *newENI.MacAddress)
		if err != nil {
			blog.Errorf("set up networkinterface %s with ip %s mac %s failed, err %s",
				eniLocalInterfaceName, primaryIP, *newENI.MacAddress, err.Error())
			os.Exit(1)
		}
		blog.Infof("set up eni with primary ip %s done", primaryIP)

		// for each newly applied network interface, create a route table in later steps,
		// here we just calculate the route table id and record it in an array
		tableID := EniRouteTableStartID + i
		routeTableIDs = append(routeTableIDs, tableID)
		routeTableIDMap[linkName] = tableID
	}

	// init network environment
	// enable system parameter ipv4.ip_forward
	blog.Infof("enable system ipv4.ip_forward ......")
	if ok := setIPForward(true); !ok {
		blog.Warnf("enable system ipv4.ip_forward failed")
	}
	// disable system parameter rp_filter for all network interface
	blog.Infof("disable system ip rp_filter ......")
	if ok := setRpFilter("all", false); !ok {
		blog.Warnf("set rp_filter %s to %s failed", "all", "0")
	}
	if ok := setRpFilter(netConf.Master, false); !ok {
		blog.Warnf("set rp_filter %s to %s failed", netConf.Master, "0")
	}
	for i := 0; i < eniNum; i++ {
		eniName := fmt.Sprintf("%s%d", netConf.ENIPrefix, i)
		if ok := setRpFilter(eniName, false); !ok {
			blog.Warnf("set rp_filter %s to %s failed", eniName, "0")
		}
	}
	blog.Infof("disable system ip rp_filter done")

	// create route tables
	blog.Infof("write new route table IDs in /etc/iproute2/rt_tables......")
	if ok := addRouteTables(routeTableIDs); !ok {
		blog.Errorf("add route tables failed")
		os.Exit(1)
	}
	blog.Infof("write route table IDs successfully")

	// get the subnet cidr of applied ip address,
	// and get gateway according to subnet cidr
	cidrIP, _, err := net.ParseCIDR(*subnet.CidrBlock)
	if err != nil {
		blog.Errorf("parse cidr %s failed, err %s", *subnet.CidrBlock, err.Error())
		os.Exit(1)
	}
	gw := ip.NextIP(cidrIP)
	// add default route into each route tables
	for eniName, id := range routeTableIDMap {
		link, err := netlink.LinkByName(eniName)
		if err != nil {
			blog.Errorf("get link by name %s failed, err %s", eniName, err.Error())
			os.Exit(1)
		}
		defaultRouteRule := &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
			Scope:     netlink.SCOPE_UNIVERSE,
			Gw:        gw,
			Table:     id,
		}
		// add default route
		err = netlink.RouteAdd(defaultRouteRule)
		if err != nil {
			blog.Errorf("add default route %v for table %d failed, err %s", defaultRouteRule, id, err.Error())
			os.Exit(1)
		}
	}

	// register ip pool for eni
	// use address of master network interface defined in config file as host index in netservice
	blog.Infof("register ip pool to netservice ......")
	hostAddr, _, _ := getIPAddrByName(netConf.Master)
	if hostAddr == "" {
		blog.Errorf("Got no ip address or mask in network interface %s", netConf.Master)
		os.Exit(1)
	}
	_, mask, err := parseCIDR(*subnet.CidrBlock)
	if err != nil {
		blog.Infof("parse cidr block %s to IPNet failed, err %s", *subnet.CidrBlock, err.Error())
		os.Exit(1)
	}
	pool := new(netsvc.NetPool)
	pool.Net = hostAddr
	pool.Cluster = netConf.ClusterID
	pool.Mask = mask
	pool.Hosts = append(pool.Hosts, hostAddr)
	pool.Gateway = "169.254.1.1"
	for _, ips := range secondaryIPMap {
		pool.Available = append(pool.Available, ips...)
	}
	// TODO: move this validation to the head of init action
	if netConf.NetService == nil {
		blog.Errorf("no netservice conf")
		os.Exit(1)
	}
	// create netservice client
	netSvcClient, err := NewNetSvcClient(netConf.NetService)
	if err != nil {
		blog.Errorf(
			"create netservice client failed, zookeeper: %s, key path %s, cert path %s, ca path %s, err %s",
			netConf.NetService.Zookeeper, netConf.NetService.Key,
			netConf.NetService.PubKey, netConf.NetService.Ca, err.Error())
		os.Exit(1)
	}
	// create ip pool in netservice
	err = netSvcClient.CreateOrUpdatePool(pool)
	if err != nil {
		blog.Errorf("create netservice pool failed, err %s", err.Error())
		os.Exit(1)
	}
	blog.Infof("register ip pool to netservice done")

	// get ip pool in netservice
	existedPool, err := netSvcClient.GetPool(pool)
	if err != nil {
		blog.Errorf("get netservice pool failed, err %s", err.Error())
		os.Exit(1)
	}
	existedIPs := make(map[string]string)
	for _, ip := range existedPool.Reserved {
		existedIPs[ip] = ip
	}
	for _, ip := range existedPool.Active {
		existedIPs[ip] = ip
	}

	// add newly ip addresses into ip pool in netservice
	blog.Infof("update ip instances to netservice")
	for mac, ips := range secondaryIPMap {
		for _, ip := range ips {
			if _, ok := existedIPs[ip]; ok {
				blog.Infof("ip %s already in pool, skip", ip)
				continue
			}
			ipIns := new(netsvc.IPInst)
			ipIns.IPAddr = ip
			ipIns.MacAddr = mac
			ipIns.Pool = pool.Net
			ipIns.Mask = pool.Mask
			ipIns.Gateway = pool.Gateway
			ipIns.Cluster = pool.Cluster
			err := netSvcClient.UpdateIPInstance(ipIns)
			if err != nil {
				blog.Errorf("update ip instance with %v failed, err %s", ipIns, err.Error())
				os.Exit(1)
			}
			blog.Infof("update ip instance with %v successfully", ipIns)
		}
	}

	blog.Infof("congratulations, node init successfully!")
}

// Recover recover applied network interface from reboot
// after cvm reboot, all the setted route table and route rule disappear,
// and all the applied network interface are down, we should take this action after reboot cvm
func (eip *EIP) Recover(file string, eniNum int) {
	blog.Infof("do recover after reboot....")
	netConf, err := conf.LoadConfFromFile(file)
	if err != nil {
		blog.Errorf("load net config from file %s failed, err %s", file, err.Error())
		os.Exit(1)
	}
	hostIPAddr, _, _ := getIPAddrByName(netConf.Master)
	if len(hostIPAddr) == 0 {
		blog.Errorf("get cvm ip addr by name %s failed", netConf)
		os.Exit(1)
	}

	//describe subnet info
	blog.Infof("describe local cvm info ......")
	instanceClient := newInstanceClient(netConf)
	if instanceClient == nil {
		blog.Errorf("create cvm instance client failed, err %s", err.Error())
		os.Exit(1)
	}
	instance, err := instanceClient.describeInstanceByIP(hostIPAddr)
	if err != nil {
		blog.Errorf("describe instance by ip %s failed, err %s", hostIPAddr, err.Error())
		os.Exit(1)
	}
	blog.Infof("describe eni subnet %s info ......", netConf.SubnetID)
	if len(netConf.SubnetID) == 0 {
		blog.Infof("miss subnetID in config, use cvm subnetID %s", *instance.VirtualPrivateCloud.SubnetId)
		netConf.SubnetID = *instance.VirtualPrivateCloud.SubnetId
	}
	vpcClient := newVPCClient(netConf, *instance.VirtualPrivateCloud.VpcId)
	subnet, err := vpcClient.querySubnet(netConf.SubnetID)
	if err != nil {
		blog.Errorf("query subnet with subnetid %s failed, err %s", netConf.SubnetID, err.Error())
		os.Exit(1)
	}

	// record route tables ids
	var routeTableIDs []int
	// eniNum from command line.
	// query each applied network interface by rule "eni-%s-%d", get primary ip address,
	// and set up network interface by (address, mac) in query result.
	for i := 0; i < eniNum; i++ {
		tableID := EniRouteTableStartID + i
		routeTableIDs = append(routeTableIDs, tableID)
		resENIName := fmt.Sprintf("eni-%s-%d", *instance.InstanceId, i)
		blog.Infof("recover eni with name %s", resENIName)
		resENIs, err := vpcClient.queryENI("", *instance.InstanceId, resENIName)
		if err != nil {
			blog.Errorf("query eni %s for instance %s failed, err %s", resENIName, *instance.InstanceId, err.Error())
			os.Exit(1)
		}
		if len(resENIs) != 1 {
			blog.Errorf("query eni %s for instance %s failed, return array length %d", resENIName, *instance.InstanceId, len(resENIs))
			os.Exit(1)
		}

		eniLocalInterfaceName := fmt.Sprintf("%s%d", netConf.ENIPrefix, i)
		blog.Infof("set up eni %s ......", eniLocalInterfaceName)
		for _, privateIPObj := range resENIs[0].PrivateIpAddressSet {
			if *privateIPObj.Primary {
				err = setupNetworkInterface(*privateIPObj.PrivateIpAddress, *subnet.CidrBlock, eniLocalInterfaceName, *resENIs[0].MacAddress)
				if err != nil {
					blog.Errorf("set up networkinterface %s with ip %s mac %s failed, err %s", eniLocalInterfaceName, *privateIPObj.PrivateIpAddress, *resENIs[0].MacAddress, err.Error())
				}
			}
		}
	}

	// init network environment
	blog.Infof("enable system ipv4.ip_forward ......")
	if ok := setIPForward(true); !ok {
		blog.Warnf("enable system ipv4.ip_forward failed")
	}
	blog.Infof("disable system ip rp_filter ......")
	if ok := setRpFilter("all", false); !ok {
		blog.Warnf("set rp_filter %s to %s failed", "all", "0")
	}
	if ok := setRpFilter(netConf.Master, false); !ok {
		blog.Warnf("set rp_filter %s to %s failed", netConf.Master, "0")
	}
	for i := 0; i < eniNum; i++ {
		eniName := fmt.Sprintf("%s%d", netConf.ENIPrefix, i)
		if ok := setRpFilter(eniName, false); !ok {
			blog.Warnf("set rp_filter %s to %s failed", eniName, "0")
		}
	}
	blog.Infof("disable system ip rp_filter done")

	// add default route into route tables
	for index, id := range routeTableIDs {
		linkName := fmt.Sprintf("%s%d", netConf.ENIPrefix, index)
		link, err := netlink.LinkByName(linkName)
		if err != nil {
			blog.Errorf("get link by name %s failed, err %s", linkName, err.Error())
			os.Exit(1)
		}
		cidrIP, _, err := net.ParseCIDR(*subnet.CidrBlock)
		if err != nil {
			blog.Errorf("parse cidr %s failed, err %s", *subnet.CidrBlock, err.Error())
			os.Exit(1)
		}
		gw := ip.NextIP(cidrIP)
		defaultRouteRule := &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
			Scope:     netlink.SCOPE_UNIVERSE,
			Gw:        gw,
			Table:     id,
		}
		// add default route
		err = netlink.RouteAdd(defaultRouteRule)
		if err != nil {
			blog.Errorf("add default route %v for table %d failed, err %s", defaultRouteRule, id, err.Error())
			os.Exit(1)
		}
	}

	blog.Infof("congratulations, node recovery successfully!")
}

// Check implements eip interface
func (eip *EIP) Check() {
	//TODO: check dirty routes and network interfaces
	blog.Infof("UNIMPLEMENTED METHOD")
}

// doDeregister deregister from netservice
func (eip *EIP) doDeregister(netConf *conf.NetConf) error {
	hostAddr, _, _ := getIPAddrByName(netConf.Master)
	if hostAddr == "" {
		blog.Errorf("get no ip address or mask in network interface %s", netConf.Master)
		return fmt.Errorf("get no ip address or mask in network interface %s", netConf.Master)
	}
	netSvcClient, err := NewNetSvcClient(netConf.NetService)
	if err != nil {
		blog.Errorf(
			"create netservice client failed, zookeeper: %s, key path %s, cert path %s, ca path %s, err %s",
			netConf.NetService.Zookeeper, netConf.NetService.Key,
			netConf.NetService.PubKey, netConf.NetService.Ca, err.Error())
		return fmt.Errorf(
			"create netservice client failed, zookeeper: %s, key path %s, cert path %s, ca path %s, err %s",
			netConf.NetService.Zookeeper, netConf.NetService.Key,
			netConf.NetService.PubKey, netConf.NetService.Ca, err.Error())
	}
	err = netSvcClient.DeletePool(netConf.ClusterID, hostAddr)
	if err != nil {
		blog.Errorf("delete host from netservice failed, err %s", err.Error())
		return fmt.Errorf("delete host from netservice failed, err %s", err.Error())
	}
	blog.Infof("de register ip pool %s/%s from net service successfully!", netConf.ClusterID, hostAddr)
	return nil
}

// Deregister implements eip interface
func (eip *EIP) Deregister(file string) {
	blog.Infof("load config %s ......", file)
	netConf, err := conf.LoadConfFromFile(file)
	if err != nil {
		blog.Errorf("load net config from file %s failed, err %s", file, err.Error())
		os.Exit(1)
	}
	// deregister ip pool for eni(by delete host in pool)
	blog.Infof("de register ip pool from service ......")
	err = eip.doDeregister(netConf)
	if err != nil {
		os.Exit(1)
	}
}

// Clean implements eip interface
// **Only to clean old qcloud plugin environment**
// 1. deregister from netservice
// 2. detach and delete each applied network interface
func (eip *EIP) Clean(file string) {
	blog.Infof("load config %s......", file)
	netConf, err := conf.LoadConfFromFile(file)
	if err != nil {
		blog.Errorf("load net config from file %s failed, err %s", file, err.Error())
		os.Exit(1)
	}
	hostIPAddr, _, _ := getIPAddrByName(netConf.Master)
	if len(hostIPAddr) == 0 {
		blog.Errorf("get cvm ip addr by name %s failed", netConf.Master)
		os.Exit(1)
	}
	blog.Infof("de register ip pool from net service ......")
	err = eip.doDeregister(netConf)
	if err != nil {
		blog.Errorf("do deregister failed, err %s", err.Error())
		//don't stop when clean, there can be no ip pool for this cvm
	}

	blog.Infof("desribe local cvm info ......")
	instanceClient := newInstanceClient(netConf)
	if instanceClient == nil {
		blog.Errorf("create cvm instance client failed, err %s", err.Error())
		os.Exit(1)
	}
	instance, err := instanceClient.describeInstanceByIP(hostIPAddr)
	if err != nil {
		blog.Errorf("describe instance by ip %s failed, err %s", hostIPAddr, err.Error())
		os.Exit(1)
	}

	blog.Infof("deleting applied enis......")
	vpcClient := newVPCClient(netConf, *instance.VirtualPrivateCloud.VpcId)
	networkInterfaceList, err := vpcClient.queryENI("", *instance.InstanceId, "")
	if err != nil {
		blog.Errorf("describe eni by ins-id %s failed, err %s", *instance.InstanceId, err.Error())
		os.Exit(1)
	}
	for _, networkInterface := range networkInterfaceList {
		if !*networkInterface.Primary {
			blog.Infof("detach eni %s from ins %s", *networkInterface.NetworkInterfaceId, *instance.InstanceId)
			err = vpcClient.detachENI(*networkInterface.NetworkInterfaceId, *instance.InstanceId)
			if err != nil {
				blog.Errorf("detach eni %s from ins %s faile, err %s, continue", *networkInterface.NetworkInterfaceId, *instance.InstanceId, err.Error())
				continue
			}
			err = vpcClient.waitForDetached(*networkInterface.NetworkInterfaceId, 5, 5)
			if err != nil {
				blog.Errorf("wait for eni %s available failed, err %s", *networkInterface.NetworkInterfaceId, err.Error())
				continue
			}
			blog.Infof("delete eni %s......", *networkInterface.NetworkInterfaceId)
			err = vpcClient.deleteENI(*networkInterface.NetworkInterfaceId)
			if err != nil {
				blog.Errorf("delete eni %s failed, err %s, continue", *networkInterface.NetworkInterfaceId, err.Error())
				continue
			}
		} else {
			blog.Infof("eni %s is primary eni, skip", *networkInterface.NetworkInterfaceId)
		}
	}

	blog.Infof("congratulations, clean cvm network successfully!")
}

// Release implements eip interface
// 1. deregister from netservice
// 2. set down, detach and delete each applied network interface
// 3. delete route tables
func (eip *EIP) Release(file string) {
	blog.Infof("load config %s ......", file)
	netConf, err := conf.LoadConfFromFile(file)
	if err != nil {
		blog.Errorf("load net config from file %s failed, err %s", file, err.Error())
		os.Exit(1)
	}
	hostIPAddr, _, _ := getIPAddrByName(netConf.Master)
	if len(hostIPAddr) == 0 {
		blog.Errorf("get cvm ip addr by name %s failed", netConf.Master)
		os.Exit(1)
	}
	blog.Infof("de register ip pool from net service ......")
	err = eip.doDeregister(netConf)
	if err != nil {
		blog.Errorf("do deregister failed, err %s", err.Error())
		os.Exit(1)
	}

	// create cvm client
	blog.Infof("desribe local cvm info ......")
	instanceClient := newInstanceClient(netConf)
	if instanceClient == nil {
		blog.Errorf("create cvm instance client failed, err %s", err.Error())
		os.Exit(1)
	}
	instance, err := instanceClient.describeInstanceByIP(hostIPAddr)
	if err != nil {
		blog.Errorf("describe instance by ip %s failed, err %s", hostIPAddr, err.Error())
		os.Exit(1)
	}
	// create vpc client
	vpcClient := newVPCClient(netConf, *instance.VirtualPrivateCloud.VpcId)

	//set down eni links
	links, err := netlink.LinkList()
	if err != nil {
		blog.Errorf("get all links in cvm failed, err %s", err.Error())
		os.Exit(1)
	}
	for _, link := range links {
		//skip the network interface without "ENIPrefix"
		linkName := link.Attrs().Name
		if !strings.HasPrefix(link.Attrs().Name, netConf.ENIPrefix) {
			blog.Infof("skip interface %s", linkName)
			continue
		}
		eniIPAddr, _, _ := getIPAddrByName(linkName)
		if len(eniIPAddr) == 0 {
			blog.Errorf("get eni ip addr by name %s failed", linkName)
			os.Exit(1)
		}

		blog.Infof("set down eni link...")
		eniLink, err := netlink.LinkByName(linkName)
		if err != nil {
			blog.Errorf("get netlink by name %s, err %s", linkName, err.Error())
			os.Exit(1)
		}
		err = netlink.LinkSetDown(eniLink)
		if err != nil {
			blog.Errorf("set down netlink with name %s, err %s", linkName, err.Error())
			os.Exit(1)
		}

		blog.Infof("detach eni %s ......", linkName)
		blog.Infof("get eni by ip %s", eniIPAddr)
		eniInstance, err := vpcClient.queryENIbyIP(eniIPAddr, *instance.InstanceId)
		if err != nil {
			blog.Errorf("get eni by ip failed, err %s", err.Error())
			os.Exit(1)
		}
		err = vpcClient.detachENI(*eniInstance.NetworkInterfaceId, *instance.InstanceId)
		if err != nil {
			blog.Errorf("detach eni %s, instance %s failed, err %s", *eniInstance.NetworkInterfaceId, *instance.InstanceId, err.Error())
			os.Exit(1)
		}
		err = vpcClient.waitForDetached(*eniInstance.NetworkInterfaceId, 5, 5)
		if err != nil {
			blog.Errorf("wait for eni %s detached from %s failed, err %s", *eniInstance.NetworkInterfaceId, *instance.InstanceId, err.Error())
			os.Exit(1)
		}

		blog.Infof("delete eni %s ......", *eniInstance.NetworkInterfaceId)
		err = vpcClient.deleteENI(*eniInstance.NetworkInterfaceId)
		if err != nil {
			blog.Errorf("delete eni %s failed, err %s", *eniInstance.NetworkInterfaceId, err.Error())
		}
	}
	ok := delRouteTables()
	if !ok {
		blog.Errorf("delete route tables failed")
	}

	blog.Infof("delete eni successfully")
}

// createVethPair create veth pair, return with cni format
func createVethPair(netns string, containerIfName string, mtu int) (*current.Interface, *current.Interface, error) {
	containerIface := &current.Interface{}
	hostIface := &current.Interface{}

	// create veth pair in container ns
	if err := ns.WithNetNSPath(netns, func(hostNS ns.NetNS) error {
		hostVeth, containerVeth, err := ip.SetupVeth(containerIfName, mtu, hostNS)
		if err != nil {
			return err
		}
		containerIface.Name = containerVeth.Name
		containerIface.Mac = containerVeth.HardwareAddr.String()
		containerIface.Sandbox = netns
		hostIface.Name = hostVeth.Name
		return nil
	}); err != nil {
		return nil, nil, err
	}

	hostVeth, err := netlink.LinkByName(hostIface.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup %q: %v", hostIface.Name, err)
	}
	hostIface.Mac = hostVeth.Attrs().HardwareAddr.String()
	return hostIface, containerIface, nil
}

// configureHostNS configure host namespace
func configureHostNS(hostIfName string, ipNet *net.IPNet, routeTableID, routeRulePriority int) error {

	// add to taskgroup route
	hostVeth, err := netlink.LinkByName(hostIfName)
	if err != nil {
		return fmt.Errorf("failed to look up %s in host ns, err %s", hostIfName, err.Error())
	}
	// add route in certain route table
	route := &netlink.Route{
		LinkIndex: hostVeth.Attrs().Index,
		Scope:     netlink.SCOPE_LINK,
		Dst:       ipNet,
		Table:     routeTableID,
	}
	err = netlink.RouteAdd(route)
	if err != nil {
		return fmt.Errorf("add route %s into host failed, err %s", route.String(), err.Error())
	}

	//add to taskgroup rule
	//**attention** do not usage &netlink.Rule{} for struct initialization
	ruleToTable := netlink.NewRule()
	ruleToTable.Dst = ipNet
	ruleToTable.Table = routeTableID
	ruleToTable.Priority = routeRulePriority - 1
	err = netlink.RuleDel(ruleToTable)
	if err != nil {
		blog.Warnf("clean old rule to table %s failed, err %s", ruleToTable.String(), err.Error())
	}
	err = netlink.RuleAdd(ruleToTable)
	if err != nil {
		return fmt.Errorf("add rule to table %s failed, err %s", ruleToTable.String(), err.Error())
	}

	//add from taskgroup rule
	ruleFromTaskgroup := netlink.NewRule()
	ruleFromTaskgroup.Src = ipNet
	ruleFromTaskgroup.Table = routeTableID
	ruleFromTaskgroup.Priority = routeRulePriority
	err = netlink.RuleDel(ruleFromTaskgroup)
	if err != nil {
		blog.Warnf("clean old rule from taskgroup %s failed, err %s", ruleToTable.String(), err.Error())
	}
	err = netlink.RuleAdd(ruleFromTaskgroup)
	if err != nil {
		return fmt.Errorf("add rule from taskgroup %s failed, err %s", ruleFromTaskgroup.String(), err.Error())
	}

	return nil
}

// configureContainerNS configure container namespace
// 1. set address for veth in container namespace
// 2. add routes
// 3. set static arp
func configureContainerNS(hostMac, netns, containerIfName string, ipNet *net.IPNet, gw net.IP) error {
	if err := ns.WithNetNSPath(netns, func(hostNS ns.NetNS) error {
		containerVeth, err := netlink.LinkByName(containerIfName)
		if err != nil {
			return fmt.Errorf("failed to look up %s in ns %s, err %s", containerIfName, netns, err.Error())
		}
		netlink.AddrAdd(containerVeth, &netlink.Addr{IPNet: ipNet})

		gwNet := &net.IPNet{IP: gw, Mask: net.CIDRMask(32, 32)}

		if err = netlink.RouteAdd(&netlink.Route{
			LinkIndex: containerVeth.Attrs().Index,
			Scope:     netlink.SCOPE_LINK,
			Dst:       gwNet,
		}); err != nil {
			return fmt.Errorf("add route to %v in ns %s failed, err %s", gwNet.String(), netns, err.Error())
		}

		defaultRoute := netlink.Route{
			LinkIndex: containerVeth.Attrs().Index,
			Dst:       &net.IPNet{IP: net.IPv4zero, Mask: net.CIDRMask(0, 32)},
			Scope:     netlink.SCOPE_UNIVERSE,
			Gw:        gw,
			Src:       ipNet.IP,
		}
		if err = netlink.RouteAdd(&defaultRoute); err != nil {
			return fmt.Errorf("add default route in ns %s failed, err %s", netns, err.Error())
		}

		hostHardwareAddr, err := net.ParseMAC(hostMac)
		if err != nil {
			return fmt.Errorf("parse mac from %s failed, err %s", hostMac, err.Error())
		}
		neigh := &netlink.Neigh{
			LinkIndex:    containerVeth.Attrs().Index,
			State:        netlink.NUD_PERMANENT,
			IP:           gwNet.IP,
			HardwareAddr: hostHardwareAddr,
		}

		if err = netlink.NeighAdd(neigh); err != nil {
			return fmt.Errorf("setup NS network: failed to add static ARP, err %s", err.Error())
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// getRouteTableIDByMac get route table id by mac address and eni prefix
func getRouteTableIDByMac(mac, eniPrefix string) (int, error) {
	links, err := netlink.LinkList()
	if err != nil {
		blog.Errorf("list links failed, err %s", err.Error())
		return -1, fmt.Errorf("list links failed, err %s", err.Error())
	}
	for _, l := range links {
		if strings.ToLower(l.Attrs().HardwareAddr.String()) == strings.ToLower(mac) {
			if !strings.HasPrefix(l.Attrs().Name, eniPrefix) {
				blog.Errorf("eni with mac %s does not has prefix %s", mac, eniPrefix)
				return -1, fmt.Errorf("eni with mac %s does not has prefix %s", mac, eniPrefix)
			}
			idString := strings.Trim(l.Attrs().Name, eniPrefix)
			id, err := strconv.Atoi(idString)
			if err != nil {
				blog.Errorf("convert %s to int failed, err %s", idString, err.Error())
				return -1, fmt.Errorf("convert %s to int failed, err %s", idString, err.Error())
			}
			return id + EniRouteTableStartID, nil
		}
	}
	return -1, fmt.Errorf("cannot find eni with mac %s", mac)
}

// CNIAdd implements cni interface
func (eip *EIP) CNIAdd(args *skel.CmdArgs) error {
	// load config from both stdin and environments variables
	netConf, cniVersion, err := conf.LoadConf(args.StdinData, args.Args)
	if err != nil {
		blog.Errorf("load config stdindata %s, args %s failed, err %s", string(args.StdinData), string(args.Args), err.Error())
		return fmt.Errorf("load config stdindata %s, args %s failed, err %s", string(args.StdinData), string(args.Args), err.Error())
	}
	// get container namespace
	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		blog.Errorf("failed to get netns %q: %s", netns, err.Error())
		return fmt.Errorf("failed to get netns %q: %s", netns, err.Error())
	}
	defer netns.Close()
	// get ip address from ipam plugin
	resultFromIPAM, err := ipam.ExecAdd(netConf.IPAM.Type, args.StdinData)
	if err != nil {
		return err
	}
	blog.Infof("get add result from ipam %s", resultFromIPAM.String())
	// if do CNIAdd failed, delete ip address
	defer func() {
		if err != nil {
			errDel := ipam.ExecDel(netConf.IPAM.Type, args.StdinData)
			if errDel != nil {
				blog.Errorf("del from ipam failed, err %s", errDel.Error())
			}
		}
	}()
	// parse result, get first ip
	result, err := current.NewResultFromResult(resultFromIPAM)
	if err != nil {
		return err
	}
	if len(result.IPs) == 0 {
		blog.Errorf("IPAM plugin %s returned missing IP config", result.String())
		return fmt.Errorf("IPAM plugin %s returned missing IP config", result.String())
	}
	if len(result.Interfaces) == 0 {
		blog.Errorf("IPAM plugin %s returned missing mac addr info", result.String())
		return fmt.Errorf("IPAM plugin %s returned missing mac addr info", result.String())
	}
	ipNet := &net.IPNet{
		IP:   result.IPs[0].Address.IP,
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}
	eniMac := result.Interfaces[0].Mac

	// find eni id according to eniMac
	routeTableID, err := getRouteTableIDByMac(eniMac, netConf.ENIPrefix)
	if err != nil {
		blog.Errorf("get route table id by mac %s with eni prefix %s failed, err %s", eniMac, netConf.ENIPrefix, err.Error())
		return fmt.Errorf("get route table id by mac %s with eni prefix %s failed, err %s", eniMac, netConf.ENIPrefix, err.Error())
	}

	hostVethInfo, containerVethInfo, err := createVethPair(netns.Path(), args.IfName, 1500)
	if err != nil {
		blog.Errorf("create veth pair failed, err %s", err.Error())
		return fmt.Errorf("create veth pair failed, err %s", err.Error())
	}
	blog.Infof("get hostVeth %v, containerVeth %v", hostVethInfo, containerVethInfo)

	err = configureContainerNS(hostVethInfo.Mac, netns.Path(), args.IfName, ipNet, result.IPs[0].Gateway)
	if err != nil {
		blog.Errorf("configure container ns network failed, err %s", err.Error())
		return fmt.Errorf("configure container ns network failed, err %s", err.Error())
	}

	err = configureHostNS(hostVethInfo.Name, ipNet, routeTableID, netConf.RouteRulePriority)
	if err != nil {
		blog.Errorf("configure host ns network failed, err %s", err.Error())
		return fmt.Errorf("configure host ns network failed, err %s", err.Error())
	}

	contIndex := 1
	ips := []*current.IPConfig{
		{
			Version:   "4",
			Address:   *ipNet,
			Interface: &contIndex,
		},
	}

	result = &current.Result{
		IPs:        ips,
		Interfaces: []*current.Interface{hostVethInfo, containerVethInfo},
	}

	return types.PrintResult(result, cniVersion)
}

// CNIDel implements cni interface
// 1. release ip address
// 2. clean container namespace
func (eip *EIP) CNIDel(args *skel.CmdArgs) error {

	blog.Infof("received cni del command: containerid %s, netns %s, ifname %s, args %s, path %s argsStdinData %s",
		args.ContainerID, args.Netns, args.IfName, args.Args, args.Path, args.StdinData)
	netConf, _, err := conf.LoadConf(args.StdinData, args.Args)
	if err != nil {
		blog.Infof("load config file failed, err %s", err.Error())
		return fmt.Errorf("load config file failed, err %s", err.Error())
	}

	err = ipam.ExecDel(netConf.IPAM.Type, args.StdinData)
	if err != nil {
		blog.Errorf("call IPAM delete function failed, err %s", err.Error())
		return fmt.Errorf("call IPAM delete function failed, err %s", err.Error())
	}

	var addrsToRelease []netlink.Addr
	var hostVethIndex int
	err = ns.WithNetNSPath(args.Netns, func(netNS ns.NetNS) error {
		link, err := netlink.LinkByName(args.IfName)
		if err != nil {
			if _, ok := err.(netlink.LinkNotFoundError); ok {
				blog.Infof("link %s not found in %v, no need to delete", args.IfName, netNS)
				return nil
			}
			blog.Errorf("get link by name %s in ns %s failed, err %s", args.IfName, args.Netns, err.Error())
			return fmt.Errorf("get link by name %s in ns %s failed, err %s", args.IfName, args.Netns, err.Error())
		}
		veth, ok := link.(*netlink.Veth)
		if !ok {
			blog.Errorf("link %s is not veth peer, failed", veth.Name)
			return fmt.Errorf("link %s is not veth peer, failed", veth.Name)
		}
		hostVethIndex, err = netlink.VethPeerIndex(veth)
		if err != nil {
			blog.Errorf("failed to get host veth peer index, err %s", err.Error())
			return fmt.Errorf("failed to get host veth peer index, err %s", err.Error())
		}
		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			blog.Errorf("get link %s addresses in ns %s failed, err %s", args.IfName, args.Netns, err.Error())
			return fmt.Errorf("get link %s addresses in ns %s failed, err %s", args.IfName, args.Netns, err.Error())
		}
		if len(addrs) == 0 {
			blog.Errorf("get link %s zero addresses in ns %s", args.IfName, args.Netns)
			return fmt.Errorf("get link %s zero addresses in ns %s", args.IfName, args.Netns)
		}
		addrsToRelease = addrs

		// shut down container veth
		_, err = ip.DelLinkByNameAddr(args.IfName, netlink.FAMILY_ALL)
		if err != nil && errors.Is(err, ip.ErrLinkNotFound) {
			blog.Errorf("delete link %s failed, err %s", args.IfName, err.Error())
			return nil
		}

		return nil
	})
	if err != nil {
		if _, ok := err.(ns.NSPathNotExistErr); ok {
			blog.Infof("ns %s is not exist", args.Netns)
		} else {
			blog.Errorf("tear ns %s failed, err %s", args.Netns, err.Error())
			return fmt.Errorf("tear ns %s failed, err %s", args.Netns, err.Error())
		}
	}

	if hostVethIndex != 0 {
		// try to delete link in host network namespace
		hostLink, err := netlink.LinkByIndex(hostVethIndex)
		if err != nil {
			if _, ok := err.(netlink.LinkNotFoundError); ok {
				blog.Infof("net link with index %d already be deleted, err %s", hostVethIndex, err.Error())
			} else {
				blog.Errorf("get netlink by index %v failed, err %s", hostVethIndex, err.Error())
				return fmt.Errorf("get netlink by index %v failed, err %s", hostVethIndex, err.Error())
			}
		} else {
			blog.Infof("delete net link %s in host ns", hostLink.Attrs().Name)
			err := netlink.LinkDel(hostLink)
			if err != nil {
				blog.Errorf("failed to delete net link %s in host ns, err %s", hostLink.Attrs().Name, err.Error())
				return fmt.Errorf("failed to delete net link %s in host ns, err %s", hostLink.Attrs().Name, err.Error())
			}
		}
	}

	// delete rule about pod
	for _, addr := range addrsToRelease {
		blog.Infof("delete addr %s route", addr.IPNet.Network())
		ipNet := &net.IPNet{
			IP:   addr.IPNet.IP,
			Mask: net.IPv4Mask(255, 255, 255, 255),
		}
		toTaskgroupRule := netlink.NewRule()
		toTaskgroupRule.Dst = ipNet
		err := netlink.RuleDel(toTaskgroupRule)
		if err != nil {
			blog.Warnf("delete to taskgroup rule %s failed, err %s", toTaskgroupRule.String(), err.Error())
		}
		fromTaskgroupRule := netlink.NewRule()
		fromTaskgroupRule.Src = ipNet
		err = netlink.RuleDel(fromTaskgroupRule)
		if err != nil {
			blog.Warnf("delete from taskgroup rule %s failed, err %s", fromTaskgroupRule.String(), err.Error())
		}
		blog.Infof("delete rules about %s complete", addr)
	}
	return nil
}
