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

package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	gocache "github.com/patrickmn/go-cache"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// IngressConverterOpt option of listener generator
type IngressConverterOpt struct {
	// DefaultRegion default cloud region for ingress converter
	DefaultRegion string
	// IsTCPUDPPortReuse if true, allow tcp listener and udp listener use same port
	IsTCPUDPPortReuse bool
}

// IngressConverter listener generator
type IngressConverter struct {
	// default cloud region for ingress converter
	defaultRegion string
	// crd client
	cli client.Client
	// ingress validater
	ingressValidater cloud.Validater
	// cloud interface for controlling cloud loadbalance
	lbClient cloud.LoadBalance
	// cache for cloud loadbalance id
	lbIDCache *gocache.Cache
	// cache for cloud loadbalance name
	lbNameCache *gocache.Cache
	// if true, allow tcp listener and udp listener use same port
	isTCPUDPPortReuse bool
}

// NewIngressConverter create ingress generator
func NewIngressConverter(opt *IngressConverterOpt,
	cli client.Client, ingressValidater cloud.Validater, lbClient cloud.LoadBalance) (*IngressConverter, error) {
	if opt == nil {
		return nil, fmt.Errorf("option cannot be empty")
	}
	return &IngressConverter{
		defaultRegion:     opt.DefaultRegion,
		isTCPUDPPortReuse: opt.IsTCPUDPPortReuse,
		cli:               cli,
		ingressValidater:  ingressValidater,
		lbClient:          lbClient,
		// set cache expire time
		lbIDCache:   gocache.New(60*time.Minute, 120*time.Minute),
		lbNameCache: gocache.New(60*time.Minute, 120*time.Minute),
	}, nil
}

// get cloud loadbalance info by cloud loadbalance id pair
// regionIDPair "ap-xxxxx:lb-xxxxxx"
func (g *IngressConverter) getLoadbalanceByID(ns, regionIDPair string) (*cloud.LoadBalanceObject, error) {
	var lbObj *cloud.LoadBalanceObject
	var err error
	strs := strings.Split(regionIDPair, ":")
	// only has id
	if len(strs) == 1 {
		obj, ok := g.lbIDCache.Get(g.defaultRegion + ":" + strs[0])
		if ok {
			if lbObj, ok = obj.(*cloud.LoadBalanceObject); !ok {
				return nil, fmt.Errorf("get obj from lb id cache is not LoadBalanceObject")
			}
			return lbObj, nil
		}
		if g.lbClient.IsNamespaced() {
			lbObj, err = g.lbClient.DescribeLoadBalancerWithNs(ns, g.defaultRegion, strs[0], "")
		} else {
			lbObj, err = g.lbClient.DescribeLoadBalancer(g.defaultRegion, strs[0], "")
		}
		if err != nil {
			return nil, err
		}
	} else if len(strs) == 2 {
		// region and id
		obj, ok := g.lbIDCache.Get(regionIDPair)
		if ok {
			if lbObj, ok = obj.(*cloud.LoadBalanceObject); !ok {
				return nil, fmt.Errorf("get obj from lb id cache is not LoadBalanceObject")
			}
			return lbObj, nil
		}
		if g.lbClient.IsNamespaced() {
			lbObj, err = g.lbClient.DescribeLoadBalancerWithNs(ns, strs[0], strs[1], "")
		} else {
			lbObj, err = g.lbClient.DescribeLoadBalancer(strs[0], strs[1], "")
		}
		if err != nil {
			return nil, err
		}
	} else {
		// invalid format
		blog.Warnf("lbid %s invalid", regionIDPair)
		return nil, fmt.Errorf("lbid %s invalid", regionIDPair)
	}
	g.lbIDCache.SetDefault(lbObj.Region+":"+lbObj.LbID, lbObj)
	g.lbNameCache.SetDefault(lbObj.Region+":"+lbObj.Name, lbObj)
	return lbObj, nil
}

// get cloud loadbalance info by cloud loadbalance name pair
// regionNamePair "ap-xxxxx:lbname"
func (g *IngressConverter) getLoadbalanceByName(ns, regionNamePair string) (*cloud.LoadBalanceObject, error) {
	var lbObj *cloud.LoadBalanceObject
	var err error
	strs := strings.Split(regionNamePair, ":")
	// only has name
	if len(strs) == 1 {
		obj, ok := g.lbNameCache.Get(g.defaultRegion + ":" + strs[0])
		if ok {
			if lbObj, ok = obj.(*cloud.LoadBalanceObject); !ok {
				return nil, fmt.Errorf("get obj from lb name cache is not LoadBalanceObject")
			}
			return lbObj, nil
		}
		if g.lbClient.IsNamespaced() {
			lbObj, err = g.lbClient.DescribeLoadBalancerWithNs(ns, g.defaultRegion, "", strs[0])
		} else {
			lbObj, err = g.lbClient.DescribeLoadBalancer(g.defaultRegion, "", strs[0])
		}
		if err != nil {
			return nil, err
		}
	} else if len(strs) == 2 {
		// region and name
		obj, ok := g.lbNameCache.Get(regionNamePair)
		if ok {
			if lbObj, ok = obj.(*cloud.LoadBalanceObject); !ok {
				return nil, fmt.Errorf("get obj from lb id cache is not LoadBalanceObject")
			}
			return lbObj, nil
		}
		if g.lbClient.IsNamespaced() {
			lbObj, err = g.lbClient.DescribeLoadBalancerWithNs(ns, strs[0], "", strs[1])
		} else {
			lbObj, err = g.lbClient.DescribeLoadBalancer(strs[0], "", strs[1])
		}
		if err != nil {
			return nil, err
		}
	} else {
		// invalid format
		blog.Warnf("lbname %s invalid", regionNamePair)
		return nil, fmt.Errorf("lbname %s invalid", regionNamePair)
	}
	g.lbIDCache.SetDefault(lbObj.Region+":"+lbObj.LbID, lbObj)
	g.lbNameCache.SetDefault(lbObj.Region+":"+lbObj.Name, lbObj)
	return lbObj, nil
}

// get ingress loadbalance objects by annotations
func (g *IngressConverter) getIngressLoadbalances(ingress *networkextensionv1.Ingress) (
	[]*cloud.LoadBalanceObject, error) {
	var lbs []*cloud.LoadBalanceObject
	lbIDStrs, idOk := ingress.Annotations[networkextensionv1.AnnotationKeyForLoadbalanceIDs]
	lbNameStrs, nameOk := ingress.Annotations[networkextensionv1.AnnotationKeyForLoadbalanceNames]
	if !idOk && !nameOk {
		blog.Errorf("ingress %+v is not associated with lb instance", ingress)
		return nil, fmt.Errorf("ingress %+v is not associated with lb instance", ingress)
	}
	// check lb id first
	// if there are both ids and names, ids is effective
	if idOk {
		lbIDs := strings.Split(lbIDStrs, ",")
		// check lb id format before request cloud
		for _, regionIDPair := range lbIDs {
			if !MatchLbStrWithId(regionIDPair) {
				// invalid format
				blog.Warnf("lbid %s invalid", regionIDPair)
				return nil, fmt.Errorf("lbid %s invalid", regionIDPair)
			}
		}
		for _, regionIDPair := range lbIDs {
			lbObj, err := g.getLoadbalanceByID(ingress.GetNamespace(), regionIDPair)
			if err != nil {
				return nil, err
			}
			lbs = append(lbs, lbObj)
		}
	} else if nameOk {
		names := strings.Split(lbNameStrs, ",")
		// check lb name format before request cloud
		for _, regionNamePair := range names {
			if !MatchLbStrWithName(regionNamePair) {
				// invalid format
				blog.Warnf("lbname %s invalid", regionNamePair)
				return nil, fmt.Errorf("lbname %s invalid", regionNamePair)
			}
		}
		for _, regionNamePair := range names {
			lbObj, err := g.getLoadbalanceByName(ingress.GetNamespace(), regionNamePair)
			if err != nil {
				return nil, err
			}
			lbs = append(lbs, lbObj)
		}
	}
	return lbs, nil
}

// ProcessUpdateIngress process newly added or updated ingress
func (g *IngressConverter) ProcessUpdateIngress(ingress *networkextensionv1.Ingress) error {
	isValid, errMsg := g.ingressValidater.IsIngressValid(ingress)
	if !isValid {
		blog.Errorf("ingress %+v ingress is invalid, err %s", ingress, errMsg)
		return fmt.Errorf("ingress %+v ingress is invalid, err %s", ingress, errMsg)
	}

	isValid, errMsg = g.ingressValidater.CheckNoConflictsInIngress(ingress)
	if !isValid {
		blog.Errorf("ingress %+v ingress has conflicts, err %s", ingress, errMsg)
		return fmt.Errorf("ingress %+v ingress has conflicts, err %s", ingress, errMsg)
	}

	lbObjs, err := g.getIngressLoadbalances(ingress)
	if err != nil {
		return err
	}

	for _, lbObj := range lbObjs {
		isConflict, inErr := g.checkConflicts(lbObj.LbID, ingress)
		if inErr != nil {
			return inErr
		}
		if isConflict {
			blog.Errorf("ingress %+v is conflict with existed listeners", ingress)
			return fmt.Errorf("ingress %+v is conflict with existed listeners", ingress)
		}
	}

	var generatedListeners []networkextensionv1.Listener
	var generatedSegListeners []networkextensionv1.Listener
	for _, rule := range ingress.Spec.Rules {
		ruleConverter := NewRuleConverter(g.cli, lbObjs, ingress.GetName(), ingress.GetNamespace(), &rule)
		ruleConverter.SetNamespaced(g.lbClient.IsNamespaced())
		ruleConverter.SetTCPUDPPortReuse(g.isTCPUDPPortReuse)
		listeners, inErr := ruleConverter.DoConvert()
		if inErr != nil {
			blog.Errorf("convert rule %+v failed, err %s", rule, inErr.Error())
			return fmt.Errorf("convert rule %+v failed, err %s", rule, inErr.Error())
		}
		generatedListeners = append(generatedListeners, listeners...)
	}
	for _, mapping := range ingress.Spec.PortMappings {
		mappingConverter := NewMappingConverter(g.cli, lbObjs, ingress.GetName(), ingress.GetNamespace(), &mapping)
		mappingConverter.SetNamespaced(g.lbClient.IsNamespaced())
		listeners, inErr := mappingConverter.DoConvert()
		if inErr != nil {
			blog.Errorf("convert mapping %+v failed, err %s", mapping, inErr.Error())
			return fmt.Errorf("convert mapping %+v failed, err %s", mapping, inErr.Error())
		}
		// if ignore segment, disable segment feature;
		// if segment length is not set or equals to 1, disable segment feature;
		if mapping.IgnoreSegment || mapping.SegmentLength == 0 || mapping.SegmentLength == 1 {
			generatedListeners = append(generatedListeners, listeners...)
		} else {
			generatedSegListeners = append(generatedSegListeners, listeners...)
		}
	}
	sort.Sort(networkextensionv1.ListenerSlice(generatedListeners))
	sort.Sort(networkextensionv1.ListenerSlice(generatedSegListeners))

	existedListeners, err := g.getListeners(ingress.GetName(), ingress.GetNamespace())
	if err != nil {
		return err
	}
	existedSegListeners, err := g.getSegmentListeners(ingress.GetName(), ingress.GetNamespace())
	if err != nil {
		return err
	}
	err = g.syncListeners(ingress.GetName(), ingress.GetNamespace(),
		existedListeners, generatedListeners, existedSegListeners, generatedSegListeners)
	if err != nil {
		blog.Errorf("syncListeners listener of ingress %s/%s failed, err %s",
			ingress.GetName(), ingress.GetNamespace(), err.Error())
		return fmt.Errorf("syncListeners listener ingress %s/%s failed, err %s",
			ingress.GetName(), ingress.GetNamespace(), err.Error())
	}
	if err = g.patchIngressStatus(ingress, lbObjs); err != nil {
		blog.Errorf("update ingress vips failed, err %s", err.Error())
		return fmt.Errorf("update ingress vips failed, err %s", err.Error())
	}
	return nil
}

// update ingress loadbalancers fields
func (g *IngressConverter) patchIngressStatus(ingress *networkextensionv1.Ingress,
	lbs []*cloud.LoadBalanceObject) error {

	newStatus := networkextensionv1.IngressStatus{}
	for _, lb := range lbs {
		newStatus.Loadbalancers = append(newStatus.Loadbalancers, networkextensionv1.IngressLoadBalancer{
			LoadbalancerName: lb.Name,
			LoadbalancerID:   lb.LbID,
			Region:           lb.Region,
			Type:             lb.Type,
			IPs:              lb.IPs,
		})
	}
	patchStruct := map[string]interface{}{
		"status": newStatus,
	}
	newStatusBytes, err := json.Marshal(patchStruct)
	if err != nil {
		return fmt.Errorf("encoding ingress status to json bytes failed, err %s", err.Error())
	}

	rawPatch := client.RawPatch(k8stypes.MergePatchType, newStatusBytes)
	updatedIngress := &networkextensionv1.Ingress{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      ingress.GetName(),
			Namespace: ingress.GetNamespace(),
		},
	}
	err = g.cli.Patch(context.TODO(), updatedIngress, rawPatch, &client.PatchOptions{})
	if err != nil {
		return fmt.Errorf("pactch ingress %s/%s status to k8s apiserver failed, err %s",
			ingress.GetName(), ingress.GetNamespace(), err.Error())
	}
	return nil
}

// ProcessDeleteIngress  process deleted ingress
func (g *IngressConverter) ProcessDeleteIngress(ingressName, ingressNamespace string) error {
	var listenerList, segListenerList []networkextensionv1.Listener
	var err error
	// get existed listeners
	listenerList, err = g.getListeners(ingressName, ingressNamespace)
	if err != nil {
		return fmt.Errorf("get listeners of ingress %s/%s failed, err %s", ingressName, ingressNamespace, err.Error())
	}
	segListenerList, err = g.getSegmentListeners(ingressName, ingressNamespace)
	if err != nil {
		return fmt.Errorf("get segment listeners of ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
	}
	if len(listenerList) == 0 && len(segListenerList) == 0 {
		blog.Infof("listeners of ingress %s/%s, ingress can be deleted", ingressName, ingressNamespace)
		return nil
	}
	// delete listeners
	if err = g.deleteListeners(ingressName, ingressNamespace); err != nil {
		return fmt.Errorf("delete listeners of ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
	}
	return fmt.Errorf("wait listeners of ingress %s/%s to be deleted", ingressName, ingressNamespace)
}

func (g *IngressConverter) deleteListeners(ingressName, ingressNamespace string) error {
	listener := &networkextensionv1.Listener{}
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
		ingressName: networkextensionv1.LabelValueForIngressName,
	})))
	if err != nil {
		blog.Errorf("get selector for deleted ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
		return fmt.Errorf("get selector for deleted ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
	}
	err = g.cli.DeleteAllOf(context.TODO(), listener,
		&client.DeleteAllOfOptions{
			ListOptions: client.ListOptions{
				LabelSelector: selector,
				Namespace:     ingressNamespace,
			},
		})
	if err != nil {
		blog.Errorf("delete listener by label selector %s, err %s", selector.String(), err.Error())
		return fmt.Errorf("delete listener by label selector %s, err %s", selector.String(), err.Error())
	}
	return nil
}

func (g *IngressConverter) getListeners(ingressName, ingressNamespace string) (
	[]networkextensionv1.Listener, error) {
	existedListenerList := &networkextensionv1.ListenerList{}
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
		ingressName: networkextensionv1.LabelValueForIngressName,
		networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
	})))
	err = g.cli.List(context.TODO(), existedListenerList, &client.ListOptions{
		Namespace:     ingressNamespace,
		LabelSelector: selector})
	if err != nil {
		blog.Errorf("list listeners ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
		return nil, fmt.Errorf("list listeners ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
	}
	return existedListenerList.Items, nil
}

func (g *IngressConverter) getSegmentListeners(ingressName, ingressNamespace string) (
	[]networkextensionv1.Listener, error) {
	existedListenerList := &networkextensionv1.ListenerList{}
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
		ingressName: networkextensionv1.LabelValueForIngressName,
		networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueTrue,
	})))
	err = g.cli.List(context.TODO(), existedListenerList, &client.ListOptions{
		Namespace:     ingressNamespace,
		LabelSelector: selector})
	if err != nil {
		blog.Errorf("list segment listeners ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
		return nil, fmt.Errorf("list segment listeners ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
	}
	return existedListenerList.Items, nil
}

func (g *IngressConverter) syncListeners(ingressName, ingressNamespace string,
	existedListeners, listeners []networkextensionv1.Listener,
	existedSegListeners, segListeners []networkextensionv1.Listener) error {

	adds, dels, olds, news := GetDiffListeners(existedListeners, listeners)
	sadds, sdels, solds, snews := GetDiffListeners(existedSegListeners, segListeners)
	adds = append(adds, sadds...)
	dels = append(dels, sdels...)
	olds = append(olds, solds...)
	news = append(news, snews...)
	for _, del := range dels {
		blog.V(3).Infof("[generator] delete listener %s/%s", del.GetNamespace(), del.GetName())
		err := g.cli.Delete(context.TODO(), &del, &client.DeleteOptions{})
		if err != nil {
			blog.Errorf("delete listener %+v failed, err %s", del, err.Error())
			return fmt.Errorf("delete listener %+v failed, err %s", del, err.Error())
		}
	}
	for _, add := range adds {
		blog.V(3).Infof("[generator] create listener %s/%s", add.GetNamespace(), add.GetName())
		err := g.cli.Create(context.TODO(), &add, &client.CreateOptions{})
		if err != nil {
			blog.Errorf("create listener %+v failed, err %s", add, err.Error())
			return fmt.Errorf("create listener %+v failed, err %s", add, err.Error())
		}
	}
	for index, new := range news {
		blog.V(3).Infof("[generator] update listener %s/%s", new.GetNamespace(), new.GetName())
		new.ResourceVersion = olds[index].ResourceVersion
		err := g.cli.Update(context.TODO(), &new, &client.UpdateOptions{})
		if err != nil {
			blog.Errorf("update listener %+v failed, err %s", new, err.Error())
			return fmt.Errorf("update listener %+v failed, err %s", new, err.Error())
		}
	}
	return nil
}
