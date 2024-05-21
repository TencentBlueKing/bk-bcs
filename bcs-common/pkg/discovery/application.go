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

// Package discovery xxx
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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/reflector"
	schedypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage/zookeeper"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
)

// ApplicationController controller for Application
type ApplicationController interface {
	// List list all application datas
	List(selector labels.Selector) (ret []*schedypes.Application, err error)
	// NamespaceList list specified applications by namespace
	// ns = namespace
	NamespaceList(ns string, selector labels.Selector) (ret []*schedypes.Application, err error)
	// GetByName xxx
	// Get application by the specified namespace, application name
	// ns = namespace
	// name = application name
	GetByName(ns, name string) (*schedypes.Application, error)
}

// applicationController for dataType resource
type applicationController struct {
	cxt          context.Context
	stopFn       context.CancelFunc
	eventStorage storage.Storage      // remote event storage
	indexer      k8scache.Indexer     // indexer
	reflector    *reflector.Reflector // reflector list/watch all datas to local memory cache
}

// List list all application datas
func (s *applicationController) List(selector labels.Selector) (ret []*schedypes.Application, err error) {
	err = ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*schedypes.Application))
	})
	return ret, err
}

// NamespaceList get specified application by namespace
func (s *applicationController) NamespaceList(ns string, selector labels.Selector) (ret []*schedypes.Application,
	err error) {
	err = ListAllByNamespace(s.indexer, ns, selector, func(m interface{}) {
		ret = append(ret, m.(*schedypes.Application))
	})
	return ret, err
}

// GetByName xxx
func (s *applicationController) GetByName(ns, name string) (*schedypes.Application, error) {
	obj, exists, err := s.indexer.GetByKey(path.Join(ns, name))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(k8sv1.Resource("Application"), fmt.Sprintf("%s.%s", ns, name))
	}
	return obj.(*schedypes.Application), nil
}

// NewApplicationController xxx
func NewApplicationController(hosts []string, eventHandler reflector.EventInterface) (ApplicationController, error) {
	indexers := k8scache.Indexers{
		meta.NamespaceIndex: meta.NamespaceIndexFunc,
	}

	ts := k8scache.NewIndexer(ApplicationObjectKeyFn, indexers)
	// create namespace client for zookeeper
	zkConfig := &zookeeper.ZkConfig{
		Hosts:         hosts,
		PrefixPath:    "/blueking/application",
		Name:          "application",
		Codec:         &meta.JsonCodec{},
		ObjectNewFunc: ApplicationObjectNewFn,
	}
	svcclient, err := zookeeper.NewNSClient(zkConfig)
	if err != nil {
		blog.Errorf("bk-bcs mesos discovery create application pod client failed, %s", err)
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
	ctl := &applicationController{
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

// ApplicationObjectKeyFn create key for application
func ApplicationObjectKeyFn(obj interface{}) (string, error) {
	application, ok := obj.(*schedypes.Application)
	if !ok {
		return "", fmt.Errorf("error object type")
	}
	return path.Join(application.GetNamespace(), application.GetName()), nil
}

// ApplicationObjectNewFn create new Application Object
func ApplicationObjectNewFn() meta.Object {
	return new(schedypes.Application)
}
