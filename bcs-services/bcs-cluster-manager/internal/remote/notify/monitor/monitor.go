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

// Package monitor xxx
package monitor

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/notify"
)

// 通过bk_monitor蓝鲸监控实现自定义metrics/事件上报

// NewMetricsNotify init metrics notify
func NewMetricsNotify(config notify.MessageServer) *MetricsNotify {
	return &MetricsNotify{basicConfig{
		server:      config.Server,
		dataId:      config.DataId,
		accessToken: config.AccessToken,
	}}
}

// MetricsNotify metrics server
type MetricsNotify struct {
	basicConfig
}

// Notify bk_monitor metrics
func (mn *MetricsNotify) Notify(ctx context.Context, message notify.MessageBody) error {
	server := serverConfig{
		basicConfig: basicConfig{
			server:      mn.server,
			dataId:      mn.dataId,
			accessToken: mn.accessToken,
		},
		customType: Metrics,
		debug:      false,
	}

	return pushCustomDataToServer(server, []*eventMetricsBody{
		{
			metrics:   message.Metrics,
			dimension: message.Dimension,
		},
	})
}

// NewEventsNotify init events notify
func NewEventsNotify(config notify.MessageServer) *EventsNotify {
	return &EventsNotify{basicConfig{
		server:      config.Server,
		dataId:      config.DataId,
		accessToken: config.AccessToken,
	}}
}

// EventsNotify events server
type EventsNotify struct {
	basicConfig
}

// Notify bk_monitor events
func (mn *EventsNotify) Notify(ctx context.Context, message notify.MessageBody) error {
	server := serverConfig{
		basicConfig: basicConfig{
			server:      mn.server,
			dataId:      mn.dataId,
			accessToken: mn.accessToken,
		},
		customType: Event,
		debug:      false,
	}

	return pushCustomDataToServer(server, []*eventMetricsBody{
		{
			eventName: message.EventName,
			eventBody: message.EventBody,
			dimension: message.Dimension,
		},
	})
}

// pushCustomDataToServer push custom data to server
func pushCustomDataToServer(server serverConfig, customData []*eventMetricsBody) error {
	const (
		path = "/v2/push/"
	)
	if len(server.server) == 0 {
		return fmt.Errorf("server empty")
	}

	if server.dataId == 0 || len(server.accessToken) == 0 {
		return fmt.Errorf("dataId or accessToken error")
	}
	if server.timeout <= 0 {
		server.timeout = time.Second * 60
	}

	reportItem := &ReportItems{
		DataId:      server.dataId,
		AccessToken: server.accessToken,
		Data:        make([]*Item, 0),
	}

	host, _ := os.Hostname()
	switch server.customType {
	case Metrics:
		for i := range customData {
			reportItem.Data = append(reportItem.Data, &Item{
				Target:    host,
				Metrics:   customData[i].metrics,
				Dimension: customData[i].dimension,
				Timestamp: time.Now().UnixMilli(),
			})
		}
	case Event:
		for i := range customData {
			reportItem.Data = append(reportItem.Data, &Item{
				Target:    host,
				EventName: customData[i].eventName,
				Event:     EventItem{Content: customData[i].eventBody},
				Dimension: customData[i].dimension,
				Timestamp: time.Now().UnixMilli(),
			})
		}
	default:
		return fmt.Errorf("not supported customType %s", server.customType)
	}

	var respData ReportStatus

	start := time.Now()
	resp, _, errs := gorequest.New().
		Timeout(server.timeout).
		Post(server.server+path).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		SetDebug(server.debug).
		Send(reportItem).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("notify", "pushCustomDataToServer", "http", metrics.LibCallStatusErr, start)
		return errs[0]
	}
	metrics.ReportLibRequestMetric("notify", "pushCustomDataToServer", "http", metrics.LibCallStatusOK, start)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http request err: %v:%v", resp.StatusCode, resp.Status)
	}

	if respData.Result != "true" {
		return fmt.Errorf("requestId[%s] error: %v", respData.RequestID, respData.Message)
	}

	return nil
}
