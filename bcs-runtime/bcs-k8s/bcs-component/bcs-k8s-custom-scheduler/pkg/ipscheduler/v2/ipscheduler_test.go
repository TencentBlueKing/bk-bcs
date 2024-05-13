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

package v2

import (
	"strconv"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	core "k8s.io/client-go/testing"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"

	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned/fake"
	informers "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/informers/externalversions"
)

var (
	alwaysReady = func() bool { return true }
)

const (
	// testCniAnnotationKey bkbcs CNI plugin annotation key
	testCniAnnotationKey = "tke.cloud.tencent.com/networks"
	// testFixedIpAnnotationKey bkbcs fixed ip request annotation key
	testFixedIpAnnotationKey = "eni.cloud.bkbcs.tencent.com"
	// testCniAnnotationValue CNI plugin annotation value
	testCniAnnotationValue = "bcs-eni-cni"
	// testFixedIpAnnotationValue fixed ip request annotation value
	testFixedIpAnnotationValue = "fixed"
)

type fixture struct {
	t testing.TB

	client *fake.Clientset
	// Objects to put in the store.
	nodeNetworkLister []*cloudv1.NodeNetwork
	cloudIPLister     []*cloudv1.CloudIP

	// Actions expected to happen on the client. Objects from here are also
	// preloaded into NewSimpleFake.
	actions []core.Action
	objects []runtime.Object
}

func newFixture(t testing.TB) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	return f
}

func (f *fixture) newIpScheduler() *IpScheduler {
	ipScheduler := &IpScheduler{
		CniAnnotationKey:     testCniAnnotationKey,
		FixedIpAnnotationKey: testFixedIpAnnotationKey,
	}

	client := fake.NewSimpleClientset([]runtime.Object{}...)
	factory := informers.NewSharedInformerFactory(client, 0)
	nodeNetworkInformer := factory.Cloud().V1().NodeNetworks()
	ipScheduler.NodeNetworkLister = nodeNetworkInformer.Lister()
	cloudIpInformer := factory.Cloud().V1().CloudIPs()
	ipScheduler.CloudIpLister = cloudIpInformer.Lister()

	for _, n := range f.nodeNetworkLister {
		factory.Cloud().V1().NodeNetworks().Informer().GetIndexer().Add(n)
	}
	for _, c := range f.cloudIPLister {
		factory.Cloud().V1().CloudIPs().Informer().GetIndexer().Add(c)
	}

	return ipScheduler
}

// TestPredicateNonUnderlayPod test whether a pod will be scheduled when it doesn't need eni ip
func TestPredicateNonUnderlayPod(t *testing.T) {
	f := newFixture(t)

	nodeNetwork1 := newNodeNetwork("127.0.0.1", 2)
	f.nodeNetworkLister = append(f.nodeNetworkLister, nodeNetwork1)
	cloudIp1 := newCloudIp("8.8.8.8", newEniID("127.0.0.1", 0), "127.0.0.1")
	f.cloudIPLister = append(f.cloudIPLister, cloudIp1)

	DefaultIpScheduler = f.newIpScheduler()

	pod1 := newPod("pod1")
	node := newNode("127.0.0.1")

	extenderArgs := schedulerapi.ExtenderArgs{
		Pod: pod1,
		Nodes: &corev1.NodeList{
			Items: []corev1.Node{
				node,
			},
		},
	}

	filterResult, err := HandleIpSchedulerPredicate(extenderArgs)
	if err != nil {
		t.Fatalf("err when test the HandleIpSchedulerPredicate func: %s", err.Error())
	}

	if len(filterResult.Nodes.Items) != 1 && filterResult.Nodes.Items[0].Name != "127.0.0.1" {
		t.Error("expected the node can be scheduable when the pod doesn't need underlay IP")
	}
}

// TestPredicateScheduable ensures a pod can be scheduled to this node which has available eni ip
func TestPredicateScheduable(t *testing.T) {
	f := newFixture(t)

	nodeNetwork1 := newNodeNetwork("127.0.0.1", 2)
	f.nodeNetworkLister = append(f.nodeNetworkLister, nodeNetwork1)
	cloudIp1 := newCloudIp("8.8.8.8", newEniID("127.0.0.1", 0), "127.0.0.1")
	f.cloudIPLister = append(f.cloudIPLister, cloudIp1)

	DefaultIpScheduler = f.newIpScheduler()

	pod1 := newPod("pod1")
	pod1.Annotations[testCniAnnotationKey] = testCniAnnotationValue
	node := newNode("127.0.0.1")

	extenderArgs := schedulerapi.ExtenderArgs{
		Pod: pod1,
		Nodes: &corev1.NodeList{
			Items: []corev1.Node{
				node,
			},
		},
	}

	filterResult, err := HandleIpSchedulerPredicate(extenderArgs)
	if err != nil {
		t.Fatalf("err when test the HandleIpSchedulerPredicate func: %s", err.Error())
	}

	if len(filterResult.Nodes.Items) != 1 && filterResult.Nodes.Items[0].Name != "127.0.0.1" {
		t.Error("expected the node can be scheduable when the node has available underlay IP")
	}
}

// TestPredicateUnscheduable ensures a pod can't be scheduable to this node which has no available ip
func TestPredicateUnscheduable(t *testing.T) {
	f := newFixture(t)

	nodeNetwork1 := newNodeNetwork("127.0.0.1", 2)
	f.nodeNetworkLister = append(f.nodeNetworkLister, nodeNetwork1)
	cloudIp1 := newCloudIp("1.1.1.1", newEniID("127.0.0.1", 0), "127.0.0.1")
	f.cloudIPLister = append(f.cloudIPLister, cloudIp1)
	cloudIp2 := newCloudIp("0.0.0.0", newEniID("127.0.0.1", 0), "127.0.0.1")
	f.cloudIPLister = append(f.cloudIPLister, cloudIp2)

	DefaultIpScheduler = f.newIpScheduler()

	pod1 := newPod("pod1")
	pod1.Annotations[testCniAnnotationKey] = testCniAnnotationValue
	node := newNode("127.0.0.1")

	extenderArgs := schedulerapi.ExtenderArgs{
		Pod: pod1,
		Nodes: &corev1.NodeList{
			Items: []corev1.Node{
				node,
			},
		},
	}

	filterResult, err := HandleIpSchedulerPredicate(extenderArgs)
	if err != nil {
		t.Fatalf("err when test the HandleIpSchedulerPredicate func: %s", err.Error())
	}

	if len(filterResult.Nodes.Items) != 0 {
		t.Error("expected the node can not be scheduable when the node has no available underlay IP")
	}
	if len(filterResult.FailedNodes) != 1 || filterResult.FailedNodes["127.0.0.1"] != "no available eni ip anymore" {
		t.Error("expected the node can not be scheduable when the node has no available underlay IP")
	}
}

// TestPredicateFixedIP1 ensure a pod need fixed ip can be scheduled to a node which already exist a CloudIP
// and their SubnetID are the same
func TestPredicateFixedIP1(t *testing.T) {
	f := newFixture(t)

	nodeNetwork1 := newNodeNetwork("127.0.0.1", 1)
	nodeNetwork1.Status.Enis[0].EniSubnetID = "12345"
	f.nodeNetworkLister = append(f.nodeNetworkLister, nodeNetwork1)

	cloudIp := newCloudIp("1.1.1.1", newEniID("127.0.0.1", 0), "127.0.0.1")
	cloudIp.Labels[IPLabelKeyForIsFixed] = strconv.FormatBool(true)
	cloudIp.Spec.PodName = "pod1"
	cloudIp.Spec.SubnetID = "12345"
	cloudIp.Spec.IsFixed = true
	cloudIp.Spec.Host = "127.0.0.1"
	f.cloudIPLister = append(f.cloudIPLister, cloudIp)

	DefaultIpScheduler = f.newIpScheduler()

	pod1 := newPod("pod1")
	pod1.Annotations[testCniAnnotationKey] = testCniAnnotationValue
	pod1.Annotations[testFixedIpAnnotationKey] = testFixedIpAnnotationValue
	node := newNode("127.0.0.1")

	extenderArgs := schedulerapi.ExtenderArgs{
		Pod: pod1,
		Nodes: &corev1.NodeList{
			Items: []corev1.Node{
				node,
			},
		},
	}

	filterResult, err := HandleIpSchedulerPredicate(extenderArgs)
	if err != nil {
		t.Fatalf("err when test the HandleIpSchedulerPredicate func: %s", err.Error())
	}

	if len(filterResult.Nodes.Items) != 1 && filterResult.Nodes.Items[0].Name != "127.0.0.1" {
		t.Error(
			"expected the node can be scheduled when the pod need fixed ip, and already exist a CloudIP on this node")
	}
}

// TestPredicateFixedIP2 ensures a pod need fixed ip can be scheduled to a node when already exist an CloudIP,
// and this CloudIP has the same SubnetID to this node, and the node has available IP
func TestPredicateFixedIP2(t *testing.T) {
	f := newFixture(t)

	nodeNetwork1 := newNodeNetwork("127.0.0.1", 1)
	nodeNetwork1.Status.FloatingIPEni.Eni.EniSubnetID = "12345"
	f.nodeNetworkLister = append(f.nodeNetworkLister, nodeNetwork1)

	cloudIp := newCloudIp("1.1.1.1", newEniID("127.0.0.1", 0), "127.0.0.2")
	cloudIp.Labels[IPLabelKeyForIsFixed] = strconv.FormatBool(true)
	cloudIp.Spec.PodName = "pod1"
	cloudIp.Spec.SubnetID = "12345"
	cloudIp.Spec.IsFixed = true
	cloudIp.Spec.Host = "127.0.0.2"
	f.cloudIPLister = append(f.cloudIPLister, cloudIp)

	DefaultIpScheduler = f.newIpScheduler()

	pod1 := newPod("pod1")
	pod1.Annotations[testCniAnnotationKey] = testCniAnnotationValue
	pod1.Annotations[testFixedIpAnnotationKey] = testFixedIpAnnotationValue
	node := newNode("127.0.0.1")

	extenderArgs := schedulerapi.ExtenderArgs{
		Pod: pod1,
		Nodes: &corev1.NodeList{
			Items: []corev1.Node{
				node,
			},
		},
	}

	filterResult, err := HandleIpSchedulerPredicate(extenderArgs)
	if err != nil {
		t.Fatalf("err when test the HandleIpSchedulerPredicate func: %s", err.Error())
	}

	if len(filterResult.Nodes.Items) != 1 && filterResult.Nodes.Items[0].Name != "127.0.0.1" {
		t.Error("expected the node can be scheduled when the pod need fixed ip," +
			"and this node's SubnetId match, and has available IP")
	}
}

// TestPredicateFixedIP3 ensures a pod need fixed ip can't be scheduled to a node when already exist an CloudIP,
// and this CloudIP has the same SubnetID to this node, but the node has no available IP
func TestPredicateFixedIP3(t *testing.T) {
	f := newFixture(t)

	nodeNetwork1 := newNodeNetwork("127.0.0.1", 1)
	nodeNetwork1.Status.FloatingIPEni.Eni.EniSubnetID = "12345"
	f.nodeNetworkLister = append(f.nodeNetworkLister, nodeNetwork1)

	cloudIp := newCloudIp("1.1.1.1", newEniID("127.0.0.1", 0), "127.0.0.2")
	cloudIp.Labels[IPLabelKeyForIsFixed] = strconv.FormatBool(true)
	cloudIp.Spec.PodName = "pod1"
	cloudIp.Spec.SubnetID = "12345"
	cloudIp.Spec.IsFixed = true
	cloudIp.Spec.Host = "127.0.0.2"
	f.cloudIPLister = append(f.cloudIPLister, cloudIp)

	cloudIp1 := newCloudIp("1.1.1.2", newEniID("127.0.0.1", 0), "127.0.0.1")
	f.cloudIPLister = append(f.cloudIPLister, cloudIp1)

	DefaultIpScheduler = f.newIpScheduler()

	pod1 := newPod("pod1")
	pod1.Annotations[testCniAnnotationKey] = testCniAnnotationValue
	pod1.Annotations[testFixedIpAnnotationKey] = testFixedIpAnnotationValue
	node := newNode("127.0.0.1")

	extenderArgs := schedulerapi.ExtenderArgs{
		Pod: pod1,
		Nodes: &corev1.NodeList{
			Items: []corev1.Node{
				node,
			},
		},
	}

	filterResult, err := HandleIpSchedulerPredicate(extenderArgs)
	if err != nil {
		t.Fatalf("err when test the HandleIpSchedulerPredicate func: %s", err.Error())
	}

	if len(filterResult.Nodes.Items) != 0 {
		t.Error("expected the node can not be scheduled when pod need fixed ip, and this node has no available ip")
	}
	if len(filterResult.FailedNodes) != 1 || filterResult.FailedNodes["127.0.0.1"] != "no available eni ip anymore" {
		t.Error("expected the node can not be scheduled when pod need fixed ip, and this node has no available ip")
	}
}

// TestPredicateFixedIP4 ensures a pod need fixed ip can't be scheduled to a node when already exist an CloudIP,
// but this CloudIP's SubnetID is different to the node
func TestPredicateFixedIP4(t *testing.T) {
	f := newFixture(t)

	nodeNetwork1 := newNodeNetwork("127.0.0.1", 1)
	nodeNetwork1.Status.FloatingIPEni.Eni.EniSubnetID = "12345"
	f.nodeNetworkLister = append(f.nodeNetworkLister, nodeNetwork1)

	cloudIp := newCloudIp("1.1.1.1", newEniID("127.0.0.1", 0), "127.0.0.2")
	cloudIp.Labels[IPLabelKeyForIsFixed] = strconv.FormatBool(true)
	cloudIp.Spec.PodName = "pod1"
	cloudIp.Spec.SubnetID = "123"
	cloudIp.Spec.IsFixed = true
	cloudIp.Spec.Host = "127.0.0.2"
	f.cloudIPLister = append(f.cloudIPLister, cloudIp)

	DefaultIpScheduler = f.newIpScheduler()

	pod1 := newPod("pod1")
	pod1.Annotations[testCniAnnotationKey] = testCniAnnotationValue
	pod1.Annotations[testFixedIpAnnotationKey] = testFixedIpAnnotationValue
	node := newNode("127.0.0.1")

	extenderArgs := schedulerapi.ExtenderArgs{
		Pod: pod1,
		Nodes: &corev1.NodeList{
			Items: []corev1.Node{
				node,
			},
		},
	}

	filterResult, err := HandleIpSchedulerPredicate(extenderArgs)
	if err != nil {
		t.Fatalf("err when test the HandleIpSchedulerPredicate func: %s", err.Error())
	}

	if len(filterResult.Nodes.Items) != 0 {
		t.Error("expected the node can not be scheduable when pod need fixed ip, and SubnetId not matched")
	}
	if len(filterResult.FailedNodes) != 1 ||
		!strings.HasPrefix(filterResult.FailedNodes["127.0.0.1"], "subnetId unmatched for fixed ip") {
		t.Error("expected the node can not be scheduable when pod need fixed ip, and SubnetId not matched")
	}
}

func newPod(name string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
		ObjectMeta: metav1.ObjectMeta{
			UID:         uuid.NewUUID(),
			Name:        name,
			Namespace:   metav1.NamespaceDefault,
			Annotations: make(map[string]string),
		},
		Spec: corev1.PodSpec{},
	}
}

func newNode(ip string) corev1.Node {
	return corev1.Node{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Node"},
		ObjectMeta: metav1.ObjectMeta{
			UID:         uuid.NewUUID(),
			Name:        ip,
			Annotations: make(map[string]string),
		},
		Spec: corev1.NodeSpec{},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    "InternalIP",
					Address: ip,
				},
			},
		},
	}
}

func newEniID(nodeName string, index int) string {
	return "eni-" + nodeName + strconv.Itoa(index)
}

func newNodeNetwork(name string, ipLimit int) *cloudv1.NodeNetwork {
	n := cloudv1.NodeNetwork{
		TypeMeta: metav1.TypeMeta{APIVersion: "cloud.bkbcs.tencent.com/v1", Kind: "CloudNetwork"},
		ObjectMeta: metav1.ObjectMeta{
			UID:         uuid.NewUUID(),
			Name:        name,
			Namespace:   BcsSystem,
			Annotations: make(map[string]string),
		},
		Spec: cloudv1.NodeNetworkSpec{
			Hostname:    name,
			ENINum:      3,
			IPNumPerENI: ipLimit,
		},
		Status: cloudv1.NodeNetworkStatus{
			Enis: []*cloudv1.ElasticNetworkInterface{
				{
					EniID:  newEniID(name, 0),
					Status: "Ready",
				},
			},
		},
	}

	return &n
}

func newCloudIp(name, eniID, nodeIp string) *cloudv1.CloudIP {
	c := cloudv1.CloudIP{
		TypeMeta: metav1.TypeMeta{APIVersion: "cloud.bkbcs.tencent.com/v1", Kind: "CloudIP"},
		ObjectMeta: metav1.ObjectMeta{
			UID:         uuid.NewUUID(),
			Name:        name,
			Namespace:   metav1.NamespaceDefault,
			Annotations: make(map[string]string),
			Labels: map[string]string{
				IPLabelKeyForHost: nodeIp,
				IPLabelKeyForEni:  eniID,
			},
		},
		Spec: cloudv1.CloudIPSpec{
			Address: name,
		},
	}

	return &c
}
