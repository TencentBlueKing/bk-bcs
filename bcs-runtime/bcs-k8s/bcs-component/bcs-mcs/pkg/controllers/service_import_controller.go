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
 *
 */

package controllers

import (
	"context"

	bcsmcsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/apis/mcs/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	mcsv1alpha1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

const (
	// ServiceImportControllerName   is the name of the controller
	ServiceImportControllerName = "service-import-controller"
)

// ServiceImportController is a controller for service import
type ServiceImportController struct {
	client.Client
	AgentID             string
	ParentClusterClient client.Client
	EventRecorder       record.EventRecorder
}

func (c *ServiceImportController) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	klog.V(4).Infof("Reconciling ServiceImport %s.", req.NamespacedName.String())
	serviceImport := &mcsv1alpha1.ServiceImport{}
	if err := c.Get(ctx, req.NamespacedName, serviceImport); err != nil {
		klog.ErrorS(err, "unable to fetch serviceImport", "serviceImport", req.NamespacedName.String())
		return controllerruntime.Result{}, client.IgnoreNotFound(err)
	}
	//删除
	if !serviceImport.DeletionTimestamp.IsZero() {
		if utils.ContainsString(serviceImport.Finalizers, utils.BcsMcsFinalizerName) {
			klog.V(4).Infof("Reconciling serviceImport %s. Deleting.", req.NamespacedName.String())
			serviceImport.ObjectMeta.Finalizers = utils.RemoveString(serviceImport.ObjectMeta.Finalizers, utils.BcsMcsFinalizerName)
			if err := c.Update(context.Background(), serviceImport); err != nil {
				return controllerruntime.Result{}, client.IgnoreNotFound(err)
			}
			return controllerruntime.Result{}, nil
		}
	}

	if !utils.ContainsString(serviceImport.ObjectMeta.Finalizers, utils.BcsMcsFinalizerName) {
		serviceImport.ObjectMeta.Finalizers = append(serviceImport.ObjectMeta.Finalizers, utils.BcsMcsFinalizerName)
		if err := c.Update(context.Background(), serviceImport); err != nil {
			klog.ErrorS(err, "unable to add finalizer to serviceImport", "serviceImport", req.NamespacedName.String())
			return controllerruntime.Result{}, client.IgnoreNotFound(err)
		}
	}

	//create or update service
	if err := c.syncToDerivedService(ctx, serviceImport); err != nil {
		klog.ErrorS(err, "unable to create or update service", "service", req.NamespacedName.String())
		return controllerruntime.Result{}, client.IgnoreNotFound(err)
	}
	//同步Endpoint
	if err := c.syncToEndpointSlice(ctx, serviceImport); err != nil {
		klog.ErrorS(err, "unable to sync endpoint", "serviceImport", req.NamespacedName.String())
		return controllerruntime.Result{}, client.IgnoreNotFound(err)
	}
	return controllerruntime.Result{}, nil
}

func (c *ServiceImportController) syncToDerivedService(ctx context.Context, serviceImport *mcsv1alpha1.ServiceImport) error {
	newDerivedService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: serviceImport.Namespace,
			Name:      utils.GenerateDerivedServiceName(serviceImport.Name),
			Annotations: map[string]string{
				utils.ConfigCreatedBy: ServiceImportControllerName,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:       serviceImport.Name,
					Kind:       utils.ServiceImportKind,
					APIVersion: mcsv1alpha1.GroupVersion.String(),
					UID:        serviceImport.UID,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeClusterIP,
			Ports: servicePorts(serviceImport),
		},
	}

	oldDerivedService := &corev1.Service{}
	err := c.Client.Get(ctx, types.NamespacedName{
		Name:      utils.GenerateDerivedServiceName(serviceImport.Name),
		Namespace: serviceImport.Namespace,
	}, oldDerivedService)
	if err != nil {
		if errors.IsNotFound(err) {
			if err = c.Client.Create(ctx, newDerivedService); err != nil {
				klog.Errorf("Create derived service(%s/%s) failed, Error: %v", newDerivedService.Namespace, newDerivedService.Name, err)
				return err
			}
			c.EventRecorder.Eventf(serviceImport, corev1.EventTypeNormal, "CreateService", "Create Service %s/%s successfully.", newDerivedService.Namespace, newDerivedService.Name)
			return c.updateServiceStatus(serviceImport, newDerivedService)
		}
		return err
	}
	//service存在，判断是否是本controller创建的
	if oldDerivedService.Annotations[utils.ConfigCreatedBy] != ServiceImportControllerName {
		//冲突
		klog.Errorf("Service(%s/%s) is created by other controller, Error: %v", oldDerivedService.Namespace, oldDerivedService.Name, err)
		c.EventRecorder.Eventf(serviceImport, corev1.EventTypeWarning, "Conflict", "Service %s is created by other controller", serviceImport.Name)
		return nil
	}

	retainServiceFields(oldDerivedService, newDerivedService)
	err = c.Client.Update(context.TODO(), newDerivedService)
	if err != nil {
		klog.Errorf("Update derived service(%s/%s) failed, Error: %v", newDerivedService.Namespace, newDerivedService.Name, err)
		return err
	}

	return c.updateServiceStatus(serviceImport, newDerivedService)
}

func (c *ServiceImportController) syncToEndpointSlice(ctx context.Context, serviceImport *mcsv1alpha1.ServiceImport) error {
	klog.V(4).Infof("syncToEndpointSlice, serviceImport is %s/%s", serviceImport.Namespace, serviceImport.Name)
	//查找本地endpoint，以及远端Manifest中的对比差异
	endpointSliceList := &discoveryv1beta1.EndpointSliceList{}
	listEndpointSliceOpts := []client.ListOption{
		client.InNamespace(serviceImport.Namespace),
		client.MatchingLabels(map[string]string{
			discoveryv1beta1.LabelServiceName: utils.GenerateDerivedServiceName(serviceImport.Name),
			utils.ConfigCreatedBy:             ServiceImportControllerName,
		}),
	}
	err := c.List(ctx, endpointSliceList, listEndpointSliceOpts...)
	if err != nil {
		return err
	}
	//查找远端Manifest中的endpoint
	manifestList := &bcsmcsv1alpha1.ManifestList{}
	listManifestOpts := []client.ListOption{
		client.MatchingLabels(map[string]string{
			utils.ConfigNamespaceLabel:        serviceImport.Namespace,
			utils.ConfigKindLabel:             utils.EndpointSliceKind,
			discoveryv1beta1.LabelServiceName: serviceImport.Name,
		}),
	}
	err = c.ParentClusterClient.List(ctx, manifestList, listManifestOpts...)
	if err != nil {
		return err
	}

	//需要删除哪些本地的endpoint
	//从manifest中解析出endpoint
	remoteEndpointSlices := make([]*discoveryv1beta1.EndpointSlice, 0)
	for _, manifest := range manifestList.Items {
		remoteEndpointSlice, err := utils.UnmarshalEndpointSlice(&manifest)
		if err != nil {
			klog.ErrorS(err, "unable to unmarshal endpointSlice", "manifest", manifest.Name)
			c.EventRecorder.Eventf(serviceImport, corev1.EventTypeWarning, "UnmarshalEndpointSliceFailed", "UnmarshalEndpointSliceFailed %s/%s", manifest.Namespace, manifest.Name)
			continue
		}
		remoteEndpointSlices = append(remoteEndpointSlices, remoteEndpointSlice)
	}
	needDeleteLocalEndpointSlices := utils.FindNeedDeleteEndpointSlice(remoteEndpointSlices, endpointSliceList.Items)
	for _, needDeleteLocalEndpointSlice := range needDeleteLocalEndpointSlices {
		klog.V(4).Infof("delete local endpointSlice %s/%s", needDeleteLocalEndpointSlice.Namespace, needDeleteLocalEndpointSlice.Name)
		err := c.Delete(ctx, &needDeleteLocalEndpointSlice)
		if err != nil && !errors.IsNotFound(err) {
			klog.ErrorS(err, "delete endpointSlice failed", "endpointSlice", needDeleteLocalEndpointSlice.Name)
			c.EventRecorder.Eventf(serviceImport, corev1.EventTypeWarning, "Delete", "Delete endpointSlice %s failed", needDeleteLocalEndpointSlice.Name)
			continue
		}
		c.EventRecorder.Eventf(serviceImport, corev1.EventTypeNormal, "Delete", "Delete endpointSlice %s successfully", needDeleteLocalEndpointSlice.Name)
	}

	for _, remoteEndpointSlice := range remoteEndpointSlices {
		generateEndpointSliceName := utils.GenerateEndpointSliceName(remoteEndpointSlice.Name, remoteEndpointSlice.Labels[utils.ConfigClusterLabel])
		desiredEndpointSlice := remoteEndpointSlice.DeepCopy()
		desiredEndpointSlice.ObjectMeta = metav1.ObjectMeta{
			Namespace: remoteEndpointSlice.Namespace,
			Name:      generateEndpointSliceName,
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:       serviceImport.Name,
					Kind:       utils.ServiceImportKind,
					APIVersion: mcsv1alpha1.GroupVersion.String(),
					UID:        serviceImport.UID,
				},
			},
		}
		desiredEndpointSlice.Labels = map[string]string{
			mcsv1alpha1.LabelServiceName:      serviceImport.Name,
			discoveryv1beta1.LabelServiceName: utils.GenerateDerivedServiceName(serviceImport.Name),
			utils.ConfigCreatedBy:             ServiceImportControllerName,
			utils.ConfigClusterLabel:          remoteEndpointSlice.Labels[utils.ConfigClusterLabel],
		}
		err := c.createOrUpdateEndpointSlice(serviceImport, desiredEndpointSlice)
		if err != nil {
			klog.ErrorS(err, "create or update endpointSlice failed", "endpointSlice", desiredEndpointSlice.Name)
			continue
		}
	}
	return nil
}

// SetupWithManager creates a controller and register to controller manager.
func (c *ServiceImportController) SetupWithManager(mgr controllerruntime.Manager, parentCluster cluster.Cluster) error {
	return controllerruntime.NewControllerManagedBy(mgr).For(&mcsv1alpha1.ServiceImport{}).
		// 同时监听父集群的manifest对象
		Watches(
			source.NewKindWithCache(&bcsmcsv1alpha1.Manifest{}, parentCluster.GetCache()),
			handler.EnqueueRequestsFromMapFunc(c.filterManifest()),
		).Complete(c)
}

func (c *ServiceImportController) filterManifest() handler.MapFunc {
	return func(a client.Object) []reconcile.Request {
		var requests []reconcile.Request
		name := a.GetLabels()[discoveryv1beta1.LabelServiceName]
		namespace := a.GetLabels()[utils.ConfigNamespaceLabel]
		serviceImport := &mcsv1alpha1.ServiceImport{}
		err := c.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, serviceImport)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			klog.ErrorS(err, "get ServiceImport failed, %s", "namespace", namespace, "name", name)
		}
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      name,
				Namespace: namespace,
			},
		})
		return requests
	}
}

// updateServiceStatus update loadbalanacer status with provided clustersetIPs
func (c *ServiceImportController) updateServiceStatus(serviceImport *mcsv1alpha1.ServiceImport, derivedService *corev1.Service) error {
	ingress := make([]corev1.LoadBalancerIngress, 0)
	for _, ip := range serviceImport.Spec.IPs {
		ingress = append(ingress, corev1.LoadBalancerIngress{
			IP: ip,
		})
	}
	derivedService.Status = corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: ingress,
		},
	}

	if err := c.Client.Status().Update(context.TODO(), derivedService); err != nil {
		klog.Errorf("Update derived service(%s/%s) status failed, Error: %v", derivedService.Namespace, derivedService.Name, err)
		return err
	}

	return nil
}

// CreateOrUpdateEndpointSlice creates a EndpointSlice object if not exist, or updates if it already exist.
func (c *ServiceImportController) createOrUpdateEndpointSlice(serviceImport *mcsv1alpha1.ServiceImport, endpointSlice *discoveryv1beta1.EndpointSlice) error {
	runtimeObject := endpointSlice.DeepCopy()
	operationResult, err := controllerutil.CreateOrUpdate(context.TODO(), c.Client, runtimeObject, func() error {
		runtimeObject.AddressType = endpointSlice.AddressType
		runtimeObject.Endpoints = endpointSlice.Endpoints
		runtimeObject.Labels = endpointSlice.Labels
		runtimeObject.Ports = endpointSlice.Ports
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to create/update EndpointSlice %s/%s. Error: %v", endpointSlice.GetNamespace(), endpointSlice.GetName(), err)
		c.EventRecorder.Eventf(serviceImport, corev1.EventTypeWarning, "CreateOrUpdateEndpointSliceFailed", "Failed to create/update EndpointSlice %s/%s. Error: %v", endpointSlice.GetNamespace(), endpointSlice.GetName(), err)
		return err
	}

	if operationResult == controllerutil.OperationResultCreated {
		klog.V(2).Infof("Create EndpointSlice %s/%s successfully.", endpointSlice.GetNamespace(), endpointSlice.GetName())
		c.EventRecorder.Eventf(serviceImport, corev1.EventTypeNormal, "CreateEndpointSlice", "Create EndpointSlice %s/%s successfully.", endpointSlice.GetNamespace(), endpointSlice.GetName())
	} else if operationResult == controllerutil.OperationResultUpdated {
		klog.V(2).Infof("Update EndpointSlice %s/%s successfully.", endpointSlice.GetNamespace(), endpointSlice.GetName())
		c.EventRecorder.Eventf(serviceImport, corev1.EventTypeNormal, "UpdateEndpointSlice", "Update EndpointSlice %s/%s successfully.", endpointSlice.GetNamespace(), endpointSlice.GetName())
	} else {
		klog.V(2).Infof("EndpointSlice %s/%s is up to date.", endpointSlice.GetNamespace(), endpointSlice.GetName())
	}

	return nil
}

func servicePorts(svcImport *mcsv1alpha1.ServiceImport) []corev1.ServicePort {
	ports := make([]corev1.ServicePort, len(svcImport.Spec.Ports))
	for i, p := range svcImport.Spec.Ports {
		ports[i] = corev1.ServicePort{
			Name:        p.Name,
			Protocol:    p.Protocol,
			Port:        p.Port,
			AppProtocol: p.AppProtocol,
		}
	}
	return ports
}

func retainServiceFields(oldSvc, newSvc *corev1.Service) {
	newSvc.Spec.ClusterIP = oldSvc.Spec.ClusterIP
	newSvc.ResourceVersion = oldSvc.ResourceVersion
}
