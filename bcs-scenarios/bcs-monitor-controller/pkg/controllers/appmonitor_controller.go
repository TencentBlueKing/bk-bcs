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

// Package controllers xxx
package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/render"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/repo"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/utils"
)

const defaultRetry = time.Second * 3

// AppMonitorReconciler reconciles a NoticeGroup object
type AppMonitorReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Ctx         context.Context
	Render      render.IRender
	RepoManager *repo.Manager
}

// Reconcile app monitor
func (r *AppMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	blog.Infof("appMonitor '%s' triggered", req.NamespacedName)
	appMonitor := &monitorextensionv1.AppMonitor{}
	if err := r.Get(context.Background(), req.NamespacedName, appMonitor); err != nil {
		if !k8serrors.IsNotFound(err) {
			blog.Errorf("Get NoticeGroup '%s' failed, err: %s", req.NamespacedName.String(), err.Error())
			return ctrl.Result{}, err
		}

		blog.Infof("NoticeGroup '%s' is deleted, skip...", req.NamespacedName.String())
		return ctrl.Result{}, nil
	}

	// 清理相关资源
	if appMonitor.DeletionTimestamp != nil {
		blog.Infof("found deleting app monitor '%s'", req.NamespacedName)
		retry, err := r.processDeleteAppMonitor(appMonitor)
		if err != nil {
			_ = r.updateSyncStatus(appMonitor, monitorextensionv1.SyncStateFailed, err)
			return ctrl.Result{}, err
		}
		if retry {
			return ctrl.Result{Requeue: true, RequeueAfter: DefaultRequeueDuration}, nil
		}

		blog.Infof("delete app monitor '%s' success", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	// check appMonitor finalizer/labels
	if err := r.checkFinalizer(appMonitor); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.checkLabels(appMonitor); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.checkRepo(appMonitor); err != nil {
		return ctrl.Result{}, err
	}

	// transfer appMonitor to sub resource, i.e. MonitorRule\NoticeGroup\Panel
	result, err := r.Render.Render(appMonitor)
	if err != nil {
		blog.Errorf("render appMonitor'%s/%s' failed, err: %s", appMonitor.Namespace, appMonitor.Name, err.Error())
		_ = r.updateSyncStatus(appMonitor, monitorextensionv1.SyncStateFailed, err)
		return ctrl.Result{}, err
	}

	needRetry, err := r.ensureResource(appMonitor, result)
	if err != nil {
		_ = r.updateSyncStatus(appMonitor, monitorextensionv1.SyncStateFailed, err)
		return ctrl.Result{}, err
	}
	if needRetry {
		blog.Infof("appmonitor'%s/%s' wait for resource synced...", appMonitor.GetNamespace(), appMonitor.GetName())
		_ = r.updateSyncStatus(appMonitor, monitorextensionv1.SyncStateNeedReSync, err)
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: defaultRetry,
		}, nil
	}
	blog.Infof(" ensure appMonitor '%s' success", req.NamespacedName)
	_ = r.updateSyncStatus(appMonitor, monitorextensionv1.SyncStateCompleted, nil)
	return ctrl.Result{}, nil
}

// eventPredicate 筛选AppMonitor Reconcile条件
func (r *AppMonitorReconciler) eventPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			monitor := createEvent.Object.(*monitorextensionv1.AppMonitor)
			// if appMonitor open IgnoreChange, do not retry when state is completed
			if monitor.DeletionTimestamp == nil &&
				monitor.Status.SyncStatus.State == monitorextensionv1.SyncStateCompleted &&
				monitor.Spec.IgnoreChange {
				blog.V(3).Infof("appMonitor '%s/%s' got create event, but is synced and ignore change",
					monitor.GetNamespace(), monitor.GetName())
				return false
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			newMonitor, okNew := e.ObjectNew.(*monitorextensionv1.AppMonitor)
			oldMonitor, okOld := e.ObjectOld.(*monitorextensionv1.AppMonitor)
			if !okNew || !okOld {
				return true
			}
			if reflect.DeepEqual(newMonitor.Spec, oldMonitor.Spec) &&
				reflect.DeepEqual(newMonitor.Finalizers, oldMonitor.Finalizers) &&
				reflect.DeepEqual(newMonitor.DeletionTimestamp, oldMonitor.DeletionTimestamp) &&
				reflect.DeepEqual(newMonitor.Annotations, oldMonitor.Annotations) {
				blog.V(5).Infof("appMonitor %+v updated, "+
					"but spec and finalizer and deletionTimestamp and annotations not change",
					newMonitor)
				return false
			}
			// if appMonitor open IgnoreChange, do not retry when state is completed
			if newMonitor.DeletionTimestamp == nil &&
				newMonitor.Status.SyncStatus.State == monitorextensionv1.SyncStateCompleted &&
				newMonitor.Spec.IgnoreChange {
				blog.V(3).Infof("appMonitor '%s/%s' updated, but is synced and ignore change",
					newMonitor.GetNamespace(), newMonitor.GetName())
				return false
			}
			return true
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitorextensionv1.AppMonitor{}).
		WithEventFilter(r.eventPredicate()).
		Complete(r)
}

func (r *AppMonitorReconciler) updateSyncStatus(appMonitor *monitorextensionv1.AppMonitor,
	state monitorextensionv1.SyncState, err error) error {
	blog.Infof("Update sync state of appMonitor (%s/%s) to %s", appMonitor.GetNamespace(), appMonitor.GetName(),
		state)
	appMonitor.Status.SyncStatus.State = state
	// err message
	if err != nil {
		appMonitor.Status.SyncStatus.Message = err.Error()
	} else {
		appMonitor.Status.SyncStatus.Message = ""
	}
	appMonitor.Status.SyncStatus.LastSyncTime = metav1.NewTime(time.Now())
	if inErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return r.Client.Status().Update(context.TODO(), appMonitor)
	}); inErr != nil {
		blog.Warnf("update appMonitor'%s/%s' failed, err: %s", appMonitor.GetNamespace(), appMonitor.GetName(), inErr.Error())
		return inErr
	}

	return nil
}

// checkFinalizer add finalizer if not exist
func (r *AppMonitorReconciler) checkFinalizer(appMonitor *monitorextensionv1.AppMonitor) error {
	if utils.ContainsString(appMonitor.Finalizers, FinalizerMonitorController) {
		return nil
	}

	appMonitor.Finalizers = append(appMonitor.Finalizers, FinalizerMonitorController)
	if err := r.Update(r.Ctx, appMonitor); err != nil {
		blog.Warnf("Update app monitor '%s/%s' failed, err: %s", appMonitor.Namespace, appMonitor.Name,
			err.Error())
		return errors.Wrapf(err, "Update app monitor '%s/%s' failed ", appMonitor.Namespace, appMonitor.Name)
	}
	return nil
}

// checkFinalizer add finalizer if not exist
func (r *AppMonitorReconciler) checkLabels(monitor *monitorextensionv1.AppMonitor) error {
	if monitor.Labels[monitorextensionv1.LabelKeyForScenarioName] != "" && monitor.Labels[monitorextensionv1.
		LabelKeyForBizID] != "" {
		return nil
	}

	patchStruct := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				monitorextensionv1.LabelKeyForScenarioName: monitor.Spec.Scenario,
				monitorextensionv1.LabelKeyForBizID:        monitor.Spec.BizId,
				// monitorextensionv1.LabelKeyForScenarioRepo: repo.GenRepoKeyFromAppMonitor(monitor),
			},
		},
	}
	patchBytes, err := json.Marshal(patchStruct)
	if err != nil {
		return errors.Wrapf(err, "marshal patchStruct for app monitor '%s/%s' failed", monitor.GetNamespace(),
			monitor.GetName())
	}
	rawPatch := client.RawPatch(k8stypes.MergePatchType, patchBytes)
	updateAppMonitor := &monitorextensionv1.AppMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      monitor.GetName(),
			Namespace: monitor.GetNamespace(),
		},
	}
	if inErr := r.Patch(context.Background(), updateAppMonitor, rawPatch, &client.PatchOptions{}); inErr != nil {
		return errors.Wrapf(err, "patch app monitor %s/%s annotation failed, patcheStruct: %s",
			monitor.GetNamespace(), monitor.GetName(), string(patchBytes))
	}
	return nil
}

func (r *AppMonitorReconciler) checkRepo(monitor *monitorextensionv1.AppMonitor) error {
	repoRef := monitor.Spec.RepoRef
	if repoRef == nil {
		return nil
	}

	repoKey := repo.GenRepoKeyFromAppMonitor(monitor)
	_, ok := r.RepoManager.GetRepo(repoKey)
	if !ok {
		if err := r.RepoManager.RegisterRepoFromArgo(repoRef.URL, repoRef.TargetRevision); err != nil {
			return fmt.Errorf("register repo failed, err: %s", err.Error())
		}
	}

	return nil
}

func (r *AppMonitorReconciler) removeFinalizer(monitor *monitorextensionv1.AppMonitor) error {
	monitor.Finalizers = utils.RemoveString(monitor.Finalizers, FinalizerMonitorController)
	if err := r.Update(context.Background(), monitor, &client.UpdateOptions{}); err != nil {
		blog.Warnf("remove finalizer for monitor %s/%s failed, err %s", monitor.GetNamespace(), monitor.GetName(),
			err.Error())
		return fmt.Errorf("remove finalizer for monitor %s/%s failed, err %s", monitor.GetNamespace(),
			monitor.GetName(), err.Error())
	}
	blog.V(3).Infof("remove finalizer for monitor %s/%s successfully", monitor.GetNamespace(), monitor.GetName())
	return nil
}

// return true/err if monitor need  retry
func (r *AppMonitorReconciler) processDeleteAppMonitor(monitor *monitorextensionv1.AppMonitor) (bool, error) {
	// if sub resource is cleaned, delete appMonitor
	canDelete, cErr := r.checkRelatedResource(monitor)
	if cErr != nil {
		blog.Errorf("check related resource for monitor '%s/%s' failed, err: %s", monitor.GetNamespace(),
			monitor.GetName(), cErr.Error())
		return true, errors.Wrapf(cErr, "check related resource for monitor '%s/%s' failed", monitor.GetNamespace(),
			monitor.GetName())
	}

	if canDelete {
		if err := r.removeFinalizer(monitor); err != nil {
			return true, err
		}
		return false, nil
	}

	blog.Infof("app monitor '%s/%s' related resource need te be cleaned", monitor.GetNamespace(), monitor.GetName())

	if err := r.deletePanel(monitor); err != nil {
		return true, err
	}
	if err := r.deleteMonitorRule(monitor); err != nil {
		return true, err
	}
	// configMap should be reused by other appMonitor with same scenario
	// if err := r.deleteConfigmap(monitor); err != nil {
	// 	return true, err
	// }
	if err := r.deleteNoticeGroup(monitor); err != nil {
		return true, err
	}

	return true, nil
}

// return true if related resource is cleaned
func (r *AppMonitorReconciler) checkRelatedResource(monitor *monitorextensionv1.AppMonitor) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForAppMonitorName: monitor.Name,
		monitorextensionv1.LabelKeyForScenarioName:   monitor.Spec.Scenario,
	}))
	if err != nil {
		blog.Errorf("build label selector failed, err: %s", err.Error())
		return false, err
	}

	monitorRuleList := &monitorextensionv1.MonitorRuleList{}
	err = r.List(r.Ctx, monitorRuleList, &client.ListOptions{
		LabelSelector: selector,
		Namespace:     monitor.GetNamespace(),
	})
	if err != nil {
		blog.Errorf("list monitorRule failed, err: %s", err.Error())
		return false, errors.Wrapf(err, "list monitorRule failed")
	}
	if len(monitorRuleList.Items) != 0 {
		return false, nil
	}

	panelList := &monitorextensionv1.PanelList{}
	err = r.List(r.Ctx, panelList, &client.ListOptions{
		LabelSelector: selector,
		Namespace:     monitor.GetNamespace(),
	})
	if err != nil {
		blog.Errorf("list panel failed, err: %s", err.Error())
		return false, errors.Wrapf(err, "list panel failed")
	}
	if len(panelList.Items) != 0 {
		return false, nil
	}

	cmList := &v1.ConfigMapList{}
	err = r.List(r.Ctx, cmList, &client.ListOptions{
		LabelSelector: selector,
		Namespace:     monitor.GetNamespace(),
	})
	if err != nil {
		blog.Errorf("list configmap failed, err: %s", err.Error())
		return false, errors.Wrapf(err, "list configmap failed")
	}
	if len(cmList.Items) != 0 {
		return false, nil
	}
	noticeGroupList := &monitorextensionv1.NoticeGroupList{}
	err = r.List(r.Ctx, noticeGroupList, &client.ListOptions{
		LabelSelector: selector,
		Namespace:     monitor.GetNamespace(),
	})
	if err != nil {
		blog.Errorf("list notice group  failed, err: %s", err.Error())
		return false, errors.Wrapf(err, "list notice group failed")
	}
	if len(noticeGroupList.Items) != 0 {
		return false, nil
	}

	return true, nil
}

// deleteMonitorRule delete related monitor rule
func (r *AppMonitorReconciler) deleteMonitorRule(monitor *monitorextensionv1.AppMonitor) error {
	monitorRule := &monitorextensionv1.MonitorRule{}
	selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForAppMonitorName: monitor.Name,
		monitorextensionv1.LabelKeyForScenarioName:   monitor.Spec.Scenario,
	}))
	if err != nil {
		blog.Errorf("get selector for deleted monitor rule '%s/%s' failed, err: %s", monitor.GetNamespace(),
			monitor.GetName(), err.Error())
		return errors.Wrapf(err, "get selector for deleted monitor rule '%s/%s' failed", monitor.GetNamespace(),
			monitor.GetName())
	}
	err = r.DeleteAllOf(r.Ctx, monitorRule, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			LabelSelector: selector,
			Namespace:     monitor.GetNamespace(),
		},
	})
	if err != nil {
		blog.Errorf("delete monitor rule by label selector '%s' failed, err: %s", selector.String(), err.Error())
		return errors.Wrapf(err, "delete monitor rule by label selector '%s' failed", selector.String())
	}
	return nil
}

// deletePanel delete related panel
func (r *AppMonitorReconciler) deletePanel(monitor *monitorextensionv1.AppMonitor) error {
	panel := &monitorextensionv1.Panel{}
	selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForAppMonitorName: monitor.Name,
		monitorextensionv1.LabelKeyForScenarioName:   monitor.Spec.Scenario,
	}))
	if err != nil {
		blog.Errorf("get selector for deleted panel '%s/%s' failed, err: %s", monitor.GetNamespace(),
			monitor.GetName(), err.Error())
		return errors.Wrapf(err, "get selector for deleted panel '%s/%s' failed", monitor.GetNamespace(),
			monitor.GetName())
	}
	err = r.DeleteAllOf(r.Ctx, panel, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			LabelSelector: selector,
			Namespace:     monitor.GetNamespace(),
		},
	})
	if err != nil {
		blog.Errorf("delete panel by label selector '%s' failed, err: %s", selector.String(), err.Error())
		return errors.Wrapf(err, "delete panel by label selector '%s' failed", selector.String())
	}
	return nil
}

// deleteNoticeGroup delete related notice group
func (r *AppMonitorReconciler) deleteNoticeGroup(monitor *monitorextensionv1.AppMonitor) error {
	noticeGroup := &monitorextensionv1.NoticeGroup{}
	selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForAppMonitorName: monitor.Name,
		monitorextensionv1.LabelKeyForScenarioName:   monitor.Spec.Scenario,
	}))
	if err != nil {
		blog.Errorf("get selector for deleted noticeGroup '%s/%s' failed, err: %s", monitor.GetNamespace(),
			monitor.GetName(), err.Error())
		return errors.Wrapf(err, "get selector for deleted noticeGroup '%s/%s' failed", monitor.GetNamespace(),
			monitor.GetName())
	}
	err = r.DeleteAllOf(r.Ctx, noticeGroup, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			LabelSelector: selector,
			Namespace:     monitor.GetNamespace(),
		},
	})
	if err != nil {
		blog.Errorf("delete noticeGroup by label selector '%s' failed, err: %s", selector.String(), err.Error())
		return errors.Wrapf(err, "delete noticeGroup by label selector '%s' failed", selector.String())
	}
	return nil
}

// deleteConfigMap delete related configmap
// nolint unused
func (r *AppMonitorReconciler) deleteConfigmap(monitor *monitorextensionv1.AppMonitor) error {
	configmap := &v1.ConfigMap{}
	selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
		// monitorextensionv1.LabelKeyForAppMonitorName: monitor.Name,
		monitorextensionv1.LabelKeyForScenarioName: monitor.Spec.Scenario,
	}))
	if err != nil {
		blog.Errorf("get selector for deleted configmap '%s/%s' failed, err: %s", monitor.GetNamespace(),
			monitor.GetName(), err.Error())
		return errors.Wrapf(err, "get selector for deleted configmap '%s/%s' failed", monitor.GetNamespace(),
			monitor.GetName())
	}
	err = r.DeleteAllOf(r.Ctx, configmap, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			LabelSelector: selector,
			Namespace:     monitor.GetNamespace(),
		},
	})
	if err != nil {
		blog.Errorf("delete configmap by label selector '%s' failed, err: %s", selector.String(), err.Error())
		return errors.Wrapf(err, "delete configmap by label selector '%s' failed", selector.String())
	}
	return nil
}

func (r *AppMonitorReconciler) ensureConfigmap(appMonitor *monitorextensionv1.AppMonitor, result *render.Result) error {
	// selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
	// 	// monitorextensionv1.LabelKeyForAppMonitorName: appMonitor.Name,
	// 	monitorextensionv1.LabelKeyForScenarioName: appMonitor.Spec.Scenario,
	// }))
	// if err != nil {
	// 	blog.Errorf("get selector for '%s/%s' failed, err: %s", appMonitor.GetNamespace(),
	// 		appMonitor.GetName(), err.Error())
	// 	return errors.Wrapf(err, "get selector for  %s/%s' failed", appMonitor.GetNamespace(),
	// 		appMonitor.GetName())
	// }

	existResources := &v1.ConfigMapList{}
	if err := r.List(context.Background(), existResources, &client.ListOptions{
		// LabelSelector: selector,
		Namespace: appMonitor.GetNamespace(),
	}); err != nil {
		blog.Errorf("list monitor '%s/%s' related configmaps failed, err: %s", appMonitor.GetNamespace(),
			appMonitor.GetName(), err.Error())
		return err
	}
	// 记录已经存在的资源
	existMap := make(map[string]bool)
	resMap := make(map[string]*v1.ConfigMap)
	for _, existResource := range existResources.Items {
		existMap[existResource.GetName()] = false
		resMap[existResource.GetName()] = &existResource
	}

	for _, configMap := range result.ConfigMaps {
		if _, ok := existMap[configMap.GetName()]; !ok {
			if inErr := r.Create(context.Background(), configMap); inErr != nil {
				blog.Errorf("create configmap '%s/%s' failed, err: %s", configMap.Namespace,
					configMap.Name, inErr.Error())
				return inErr
			}
		} else {
			existMap[configMap.GetName()] = true
			existRes := resMap[configMap.GetName()]
			// 如果spec没有变化， 则跳过
			if reflect.DeepEqual(existRes.Data, configMap.Data) {
				continue
			}
			blog.Infof("found updated configmap: %s/%s", existRes.GetNamespace(), existRes.GetName())
			existRes.Data = configMap.Data
			if inErr := r.Update(context.Background(), configMap); inErr != nil {
				blog.Errorf("update configmap '%s/%s' failed, err: %s", configMap.Namespace,
					configMap.Name, inErr.Error())
				return inErr
			}
		}
	}
	return nil
}

func (r *AppMonitorReconciler) ensureMonitorRule(appMonitor *monitorextensionv1.AppMonitor,
	result *render.Result) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForAppMonitorName: appMonitor.Name,
		monitorextensionv1.LabelKeyForScenarioName:   appMonitor.Spec.Scenario,
	}))
	if err != nil {
		blog.Errorf("get selector for '%s/%s' failed, err: %s", appMonitor.GetNamespace(),
			appMonitor.GetName(), err.Error())
		return true, errors.Wrapf(err, "get selector for  %s/%s' failed", appMonitor.GetNamespace(),
			appMonitor.GetName())
	}

	existResources := &monitorextensionv1.MonitorRuleList{}
	if err = r.List(context.Background(), existResources, &client.ListOptions{
		LabelSelector: selector,
		Namespace:     appMonitor.GetNamespace(),
	}); err != nil {
		blog.Errorf("list monitor '%s/%s' related monitor rule failed, err: %s", appMonitor.GetNamespace(),
			appMonitor.GetName(), err.Error())
		return true, err
	}
	// 记录已经存在的资源
	existMap := make(map[string]bool)
	resMap := make(map[string]*monitorextensionv1.MonitorRule)
	needRetry := false
	for _, existResource := range existResources.Items {
		existMap[existResource.GetName()] = false
		resMap[existResource.GetName()] = &existResource
	}

	for _, res := range result.MonitorRule {
		if _, ok := existMap[res.GetName()]; !ok {
			// if create resource, need to retry to wait
			needRetry = true
			if inErr := r.Create(context.Background(), res); inErr != nil {
				blog.Errorf("create monitor rule '%s/%s' failed, err: %s", res.Namespace,
					res.Name, inErr.Error())
				return true, inErr
			}
		} else {
			existMap[res.GetName()] = true
			existRes := resMap[res.GetName()]
			// 如果spec没有变化， 则跳过
			if reflect.DeepEqual(existRes.Spec, res.Spec) {
				continue
			}
			// if update resource, need to retry to wait
			needRetry = true
			blog.Infof("found updated monitorRule: %s/%s", existRes.GetNamespace(), existRes.GetName())
			existRes.Spec = res.Spec
			if inErr := r.Update(context.Background(), existRes); inErr != nil {
				blog.Errorf("update monitor rule '%s/%s' failed, err: %s", res.Namespace,
					res.Name, inErr.Error())
				return true, inErr
			}
		}
	}

	for name, used := range existMap {
		if !used {
			if inErr := r.Delete(context.Background(), &monitorextensionv1.MonitorRule{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: appMonitor.GetNamespace()},
			}); inErr != nil {
				blog.Errorf("delete monitor rule '%s/%s' failed, err: %s", appMonitor.GetNamespace(), name, inErr.Error())
				return true, inErr
			}
		}
	}
	return needRetry, nil
}

func (r *AppMonitorReconciler) ensurePanel(appMonitor *monitorextensionv1.AppMonitor,
	result *render.Result) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForAppMonitorName: appMonitor.Name,
		monitorextensionv1.LabelKeyForScenarioName:   appMonitor.Spec.Scenario,
	}))
	if err != nil {
		blog.Errorf("get selector for '%s/%s' failed, err: %s", appMonitor.GetNamespace(),
			appMonitor.GetName(), err.Error())
		return true, errors.Wrapf(err, "get selector for  %s/%s' failed", appMonitor.GetNamespace(),
			appMonitor.GetName())
	}

	existResources := &monitorextensionv1.PanelList{}
	if err = r.List(context.Background(), existResources, &client.ListOptions{
		LabelSelector: selector,
		Namespace:     appMonitor.GetNamespace(),
	}); err != nil {
		blog.Errorf("list monitor '%s/%s' related panel failed, err: %s", appMonitor.GetNamespace(),
			appMonitor.GetName(), err.Error())
		return true, err
	}
	// 记录已经存在的资源
	existMap := make(map[string]bool)
	resMap := make(map[string]*monitorextensionv1.Panel)
	needRetry := false
	for _, existResource := range existResources.Items {
		existMap[existResource.GetName()] = false
		resMap[existResource.GetName()] = &existResource
	}

	for _, res := range result.Panel {
		if _, ok := existMap[res.GetName()]; !ok {
			// if create new resource, need to retry to wait
			needRetry = true
			if inErr := r.Create(context.Background(), res); inErr != nil {
				blog.Errorf("create panel '%s/%s' failed, err: %s", res.Namespace,
					res.Name, inErr.Error())
				return true, inErr
			}
		} else {
			existMap[res.GetName()] = true
			existRes := resMap[res.GetName()]
			// 如果spec没有变化， 则跳过
			if reflect.DeepEqual(existRes.Spec, res.Spec) {
				continue
			}
			// if update resource, need to retry to wait
			needRetry = true
			blog.Infof("found updated panel: %s/%s", existRes.GetNamespace(), existRes.GetName())
			existRes.Spec = res.Spec
			if inErr := r.Update(context.Background(), existRes); inErr != nil {
				blog.Errorf("update panel '%s/%s' failed, err: %s", res.Namespace,
					res.Name, inErr.Error())
				return true, inErr
			}
		}
	}

	for name, used := range existMap {
		if !used {
			if inErr := r.Delete(context.Background(), &monitorextensionv1.Panel{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: appMonitor.GetNamespace()},
			}); inErr != nil {
				blog.Errorf("delete panel '%s/%s' failed, err: %s", appMonitor.GetNamespace(), name, inErr.Error())
				return true, inErr
			}
		}
	}
	return needRetry, nil
}

// return true if need retry
func (r *AppMonitorReconciler) ensureNoticeGroup(appMonitor *monitorextensionv1.AppMonitor,
	result *render.Result) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
		monitorextensionv1.LabelKeyForAppMonitorName: appMonitor.Name,
		monitorextensionv1.LabelKeyForScenarioName:   appMonitor.Spec.Scenario,
	}))
	if err != nil {
		blog.Errorf("get selector for '%s/%s' failed, err: %s", appMonitor.GetNamespace(),
			appMonitor.GetName(), err.Error())
		return true, errors.Wrapf(err, "get selector for  %s/%s' failed", appMonitor.GetNamespace(),
			appMonitor.GetName())
	}

	existResources := &monitorextensionv1.NoticeGroupList{}
	if err = r.List(context.Background(), existResources, &client.ListOptions{
		LabelSelector: selector,
		Namespace:     appMonitor.GetNamespace(),
	}); err != nil {
		blog.Errorf("list monitor '%s/%s' related notice group failed, err: %s", appMonitor.GetNamespace(),
			appMonitor.GetName(), err.Error())
		return true, err
	}
	// 记录已经存在的资源
	existMap := make(map[string]bool)
	resMap := make(map[string]*monitorextensionv1.NoticeGroup)
	needRetry := false
	for _, existResource := range existResources.Items {
		existMap[existResource.GetName()] = false
		resMap[existResource.GetName()] = &existResource
	}

	for _, res := range result.NoticeGroup {
		if _, ok := existMap[res.GetName()]; !ok {
			// if create new resource, need to retry to wait
			needRetry = true
			if inErr := r.Create(context.Background(), res); inErr != nil {
				blog.Errorf("create notice group '%s/%s' failed, err: %s", res.Namespace,
					res.Name, inErr.Error())
				return true, inErr
			}
		} else {
			existMap[res.GetName()] = true
			existRes := resMap[res.GetName()]
			// 如果spec没有变化， 则跳过
			if reflect.DeepEqual(existRes.Spec, res.Spec) {
				continue
			}
			// if update resource, need to retry to wait
			needRetry = true
			blog.Infof("found updated notice group: %s/%s", existRes.GetNamespace(), existRes.GetName())
			existRes.Spec = res.Spec
			if inErr := r.Update(context.Background(), existRes); inErr != nil {
				blog.Errorf("update notice group '%s/%s' failed, err: %s", res.Namespace,
					res.Name, inErr.Error())
				return true, inErr
			}
		}
	}

	for name, used := range existMap {
		if !used {
			if inErr := r.Delete(context.Background(), &monitorextensionv1.NoticeGroup{
				ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: appMonitor.GetNamespace()},
			}); inErr != nil {
				blog.Errorf("delete notice group '%s/%s' failed, err: %s", appMonitor.GetNamespace(), name, inErr.Error())
				return true, inErr
			}
		}
	}
	return needRetry, nil
}

// return true if need retry
func (r *AppMonitorReconciler) ensureResource(appMonitor *monitorextensionv1.AppMonitor,
	result *render.Result) (bool, error) {
	// sync sub resource
	if err := r.ensureConfigmap(appMonitor, result); err != nil {
		return true, err
	}

	if needRetry, err := r.ensureNoticeGroup(appMonitor, result); needRetry || err != nil {
		return true, err
	}

	if needRetry, err := r.ensureMonitorRule(appMonitor, result); needRetry || err != nil {
		return true, err
	}

	if needRetry, err := r.ensurePanel(appMonitor, result); needRetry || err != nil {
		return true, err
	}
	return false, nil
}
