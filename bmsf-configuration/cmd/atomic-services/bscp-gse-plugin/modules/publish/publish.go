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

	"bk-bscp/cmd/atomic-services/bscp-gse-plugin/modules/session"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/internal/safeviper"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// Manager handles all publish events.
type Manager struct {
	// config viper as context here.
	viper *safeviper.SafeViper

	// session manager, used for push notification to sidecar.
	sessionMgr *session.Manager

	// strategy handler, check strategies when publish event coming.
	strategyHandler *strategy.Handler
}

// NewManager creates new Manager.
func NewManager(viper *safeviper.SafeViper, sessionMgr *session.Manager, strategyHandler *strategy.Handler) *Manager {
	return &Manager{
		viper:           viper,
		sessionMgr:      sessionMgr,
		strategyHandler: strategyHandler,
	}
}

// Process is a callback func used for processing publish events.
func (mgr *Manager) Process(msg *pbcommon.Signalling) {
	logger.V(2).Infof("process new notification, %+v", msg.Publishing)

	switch msg.Type {
	case pbcommon.SignallingType_ST_SignallingTypePublish:
		if err := mgr.processPublishing(msg); err != nil {
			logger.Error("process release publishing, %+v", err)
		}

	case pbcommon.SignallingType_ST_SignallingTypeRollback:
		if err := mgr.processRollbackPublishing(msg); err != nil {
			logger.Error("process release rollback publishing, %+v", err)
		}

	case pbcommon.SignallingType_ST_SignallingTypeReload:
		if err := mgr.processReloadPublishing(msg); err != nil {
			logger.Error("process release reload publishing, %+v", err)
		}

	default:
		logger.Error("process publish message, unknow signalling type[%+v]", msg.Type)
	}
}

func (mgr *Manager) getSessions(msg *pbcommon.Signalling) ([]*session.Session, error) {
	sessions, err := mgr.sessionMgr.GetSessions(msg.Publishing.AppId)
	if err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("appid[%s] have no available instance session", msg.Publishing.AppId)
	}

	logger.V(2).Infof("process publishing notification, appid[%s] available sessions count[%d]",
		msg.Publishing.AppId, len(sessions))

	// matched strategies.
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
				AppId:   session.Sidecar.AppID,
				CloudId: session.Sidecar.CloudID,
				Ip:      session.Sidecar.IP,
				Path:    session.Sidecar.Path,
				Labels:  session.Sidecar.Labels,
			}
			matcher := mgr.strategyHandler.Matcher()
			if matcher(&strategies, ins) {
				targets = append(targets, session)
			}
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("appid[%s] have no strategy matched instance session", msg.Publishing.AppId)
	}

	return targets, nil
}

func (mgr *Manager) getPublishingRate(nice float64, sessionCount int) time.Duration {
	limit := nice * float64(sessionCount)
	rate := float64(time.Second) / limit
	return time.Duration(rate)
}

// processPublishing processes publishing event message.
func (mgr *Manager) processPublishing(msg *pbcommon.Signalling) error {
	sessions, err := mgr.getSessions(msg)
	if err != nil {
		return err
	}
	rate := mgr.getPublishingRate(msg.Publishing.Nice, len(sessions))

	logger.V(2).Infof("process publish notification, appid[%s] final instance count[%d], nice[%+v] rate[%+v]",
		msg.Publishing.AppId, len(sessions), msg.Publishing.Nice, rate)

	for idx, session := range sessions {
		if idx == 0 {
			common.DelayRandomMS(1000)
		} else {
			time.Sleep(rate)
		}
		mgr.pushNotification(session, msg)
	}

	return nil
}

// processRollbackPublishing processes rollback publishing event message.
func (mgr *Manager) processRollbackPublishing(msg *pbcommon.Signalling) error {
	sessions, err := mgr.getSessions(msg)
	if err != nil {
		return err
	}
	rate := mgr.getPublishingRate(msg.Publishing.Nice, len(sessions))

	logger.V(2).Infof("process rollback notification, appid[%s] final instance count[%d], nice[%+v] rate[%+v]",
		msg.Publishing.AppId, len(sessions), msg.Publishing.Nice, rate)

	for idx, session := range sessions {
		if idx == 0 {
			common.DelayRandomMS(1000)
		} else {
			time.Sleep(rate)
		}
		mgr.pushNotification(session, msg)
	}

	return nil
}

// processReloadPublishing processes reload publishing event message.
func (mgr *Manager) processReloadPublishing(msg *pbcommon.Signalling) error {
	sessions, err := mgr.getSessions(msg)
	if err != nil {
		return err
	}
	rate := mgr.getPublishingRate(msg.Publishing.Nice, len(sessions))

	logger.V(2).Infof("process reload notification, appid[%s] final instance count[%d], nice[%+v] rate[%+v]",
		msg.Publishing.AppId, len(sessions), msg.Publishing.Nice, rate)

	for _, session := range sessions {
		// NOTE: reload notification, not step publishing.
		mgr.pushNotification(session, msg)
	}

	return nil
}

// pushNotification pushs publishing notification to target sidecar base on session information.
func (mgr *Manager) pushNotification(target *session.Session, msg *pbcommon.Signalling) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn("send publish notification to channel, channel is closed and recover success, %+v", err)
		}
	}()

	var notification interface{}

	switch msg.Type {
	case pbcommon.SignallingType_ST_SignallingTypePublish:
		notification = &pb.SCCMDPushNotification{
			BizId:     msg.Publishing.BizId,
			AppId:     msg.Publishing.AppId,
			CfgId:     msg.Publishing.CfgId,
			CfgName:   msg.Publishing.CfgName,
			CfgFpath:  msg.Publishing.CfgFpath,
			Serialno:  msg.Publishing.Serialno,
			ReleaseId: msg.Publishing.ReleaseId,
		}

	case pbcommon.SignallingType_ST_SignallingTypeRollback:
		notification = &pb.SCCMDPushRollbackNotification{
			BizId:     msg.Publishing.BizId,
			AppId:     msg.Publishing.AppId,
			CfgId:     msg.Publishing.CfgId,
			CfgName:   msg.Publishing.CfgName,
			CfgFpath:  msg.Publishing.CfgFpath,
			Serialno:  msg.Publishing.Serialno,
			ReleaseId: msg.Publishing.ReleaseId,
		}

	case pbcommon.SignallingType_ST_SignallingTypeReload:
		notification = &pb.SCCMDPushReloadNotification{
			BizId:      msg.Publishing.BizId,
			AppId:      msg.Publishing.AppId,
			ReloadSpec: msg.Publishing.ReloadSpec,
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
