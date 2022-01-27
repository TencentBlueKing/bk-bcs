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

package admin

import (
	"crypto/tls"
	"encoding/json"
)

//Option client option
type Option struct {
	AdminToken string
	TLSConfig  *tls.Config
	Addrs      []string
}

// Basic response from apisix admin interface
type Basic struct {
	Action  string `json:"action"`
	Count   int    `json:"count"`
	Data    *Data  `json:"node,omitempty"`
	Message string `json:"message,omitempty"`
	Err     string `json:"error_msg,omitempty"`
}

// Data information from apisix admin interface
type Data struct {
	Directory     bool   `json:"dir,omitempty"`
	Key           string `json:"key"`
	CreateIndex   uint   `json:"createdIndex"`
	ModifiedIndex uint   `json:"modifiedIndex"`
	//! when no data response from apisix, response.Data.Nodes is {} and Basic.Count is 1.
	//* if any services response from apisix, response.Data.Node is slice
	Nodes json.RawMessage `json:"nodes,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

// Node data holder
type Node struct {
	Key           string          `json:"key"`
	CreateIndex   uint            `json:"createdIndex"`
	ModifiedIndex uint            `json:"modifiedIndex"`
	Value         json.RawMessage `json:"value"`
}

//Nodes for Data.Nodes
type Nodes []*Node

// Client definition for apisix admin api
type Client interface {
	//upstream operation
	GetUpstream(id string) (*Upstream, error)
	ListUpstream() ([]*Upstream, error)
	//create upstream, upstream id will auto generate by apisix when not setting
	CreateUpstream(upstr *Upstream) error
	UpdateUpstream(upstr *Upstream) error
	DeleteUpstream(id string) error

	//service operation
	GetService(id string) (*Service, error)
	ListService() ([]*Service, error)
	CreateService(svc *Service) error
	UpdateService(svc *Service) error
	DeleteService(id string) error

	//route operation
	GetRoute(id string) (*Route, error)
	ListRoute() ([]*Route, error)
	CreateRoute(route *Route) error
	UpdateRoute(route *Route) error
	DeleteRoute(id string) error
}
