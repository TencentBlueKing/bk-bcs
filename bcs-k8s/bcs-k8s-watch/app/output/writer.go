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
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/bcs"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output/action"
)

// writer queue -> handler queue -> action func

type Writer struct {
	queue      chan *action.SyncData
	alarmQueue chan *action.SyncData
	stop       <-chan struct{}
	handlers   map[string]*Handler
	alertor    *action.Alertor
}

func NewWriter(clusterID string, storageService *bcs.StorageService, alertor *action.Alertor) (*Writer, error) {

	// FIXME: 1024, will stuck while there are a log of resources add/update comming
	// 2018-05-20 queue size change to 10240
	w := &Writer{
		queue:      make(chan *action.SyncData, 10240),
		handlers:   make(map[string]*Handler),
		alarmQueue: make(chan *action.SyncData, 2048),
		alertor:    alertor,
	}
	if err := w.init(clusterID, storageService); err != nil {
		return nil, err
	}
	return w, nil
}

func (writer *Writer) init(clusterID string, storageService *bcs.StorageService) error {
	resourceList := []string{"Service", "EndPoints", "Node", "Pod", "ReplicationController", "ConfigMap", "Secret", "Namespace", "Event",
		"Deployment", "DaemonSet",
		"Job", "StatefulSet",
		"Ingress", "ReplicaSet", "ExportService"}

	for _, resource := range resourceList {
		writer.handlers[resource] = &Handler{
			dataType: resource,
			// FIXME: 1024, maybe the limit
			queue: make(chan *action.SyncData, 1024),
			action: &action.StorageAction{
				Name:           resource,
				ClusterID:      clusterID,
				StorageService: storageService,
			},
		}
	}
	return nil
}

func (writer *Writer) Sync(data *action.SyncData) {
	if data == nil {
		glog.Error("Writer got nil data")
	}
	writer.queue <- data
}

func (writer *Writer) SyncAlarmEvent(data *action.SyncData) {
	if data == nil {
		glog.Error("Writer got nil alarm data")
	}
	writer.alarmQueue <- data
}

func (writer *Writer) Run(stop <-chan struct{}) {
	writer.stop = stop
	for name, handler := range writer.handlers {
		glog.Infof("Writer starting %s data channel", name)
		go handler.Run()
	}
	wait.Until(writer.route, time.Second, wait.NeverStop)
}

// Route from writer.queue To handler.queue
func (writer *Writer) route() {
	glog.Info("Writer ready to go into worker!")
	for {
		select {
		case <-writer.stop:
			glog.Info("Writer Got exit signal, ready to exit")
			return
		case syncData := <-writer.queue:
			// notify 10 item once, decrease the log amount
			// TODO: use groutine to print queue length, every 10 seconds or 1minss
			currentQueueLen := len(writer.queue)
			if currentQueueLen != 0 && currentQueueLen%10 == 0 {
				glog.Infof("Data in writer's queue: %d", currentQueueLen)
			}

			// FIXME: 如果某个handler的channel stuck了, 则这里会stuck
			if handler, ok := writer.handlers[syncData.Kind]; ok {
				handler.Handle(syncData)
			} else {
				glog.Errorf("Got unknown DataType: %s", syncData.Kind)
			}
		case syncData := <-writer.alarmQueue:
			writer.alertor.DoAlarm(syncData)
		}
	}

}
