package discovery

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/pkg/meta"
	"bk-bcs/bcs-common/pkg/queue"
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
	NamespaceList(ns string, selector labels.Selector) (ret []*schedtypes.TaskGroup, err error)
	// ApplicationList list specified taskgroups by application name.namespace
	ApplicationList(ns, appname string, selector labels.Selector) (ret []*schedtypes.TaskGroup, err error)
	// Get taskgroup by the specified taskgroupID
	GetById(taskgroupId string) (*schedtypes.TaskGroup, error)
	// Get taskgroup by the specified namespace, taskgroup name
	GetByName(ns, name string) (*schedtypes.TaskGroup, error)
	// RegisterTaskgroupQueue register event callback for Taskgroup
	RegisterTaskgroupQueue(handler queue.Queue)
}

type taskgroupEvent struct {
	EventType watch.EventType
	Old       *schedtypes.TaskGroup
	Cur       *schedtypes.TaskGroup
}

//taskgroupController for dataType resource
type taskgroupController struct {
	cxt          context.Context
	stopFn       context.CancelFunc
	eventStorage storage.Storage      //remote event storage
	indexer      k8scache.Indexer     //indexer
	reflector    *reflector.Reflector //reflector list/watch all datas to local memory cache

	eventCh    chan *taskgroupEvent //event for service==>AppSvc
	eventQueue queue.Queue          //queue for event message
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

// RegisterAppNodeHandler register event callback for AppNode
func (s *taskgroupController) RegisterTaskgroupQueue(handler queue.Queue) {
	s.eventQueue = handler
	blog.Infof("taskGroup Controller starting backend goroutine for queue handling")
	go s.handleTaskGroup()
}

func (s *taskgroupController) taskOnAdd(obj interface{}) {
	if obj == nil {
		return
	}
	group, ok := obj.(*schedtypes.TaskGroup)
	if !ok {
		blog.Errorf("bk-bcs mesos TaskGroup[Pod] plugin AddEvent get error Object data")
		return
	}

	e := &taskgroupEvent{
		EventType: watch.EventAdded,
		Cur:       group,
	}
	blog.V(3).Infof("bk-bcs taskgroup %s/%s trigger Add Event", group.GetNamespace(), group.GetName())
	s.eventCh <- e
}

func (s *taskgroupController) taskOnUpdate(old, cur interface{}) {
	if old == nil || cur == nil {
		return
	}
	//check if old service is different with current one
	oldGroup, ok := old.(*schedtypes.TaskGroup)
	curGroup, cok := cur.(*schedtypes.TaskGroup)
	if !ok || !cok {
		blog.Errorf("bk-bcs TaskGroup[Pod] plugin AppOnUpdate got error CurrentObject type")
		return
	}
	if reflect.DeepEqual(oldGroup.Taskgroup, curGroup.Taskgroup) {
		blog.V(3).Infof("bk-bcs TaskGroup %s/%s is nothing different on EventUpdate", curGroup.GetNamespace(), curGroup.GetName())
		return
	}

	e := &taskgroupEvent{
		EventType: watch.EventUpdated,
		Old:       oldGroup,
		Cur:       curGroup,
	}
	blog.V(3).Infof("bk-bcs taskgroup %s/%s trigger Update Event", curGroup.GetNamespace(), curGroup.GetName())
	s.eventCh <- e
}

func (s *taskgroupController) taskOnDelete(obj interface{}) {
	if obj == nil {
		return
	}
	group, ok := obj.(*schedtypes.TaskGroup)
	if !ok {
		blog.Errorf("bk-bcs taskgroup[pod] plugin OnDelete get error Object data")
		return
	}
	e := &taskgroupEvent{
		EventType: watch.EventDeleted,
		Old:       nil,
		Cur:       group,
	}
	blog.V(3).Infof("bk-bcs taskgroup %s/%s trigger Delete Event", group.GetNamespace(), group.GetName())
	s.eventCh <- e
}

func (s *taskgroupController) handleTaskGroup() {
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
				Data: event.Cur,
			}
			s.eventQueue.Push(e)
		}
	}
}

func NewTaskgroupController(hosts []string) (TaskgroupController, error) {
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
		eventCh:      make(chan *taskgroupEvent, 1024),
	}
	handler := &reflector.EventHandler{
		AddFn:    ctl.taskOnAdd,
		UpdateFn: ctl.taskOnUpdate,
		DeleteFn: ctl.taskOnDelete,
	}
	//create reflector
	ctl.reflector = reflector.NewReflector(fmt.Sprintf("Reflector-%s", zkConfig.Name), ts, listwatcher, time.Second*600, handler)
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
