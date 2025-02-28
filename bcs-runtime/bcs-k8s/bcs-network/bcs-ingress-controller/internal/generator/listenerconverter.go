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

package generator

import (
	"context"
	"fmt"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
)

// IngressListenerConverter convert listeners by ingress, do not operator listener directly
type IngressListenerConverter struct {
	// crd client
	Cli client.Client

	// if lb client is namespaced scope
	IsNamespaced bool
	// if true, allow tcp listener and udp listener use same port
	IsTCPUDPPortReuse bool
	// send event
	Eventer record.EventRecorder
}

// CheckIngressUpdateFinish return true if ingress update finished
func (c *IngressListenerConverter) CheckIngressUpdateFinish(ingress *networkextensionv1.Ingress) (bool, error) {
	lbObjs := make([]*cloud.LoadBalanceObject, 0, len(ingress.Status.Loadbalancers))
	for _, lb := range ingress.Status.Loadbalancers {
		lbObjs = append(lbObjs, &cloud.LoadBalanceObject{
			LbID:   lb.LoadbalancerID,
			Region: lb.Region,
		})
	}

	generateListeners, generateSegListeners, err := c.GenerateListeners(ingress,
		lbObjs)
	if err != nil {
		return false, fmt.Errorf("generate ingress '%s/%s'  listeners failed, err: %s",
			ingress.Namespace, ingress.Name, err.Error())
	}

	existedListeners, existedSegListeners, err := c.GetExistedListeners(ingress.GetName(),
		ingress.GetNamespace())
	if err != nil {
		return false, fmt.Errorf("get existed listeners from ingress '%s/%s' failed, err: %s",
			ingress.Namespace, ingress.Name, err.Error())
	}

	adds, dels, olds, news := GetDiffListeners(existedListeners, generateListeners)
	if len(adds) != 0 || len(dels) != 0 || len(olds) != 0 || len(news) != 0 {
		blog.Infof("ingress '%s/%s' update not finish, listeners diff: adds: %d, dels: %d, olds: %d, news: %d",
			ingress.Namespace, ingress.Name, len(adds), len(dels), len(olds), len(news))
		return false, nil
	}

	sadds, sdels, solds, snews := GetDiffListeners(existedSegListeners, generateSegListeners)
	if len(sadds) != 0 || len(sdels) != 0 || len(solds) != 0 || len(snews) != 0 {
		blog.Infof("ingress '%s/%s' update not finish, seg listeners diff: adds: %d, dels: %d, olds: %d, news: %d",
			ingress.Namespace, ingress.Name, len(sadds), len(sdels), len(solds), len(snews))
		return false, nil
	}

	for _, li := range existedListeners {
		if li.Status.Status != networkextensionv1.ListenerStatusSynced {
			return false, nil
		}
	}

	for _, li := range existedSegListeners {
		if li.Status.Status != networkextensionv1.ListenerStatusSynced {
			return false, nil
		}
	}

	return true, nil
}

// GenerateListeners generate listeners by ingress, lbObjs is got from ingress
func (c *IngressListenerConverter) GenerateListeners(ingress *networkextensionv1.Ingress,
	lbObjs []*cloud.LoadBalanceObject) ([]networkextensionv1.Listener, []networkextensionv1.Listener, error) {

	var generatedListeners []networkextensionv1.Listener
	var generatedSegListeners []networkextensionv1.Listener
	for i, rule := range ingress.Spec.Rules {
		ruleConverter := NewRuleConverter(c.Cli, lbObjs, ingress, ingress.GetName(), ingress.GetNamespace(), &rule,
			c.Eventer)
		ruleConverter.SetNamespaced(c.IsNamespaced)
		ruleConverter.SetTCPUDPPortReuse(c.IsTCPUDPPortReuse)
		listeners, inErr := ruleConverter.DoConvert()
		if inErr != nil {
			blog.Errorf("convert rule[%d] failed, err %s", i, inErr.Error())
			return nil, nil, fmt.Errorf("convert rule %d failed, err %s", i, inErr.Error())
		}
		generatedListeners = append(generatedListeners, listeners...)
	}
	for i, mapping := range ingress.Spec.PortMappings {
		mappingConverter := NewMappingConverter(c.Cli, lbObjs, ingress.GetName(), ingress.GetNamespace(), &mapping)
		mappingConverter.SetNamespaced(c.IsNamespaced)
		listeners, inErr := mappingConverter.DoConvert()
		if inErr != nil {
			blog.Errorf("convert mapping %d failed, err %s", i, inErr.Error())
			return nil, nil, fmt.Errorf("convert mapping %d failed, err %s", i, inErr.Error())
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

	return generatedListeners, generatedSegListeners, nil
}

// GetExistedListeners get related listeners of ingress, return listeners and segment listeners
func (c *IngressListenerConverter) GetExistedListeners(ingressName, ingressNs string) ([]networkextensionv1.Listener,
	[]networkextensionv1.Listener, error) {
	listeners, err := c.getListeners(ingressName, ingressNs)
	if err != nil {
		return nil, nil, err
	}

	segListeners, err := c.getSegmentListeners(ingressName, ingressNs)
	if err != nil {
		return nil, nil, err
	}

	return listeners, segListeners, nil
}

func (c *IngressListenerConverter) getListeners(ingressName, ingressNamespace string) (
	[]networkextensionv1.Listener, error) {
	existedListenerList := &networkextensionv1.ListenerList{}
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
		ingressName: networkextensionv1.LabelValueForIngressName,
		networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
	})))
	err = c.Cli.List(context.TODO(), existedListenerList, &client.ListOptions{
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

func (c *IngressListenerConverter) getSegmentListeners(ingressName, ingressNamespace string) (
	[]networkextensionv1.Listener, error) {
	existedListenerList := &networkextensionv1.ListenerList{}
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
		ingressName: networkextensionv1.LabelValueForIngressName,
		networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueTrue,
	})))
	err = c.Cli.List(context.TODO(), existedListenerList, &client.ListOptions{
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
