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
// package controllers contains the reconcile logic for the istio-policy-controller.
package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/istio-policy-controller/internal/metric"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/istio-policy-controller/internal/option"
	"github.com/go-logr/logr"
	"istio.io/api/networking/v1alpha3"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	"istio.io/client-go/pkg/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8scorev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceReconciler reconciler for k8s svc
type ServiceReconciler struct {
	// Client client for reconciler
	client.Client
	IstioClient *versioned.Clientset
	Log         logr.Logger
	// Option option for controller
	Option *option.ControllerOption
}

// SetupWithManager set node reconciler
func (sr *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&k8scorev1.Service{}).
		WithEventFilter(getServicePredicate()).
		Complete(sr)
	if err != nil {
		return err
	}

	return nil
}

// Reconcile reconcile for k8s svc
func (sr *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// 尝试获取 Service 对象
	var svc corev1.Service
	if err := sr.Get(ctx, req.NamespacedName, &svc); err != nil {
		// 如果 err 是 NotFound，说明是删除事件
		if client.IgnoreNotFound(err) != nil {
			sr.Log.Error(err, "unable to fetch Service")
			return ctrl.Result{}, err
		}

		sr.Log.Info(fmt.Sprintf("Service deleted, name: %s, namespace: %s", req.Name, req.Namespace))
		err = sr.deletePolicy(ctx, req.Namespace, req.Name)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	sr.Log.Info(fmt.Sprintf("Service created or updated, name: %s, namespace: %s", req.Name, req.Namespace))
	err := sr.createOrUpdatePolicy(ctx, req.Namespace, req.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// createOrUpdatePolicy 创建或更新策略
func (sr *ServiceReconciler) createOrUpdatePolicy(ctx context.Context, namespace, name string) error {
	dr, err := sr.IstioClient.NetworkingV1().DestinationRules(namespace).Get(ctx,
		name, metav1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			sr.Log.Error(err, "failed to get DestinationRule")
			return err
		}

		metric.PolicyGeneratedTotal.Inc()
		sr.Log.Info("Creating DestinationRule", "name", name, "namespace", namespace)
		err = sr.createDr(ctx, namespace, name)
		if err != nil {
			sr.Log.Error(err, "failed to create DestinationRule")
			return err
		}

		metric.PolicySuccessTotal.Inc()
		sr.Log.Info("DestinationRule created successfully")
	} else if dr != nil && dr.GetName() != "" {
		metric.PolicyConflictTotal.Inc()
		sr.Log.Info("Updating DestinationRule", "name", name, "namespace", namespace)
		err = sr.updateDr(ctx, dr)
		if err != nil {
			sr.Log.Error(err, "failed to update DestinationRule")
			return err
		}

		sr.Log.Info("DestinationRule updated successfully")
	}

	_, err = sr.IstioClient.NetworkingV1().VirtualServices(namespace).Get(ctx,
		name, metav1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			sr.Log.Error(err, "failed to get VirtualServices")
			return err
		}

		// 创建 VirtualServices
		sr.Log.Info("Creating VirtualServices", "name", name, "namespace", namespace)
		err = sr.createVs(ctx, namespace, name)
		if err != nil {
			sr.Log.Error(err, "failed to create VirtualServices")
			return err
		}

		sr.Log.Info("VirtualServices created successfully")
	}

	return nil
}

// createDr 创建 DestinationRule
func (sr *ServiceReconciler) createDr(ctx context.Context, namespace, name string) error {
	dr := &networkingv1.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				LabelKeyManagedBy:        ControllerName,
				LabelKeyServiceNamespace: namespace,
				LabelKeyServiceName:      name,
			},
		},
		Spec: v1alpha3.DestinationRule{
			Host: sprintfHost(name, namespace),
		},
	}

	for _, svc := range sr.Option.Cfg.Services {
		if svc.Name == name && svc.Namespace == namespace {
			if svc.TrafficPolicy == nil {
				return errors.New("no traffic policy found in service config")
			}

			dr.Spec.TrafficPolicy = svc.TrafficPolicy
			overrideDrPolicy(dr)
			_, err := sr.IstioClient.NetworkingV1().DestinationRules(namespace).
				Create(ctx, dr, metav1.CreateOptions{})

			return err
		}
	}

	if sr.Option.Cfg.Global.TrafficPolicy == nil {
		return errors.New("no traffic policy found in global config")
	}

	dr.Spec.TrafficPolicy = sr.Option.Cfg.Global.TrafficPolicy
	overrideDrPolicy(dr)
	_, err := sr.IstioClient.NetworkingV1().DestinationRules(namespace).
		Create(ctx, dr, metav1.CreateOptions{})

	return err
}

// createVs 创建 VirtualService
func (sr *ServiceReconciler) createVs(ctx context.Context, namespace, name string) error {
	vs := &networkingv1.VirtualService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				LabelKeyManagedBy:        ControllerName,
				LabelKeyServiceNamespace: namespace,
				LabelKeyServiceName:      name,
			},
		},
		Spec: v1alpha3.VirtualService{
			Hosts: []string{sprintfHost(name, namespace)},
			Http: []*v1alpha3.HTTPRoute{
				{
					Route: []*v1alpha3.HTTPRouteDestination{
						{
							Destination: &v1alpha3.Destination{
								Host: sprintfHost(name, namespace),
							},
						},
					},
				},
			},
		},
	}

	for _, svc := range sr.Option.Cfg.Services {
		if svc.Name == name && svc.Namespace == namespace {
			if svc.Setting.AutoGenerateVS {
				_, err := sr.IstioClient.NetworkingV1().VirtualServices(namespace).
					Create(ctx, vs, metav1.CreateOptions{})
				if err != nil {
					return err
				}
			}

			return nil
		}
	}

	if sr.Option.Cfg.Global.Setting.AutoGenerateVS {
		_, err := sr.IstioClient.NetworkingV1().VirtualServices(namespace).
			Create(ctx, vs, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// updateDr 更新 DestinationRule
func (sr *ServiceReconciler) updateDr(ctx context.Context, dr *networkingv1.DestinationRule) error {
	label := dr.GetLabels()
	if len(label) == 0 {
		label = map[string]string{}
	}
	label[LabelKeyManagedBy] = ControllerName
	label[dr.GetName()] = dr.GetName()
	label[dr.GetNamespace()] = dr.GetNamespace()

	for _, svc := range sr.Option.Cfg.Services {
		if svc.Name == dr.GetName() && svc.Namespace == dr.GetNamespace() {
			if v, ok := dr.GetLabels()[LabelKeyManagedBy]; (ok && v == ControllerName) ||
				svc.Setting.UpdateUnmanagedResources {
				if svc.TrafficPolicy == nil {
					return errors.New("no traffic policy found in service config")
				}

				if svc.Setting.MergeMode == MergeModeMerge {
					mergeDrPolicy(dr, svc.TrafficPolicy)
				} else {
					dr.Spec.TrafficPolicy = svc.TrafficPolicy
					overrideDrPolicy(dr)
				}
				dr.SetLabels(label)

				_, err := sr.IstioClient.NetworkingV1().DestinationRules(dr.GetNamespace()).
					Update(ctx, dr, metav1.UpdateOptions{})

				return err
			}

			return nil
		}
	}

	if v, ok := dr.GetLabels()[LabelKeyManagedBy]; (ok && v == ControllerName) ||
		sr.Option.Cfg.Global.Setting.UpdateUnmanagedResources {
		if sr.Option.Cfg.Global.TrafficPolicy == nil {
			return errors.New("no traffic policy found in global config")
		}

		if sr.Option.Cfg.Global.Setting.MergeMode == MergeModeMerge {
			mergeDrPolicy(dr, sr.Option.Cfg.Global.TrafficPolicy)
		} else {
			dr.Spec.TrafficPolicy = sr.Option.Cfg.Global.TrafficPolicy
			overrideDrPolicy(dr)
		}
		dr.SetLabels(label)

		_, err := sr.IstioClient.NetworkingV1().DestinationRules(dr.GetNamespace()).
			Update(ctx, dr, metav1.UpdateOptions{})

		return err
	}

	return nil
}

// deletePolicy 删除策略
func (sr *ServiceReconciler) deletePolicy(ctx context.Context, namespace, name string) error {
	dr, err := sr.IstioClient.NetworkingV1().DestinationRules(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			sr.Log.Error(err, "failed to get DestinationRule")
		}
	}

	vs, err := sr.IstioClient.NetworkingV1().VirtualServices(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			sr.Log.Error(err, "failed to get VirtualService")
		}
	}

	for _, svc := range sr.Option.Cfg.Services {
		if svc.Name == name && svc.Namespace == namespace {
			if svc.Setting.DeletePolicyOnServiceDelete {
				return sr.deleteDrAndVs(ctx, dr, vs)
			}

			return nil
		}
	}

	if sr.Option.Cfg.Global.Setting.DeletePolicyOnServiceDelete {
		return sr.deleteDrAndVs(ctx, dr, vs)
	}

	return nil
}

// deleteDrAndVs 删除 DestinationRule 和 VirtualService
func (sr *ServiceReconciler) deleteDrAndVs(ctx context.Context, dr *networkingv1.DestinationRule,
	vs *networkingv1.VirtualService) error {

	var drErr, vsErr error
	if dr != nil {
		if v, ok := dr.GetLabels()[LabelKeyManagedBy]; ok && v == ControllerName {
			drErr = sr.IstioClient.NetworkingV1().DestinationRules(dr.Namespace).Delete(ctx,
				dr.Name, metav1.DeleteOptions{})
			if drErr != nil {
				sr.Log.Error(drErr, "failed to delete DestinationRule")
			}

			sr.Log.Info("DestinationRule deleted successfully")
		}
	}

	if vs != nil {
		if v, ok := vs.GetLabels()[LabelKeyManagedBy]; ok && v == ControllerName {
			vsErr = sr.IstioClient.NetworkingV1().VirtualServices(vs.Namespace).Delete(ctx,
				vs.Name, metav1.DeleteOptions{})
			if vsErr != nil {
				sr.Log.Error(vsErr, "failed to delete VirtualService")
				return vsErr
			}

			sr.Log.Info("VirtualService deleted successfully")
		}
	}

	return errors.Join(drErr, vsErr)
}
