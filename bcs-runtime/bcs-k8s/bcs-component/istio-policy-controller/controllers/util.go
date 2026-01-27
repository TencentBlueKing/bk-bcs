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

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/istio-policy-controller/pkg/config"
	"istio.io/api/networking/v1alpha3"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LabelKey   = "managed-by"
	LabelValue = "istio-policy-controller"

	MergeModeMerge    = "merge"
	MergeModeOverride = "override"
)

func (sr *ServiceReconciler) createDr(ctx context.Context, namespace, name string) error {
	dr := &networkingv1.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha3.DestinationRule{
			Host:          fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace),
			TrafficPolicy: &v1alpha3.TrafficPolicy{},
		},
	}

	var tp *v1alpha3.TrafficPolicy
	for _, svc := range config.G.Services {
		if svc.Name == name && svc.Namespace == namespace {
			tp = svc.TrafficPolicy
		}
	}

	if tp == nil {
		tp = config.G.Global.TrafficPolicy
	}

	_, err := sr.IstioClient.NetworkingV1().DestinationRules(namespace).
		Create(context.Background(), dr, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (sr *ServiceReconciler) createVs(ctx context.Context, namespace, name string) error {
	vs := &networkingv1.VirtualService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha3.VirtualService{
			Hosts: []string{fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace)},
			Http: []*v1alpha3.HTTPRoute{
				{
					Route: []*v1alpha3.HTTPRouteDestination{
						{
							Destination: &v1alpha3.Destination{
								Host: fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace),
							},
						},
					},
				},
			},
		},
	}

	for _, svc := range config.G.Services {
		if svc.Name == name && svc.Namespace == namespace {
			if svc.Setting.AutoGenerateVS {
				_, err := sr.IstioClient.NetworkingV1().VirtualServices(namespace).
					Create(context.Background(), vs, metav1.CreateOptions{})
				if err != nil {
					return err
				}
			}
		}
	}

	if config.G.Global.Setting.AutoGenerateVS {
		_, err := sr.IstioClient.NetworkingV1().VirtualServices(namespace).
			Create(context.Background(), vs, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (sr *ServiceReconciler) deleteDrAndVs(ctx context.Context, dr *networkingv1.DestinationRule,
	vs *networkingv1.VirtualService) error {

	var drErr, vsErr error
	if dr != nil {
		if v, ok := dr.GetLabels()[LabelKey]; ok && v == LabelValue {
			drErr = sr.IstioClient.NetworkingV1().DestinationRules(dr.Namespace).Delete(ctx,
				dr.Name, metav1.DeleteOptions{})
			if drErr != nil {
				sr.Log.Error(drErr, "failed to delete DestinationRule")
			}
		}
	}

	if vs != nil {
		if v, ok := vs.GetLabels()[LabelKey]; ok && v == LabelValue {
			vsErr = sr.IstioClient.NetworkingV1().VirtualServices(vs.Namespace).Delete(ctx,
				vs.Name, metav1.DeleteOptions{})
			if vsErr != nil {
				sr.Log.Error(vsErr, "failed to delete VirtualService")
				return vsErr
			}
		}
	}

	if drErr != nil || vsErr != nil {
		errs := make([]string, 0)
		if drErr != nil {
			errs = append(errs, drErr.Error())
		}
		if vsErr != nil {
			errs = append(errs, vsErr.Error())
		}

		return fmt.Errorf(strings.Join(errs, ";"))
	}

	return nil
}
