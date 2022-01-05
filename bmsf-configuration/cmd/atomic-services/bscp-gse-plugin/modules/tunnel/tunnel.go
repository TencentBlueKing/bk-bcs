/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tunnel

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"bk-bscp/cmd/atomic-services/bscp-gse-plugin/modules/gseagent"
	pbcommon "bk-bscp/internal/protocol/common"
	pbtunnelserver "bk-bscp/internal/protocol/tunnelserver"
	"bk-bscp/internal/safeviper"
	"bk-bscp/internal/types"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// Message is message from gse tunnel.
type Message struct {
	// NOTE: DO NOT reuse GTCMD protocol sequence here, it's not the same semantic.
	// In general tunnel stream protocol, the sequence is used to make GSE task Cmd
	// consistent(change async to sync). In GTCMD protocol, the sequence is used for tracing
	// data flow, it cloud be repeated in some flow. If reuse sequence here, it would
	// misunderstand response and handle the wrong data for upstream sidecar/plugin instance.

	// Seq is gse tunnel message sequence.
	Seq string

	// Data is gse tunnel message data.
	Data []byte
}

// Tunnel is gse tunnel impls to communicate with tunnel server.
type Tunnel struct {
	// viper as context.
	viper *safeviper.SafeViper

	// gse agent.
	gseAgent *gseagent.GSEAgent

	// publishing processer.
	pubProcesser func(*pbcommon.Signalling)

	// message map records for pullConfigList.
	pullConfigListMsgMap map[string]chan *Message
	pullConfigListMutex  sync.RWMutex

	// message map records for pullRelease.
	pullReleaseMsgMap map[string]chan *Message
	pullReleaseMutex  sync.RWMutex

	// message map records for queryAppMetadata.
	queryAppMetadataMsgMap map[string]chan *Message
	queryAppMetadataMutex  sync.RWMutex

	// message map records for queryHostProcAttrList.
	queryHostProcAttrListMsgMap map[string]chan *Message
	queryHostProcAttrListMutex  sync.RWMutex
}

// NewTunnel create a new Tunnel object.
func NewTunnel(viper *safeviper.SafeViper, gseAgent *gseagent.GSEAgent, processer func(*pbcommon.Signalling)) *Tunnel {
	return &Tunnel{
		viper:                       viper,
		gseAgent:                    gseAgent,
		pubProcesser:                processer,
		pullConfigListMsgMap:        make(map[string]chan *Message),
		pullReleaseMsgMap:           make(map[string]chan *Message),
		queryAppMetadataMsgMap:      make(map[string]chan *Message),
		queryHostProcAttrListMsgMap: make(map[string]chan *Message),
	}
}

// RecvMessage is a callback func for normal message data fron gse tunnel
func (t *Tunnel) RecvMessage(message *gseagent.Message) {
	// #lizard forgives
	logger.V(4).Infof("TUNNEL| recv message from gse tunnel, meta[%s], len[%d]", message.Meta, len(message.Data))

	// unmarshal GeneralTunnelUpStream base on message.
	msg := pbtunnelserver.GeneralTunnelUpStream{}
	err := proto.Unmarshal(message.Data, &msg)
	if err != nil {
		logger.Errorf("TUNNEL| unmarshal message from gse tunnel failed, meta[%s], %+v", message.Meta, err)
		return
	}
	logger.V(3).Infof("TUNNEL| recv new GeneralTunnelUpStream, meta[%s], Seq[%s] Cmd[%+v] len[%d]",
		message.Meta, msg.Seq, msg.Cmd, len(msg.Data))

	// process GeneralTunnelUpStream.
	var ch chan *Message
	var ok bool

	switch msg.Cmd {
	case pbtunnelserver.GeneralTunnelCmd_GTCMD_C2S_PULL_CONFIGLIST:
		// get message record channel of pullConfigList request.

		t.pullConfigListMutex.Lock()
		if ch, ok = t.pullConfigListMsgMap[msg.Seq]; !ok {
			logger.Errorf("TUNNEL| can't get PullConfigList record channel for meta[%s] seq[%s] cmd[%+v] len[%d]",
				message.Meta, msg.Seq, msg.Cmd, len(msg.Data))

			t.pullConfigListMutex.Unlock()
			return
		}
		// get message record channel of pullConfigList request success.

		delete(t.pullConfigListMsgMap, msg.Seq)
		t.pullConfigListMutex.Unlock()

	case pbtunnelserver.GeneralTunnelCmd_GTCMD_C2S_PULL_RELEASE:
		// get message record channel of pullRelease request.

		t.pullReleaseMutex.Lock()
		if ch, ok = t.pullReleaseMsgMap[msg.Seq]; !ok {
			logger.Errorf("TUNNEL| can't get PullRelease record channel for meta[%s] seq[%s] cmd[%+v] len[%d]",
				message.Meta, msg.Seq, msg.Cmd, len(msg.Data))

			t.pullReleaseMutex.Unlock()
			return
		}
		// get message record channel of pullRelease request success.

		delete(t.pullReleaseMsgMap, msg.Seq)
		t.pullReleaseMutex.Unlock()

	case pbtunnelserver.GeneralTunnelCmd_GTCMD_C2S_QUERY_APP_METADATA:
		// get message record channel of queryAppMetadata request.

		t.queryAppMetadataMutex.Lock()
		if ch, ok = t.queryAppMetadataMsgMap[msg.Seq]; !ok {
			logger.Errorf("TUNNEL| can't get QueryAppMetadata record channel for meta[%s] seq[%s] cmd[%+v] len[%d]",
				message.Meta, msg.Seq, msg.Cmd, len(msg.Data))

			t.queryAppMetadataMutex.Unlock()
			return
		}
		// get message record channel of queryAppMetadata request success.

		delete(t.queryAppMetadataMsgMap, msg.Seq)
		t.queryAppMetadataMutex.Unlock()

	case pbtunnelserver.GeneralTunnelCmd_GTCMD_C2S_QUERY_HOST_PROCATTR_LIST:
		// get message record channel of queryHostProcAttrList request.

		t.queryHostProcAttrListMutex.Lock()
		if ch, ok = t.queryHostProcAttrListMsgMap[msg.Seq]; !ok {
			logger.Errorf("TUNNEL| can't get QueryHostProcAttrList record chan for meta[%s] seq[%s] cmd[%+v] len[%d]",
				message.Meta, msg.Seq, msg.Cmd, len(msg.Data))

			t.queryHostProcAttrListMutex.Unlock()
			return
		}
		// get message record channel of queryHostProcAttrList request success.

		delete(t.queryHostProcAttrListMsgMap, msg.Seq)
		t.queryHostProcAttrListMutex.Unlock()

	case pbtunnelserver.GeneralTunnelCmd_GTCMD_S2C_PUBLISH:
		// process publish message.
		publishStruct := &pbtunnelserver.GTCMDPublish{}
		if err := proto.Unmarshal(msg.Data, publishStruct); err != nil {
			logger.Errorf("TUNNEL| unmashal message[%d] from gse tunnel to GTCMDPublish failed, meta[%s], %+v",
				len(msg.Data), message.Meta, err)
			return
		}
		go t.processPublish(publishStruct)
		return

	case pbtunnelserver.GeneralTunnelCmd_GTCMD_S2C_ROLLBACK:
		// process rollback message.
		rollbackStruct := &pbtunnelserver.GTCMDRollback{}
		if err := proto.Unmarshal(msg.Data, rollbackStruct); err != nil {
			logger.Errorf("TUNNEL| unmashal message[%d] from gse tunnel to GTCMDRollback failed, meta[%s], %+v",
				len(msg.Data), message.Meta, err)
			return
		}
		go t.processRollback(rollbackStruct)
		return

	case pbtunnelserver.GeneralTunnelCmd_GTCMD_S2C_RELOAD:
		// process reload message.
		reloadStruct := &pbtunnelserver.GTCMDReload{}
		if err := proto.Unmarshal(msg.Data, reloadStruct); err != nil {
			logger.Errorf("TUNNEL| unmashal message[%d] from gse tunnel to GTCMDReload failed, meta[%s], %+v",
				len(msg.Data), message.Meta, err)
			return
		}
		go t.processReload(reloadStruct)
		return

	default:
		logger.Warnf("TUNNEL| recved invalid message type[%+v] from gse tunnel, meta[%s]", msg.Cmd, message.Meta)
		return
	}

	select {
	case ch <- &Message{Seq: msg.Seq, Data: msg.Data}:
	case <-time.After(time.Second):
		logger.Warnf("TUNNEL| send normal message type[%+v] from gse tunnel to record cahnnel timeout, meta[%s]",
			msg.Cmd, message.Meta)
	}
}

// processPublish processes publish message.
func (t *Tunnel) processPublish(msg *pbtunnelserver.GTCMDPublish) {
	t.pubProcesser(&pbcommon.Signalling{
		Type: pbcommon.SignallingType_ST_SignallingTypePublish,
		Publishing: &pbcommon.Publishing{
			BizId:      msg.BizId,
			AppId:      msg.AppId,
			CfgId:      msg.CfgId,
			CfgName:    msg.CfgName,
			CfgFpath:   msg.CfgFpath,
			Serialno:   msg.Serialno,
			ReleaseId:  msg.ReleaseId,
			Strategies: msg.Strategies,
			Nice:       msg.Nice,
		},
	})
}

// processRollback processes rollback message.
func (t *Tunnel) processRollback(msg *pbtunnelserver.GTCMDRollback) {
	t.pubProcesser(&pbcommon.Signalling{
		Type: pbcommon.SignallingType_ST_SignallingTypeRollback,
		Publishing: &pbcommon.Publishing{
			BizId:      msg.BizId,
			AppId:      msg.AppId,
			CfgId:      msg.CfgId,
			CfgName:    msg.CfgName,
			CfgFpath:   msg.CfgFpath,
			Serialno:   msg.Serialno,
			ReleaseId:  msg.ReleaseId,
			Strategies: msg.Strategies,
			Nice:       msg.Nice,
		},
	})
}

// processReload processes reload message.
func (t *Tunnel) processReload(msg *pbtunnelserver.GTCMDReload) {
	t.pubProcesser(&pbcommon.Signalling{
		Type: pbcommon.SignallingType_ST_SignallingTypeReload,
		Publishing: &pbcommon.Publishing{
			BizId:      msg.BizId,
			AppId:      msg.AppId,
			Strategies: msg.Strategies,
			ReloadSpec: msg.ReloadSpec,
			Nice:       msg.Nice,
		},
	})
}

// PullConfigList returns config list.
func (t *Tunnel) PullConfigList(messageID uint64,
	req *pbtunnelserver.GTCMDPullConfigListReq) (*pbtunnelserver.GTCMDPullConfigListResp, error) {

	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	innerReq := &pbtunnelserver.GeneralTunnelDownStream{
		// NOTE: DO NOT reuse GTCMD protocol sequence here, it's not the same semantic.
		// In general tunnel stream protocol, the sequence is used to make GSE task Cmd
		// consistent(change async to sync). In GTCMD protocol, the sequence is used for tracing
		// data flow, it cloud be repeated in some flow. If reuse sequence here, it would
		// misunderstand response and handle the wrong data for upstream sidecar/plugin instance.
		Seq:  common.Sequence(),
		Cmd:  pbtunnelserver.GeneralTunnelCmd_GTCMD_C2S_PULL_CONFIGLIST,
		Data: reqBytes,
	}
	bytes, err := proto.Marshal(innerReq)
	if err != nil {
		return nil, err
	}

	// channel to receive response.
	ch := make(chan *Message, 1)

	t.pullConfigListMutex.Lock()
	t.pullConfigListMsgMap[innerReq.Seq] = ch
	t.pullConfigListMutex.Unlock()

	// request base on gse tunnel.
	logger.V(3).Infof("TUNNEL| send PullConfigList[%d][%s] request[%d] transmitType[%+v]",
		messageID, innerReq.Seq, len(bytes), gseagent.TTRAND)

	t.gseAgent.SendMessage(bytes, 0, messageID, gseagent.TTRAND)

	// wait for response with timeout.
	timer := time.NewTimer(t.viper.GetDuration("server.gseTunnelTimeout"))

	select {
	case <-timer.C:
		// timeout now, delete message record channel, and response can cancel base on it earlier.
		t.pullConfigListMutex.Lock()
		delete(t.pullConfigListMsgMap, innerReq.Seq)
		t.pullConfigListMutex.Unlock()

		return nil, types.ErrorTimeout

	case msg := <-ch:
		resp := &pbtunnelserver.GTCMDPullConfigListResp{}
		if err := proto.Unmarshal(msg.Data, resp); err != nil {
			return nil, err
		}

		if resp.Seq != req.Seq {
			return nil, fmt.Errorf("recved inconformity sequence")
		}
		return resp, nil
	}
}

// PullRelease returns release.
func (t *Tunnel) PullRelease(messageID uint64,
	req *pbtunnelserver.GTCMDPullReleaseReq) (*pbtunnelserver.GTCMDPullReleaseResp, error) {

	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	innerReq := &pbtunnelserver.GeneralTunnelDownStream{
		// NOTE: DO NOT reuse GTCMD protocol sequence here, it's not the same semantic.
		// In general tunnel stream protocol, the sequence is used to make GSE task Cmd
		// consistent(change async to sync). In GTCMD protocol, the sequence is used for tracing
		// data flow, it cloud be repeated in some flow. If reuse sequence here, it would
		// misunderstand response and handle the wrong data for upstream sidecar/plugin instance.
		Seq:  common.Sequence(),
		Cmd:  pbtunnelserver.GeneralTunnelCmd_GTCMD_C2S_PULL_RELEASE,
		Data: reqBytes,
	}
	bytes, err := proto.Marshal(innerReq)
	if err != nil {
		return nil, err
	}

	// channel to receive response.
	ch := make(chan *Message, 1)

	t.pullReleaseMutex.Lock()
	t.pullReleaseMsgMap[innerReq.Seq] = ch
	t.pullReleaseMutex.Unlock()

	// request base on gse tunnel.
	logger.V(3).Infof("TUNNEL| send PullRelease[%d][%s] request[%d] transmitType[%+v] GTCMDReq[%s]",
		messageID, innerReq.Seq, len(bytes), gseagent.TTRAND, req.Seq)

	t.gseAgent.SendMessage(bytes, 0, messageID, gseagent.TTRAND)

	// wait for response with timeout.
	timer := time.NewTimer(t.viper.GetDuration("server.gseTunnelTimeout"))

	select {
	case <-timer.C:
		// timeout now, delete message record channel, and response can cancel base on it earlier.
		t.pullReleaseMutex.Lock()
		delete(t.pullReleaseMsgMap, innerReq.Seq)
		t.pullReleaseMutex.Unlock()

		return nil, types.ErrorTimeout

	case msg := <-ch:
		resp := &pbtunnelserver.GTCMDPullReleaseResp{}
		if err := proto.Unmarshal(msg.Data, resp); err != nil {
			return nil, err
		}

		if resp.Seq != req.Seq {
			return nil, fmt.Errorf("recved inconformity sequence")
		}
		return resp, nil
	}
}

// QueryAppMetadata query app metadata.
func (t *Tunnel) QueryAppMetadata(messageID uint64,
	req *pbtunnelserver.GTCMDQueryAppMetadataReq) (*pbtunnelserver.GTCMDQueryAppMetadataResp, error) {

	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	innerReq := &pbtunnelserver.GeneralTunnelDownStream{
		// NOTE: DO NOT reuse GTCMD protocol sequence here, it's not the same semantic.
		// In general tunnel stream protocol, the sequence is used to make GSE task Cmd
		// consistent(change async to sync). In GTCMD protocol, the sequence is used for tracing
		// data flow, it cloud be repeated in some flow. If reuse sequence here, it would
		// misunderstand response and handle the wrong data for upstream sidecar/plugin instance.
		Seq:  common.Sequence(),
		Cmd:  pbtunnelserver.GeneralTunnelCmd_GTCMD_C2S_QUERY_APP_METADATA,
		Data: reqBytes,
	}
	bytes, err := proto.Marshal(innerReq)
	if err != nil {
		return nil, err
	}

	// channel to receive response.
	ch := make(chan *Message, 1)

	t.queryAppMetadataMutex.Lock()
	t.queryAppMetadataMsgMap[innerReq.Seq] = ch
	t.queryAppMetadataMutex.Unlock()

	// request base on gse tunnel.
	logger.V(3).Infof("TUNNEL| send QueryAppMetadata[%d][%s] request[%d] transmitType[%+v] GTCMDReq[%s]",
		messageID, innerReq.Seq, len(bytes), gseagent.TTRAND, req.Seq)

	t.gseAgent.SendMessage(bytes, 0, messageID, gseagent.TTRAND)

	// wait for response with timeout.
	timer := time.NewTimer(t.viper.GetDuration("server.gseTunnelTimeout"))

	select {
	case <-timer.C:
		// timeout now, delete message record channel, and response can cancel base on it earlier.
		t.queryAppMetadataMutex.Lock()
		delete(t.queryAppMetadataMsgMap, innerReq.Seq)
		t.queryAppMetadataMutex.Unlock()

		return nil, types.ErrorTimeout

	case msg := <-ch:
		resp := &pbtunnelserver.GTCMDQueryAppMetadataResp{}
		if err := proto.Unmarshal(msg.Data, resp); err != nil {
			return nil, err
		}

		if resp.Seq != req.Seq {
			return nil, fmt.Errorf("recved inconformity sequence")
		}
		return resp, nil
	}
}

// QueryHostProcAttrList query proc attrs on host.
func (t *Tunnel) QueryHostProcAttrList(messageID uint64,
	req *pbtunnelserver.GTCMDQueryHostProcAttrListReq) (*pbtunnelserver.GTCMDQueryHostProcAttrListResp, error) {

	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	innerReq := &pbtunnelserver.GeneralTunnelDownStream{
		// NOTE: DO NOT reuse GTCMD protocol sequence here, it's not the same semantic.
		// In general tunnel stream protocol, the sequence is used to make GSE task Cmd
		// consistent(change async to sync). In GTCMD protocol, the sequence is used for tracing
		// data flow, it cloud be repeated in some flow. If reuse sequence here, it would
		// misunderstand response and handle the wrong data for upstream sidecar/plugin instance.
		Seq:  common.Sequence(),
		Cmd:  pbtunnelserver.GeneralTunnelCmd_GTCMD_C2S_QUERY_HOST_PROCATTR_LIST,
		Data: reqBytes,
	}
	bytes, err := proto.Marshal(innerReq)
	if err != nil {
		return nil, err
	}

	// channel to receive response.
	ch := make(chan *Message, 1)

	t.queryHostProcAttrListMutex.Lock()
	t.queryHostProcAttrListMsgMap[innerReq.Seq] = ch
	t.queryHostProcAttrListMutex.Unlock()

	// request base on gse tunnel.
	logger.V(3).Infof("TUNNEL| send QueryHostProcAttrList[%d][%s] request[%d] transmitType[%+v] GTCMDReq[%s]",
		messageID, innerReq.Seq, len(bytes), gseagent.TTRAND, req.Seq)

	t.gseAgent.SendMessage(bytes, 0, messageID, gseagent.TTRAND)

	// wait for response with timeout.
	timer := time.NewTimer(t.viper.GetDuration("server.gseTunnelTimeout"))

	select {
	case <-timer.C:
		// timeout now, delete message record channel, and response can cancel base on it earlier.
		t.queryHostProcAttrListMutex.Lock()
		delete(t.queryHostProcAttrListMsgMap, innerReq.Seq)
		t.queryHostProcAttrListMutex.Unlock()

		return nil, types.ErrorTimeout

	case msg := <-ch:
		resp := &pbtunnelserver.GTCMDQueryHostProcAttrListResp{}
		if err := proto.Unmarshal(msg.Data, resp); err != nil {
			return nil, err
		}

		if resp.Seq != req.Seq {
			return nil, fmt.Errorf("recved inconformity sequence")
		}
		return resp, nil
	}
}

// EffectReport reports effect infos.
func (t *Tunnel) EffectReport(messageID uint64, req *pbtunnelserver.GTCMDEffectReport) error {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	innerReq := &pbtunnelserver.GeneralTunnelDownStream{
		// NOTE: DO NOT reuse GTCMD protocol sequence here, it's not the same semantic.
		// In general tunnel stream protocol, the sequence is used to make GSE task Cmd
		// consistent(change async to sync). In GTCMD protocol, the sequence is used for tracing
		// data flow, it cloud be repeated in some flow. If reuse sequence here, it would
		// misunderstand response and handle the wrong data for upstream sidecar/plugin instance.
		Seq:  common.Sequence(),
		Cmd:  pbtunnelserver.GeneralTunnelCmd_GTCMD_C2S_EFFECT_REPORT,
		Data: reqBytes,
	}
	bytes, err := proto.Marshal(innerReq)
	if err != nil {
		return err
	}

	// request base on gse tunnel.
	logger.V(3).Infof("TUNNEL| send EffectReport[%d][%s] request[%d] transmitType[%+v] GTCMDReq[%s]",
		messageID, innerReq.Seq, len(bytes), gseagent.TTRAND, req.Seq)

	t.gseAgent.SendMessage(bytes, 0, messageID, gseagent.TTRAND)

	return nil
}

// PluginInfo reports plugin infos.
func (t *Tunnel) PluginInfo(messageID uint64, req *pbtunnelserver.GTCMDPluginInfo) error {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	innerReq := &pbtunnelserver.GeneralTunnelDownStream{
		// NOTE: DO NOT reuse GTCMD protocol sequence here, it's not the same semantic.
		// In general tunnel stream protocol, the sequence is used to make GSE task Cmd
		// consistent(change async to sync). In GTCMD protocol, the sequence is used for tracing
		// data flow, it cloud be repeated in some flow. If reuse sequence here, it would
		// misunderstand response and handle the wrong data for upstream sidecar/plugin instance.
		Seq:  common.Sequence(),
		Cmd:  pbtunnelserver.GeneralTunnelCmd_GTCMD_C2S_PLUGIN_INFO,
		Data: reqBytes,
	}
	bytes, err := proto.Marshal(innerReq)
	if err != nil {
		return err
	}

	// request base on gse tunnel.
	logger.V(3).Infof("TUNNEL| send PluginInfo[%d][%s] request[%d] transmitType[%+v]",
		messageID, innerReq.Seq, len(bytes), gseagent.TTBROADCAST)

	t.gseAgent.SendMessage(bytes, 0, messageID, gseagent.TTBROADCAST)

	return nil
}

// Close closes tunnel.
func (t *Tunnel) Close() {
}
