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

package discovery

import (
	"context"
	"fmt"
	"path"
	"time"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	k8scache "k8s.io/client-go/tools/cache"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/reflector"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage/zookeeper"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
)

// ServiceController controller for Service
type ServiceController interface {
	// List list all service datas
	List(selector labels.Selector) (ret []*commtypes.BcsService, err error)
	// NamespaceList list specified services by namespace
	// ns = namespace
	NamespaceList(ns string, selector labels.Selector) (ret []*commtypes.BcsService, err error)
	// GetByName xxx
	// Get service by the specified namespace, service name
	// ns = namespace
	// name = service name
	GetByName(ns, name string) (*commtypes.BcsService, error)
}

// serviceController for dataType resource
type serviceController struct {
	cxt          context.Context
	stopFn       context.CancelFunc
	eventStorage storage.Storage      // remote event storage
	indexer      k8scache.Indexer     // indexer
	reflector    *reflector.Reflector // reflector list/watch all datas to local memory cache
}

// List list all service datas
func (s *serviceController) List(selector labels.Selector) (ret []*commtypes.BcsService, err error) {
	err = ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*commtypes.BcsService))
	})
	return ret, err
}

// NamespaceList get specified service by namespace
func (s *serviceController) NamespaceList(ns string, selector labels.Selector) (ret []*commtypes.BcsService,
	err error) {
	err = ListAllByNamespace(s.indexer, ns, selector, func(m interface{}) {
		ret = append(ret, m.(*commtypes.BcsService))
	})
	return ret, err
}

// GetByName xxx
func (s *serviceController) GetByName(ns, name string) (*commtypes.BcsService, error) {
	obj, exists, err := s.indexer.GetByKey(path.Join(ns, name))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(k8sv1.Resource("Service"), fmt.Sprintf("%s.%s", ns, name))
	}
	return obj.(*commtypes.BcsService), nil
}

// NewServiceController xxx
func NewServiceController(hosts []string, eventHandler reflector.EventInterface) (ServiceController, error) {
	indexers := k8scache.Indexers{
		meta.NamespaceIndex: meta.NamespaceIndexFunc,
	}

	ts := k8scache.NewIndexer(ServiceObjectKeyFn, indexers)
	// create namespace client for zookeeper
	zkConfig := &zookeeper.ZkConfig{
		Hosts:         hosts,
		PrefixPath:    "/blueking/service",
		Name:          "service",
		Codec:         &meta.JsonCodec{},
		ObjectNewFunc: ServiceObjectNewFn,
	}
	svcclient, err := zookeeper.NewNSClient(zkConfig)
	if err != nil {
		blog.Errorf("bk-bcs mesos discovery create service pod client failed, %s", err)
		return nil, err
	}
	// create listwatcher
	listwatcher := &reflector.ListWatch{
		ListFn: func() ([]meta.Object, error) {
			return svcclient.List(context.Background(), "", nil)
		},
		WatchFn: func() (watch.Interface, error) {
			return svcclient.Watch(context.Background(), "", "", nil)
		},
	}
	cxt, stopfn := context.WithCancel(context.Background())
	ctl := &serviceController{
		cxt:          cxt,
		stopFn:       stopfn,
		eventStorage: svcclient,
		indexer:      ts,
	}

	// create reflector
	ctl.reflector = reflector.NewReflector(fmt.Sprintf("Reflector-%s", zkConfig.Name), ts, listwatcher, time.Second*600,
		eventHandler)
	// sync all data object from remote event storage
	err = ctl.reflector.ListAllData()
	if err != nil {
		return nil, err
	}
	// run reflector controller
	go ctl.reflector.Run()

	return ctl, nil
}

// ServiceObjectKeyFn create key for service
func ServiceObjectKeyFn(obj interface{}) (string, error) {
	service, ok := obj.(*commtypes.BcsService)
	if !ok {
		return "", fmt.Errorf("error object type")
	}
	return path.Join(service.GetNamespace(), service.GetName()), nil
}

// ServiceObjectNewFn create new Service Object
func ServiceObjectNewFn() meta.Object {
	return new(commtypes.BcsService)
}
