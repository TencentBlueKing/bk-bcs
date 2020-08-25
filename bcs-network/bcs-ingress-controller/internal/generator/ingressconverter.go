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
	"fmt"
	"sort"
	"strings"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud"
)

// IngressConverter listener generator
type IngressConverter struct {
	cli              client.Client
	ingressValidater cloud.Validater
}

// NewIngressConverter create ingress generator
func NewIngressConverter(cli client.Client, ingressValidater cloud.Validater) *IngressConverter {
	return &IngressConverter{
		cli:              cli,
		ingressValidater: ingressValidater,
	}
}

// ProcessUpdateIngress process newly added or updated ingress
func (g *IngressConverter) ProcessUpdateIngress(ingress *networkextensionv1.Ingress) error {

	isValid, errMsg := g.ingressValidater.IsIngressValid(ingress)
	if !isValid {
		blog.Errorf("ingress %+v ingress is invalid, err %s", ingress, errMsg)
		return fmt.Errorf("ingress %+v ingress is invalid, err %s", ingress, errMsg)
	}

	isValid, errMsg = checkConflictsInIngress(ingress)
	if !isValid {
		blog.Errorf("ingress %+v ingress has conflicts, err %s", ingress, errMsg)
		return fmt.Errorf("ingress %+v ingress has conflicts, err %s", ingress, errMsg)
	}

	isConflict, err := g.checkConflicts(ingress)
	if err != nil {
		return err
	}
	if isConflict {
		blog.Errorf("ingress %+v is conflict with existed listeners", ingress)
		return fmt.Errorf("ingress %+v is conflict with existed listeners", ingress)
	}

	var lbIDs []string
	lbIDStrs, ok := ingress.Annotations[networkextensionv1.AnnotationKeyForLoadbalanceIDs]
	if !ok {
		blog.Warnf("ingress %+v is not associated with lb instance")
	} else {
		lbIDs = strings.Split(lbIDStrs, ",")
	}

	var generatedListeners []networkextensionv1.Listener
	var generatedSegListeners []networkextensionv1.Listener
	for _, rule := range ingress.Spec.Rules {
		ruleConverter := NewRuleConverter(g.cli, lbIDs, ingress.GetName(), ingress.GetNamespace(), &rule)
		listeners, err := ruleConverter.DoConvert()
		if err != nil {
			blog.Errorf("convert rule %+v failed, err %s", rule, err.Error())
			return fmt.Errorf("convert rule %+v failed, err %s", rule, err.Error())
		}
		generatedListeners = append(generatedListeners, listeners...)
	}
	for _, mapping := range ingress.Spec.PortMappings {
		MappingConverter := NewMappingConverter(g.cli, lbIDs, ingress.GetName(), ingress.GetNamespace(), &mapping)
		listeners, err := MappingConverter.DoConvert()
		if err != nil {
			blog.Errorf("convert mapping %+v failed, err %s", mapping, err.Error())
			return fmt.Errorf("convert mapping %+v failed, err %s", mapping, err.Error())
		}
		if mapping.IgnoreSegment {
			generatedListeners = append(generatedListeners, listeners...)
		} else {
			generatedSegListeners = append(generatedSegListeners, listeners...)
		}
	}
	sort.Sort(networkextensionv1.ListenerSlice(generatedListeners))
	sort.Sort(networkextensionv1.ListenerSlice(generatedSegListeners))

	existedListeners, err := g.getListeners(ingress.GetName(), ingress.GetNamespace())
	err = g.syncListeners(ingress.GetName(), ingress.GetNamespace(), existedListeners, generatedListeners)
	if err != nil {
		blog.Errorf("syncListeners listener failed, err %s", err.Error())
		return fmt.Errorf("syncListeners listener failed, err %s", err.Error())
	}

	existedSegListeners, err := g.getSegmentListeners(ingress.GetName(), ingress.GetNamespace())
	err = g.syncListeners(ingress.GetName(), ingress.GetNamespace(), existedSegListeners, generatedSegListeners)
	if err != nil {
		blog.Errorf("syncListeners listener failed, err %s", err.Error())
		return fmt.Errorf("syncListeners listener failed, err %s", err.Error())
	}
	return nil
}

// ProcessDeleteIngress  process deleted ingress
func (g *IngressConverter) ProcessDeleteIngress(ingressName, ingressNamespace string) error {
	listener := &networkextensionv1.Listener{}
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
		ingressName:      networkextensionv1.LabelValueForIngressName,
		ingressNamespace: networkextensionv1.LabelValueForIngressNamespace,
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
		ingressName:      networkextensionv1.LabelValueForIngressName,
		ingressNamespace: networkextensionv1.LabelValueForIngressNamespace,
		networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
	})))
	err = g.cli.List(context.TODO(), existedListenerList, &client.ListOptions{LabelSelector: selector})
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
		ingressName:      networkextensionv1.LabelValueForIngressName,
		ingressNamespace: networkextensionv1.LabelValueForIngressNamespace,
		networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueTrue,
	})))
	err = g.cli.List(context.TODO(), existedListenerList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		blog.Errorf("list segment listeners ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
		return nil, fmt.Errorf("list segment listeners ingress %s/%s failed, err %s",
			ingressName, ingressNamespace, err.Error())
	}
	return existedListenerList.Items, nil
}

func (g *IngressConverter) syncListeners(ingressName, ingressNamespace string,
	existedListeners, listeners []networkextensionv1.Listener) error {

	adds, dels, olds, news := GetDiffListeners(existedListeners, listeners)
	for _, del := range dels {
		err := g.cli.Delete(context.TODO(), &del, &client.DeleteOptions{})
		if err != nil {
			blog.Errorf("delete listener %+v failed, err %s", del, err.Error())
			return fmt.Errorf("delete listener %+v failed, err %s", del, err.Error())
		}
	}
	for _, add := range adds {
		err := g.cli.Create(context.TODO(), &add, &client.CreateOptions{})
		if err != nil {
			blog.Errorf("create listener %+v failed, err %s", add, err.Error())
			return fmt.Errorf("create listener %+v failed, err %s", add, err.Error())
		}
	}
	for index, new := range news {
		new.ResourceVersion = olds[index].ResourceVersion
		err := g.cli.Update(context.TODO(), &new, &client.UpdateOptions{})
		if err != nil {
			blog.Errorf("update listener %+v failed, err %s", new, err.Error())
			return fmt.Errorf("update listener %+v failed, err %s", new, err.Error())
		}
	}
	return nil
}
