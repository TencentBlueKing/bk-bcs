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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	pluginmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin xxx
type Plugin struct {
	opt     *Options
	dnsLock sync.Mutex
	ready   bool
	Detail  Detail
	pluginmanager.NodePlugin
}

// Detail xxx
type Detail struct {
}

var (
	netAvailabilityLabels = []string{"node", "targetnode", "status"}
	netAvailability       = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "net_availability",
		Help: "net_availability, 1 means OK",
	}, netAvailabilityLabels)
	clusterApiserverCertificateExpirationLabels = []string{"type"}
	clusterApiserverCertificateExpiration       = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterApiserverCertExpirationMetricName,
		Help: ClusterApiserverCertExpirationMetricName,
	}, clusterApiserverCertificateExpirationLabels)
)

func init() {
	metricmanager.Register(netAvailability)
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

	if p.opt.CheckCert {
		metricmanager.Register(clusterApiserverCertificateExpiration)
	}

	// run as daemon
	if runMode == pluginmanager.RunModeDaemon {
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
				case <-time.After(time.Duration(p.opt.Interval) * time.Second):
					continue
				}
			}
		}()
	} else if runMode == pluginmanager.RunModeOnce {
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
	result := make([]pluginmanager.CheckItem, 0, 0)
	nodeconfig := pluginmanager.Pm.GetConfig().NodeConfig
	nodeName := nodeconfig.NodeName

	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("netcheck failed: %s, stack: %v\n", r, string(debug.Stack()))
		}
		p.Result = pluginmanager.CheckResult{
			Items: result,
		}

		if !p.ready {
			p.ready = true
		}
	}()

	gaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
	defer func() {
		metricmanager.RefreshMetric(netAvailability, gaugeVecSetList)
	}()

	cidr := nodeconfig.Node.Spec.PodCIDR
	for key, val := range nodeconfig.Node.Annotations {
		if key == "tke.cloud.tencent.com/pod-cidrs" {
			cidr = val
			break
		}
	}

	// 检测网卡配置
	if cidr != "" {
		devStatus, err := CheckDevIP(cidr)
		if err != nil {
			klog.Errorf(err.Error())
			result = append(result, pluginmanager.CheckItem{
				ItemName:   pluginName,
				ItemTarget: nodeName,
				Level:      pluginmanager.RISKLevel,
				Normal:     false,
				Detail:     fmt.Sprintf("check interface failed: %s", err.Error()),
				Status:     devStatus,
			})

			gaugeVecSetList = append(gaugeVecSetList, &metricmanager.GaugeVecSet{
				Labels: []string{nodeconfig.NodeName, nodeconfig.NodeName, devStatus},
				Value:  float64(1),
			})
			return
		}
	}

	// 检查节点的容器网络
	// checkitem上报是否有pod ping不通
	checkItem := pluginmanager.CheckItem{
		ItemName:   pluginName,
		ItemTarget: nodeName,
		Level:      pluginmanager.RISKLevel,
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
	gaugeVecSetList = append(gaugeVecSetList, &metricmanager.GaugeVecSet{
		Labels: []string{nodeconfig.NodeName, nodeconfig.NodeName, status},
		Value:  float64(1),
	})

	// 检查apiserver证书
	if p.opt.CheckCert {
		checkItemList, gvsList, err := getApiserverCert(nodeconfig.KubernetesSvc)
		if err != nil {
			klog.Errorf("check apiserver cert expiration failed: %s", err.Error())
		} else {
			result = append(result, checkItemList...)
			metricmanager.RefreshMetric(clusterApiserverCertificateExpiration, gvsList)
		}
	}
}

// CheckOverLay xxx
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

// PINGCheck xxx
func PINGCheck(ip string) string {
	pingCmd := exec.Command("ping", "-c1", "-W1", ip)
	output, err := pingCmd.CombinedOutput()
	if err != nil {
		klog.Error(string(output), err.Error())
		return PingFailedStatus
	}

	return NormalStatus
}

// GetIFAddr xxx
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

// CheckDevIP 目前只兼容cni0与flannel的场景
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

// GetLinkIp xxx
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

// getApiserverCert get apsierver cert expiration through api port
func getApiserverCert(svcIP string) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, error) {
	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	gvsList := make([]*metricmanager.GaugeVecSet, 0, 0)
	// 检查自签证书
	expiration, err := util.GetServerCert("apiserver-loopback-client", svcIP, "443")
	if err != nil {
		klog.Errorf("check apiserver self-signed cert expiration failed: %s", err.Error())
		return checkItemList, gvsList, err
	}

	checkItem := pluginmanager.CheckItem{
		ItemName:   ClusterApiserverCertExpirationCheckItem,
		ItemTarget: ApiserverTarget,
		Normal:     true,
		Status:     NormalStatus,
		Detail:     fmt.Sprintf(StringMap[AboutToExpireDetail], "self signed", expiration.Sub(time.Now())/time.Second),
		Level:      pluginmanager.WARNLevel,
		Tags:       nil,
	}

	// 时间在1周以内则返回异常
	if expiration.Sub(time.Now()) < 604800*time.Second {
		checkItem.Normal = false
		checkItem.Status = AboutToExpireStatus
		checkItem.SetDetail(fmt.Sprintf(StringMap[AboutToExpireDetail], "self signed", expiration.Sub(time.Now())/time.Second))
	}

	checkItemList = append(checkItemList, checkItem)

	gvsList = append(gvsList, &metricmanager.GaugeVecSet{
		Labels: []string{"self signed"},
		Value:  float64(expiration.Sub(time.Now()) / time.Second),
	})

	// 检查apiserver证书
	expiration, err = util.GetServerCert("kubernetes.default.svc.cluster.local", svcIP, "443")
	if err != nil {
		klog.Errorf("check apiserver cert expiration failed: %s", err.Error())
		return checkItemList, gvsList, err
	}

	checkItem = pluginmanager.CheckItem{
		ItemName:   ClusterApiserverCertExpirationCheckItem,
		ItemTarget: ApiserverTarget,
		Normal:     true,
		Status:     NormalStatus,
		Detail:     fmt.Sprintf(StringMap[AboutToExpireDetail], "apiserver", expiration.Sub(time.Now())/time.Second),
		Level:      pluginmanager.WARNLevel,
		Tags:       nil,
	}

	// 时间在1周以内则返回异常
	if expiration.Sub(time.Now()) < 604800*time.Second {
		checkItem.Normal = false
		checkItem.Status = AboutToExpireStatus
		checkItem.SetDetail(fmt.Sprintf(StringMap[AboutToExpireDetail], "apiserver", expiration.Sub(time.Now())/time.Second))
	}

	checkItemList = append(checkItemList, checkItem)

	gvsList = append(gvsList, &metricmanager.GaugeVecSet{
		Labels: []string{"apiserver"},
		Value:  float64(expiration.Sub(time.Now()) / time.Second),
	})

	return checkItemList, gvsList, err
}

// Ready xxx
func (p *Plugin) Ready(string) bool {
	return p.ready
}

// GetResult xxx
func (p *Plugin) GetResult(string) pluginmanager.CheckResult {
	return p.Result
}

// Execute xxx
func (p *Plugin) Execute() {
	p.Check()
}

// GetDetail xxx
func (p *Plugin) GetDetail() interface{} {
	return p.Detail
}
