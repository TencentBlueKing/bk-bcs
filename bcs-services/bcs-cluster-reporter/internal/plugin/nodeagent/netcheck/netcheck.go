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

// Package netcheck xxx
package netcheck

import (
	"fmt"
	"net"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metric_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin_manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin xxx
type Plugin struct {
	opt     *Options
	dnsLock sync.Mutex
	ready   bool
	Detail  Detail
	plugin_manager.NodePlugin
}

type Detail struct {
}

var (
	netAvailabilityLabels = []string{"node", "targetnode", "status"}
	netAvailability       = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "net_availability",
		Help: "net_availability, 1 means OK",
	}, netAvailabilityLabels)
)

func init() {
	metric_manager.Register(netAvailability)
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	p.opt = &Options{}
	err := util.ReadorInitConf(configFilePath, p.opt, initContent)
	if err != nil {
		return err
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	p.StopChan = make(chan int)
	interval := p.opt.Interval
	if interval == 0 {
		interval = 60
	}

	// run as daemon
	if runMode == plugin_manager.RunModeDaemon {
		go func() {
			for {
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()
					go p.Check()
				} else {
					klog.Infof("the former %s didn't over, skip in this loop", p.Name())
				}
				select {
				case result := <-p.StopChan:
					klog.Infof("stop plugin %s by signal %d", p.Name(), result)
					return
				case <-time.After(time.Duration(interval) * time.Second):
					continue
				}
			}
		}()
	} else if runMode == plugin_manager.RunModeOnce {
		p.Check()
	}

	return nil
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.CheckLock.Lock()
	p.StopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	p.CheckLock.Unlock()
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return "netcheck"
}

// Check xxx
func (p *Plugin) Check() {
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		p.CheckLock.Unlock()
	}()
	p.ready = false
	result := make([]plugin_manager.CheckItem, 0, 0)
	nodeconfig := plugin_manager.Pm.GetConfig().NodeConfig
	nodeName := nodeconfig.NodeName

	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("netcheck failed: %s, stack: %v\n", r, string(debug.Stack()))
		}
		p.Result = plugin_manager.CheckResult{
			Items: result,
		}

		if !p.ready {
			p.ready = true
		}
	}()

	gaugeVecSetList := make([]*metric_manager.GaugeVecSet, 0, 0)
	defer func() {
		metric_manager.RefreshMetric(netAvailability, gaugeVecSetList)
	}()

	cidr := nodeconfig.Node.Spec.PodCIDR
	for key, val := range nodeconfig.Node.Annotations {
		if key == "tke.cloud.tencent.com/pod-cidrs" {
			cidr = val
			break
		}
	}
	// 检测网卡配置
	devStatus, err := CheckDevIP(cidr)
	if err != nil {
		klog.Errorf(err.Error())
		result = append(result, plugin_manager.CheckItem{
			ItemName:   pluginName,
			ItemTarget: nodeName,
			Level:      plugin_manager.RISKLevel,
			Normal:     false,
			Detail:     fmt.Sprintf("check interface failed: %s", err.Error()),
			Status:     devStatus,
		})

		gaugeVecSetList = append(gaugeVecSetList, &metric_manager.GaugeVecSet{
			Labels: []string{nodeconfig.NodeName, nodeconfig.NodeName, devStatus},
			Value:  float64(1),
		})
		return
	}

	// 检查节点的容器网络
	// checkitem上报是否有pod ping不通
	checkItem := plugin_manager.CheckItem{
		ItemName:   pluginName,
		ItemTarget: nodeName,
		Level:      plugin_manager.RISKLevel,
		Normal:     true,
		Detail:     "ping dns pod success",
		Status:     NormalStatus,
	}

	status, err := CheckOverLay(nodeconfig.ClientSet)
	if err != nil {
		checkItem.Normal = false
		checkItem.Detail = err.Error()
		checkItem.Status = status
		klog.Errorf(err.Error())
	}

	result = append(result, checkItem)
	gaugeVecSetList = append(gaugeVecSetList, &metric_manager.GaugeVecSet{
		Labels: []string{nodeconfig.NodeName, nodeconfig.NodeName, status},
		Value:  float64(1),
	})
}

func CheckOverLay(clientSet *kubernetes.Clientset) (string, error) {
	//测试访问dns pod是否OK
	ipList := make([]string, 0, 0)
	ep, err := clientSet.CoreV1().Endpoints("kube-system").Get(util.GetCtx(10*time.Second), "kube-dns", v1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		return errorStatus, err
	}

	for _, subset := range ep.Subsets {
		for _, address := range subset.Addresses {
			ipList = append(ipList, address.IP)
		}
	}

	for _, ip := range ipList {
		pingStatus := PINGCheck(ip)
		if pingStatus != NormalStatus {
			return pingStatus, fmt.Errorf("ping failed: %s", ip)
		}
	}

	klog.Infof("olveray ping %v success", ipList)

	return NormalStatus, nil
}

func PINGCheck(ip string) (status string) {
	pingCmd := exec.Command("ping", "-c1", "-W1", ip)
	output, err := pingCmd.CombinedOutput()
	if err != nil {
		status = PingFailedStatus
		klog.Error(string(output), err.Error())
		return
	}

	status = NormalStatus
	return
}

func GetIFAddr(ifName string) (net.IP, error) {
	sock, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = syscall.Close(sock)
		if err != nil {
			klog.Infof("close sock failed: %s ", err.Error())
		}
	}()

	ifreq, err := unix.NewIfreq(ifName)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(sock), uintptr(unix.SIOCGIFADDR), uintptr(unsafe.Pointer(ifreq)))
	if errno != 0 {
		return nil, errno
	}

	addr, err := ifreq.Inet4Addr()
	if err != nil {
		return nil, err
	}

	ip := net.IPv4(addr[0], addr[1], addr[2], addr[3])
	return ip, err
}

// 目前只兼容cni0与flannel的场景
func CheckDevIP(cidr string) (string, error) {
	_, subnet, _ := net.ParseCIDR(cidr)

	ip, _, linkName, err := GetLinkIp("bridge")
	if err != nil {
		if err.Error() == "not found" {
			return NormalStatus, nil
		}
		return errorStatus, err
	}

	if !subnet.Contains(ip) {
		return devDistinctStatus, fmt.Errorf("node cidr is %s, bridge %s is %s", cidr, linkName, ip)
	} else {
		klog.Infof("check netinterface success, cidr: %s, bridge %s: %s", cidr, linkName, ip)
	}

	vxLanIp, _, linkName, err := GetLinkIp("vxlan")
	if err != nil {
		if err.Error() == "not found" {
			return NormalStatus, nil
		}
		return errorStatus, err
	}

	if !subnet.Contains(vxLanIp) {
		return devDistinctStatus, fmt.Errorf("node cidr is %s, vxlan %s is %s", cidr, linkName, vxLanIp)
	} else {
		klog.Infof("check netinterface success, cidr: %s, vxlanIP %s: %s", cidr, linkName, vxLanIp)
	}

	return NormalStatus, nil
}

func GetLinkIp(deviceType string) (net.IP, net.IPMask, string, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, nil, "", err
	}

	for _, link := range links {
		if strings.Contains(link.Attrs().Name, "docker") {
			continue
		}
		if link.Type() == deviceType {

			addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
			if err != nil {
				return nil, nil, link.Attrs().Name, err
			}
			for _, addr := range addrs {
				return addr.IP, addr.Mask, link.Attrs().Name, nil
			}
		}
	}

	return nil, nil, "", fmt.Errorf("not found")
}

func (p *Plugin) Ready(string) bool {
	return p.ready
}

func (p *Plugin) GetResult(string) plugin_manager.CheckResult {
	return p.Result
}

func (p *Plugin) Execute() {
	p.Check()
}

func (p *Plugin) GetDetail() interface{} {
	return p.Detail
}

func (p *Plugin) GetString(key string) string {
	return StringMap[key]
}
