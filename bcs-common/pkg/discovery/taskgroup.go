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

import (
	"fmt"
	"path"
	"strings"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/pkg/meta"
	"bk-bcs/bcs-common/pkg/reflector"
	schedtypes "bk-bcs/bcs-common/pkg/scheduler/types"
	"bk-bcs/bcs-common/pkg/storage"
	"bk-bcs/bcs-common/pkg/storage/zookeeper"
	"bk-bcs/bcs-common/pkg/watch"

	"golang.org/x/net/context"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	k8scache "k8s.io/client-go/tools/cache"
)

// TaskgroupController controller for Taskgroup
type TaskgroupController interface {
	// List list all taskgroup datas
	List(selector labels.Selector) (ret []*schedtypes.TaskGroup, err error)
	// NamespaceList list specified taskgroups by namespace
	// ns = namespace
	NamespaceList(ns string, selector labels.Selector) (ret []*schedtypes.TaskGroup, err error)
	// ApplicationList list specified taskgroups by application name.namespace
	// ns = namespace
	// appname = application name
	ApplicationList(ns, appname string, selector labels.Selector) (ret []*schedtypes.TaskGroup, err error)
	// Get taskgroup by the specified taskgroupID
	// taskgroupID, for examples: 0.test-app.defaultGroup.10001.1564385547430065980
	GetById(taskgroupId string) (*schedtypes.TaskGroup, error)
	// Get taskgroup by the specified namespace, taskgroup name
	// ns = namespace
	// name = taskgroup name, appname-index, for examples: test-app-0
	GetByName(ns, name string) (*schedtypes.TaskGroup, error)
}

//taskgroupController for dataType resource
type taskgroupController struct {
	cxt          context.Context
	stopFn       context.CancelFunc
	eventStorage storage.Storage      //remote event storage
	indexer      k8scache.Indexer     //indexer
	reflector    *reflector.Reflector //reflector list/watch all datas to local memory cache
}

// List list all taskgroup datas
func (s *taskgroupController) List(selector labels.Selector) (ret []*schedtypes.TaskGroup, err error) {
	err = ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*schedtypes.TaskGroup))
	})
	return ret, err
}

// NamespaceList get specified taskgroup by namespace
func (s *taskgroupController) NamespaceList(ns string, selector labels.Selector) (ret []*schedtypes.TaskGroup, err error) {
	err = ListAllByNamespace(s.indexer, ns, selector, func(m interface{}) {
		ret = append(ret, m.(*schedtypes.TaskGroup))
	})
	return ret, err
}

func (s *taskgroupController) ApplicationList(ns, appname string,
	selector labels.Selector) (ret []*schedtypes.TaskGroup, err error) {

	err = ListAllByApplication(s.indexer, ns, appname, selector, func(m interface{}) {
		ret = append(ret, m.(*schedtypes.TaskGroup))
	})
	return ret, err
}

func (s *taskgroupController) GetById(taskgroupId string) (*schedtypes.TaskGroup, error) {
	//taskgroupId example: 0.appname.namespace.10001.1505188438633098965
	arr := strings.Split(taskgroupId, ".")
	if len(arr) != 5 {
		return nil, fmt.Errorf("taskgroupId %s is invalid", taskgroupId)
	}

	obj, exists, err := s.indexer.GetByKey(path.Join(arr[2], arr[1]))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("Taskgroup"), taskgroupId)
	}
	return obj.(*schedtypes.TaskGroup), nil
}

func (s *taskgroupController) GetByName(ns, name string) (*schedtypes.TaskGroup, error) {
	obj, exists, err := s.indexer.GetByKey(path.Join(ns, name))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("Taskgroup"), fmt.Sprintf("%s.%s", ns, name))
	}
	return obj.(*schedtypes.TaskGroup), nil
}

func NewTaskgroupController(hosts []string, eventHandler reflector.EventInterface) (TaskgroupController, error) {
	indexers := k8scache.Indexers{
		meta.NamespaceIndex:   meta.NamespaceIndexFunc,
		meta.ApplicationIndex: meta.ApplicationIndexFunc,
	}

	ts := k8scache.NewIndexer(TaskGroupObjectKeyFn, indexers)
	//create namespace client for zookeeper
	zkConfig := &zookeeper.ZkConfig{
		Hosts:         hosts,
		PrefixPath:    "/blueking/application",
		Name:          "taskgroup",
		Codec:         &meta.JsonCodec{},
		ObjectNewFunc: TaskGroupObjectNewFn,
	}
	podclient, err := zookeeper.NewPodClient(zkConfig)
	if err != nil {
		blog.Errorf("bk-bcs mesos discovery create taskgroup pod client failed, %s", err)
		return nil, err
	}
	//create listwatcher
	listwatcher := &reflector.ListWatch{
		ListFn: func() ([]meta.Object, error) {
			return podclient.List(context.Background(), "", nil)
		},
		WatchFn: func() (watch.Interface, error) {
			return podclient.Watch(context.Background(), "", "", nil)
		},
	}
	cxt, stopfn := context.WithCancel(context.Background())
	ctl := &taskgroupController{
		cxt:          cxt,
		stopFn:       stopfn,
		eventStorage: podclient,
		indexer:      ts,
	}

	//create reflector
	ctl.reflector = reflector.NewReflector(fmt.Sprintf("Reflector-%s", zkConfig.Name), ts, listwatcher, time.Second*600, eventHandler)
	//sync all data object from remote event storage
	err = ctl.reflector.ListAllData()
	if err != nil {
		return nil, err
	}
	//run reflector controller
	go ctl.reflector.Run()

	return ctl, nil
}

//TaskGroupObjectKeyFn create key for taskgroup
func TaskGroupObjectKeyFn(obj interface{}) (string, error) {
	task, ok := obj.(*schedtypes.TaskGroup)
	if !ok {
		return "", fmt.Errorf("error object type")
	}
	return path.Join(task.GetNamespace(), task.GetName()), nil
}

//TaskGroupObjectNewFn create new Service Object
func TaskGroupObjectNewFn() meta.Object {
	return new(schedtypes.TaskGroup)
}
