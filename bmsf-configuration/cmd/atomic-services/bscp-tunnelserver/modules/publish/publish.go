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
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"

	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/metrics"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/session"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// Manager handles all publish events.
type Manager struct {
	// config viper as context here.
	viper *viper.Viper

	// session manager, used for push notification to app instance.
	sessionMgr *session.Manager

	// strategy handler, check strategies when publish event coming.
	strategyHandler *strategy.Handler

	// prometheus metrics collector.
	collector *metrics.Collector
}

// NewManager creates new Manager.
func NewManager(viper *viper.Viper, sessionMgr *session.Manager,
	strategyHandler *strategy.Handler, collector *metrics.Collector) *Manager {
	return &Manager{
		viper:           viper,
		sessionMgr:      sessionMgr,
		strategyHandler: strategyHandler,
		collector:       collector,
	}
}

// getSessions returns app instances that available and matched the strategies.
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

	return sessions, nil
}

func (mgr *Manager) agentKey(pluginID string, cloudID int32) string {
	return fmt.Sprintf("%s-%d", pluginID, cloudID)
}

func (mgr *Manager) getAgents(sessions []*session.Session) ([]*modules.AgentInformation, map[string]int) {
	// agent -> sessions.
	sessionMapping := make(map[string][]*session.Session)

	for _, session := range sessions {
		agentKey := mgr.agentKey(session.PluginID, session.CloudID)
		sessionMapping[agentKey] = append(sessionMapping[agentKey], session)
	}

	// agent list.
	agents := []*modules.AgentInformation{}

	// agent -> session count.
	sessionCountMapping := make(map[string]int)

	for agentKey, sessions := range sessionMapping {
		agents = append(agents, &modules.AgentInformation{
			HostIP:  sessions[0].PluginID,
			CloudID: sessions[0].CloudID,
		})
		sessionCountMapping[agentKey] = len(sessions)
	}

	return agents, sessionCountMapping
}

// Publish pubs signalling to sidecars.
func (mgr *Manager) Publish(msg *pbcommon.Signalling) error {
	sessions, err := mgr.getSessions(msg)
	if err != nil {
		return err
	}
	agents, _ := mgr.getAgents(sessions)

	mgr.pushNotification(agents, len(sessions), sessions[0].PubFunc, msg)

	return nil
}

func (mgr *Manager) pushNotification(agents []*modules.AgentInformation, sessionCount int,
	pubFunc func(sendProcesserMessage *modules.GSESendProcesserMessage) error, msg *pbcommon.Signalling) {

	startTime := time.Now()

	defer func() {
		logger.V(2).Infof("process publishing notification done, appid[%s] agent count[%d], session count[%d], cost: %+v",
			msg.Publishing.AppId, len(agents), sessionCount, time.Since(startTime))

		if err := recover(); err != nil {
			logger.Warn("send publishing notification to channel, channel is closed and recover success, %+v", err)
		}
	}()

	upStream := &pb.GeneralTunnelUpStream{Seq: common.Sequence()}

	switch msg.Type {
	case pbcommon.SignallingType_ST_SignallingTypePublish:
		notification := &pb.GTCMDPublish{
			BizId:      msg.Publishing.BizId,
			AppId:      msg.Publishing.AppId,
			CfgId:      msg.Publishing.CfgId,
			CfgName:    msg.Publishing.CfgName,
			CfgFpath:   msg.Publishing.CfgFpath,
			Serialno:   msg.Publishing.Serialno,
			ReleaseId:  msg.Publishing.ReleaseId,
			Strategies: msg.Publishing.Strategies,
			Nice:       msg.Publishing.Nice / float64(sessionCount),
		}

		logger.V(2).Infof("process publish notification, appid[%s] nice[%+v] totalSession[%d]",
			msg.Publishing.AppId, notification.Nice, sessionCount)

		data, err := proto.Marshal(notification)
		if err != nil {
			logger.Error("marshal notification to upstream, %+v", err)
			return
		}
		upStream.Data = data
		upStream.Cmd = pb.GeneralTunnelCmd_GTCMD_S2C_PUBLISH

	case pbcommon.SignallingType_ST_SignallingTypeRollback:
		notification := &pb.GTCMDRollback{
			BizId:      msg.Publishing.BizId,
			AppId:      msg.Publishing.AppId,
			CfgId:      msg.Publishing.CfgId,
			CfgName:    msg.Publishing.CfgName,
			CfgFpath:   msg.Publishing.CfgFpath,
			Serialno:   msg.Publishing.Serialno,
			ReleaseId:  msg.Publishing.ReleaseId,
			Strategies: msg.Publishing.Strategies,
			Nice:       msg.Publishing.Nice / float64(sessionCount),
		}

		logger.V(2).Infof("process rollback notification, appid[%s] nice[%+v] totalSession[%d]",
			msg.Publishing.AppId, notification.Nice, sessionCount)

		data, err := proto.Marshal(notification)
		if err != nil {
			logger.Error("marshal notification to upstream, %+v", err)
			return
		}
		upStream.Data = data
		upStream.Cmd = pb.GeneralTunnelCmd_GTCMD_S2C_ROLLBACK

	case pbcommon.SignallingType_ST_SignallingTypeReload:
		notification := &pb.GTCMDReload{
			BizId:      msg.Publishing.BizId,
			AppId:      msg.Publishing.AppId,
			Strategies: msg.Publishing.Strategies,
			ReloadSpec: msg.Publishing.ReloadSpec,
			Nice:       msg.Publishing.Nice / float64(sessionCount),
		}

		logger.V(2).Infof("process reload notification, appid[%s] nice[%+v] totalSession[%d]",
			msg.Publishing.AppId, notification.Nice, sessionCount)

		data, err := proto.Marshal(notification)
		if err != nil {
			logger.Error("marshal notification to upstream, %+v", err)
			return
		}
		upStream.Data = data
		upStream.Cmd = pb.GeneralTunnelCmd_GTCMD_S2C_RELOAD

	default:
		logger.Error("process notification message, unknow signalling type[%+v]", msg.Type)
		return
	}

	// publish to plugin.
	pubFunc(&modules.GSESendProcesserMessage{Agents: agents, UpStream: upStream})
}
