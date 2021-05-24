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

package bcs

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/queue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/reflector"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage/zookeeper"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
	"github.com/Tencent/bk-bcs/bmsf-mesh/bmsf-mesos-adapter/pkg/util/str"
	v1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/mesh/v1"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8scache "k8s.io/client-go/tools/cache"
)

//svcController for dataType resource
type svcController struct {
	cxt          context.Context
	stopFn       context.CancelFunc
	eventStorage storage.Storage      //remote event storage
	svcCache     k8scache.Store       //local cache
	reflector    *reflector.Reflector //reflector list/watch all datas to local memory cache
	svcCh        chan *svcEvent       //event for service==>AppSvc
	svcQueue     queue.Queue          //queue for event message
}

func (s *svcController) run() {
	if err := s.reflector.ListAllData(); err != nil {
		blog.Errorf("list all bcs service failed, err %s", err.Error())
	}
	go s.reflector.Run()
}

func (s *svcController) stop() {
	s.reflector.Stop()
	s.eventStorage.Close()
	close(s.svcCh)
}

// GetAppSvc get specified AppSvc by namespace, name
func (s *svcController) GetAppSvc(ns, name string) (*v1.AppSvc, error) {
	return nil, nil
}

// ListAppSvcs List all AppSvc datas
func (s *svcController) ListAppSvcs(selector labels.Selector) ([]*v1.AppSvc, error) {
	return nil, nil
}

// RegisterAppSvcHandler register event callback for AppSvc
func (s *svcController) RegisterAppSvcQueue(handler queue.Queue) {
	s.svcQueue = handler
	blog.Infof("bk-bcs service Controller starting backend goroutine for queue handling")
	go s.handleService()
}

func (s *svcController) OnAdd(obj interface{}) {
	if obj == nil {
		return
	}
	svc, ok := obj.(*bcstypes.BcsService)
	if !ok {
		blog.Errorf("bk-bcs service plugin get error object type when ServiceOnAdd")
		return
	}
	if !s.isValid(svc) {
		return
	}
	e := &svcEvent{
		EventType: watch.EventAdded,
		Cur:       svc,
	}
	blog.V(3).Infof("bk-bcs service %s/%s trigger Add Event", svc.GetNamespace(), svc.GetName())
	s.svcCh <- e
}

func (s *svcController) OnUpdate(old, cur interface{}) {
	if cur == nil || old == nil {
		return
	}
	svc, ok := cur.(*bcstypes.BcsService)
	oldSvc, ook := old.(*bcstypes.BcsService)
	if !ok || !ook {
		blog.Errorf("bk-bcs service plugin get error object type when ServiceOnUpdate")
		return
	}
	if reflect.DeepEqual(oldSvc.Spec, svc.Spec) && reflect.DeepEqual(oldSvc.ObjectMeta, svc.ObjectMeta) {
		blog.Warnf("bk-bcs service %s/%s nothing different on EventUpdate.", svc.GetNamespace(), svc.GetName())
		return
	}
	if !s.isValid(svc) || !s.isValid(oldSvc) {
		return
	}
	e := &svcEvent{
		EventType: watch.EventUpdated,
		Old:       oldSvc,
		Cur:       svc,
	}
	blog.V(3).Infof("bk-bcs service %s/%s trigger Update Event", svc.GetNamespace(), svc.GetName())
	s.svcCh <- e
}

func (s *svcController) OnDelete(obj interface{}) {
	if obj == nil {
		return
	}
	svc, ok := obj.(*bcstypes.BcsService)
	if !ok {
		blog.Errorf("bk-bcs service plugin get error object type when ServiceOnDelete")
		return
	}
	if !s.isValid(svc) {
		return
	}
	e := &svcEvent{
		EventType: watch.EventDeleted,
		Cur:       svc,
	}
	blog.V(3).Infof("bk-bcs service %s/%s trigger Delete Event", svc.GetNamespace(), svc.GetName())
	s.svcCh <- e
}

func (s *svcController) isValid(svc *bcstypes.BcsService) bool {
	//verify BcsService data
	if len(svc.Spec.Selector) == 0 {
		blog.Errorf("bk-bcs service %s/%s lost Selector info", svc.GetNamespace(), svc.GetName())
		return false
	}
	if len(svc.Spec.Ports) == 0 {
		blog.Errorf("bk-bcs service %s/%s lost Service Ports info", svc.GetNamespace(), svc.GetName())
		return false
	}
	return true
}

func (s *svcController) handleService() {
	blog.Infof("bk-bcs service backgroup goroutine starting...")
	for {
		select {
		case <-s.cxt.Done():
			blog.Infof("bk-bcs service event goroutine is asked exit.")
			return
		case event, ok := <-s.svcCh:
			if !ok {
				blog.Infof("bk-bcs service event channel broken, ready to exit.")
				return
			}
			//event can only watch.EventAdded, watch.EventUpdated, watch.EventDeleted:
			e := &queue.Event{
				Type: event.EventType,
			}
			appSvc := s.convertBkToAppSvc(event.Cur)
			appSvc.Status.LastUpdateTime = metav1.Now()
			e.Data = appSvc
			s.svcQueue.Push(e)
		}
	}
}

func (s *svcController) convertBkToAppSvc(bkSvc *bcstypes.BcsService) *v1.AppSvc {
	out := &v1.AppSvc{
		TypeMeta: metav1.TypeMeta{
			Kind:       "appsvc",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        str.ReplaceSpecialCharForSvcName(bkSvc.GetName()),
			Namespace:   bkSvc.GetNamespace(),
			Labels:      str.ReplaceSpecialCharForLabel(bkSvc.GetLabels()),
			Annotations: bkSvc.GetAnnotations(),
		},
		Spec: v1.AppSvcSpec{
			Selector: str.ReplaceSpecialCharForLabel(bkSvc.Spec.Selector),
			Type:     bkSvc.Spec.Type,
			Frontend: bkSvc.Spec.ClusterIP,
		},
		Status: v1.AppSvcStatus{},
	}
	out.Spec.ServicePorts = s.convertBkServicePort(bkSvc.Spec.Ports)
	return out
}

func (s *svcController) convertBkServicePort(ports []bcstypes.ServicePort) []v1.ServicePort {
	var svcPorts []v1.ServicePort
	for _, p := range ports {
		port := v1.ServicePort{
			Name:        p.Name,
			Protocol:    strings.ToLower(p.Protocol),
			Domain:      p.DomainName,
			Path:        p.Path,
			ServicePort: p.Port,
			ProxyPort:   p.NodePort,
		}
		svcPorts = append(svcPorts, port)
	}
	return svcPorts
}

func newServiceCache(hosts []string) (*svcController, error) {
	ss := k8scache.NewIndexer(ServiceObjectKeyFn, k8scache.Indexers{})
	//create namespace client for zookeeper
	zkConfig := &zookeeper.ZkConfig{
		Hosts:         hosts,
		PrefixPath:    "/blueking/service",
		Name:          "service",
		Codec:         &meta.JsonCodec{},
		ObjectNewFunc: ServiceObjectNewFn,
	}
	nsclient, err := zookeeper.NewStorage(zkConfig)
	if err != nil {
		blog.Errorf("bcs mesos discovery create service namespace client failed, %s", err)
		return nil, err
	}
	//create listwatcher
	listwatcher := &reflector.ListWatch{
		ListFn: func() ([]meta.Object, error) {
			return nsclient.List(context.Background(), "", nil)
		},
		WatchFn: func() (watch.Interface, error) {
			return nsclient.Watch(context.Background(), "", "", nil)
		},
	}
	cxt, stopfn := context.WithCancel(context.Background())
	ctl := &svcController{
		cxt:          cxt,
		stopFn:       stopfn,
		eventStorage: nsclient,
		svcCache:     ss,
		svcCh:        make(chan *svcEvent, 1024),
	}
	//create reflector
	ctl.reflector = reflector.NewReflector(fmt.Sprintf("Reflector-%s", zkConfig.Name), ss, listwatcher, time.Second*600, ctl)
	return ctl, nil
}

//ServiceObjectKeyFn create key for ServiceObject
func ServiceObjectKeyFn(obj interface{}) (string, error) {
	svc, ok := obj.(*bcstypes.BcsService)
	if !ok {
		return "", fmt.Errorf("error object type")
	}
	return path.Join(svc.NameSpace, svc.Name), nil
}

//ServiceObjectNewFn create new Service Object
func ServiceObjectNewFn() meta.Object {
	return new(bcstypes.BcsService)
}
