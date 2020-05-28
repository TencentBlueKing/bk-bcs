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

package signalling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/viper"

	"bk-bscp/cmd/bscp-connserver/modules/metrics"
	"bk-bscp/cmd/bscp-connserver/modules/session"
	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// SignallingAction handles signalling channel.
type SignallingAction struct {
	viper      *viper.Viper
	dataMgrCli pbdatamanager.DataManagerClient

	// sidecar session manager.
	sessionMgr *session.Manager

	// prometheus metrics collector.
	collector *metrics.Collector

	// current signalling stream.
	stream pb.Connection_SignallingChannelServer

	// current sidecar instance, used for connection session operations.
	sidecar session.SidecarInstance

	// mark if the sidecar instance is updated already.
	isSidecarUpdated bool

	// last sidecar instance update time.
	lastUpdateTime time.Time

	// channel for publish notification.
	pubCh chan interface{}

	// recvice stopping-signal from this channel, and exit processing coroutine.
	stopCh chan bool
}

// NewSignallingAction creates new SignallingAction.
func NewSignallingAction(viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	sessionMgr *session.Manager, collector *metrics.Collector, stream pb.Connection_SignallingChannelServer) *SignallingAction {
	action := &SignallingAction{viper: viper, dataMgrCli: dataMgrCli, sessionMgr: sessionMgr, collector: collector, stream: stream}
	return action
}

// Err setup error code message in response and return the error.
func (act *SignallingAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *SignallingAction) Input() error {
	act.pubCh = make(chan interface{})
	act.stopCh = make(chan bool, 1)
	return nil
}

// Output handles the output messages.
func (act *SignallingAction) Output() error {
	logger.Info("SignallingChannel| signalling channel shutdown now, %+v", act.sidecar)

	// stop publish notification, SignallingChannel shutdown now.
	act.stopCh <- true

	// delete session, ignore error when sidecar instance content missing.
	act.sessionMgr.DeleteSession(&act.sidecar)

	// sidecar instance offline.
	if act.isSidecarUpdated {
		act.onSidecarOffline(&act.sidecar)
	}

	return nil
}

func (act *SignallingAction) verify(r interface{}) error {
	// #lizard forgives
	switch r.(type) {
	case *pb.SCCMDPing:
		req := r.(*pb.SCCMDPing)

		length := len(req.Bid)
		if length == 0 {
			return errors.New("invalid params, bid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, bid too long")
		}

		length = len(req.Appid)
		if length == 0 {
			return errors.New("invalid params, appid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, appid too long")
		}

		length = len(req.Clusterid)
		if length == 0 {
			return errors.New("invalid params, clusterid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, clusterid too long")
		}

		length = len(req.Zoneid)
		if length == 0 {
			return errors.New("invalid params, zoneid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, zoneid too long")
		}

		length = len(req.Dc)
		if length == 0 {
			return errors.New("invalid params, dc missing")
		}
		if length > database.BSCPNAMELENLIMIT {
			return errors.New("invalid params, dc too long")
		}

		length = len(req.IP)
		if length == 0 {
			return errors.New("invalid params, ip missing")
		}
		if length > database.BSCPNORMALSTRLENLIMIT {
			return errors.New("invalid params, ip too long")
		}

		if req.Timeout == 0 {
			return errors.New("invalid params, timeout missing")
		}

		if req.Labels == "" {
			req.Labels = strategy.EmptySidecarLabels
		} else {
			labels := strategy.SidecarLabels{}
			if err := json.Unmarshal([]byte(req.Labels), &labels); err != nil {
				return fmt.Errorf("invalid params, labels[%+v], %+v", req.Labels, err)
			}
		}

	case *pb.SCCMDPushNotification:
		req := r.(*pb.SCCMDPushNotification)

		length := len(req.Bid)
		if length == 0 {
			return errors.New("invalid params, bid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, bid too long")
		}

		length = len(req.Appid)
		if length == 0 {
			return errors.New("invalid params, appid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, appid too long")
		}

		length = len(req.Cfgsetid)
		if length == 0 {
			return errors.New("invalid params, cfgsetid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, cfgsetid too long")
		}

		length = len(req.CfgsetName)
		if length == 0 {
			return errors.New("invalid params, cfgsetname missing")
		}
		if length > database.BSCPNAMELENLIMIT {
			return errors.New("invalid params, cfgsetname too long")
		}
		if len(req.CfgsetFpath) > database.BSCPCFGSETFPATHLENLIMIT {
			return errors.New("invalid params, cfgsetFpath too long")
		}

		length = len(req.Releaseid)
		if length == 0 {
			return errors.New("invalid params, releaseid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, releaseid too long")
		}

	case *pb.SCCMDPushRollbackNotification:
		req := r.(*pb.SCCMDPushRollbackNotification)

		length := len(req.Bid)
		if length == 0 {
			return errors.New("invalid params, bid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, bid too long")
		}

		length = len(req.Appid)
		if length == 0 {
			return errors.New("invalid params, appid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, appid too long")
		}

		length = len(req.Cfgsetid)
		if length == 0 {
			return errors.New("invalid params, cfgsetid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, cfgsetid too long")
		}

		length = len(req.CfgsetName)
		if length == 0 {
			return errors.New("invalid params, cfgsetname missing")
		}
		if length > database.BSCPNAMELENLIMIT {
			return errors.New("invalid params, cfgsetname too long")
		}
		if len(req.CfgsetFpath) > database.BSCPCFGSETFPATHLENLIMIT {
			return errors.New("invalid params, cfgsetFpath too long")
		}

		length = len(req.Releaseid)
		if length == 0 {
			return errors.New("invalid params, releaseid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, releaseid too long")
		}

	case *pb.SCCMDPushReloadNotification:
		req := r.(*pb.SCCMDPushReloadNotification)

		length := len(req.Bid)
		if length == 0 {
			return errors.New("invalid params, bid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, bid too long")
		}

		length = len(req.Appid)
		if length == 0 {
			return errors.New("invalid params, appid missing")
		}
		if length > database.BSCPIDLENLIMIT {
			return errors.New("invalid params, appid too long")
		}
		if req.ReloadSpec == nil || len(req.ReloadSpec.Info) == 0 {
			return errors.New("invalid params, reloadSpec missing")
		}

	default:
		return fmt.Errorf("invalid request type[%+v]", r)
	}
	return nil
}

// onSidecarOnline creates or updates app instance information when the signalling channel is setuped.
func (act *SignallingAction) onSidecarOnline(sidecar *session.SidecarInstance) error {
	if time.Now().Sub(act.lastUpdateTime) <= time.Minute {
		return nil
	}
	act.lastUpdateTime = time.Now()

	logger.Info("new sidecar connection, %+v", sidecar)

	r := &pbdatamanager.CreateAppInstanceReq{
		Seq:       common.Sequence(),
		Bid:       sidecar.Bid,
		Appid:     sidecar.Appid,
		Clusterid: sidecar.Clusterid,
		Zoneid:    sidecar.Zoneid,
		Dc:        sidecar.Dc,
		IP:        sidecar.IP,
		Labels:    sidecar.Labels,
		State:     int32(pbcommon.AppInstanceState_INSS_ONLINE),
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("request to datamanager CreateAppInstance, %+v", r)

	resp, err := act.dataMgrCli.CreateAppInstance(ctx, r)
	if err != nil {
		logger.Warn("can't request to datamanager CreateAppInstance, %+v", err)
		return err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		logger.Warn("can't request to datamanager CreateAppInstance, %+v", resp)
		return errors.New(resp.ErrMsg)
	}
	return nil
}

// onSidecarOffline updates sidecar information when the signalling channel is canceled.
func (act *SignallingAction) onSidecarOffline(sidecar *session.SidecarInstance) error {
	logger.Info("sidecar connection disconnected, %+v", sidecar)

	r := &pbdatamanager.UpdateAppInstanceReq{
		Seq:       common.Sequence(),
		Bid:       sidecar.Bid,
		Appid:     sidecar.Appid,
		Clusterid: sidecar.Clusterid,
		Zoneid:    sidecar.Zoneid,
		Dc:        sidecar.Dc,
		IP:        sidecar.IP,
		Labels:    sidecar.Labels,
		State:     int32(pbcommon.AppInstanceState_INSS_OFFLINE),
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.calltimeout"))
	defer cancel()

	logger.V(2).Infof("request to datamanager UpdateAppInstance, %+v", r)

	resp, err := act.dataMgrCli.UpdateAppInstance(ctx, r)
	if err != nil {
		logger.Warn("can't request to datamanager UpdateAppInstance, %+v", err)
		return err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		logger.Warn("can't request to datamanager UpdateAppInstance, %+v", resp)
		return errors.New(resp.ErrMsg)
	}
	return nil
}

// closePubCh closes pub-chan in safely.
func (act *SignallingAction) closePubCh(ch chan interface{}) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn("close publish channel, repeated close operation, recover success, %+v", err)
		}
	}()
	close(ch)
}

// handleNotification processes publish notification in one coroutine,
// exits when stream call return or any error happens in signalling channel.
func (act *SignallingAction) handleNotification(stream pb.Connection_SignallingChannelServer, pubCh chan interface{}, stopCh chan bool) {
	for {
		select {
		case <-stopCh:
			// stop processing publish notification.
			act.closePubCh(pubCh)
			return

		case notification := <-pubCh:
			// push notification to sidecar.
			if err := act.verify(notification); err != nil {
				logger.Error("handleNotification| verify proto, %+v, %+v", notification, err)
				continue
			}

			switch notification.(type) {
			case *pb.SCCMDPushNotification:
				msg := notification.(*pb.SCCMDPushNotification)

				err := stream.Send(&pb.SignallingChannelUpStream{
					Seq:     common.Sequence(),
					Cmd:     pb.SignallingChannelCmd_SCCMD_S2C_PUSH_NOTIFICATION,
					CmdPush: msg,
				})
				if err != nil {
					logger.Error("handleNotification| send publish notification to sidecar, notification[%v], %+v", msg, err)
				} else {
					logger.Info("handleNotification| send publish notification to sidecar success, notification[%v]", msg)
				}
				act.collector.StatPublishing(err == nil)
				continue

			case *pb.SCCMDPushRollbackNotification:
				msg := notification.(*pb.SCCMDPushRollbackNotification)

				err := stream.Send(&pb.SignallingChannelUpStream{
					Seq:         common.Sequence(),
					Cmd:         pb.SignallingChannelCmd_SCCMD_S2C_PUSH_ROLLBACK_NOTIFICATION,
					CmdRollback: msg,
				})
				if err != nil {
					logger.Error("handleNotification| send rollback publish notification to sidecar, notification[%v], %+v", msg, err)
				} else {
					logger.Info("handleNotification| send rollback publish notification to sidecar success, notification[%v]", msg)
				}
				act.collector.StatPublishing(err == nil)
				continue

			case *pb.SCCMDPushReloadNotification:
				msg := notification.(*pb.SCCMDPushReloadNotification)

				err := stream.Send(&pb.SignallingChannelUpStream{
					Seq:       common.Sequence(),
					Cmd:       pb.SignallingChannelCmd_SCCMD_S2C_PUSH_RELOAD_NOTIFICATION,
					CmdReload: msg,
				})
				if err != nil {
					logger.Error("handleNotification| send reload publish notification to sidecar, notification[%v], %+v", msg, err)
				} else {
					logger.Info("handleNotification| send reload publish notification to sidecar success, notification[%v]", msg)
				}
				act.collector.StatPublishing(err == nil)
				continue

			default:
				logger.Error("handleNotification| invalid proto, %+v", notification)
				continue
			}
		}
	}
}

// Do makes the workflows of this action base on input messages.
func (act *SignallingAction) Do() error {
	// watch publish channel, then push to sidecar when notification coming.
	go act.handleNotification(act.stream, act.pubCh, act.stopCh)

	// processing signalling channel.
	for {
		// TODO cancel quiet stream.
		r, err := act.stream.Recv()
		if err == io.EOF {
			logger.Warn("SignallingChannel| signalling channel closing, %+v", act.stream)
			break
		}

		if err != nil {
			logger.Error("SignallingChannel| recvice stream from sidecar, %+v, %+v", act.stream, err)
			return nil
		}

		switch r.Cmd {
		case pb.SignallingChannelCmd_SCCMD_C2S_PING:
			if err := act.verify(r.CmdPing); err != nil {
				logger.Error("SignallingChannel-PING[%d]| %+v, %+v", r.Seq, r, err)
				return nil
			}

			// save sidecar instance content.
			act.sidecar = session.SidecarInstance{
				Bid:       r.CmdPing.Bid,
				Appid:     r.CmdPing.Appid,
				Clusterid: r.CmdPing.Clusterid,
				Zoneid:    r.CmdPing.Zoneid,
				Dc:        r.CmdPing.Dc,
				IP:        r.CmdPing.IP,
				Labels:    r.CmdPing.Labels,
			}

			// flush sidecar instance connection session.
			if err := act.sessionMgr.FlushSession(r.CmdPing, act.pubCh); err != nil {
				logger.Error("SignallingChannel-PING[%d]| flush session, %+v, %+v", r.Seq, act.sidecar, err)
				return nil
			}

			// update sidecar instance state.
			if err := act.onSidecarOnline(&act.sidecar); err == nil {
				act.isSidecarUpdated = true
			}

			// send PONG back to keepalive connection.
			if err := act.stream.Send(&pb.SignallingChannelUpStream{Seq: r.Seq, Cmd: pb.SignallingChannelCmd_SCCMD_S2C_PONG}); err != nil {
				logger.Error("SignallingChannel-PING[%d]| send PONG back, %+v, %+v", r.Seq, act.sidecar, err)
				return nil
			}
			logger.V(2).Infof("SignallingChannel-PING[%d]| PING success, %+v", r.Seq, act.sidecar)

		default:
			logger.Warn("SignallingChannel| unknow CMD, %+v", r)
		}
	}

	// ultimately, close signalling channel and stop publishing.
	return nil
}
