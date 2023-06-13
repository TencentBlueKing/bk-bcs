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

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/errgroup"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sunstruct "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/listenercontroller"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

const (
	// 查询loadBalancer并发数限制
	concurrentLBGetLimit = 5
)

// IngressConverterOpt option of listener generator
type IngressConverterOpt struct {
	// DefaultRegion default cloud region for ingress converter
	DefaultRegion string
	// IsTCPUDPPortReuse if true, allow tcp listener and udp listener use same port
	IsTCPUDPPortReuse bool
	// Cloud cloud mod, e.g. tencentcloud aws gcp
	Cloud string
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
	// cloud e.g. tencentcloud aws gcp
	cloud string

	listenerHelper *listenercontroller.ListenerHelper
}

// NewIngressConverter create ingress generator
func NewIngressConverter(opt *IngressConverterOpt,
	cli client.Client, ingressValidater cloud.Validater, lbClient cloud.LoadBalance,
	listenerHelper *listenercontroller.ListenerHelper, lbIDCache, lbNameCache *gocache.Cache) (*IngressConverter,
	error) {
	if opt == nil {
		return nil, fmt.Errorf("option cannot be empty")
	}
	return &IngressConverter{
		defaultRegion:     opt.DefaultRegion,
		isTCPUDPPortReuse: opt.IsTCPUDPPortReuse,
		cloud:             opt.Cloud,
		cli:               cli,
		ingressValidater:  ingressValidater,
		lbClient:          lbClient,
		// set cache expire time
		lbIDCache:      lbIDCache,
		lbNameCache:    lbNameCache,
		listenerHelper: listenerHelper,
	}, nil
}

// get cloud loadbalance info by cloud loadbalance id pair
// regionIDPair "ap-xxxxx:lb-xxxxxx" "arn:aws:elasticloadbalancing:xxx:xxx:xxx"
func (g *IngressConverter) getLoadBalancerByID(ns, regionIDPair, protocolLayer string) (*cloud.LoadBalanceObject, error) {
	var lbObj *cloud.LoadBalanceObject
	var err error
	strs := g.splitRegionIDPair(regionIDPair)
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
			lbObj, err = g.lbClient.DescribeLoadBalancerWithNs(ns, g.defaultRegion, strs[0], "", protocolLayer)
		} else {
			lbObj, err = g.lbClient.DescribeLoadBalancer(g.defaultRegion, strs[0], "", protocolLayer)
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
			lbObj, err = g.lbClient.DescribeLoadBalancerWithNs(ns, strs[0], strs[1], "", protocolLayer)
		} else {
			lbObj, err = g.lbClient.DescribeLoadBalancer(strs[0], strs[1], "", protocolLayer)
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

func (g *IngressConverter) getLoadBalancerByIDWrapper(ns, regionIDPair, protocolLayer string,
	lbCh chan *cloud.LoadBalanceObject) func() error {
	return func() error {
		lb, err := g.getLoadBalancerByID(ns, regionIDPair, protocolLayer)
		if err != nil {
			return err
		}
		lbCh <- lb
		return nil
	}
}

// split regionIDPair
// regionIDPair "ap-xxxxx:lb-xxxxxx" "arn:aws:elasticloadbalancing:xxx:xxx:xxx"
// if regionIDPair has region and id, return 2 string
// else return 1 string
func (g *IngressConverter) splitRegionIDPair(regionIDPair string) []string {
	if a, err := arn.Parse(regionIDPair); err == nil {
		return []string{a.Region, regionIDPair}
	}
	return strings.Split(regionIDPair, ":")
}

// get cloud loadbalance info by cloud loadbalance name pair
// regionNamePair "ap-xxxxx:lbname"
func (g *IngressConverter) getLoadBalancerByName(ns, regionNamePair, protocolLayer string) (*cloud.LoadBalanceObject, error) {
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
			lbObj, err = g.lbClient.DescribeLoadBalancerWithNs(ns, g.defaultRegion, "", strs[0], protocolLayer)
		} else {
			lbObj, err = g.lbClient.DescribeLoadBalancer(g.defaultRegion, "", strs[0], protocolLayer)
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
			lbObj, err = g.lbClient.DescribeLoadBalancerWithNs(ns, strs[0], "", strs[1], protocolLayer)
		} else {
			lbObj, err = g.lbClient.DescribeLoadBalancer(strs[0], "", strs[1], protocolLayer)
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

func (g *IngressConverter) getLoadBalancerByNameWrapper(ns, regionNamePair, protocolLayer string,
	lbCh chan *cloud.LoadBalanceObject) func() error {
	return func() error {
		lb, err := g.getLoadBalancerByName(ns, regionNamePair, protocolLayer)
		if err != nil {
			return err
		}
		lbCh <- lb
		return nil
	}
}

// GetIngressLoadBalancers get ingress loadBalancer objects by annotations
func (g *IngressConverter) GetIngressLoadBalancers(ingress *networkextensionv1.Ingress) (
	[]*cloud.LoadBalanceObject, error) {
	protocolLayer := common.GetIngressProtocolLayer(ingress)
	var lbs []*cloud.LoadBalanceObject
	var lbCh chan *cloud.LoadBalanceObject
	workGroup := &errgroup.Group{}
	workGroup.SetLimit(concurrentLBGetLimit)
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
			if !MatchLbStrWithID(g.cloud, regionIDPair) {
				// invalid format
				blog.Warnf("lbid %s invalid", regionIDPair)
				return nil, fmt.Errorf("lbid %s invalid", regionIDPair)
			}
		}
		lbCh = make(chan *cloud.LoadBalanceObject, len(lbIDs))
		for _, regionIDPair := range lbIDs {
			workGroup.Go(g.getLoadBalancerByIDWrapper(ingress.GetNamespace(), regionIDPair, protocolLayer, lbCh))
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
		lbCh = make(chan *cloud.LoadBalanceObject, len(names))
		for _, regionNamePair := range names {
			workGroup.Go(g.getLoadBalancerByNameWrapper(ingress.GetNamespace(), regionNamePair, protocolLayer, lbCh))
		}
	}

	err := workGroup.Wait()
	close(lbCh)
	if err != nil {
		return nil, err
	}
	for lb := range lbCh {
		lbs = append(lbs, lb)
	}
	return lbs, nil
}

// ProcessUpdateIngress process newly added or updated ingress, return warnings([]string) and error
func (g *IngressConverter) ProcessUpdateIngress(ingress *networkextensionv1.Ingress) ([]string, error) {
	var warnings []string
	warnings = append(warnings, g.CheckIngressServiceAvailable(ingress)...)

	lbObjs, err := g.GetIngressLoadBalancers(ingress)
	if err != nil {
		return warnings, err
	}

	var generatedListeners []networkextensionv1.Listener
	var generatedSegListeners []networkextensionv1.Listener
	for i, rule := range ingress.Spec.Rules {
		ruleConverter := NewRuleConverter(g.cli, lbObjs, ingress.GetName(), ingress.GetNamespace(), &rule)
		ruleConverter.SetNamespaced(g.lbClient.IsNamespaced())
		ruleConverter.SetTCPUDPPortReuse(g.isTCPUDPPortReuse)
		listeners, inErr := ruleConverter.DoConvert()
		if inErr != nil {
			blog.Errorf("convert rule[%d] failed, err %s", i, inErr.Error())
			return warnings, fmt.Errorf("convert rule %d failed, err %s", i, inErr.Error())
		}
		generatedListeners = append(generatedListeners, listeners...)
	}
	for i, mapping := range ingress.Spec.PortMappings {
		mappingConverter := NewMappingConverter(g.cli, lbObjs, ingress.GetName(), ingress.GetNamespace(), &mapping)
		mappingConverter.SetNamespaced(g.lbClient.IsNamespaced())
		listeners, inErr := mappingConverter.DoConvert()
		if inErr != nil {
			blog.Errorf("convert mapping %d failed, err %s", i, inErr.Error())
			return warnings, fmt.Errorf("convert mapping %d failed, err %s", i, inErr.Error())
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
		return warnings, err
	}
	existedSegListeners, err := g.getSegmentListeners(ingress.GetName(), ingress.GetNamespace())
	if err != nil {
		return warnings, err
	}
	err = g.syncListeners(ingress.GetName(), ingress.GetNamespace(),
		existedListeners, generatedListeners, existedSegListeners, generatedSegListeners)
	if err != nil {
		blog.Errorf("syncListeners listener of ingress %s/%s failed, err %s",
			ingress.GetName(), ingress.GetNamespace(), err.Error())
		return warnings, fmt.Errorf("syncListeners listener ingress %s/%s failed, err %s",
			ingress.GetName(), ingress.GetNamespace(), err.Error())
	}
	if err = g.patchIngressStatus(ingress, lbObjs); err != nil {
		blog.Errorf("update ingress vips failed, err %s", err.Error())
		return warnings, fmt.Errorf("update ingress vips failed, err %s", err.Error())
	}
	return warnings, nil
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
			DNSName:          lb.DNSName,
			Scheme:           lb.Scheme,
			AWSLBType:        lb.AWSLBType,
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
func (g *IngressConverter) ProcessDeleteIngress(ingressName, ingressNamespace string) (bool, error) {
	var listenerList, segListenerList []networkextensionv1.Listener
	var err error
	// get existed listeners
	listenerList, err = g.getListeners(ingressName, ingressNamespace)
	if err != nil {
		return true, fmt.Errorf("get listeners of ingress %s/%s failed, err %s", ingressName, ingressNamespace,
			err.Error())
	}
	segListenerList, err = g.getSegmentListeners(ingressName, ingressNamespace)
	if err != nil {
		return true, fmt.Errorf("get segment listeners of ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
	}
	if len(listenerList) == 0 && len(segListenerList) == 0 {
		blog.Infof("listeners of ingress %s/%s, ingress can be deleted", ingressName, ingressNamespace)
		return false, nil
	}

	g.listenerHelper.SetDeleteListeners(append(listenerList, segListenerList...))

	return true, nil
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

// CheckIngressServiceAvailable ingress service is unavailable if len([]string) != 0
func (g *IngressConverter) CheckIngressServiceAvailable(ingress *networkextensionv1.Ingress) []string {
	// use set to avoid repeat message
	msgSet := make(map[string]struct{})
	for i, rule := range ingress.Spec.Rules {
		if len(rule.Services) == 0 {
			msgSet[fmt.Sprintf(constant.ValidateMsgEmptySvc, i+1)] = struct{}{}
			continue
		}
		for _, service := range rule.Services {
			svc := &k8scorev1.Service{}
			err := g.cli.Get(context.TODO(), k8stypes.NamespacedName{Namespace: service.ServiceNamespace,
				Name: service.ServiceName}, svc)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					msgSet[fmt.Sprintf(constant.ValidateMsgNotFoundSvc, i+1, service.ServiceNamespace,
						service.ServiceName)] = struct{}{}
				} else {
					blog.Errorf("k8s get resource failed, err: %+v", err)
					msgSet[fmt.Sprintf(constant.ValidateMsgUnknownErr, err)] = struct{}{}
				}
			}
		}
	}

	for i, portMapping := range ingress.Spec.PortMappings {
		if portMapping.WorkloadKind == "" || portMapping.WorkloadName == "" || portMapping.WorkloadNamespace == "" {
			msgSet[fmt.Sprintf(constant.ValidateMsgInvalidWorkload, i+1)] = struct{}{}
			continue
		}

		switch portMapping.WorkloadKind {
		case "GameStatefulSet":
			gsts := &k8sunstruct.Unstructured{}
			gsts.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "tkex.tencent.com",
				Version: "v1alpha1",
				Kind:    "GameStatefulSet",
			})
			err := g.cli.Get(context.TODO(), k8stypes.NamespacedName{
				Namespace: portMapping.WorkloadNamespace,
				Name:      portMapping.WorkloadName,
			}, gsts)

			if err != nil {
				if k8serrors.IsNotFound(err) {
					msgSet[fmt.Sprintf(constant.ValidateMsgEmptyWorkload, i+1)] = struct{}{}
				} else {
					blog.Errorf("k8s get resource failed, err: %+v", err)
					msgSet[fmt.Sprintf(constant.ValidateMsgUnknownErr, err)] = struct{}{}
				}
			}
		case "StatefulSet":
			sts := &k8sappsv1.StatefulSet{}
			err := g.cli.Get(context.TODO(), k8stypes.NamespacedName{
				Namespace: portMapping.WorkloadNamespace,
				Name:      portMapping.WorkloadName,
			}, sts)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					msgSet[fmt.Sprintf(constant.ValidateMsgEmptyWorkload, i+1)] = struct{}{}
				} else {
					blog.Errorf("k8s get resource failed, err: %+v", err)
					msgSet[fmt.Sprintf(constant.ValidateMsgUnknownErr, err)] = struct{}{}
				}
			}
		default:
			msgSet[fmt.Sprintf("port mapping[%d] has invalid workload kind", i+1)] = struct{}{}
		}
	}

	var msgList []string
	for msg := range msgSet {
		msgList = append(msgList, msg)
	}

	return msgList
}
