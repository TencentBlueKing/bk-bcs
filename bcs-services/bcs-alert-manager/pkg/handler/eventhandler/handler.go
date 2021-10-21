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

package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/remote/alert"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/remote/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/utils/concurrency"

	v1 "k8s.io/api/core/v1"
)

const (
	// ResourceType event
	ResourceType = "Event"
)

const (
	handlerName       = "eventhandler"
	handlerHandleFunc = "HandleQueueEvent"
)

// event metaInfo
const (
	clusterIDTag    = "clusterId"
	nameSpaceTag    = "namespace"
	resourceKindTag = "resourceKind"
	resourceNameTag = "resourceName"
	resourceTypeTag = "resourceType" // event
	levelTag        = "level"        // Normal, Warning
	typeTag         = "type"         // event.Reason

	clusterIDKey    = "cluster_id"
	nameSpaceKey    = "namespace"
	resourceKindKey = "resource_kind"
	resourceNameKey = "resource_name"
	resourceTypeKey = "resource_type"
	levelKey        = "level"
	typeKey         = "type"
)

func generateAlertName(labels map[string]string) string {
	return fmt.Sprintf("集群[%s]命名空间[%s]资源[%s-%s]产生类型[%s]",
		labels[clusterIDKey], labels[nameSpaceKey], labels[resourceKindKey], labels[resourceNameKey],
		labels[typeKey])
}

var eventFeatTagsMapLabelKeys = map[string]string{
	clusterIDTag:    clusterIDKey,
	nameSpaceTag:    nameSpaceKey,
	resourceKindTag: resourceKindKey,
	resourceNameTag: resourceNameKey,
	resourceTypeTag: resourceTypeKey,
	levelTag:        levelKey,
	typeTag:         typeKey,
}

// SyncEventHandler event handler
type SyncEventHandler struct {
	unSub              func()
	stopCtx            context.Context
	stopCancel         context.CancelFunc
	alertClient        alert.BcsAlarmInterface
	eventListCh        chan msgqueue.HandlerData
	filters            []msgqueue.Filter
	alertBarrier       *concurrency.Concurrency
	alertBatchNum      int
	isBatchAggregation bool
}

// Options for eventHandler
type Options struct {
	IsBatchAggregation bool
	AlertEventBatchNum int
	ConcurrencyNum     int
	ChanQueueNum       int
	Client             alert.BcsAlarmInterface
}

// NewSyncEventHandler create event handler object
func NewSyncEventHandler(opt Options) *SyncEventHandler {
	ctx, cancel := context.WithCancel(context.Background())

	return &SyncEventHandler{
		stopCtx:            ctx,
		stopCancel:         cancel,
		alertClient:        opt.Client,
		eventListCh:        make(chan msgqueue.HandlerData, opt.ChanQueueNum),
		alertBarrier:       concurrency.NewConcurrency(opt.ConcurrencyNum),
		alertBatchNum:      opt.AlertEventBatchNum,
		isBatchAggregation: opt.IsBatchAggregation,
	}
}

func (eh *SyncEventHandler) backgroundBatchSync() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	alertEventList := []msgqueue.HandlerData{}
	for {
		select {
		case <-eh.stopCtx.Done():
			glog.Info("backgroundBatchSync has been stopped")
			return
		case event := <-eh.eventListCh:
			alertEventList = append(alertEventList, event)
			if len(alertEventList) < eh.alertBatchNum {
				continue
			}
		case <-ticker.C:
		}

		if len(alertEventList) == 0 {
			continue
		}

		eh.alertBarrier.Add()
		go func(eventList []msgqueue.HandlerData) {
			defer eh.alertBarrier.Done()
			defer func() {
				if r := recover(); r != nil {
					glog.Errorf("[monitor][event] panic: %v\n", r)
				}
			}()

			alertReqDataList := eh.transEventListToAlertData(eventList)
			if len(alertReqDataList) == 0 {
				return
			}

			//fmt.Println(alertReqDataList)
			err := eh.alertClient.SendAlarmInfoToAlertServer(alertReqDataList, time.Second*10)
			if err != nil {
				glog.Errorf("event handler backgroundSync sendEvenDataToAlert failed: %v", err)
			}

		}(alertEventList)

		alertEventList = []msgqueue.HandlerData{}
	}
}

func (eh *SyncEventHandler) transEventMetaToAlertData(eventMeta msgqueue.HandlerData) (map[string]string, map[string]string) {
	if eh == nil {
		return nil, nil
	}

	if len(eventMeta.Body) == 0 {
		return nil, nil
	}

	event := &v1.Event{}
	err := json.Unmarshal(eventMeta.Body, event)
	if err != nil {
		return nil, nil
	}

	if len(event.Message) == 0 {
		return nil, nil
	}

	labels := map[string]string{}
	for tagKey, labelKey := range eventFeatTagsMapLabelKeys {
		if val, ok := eventMeta.Meta[tagKey]; ok {
			labels[labelKey] = val
		}
	}

	labels[string(alert.AlarmLabelsAlarmName)] = generateAlertName(labels)
	// defaultProjectID parse by clusterID
	// labels[string(alert.AlarmLabelsAlarmProjectID)] = alert.DefaultAlarmProjectID

	// parse event metadata name
	annotations := map[string]string{
		string(alert.AlarmAnnotationsUUID):   event.ObjectMeta.Namespace + "/" + event.ObjectMeta.Name,
		string(alert.AlarmAnnotationsBody):   event.Message,
		string(alert.AlarmAnnotationsHostIP): event.Source.Host,
	}

	return annotations, labels
}

func (eh *SyncEventHandler) transEventListToAlertData(eventList []msgqueue.HandlerData) []alert.AlarmReqData {
	alarmDataList := []alert.AlarmReqData{}

	if len(eventList) == 0 {
		return alarmDataList
	}

	for i := range eventList {
		if len(eventList[i].Body) == 0 {
			continue
		}

		// parse event meta & body
		annotations, labels := eh.transEventMetaToAlertData(eventList[i])
		if annotations == nil || labels == nil {
			continue
		}

		alarmDataList = append(alarmDataList, alert.AlarmReqData{
			StartsTime:  time.Now(),
			Annotations: annotations,
			Labels:      labels,
		})
	}

	return alarmDataList
}

func (eh *SyncEventHandler) backgroundSync() {

	for eventData := range eh.eventListCh {
		select {
		case <-eh.stopCtx.Done():
			glog.Info("backgroundSync has been stopped")
			return
		default:
		}

		eh.alertBarrier.Add()
		go func(event msgqueue.HandlerData) {
			defer eh.alertBarrier.Done()
			defer func() {
				if r := recover(); r != nil {
					glog.Errorf("[monitor][event] panic: %v\n", r)
				}
			}()
			err := eh.sendEvenDataToAlert(event)
			if err != nil {
				glog.Errorf("event handler backgroundSync sendEvenDataToAlert failed: %v", err)
			}
		}(eventData)
	}
}

func (eh *SyncEventHandler) sendEvenDataToAlert(event msgqueue.HandlerData) error {

	if len(event.Body) == 0 {
		return nil
	}
	// parse event meta & body
	annotations, labels := eh.transEventMetaToAlertData(event)
	if annotations == nil || labels == nil {
		errMsg := fmt.Errorf("sendEvenDataToAlert annotation || labels is nil")
		return errMsg
	}

	data := []alert.AlarmReqData{
		{
			StartsTime:  time.Now(),
			Annotations: annotations,
			Labels:      labels,
		},
	}

	err := eh.alertClient.SendAlarmInfoToAlertServer(data, time.Second*10)
	if err != nil {
		return err
	}

	return nil
}

// HandleQueueEvent register queue for event callback
func (eh *SyncEventHandler) HandleQueueEvent(ctx context.Context, data []byte) error {
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("event handle queueEvent panic: %v\n", r)
		}
	}()

	started := time.Now()

	select {
	case <-ctx.Done():
	case <-eh.stopCtx.Done():
		glog.Errorf("event handler has been closed")
		metrics.ReportHandlerFuncLatency(handlerName, handlerHandleFunc, metrics.SucStatus, started)
		return nil
	default:
	}
	eventHandlerData := &msgqueue.HandlerData{}
	err := json.Unmarshal(data, eventHandlerData)
	if err != nil {
		glog.Errorf("Unmarshal event handler data failed: %v", err)
		metrics.ReportHandlerFuncLatency(handlerName, handlerHandleFunc, metrics.ErrStatus, started)
		return err
	}

	if !validateResourceType(eventHandlerData.ResourceType) {
		metrics.ReportHandlerFuncLatency(handlerName, handlerHandleFunc, metrics.SucStatus, started)
		return nil
	}

	select {
	case eh.eventListCh <- *eventHandlerData:
	case <-time.After(time.Second * 1):
		glog.Info("handle queue event has been discarded")
	}
	metrics.ReportHandlerFuncLatency(handlerName, handlerHandleFunc, metrics.SucStatus, started)

	return nil
}

func validateResourceType(resourceType string) bool {
	if resourceType != ResourceType {
		return false
	}

	return true
}

// Consume subscribe Event queue & backgroundSync
func (eh *SyncEventHandler) Consume(ctx context.Context, sub msgqueue.MessageQueue) error {
	unSub, _ := sub.Subscribe(msgqueue.HandlerWrap("event-handler", eh.HandleQueueEvent), eh.filters, msgqueue.EventSubscribeType)

	eh.unSub = func() {
		unSub.Unsubscribe()
	}
	glog.Infof("SyncEventHandler backgroundBatchSync switch: %v", eh.isBatchAggregation)

	if eh.isBatchAggregation {
		go eh.backgroundBatchSync()
	} else {
		go eh.backgroundSync()
	}

	go eh.monitor()

	return nil
}

func (eh *SyncEventHandler) monitor() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-eh.stopCtx.Done():
			return
		case <-ticker.C:
		}

		// report chan queue length
		metrics.ReportHandlerQueueLength(handlerName, float64(len(eh.eventListCh)))
	}
}

// Stop close chanQueue & unSub
func (eh *SyncEventHandler) Stop() error {
	eh.unSub()
	eh.stopCancel()
	close(eh.eventListCh)
	time.Sleep(time.Second * 3)

	return nil
}
