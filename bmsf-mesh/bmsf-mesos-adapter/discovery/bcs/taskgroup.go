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
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/queue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/reflector"
	schetypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage/zookeeper"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
	v1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/mesh/v1"
	"github.com/Tencent/bk-bcs/bmsf-mesh/bmsf-mesos-adapter/pkg/util/str"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8scache "k8s.io/client-go/tools/cache"
)

//TaskGroup adaptor for bcs-scheduler TaskGroup to object data
type TaskGroup struct {
	schetypes.TaskGroup `json:",inline"`
}

//GetName for ObjectMeta
func (om *TaskGroup) GetName() string {
	return om.Name
}

//SetName set object name
func (om *TaskGroup) SetName(name string) {
	om.Name = name
}

//GetNamespace for ObjectMeta
func (om *TaskGroup) GetNamespace() string {
	return om.ObjectMeta.NameSpace
}

//SetNamespace set object namespace
func (om *TaskGroup) SetNamespace(ns string) {
	om.ObjectMeta.NameSpace = ns
}

//GetCreationTimestamp get create timestamp
func (om *TaskGroup) GetCreationTimestamp() time.Time {
	return om.ObjectMeta.CreationTimestamp
}

//SetCreationTimestamp set creat timestamp
func (om *TaskGroup) SetCreationTimestamp(timestamp time.Time) {
	om.ObjectMeta.CreationTimestamp = timestamp
}

//GetLabels for ObjectMeta
func (om *TaskGroup) GetLabels() map[string]string {
	return om.ObjectMeta.Labels
}

//SetLabels set objec labels
func (om *TaskGroup) SetLabels(labels map[string]string) {
	om.ObjectMeta.Labels = labels
}

//GetAnnotations for ObjectMeta
func (om *TaskGroup) GetAnnotations() map[string]string {
	return om.ObjectMeta.Annotations
}

//SetAnnotations get annotation name
func (om *TaskGroup) SetAnnotations(annotation map[string]string) {
	om.ObjectMeta.Annotations = annotation
}

//GetClusterName get cluster name
func (om *TaskGroup) GetClusterName() string {
	return om.ObjectMeta.ClusterName
}

//SetClusterName set cluster name
func (om *TaskGroup) SetClusterName(clusterName string) {
	om.ObjectMeta.ClusterName = clusterName
}

//svcController for dataType resource
type taskGroupController struct {
	cxt          context.Context
	stopFn       context.CancelFunc
	eventStorage storage.Storage      //remote event storage
	localCache   k8scache.Store       //local cache
	reflector    *reflector.Reflector //reflector list/watch all datas to local memory cache
	eventCh      chan *taskGroupEvent //event for service==>AppSvc
	eventQueue   queue.Queue          //queue for event message
}

func (s *taskGroupController) run() {
	if err := s.reflector.ListAllData(); err != nil {
		blog.Errorf("list all bcs taskgroup failed, err %s", err.Error())
	}
	go s.reflector.Run()
}

func (s *taskGroupController) stop() {
	s.reflector.Stop()
	s.eventStorage.Close()
	close(s.eventCh)
}

// ListAppNodes list all appNode datas
func (s *taskGroupController) ListAppNodes(selector labels.Selector) ([]*v1.AppNode, error) {
	return nil, nil
}

// GetAppNode get specified AppNode by namespace, name
func (s *taskGroupController) GetAppNode(ns, name string) (*v1.AppNode, error) {
	return nil, nil
}

// RegisterAppNodeHandler register event callback for AppNode
func (s *taskGroupController) RegisterAppNodeQueue(handler queue.Queue) {
	s.eventQueue = handler
	blog.Infof("taskGroup Controller starting backend goroutine for queue handling")
	go s.handleTaskGroup()
}

func (s *taskGroupController) OnAdd(obj interface{}) {
	if obj == nil {
		return
	}
	group, ok := obj.(*TaskGroup)
	if !ok {
		blog.Errorf("bk-bcs mesos TaskGroup[Pod] plugin AddEvent get error Object data")
		return
	}
	if !s.isValid(group) {
		blog.Errorf("bk-bcs TaskGroup %s/%s is invalid in UpdateEvent.", group.GetNamespace(), group.GetName())
		return
	}
	e := &taskGroupEvent{
		EventType: watch.EventAdded,
		Cur:       group,
	}
	blog.V(3).Infof("bk-bcs taskgroup %s/%s trigger Add Event", group.GetNamespace(), group.GetName())
	s.eventCh <- e
}

func (s *taskGroupController) OnUpdate(old, cur interface{}) {
	if old == nil || cur == nil {
		return
	}
	//check if old service is different with current one
	oldGroup, ok := old.(*TaskGroup)
	curGroup, cok := cur.(*TaskGroup)
	if !ok || !cok {
		blog.Errorf("bk-bcs TaskGroup[Pod] plugin AppOnUpdate got error CurrentObject type")
		return
	}
	if reflect.DeepEqual(oldGroup.Taskgroup, curGroup.Taskgroup) {
		blog.Warnf("bk-bcs TaskGroup %s/%s is nothing different on EventUpdate", curGroup.GetNamespace(), curGroup.GetName())
		return
	}
	if !s.isValid(curGroup) {
		blog.Errorf("bk-bcs TaskGroup %s/%s is invalid in UpdateEvent.", curGroup.GetNamespace(), curGroup.GetName())
		return
	}
	e := &taskGroupEvent{
		EventType: watch.EventUpdated,
		Old:       oldGroup,
		Cur:       curGroup,
	}
	blog.V(3).Infof("bk-bcs taskgroup %s/%s trigger Update Event", curGroup.GetNamespace(), curGroup.GetName())
	s.eventCh <- e
}

func (s *taskGroupController) OnDelete(obj interface{}) {
	if obj == nil {
		return
	}
	group, ok := obj.(*TaskGroup)
	if !ok {
		blog.Errorf("bk-bcs taskgroup[pod] plugin OnDelete get error Object data")
		return
	}
	e := &taskGroupEvent{
		EventType: watch.EventDeleted,
		Old:       nil,
		Cur:       group,
	}
	blog.V(3).Infof("bk-bcs taskgroup %s/%s trigger Delete Event", group.GetNamespace(), group.GetName())
	s.eventCh <- e
}

func (s *taskGroupController) isValid(task *TaskGroup) bool {
	//verify BcsService data
	if len(task.Taskgroup) == 0 {
		blog.Errorf("bk-bcs TaskGroup %s/%s lost pod info", task.GetNamespace(), task.GetName())
		return false
	}
	if !(task.Status == schetypes.TASK_STATUS_RUNNING || task.Status == schetypes.TASK_STATUS_LOST) {
		blog.Errorf("bk-bcs TaskGroup %s/%s Status[%s] lost IPAddress info, wait for RUNNING", task.GetNamespace(), task.GetName(), task.Status)
		return false
	}
	return true
}

func (s *taskGroupController) handleTaskGroup() {
	blog.Infof("bk-bcs taskgroup backgroup goroutine starting...")
	for {
		select {
		case <-s.cxt.Done():
			blog.Infof("bk-bcs TaskGroup[Pod] plugin event goroutine is asked exit.")
			return
		case event, ok := <-s.eventCh:
			if !ok {
				blog.Infof("bk-bcs TaskGroup event channel broken, ready to exit.")
				return
			}
			//convert TaskGroup to V1.AppNode
			e := &queue.Event{
				Type: event.EventType,
			}
			node := s.convertTaskGroupToAppNode(event.Cur)
			if node == nil {
				continue
			}
			node.Status.LastUpdateTime = metav1.Now()
			if node != nil {
				e.Data = node
				s.eventQueue.Push(e)
			}
		}
	}
}

func (s *taskGroupController) convertTaskGroupToAppNode(taskGroup *TaskGroup) *v1.AppNode {
	node := &v1.AppNode{
		TypeMeta: metav1.TypeMeta{
			Kind:       "appnode",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        str.ReplaceSpecialCharForAppName(taskGroup.GetName()),
			Namespace:   taskGroup.GetNamespace(),
			Labels:      str.ReplaceSpecialCharForLabel(taskGroup.GetLabels()),
			Annotations: taskGroup.GetAnnotations(),
		},
		Spec:   v1.AppNodeSpec{},
		Status: v1.AppNodeStatus{},
	}
	node.Spec.Index = taskGroup.GetName()
	//node.Spec.ProxyIP = taskGroup.HostName
	//convert ports & IP Address
	for _, task := range taskGroup.Taskgroup {
		if task.PortMappings != nil {
			ports := s.convertPortMapping(task.PortMappings)
			node.Spec.Ports = append(node.Spec.Ports, ports...)
		}
		if node.Spec.Network == "" {
			node.Spec.Network = task.Network
		}
		if len(node.Spec.NodeIP) != 0 {
			continue
		}
		if len(task.StatusData) == 0 {
			blog.Warnf("TaskGroup %s/%s ID %s lost Status data", taskGroup.GetNamespace(), taskGroup.GetName(), task.ID)
			continue
		}
		info := new(containerInfo)
		if err := json.Unmarshal([]byte(task.StatusData), info); err != nil {
			blog.Errorf("TaskGroup %s/%s decode Container %s Status data failed, %s", taskGroup.GetNamespace(), taskGroup.GetName(), task.ID, err)
			continue
		}
		if len(info.IPAddress) != 0 {
			node.Spec.NodeIP = info.IPAddress
			node.Spec.ProxyIP = info.NodeAddress
			//fix(DeveloperJim): for iterating all PortMappings
			continue
		}
		if len(info.IPAddress) == 0 && len(info.NodeAddress) != 0 {
			node.Spec.NodeIP = info.NodeAddress
			node.Spec.ProxyIP = info.NodeAddress
		}
	}
	if len(node.Spec.NodeIP) == 0 {
		blog.Errorf("bk-bcs convert TaskGroup %s/%s to AppNode finnally failed. lost node ip address, detail: %s", taskGroup.GetNamespace(), taskGroup.GetName(), taskGroup.ID)
		return nil
	}
	blog.Infof("bk-bcs convert TaskGroup %s/%s to AppNode successfluly", taskGroup.GetNamespace(), taskGroup.GetName())
	return node
}

func (s *taskGroupController) convertPortMapping(ports []*schetypes.PortMapping) (out []v1.NodePort) {
	for _, port := range ports {
		nport := v1.NodePort{
			Name:      port.Name,
			Protocol:  strings.ToLower(port.Protocol),
			NodePort:  int(port.ContainerPort),
			ProxyPort: int(port.HostPort),
		}
		out = append(out, nport)
	}
	return
}

func newTaskGroupCache(hosts []string) (*taskGroupController, error) {
	ts := k8scache.NewIndexer(TaskGroupObjectKeyFn, k8scache.Indexers{})
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
	ctl := &taskGroupController{
		cxt:          cxt,
		stopFn:       stopfn,
		eventStorage: podclient,
		localCache:   ts,
		eventCh:      make(chan *taskGroupEvent, 1024),
	}
	//create reflector
	ctl.reflector = reflector.NewReflector(fmt.Sprintf("Reflector-%s", zkConfig.Name), ts, listwatcher, time.Second*600, ctl)
	return ctl, nil
}

//TaskGroupObjectKeyFn create key for taskgroup
func TaskGroupObjectKeyFn(obj interface{}) (string, error) {
	task, ok := obj.(*TaskGroup)
	if !ok {
		return "", fmt.Errorf("error object type")
	}
	return path.Join(task.GetNamespace(), task.GetName()), nil
}

//TaskGroupObjectNewFn create new Service Object
func TaskGroupObjectNewFn() meta.Object {
	return new(TaskGroup)
}
