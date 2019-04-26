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

package discovery

import "errors"

//IMeta meta interface
type IMeta interface {
	Key() string
}

var errLostKey = errors.New("object lost key")

//MetaKeyFunc key func for node cache
func MetaKeyFunc(obj interface{}) (string, error) {
	m, ok := obj.(IMeta)
	if !ok {
		return "", errLostKey
	}
	return m.Key(), nil
}

//Meta for location
type Meta struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

//Key get key of meta data
func (m *Meta) Key() string {
	return m.Cluster + "." + m.Namespace + "." + m.Name
}

//IsValid check meta is valid
func (m *Meta) IsValid() bool {
	if len(m.Cluster) == 0 {
		return false
	}
	if len(m.Namespace) == 0 {
		return false
	}
	if len(m.Name) == 0 {
		return false
	}
	return true
}

//IsEqual check two Meta is equal
func (m *Meta) IsEqual(other *Meta) bool {
	if m.Namespace != other.Namespace {
		return false
	}
	if m.Name != other.Name {
		return false
	}
	if m.Cluster != other.Cluster {
		return false
	}
	return true
}

//GetName return meta name
func (m *Meta) GetName() string {
	return m.Name
}

//GetNamespace return meta namespace
func (m *Meta) GetNamespace() string {
	return m.Namespace
}

//GetCluster return meta cluster
func (m *Meta) GetCluster() string {
	return m.Cluster
}
