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

// Package storage xxx
package storage

import "github.com/coredns/coredns/plugin/etcd/msg"

// storage for data persistence

// ServiceList define sort interface for msg.Service list
type ServiceList []msg.Service

// Len is the number of elements in the collection.
func (sl ServiceList) Len() int {
	return len(sl)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (sl ServiceList) Less(i, j int) bool {
	if sl[i].Host < sl[j].Host {
		return true
	} else if sl[i].Host > sl[j].Host {
		return false
	} else {
		return sl[i].Port < sl[j].Port
	}
}

// Swap swaps the elements with indexes i and j.
func (sl ServiceList) Swap(i, j int) {
	sl[i], sl[j] = sl[j], sl[i]
}

// Storage interface for dns data persistence
type Storage interface {
	AddService(domain string, msg []msg.Service) error
	UpdateService(domain string, old, cur []msg.Service) error
	DeleteService(domain string, msg []msg.Service) error
	ListServiceByName(domain string) ([]msg.Service, error)
	ListServiceByNamespace(namespace, cluster, zone string) ([]msg.Service, error)
	ListService(cluster, zone string) ([]msg.Service, error)
	Close()
}

// //EmptyStorage for no storage item in configuration
// type EmptyStorage struct {
// }

// //AddService add service dns data
// func (es *EmptyStorage) AddService(domain string, msgs []msg.Service) error {
// 	return nil
// }

// //UpdateService update service dns data
// func (es *EmptyStorage) UpdateService(domain string, old, cur []msg.Service) error {
// 	return nil
// }

// //DeleteService update service dns data
// func (es *EmptyStorage) DeleteService(domain string, msg []msg.Service) error {
// 	return nil
// }

// //ListServiceByName list service dns data by service name
// func (es *EmptyStorage) ListServiceByName(domain string) (svcList []msg.Service, err error) {
// 	return svcList, err
// }

// //ListServiceByNamespace list service dns data under namespace
// func (es *EmptyStorage) ListServiceByNamespace(namespace, cluster, zone string) (svcList []msg.Service, err error) {
// 	return svcList, err
// }

// //ListService list all service dns data from etcd
// func (es *EmptyStorage) ListService(cluster, zone string) ([]msg.Service, error) {
// 	return nil, nil
// }

// //Close close connection, release all context
// func (es *EmptyStorage) Close() {
// }
