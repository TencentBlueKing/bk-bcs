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

package types

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	// TagResourceType tag for resource type
	TagResourceType = "resourceType"
	// TagResourceName tag for resource name
	TagResourceName = "resourceName"
	// TagNamespace tag for namespace
	TagNamespace = "namespace"
	// TagClusterID tag for cluster id
	TagClusterID = "clusterId"
	// TagCreateTime tag for create time
	TagCreateTime = "createTime"
	// TagUpdateTime tag for update time
	TagUpdateTime = "updateTime"
)

// ClusterNamespacedName comprise a resource name
type ClusterNamespacedName struct {
	ClusterID string
	Namespace string
	Name      string
}

// ObjectKey key for object
type ObjectKey ClusterNamespacedName

// ObjectType type of object
type ObjectType string

// Meta metadata for object
type Meta struct {
	Type       ObjectType `json:"resourceType" bson:"resourceType"`
	Name       string     `json:"resourceName" bson:"resourceName"`
	Namespace  string     `json:"namespace" bson:"namespace"`
	ClusterID  string     `json:"clusterId" bson:"clusterId"`
	CreateTime time.Time  `json:"createTime" bson:"createTime"`
	UpdateTime time.Time  `json:"updateTime" bson:"updateTime"`
}

// GetObjectType get object type
func (om *Meta) GetObjectType() ObjectType {
	return om.Type
}

// SetObjectType set object type
func (om *Meta) SetObjectType(t ObjectType) {
	om.Type = t
}

// GetName get name
func (om *Meta) GetName() string {
	return om.Name
}

// SetName set name
func (om *Meta) SetName(n string) {
	om.Name = n
}

// GetNamespace get namespace
func (om *Meta) GetNamespace() string {
	return om.Namespace
}

// SetNamespace set namespace
func (om *Meta) SetNamespace(ns string) {
	om.Namespace = ns
}

// GetClusterID get cluster id
func (om *Meta) GetClusterID() string {
	return om.ClusterID
}

// SetClusterID set cluster id
func (om *Meta) SetClusterID(cid string) {
	om.ClusterID = cid
}

// GetCreateTime get create time
func (om *Meta) GetCreateTime() time.Time {
	return om.CreateTime
}

// SetCreateTime set create time
func (om *Meta) SetCreateTime(t time.Time) {
	om.CreateTime = t
}

// GetUpdateTime get update time
func (om *Meta) GetUpdateTime() time.Time {
	return om.UpdateTime
}

// SetUpdateTime set update time
func (om *Meta) SetUpdateTime(t time.Time) {
	om.UpdateTime = t
}

// Object interface for object
// type Object interface {
// 	GetObjectType() ObjectType
// 	SetObjectType(t ObjectType)
// 	GetName() string
// 	SetName(n string)
// 	GetNamespace() string
// 	SetNamespace(ns string)
// 	GetClusterID() string
// 	SetClusterID(cid string)
// 	GetCreateTime() time.Time
// 	SetCreateTime(t time.Time)
// 	GetUpdateTime() time.Time
// 	SetUpdateTime(t time.Time)
// 	GetData() map[string]interface{}
// 	SetData(map[string]interface{}) error
// 	ToString() string
// }

// ObjectList list of object
// type ObjectList []Object

// RawObject store object
type RawObject struct {
	Meta `json:",inline" bson:",inline"`
	Data map[string]interface{} `json:"data" bson:"data"`
}

// GetData get data
func (obj *RawObject) GetData() map[string]interface{} {
	return obj.Data
}

// SetData set data
func (obj *RawObject) SetData(data map[string]interface{}) error {
	obj.Data = data
	return nil
}

// ToString convert object to string
func (obj *RawObject) ToString() string {
	bytes, _ := json.Marshal(obj.GetData())
	return fmt.Sprintf("meta: %s/%s/%s, data: %s",
		obj.GetClusterID(), obj.GetNamespace(), obj.GetName(), string(bytes))
}

// Key get key of object
func (obj *RawObject) Key() ObjectKey {
	return ObjectKey{ClusterID: obj.GetClusterID(), Namespace: obj.GetNamespace(), Name: obj.GetName()}
}

// ValueSelector selector for searching objects with certain value in their path
type ValueSelector struct {
	Pairs map[string]interface{}
}

// NewValueSelector create value selector
func NewValueSelector() *ValueSelector {
	m := make(map[string]interface{})
	return &ValueSelector{
		Pairs: m,
	}
}

// Set set selector value pairs
func (vs *ValueSelector) Set(path string, value interface{}) error {
	vs.Pairs[path] = value
	return nil
}

// Get get selector value
func (vs *ValueSelector) Get(path string) (interface{}, bool) {
	value, ok := vs.Pairs[path]
	return value, ok
}

// GetPairs return map[string]interface{}
func (vs *ValueSelector) GetPairs() map[string]interface{} {
	return vs.Pairs
}
