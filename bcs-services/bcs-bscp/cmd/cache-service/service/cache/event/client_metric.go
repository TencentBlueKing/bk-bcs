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

// Package event handle client metric
package event

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/cache-service/service/cache/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/bedis"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/dao"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbclient "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client"
	pbce "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client-event"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
)

const (
	defaultClientMetricTaskInterval = 10 * time.Second
	clientMetricPattern             = `*bscp:client-metric:*`
	clientMetricKey                 = `\{\d+\}bscp:client-metric:\d+`
)

// ClientMetric xxx
type ClientMetric struct {
	set   dao.Set
	state serviced.State
	bds   bedis.Client
	op    client.Interface
}

// NewClientMetric init client metric
func NewClientMetric(set dao.Set, state serviced.State, bds bedis.Client, op client.Interface) ClientMetric {
	return ClientMetric{
		set:   set,
		state: state,
		bds:   bds,
		op:    op,
	}
}

// Run the client metric task
func (cm *ClientMetric) Run() {
	logs.Infof("start client metric task")
	notifier := shutdown.AddNotifier()
	go func() {
		ticker := time.NewTicker(defaultClientMetricTaskInterval)
		defer ticker.Stop()
		for {
			kt := kit.New()
			ctx, cancel := context.WithCancel(kt.Ctx)
			kt.Ctx = ctx

			select {
			case <-notifier.Signal:
				logs.Infof("stop handle client metric data success")
				cancel()
				notifier.Done()
				return
			case <-ticker.C:
				logs.Infof("start handle client metric data")

				if !cm.state.IsMaster() {
					logs.V(2).Infof("this is slave, do not need to handle, skip. rid: %s", kt.Rid)
					time.Sleep(sleepTime)
					continue
				}
				cm.consumeClientMetricData(kt)
			}
		}
	}()
}

// 消费队列中的 client metric 数据
func (cm *ClientMetric) consumeClientMetricData(kt *kit.Kit) {
	// 先获取符合规则的key
	keys, err := cm.matchKeys()
	if err != nil {
		logs.Errorf("the KEY is not matched, err: %s, rid: %s", err.Error(), kt.Rid)
		return
	}
	if len(keys) == 0 {
		logs.V(2).Infof("there is no matching KEY, rid: %s", kt.Rid)
		return
	}
	for _, key := range keys {
		lLen, err := cm.bds.LLen(kt.Ctx, key)
		if err != nil {
			logs.Errorf("get key: %s list length failed, err: %s", key, err.Error())
			continue
		}
		if lLen != 0 {
			cm.getClientMetricList(kt, key, lLen)
		}
	}
}

// 获取 client metric 数据列表
func (cm *ClientMetric) getClientMetricList(kt *kit.Kit, key string, listLen int64) {
	batchSize := 1000
	for i := 0; i < int(listLen); i += batchSize {
		startIndex := int64(i)
		endIndex := int64(i + batchSize - 1)
		if endIndex >= listLen {
			endIndex = listLen - 1
		}
		list, err := cm.bds.LRange(kt.Ctx, key, startIndex, endIndex)
		if err != nil {
			logs.Errorf("get key: %s  %v to %v client metric data failed, rid: %s, err: %s ", key,
				startIndex, endIndex, kt.Rid, err.Error())
			continue
		}
		err = cm.handleClientMetricData(kt, list)
		if err != nil {
			logs.Errorf("handle client metric data failed, rid: %s, err: %s", kt.Rid, err.Error())
			continue
		}

		_, err = cm.bds.LTrim(kt.Ctx, key, endIndex+1, -1)
		if err != nil {
			logs.Errorf("delete the Specify keys values data failed, key: %s, rid: %s, err: %s", key, kt.Rid, err.Error())
			continue
		}
	}

}

// 处理 client metric 数据
// client 表是按照 业务+服务+客户端 维度：数据做聚合
// 多条心跳把每条每一列中的最大值取出来，组合成一条
// 多条变更数据只需要最后一条
// client event 表是按照 业务+服务+事件ID 维度：数据做聚合
func (cm *ClientMetric) handleClientMetricData(kt *kit.Kit, payload []string) error { // nolint
	vc := new(sfs.VersionChangePayload)
	hb := new(sfs.HeartbeatItem)
	clientData := make([]*pbclient.Client, 0)
	clientEventData := make([]*pbce.ClientEvent, 0)

	vcClientEvent := map[string]*pbce.ClientEvent{}
	hbClientEvent := map[string]*pbce.ClientEvent{}

	hbClient := map[string]*pbclient.Client{}
	vcClient := map[string]*pbclient.Client{}

	maxResourceUsageValues := make(map[string]*pbclient.ClientResource)

	clientMetricData := sfs.ClientMetricData{}
	for _, v := range payload {
		err := jsoni.Unmarshal([]byte(v), &clientMetricData)
		if err != nil {
			return err
		}
		switch sfs.MessagingType(clientMetricData.MessagingType) {
		case sfs.Heartbeat:
			err = jsoni.Unmarshal(clientMetricData.Payload, hb)
			if err != nil {
				return err
			}

			hb.Application.AppID = clientMetricData.AppID
			clientMetric, errHb := hb.PbClientMetric()
			if errHb != nil {
				return errHb
			}
			if clientMetric == nil {
				continue
			}
			key := fmt.Sprintf("%d-%d-%s", clientMetric.Attachment.BizId,
				clientMetric.Attachment.AppId, clientMetric.Attachment.Uid)
			// 如果 key 已存在，比较并更新最大值
			if existing, ok := maxResourceUsageValues[key]; ok {
				if clientMetric.Spec.Resource.CpuMaxUsage > existing.CpuMaxUsage {
					existing.CpuMaxUsage = clientMetric.Spec.Resource.CpuMaxUsage
				}
				if clientMetric.Spec.Resource.CpuUsage > existing.CpuUsage {
					existing.CpuUsage = clientMetric.Spec.Resource.CpuUsage
				}
				if clientMetric.Spec.Resource.MemoryMaxUsage > existing.MemoryMaxUsage {
					existing.MemoryMaxUsage = clientMetric.Spec.Resource.MemoryMaxUsage
				}
				if clientMetric.Spec.Resource.MemoryUsage > existing.MemoryUsage {
					existing.MemoryUsage = clientMetric.Spec.Resource.MemoryUsage
				}
				maxResourceUsageValues[key] = existing
			} else {
				maxResourceUsageValues[key] = &pbclient.ClientResource{
					CpuMaxUsage:    clientMetric.Spec.Resource.CpuMaxUsage,
					CpuUsage:       clientMetric.Spec.Resource.CpuUsage,
					MemoryUsage:    clientMetric.Spec.Resource.MemoryUsage,
					MemoryMaxUsage: clientMetric.Spec.Resource.MemoryMaxUsage,
				}
			}
			clientMetric.Spec.Resource = maxResourceUsageValues[key]
			hbClient[key] = clientMetric
			// 处理clientEvent数据
			clientEventMetric, ceErr := hb.PbClientEventMetric()
			if ceErr != nil {
				return ceErr
			}
			hbClientEvent = lastClientEventData(clientEventMetric, hbClientEvent)
		case sfs.VersionChangeMessage:
			err = vc.Decode(clientMetricData.Payload)
			if err != nil {
				return err
			}
			vc.Application.AppID = clientMetricData.AppID
			clientMetric, errCeVc := vc.PbClientMetric()
			if errCeVc != nil {
				return errCeVc
			}
			vcClient = lastClientData(clientMetric, vcClient)
			clientEventMetric, errVc := vc.PbClientEventMetric()
			if errVc != nil {
				return errVc
			}
			vcClientEvent = lastClientEventData(clientEventMetric, vcClientEvent)
		}
	}

	for _, v := range vcClient {
		clientData = append(clientData, v)
	}
	for _, v := range hbClient {
		clientData = append(clientData, v)
	}
	for _, v := range vcClientEvent {
		clientEventData = append(clientEventData, v)
	}
	for _, v := range hbClientEvent {
		clientEventData = append(clientEventData, v)
	}

	err := cm.op.BatchUpsertClientMetrics(kt, clientData, clientEventData)
	if err != nil {
		logs.Errorf("batch upsert client metrics failed, rid: %s, err: %s", kt.Rid, err.Error())
		return err
	}

	return nil
}

// matchKeys xxx
func (cm *ClientMetric) matchKeys() ([]string, error) {
	kt := kit.New()
	keys, err := cm.bds.Keys(kt.Ctx, clientMetricPattern)
	if err != nil {
		return nil, err
	}
	// 再次过滤
	keys, err = filterKeysByRegex(keys, clientMetricKey)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// filterKeysByRegex 使用正则表达式筛选符合规则的键
func filterKeysByRegex(keys []string, pattern string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var matchingKeys []string
	for _, key := range keys {
		if re.MatchString(key) {
			matchingKeys = append(matchingKeys, key)
		}
	}
	return matchingKeys, nil
}

// 过滤出最后一条数据
func lastClientData(clientMetric *pbclient.Client, clientMap map[string]*pbclient.Client) map[string]*pbclient.Client {
	if clientMetric == nil {
		return nil
	}
	key := fmt.Sprintf("%d-%d-%s", clientMetric.Attachment.BizId,
		clientMetric.Attachment.AppId, clientMetric.Attachment.Uid)
	if p, ok := clientMap[key]; ok {
		if p.Spec.LastHeartbeatTime.AsTime().After(clientMetric.Spec.LastHeartbeatTime.AsTime()) {
			clientMap[key] = p
		}
	} else {
		clientMap[key] = clientMetric
	}
	return clientMap
}

func lastClientEventData(clientEventMetric *pbce.ClientEvent,
	clientEventMap map[string]*pbce.ClientEvent) map[string]*pbce.ClientEvent {
	if clientEventMetric == nil {
		return nil
	}
	key := fmt.Sprintf("%d-%d-%s", clientEventMetric.Attachment.BizId,
		clientEventMetric.Attachment.AppId, clientEventMetric.Attachment.CursorId)
	if p, ok := clientEventMap[key]; ok {
		if p.HeartbeatTime.AsTime().After(clientEventMetric.HeartbeatTime.AsTime()) {
			clientEventMap[key] = p
		}
	} else {
		clientEventMap[key] = clientEventMetric
	}
	return clientEventMap
}
