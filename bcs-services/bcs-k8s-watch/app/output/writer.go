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

package output

import (
	"errors"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/k8s/resources"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/output/action"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/pkg/metrics"
)

const (
	// defaultQueueSizeNormalMetadata is default queue size of Writer for normal metadata.
	defaultQueueSizeNormalMetadata = 10 * 1024

	// defaultQueueSizeAlarmMetadata is default queue size of Writer for alarm metadata.
	defaultQueueSizeAlarmMetadata = 2 * 1024

	// defaultQueueTimeout is default timeout of queue.
	defaultQueueTimeout = 1 * time.Second

	// defaultDistributeInterval is default interval of distribution.
	defaultDistributeInterval = 500 * time.Millisecond

	// debugInterval is interval of debug.
	debugInterval = 10 * time.Second

	// defaultQueueNum is default queue num for Pod kind
	defaultQueueNum = 10
)

const (
	// NormalQueue for normalQueue handlerLabel
	NormalQueue = "writer_normal_queue"
)

const (
	// Pod pod resource
	Pod = "Pod"
	// PodPrefix queue key prefix
	PodPrefix = "Pod_"
	// Event event resource
	Event = "Event"
	// EventPrefix queue key prefix
	EventPrefix = "Event_"
)

var (
	// writerResources is resource list could be handled by the writer.
	writerResources = []string{
		"Service",
		"EndPoints",
		"Node",
		"Pod",
		"ReplicationController",
		"ConfigMap",
		"Secret",
		"Namespace",
		"Event",
		"Deployment",
		"DaemonSet",
		"Job",
		"StatefulSet",
		"Ingress",
		"ReplicaSet",
		"ExportService",
		"BcsLogConfig",
		"BcsDbPrivConfig",
		"GameDeployment",
		"GameStatefulSet",
	}
)

// ResourceQueueDistributeNum for resource queueNum
type resourceQueueDistributeNum struct {
	// PodChanQueueNum kind pod queueNum
	podChanQueueNum int
}

// Writer writes the metadata to target storage service.
// There are queues for normal data and alarm message data, every
// metadata in queues would be distributed to settled handler.
type Writer struct {
	// clusterID
	clusterID string
	// normal metadata queue.
	queue chan *action.SyncData

	// settled handlers.
	Handlers map[string]*Handler

	// getResourceName get resourceName by data
	getResourceName func(data *action.SyncData) string
	// resourceQueueNum for resource queueNum
	resourceQueueNum resourceQueueDistributeNum
	// goroutine stop channel.
	stopCh <-chan struct{}
}

// NewWriter creates a new Writer instance which base on bcs-storage service and alarm sender.
func NewWriter(clusterID string, storageService *bcs.InnerService, bcsConfig options.BCSConfig) (*Writer, error) {
	var writerQueueLength int64 = defaultQueueSizeNormalMetadata
	if bcsConfig.WriterQueueLen > defaultQueueSizeNormalMetadata {
		writerQueueLength = bcsConfig.WriterQueueLen
	}

	w := &Writer{
		queue:     make(chan *action.SyncData, writerQueueLength),
		Handlers:  make(map[string]*Handler),
		clusterID: clusterID,
		resourceQueueNum: resourceQueueDistributeNum{
			podChanQueueNum: bcsConfig.PodQueueNum,
		},
		getResourceName: getResourceDataName,
	}

	if err := w.init(clusterID, storageService); err != nil {
		return nil, err
	}
	return w, nil
}

// initWatcherResourceDistributeQueue init resource extra distribute queue according to w.resourceQueueNum
func (w *Writer) initWatcherResourceDistributeQueue(clusterID string, resource string, action *action.StorageAction) {
	switch resource {
	case Pod:
		if w.resourceQueueNum.podChanQueueNum > 0 {
			glog.Infof("resource %s create %d handlerQueue", Pod, w.resourceQueueNum.podChanQueueNum)
			for i := 0; i < w.resourceQueueNum.podChanQueueNum; i++ {
				handlerChanKey := PodPrefix + strconv.Itoa(i)
				w.Handlers[handlerChanKey] = NewHandler(clusterID, handlerChanKey, action)
			}
		}
	case Event:
		// 4 times of podChanQueueNum for event
		glog.Infof("resource %s create %d handlerQueue", Event, 4*w.resourceQueueNum.podChanQueueNum)
		for i := 0; i < 4*w.resourceQueueNum.podChanQueueNum; i++ {
			handlerChanKey := EventPrefix + strconv.Itoa(i)
			w.Handlers[handlerChanKey] = NewHandler(clusterID, handlerChanKey, action)
		}
	default:
	}
}

func (w *Writer) init(clusterID string, storageService *bcs.InnerService) error {
	for resource := range resources.WatcherConfigList {
		action := action.NewStorageAction(clusterID, resource, storageService)
		w.Handlers[resource] = NewHandler(clusterID, resource, action)
		w.initWatcherResourceDistributeQueue(clusterID, resource, action)
	}

	for resource := range resources.BkbcsWatcherConfigList {
		action := action.NewStorageAction(clusterID, resource, storageService)
		w.Handlers[resource] = NewHandler(clusterID, resource, action)
	}
	return nil
}

// Sync syncs normal metadata by sending into queue.
func (w *Writer) Sync(data *action.SyncData) {
	if data == nil {
		glog.Warn("can't sync the nil data")
		return
	}

	select {
	case w.queue <- data:
		metrics.ReportK8sWatchHandlerQueueLengthInc(w.clusterID, NormalQueue)
	case <-time.After(defaultQueueTimeout):
		metrics.ReportK8sWatchHandlerDiscardEvents(w.clusterID, NormalQueue)
		glog.Warn("can't sync data, queue timeout")
	}
}

// distributeNormal distributes normal metadata from queue. The distribute
// func is invoked by wait.NonSlidingUntil with a stop channel, do not block to
// recv the queue here in order to make it have runtime to handle the stop channel.
func (w *Writer) distributeNormal() {
	// try to keep reading from queue until there is no more data every period.
	for {
		select {
		case data := <-w.queue:
			metrics.ReportK8sWatchHandlerQueueLengthDec(w.clusterID, NormalQueue)
			// observe writer queue length
			if len(w.queue)+1024 > cap(w.queue) {
				glog.Warnf("Writer queue is busy, current task queue(%d/%d)", len(w.queue), cap(w.queue))
			} else {
				glog.V(3).Infof("write queue receive task, current queue(%d/%d)", len(w.queue), cap(w.queue))
			}

			handlerKey := w.GetHandlerKeyBySyncData(data)
			if handler, ok := w.Handlers[handlerKey]; ok {
				handler.HandleWithTimeout(data, defaultQueueTimeout)
			} else {
				glog.Errorf("can't distribute the normal metadata, unknown DataType[%+v]", data.Kind)
			}

		case <-time.After(defaultQueueTimeout):
			// no more data, break loop.
			return
		}
	}
}

// GetHandlerKeyBySyncData returns the handler key by sync data
func (w *Writer) GetHandlerKeyBySyncData(data *action.SyncData) string {
	if w == nil || data == nil {
		return ""
	}

	// default handlerKey
	handlerKey := data.Kind
	switch data.Kind {
	case Pod:
		resourceName := w.getResourceName(data)
		if len(resourceName) > 0 {
			index := getHashId(resourceName, w.resourceQueueNum.podChanQueueNum)
			if index >= 0 {
				handlerKey = PodPrefix + strconv.Itoa(index)
			}
			glog.V(5).Infof("Pod resource[%s], handlerKey[%d: %s]", resourceName, index, handlerKey)
		}
	case Event:
		resourceName := w.getResourceName(data)
		if len(resourceName) > 0 {
			index := getHashId(resourceName, 4*w.resourceQueueNum.podChanQueueNum)
			if index >= 0 {
				handlerKey = EventPrefix + strconv.Itoa(index)
			}
			glog.V(5).Infof("Event resource[%s], handlerKey[%d: %s]", resourceName, index, handlerKey)
		}
	default:
	}

	return handlerKey
}

// debugs here.
func (w *Writer) debug() {
	for {
		time.Sleep(debugInterval)
		glog.Infof("Writer debug: NormalQueueLen[%d] AlarmQueueLen[%d]", len(w.queue))
	}
}

// reportQueueLength report writer module queueInfo to prometheus metrics
func (w *Writer) reportWriterQueueLength() {
	metrics.ReportK8sWatchHandlerQueueLength(w.clusterID, NormalQueue, float64(len(w.queue)))
}

// Run runs the Writer instance with target stop channel, and starts all handlers.
// There is a goroutine which keep consuming metadata in queues and distributes data
// to settled handler until stop channel is activated.
func (w *Writer) Run(stopCh <-chan struct{}) error {
	if stopCh != nil {
		w.stopCh = stopCh
	}

	if w.stopCh == nil {
		return errors.New("can't run the writer with nil stop channel")
	}

	// start all handlers.
	for _, handler := range w.Handlers {
		handler.Run(stopCh)
	}

	// keep consuming metadata from queues.
	glog.Info("Writer keeps consuming/distributing metadata now")
	go wait.NonSlidingUntil(w.distributeNormal, defaultDistributeInterval, w.stopCh)

	// report writer module queueLen metrics
	go wait.Until(w.reportWriterQueueLength, defaultHandlerReportPeriod, w.stopCh)
	// setup debug.
	//go w.debug()

	return nil
}

// getResourceDataName get resource name by SyncData
func getResourceDataName(data *action.SyncData) string {
	if data == nil {
		return ""
	}
	if len(data.Namespace) > 0 {
		return data.Namespace + "/" + data.Name
	}
	return data.Name
}

// getHashId get string hashID for distribute to same queue according to hashID
// if queueNum maxInt <= 0, use source queue
// if queueNum maxInt > 0 , distribute same resource to the same queue according to handID
func getHashId(s string, maxInt int) int {
	if maxInt <= 0 {
		return -1
	}

	seed := 131
	hash := 0
	char := []byte(s)

	for _, c := range char {
		hash = hash*seed + int(c)
	}

	return (hash & 0x7FFFFFFF) % maxInt
}
