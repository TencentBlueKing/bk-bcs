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

package config

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
)

// AlertManagerOptions parse command-line parameters to options
type AlertManagerOptions struct {
	conf.FileConfig
	registry.CMDOptions `json:"etcdOptions"`
	conf.ServiceConfig  `json:"serviceOptions"`
	conf.LogConfig      `json:"logOptions"`
	conf.MetricConfig   `json:"metricOptions"`
	conf.CertConfig     `json:"certOptions"`

	SwaggerConfigDir   SwaggerConfig      `json:"swaggerConfigDir"`
	AlertServerOptions AlertServerOptions `json:"alertServerOptions"`
	DebugMode          bool               `json:"debug_mode" value:"false" usage:"Debug mode, use pprof."`
	HandlerConfig      HandlerOptions     `json:"handler_config"`

	ResourceSubs []ResourceSubType `json:"resourceSubs" value:"" usage:"ResourceSubs consumer"`
	QueueConfig  QueueConfig       `json:"queue_config"`
}

// QueueConfig option for queue
type QueueConfig struct {
	// commonOpts
	QueueFlag bool   `json:"queueFlag"`
	QueueKind string `json:"queueKind"`
	Resource  string `json:"resource"`
	Address   string `json:"address"`

	// exchangeOpts
	ExchangeName           string `json:"exchangeName"`
	ExchangeDurable        bool   `json:"exchangeDurable"`
	ExchangePrefetchCount  int    `json:"exchangePrefetchCount"`
	ExchangePrefetchGlobal bool   `json:"exchangePrefetchGlobal"`

	// nats-streaming
	ClusterID      string `json:"clusterID"`
	ConnectTimeout int    `json:"connectTimeout"`
	ConnectRetry   bool   `json:"connectRetry"`

	// publishOpts
	PublishDelivery int `json:"publishDelivery"`

	// subscribeOpts
	SubDurable           bool                   `json:"subDurable"`
	SubDisableAutoAck    bool                   `json:"subDisableAutoAck"`
	SubAckOnSuccess      bool                   `json:"subAckOnSuccess"`
	SubRequeueOnError    bool                   `json:"subRequeueOnError"`
	SubDeliverAllMessage bool                   `json:"subDeliverAllMessage"`
	SubManualAckMode     bool                   `json:"subManualAckMode"`
	SubEnableAckWait     bool                   `json:"subEnableAckWait"`
	SubAckWaitDuration   int                    `json:"subAckWaitDuration"`
	SubMaxInFlight       int                    `json:"subMaxInFlight"`
	QueueArguments       map[string]interface{} `json:"queueArguments"`
}

// SwaggerConfig option for swagger
type SwaggerConfig struct {
	Dir string `json:"dir"`
}

// AlertServerOptions option for alert-system server
type AlertServerOptions struct {
	Server      string `json:"server"`
	AppCode     string `json:"appCode"`
	AppSecret   string `json:"appSecret"`
	ServerDebug bool   `json:"debugLevel"`
}

// ResourceSubType for subscribe handler type switch
type ResourceSubType struct {
	Switch   string `json:"switch"`
	Category string `json:"category"`
}

// HandlerOptions for all handler options
type HandlerOptions struct {
	EventHandlerOptions
}

// EventHandlerOptions for event handler options
type EventHandlerOptions struct {
	ConcurrencyNum     int  `json:"concurrencyNum"`
	AlertEventNum      int  `json:"alertEventNum"`
	ChanQueueNum       int  `json:"chanQueueNum"`
	IsBatchAggregation bool `json:"isBatchAggregation"`
}

// NewAlertManagerOptions create AlertManagerOptions object
func NewAlertManagerOptions() *AlertManagerOptions {
	return &AlertManagerOptions{}
}
