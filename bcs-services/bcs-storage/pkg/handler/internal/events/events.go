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

package events

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/events"
)

const (
	KindTag = "kind"
)

var conditionTagList = []string{constants.IdTag, constants.EnvTag, KindTag, constants.LevelTag,
	constants.ComponentTag, constants.TypeTag, constants.ClusterIDTag, "extraInfo.name", "extraInfo.namespace",
	"extraInfo.kind"}

type general struct {
	ctx   context.Context
	event *storage.BcsStorageEvent
	//
	field        []string
	extra        operator.M
	extraContain operator.M
	timeBegin    int64
	timeEnd      int64
	offset       int64
	limit        int64
}

func (g *general) creatEvent() error {
	data := util.StructToMap(g.event)
	eventTime, _ := strconv.ParseInt(g.event.EventTime, 10, 64)
	data[constants.EventTimeTag] = time.Unix(eventTime, 0)

	resourceType := events.TablePrefix + data[constants.ClusterIDTag].(string)

	opt := &lib.StorePutOption{
		UniqueKey: events.EventIndexKeys,
	}

	return events.AddEvent(g.ctx, resourceType, data, opt)
}

func (g *general) getCondition() *operator.Condition {
	var condition *operator.Condition
	timeCondition := g.getTimeCondition()
	commonCondition := append(g.getCommonCondition(), timeCondition...)

	if len(commonCondition) != 0 {
		condition = operator.NewBranchCondition(operator.And, commonCondition...)
	} else {
		condition = operator.EmptyCondition
	}

	// handle the extra field
	var extraCondition []*operator.Condition
	extra := g.getExtra()
	features := make(operator.M)
	for k, v := range extra {
		if _, ok := v.([]interface{}); !ok {
			features[k] = []interface{}{v}
			continue
		}
		features[k] = v
	}

	if len(features) > 0 {
		extraCondition = append(extraCondition, operator.NewLeafCondition(operator.In, features))
	}

	// handle the extra contain field
	extraCon := g.getExtraContain()
	featuresCon := make(operator.M)
	for k, v := range extraCon {
		if _, ok := v.(string); !ok {
			continue
		}
		featuresCon[k] = v.(string)
	}

	if len(featuresCon) > 0 {
		extraCondition = append(extraCondition, operator.NewLeafCondition(operator.Con, featuresCon))
	}
	if len(extraCondition) != 0 {
		condition = operator.NewBranchCondition(operator.And, extraCondition...)
	}
	return condition
}

func (g *general) getTimeCondition() []*operator.Condition {
	var condList []*operator.Condition
	if g.timeBegin > 0 {
		condList = append(
			condList,
			operator.NewLeafCondition(
				operator.Gt,
				operator.M{
					constants.EventTimeTag: time.Unix(g.timeBegin, 0),
				},
			),
		)
	}

	if g.timeEnd > 0 {
		condList = append(
			condList,
			operator.NewLeafCondition(
				operator.Lt,
				operator.M{
					constants.EventTimeTag: time.Unix(g.timeEnd, 0),
				},
			),
		)
	}

	return condList
}

func (g *general) getExtra() operator.M {
	return g.extra
}

func (g *general) getCommonCondition() []*operator.Condition {
	var condList []*operator.Condition
	var temp map[string]interface{}
	data := util.StructToMap(g.event)

	for _, k := range conditionTagList {
		temp = data
		keys := []string{k}
		var result interface{}

		if strings.Contains(k, ".") {
			keys = strings.Split(k, ".")
		}

		for _, key := range keys {
			switch temp[key].(type) {
			case map[string]interface{}:
				temp = temp[key].(map[string]interface{})
			default:
				result = temp[key]
			}
		}

		if v, ok := result.(string); ok && v != "" {
			condList = append(
				condList, operator.NewLeafCondition(operator.In,
					operator.M{
						k: strings.Split(v, ","),
					},
				),
			)
		}
	}
	return condList
}

func (g *general) getExtraContain() operator.M {
	return g.extraContain
}

func (g *general) getEvents() ([]operator.M, int64, error) {
	condition := g.getCondition()

	resourceType := strings.Split(g.event.ClusterId, ",")

	opt := &lib.StoreGetOption{
		Fields: g.field,
		Sort: map[string]int{
			constants.EventTimeTag: -1,
		},
		Cond:   condition,
		Offset: g.offset,
		Limit:  g.limit,
	}

	return events.GetEventList(g.ctx, resourceType, opt)
}

// HandlerPutEvent PutEvent 业务方法
func HandlerPutEvent(ctx context.Context, req *storage.PutEventRequest) error {
	event := &storage.BcsStorageEvent{
		ClusterId: req.ClusterId,
		Kind:      req.Kind,
		ExtraInfo: req.ExtraInfo,
		Env:       req.Env,
		Component: req.Component,
		Type:      req.Type,
		Data:      req.Data,
		Level:     req.Level,
		Describe:  req.Describe,
		EventTime: strconv.FormatInt(req.EventTime, 10),
	}
	g := &general{
		ctx:   ctx,
		event: event,
	}

	return g.creatEvent()
}

// HandlerListEvent ListEvent 业务方法
func HandlerListEvent(ctx context.Context, req *storage.ListEventRequest) ([]operator.M, int64, error) {
	event := &storage.BcsStorageEvent{
		XId:       req.Id,
		Env:       req.Env,
		Kind:      req.Kind,
		Level:     req.Level,
		Component: req.Component,
		Type:      req.Type,
		ClusterId: req.ClusterId,
		ExtraInfo: req.ExtraInfo,
	}
	g := &general{
		ctx:   ctx,
		event: event,
		//
		field:        req.Field,
		extra:        req.Extra.AsMap(),
		timeBegin:    req.TimeBegin,
		timeEnd:      req.TimeEnd,
		offset:       int64(req.Offset),
		limit:        int64(req.Length),
		extraContain: req.ExtraContain.AsMap(),
	}

	return g.getEvents()
}

// HandlerWatch Watch业务方法
func HandlerWatch(ctx context.Context, req *storage.WatchEventRequest) (chan *lib.Event, error) {
	store := events.GetStore()
	watchOption := &lib.StoreWatchOption{
		Cond:      req.Option.Cond.AsMap(),
		SelfOnly:  req.Option.SelfOnly,
		MaxEvents: uint(req.Option.MaxEvents),
		Timeout:   req.Option.Timeout.AsDuration(),
		MustDiff:  req.Option.MustDiff,
	}
	return store.Watch(ctx, events.TablePrefix+req.ClusterId, watchOption)
}
