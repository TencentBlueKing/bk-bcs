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
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/bcs"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/k8s/resources"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output/action"
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
	}
)

// Writer writes the metadata to target storage service.
// There are queues for normal data and alarm message data, every
// metadata in queues would be distributed to settled handler.
type Writer struct {
	// normal metadata queue.
	queue chan *action.SyncData

	// alarm message queue.
	alarmQueue chan *action.SyncData

	// settled handlers.
	Handlers map[string]*Handler

	// alarm sender.
	alertor *action.Alertor

	// groutine stop channel.
	stopCh <-chan struct{}
}

// NewWriter creates a new Writer instance which base on bcs-storage service and alarm sender.
func NewWriter(clusterID string, storageService *bcs.InnerService, alertor *action.Alertor) (*Writer, error) {
	w := &Writer{
		queue:      make(chan *action.SyncData, defaultQueueSizeNormalMetadata),
		alarmQueue: make(chan *action.SyncData, defaultQueueSizeAlarmMetadata),
		Handlers:   make(map[string]*Handler),
		alertor:    alertor,
	}

	if err := w.init(clusterID, storageService); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Writer) init(clusterID string, storageService *bcs.InnerService) error {
	for resource := range resources.WatcherConfigList {
		action := action.NewStorageAction(clusterID, resource, storageService)
		w.Handlers[resource] = NewHandler(resource, action)
	}

	for resource := range resources.BkbcsWatcherConfigList {
		action := action.NewStorageAction(clusterID, resource, storageService)
		w.Handlers[resource] = NewHandler(resource, action)
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
	case <-time.After(defaultQueueTimeout):
		glog.Warn("can't sync data, queue timeout")
	}
}

// SyncAlarmEvent syncs alarm message data by sending into alarm queue.
func (w *Writer) SyncAlarmEvent(data *action.SyncData) {
	if data == nil {
		glog.Error("can't sync the nil alarm data")
		return
	}

	select {
	case w.alarmQueue <- data:
	case <-time.After(defaultQueueTimeout):
		glog.Warn("can't sync data, alarm queue timeout")
	}
}

// distributeNormal distributes normal metadata from queue. The distribute
// func is drived by wait.NonSlidingUntil with a stop channel, do not block to
// recv the queue here in order to make it have runtime to handle the stop channel.
func (w *Writer) distributeNormal() {
	// try to keep reading from queue until there is no more data every period.
	for {
		select {
		case data := <-w.queue:
			if handler, ok := w.Handlers[data.Kind]; ok {
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

// distributeAlarm distributes alarm metadata from alarm queue. The distribute
// func is drived by wait.NonSlidingUntil with a stop channel, do not block to
// recv the queue here in order to make it have runtime to handle the stop channel.
func (w *Writer) distributeAlarm() {
	// try to keep reading from queue until there is no more data every period.
	for {
		select {
		case data := <-w.alarmQueue:
			w.alertor.DoAlarm(data)

		case <-time.After(defaultQueueTimeout):
			// no more data, break loop.
			return
		}
	}
}

// debugs here.
func (w *Writer) debug() {
	for {
		time.Sleep(debugInterval)
		glog.Infof("Writer debug: NormalQueueLen[%d] AlarmQueueLen[%d]", len(w.queue), len(w.alarmQueue))
	}
}

// Run runs the Writer instance with target stop channel, and starts all handlers.
// There is a groutine which keep consuming metadata in queues and distributes data
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
	go wait.NonSlidingUntil(w.distributeAlarm, defaultDistributeInterval, w.stopCh)

	// setup debug.
	//go w.debug()

	return nil
}
