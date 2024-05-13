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

package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	netsvc "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/cloud-network-agent/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/eni"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/netservice"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/networkutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/nodenetwork"
	cloud "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"
)

const (
	eventTypeAdd    = "add"
	eventTypeDel    = "del"
	eventTypeUpdate = "update"

	statusSuccess = "success"
	statusFailed  = "failed"
)

var (
	// ControllerEventCounter node network event counter for network controller
	ControllerEventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cloudnetwork_agent",
			Subsystem: "controller",
			Name:      "event_counter",
			Help:      "controller node network event counter",
		},
		[]string{"event_type"})

	// ControllerReconcilNetworkCounter counter for reconcile network
	ControllerReconcilNetworkCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cloudnetwork_agent",
			Subsystem: "controller",
			Name:      "reconcile_counter",
			Help:      "counter for controller reconcile network",
		},
		[]string{"status"})
)

func init() {
	prometheus.MustRegister(ControllerEventCounter)
	prometheus.MustRegister(ControllerReconcilNetworkCounter)
}

func reportEventMetric(eventType string) {
	ControllerEventCounter.WithLabelValues(eventType).Inc()
}

func reportReconcileMetric(status string) {
	ControllerReconcilNetworkCounter.WithLabelValues(status).Inc()
}

// NetworkController controller for cloud network
type NetworkController struct {
	// eth name for identifying vm instance
	instanceEth string

	// vm instance hostname
	hostname string

	// extra elastic network interface number
	eniNum int

	// ip number for each extra elastic network interface
	ipNum int

	// options for network agent
	options *options.NetworkOption

	// infor for current vm node
	vmInfo *cloud.VMInfo

	// node network object for current node
	nodeNetwork     *cloud.NodeNetwork
	nodeNetworkMutx sync.Mutex

	// netservice client
	netsvcClient netservice.Interface

	// crd client
	nodeNetClient nodenetwork.Interface

	// eni cloud client
	eniClient eni.Interface

	// network util client
	netUtil networkutil.Interface

	// lock for controller
	// ensure that reconcle and release not do at the same time
	controllerLock sync.Mutex
}

// New create new network controller
func New(instanceEth, hostname string, op *options.NetworkOption,
	netsvcClient netservice.Interface, nodeNetClient nodenetwork.Interface,
	eniClient eni.Interface, netUtil networkutil.Interface) *NetworkController {
	return &NetworkController{
		instanceEth:   instanceEth,
		hostname:      hostname,
		options:       op,
		netsvcClient:  netsvcClient,
		nodeNetClient: nodeNetClient,
		eniClient:     eniClient,
		netUtil:       netUtil,
	}
}

// OnAdd add event
func (nc *NetworkController) OnAdd(obj interface{}) {
	nodeNetwork, ok := obj.(*cloud.NodeNetwork)
	if !ok {
		blog.Errorf("obj %+v is not *cloud.NodeNetwork", obj)
		return
	}
	blog.V(3).Infof("new NodeNetwork added: %+v", nodeNetwork)
	reportEventMetric(eventTypeAdd)
}

// OnUpdate update event
// TODO: to deal with node network changes
func (nc *NetworkController) OnUpdate(oldObj, newObj interface{}) {
	_, okOld := oldObj.(*cloud.NodeNetwork)
	if !okOld {
		blog.Errorf("oldObj %+v is not *cloud.NodeNetwork", oldObj)
		return
	}
	_, okNew := newObj.(*cloud.NodeNetwork)
	if !okNew {
		blog.Errorf("newObj %+v is not *cloud.NodeNetwork", newObj)
		return
	}

	reportEventMetric(eventTypeUpdate)
}

// OnDelete delete event
func (nc *NetworkController) OnDelete(obj interface{}) {
	nodeNetwork, ok := obj.(*cloud.NodeNetwork)
	if !ok {
		blog.Errorf("get nodeNetwork %s/%s deleted", nodeNetwork.GetNamespace(), nodeNetwork.GetName())
		return
	}

	reportEventMetric(eventTypeDel)

	// when deleted node network was on current node, start release network
	if nodeNetwork.GetNamespace() == nc.nodeNetwork.GetNamespace() &&
		nodeNetwork.GetName() == nc.nodeNetwork.GetName() {
		go func() {
			if err := nc.releaseNodeNetwork(); err != nil {
				blog.Errorf("releaseNodeNetwork failed, err %s", err.Error())
			}
		}()
	}
}

// getNodeNetwork get node network config from etcd
func (nc *NetworkController) getNodeNetwork() error {
	nodeNetwork, err := nc.nodeNetClient.Get(constant.CRD_NAMESPACES, nc.hostname)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Warnf("node network not found in etcd")
			return nil
		}
		blog.Errorf("get node network failed, err %s", err.Error())
		return err
	}
	nc.nodeNetworkMutx.Lock()
	nc.nodeNetwork = nodeNetwork
	nc.nodeNetworkMutx.Unlock()
	return nil
}

func (nc *NetworkController) getEniQuota() error {
	// get limitation according to vm info
	eniNum, ipNum, err := nc.eniClient.GetENILimit()
	if err != nil {
		blog.Infof("get eni quota limit, err %s", err.Error())
	}
	eniNum = eniNum - 1
	ipNum = ipNum - 1

	// when request eni number is 0 or over limitation,
	// apply for as many network cards as possible
	if nc.options.EniNum == 0 || int(nc.options.EniNum) > eniNum {
		nc.eniNum = eniNum
	} else {
		nc.eniNum = int(nc.options.EniNum)
	}
	// when request ip number is 0 or over limitation,
	// apply for as many ip addresses as possible
	if nc.options.IPNumPerEni == 0 || int(nc.options.IPNumPerEni-1) > ipNum {
		nc.ipNum = ipNum
	} else {
		nc.ipNum = int(nc.options.IPNumPerEni)
	}
	return nil
}

// generate eni name by vm instance id and eni index
func getEniName(instanceid string, index int) string {
	return instanceid + "-" + constant.ENI_PREFIX + strconv.Itoa(index)
}

// get eni interface name
func getEniIfaceName(index int) string {
	return constant.ENI_PREFIX + strconv.Itoa(index)
}

func (nc *NetworkController) createEnis() error {
	// create new node network object
	newNode := &cloud.NodeNetwork{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: cloud.SchemeGroupVersion.Version,
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      nc.hostname,
			Namespace: constant.CRD_NAMESPACES,
		},
		Spec: cloud.NodeNetworkSpec{
			Cluster:     nc.options.Cluster,
			Hostname:    nc.hostname,
			NodeAddress: nc.vmInfo.InstanceIP,
			VM:          nc.vmInfo,
			ENINum:      nc.eniNum,
			IPNumPerENI: nc.ipNum,
		},
	}

	maxIndex, err := nc.eniClient.GetMaxENIIndex()
	if err != nil {
		blog.Errorf("get max eni index failed, err %s", err.Error())
		return err
	}
	blog.Infof("get current max index %d", maxIndex)

	for i := 0; i < nc.eniNum; i++ {
		eniName := getEniName(nc.vmInfo.InstanceID, i)

		// createENI
		newIf, err := nc.eniClient.CreateENI(eniName, nc.ipNum)
		if err != nil {
			blog.Errorf("create eni failed, err %s", err.Error())
			return err
		}
		blog.Infof("create eni %s done", eniName)

		if newIf.Attachment == nil {
			// attachENI
			attachment, err := nc.eniClient.AttachENI(
				maxIndex+i+1,
				newIf.EniID,
				nc.vmInfo.InstanceID,
				newIf.MacAddress)

			if err != nil {
				blog.Errorf("attach eni %s failed, err %s", newIf.EniID, err.Error())
				return err
			}
			newIf.Attachment = attachment
		}

		newIf.Index = i
		newIf.EniIfaceName = getEniIfaceName(i)
		newIf.RouteTableID = constant.START_ROUTE_TABLE + i
		blog.Infof("attach eni %s done", eniName)

		newNode.Status.Enis = append(newNode.Status.Enis, newIf)
	}

	if err := nc.createNodeNetwork(newNode); err != nil {
		return err
	}

	return nil
}

func (nc *NetworkController) deleteEnis() error {
	for _, eni := range nc.nodeNetwork.Status.Enis {
		if err := nc.eniClient.DetachENI(eni.Attachment); err != nil {
			blog.Errorf("detach eni by %+v failed, err %s", eni.Attachment, err.Error())
			return fmt.Errorf("detach eni by %+v, err %s", eni.Attachment, err.Error())
		}
		time.Sleep(2 * time.Second)
		blog.Infof("detach eni by %+v successfully", eni.Attachment)
		if err := nc.eniClient.DeleteENI(eni.EniID); err != nil {
			blog.Errorf("delete eni by eni id %s failed, err %s", eni.EniID, err.Error())
			return fmt.Errorf("delete eni by eni id %s failed, err %s", eni.EniID, err.Error())
		}
		blog.Infof("delete eni by eni id %s successfully", eni.EniID)
	}
	return nil
}

func (nc *NetworkController) deleteNodeNetwork() {
	nc.nodeNetworkMutx.Lock()
	nc.nodeNetwork = nil
	nc.nodeNetworkMutx.Unlock()
}

func (nc *NetworkController) createNodeNetwork(newNode *cloud.NodeNetwork) error {
	if err := nc.nodeNetClient.Create(newNode); err != nil {
		blog.Errorf("write node network %+v to apiserver failed, err %s", newNode, err.Error())
		return err
	}

	nc.nodeNetworkMutx.Lock()
	nc.nodeNetwork = newNode
	nc.nodeNetworkMutx.Unlock()
	return nil
}

func getSubnetMask(cidr string) (int, error) {
	strs := strings.Split(cidr, "/")
	if len(strs) == 2 {
		mask, err := strconv.Atoi(strs[1])
		if err != nil {
			return -1, fmt.Errorf("convert mask %s of cidr %s to int failed, err %s",
				strs[1], cidr, err.Error())
		}
		return mask, nil
	}
	return -1, fmt.Errorf("invalid cidr %s", cidr)
}

func (nc *NetworkController) createNetservicePool() error {
	if len(nc.nodeNetwork.Status.Enis) == 0 {
		return fmt.Errorf("no eni in node network, cannot create netservice pool")
	}

	subnetCidr := nc.nodeNetwork.Status.Enis[0].EniSubnetCidr
	mask, err := getSubnetMask(subnetCidr)
	if err != nil {
		blog.Errorf("get subnet mask of cidr %s", subnetCidr)
		return fmt.Errorf("get subnet mask of cidr %s", subnetCidr)
	}
	pool := new(netsvc.NetPool)
	pool.Net = nc.nodeNetwork.Spec.NodeAddress
	pool.Cluster = nc.options.Cluster
	pool.Mask = mask
	pool.Hosts = append(pool.Hosts, nc.nodeNetwork.Spec.NodeAddress)
	pool.Gateway = "169.254.1.1"
	var ipInstances []*netsvc.IPInst
	for _, eni := range nc.nodeNetwork.Status.Enis {
		for _, ip := range eni.SecondaryAddresses {
			if !ip.IsPrimary {
				pool.Available = append(pool.Available, ip.IP)
				ipIns := new(netsvc.IPInst)
				ipIns.IPAddr = ip.IP
				ipIns.MacAddr = eni.MacAddress
				ipIns.Pool = pool.Net
				ipIns.Mask = pool.Mask
				ipIns.Gateway = pool.Gateway
				ipIns.Cluster = pool.Cluster
				ipInstances = append(ipInstances, ipIns)
			}
		}
	}
	err = nc.netsvcClient.CreateOrUpdatePool(pool)
	if err != nil {
		blog.Errorf("create netservice pool %+v failed, err %s", pool, err.Error())
		return err
	}
	blog.Infof("create netservice pool %+v successfully", pool)
	// add newly ip address into ip pool in service
	for _, ipIns := range ipInstances {
		err := nc.netsvcClient.UpdateIPInstance(ipIns)
		if err != nil {
			blog.Errorf("update ip instance with %+v failed, err %s", ipIns, err.Error())
			return err
		}
		blog.Infof("update ip instance with %+v successfully", ipIns)
	}

	blog.Infof("create netservice pool done")
	return nil
}

func (nc *NetworkController) deleteNetservicePool() error {
	if len(nc.nodeNetwork.Status.Enis) == 0 {
		return fmt.Errorf("no eni in node network, cannot create netservice pool")
	}

	err := nc.netsvcClient.DeletePool(nc.options.Cluster, nc.nodeNetwork.Spec.NodeAddress)
	if err != nil {
		return fmt.Errorf("delete netservice pool failed, err %s", err.Error())
	}
	blog.Infof("delete netservice pool %s/%s successfully", nc.options.Cluster, nc.nodeNetwork.Spec.NodeAddress)

	return nil
}

// Init node init
func (nc *NetworkController) Init() error {

	if err := nc.getNodeNetwork(); err != nil {
		blog.Errorf("get node network failed, err %s", err.Error())
		return err
	}

	if err := nc.eniClient.Init(); err != nil {
		blog.Errorf("aws client init failed, err %s", err.Error())
		return err
	}

	vmInfo, err := nc.eniClient.GetVMInfo()
	if err != nil {
		blog.Errorf("get vm info failed, err %s", err.Error())
		return err
	}
	nc.vmInfo = vmInfo

	if err := nc.getEniQuota(); err != nil {
		blog.Errorf("get eni quota failed, err %s", err.Error())
		return err
	}

	// Determine if the node network exists
	nc.nodeNetworkMutx.Lock()
	if nc.nodeNetwork != nil {
		nc.nodeNetworkMutx.Unlock()
		blog.Infof("node network alrady exists %+v", nc.nodeNetwork)
		return nil
	}
	nc.nodeNetworkMutx.Unlock()

	if err := nc.createEnis(); err != nil {
		blog.Errorf("create enis failed, err %s", err.Error())
		return err
	}

	if err := nc.createNetservicePool(); err != nil {
		blog.Errorf("create netservice pool failed, err %s", err.Error())
		return err
	}

	return nil
}

func (nc *NetworkController) getRouteIDMap() map[string]string {
	ret := make(map[string]string)
	for _, eni := range nc.nodeNetwork.Status.Enis {
		ret[strconv.Itoa(eni.RouteTableID)] = constant.ENI_PREFIX + strconv.Itoa(eni.Index)
	}
	return ret
}

// Run run controller
func (nc *NetworkController) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := nc.reconcileNodeNetwork(); err != nil {
		reportReconcileMetric(statusFailed)
		blog.Infof("first reconcile node network failed, err %s", err.Error())
		return
	}
	reportReconcileMetric(statusSuccess)

	routeIDMap := nc.getRouteIDMap()
	if err := nc.netUtil.SetHostNetwork(nc.instanceEth, routeIDMap); err != nil {
		blog.Infof("set host network failed, err %s", err.Error())
		return
	}

	tick := time.NewTicker(time.Duration(nc.options.CheckInterval) * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Infof("it's time to check node network and enis!")

			if err := nc.reconcileNodeNetwork(); err != nil {
				reportReconcileMetric(statusFailed)
				blog.Warnf("reconcile node network failed, err %s", err.Error())
			} else {
				reportReconcileMetric(statusSuccess)
			}

		case <-ctx.Done():
			blog.Infof("stop controller...")
			return
		}
	}
}

// reconcileNodeNetwork restore network interface on vm
func (nc *NetworkController) reconcileNodeNetwork() error {
	nc.controllerLock.Lock()
	defer nc.controllerLock.Unlock()

	if nc.nodeNetwork == nil {
		blog.Errorf("no node network found")
		return fmt.Errorf("no node network found")
	}

	rules, err := nc.netUtil.RuleList()
	if err != nil {
		blog.Errorf("list rule failed, err %s", err.Error())
		return fmt.Errorf("list rule failed, err %s", err.Error())
	}

	for _, netiface := range nc.nodeNetwork.Status.Enis {
		err := nc.netUtil.SetUpNetworkInterface(
			netiface.Address.IP,
			netiface.EniSubnetCidr,
			netiface.MacAddress,
			netiface.EniIfaceName,
			netiface.RouteTableID,
			nc.options.EniMTU,
			rules,
		)
		if err != nil {
			blog.Errorf("sync network interface failed, err %s", err.Error())
		}
	}

	return nil
}

// releaseNodeNetwork release node network
func (nc *NetworkController) releaseNodeNetwork() error {
	nc.controllerLock.Lock()
	defer nc.controllerLock.Unlock()

	var err error

	// delete netservice pool
	if err = nc.deleteNetservicePool(); err != nil {
		blog.Errorf(
			"failed delete netservice pool when release node network, err %s",
			err.Error())

		blog.Warnf("try to restore node network to etcd")
		// try to restore node network to etcd
		if restoreErr := nc.createNodeNetwork(nc.nodeNetwork); restoreErr != nil {
			blog.Warnf("falied to restore node network")
		}

		return fmt.Errorf(
			"failed delete netservice pool when release node network, err %s",
			err.Error())
	}

	// set down eni
	rules, err := nc.netUtil.RuleList()
	if err != nil {
		blog.Errorf("list rule failed, err %s", err.Error())
		return fmt.Errorf("list rule failed, err %s", err.Error())
	}
	for _, netiface := range nc.nodeNetwork.Status.Enis {
		err := nc.netUtil.SetDownNetworkInterface(
			netiface.Address.IP,
			netiface.EniSubnetCidr,
			netiface.MacAddress,
			netiface.EniIfaceName,
			netiface.RouteTableID,
			rules,
		)
		if err != nil {
			blog.Errorf("set down network interface failed, err %s", err.Error())
		}
	}

	// delete enis
	if err := nc.deleteEnis(); err != nil {
		blog.Errorf("failed to delete enis, err %s", err.Error())
		return err
	}

	// delete node work network
	nc.deleteNodeNetwork()

	return nil
}
