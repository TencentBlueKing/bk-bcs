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

	"bk-bscp/cmd/atomic-services/bscp-connserver/modules/metrics"
	"bk-bscp/cmd/atomic-services/bscp-connserver/modules/session"
	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	pbdatamanager "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// SignalAction handles signalling channel.
type SignalAction struct {
	ctx        context.Context
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

// NewSignalAction creates new SignalAction.
func NewSignalAction(ctx context.Context, viper *viper.Viper, dataMgrCli pbdatamanager.DataManagerClient,
	sessionMgr *session.Manager, collector *metrics.Collector,
	stream pb.Connection_SignallingChannelServer) *SignalAction {

	action := &SignalAction{ctx: ctx, viper: viper, dataMgrCli: dataMgrCli,
		sessionMgr: sessionMgr, collector: collector, stream: stream}

	return action
}

// Err setup error code message in response and return the error.
func (act *SignalAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	return errors.New(errMsg)
}

// Input handles the input messages.
func (act *SignalAction) Input() error {
	act.pubCh = make(chan interface{})
	act.stopCh = make(chan bool, 1)
	return nil
}

// Output handles the output messages.
func (act *SignalAction) Output() error {
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

func (act *SignalAction) verify(r interface{}) error {
	// #lizard forgives
	switch r.(type) {
	case *pb.SCCMDPing:
		req := r.(*pb.SCCMDPing)

		var err error

		if err = common.ValidateString("biz_id", req.BizId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("app_id", req.AppId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("cloud_id", req.CloudId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("ip", req.Ip,
			database.BSCPNOTEMPTY, database.BSCPNORMALSTRLENLIMIT); err != nil {
			return err
		}
		if req.Timeout == 0 {
			return errors.New("invalid input data, timeout is required")
		}

		if len(req.Labels) == 0 {
			req.Labels = strategy.EmptySidecarLabels
		} else {
			labels := strategy.SidecarLabels{}
			if err := json.Unmarshal([]byte(req.Labels), &labels); err != nil {
				return fmt.Errorf("invalid input data, labels[%+v], %+v", req.Labels, err)
			}
		}

	case *pb.SCCMDPushNotification:
		req := r.(*pb.SCCMDPushNotification)

		var err error

		if err = common.ValidateString("biz_id", req.BizId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("app_id", req.AppId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("cfg_id", req.CfgId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("cfg_name", req.CfgName,
			database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("cfg_fpath", req.CfgFpath,
			database.BSCPEMPTY, database.BSCPCFGFPATHLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("release_id", req.ReleaseId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}

	case *pb.SCCMDPushRollbackNotification:
		req := r.(*pb.SCCMDPushRollbackNotification)

		var err error

		if err = common.ValidateString("biz_id", req.BizId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("app_id", req.AppId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("cfg_id", req.CfgId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("cfg_name", req.CfgName,
			database.BSCPNOTEMPTY, database.BSCPNAMELENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("cfg_fpath", req.CfgFpath,
			database.BSCPEMPTY, database.BSCPCFGFPATHLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("release_id", req.ReleaseId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}

	case *pb.SCCMDPushReloadNotification:
		req := r.(*pb.SCCMDPushReloadNotification)

		var err error

		if err = common.ValidateString("biz_id", req.BizId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if err = common.ValidateString("app_id", req.AppId,
			database.BSCPNOTEMPTY, database.BSCPIDLENLIMIT); err != nil {
			return err
		}
		if req.ReloadSpec == nil || len(req.ReloadSpec.Info) == 0 {
			return errors.New("invalid input data, reload_spec is required")
		}

	default:
		return fmt.Errorf("invalid request type[%+v]", r)
	}
	return nil
}

// onSidecarOnline creates or updates app instance information when the signalling channel is setuped.
func (act *SignalAction) onSidecarOnline(sidecar *session.SidecarInstance) error {
	if time.Now().Sub(act.lastUpdateTime) <= time.Minute {
		return nil
	}
	act.lastUpdateTime = time.Now()

	logger.Info("new sidecar connection, %+v", sidecar)

	r := &pbdatamanager.CreateAppInstanceReq{
		Seq:     common.Sequence(),
		BizId:   sidecar.BizID,
		AppId:   sidecar.AppID,
		CloudId: sidecar.CloudID,
		Ip:      sidecar.IP,
		Path:    sidecar.Path,
		Labels:  sidecar.Labels,
		State:   int32(pbcommon.AppInstanceState_INSS_ONLINE),
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("request to datamanager, %+v", r)

	resp, err := act.dataMgrCli.CreateAppInstance(ctx, r)
	if err != nil {
		logger.Warn("can't request to datamanager CreateAppInstance, %+v", err)
		return err
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		logger.Warn("can't request to datamanager CreateAppInstance, %+v", resp)
		return errors.New(resp.Message)
	}
	return nil
}

// onSidecarOffline updates sidecar information when the signalling channel is canceled.
func (act *SignalAction) onSidecarOffline(sidecar *session.SidecarInstance) error {
	logger.Info("sidecar connection disconnected, %+v", sidecar)

	r := &pbdatamanager.UpdateAppInstanceReq{
		Seq:     common.Sequence(),
		BizId:   sidecar.BizID,
		AppId:   sidecar.AppID,
		CloudId: sidecar.CloudID,
		Ip:      sidecar.IP,
		Path:    sidecar.Path,
		Labels:  sidecar.Labels,
		State:   int32(pbcommon.AppInstanceState_INSS_OFFLINE),
	}

	ctx, cancel := context.WithTimeout(context.Background(), act.viper.GetDuration("datamanager.callTimeout"))
	defer cancel()

	logger.V(4).Infof("request to datamanager, %+v", r)

	resp, err := act.dataMgrCli.UpdateAppInstance(ctx, r)
	if err != nil {
		logger.Warn("can't request to datamanager UpdateAppInstance, %+v", err)
		return err
	}
	if resp.Code != pbcommon.ErrCode_E_OK {
		logger.Warn("can't request to datamanager UpdateAppInstance, %+v", resp)
		return errors.New(resp.Message)
	}
	return nil
}

// closePubCh closes pub-chan in safely.
func (act *SignalAction) closePubCh(ch chan interface{}) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn("close publish channel, repeated close operation, recover success, %+v", err)
		}
	}()
	close(ch)
}

// handleNotification processes publish notification in one coroutine,
// exits when stream call return or any error happens in signalling channel.
func (act *SignalAction) handleNotification(stream pb.Connection_SignallingChannelServer,
	pubCh chan interface{}, stopCh chan bool) {

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
					logger.Error("handleNotification| send publish notification to sidecar[%v], %+v", msg, err)
				} else {
					logger.Info("handleNotification| send publish notification to sidecar success[%v]", msg)
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
					logger.Error("handleNotification| send rollback notification to sidecar[%v], %+v", msg, err)
				} else {
					logger.Info("handleNotification| send rollback notification to sidecar success[%v]", msg)
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
					logger.Error("handleNotification| send reload notification to sidecar[%v], %+v", msg, err)
				} else {
					logger.Info("handleNotification| send reload notification to sidecar success[%v]", msg)
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
func (act *SignalAction) Do() error {
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
				logger.Error("SignallingChannel-PING[%s]| %+v, %+v", r.Seq, r, err)
				return nil
			}

			// save sidecar instance content.
			act.sidecar = session.SidecarInstance{
				BizID:   r.CmdPing.BizId,
				AppID:   r.CmdPing.AppId,
				CloudID: r.CmdPing.CloudId,
				IP:      r.CmdPing.Ip,
				Path:    r.CmdPing.Path,
				Labels:  r.CmdPing.Labels,
			}

			// flush sidecar instance connection session.
			if err := act.sessionMgr.FlushSession(r.CmdPing, act.pubCh); err != nil {
				logger.Error("SignallingChannel-PING[%s]| flush session, %+v, %+v", r.Seq, act.sidecar, err)
				return nil
			}

			// update sidecar instance state.
			if err := act.onSidecarOnline(&act.sidecar); err != nil {
				logger.Error("SignallingChannel-PING[%s]| report online state, %+v, %+v", r.Seq, act.sidecar, err)
			} else {
				act.isSidecarUpdated = true
			}

			// send PONG back to keepalive connection.
			if err := act.stream.Send(&pb.SignallingChannelUpStream{
				Seq: r.Seq,
				Cmd: pb.SignallingChannelCmd_SCCMD_S2C_PONG,
			}); err != nil {
				logger.Error("SignallingChannel-PING[%s]| send PONG back, %+v, %+v", r.Seq, act.sidecar, err)
				return nil
			}
			logger.V(4).Infof("SignallingChannel-PING[%s]| PING success, %+v", r.Seq, act.sidecar)

		default:
			logger.Warn("SignallingChannel| unknow CMD, %+v", r)
		}
	}

	// ultimately, close signalling channel and stop publishing.
	return nil
}
