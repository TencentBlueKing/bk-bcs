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

package meta

import (
	"time"
)

const (
	// NamespaceAll is the default argument to specify on a context when you want to list or filter resources across all namespaces
	NamespaceAll string = ""
)

//TypeMeta define version & type
type TypeMeta struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
}

//ObjectMeta common meta info for all Object
type ObjectMeta struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	CreationTimestamp time.Time         `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	ClusterName       string            `json:"clusterName,omitempty"`
}

//GetName get object name
func (obj *ObjectMeta) GetName() string {
	return obj.Name
}

//SetName set object name
func (obj *ObjectMeta) SetName(name string) {
	obj.Name = name
}

//GetNamespace get object namespace
func (obj *ObjectMeta) GetNamespace() string {
	return obj.Namespace
}

//SetNamespace set object namespace
func (obj *ObjectMeta) SetNamespace(ns string) {
	obj.Namespace = ns
}

//GetCreationTimestamp get create timestamp
func (obj *ObjectMeta) GetCreationTimestamp() time.Time {
	return obj.CreationTimestamp
}

//SetCreationTimestamp set creat timestamp
func (obj *ObjectMeta) SetCreationTimestamp(timestamp time.Time) {
	obj.CreationTimestamp = timestamp
}

//GetLabels get object labels
func (obj *ObjectMeta) GetLabels() map[string]string {
	return obj.Labels
}

//SetLabels set objec labels
func (obj *ObjectMeta) SetLabels(labels map[string]string) {
	obj.Labels = labels
}

//GetAnnotations get object annotation
func (obj *ObjectMeta) GetAnnotations() map[string]string {
	return obj.Annotations
}

//SetAnnotations get annotation name
func (obj *ObjectMeta) SetAnnotations(annotation map[string]string) {
	obj.Annotations = annotation
}

//GetClusterName get cluster name
func (obj *ObjectMeta) GetClusterName() string {
	return obj.ClusterName
}

//SetClusterName set cluster name
func (obj *ObjectMeta) SetClusterName(clusterName string) {
	obj.ClusterName = clusterName
}

//Objects define list for object
type Objects struct {
	ObjectMeta `json:"meta"`
	Items      []Object `json:"items"`
}

//GetItems implements List interface
func (objs *Objects) GetItems() []Object {
	return objs.Items
}

//SetItems implements List interface
func (objs *Objects) SetItems(list []Object) {
	objs.Items = list
}

func Accessor(obj interface{}) (Object, error) {
	switch t := obj.(type) {
	case Object:
		return t, nil
	default:
		return nil, errNotObject
	}
}
