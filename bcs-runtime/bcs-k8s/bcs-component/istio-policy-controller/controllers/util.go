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
	"fmt"
	"reflect"

	"istio.io/api/networking/v1alpha3"
	v1 "istio.io/client-go/pkg/apis/networking/v1"
	k8scorev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	// ControllerName controller name
	ControllerName = "istio-policy-controller"

	// LabelKeyManagedBy label key for istio-policy-controller
	LabelKeyManagedBy = "managed-by"
	// LabelKeyServiceNamespace label key for service namespace
	LabelKeyServiceNamespace = "service-namespace"
	// LabelKeyServiceName label key for service name
	LabelKeyServiceName = "service-name"

	// MergeModeMerge merge mode merge
	MergeModeMerge = "merge"
	// MergeModeOverride merge mode override
	MergeModeOverride = "override"
)

// sprintfHost 格式化 host 字符串
func sprintfHost(name, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace)
}

// getServicePredicate 获取 Service 事件的 Predicate
func getServicePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			svc, ok := e.Object.(*k8scorev1.Service)
			if !ok {
				return false
			}
			ctrl.Log.WithName("event").Info(fmt.Sprintf("Create service, name: %s, namespace: %s",
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

			ctrl.Log.WithName("event").Info(fmt.Sprintf(
				"Update new service, new service name: %s, old service name: %s, namespace: %s",
				newSvc.GetName(), oldSvc.GetName(), newSvc.GetNamespace()))
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			svc, ok := e.Object.(*k8scorev1.Service)
			if !ok {
				return false
			}

			ctrl.Log.WithName("event").Info(fmt.Sprintf("Delete service, name: %s, namespace: %s",
				svc.GetName(), svc.GetNamespace()))
			return true
		},
	}
}

// overrideDrPolicy override destination policy
func overrideDrPolicy(dr *v1.DestinationRule) {
	if isEmptyStruct(dr.Spec.TrafficPolicy.LoadBalancer) {
		dr.Spec.TrafficPolicy.LoadBalancer = nil
	}
	if isEmptyStruct(dr.Spec.TrafficPolicy.ConnectionPool) {
		dr.Spec.TrafficPolicy.ConnectionPool = nil
	}
	if isEmptyStruct(dr.Spec.TrafficPolicy.OutlierDetection) {
		dr.Spec.TrafficPolicy.OutlierDetection = nil
	}
	if isEmptyStruct(dr.Spec.TrafficPolicy.Tls) {
		dr.Spec.TrafficPolicy.Tls = nil
	}
	if len(dr.Spec.TrafficPolicy.PortLevelSettings) == 0 {
		dr.Spec.TrafficPolicy.PortLevelSettings = nil
	}
	if isEmptyStruct(dr.Spec.TrafficPolicy.Tunnel) {
		dr.Spec.TrafficPolicy.Tunnel = nil
	}
	if isEmptyStruct(dr.Spec.TrafficPolicy.ProxyProtocol) {
		dr.Spec.TrafficPolicy.ProxyProtocol = nil
	}
	if isEmptyStruct(dr.Spec.TrafficPolicy.RetryBudget) {
		dr.Spec.TrafficPolicy.RetryBudget = nil
	}
}

// mergeDrPolicy merge destination policy
func mergeDrPolicy(dr *v1.DestinationRule, tp *v1alpha3.TrafficPolicy) {
	if tp == nil {
		return
	}

	if dr.Spec.TrafficPolicy == nil {
		dr.Spec.TrafficPolicy = &v1alpha3.TrafficPolicy{}
	}

	if tp.LoadBalancer != nil {
		if isEmptyStruct(tp.LoadBalancer) {
			dr.Spec.TrafficPolicy.LoadBalancer = nil
		} else {
			dr.Spec.TrafficPolicy.LoadBalancer = tp.LoadBalancer
		}
	}
	if tp.ConnectionPool != nil {
		if isEmptyStruct(tp.ConnectionPool) {
			dr.Spec.TrafficPolicy.ConnectionPool = nil
		} else {
			dr.Spec.TrafficPolicy.ConnectionPool = tp.ConnectionPool
		}
	}
	if tp.OutlierDetection != nil {
		if isEmptyStruct(tp.OutlierDetection) {
			dr.Spec.TrafficPolicy.OutlierDetection = nil
		} else {
			dr.Spec.TrafficPolicy.OutlierDetection = tp.OutlierDetection
		}
	}
	if tp.Tls != nil {
		if isEmptyStruct(tp.Tls) {
			dr.Spec.TrafficPolicy.Tls = nil
		} else {
			dr.Spec.TrafficPolicy.Tls = tp.Tls
		}
	}
	if tp.PortLevelSettings != nil {
		if len(tp.PortLevelSettings) == 0 {
			dr.Spec.TrafficPolicy.PortLevelSettings = nil
		} else {
			dr.Spec.TrafficPolicy.PortLevelSettings = tp.PortLevelSettings
		}
	}
	if tp.Tunnel != nil {
		if isEmptyStruct(tp.Tunnel) {
			dr.Spec.TrafficPolicy.Tunnel = nil
		} else {
			dr.Spec.TrafficPolicy.Tunnel = tp.Tunnel
		}
	}
	if tp.ProxyProtocol != nil {
		if isEmptyStruct(tp.ProxyProtocol) {
			dr.Spec.TrafficPolicy.ProxyProtocol = nil
		} else {
			dr.Spec.TrafficPolicy.ProxyProtocol = tp.ProxyProtocol
		}
	}
	if tp.RetryBudget != nil {
		if isEmptyStruct(tp.RetryBudget) {
			dr.Spec.TrafficPolicy.RetryBudget = nil
		} else {
			dr.Spec.TrafficPolicy.RetryBudget = tp.RetryBudget
		}
	}
}

// isEmptyStruct 判断任意结构体是否逻辑上为零(注意: 不能用于非结构体类型,如 chan、func、interface、array、map等)
func isEmptyStruct(v interface{}) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)

	// 解引用指针（支持多层）
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return true
		}
		rv = rv.Elem()
		rt = rt.Elem()
	}

	if rv.Kind() != reflect.Struct {
		panic("isEmptyStruct only accepts struct or pointer to struct")
	}

	return isLogicallyZero(rv, rt)
}

// isLogicallyZero 递归判断任意 reflect.Value 是否逻辑上为零
func isLogicallyZero(val reflect.Value, typ reflect.Type) bool {
	switch val.Kind() {
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return val.Complex() == 0
	case reflect.String:
		return val.String() == ""
	case reflect.Ptr:
		if val.IsNil() {
			return true
		}
		return isLogicallyZero(val.Elem(), typ.Elem())
	case reflect.Slice, reflect.Map:
		return val.IsNil() // 若想认为空 slice/map 为零，改为 val.Len() == 0
	case reflect.Array:
		elemType := typ.Elem()
		for i := 0; i < val.Len(); i++ {
			if !isLogicallyZero(val.Index(i), elemType) {
				return false
			}
		}
		return true
	case reflect.Struct:
		// 遍历所有字段，跳过未导出的
		for i := 0; i < val.NumField(); i++ {
			fieldVal := val.Field(i)
			fieldType := typ.Field(i)

			// 跳过未导出字段：PkgPath 非空表示未导出
			if fieldType.PkgPath != "" {
				continue
			}

			if !isLogicallyZero(fieldVal, fieldType.Type) {
				return false
			}
		}
		return true
	case reflect.Interface:
		if val.IsNil() {
			return true
		}
		// 获取接口中实际存储的值
		actualVal := val.Elem()
		// 递归判断实际值是否为零，使用其自身的 Type
		return isLogicallyZero(actualVal, actualVal.Type())
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return val.IsNil()
	default:
		return false // 未知类型视为非零
	}
}
