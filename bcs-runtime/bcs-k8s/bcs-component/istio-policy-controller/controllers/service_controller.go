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
// package xxx
package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/istio-policy-controller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/istio-policy-controller/pkg/config"
	"github.com/go-logr/logr"
	"istio.io/client-go/pkg/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8scorev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

func getServicePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			svc, ok := e.Object.(*k8scorev1.Service)
			if !ok {
				return false
			}
			ctrl.Log.WithName("event").Info(fmt.Sprintf("Create event svc name: %s, namespace: %s",
				svc.GetName(), svc.GetNamespace()))
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSvc, newOk := e.ObjectNew.(*k8scorev1.Service)
			oldSvc, oldOk := e.ObjectOld.(*k8scorev1.Service)
			if !newOk || !oldOk {
				return false
			}
			if newSvc.DeletionTimestamp != nil {
				return true
			}

			ctrl.Log.WithName("event").Info(fmt.Sprintf("Update event new svc name: %s, old svc name: %s, namespace: %s",
				newSvc.GetName(), oldSvc.GetName(), newSvc.GetNamespace()))
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			svc, ok := e.Object.(*k8scorev1.Service)
			if !ok {
				return false
			}

			ctrl.Log.WithName("event").Info(fmt.Sprintf("Delete event svc name: %s, namespace: %s",
				svc.GetName(), svc.GetNamespace()))
			return true
		},
	}
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

	// 开启一个线程,遍历所有svc
	go sr.updateExistSvcs()

	return nil
}

// Reconcile reconcile k8s node info
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
		// 在这里处理删除逻辑
		err = sr.deletePolicy(ctx, req.Namespace, req.Name)
		if err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// 如果走到这里，说明是创建或更新
	sr.Log.Info(fmt.Sprintf("Service created or updated, name: %s, namespace: %s", req.Name, req.Namespace))
	// 在这里添加你的业务逻辑
	err := sr.createOrUpdatePolicy(req.Namespace, req.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (sr *ServiceReconciler) createOrUpdatePolicy(namespace, name string) error {
	dr, err := sr.IstioClient.NetworkingV1().DestinationRules(namespace).Get(context.Background(),
		name, metav1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			sr.Log.Error(err, "failed to get DestinationRule")
			return err
		}

		// 创建 DestinationRule 的逻辑
		sr.Log.Info("Creating DestinationRule", "name", name, "namespace", namespace)
		err = sr.createDr(context.Background(), namespace, name)
		if err != nil {
			return err
		}
	}

	_, err = sr.IstioClient.NetworkingV1().VirtualServices(namespace).Get(context.Background(),
		name, metav1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			sr.Log.Error(err, "failed to get VirtualServices")
			return err
		}

		// 创建 VirtualServices 的逻辑
		sr.Log.Info("Creating VirtualServices", "name", name, "namespace", namespace)
		err = sr.createVs(context.Background(), namespace, name)
		if err != nil {
			return err
		}
	}

	// 更新 DestinationRule 的逻辑
	sr.Log.Info("Updating DestinationRule", "name", name, "namespace", namespace)
	for _, svc := range config.G.Services {
		if svc.Name == name && svc.Namespace == namespace {
			if svc.Setting.MergeMode == MergeModeMerge {
				if svc.TrafficPolicy != nil {
					if svc.TrafficPolicy.LoadBalancer != nil {
						dr.Spec.TrafficPolicy.LoadBalancer = svc.TrafficPolicy.LoadBalancer
					}
					if svc.TrafficPolicy.ConnectionPool != nil {
						dr.Spec.TrafficPolicy.ConnectionPool = svc.TrafficPolicy.ConnectionPool
					}
					if svc.TrafficPolicy.OutlierDetection != nil {
						dr.Spec.TrafficPolicy.OutlierDetection = svc.TrafficPolicy.OutlierDetection
					}
					if svc.TrafficPolicy.Tls != nil {
						dr.Spec.TrafficPolicy.Tls = svc.TrafficPolicy.Tls
					}
					if svc.TrafficPolicy.PortLevelSettings != nil {
						dr.Spec.TrafficPolicy.PortLevelSettings = svc.TrafficPolicy.PortLevelSettings
					}
					if svc.TrafficPolicy.Tunnel != nil {
						dr.Spec.TrafficPolicy.Tunnel = svc.TrafficPolicy.Tunnel
					}
					if svc.TrafficPolicy.ProxyProtocol != nil {
						dr.Spec.TrafficPolicy.ProxyProtocol = svc.TrafficPolicy.ProxyProtocol
					}
					// if svc.TrafficPolicy.RetryBudget != nil {
					// 	dr.Spec.TrafficPolicy.RetryBudget = svc.TrafficPolicy.RetryBudget
					// }
				}
			} else {
				dr.Spec.TrafficPolicy = svc.TrafficPolicy
			}
		}
	}

	_, err = sr.IstioClient.NetworkingV1().DestinationRules(dr.Namespace).
		Update(context.Background(), dr, metav1.UpdateOptions{})

	return err
}

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

	for _, svc := range config.G.Services {
		if svc.Name == name && svc.Namespace == namespace && svc.Setting.DeletePolicyOnServiceDelete {
			return sr.deleteDrAndVs(ctx, dr, vs)
		}
	}

	if config.G.Global.Setting.DeletePolicyOnServiceDelete {
		return sr.deleteDrAndVs(ctx, dr, vs)
	}

	return nil
}

// updateExistSvcs update exist services
func (sr *ServiceReconciler) updateExistSvcs() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			svcList := &corev1.ServiceList{}
			if err := sr.Client.List(context.Background(), svcList); err != nil {
				sr.Log.Error(err, "failed to list services")
				continue
			}
			// 处理逻辑

			for _, svc := range svcList.Items {
				// 处理每个 Service
				sr.Log.Info("Found service", "namespace", svc.Namespace, "name", svc.Name)
				// 你的业务逻辑...
				err := sr.createOrUpdatePolicy(svc.Namespace, svc.Name)
				if err != nil {
					sr.Log.Error(err, "failed to create or update policy",
						"namespace", svc.Namespace, "name", svc.Name)
					continue
				}
			}

			return
		}
	}
}
