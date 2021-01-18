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

package apiserver

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"strconv"
	"strings"
	"time"
)

func getPublishOptions(key string, queueConf *conf.Config) (msgqueue.QueueOption, error) {
	publishDeliveryRaw := queueConf.Read(key, "PublishDelivery")
	publishDelivery, err := strconv.Atoi(publishDeliveryRaw)
	if err != nil {
		return nil, err
	}

	return msgqueue.PublishOpts(
		&msgqueue.PublishOptions{
			DeliveryMode: uint8(publishDelivery),
		}), nil
}

func getNatStreamingOptions(key string, queueConf *conf.Config) (msgqueue.QueueOption, error) {
	clusterID := queueConf.Read(key, "ClusterId")
	connectTimeoutRaw := queueConf.Read(key, "ConnectTimeout")
	connectTimeout, err := strconv.Atoi(connectTimeoutRaw)
	if err != nil {
		return nil, err
	}
	connectRetryRaw := queueConf.Read(key, "ConnectRetry")
	connectRetry, err := strconv.ParseBool(connectRetryRaw)
	if err != nil {
		return nil, err
	}

	return msgqueue.NatsOpts(
		&msgqueue.NatsOptions{
			ClusterID:      clusterID,
			ConnectTimeout: time.Duration(connectTimeout) * time.Second,
			ConnectRetry:   connectRetry,
		}), nil
}

func getQueueCommonOptions(key string, queueConf *conf.Config) (msgqueue.QueueOption, error) {
	flagRaw := queueConf.Read(key, "QueueFlag")
	kind := queueConf.Read(key, "QueueKind")

	flag, err := strconv.ParseBool(flagRaw)
	if err != nil {
		return nil, err
	}

	resource := queueConf.Read(key, "Resource")
	resourceToQueue := map[string]string{}
	arrayResource := strings.Split(resource, ",")
	for _, r := range arrayResource {
		resourceToQueue[r] = r
	}

	address := queueConf.Read(key, "Address")

	return msgqueue.CommonOpts(
		&msgqueue.CommonOptions{
			QueueFlag:       flag,
			QueueKind:       msgqueue.QueueKind(kind),
			ResourceToQueue: resourceToQueue,
			Address:         address,
		}), nil
}

func getQueueExchangeOptions(key string, queueConf *conf.Config) (msgqueue.QueueOption, error) {
	exchangeName := queueConf.Read(key, "ExchangeName")
	exchangeDurableRaw := queueConf.Read(key, "ExchangeDurable")
	exchangeDurable, err := strconv.ParseBool(exchangeDurableRaw)
	if err != nil {
		return nil, err
	}
	exchagePrefetchCountRaw := queueConf.Read(key, "ExchangePrefetchCount")
	exchagePrefetchCount, err := strconv.Atoi(exchagePrefetchCountRaw)
	if err != nil {
		return nil, err
	}
	exchangePrefetchGlobalRaw := queueConf.Read(key, "ExchangePrefetchGlobal")
	exchangePrefetchGlobal, err := strconv.ParseBool(exchangePrefetchGlobalRaw)
	if err != nil {
		return nil, err
	}

	return msgqueue.Exchange(
		&msgqueue.ExchangeOptions{
			Name:           exchangeName,
			Durable:        exchangeDurable,
			PrefetchCount:  exchagePrefetchCount,
			PrefetchGlobal: exchangePrefetchGlobal,
		}), nil

}

func getQueueSubscribeOptions(key string, queueConf *conf.Config) (msgqueue.QueueOption, error) {
	subDurableRaw := queueConf.Read(key, "SubDurable")
	subDurable, err := strconv.ParseBool(subDurableRaw)
	if err != nil {
		return nil, err
	}
	subDisableAutoAckRaw := queueConf.Read(key, "SubDisableAutoAck")
	subDisableAutoAck, err := strconv.ParseBool(subDisableAutoAckRaw)
	if err != nil {
		return nil, err
	}
	subAckOnSuccessRaw := queueConf.Read(key, "SubAckOnSuccess")
	subAckOnSuccess, err := strconv.ParseBool(subAckOnSuccessRaw)
	if err != nil {
		return nil, err
	}

	subRequeueOnErrorRaw := queueConf.Read(key, "SubRequeueOnError")
	subRequeueOnError, err := strconv.ParseBool(subRequeueOnErrorRaw)
	if err != nil {
		return nil, err
	}

	subDeliverAllMessageRaw := queueConf.Read(key, "SubDeliverAllMessage")
	subDeliverAllMessage, err := strconv.ParseBool(subDeliverAllMessageRaw)
	if err != nil {
		return nil, err
	}

	subManualAckModeRaw := queueConf.Read(key, "SubManualAckMode")
	subManualAckMode, err := strconv.ParseBool(subManualAckModeRaw)
	if err != nil {
		return nil, err
	}
	subEnableAckWaitRaw := queueConf.Read(key, "SubEnableAckWait")
	subEnableAckWait, err := strconv.ParseBool(subEnableAckWaitRaw)
	if err != nil {
		return nil, err
	}

	subAckWaitDurationRaw := queueConf.Read(key, "SubAckWaitDuration")
	subAckWaitDuration, err := strconv.Atoi(subAckWaitDurationRaw)
	if err != nil {
		return nil, err
	}

	subMaxInFlightRaw := queueConf.Read(key, "SubMaxInFlight")
	subMaxInFlight, err := strconv.Atoi(subMaxInFlightRaw)
	if err != nil {
		return nil, err
	}

	// parse queueArguments
	arguments := make(map[string]interface{})
	queueArgumentsRaw := queueConf.Read(key, "QueueArguments")
	queueArguments := strings.Split(queueArgumentsRaw, ";")
	if len(queueArguments) > 0 {
		for _, data := range queueArguments {
			dList := strings.Split(data, ":")
			if len(dList) == 2 {
				arguments[dList[0]] = dList[1]
			}
		}
	}

	return msgqueue.SubscribeOpts(
		&msgqueue.SubscribeOptions{
			DisableAutoAck:    subDisableAutoAck,
			Durable:           subDurable,
			AckOnSuccess:      subAckOnSuccess,
			RequeueOnError:    subRequeueOnError,
			DeliverAllMessage: subDeliverAllMessage,
			ManualAckMode:     subManualAckMode,
			EnableAckWait:     subEnableAckWait,
			AckWaitDuration:   time.Duration(subAckWaitDuration) * time.Second,
			MaxInFlight:       subMaxInFlight,
			QueueArguments:    arguments,
		}), nil
}