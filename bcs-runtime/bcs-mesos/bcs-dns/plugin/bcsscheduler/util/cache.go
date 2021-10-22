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

package util

import (
	"encoding/json"
	"fmt"
	"strings"

	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
)

//storage is cache for storage DNS data

//DataMeta interface for operator namespace/name
type DataMeta interface {
	GetName() string
	GetNamespace() string
	GetLabels() map[string]string
	GetAnnotations() map[string]string
}

//Decoder decoder for detail data type
type Decoder interface {
	Decode([]byte) (interface{}, error)
}

//DNSDataKeyFunc function for data uniq key
func DNSDataKeyFunc(obj interface{}) (string, error) {
	meta, ok := obj.(DataMeta)
	if !ok {
		return "", fmt.Errorf("DataMeta type Assert failed")
	}
	ns := strings.ToLower(meta.GetNamespace())
	name := strings.ToLower(meta.GetName())
	return ns + "/" + name, nil
}

//SvcDecoder decoder for ServiceCache
type SvcDecoder struct{}

//Decode implementation of Decoder
func (svcd *SvcDecoder) Decode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	srv := new(bcstypes.BcsService)
	if err := json.Unmarshal(data, srv); err != nil {
		return nil, err
	}
	srv.Name = strings.ToLower(srv.Name)
	srv.NameSpace = strings.ToLower(srv.NameSpace)
	return srv, nil
}

//ServiceCache cache for bcs-scheduler service data
type ServiceCache struct {
	Store cache.Store
}

//GetServiceByEndpoint get service info by Endpoint
func (srv *ServiceCache) GetServiceByEndpoint(endpoint *bcstypes.BcsEndpoint) *bcstypes.BcsService {
	key := endpoint.GetNamespace() + "/" + endpoint.GetName()
	item, ok, _ := srv.Store.GetByKey(key)
	if ok {
		return item.(*bcstypes.BcsService)
	}
	return nil
}

//GetService get Service by name and namespace
func (srv *ServiceCache) GetService(namespace, name string) *bcstypes.BcsService {
	key := namespace + "/" + name
	item, ok, _ := srv.Store.GetByKey(key)
	if ok {
		return item.(*bcstypes.BcsService)
	}
	return nil
}

//EndpointDecoder decoder for EndpointCache
type EndpointDecoder struct{}

//Decode implementation of Decoder
func (epd *EndpointDecoder) Decode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	ep := new(bcstypes.BcsEndpoint)
	if err := json.Unmarshal(data, ep); err != nil {
		return nil, err
	}
	ep.Name = strings.ToLower(ep.Name)
	ep.NameSpace = strings.ToLower(ep.NameSpace)
	return ep, nil
}

//EndpointCache cache for bcs-scheduler endpoint data
type EndpointCache struct {
	Store cache.Store //Store for detail
}

//GetEndpointByService get endpoint info from BcsService
func (ep *EndpointCache) GetEndpointByService(svc *bcstypes.BcsService) *bcstypes.BcsEndpoint {
	key := svc.GetNamespace() + "/" + svc.GetName()
	item, ok, _ := ep.Store.GetByKey(key)
	if ok {
		return item.(*bcstypes.BcsEndpoint)
	}
	return nil
}

//ListEndpoints list all endpoints, change interface into *BcsEndpoint
func (ep *EndpointCache) ListEndpoints() (epList []*bcstypes.BcsEndpoint) {
	for _, item := range ep.Store.List() {
		epList = append(epList, item.(*bcstypes.BcsEndpoint))
	}
	return epList
}
