/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package publish

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-connserver/modules/metrics"
	"bk-bscp/cmd/bscp-connserver/modules/session"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/internal/strategy"
	"bk-bscp/internal/structs"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/natsmq"
)

// Manager handles all publish events.
type Manager struct {
	// config viper as context here.
	viper *viper.Viper

	// subscriber of publish message queue with target topic.
	subscriber *mq.Subscriber

	// session manager, used for push notification to sidecar.
	sessionMgr *session.Manager

	// strategy handler, check strategies when publish event coming.
	strategyHandler *strategy.Handler

	// prometheus metrics collector.
	collector *metrics.Collector
}

// NewManager creates new Manager.
func NewManager(viper *viper.Viper, subscriber *mq.Subscriber, sessionMgr *session.Manager,
	strategyHandler *strategy.Handler, collector *metrics.Collector) *Manager {
	return &Manager{
		viper:           viper,
		subscriber:      subscriber,
		sessionMgr:      sessionMgr,
		strategyHandler: strategyHandler,
		collector:       collector,
	}
}

// Init starts and keeps subscribing publish notification message here, and processed by the callback func.
func (mgr *Manager) Init() error {
	return mgr.subscriber.Subscribe(mgr.viper.GetString("natsmqCluster.publishtopic"), mgr.cb)
}

func (mgr *Manager) cb(bytes []byte) {
	msg := structs.Signalling{}
	if err := json.Unmarshal(bytes, &msg); err != nil {
		logger.Error("process publish signalling message, %+v", err)
		return
	}
	logger.Info("process publish signalling message, %+v", msg)

	go mgr.process(&msg)
}

// process is a callback func used for processing publish events.
func (mgr *Manager) process(msg *structs.Signalling) {
	mgr.collector.StatPublishingTask(true)

	switch msg.Type {
	case structs.SignallingTypePublish:
		if err := mgr.processPublishing(msg, NewSimpleRateController()); err != nil {
			logger.Error("process release publishing, %+v", err)
		}
	case structs.SignallingTypeRollback:
		if err := mgr.processRollbackPublishing(msg, NewSimpleRateController()); err != nil {
			logger.Error("process release rollback publishing, %+v", err)
		}
	default:
		logger.Error("process publish message, unknow signalling type[%+v]", msg.Type)
	}

	mgr.collector.StatPublishingTask(false)
}

// getSessions returns app instances that available and matched the strategies.
func (mgr *Manager) getSessions(msg *structs.Signalling) ([]*session.Session, error) {
	sessions, err := mgr.sessionMgr.GetSessions(msg.Publishing.Appid)
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, fmt.Errorf("appid[%s] empty sessions", msg.Publishing.Appid)
	}
	logger.Info("process notification message, appid[%s] available sessions count[%d]", msg.Publishing.Appid, len(sessions))

	// unmarshal strategies.
	strategies := strategy.Strategy{}
	if msg.Publishing.Strategies != strategy.EmptyStrategy {
		if err := json.Unmarshal([]byte(msg.Publishing.Strategies), &strategies); err != nil {
			return nil, err
		}
	}

	// range session list, and check publish strategies to get final targets.
	targets := []*session.Session{}

	for _, session := range sessions {
		if msg.Publishing.Strategies == strategy.EmptyStrategy {
			// empty strategy, all sidecars would be accepted.
			targets = append(targets, session)
		} else {
			ins := &pbcommon.AppInstance{
				Appid:     session.Sidecar.Appid,
				Clusterid: session.Sidecar.Clusterid,
				Zoneid:    session.Sidecar.Zoneid,
				Dc:        session.Sidecar.Dc,
				IP:        session.Sidecar.IP,
				Labels:    session.Sidecar.Labels,
			}
			matcher := mgr.strategyHandler.Matcher()
			if matcher(&strategies, ins) {
				targets = append(targets, session)
			}
		}
	}

	return targets, nil
}

// processPublishing processes publishing event message.
func (mgr *Manager) processPublishing(msg *structs.Signalling, rateController RateController) error {
	targets, err := mgr.getSessions(msg)
	if err != nil {
		return err
	}
	logger.Info("process publish notification message, final sidecar targets count[%d], %+v", len(targets), msg.Publishing)

	// step-publishing.
	rateController.Arrange(targets)
	for {
		targets := rateController.Next()
		if targets == nil {
			logger.V(3).Infof("step publishing done, %+v", msg.Publishing)
			break
		}

		logger.V(3).Infof("step publishing, count[%d]: %+v", len(targets), msg.Publishing)
		for _, target := range targets {
			mgr.pushNotification(target, msg)
		}
	}
	return nil
}

// processRollbackPublishing processes rollback publishing event message.
func (mgr *Manager) processRollbackPublishing(msg *structs.Signalling, rateController RateController) error {
	targets, err := mgr.getSessions(msg)
	if err != nil {
		return err
	}
	logger.Info("process rollback publish notification message, final sidecar targets count[%d], %+v", len(targets), msg.Publishing)

	// step-publishing.
	rateController.Arrange(targets)
	for {
		targets := rateController.Next()
		if targets == nil {
			logger.V(3).Infof("step rollback publishing done, %+v", msg.Publishing)
			break
		}

		logger.V(3).Infof("step rollback publishing, count[%d]: %+v", len(targets), msg.Publishing)
		for _, target := range targets {
			mgr.pushNotification(target, msg)
		}
	}
	return nil
}

// pushNotification pushs publishing notification to target sidecar base on session information.
func (mgr *Manager) pushNotification(target *session.Session, msg *structs.Signalling) {

	// TODO

	defer func() {
		if err := recover(); err != nil {
			logger.Warn("send publish notification to channel, channel is closed and recover success, %+v", err)
		}
	}()

	var notification interface{}

	switch msg.Type {
	case structs.SignallingTypePublish:
		notification = &pb.SCCMDPushNotification{
			Bid:         msg.Publishing.Bid,
			Appid:       msg.Publishing.Appid,
			Cfgsetid:    msg.Publishing.Cfgsetid,
			CfgsetName:  msg.Publishing.CfgsetName,
			CfgsetFpath: msg.Publishing.CfgsetFpath,
			Serialno:    msg.Publishing.Serialno,
			Releaseid:   msg.Publishing.Releaseid,
		}

	case structs.SignallingTypeRollback:
		notification = &pb.SCCMDPushRollbackNotification{
			Bid:         msg.Publishing.Bid,
			Appid:       msg.Publishing.Appid,
			Cfgsetid:    msg.Publishing.Cfgsetid,
			CfgsetName:  msg.Publishing.CfgsetName,
			CfgsetFpath: msg.Publishing.CfgsetFpath,
			Serialno:    msg.Publishing.Serialno,
			Releaseid:   msg.Publishing.Releaseid,
		}

	default:
		logger.Error("process notification message, unknow signalling type[%+v]", msg.Type)
		return
	}

	// send to channel in safe mode.
	select {
	case target.PubCh <- notification:
	case <-time.After(mgr.viper.GetDuration("server.pubChanTimeout")):
		logger.Warn("send notification to channel timeout, target[%+v], %+v", target, msg.Publishing)
	}
}
