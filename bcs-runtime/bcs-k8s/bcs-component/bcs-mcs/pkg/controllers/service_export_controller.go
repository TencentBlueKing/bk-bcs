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
	"fmt"
	"reflect"

	bcsmcsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/apis/mcs/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	mcsv1alpha1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

const (
	// ServiceExportControllerName is the name of the controller
	ServiceExportControllerName = "service-export-controller"
)

// ServiceExportController is a controller for service export
type ServiceExportController struct {
	client.Client
	AgentID             string
	ParentClusterClient client.Client
	EventRecorder       record.EventRecorder
}

func (c *ServiceExportController) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	klog.V(4).Infof("Reconciling ServiceExport %s.", req.NamespacedName.String())
	serviceExport := &mcsv1alpha1.ServiceExport{}
	if err := c.Get(ctx, req.NamespacedName, serviceExport); err != nil {
		klog.ErrorS(err, "unable to fetch serviceExport", "serviceExport", req.NamespacedName.String())
		return controllerruntime.Result{}, client.IgnoreNotFound(err)
	}
	//删除
	if !serviceExport.DeletionTimestamp.IsZero() {
		if utils.ContainsString(serviceExport.Finalizers, utils.BcsMcsFinalizerName) {
			klog.V(4).Infof("Reconciling ServiceExport %s. Deleting.", req.NamespacedName.String())
			if err := c.deleteManifest(ctx, serviceExport); err != nil {
				klog.ErrorS(err, "unable to delete manifest", "serviceExport", req.NamespacedName.String())
				return controllerruntime.Result{}, client.IgnoreNotFound(err)
			}
			serviceExport.ObjectMeta.Finalizers = utils.RemoveString(serviceExport.ObjectMeta.Finalizers, utils.BcsMcsFinalizerName)
			if err := c.Update(context.Background(), serviceExport); err != nil {
				return controllerruntime.Result{}, client.IgnoreNotFound(err)
			}
			return controllerruntime.Result{}, nil
		}
	}

	if !utils.ContainsString(serviceExport.ObjectMeta.Finalizers, utils.BcsMcsFinalizerName) {
		serviceExport.ObjectMeta.Finalizers = append(serviceExport.ObjectMeta.Finalizers, utils.BcsMcsFinalizerName)
		if err := c.Update(context.Background(), serviceExport); err != nil {
			klog.ErrorS(err, "unable to add finalizer to serviceExport", "serviceExport", req.NamespacedName.String())
			return controllerruntime.Result{}, client.IgnoreNotFound(err)
		}
	}

	//同步至manifest
	if err := c.syncToManifest(ctx, serviceExport); err != nil {
		klog.ErrorS(err, "unable to sync endpointSlice to manifest", "serviceExport", req.NamespacedName.String())
		c.EventRecorder.Event(serviceExport, corev1.EventTypeWarning, "Error", fmt.Sprintf("sync endpointSlice to manifest failed: %s", err.Error()))
		return controllerruntime.Result{}, client.IgnoreNotFound(err)
	}
	c.EventRecorder.Event(serviceExport, corev1.EventTypeNormal, "Synced", "Synced ServiceExport")
	return controllerruntime.Result{}, nil
}

func (c *ServiceExportController) deleteManifest(ctx context.Context, serviceExport *mcsv1alpha1.ServiceExport) error {
	matchManifestLabel := c.matchEndpointSliceManifestLabel(serviceExport)
	klog.Infof("delete manifest with labelsSelector: %v", matchManifestLabel)
	manifestNamespace := utils.GenManifestNamespace(c.AgentID)
	matchManifestOpts := []client.DeleteAllOfOption{
		client.InNamespace(manifestNamespace),
		client.MatchingLabels(matchManifestLabel),
	}
	err := c.ParentClusterClient.DeleteAllOf(ctx, &bcsmcsv1alpha1.Manifest{}, matchManifestOpts...)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	c.EventRecorder.Eventf(serviceExport, corev1.EventTypeNormal, "DeleteManifest", "delete manifest by labels %v successful", matchManifestLabel)
	return nil
}

func (c *ServiceExportController) syncToManifest(ctx context.Context, serviceExport *mcsv1alpha1.ServiceExport) error {
	klog.V(4).Infof("Syncing endpointSlice to manifest. serviceExport is %s/%s", serviceExport.Namespace, serviceExport.Name)
	//收集EndpointSlices,并更新至manifest
	endpointSliceList := &discoveryv1beta1.EndpointSliceList{}
	listEndpointSliceOpts := []client.ListOption{
		client.InNamespace(serviceExport.Namespace),
		client.MatchingLabels(map[string]string{discoveryv1beta1.LabelServiceName: serviceExport.Name}),
	}
	if err := c.List(ctx, endpointSliceList, listEndpointSliceOpts...); err != nil {
		klog.ErrorS(err, "unable to get endpointSliceList", "serviceExport", types.NamespacedName{Namespace: serviceExport.Namespace, Name: serviceExport.Name}.String())
		c.EventRecorder.Event(serviceExport, corev1.EventTypeWarning, "Error", fmt.Sprintf("get endpointSlice failed: %s", err.Error()))
		return client.IgnoreNotFound(err)
	}

	//将endpointSliceList 同步至manifest
	if endpointSliceList == nil {
		return c.deleteManifest(ctx, serviceExport)
	}
	if len(endpointSliceList.Items) == 0 {
		return c.deleteManifest(ctx, serviceExport)
	}

	manifestList := &bcsmcsv1alpha1.ManifestList{}
	manifestNamespace := utils.GenManifestNamespace(c.AgentID)
	matchManifestOpts := []client.ListOption{
		client.InNamespace(manifestNamespace),
		client.MatchingLabels(c.matchEndpointSliceManifestLabel(serviceExport)),
	}
	if err := c.ParentClusterClient.List(ctx, manifestList, matchManifestOpts...); err != nil {
		klog.ErrorS(err, "unable to get manifestList", "serviceExport", serviceExport.Name)
		return err
	}

	//判断哪些需要被删除
	toDeleteManifests := utils.FindNeedDeleteManifest(manifestList, endpointSliceList)
	for _, toDeleteManifest := range toDeleteManifests {
		klog.Infof("delete manifest: %s/%s", toDeleteManifest.Namespace, toDeleteManifest.Name)
		if err := c.ParentClusterClient.Delete(ctx, toDeleteManifest); err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			klog.ErrorS(err, "unable to delete manifest", "manifest", toDeleteManifest.Name)
			c.EventRecorder.Event(serviceExport, corev1.EventTypeWarning, "DeleteManifest", fmt.Sprintf("delete manifest %s/%s failed: %s", toDeleteManifest.Namespace, toDeleteManifest.Name, err.Error()))
			return err
		}
		c.EventRecorder.Event(serviceExport, corev1.EventTypeNormal, "DeleteManifest", fmt.Sprintf("delete manifest %s/%s successful", toDeleteManifest.Namespace, toDeleteManifest.Name))
	}
	//需要更新或创建的
	for _, endpointSlice := range endpointSliceList.Items {
		manifestName := utils.GenManifestName(utils.EndpointsSliceResourceName, endpointSlice.Namespace, endpointSlice.Name)
		manifestNamespace := utils.GenManifestNamespace(c.AgentID)
		manifest := &bcsmcsv1alpha1.Manifest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      manifestName,
				Namespace: manifestNamespace,
				Labels:    endpointSlice.GetLabels(),
			},
		}
		if manifest.Labels == nil {
			manifest.Labels = map[string]string{}
		}
		manifest.Labels[utils.ConfigGroupLabel] = endpointSlice.GroupVersionKind().Group
		manifest.Labels[utils.ConfigVersionLabel] = endpointSlice.GroupVersionKind().Version
		manifest.Labels[utils.ConfigKindLabel] = endpointSlice.GroupVersionKind().Kind
		manifest.Labels[utils.ConfigNameLabel] = endpointSlice.Name
		manifest.Labels[utils.ConfigNamespaceLabel] = endpointSlice.Namespace
		manifest.Labels[utils.ConfigUIDLabel] = string(endpointSlice.UID)
		manifest.Labels[utils.ConfigClusterLabel] = c.AgentID

		//update template
		if reflect.DeepEqual(manifest.Template.Object, &endpointSlice) {
			continue
		}
		manifest.Template = runtime.RawExtension{
			Object: &endpointSlice,
		}

		err := c.createOrUpdateManifest(serviceExport, manifest)
		if err != nil {
			klog.ErrorS(err, "create or update manifest failed", "manifest", manifest.Name)
			return err
		}
	}

	return nil
}

// createOrUpdateManifest 创建或更新Manifest
func (c *ServiceExportController) createOrUpdateManifest(serviceExport *mcsv1alpha1.ServiceExport, manifest *bcsmcsv1alpha1.Manifest) error {
	runtimeObject := manifest.DeepCopy()
	operationResult, err := controllerutil.CreateOrUpdate(context.TODO(), c.ParentClusterClient, runtimeObject, func() error {
		runtimeObject.Template = manifest.Template
		runtimeObject.Labels = manifest.Labels
		return nil
	})
	if err != nil {
		klog.Errorf("Failed to create/update Manifest %s/%s. Error: %v", manifest.GetNamespace(), manifest.GetName(), err)
		c.EventRecorder.Eventf(serviceExport, corev1.EventTypeWarning, "CreateOrUpdateManifestFailed", "Failed to create/update Manifest %s/%s. Error: %v", manifest.GetNamespace(), manifest.GetName(), err)
		return err
	}

	if operationResult == controllerutil.OperationResultCreated {
		klog.V(2).Infof("Create Manifest %s/%s successfully.", manifest.GetNamespace(), manifest.GetName())
		c.EventRecorder.Eventf(serviceExport, corev1.EventTypeNormal, "CreateManifest", "Create Manifest %s/%s successfully.", manifest.GetNamespace(), manifest.GetName())
	} else if operationResult == controllerutil.OperationResultUpdated {
		klog.V(2).Infof("Update Manifest %s/%s successfully.", manifest.GetNamespace(), manifest.GetName())
		c.EventRecorder.Eventf(serviceExport, corev1.EventTypeNormal, "UpdateManifest", "Update Manifest %s/%s successfully.", manifest.GetNamespace(), manifest.GetName())
	} else {
		klog.V(2).Infof("Manifest %s/%s is up to date.", manifest.GetNamespace(), manifest.GetName())
	}

	return nil
}

func (c *ServiceExportController) matchEndpointSliceManifestLabel(serviceExport *mcsv1alpha1.ServiceExport) map[string]string {
	return map[string]string{
		utils.ConfigClusterLabel:          c.AgentID,
		discoveryv1beta1.LabelServiceName: serviceExport.Name,
		utils.ConfigNamespaceLabel:        serviceExport.Namespace,
		utils.ConfigKindLabel:             utils.EndpointSliceKind,
	}
}

// SetupWithManager creates a controller and register to controller manager.
func (c *ServiceExportController) SetupWithManager(mgr controllerruntime.Manager) error {
	return controllerruntime.NewControllerManagedBy(mgr).For(&mcsv1alpha1.ServiceExport{}).
		// 同时监听EndpointSlice，通过  filterEndpointSliceFunc 方法过滤掉不需要的对象
		Watches(&source.Kind{Type: &discoveryv1beta1.EndpointSlice{}}, handler.EnqueueRequestsFromMapFunc(c.filterEndpointSliceFunc())).Complete(c)
}

func (c *ServiceExportController) filterEndpointSliceFunc() handler.MapFunc {
	return func(a client.Object) []reconcile.Request {
		name := a.GetName()
		namespace := a.GetNamespace()
		klog.V(5).Infof("watch EndpointSlice %s/%s change", namespace, name)
		var requests []reconcile.Request
		serviceExportName := a.GetLabels()[discoveryv1beta1.LabelServiceName]
		if serviceExportName == "" {
			klog.V(5).Infof("EndpointSlice %s/%s label service-name is empty", namespace, name)
			return nil
		}
		serviceExport := &mcsv1alpha1.ServiceExport{}
		err := c.Client.Get(context.TODO(), types.NamespacedName{Name: serviceExportName, Namespace: namespace}, serviceExport)
		if err != nil {
			if errors.IsNotFound(err) {
				klog.V(5).Infof("serviceExport %s/%s is not exists, skip requeue", namespace, name)
				return nil
			}
			klog.ErrorS(err, "get serviceExport failed, %s", "namespace", namespace, "name", serviceExportName)
		}

		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      serviceExportName,
				Namespace: namespace,
			},
		})
		return requests
	}
}
