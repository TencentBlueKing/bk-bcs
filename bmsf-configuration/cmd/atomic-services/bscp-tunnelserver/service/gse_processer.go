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

package service

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"

	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/gse"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/metrics"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/session"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/pkg/logger"
)

// GSEProcesser is gse message processer.
type GSEProcesser struct {
	// config viper as context here.
	viper *viper.Viper

	// plat service protocol factory.
	platService *gse.PlatService

	// session manager, handles gse plugin sessions.
	sessionMgr *session.Manager

	// pushToPlugin is push to plugin callback.
	pushToPlugin func(sendProcesserMessage *modules.GSESendProcesserMessage) error

	// tunnelserver handles normal request from plugin.
	tunnelServer *TunnelServer

	// prometheus metrics collector.
	collector *metrics.Collector

	// GTCMD_C2S_QUERY_APP_METADATA message channel.
	queryAppMetadataMessageChan chan *modules.GSERecvProcesserMessage

	// GTCMD_C2S_QUERY_HOST_PROCATTR_LIST message channel.
	queryProcAttrListMessageChan chan *modules.GSERecvProcesserMessage

	// GTCMD_C2S_PLUGIN_INFO message channel.
	pluginInfoMessageChan chan *modules.GSERecvProcesserMessage

	// GTCMD_C2S_EFFECT_REPORT message channel.
	effectReportMessageChan chan *modules.GSERecvProcesserMessage

	// GTCMD_C2S_PULL_RELEASE message channel.
	pullReleaseMessageChan chan *modules.GSERecvProcesserMessage

	// GTCMD_C2S_PULL_CONFIGLIST message channel.
	pullConfigListMessageChan chan *modules.GSERecvProcesserMessage
}

// NewGSEProcesser creates a new GSEProcesser.
func NewGSEProcesser(viper *viper.Viper, platService *gse.PlatService, tunnelServer *TunnelServer,
	pushToPlugin func(sendProcesserMessage *modules.GSESendProcesserMessage) error) *GSEProcesser {

	return &GSEProcesser{
		viper:        viper,
		platService:  platService,
		pushToPlugin: pushToPlugin,
		sessionMgr:   tunnelServer.sessionMgr,
		tunnelServer: tunnelServer,
		collector:    tunnelServer.collector,

		queryAppMetadataMessageChan: make(chan *modules.GSERecvProcesserMessage,
			viper.GetInt("gseTaskServer.processerMessageChanSize")),

		queryProcAttrListMessageChan: make(chan *modules.GSERecvProcesserMessage,
			viper.GetInt("gseTaskServer.processerMessageChanSize")),

		pluginInfoMessageChan: make(chan *modules.GSERecvProcesserMessage,
			viper.GetInt("gseTaskServer.processerMessageChanSize")),

		effectReportMessageChan: make(chan *modules.GSERecvProcesserMessage,
			viper.GetInt("gseTaskServer.processerMessageChanSize")),

		pullReleaseMessageChan: make(chan *modules.GSERecvProcesserMessage,
			viper.GetInt("gseTaskServer.processerMessageChanSize")),

		pullConfigListMessageChan: make(chan *modules.GSERecvProcesserMessage,
			viper.GetInt("gseTaskServer.processerMessageChanSize")),
	}
}

// Run runs the gse recv processer.
func (p *GSEProcesser) Run() {
	// keep processing recved message.
	for i := 0; i < p.viper.GetInt("gseTaskServer.protocolProcesserNum"); i++ {
		// process query app metadata.
		go p.processQueryAppMetadata()

		// process query procattr list.
		go p.processQueryProcAttrList()

		// process plugin info.
		go p.processPluginInfo()

		// process effect report.
		go p.processEffectReport()

		// process pull release.
		go p.processPullRelease()

		// process pull config list.
		go p.processPullConfigList()
	}
}

// Handle handles gse recved message for target GTCMD type.
func (p *GSEProcesser) Handle(message *gse.Message) error {
	// process recved message.
	switch message.Header.MainHeader.MessageType {
	case gse.PlatServiceCMDSubscribeFromPluginReq:
		if err := p.handlePluginRequest(message); err != nil {
			logger.Warnf("GSEPROCESSER| handle plugin request failed, %+v", err)
		}

	case gse.PlatServiceCMDPushToPluginResp:
		if err := p.handlePushResponse(message); err != nil {
			logger.Warnf("GSEPROCESSER| handle push response failed, %+v", err)
		}

	default:
		logger.Warnf("GSEPROCESSER| recv not matched message type[%d], %+v, %+v",
			message.Header.MainHeader.MessageType, message.Header.MainHeader, message.Header.ExtraHeader)
	}
	return nil
}

func (p *GSEProcesser) handlePluginRequest(message *gse.Message) error {
	p.collector.StatGSEPlatMessageProcessedCount("PLUGIN_REQ")

	pluginMessage := &gse.PlatServiceSubscribeFromPluginBody{}

	// decode message.
	if err := p.platService.DecodeJSONBody(message.Body, pluginMessage); err != nil {
		return fmt.Errorf("decode plugin request message failed, %+v, %+v, %+v",
			message.Header.MainHeader, message.Header.ExtraHeader, err)
	}

	decodeMessage, err := base64.StdEncoding.DecodeString(pluginMessage.Message)
	if err != nil {
		return fmt.Errorf("base64 decode plugin request message failed, %+v, %+v, %+v",
			message.Header.MainHeader, message.Header.ExtraHeader, err)
	}
	pluginMessage.Message = string(decodeMessage)

	// general tunnel down stream request.
	downStream := &pb.GeneralTunnelDownStream{}

	// unmarshal protobuf.
	if err := proto.Unmarshal([]byte(pluginMessage.Message), downStream); err != nil {
		return fmt.Errorf("decode general tunnel down stream message failed, %+v, %+v, %+v",
			message.Header.MainHeader, message.Header.ExtraHeader, err)
	}

	// new recv processer message.
	recvProcesserMessage := &modules.GSERecvProcesserMessage{
		MsgSeqID:   message.Header.ExtraHeader.MessageSequenceID,
		Agent:      pluginMessage.Agent,
		DownStream: downStream,
	}

	// add down stream message to processer channel.
	if err := p.addMessageToProcesserChan(recvProcesserMessage); err != nil {
		return fmt.Errorf("add to gse recv processer channel failed, %+v, %+v, %+v",
			message.Header.MainHeader, message.Header.ExtraHeader, err)
	}

	return nil
}

func (p *GSEProcesser) addMessageToProcesserChan(recvProcesserMessage *modules.GSERecvProcesserMessage) error {
	switch recvProcesserMessage.DownStream.Cmd {

	// query app metadata request.
	case pb.GeneralTunnelCmd_GTCMD_C2S_QUERY_APP_METADATA:
		// stat message channel runtime size.
		p.collector.StatGSEPlatMessageChanRuntime("QUERY_APP_METADATA", int64(len(p.queryAppMetadataMessageChan)))

		select {
		case p.queryAppMetadataMessageChan <- recvProcesserMessage:
			p.collector.StatGSEPlatMessageCount("QUERY_APP_METADATA")

		case <-time.After(p.viper.GetDuration("gseTaskServer.processerMessageChanTimeout")):
			p.collector.StatGSEPlatMessageFuseCount("QUERY_APP_METADATA")
			return fmt.Errorf("query app metadata channel timeout, current size[%d]", len(p.queryAppMetadataMessageChan))
		}

	// query host procattrs request.
	case pb.GeneralTunnelCmd_GTCMD_C2S_QUERY_HOST_PROCATTR_LIST:
		// stat message channel runtime size.
		p.collector.StatGSEPlatMessageChanRuntime("QUERY_HOST_PROCATTR_LIST", int64(len(p.queryProcAttrListMessageChan)))

		select {
		case p.queryProcAttrListMessageChan <- recvProcesserMessage:
			p.collector.StatGSEPlatMessageCount("QUERY_HOST_PROCATTR_LIST")

		case <-time.After(p.viper.GetDuration("gseTaskServer.processerMessageChanTimeout")):
			p.collector.StatGSEPlatMessageFuseCount("QUERY_HOST_PROCATTR_LIST")
			return fmt.Errorf("query procattr list channel timeout, current size[%d]", len(p.queryProcAttrListMessageChan))
		}

	// handle plugin info(flush sidecar instance session).
	case pb.GeneralTunnelCmd_GTCMD_C2S_PLUGIN_INFO:
		// stat message channel runtime size.
		p.collector.StatGSEPlatMessageChanRuntime("PLUGIN_INFO", int64(len(p.pluginInfoMessageChan)))

		select {
		case p.pluginInfoMessageChan <- recvProcesserMessage:
			p.collector.StatGSEPlatMessageCount("PLUGIN_INFO")

		case <-time.After(p.viper.GetDuration("gseTaskServer.processerMessageChanTimeout")):
			p.collector.StatGSEPlatMessageFuseCount("PLUGIN_INFO")
			return fmt.Errorf("plugin info channel timeout, current size[%d]", len(p.pluginInfoMessageChan))
		}

	// handle sidecar effect reports.
	case pb.GeneralTunnelCmd_GTCMD_C2S_EFFECT_REPORT:
		// stat message channel runtime size.
		p.collector.StatGSEPlatMessageChanRuntime("EFFECT_REPORT", int64(len(p.effectReportMessageChan)))

		select {
		case p.effectReportMessageChan <- recvProcesserMessage:
			p.collector.StatGSEPlatMessageCount("EFFECT_REPORT")

		case <-time.After(p.viper.GetDuration("gseTaskServer.processerMessageChanTimeout")):
			p.collector.StatGSEPlatMessageFuseCount("EFFECT_REPORT")
			return fmt.Errorf("effect report channel timeout, current size[%d]", len(p.effectReportMessageChan))
		}

	// pull release information.
	case pb.GeneralTunnelCmd_GTCMD_C2S_PULL_RELEASE:
		// stat message channel runtime size.
		p.collector.StatGSEPlatMessageChanRuntime("PULL_RELEASE", int64(len(p.pullReleaseMessageChan)))

		select {
		case p.pullReleaseMessageChan <- recvProcesserMessage:
			p.collector.StatGSEPlatMessageCount("PULL_RELEASE")

		case <-time.After(p.viper.GetDuration("gseTaskServer.processerMessageChanTimeout")):
			p.collector.StatGSEPlatMessageFuseCount("PULL_RELEASE")
			return fmt.Errorf("pull release channel timeout, current size[%d]", len(p.pullReleaseMessageChan))
		}

	// pull config list.
	case pb.GeneralTunnelCmd_GTCMD_C2S_PULL_CONFIGLIST:
		// stat message channel runtime size.
		p.collector.StatGSEPlatMessageChanRuntime("PULL_CONFIGLIST", int64(len(p.pullConfigListMessageChan)))

		select {
		case p.pullConfigListMessageChan <- recvProcesserMessage:
			p.collector.StatGSEPlatMessageCount("PULL_CONFIGLIST")
		case <-time.After(p.viper.GetDuration("gseTaskServer.processerMessageChanTimeout")):
			p.collector.StatGSEPlatMessageFuseCount("PULL_CONFIGLIST")
			return fmt.Errorf("pull config list channel timeout, current size[%d]", len(p.pullConfigListMessageChan))
		}

	default:
		return fmt.Errorf("unknow general tunnel cmd, %+v", recvProcesserMessage.DownStream.Cmd)
	}

	return nil
}

func (p *GSEProcesser) processQueryAppMetadata() {
	startTime := time.Now()

	for {
		p.collector.StatProcess("QUERY_APP_METADATA", startTime, time.Now())

		recvProcesserMessage := <-p.queryAppMetadataMessageChan

		startTime = time.Now()

		p.collector.StatGSEPlatMessageProcessedCount("QUERY_APP_METADATA")

		req := &pb.GTCMDQueryAppMetadataReq{}
		if err := proto.Unmarshal([]byte(recvProcesserMessage.DownStream.Data), req); err != nil {
			logger.Errorf("GSEPROCESSER| decode QueryAppMetadata failed, %+v", err)
			continue
		}

		resp, err := p.tunnelServer.QueryAppMetadata(recvProcesserMessage.MsgSeqID, recvProcesserMessage.Agent, req)
		if err != nil {
			logger.Errorf("GSEPROCESSER| tunnelserver process QueryAppMetadata failed, %+v", err)
			continue
		}

		respData, err := proto.Marshal(resp)
		if err != nil {
			logger.Errorf("GSEPROCESSER| tunnelserver process QueryAppMetadata response failed, %+v", err)
			continue
		}

		// response(if need) with same seq/cmd.
		upStream := &pb.GeneralTunnelUpStream{
			Seq:  recvProcesserMessage.DownStream.Seq,
			Cmd:  recvProcesserMessage.DownStream.Cmd,
			Data: respData,
		}

		sendProcesserMessage := &modules.GSESendProcesserMessage{
			MsgSeqID: recvProcesserMessage.MsgSeqID,
			Agents:   []*modules.AgentInformation{recvProcesserMessage.Agent},
			UpStream: upStream,
		}
		p.pushToPlugin(sendProcesserMessage)
	}
}

func (p *GSEProcesser) processQueryProcAttrList() {
	startTime := time.Now()

	for {
		p.collector.StatProcess("QUERY_HOST_PROCATTR_LIST", startTime, time.Now())

		recvProcesserMessage := <-p.queryProcAttrListMessageChan

		startTime = time.Now()

		p.collector.StatGSEPlatMessageProcessedCount("QUERY_HOST_PROCATTR_LIST")

		req := &pb.GTCMDQueryHostProcAttrListReq{}
		if err := proto.Unmarshal([]byte(recvProcesserMessage.DownStream.Data), req); err != nil {
			logger.Errorf("GSEPROCESSER| decode QueryHostProcAttrList failed, %+v", err)
			continue
		}
		req.CloudId = fmt.Sprintf("%d", recvProcesserMessage.Agent.CloudID)
		req.Ip = recvProcesserMessage.Agent.HostIP

		resp, err := p.tunnelServer.QueryHostProcAttrList(recvProcesserMessage.MsgSeqID, recvProcesserMessage.Agent, req)
		if err != nil {
			logger.Errorf("GSEPROCESSER| tunnelserver process QueryHostProcAttrList failed, %+v", err)
			continue
		}

		respData, err := proto.Marshal(resp)
		if err != nil {
			logger.Errorf("GSEPROCESSER| tunnelserver process QueryHostProcAttrList response failed, %+v", err)
			continue
		}

		// response(if need) with same seq/cmd.
		upStream := &pb.GeneralTunnelUpStream{
			Seq:  recvProcesserMessage.DownStream.Seq,
			Cmd:  recvProcesserMessage.DownStream.Cmd,
			Data: respData,
		}

		sendProcesserMessage := &modules.GSESendProcesserMessage{
			MsgSeqID: recvProcesserMessage.MsgSeqID,
			Agents:   []*modules.AgentInformation{recvProcesserMessage.Agent},
			UpStream: upStream,
		}
		p.pushToPlugin(sendProcesserMessage)
	}
}

func (p *GSEProcesser) processPluginInfo() {
	startTime := time.Now()

	for {
		p.collector.StatProcess("PLUGIN_INFO", startTime, time.Now())

		recvProcesserMessage := <-p.pluginInfoMessageChan

		startTime = time.Now()

		p.collector.StatGSEPlatMessageProcessedCount("PLUGIN_INFO")

		info := &pb.GTCMDPluginInfo{}
		if err := proto.Unmarshal([]byte(recvProcesserMessage.DownStream.Data), info); err != nil {
			logger.Errorf("GSEPROCESSER| decode PluginInfo failed, %+v", err)
			continue
		}

		for _, instance := range info.Instances {
			if info.Timeout <= 0 {
				// delete session of offline sidecar.
				if err := p.sessionMgr.DeleteSession(instance); err != nil {
					logger.Warnf("GSEPROCESSER| delete session, pluginid[%s] cloudid[%d] %+v, %+v",
						recvProcesserMessage.Agent.HostIP, recvProcesserMessage.Agent.CloudID, instance, err)
				}
				continue
			}

			// flush session of online sidecar.
			if err := p.sessionMgr.FlushSession(instance,
				recvProcesserMessage.Agent.HostIP, recvProcesserMessage.Agent.CloudID,
				p.pushToPlugin,
				func(key, value interface{}) error {
					session, ok := value.(*session.Session)
					if ok {
						instance := &pbcommon.AppInstance{
							BizId:   session.Sidecar.BizID,
							AppId:   session.Sidecar.AppID,
							CloudId: session.Sidecar.CloudID,
							Ip:      session.Sidecar.IP,
							Path:    session.Sidecar.Path,
							Labels:  session.Sidecar.Labels,
						}
						return p.tunnelServer.CreateAppInstance(instance)
					}
					return nil
				},
				func(key, value interface{}) {
					session, ok := value.(*session.Session)
					if ok {
						instance := &pbcommon.AppInstance{
							BizId:   session.Sidecar.BizID,
							AppId:   session.Sidecar.AppID,
							CloudId: session.Sidecar.CloudID,
							Ip:      session.Sidecar.IP,
							Path:    session.Sidecar.Path,
							Labels:  session.Sidecar.Labels,
						}
						p.tunnelServer.UpdateAppInstance(instance)
					}
				}, time.Duration(info.Timeout)*time.Second); err != nil {
				logger.Warnf("GSEPROCESSER| flush session, pluginid[%s] cloudid[%d] %+v, %+v",
					recvProcesserMessage.Agent.HostIP, recvProcesserMessage.Agent.CloudID, instance, err)
			}
		}

		logger.V(2).Infof("GSEPROCESSER| flush session, pluginid[%s] cloudid[%d] count[%d] done, cost[%+v]",
			recvProcesserMessage.Agent.HostIP, recvProcesserMessage.Agent.CloudID, len(info.Instances), time.Since(startTime))
	}
}

func (p *GSEProcesser) processEffectReport() {
	startTime := time.Now()

	for {
		p.collector.StatProcess("EFFECT_REPORT", startTime, time.Now())

		recvProcesserMessage := <-p.effectReportMessageChan

		startTime = time.Now()

		p.collector.StatGSEPlatMessageProcessedCount("EFFECT_REPORT")

		report := &pb.GTCMDEffectReport{}
		if err := proto.Unmarshal([]byte(recvProcesserMessage.DownStream.Data), report); err != nil {
			logger.Errorf("GSEPROCESSER| decode EffectReport failed, %+v", err)
			continue
		}

		if err := p.tunnelServer.Report(recvProcesserMessage.MsgSeqID, recvProcesserMessage.Agent, report); err != nil {
			logger.Errorf("GSEPROCESSER| tunnelserver process EffectReport failed, %+v", err)
			continue
		}

		logger.V(2).Infof("GSEPROCESSER| handle effect reports, pluginid[%s] cloudid[%d]",
			recvProcesserMessage.Agent.HostIP, recvProcesserMessage.Agent.CloudID)
	}
}

func (p *GSEProcesser) processPullRelease() {
	startTime := time.Now()

	for {
		p.collector.StatProcess("PULL_RELEASE", startTime, time.Now())

		recvProcesserMessage := <-p.pullReleaseMessageChan

		startTime = time.Now()

		p.collector.StatGSEPlatMessageProcessedCount("PULL_RELEASE")

		req := &pb.GTCMDPullReleaseReq{}
		if err := proto.Unmarshal([]byte(recvProcesserMessage.DownStream.Data), req); err != nil {
			logger.Errorf("GSEPROCESSER| decode PullRelease failed, %+v", err)
			continue
		}

		resp, err := p.tunnelServer.PullRelease(recvProcesserMessage.MsgSeqID, recvProcesserMessage.Agent, req)
		if err != nil {
			logger.Errorf("GSEPROCESSER| tunnelserver process PullRelease failed, %+v", err)
			continue
		}

		respData, err := proto.Marshal(resp)
		if err != nil {
			logger.Errorf("GSEPROCESSER| tunnelserver process PullRelease response failed, %+v", err)
			continue
		}

		// response(if need) with same seq/cmd.
		upStream := &pb.GeneralTunnelUpStream{
			Seq:  recvProcesserMessage.DownStream.Seq,
			Cmd:  recvProcesserMessage.DownStream.Cmd,
			Data: respData,
		}

		sendProcesserMessage := &modules.GSESendProcesserMessage{
			MsgSeqID: recvProcesserMessage.MsgSeqID,
			Agents:   []*modules.AgentInformation{recvProcesserMessage.Agent},
			UpStream: upStream,
		}
		p.pushToPlugin(sendProcesserMessage)
	}
}

func (p *GSEProcesser) processPullConfigList() {
	startTime := time.Now()

	for {
		p.collector.StatProcess("PULL_CONFIGLIST", startTime, time.Now())

		recvProcesserMessage := <-p.pullConfigListMessageChan

		startTime = time.Now()

		p.collector.StatGSEPlatMessageProcessedCount("PULL_CONFIGLIST")

		req := &pb.GTCMDPullConfigListReq{}
		if err := proto.Unmarshal([]byte(recvProcesserMessage.DownStream.Data), req); err != nil {
			logger.Errorf("GSEPROCESSER| decode PullConfigList failed, %+v", err)
			continue
		}

		resp, err := p.tunnelServer.PullConfigList(recvProcesserMessage.MsgSeqID, recvProcesserMessage.Agent, req)
		if err != nil {
			logger.Errorf("GSEPROCESSER| tunnelserver process PullConfigList failed, %+v", err)
			continue
		}

		respData, err := proto.Marshal(resp)
		if err != nil {
			logger.Errorf("GSEPROCESSER| tunnelserver process PullConfigList response failed, %+v", err)
			continue
		}

		// response(if need) with same seq/cmd.
		upStream := &pb.GeneralTunnelUpStream{
			Seq:  recvProcesserMessage.DownStream.Seq,
			Cmd:  recvProcesserMessage.DownStream.Cmd,
			Data: respData,
		}

		sendProcesserMessage := &modules.GSESendProcesserMessage{
			MsgSeqID: recvProcesserMessage.MsgSeqID,
			Agents:   []*modules.AgentInformation{recvProcesserMessage.Agent},
			UpStream: upStream,
		}
		p.pushToPlugin(sendProcesserMessage)
	}
}

func (p *GSEProcesser) handlePushResponse(message *gse.Message) error {
	p.collector.StatGSEPlatMessageProcessedCount("PUSH_RESP")
	pushRespMessage := &gse.PlatServiceSimpleBody{}

	if err := p.platService.DecodeJSONBody(message.Body, pushRespMessage); err != nil {
		return fmt.Errorf("decode push response message failed, %+v, %+v, %+v",
			message.Header.MainHeader, message.Header.ExtraHeader, err)
	}

	if pushRespMessage.ErrCode != gse.PlatServiceErrCodeOK {
		return fmt.Errorf("push message failed, %+v, %+v, %d",
			message.Header.MainHeader, message.Header.ExtraHeader, pushRespMessage.ErrCode)
	}

	return nil
}
