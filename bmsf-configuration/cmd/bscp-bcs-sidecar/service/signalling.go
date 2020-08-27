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
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// SignallingChannel handles signalling channel
// between connserver and sidecar.
type SignallingChannel struct {
	// configs handler.
	viper *viper.Viper

	businessName string
	appName      string
	path         string

	// config handler.
	handler *Handler
}

// NewSignallingChannel creates new SignallingChannel.
func NewSignallingChannel(viper *viper.Viper, businessName, appName, path string, handler *Handler) *SignallingChannel {
	return &SignallingChannel{
		viper:        viper,
		businessName: businessName,
		appName:      appName,
		path:         path,
		handler:      handler,
	}
}

func (sc *SignallingChannel) makeConnectionClient() (*grpc.ClientConn, pb.ConnectionClient, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(sc.viper.GetDuration("connserver.dialtimeout")),
	}

	endpoint := sc.viper.GetString("connserver.hostname") + ":" + sc.viper.GetString("connserver.port")
	c, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, nil, err
	}
	client := pb.NewConnectionClient(c)

	return c, client, nil
}

func (sc *SignallingChannel) access() ([]string, error) {
	c, client, err := sc.makeConnectionClient()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	modKey := ModKey(sc.businessName, sc.appName, sc.path)

	sidecarLabels := &strategy.SidecarLabels{Labels: sc.viper.GetStringMapString(fmt.Sprintf("appmod.%s.labels", modKey))}
	labels, err := json.Marshal(sidecarLabels)
	if err != nil {
		return nil, err
	}

	r := &pb.AccessReq{
		Seq:       common.Sequence(),
		Bid:       sc.viper.GetString(fmt.Sprintf("appmod.%s.bid", modKey)),
		Appid:     sc.viper.GetString(fmt.Sprintf("appmod.%s.appid", modKey)),
		Clusterid: sc.viper.GetString(fmt.Sprintf("appmod.%s.clusterid", modKey)),
		Zoneid:    sc.viper.GetString(fmt.Sprintf("appmod.%s.zoneid", modKey)),
		Dc:        sc.viper.GetString(fmt.Sprintf("appmod.%s.dc", modKey)),
		IP:        sc.viper.GetString("appinfo.ip"),
		Labels:    string(labels),
	}

	ctx, cancel := context.WithTimeout(context.Background(), sc.viper.GetDuration("connserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("SignallingChannel[%s %s %s]| request to connserver Access, %+v", sc.businessName, sc.appName, sc.path, r)

	resp, err := client.Access(ctx, r)
	if err != nil {
		return nil, err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return nil, fmt.Errorf("can't access to connserver, %+v", resp)
	}
	if resp.Endpoints == nil || len(resp.Endpoints) == 0 {
		return nil, fmt.Errorf("can't access to connserver, no available node")
	}

	nodes := []string{}
	for _, endpoint := range resp.Endpoints {
		nodes = append(nodes, fmt.Sprintf("%s:%d", endpoint.IP, endpoint.Port))
	}
	return nodes, nil
}

func (sc *SignallingChannel) ping(stream pb.Connection_SignallingChannelClient) error {
	modKey := ModKey(sc.businessName, sc.appName, sc.path)

	// sidecar labels.
	sidecarLabels := &strategy.SidecarLabels{Labels: sc.viper.GetStringMapString(fmt.Sprintf("appmod.%s.labels", modKey))}
	labels, err := json.Marshal(sidecarLabels)
	if err != nil {
		return err
	}

	// session timeout with coefficient.
	sessionTimeout := sc.viper.GetInt64("sidecar.sessionCoefficient") *
		int64(sc.viper.GetDuration("sidecar.sessionTimeout")/time.Second)

	// PING command.
	r := &pb.SignallingChannelDownStream{
		Seq: common.Sequence(),
		Cmd: pb.SignallingChannelCmd_SCCMD_C2S_PING,
		CmdPing: &pb.SCCMDPing{
			Bid:       sc.viper.GetString(fmt.Sprintf("appmod.%s.bid", modKey)),
			Appid:     sc.viper.GetString(fmt.Sprintf("appmod.%s.appid", modKey)),
			Clusterid: sc.viper.GetString(fmt.Sprintf("appmod.%s.clusterid", modKey)),
			Zoneid:    sc.viper.GetString(fmt.Sprintf("appmod.%s.zoneid", modKey)),
			Dc:        sc.viper.GetString(fmt.Sprintf("appmod.%s.dc", modKey)),
			IP:        sc.viper.GetString("appinfo.ip"),
			Labels:    string(labels),
			Timeout:   sessionTimeout,
		},
	}
	// send PING command.
	return stream.Send(r)
}

func (sc *SignallingChannel) pinging(ctx context.Context, switchCh chan struct{}, stream pb.Connection_SignallingChannelClient) error {
	// ping at first.
	if err := sc.ping(stream); err != nil {
		logger.Error("SignallingChannel[%s %s %s]| pinging failed at first, switch stream, %+v", sc.businessName, sc.appName, sc.path, err)
		switchCh <- struct{}{}
		return err
	}

	// keep pinging.
	for {
		if sc.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(sc.businessName, sc.appName, sc.path))) {
			logger.Info("SignallingChannel[%s %s %s]| signalling stop pinging now!", sc.businessName, sc.appName, sc.path)
			switchCh <- struct{}{}
			return nil
		}

		select {
		case <-ctx.Done():
			logger.Info("SignallingChannel[%s %s %s]| cancel pinging now.", sc.businessName, sc.appName, sc.path)
			return nil

		case <-time.After(sc.viper.GetDuration("sidecar.sessionTimeout")):
			if err := sc.ping(stream); err != nil {
				logger.Error("SignallingChannel[%s %s %s]| pinging failed, switch stream, %+v", sc.businessName, sc.appName, sc.path, err)
				switchCh <- struct{}{}
				return err
			}
			logger.Info("SignallingChannel[%s %s %s]| CMD -- sent PING.", sc.businessName, sc.appName, sc.path)
		}
	}
}

func (sc *SignallingChannel) signalling(ctx context.Context, switchCh chan struct{}, stream pb.Connection_SignallingChannelClient) error {
	// keep pinging.
	go sc.pinging(ctx, switchCh, stream)

	// stop recving signal.
	stopRecving := false

	go func(ctx context.Context, stopRecving *bool) {
		<-ctx.Done()
		*stopRecving = true
	}(ctx, &stopRecving)

	for {
		if stopRecving || sc.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(sc.businessName, sc.appName, sc.path))) {
			logger.Info("SignallingChannel[%s %s %s]| cancel and stop recving now.", sc.businessName, sc.appName, sc.path)
			stream.CloseSend()
			switchCh <- struct{}{}
			break
		}

		// TODO cancel quiet stream.
		resp, err := stream.Recv()
		if err == io.EOF {
			logger.Info("SignallingChannel[%s %s %s]| stream recv closing now.", sc.businessName, sc.appName, sc.path)
			stream.CloseSend()
			switchCh <- struct{}{}
			break
		}

		if err != nil {
			logger.Info("SignallingChannel[%s %s %s]| stream recv %+v.", sc.businessName, sc.appName, sc.path, err)
			switchCh <- struct{}{}
			return err
		}

		// process commands from connserver.
		switch resp.Cmd {
		case pb.SignallingChannelCmd_SCCMD_S2C_PONG:
			logger.Info("SignallingChannel[%s %s %s]| CMD -- recviced PONG.", sc.businessName, sc.appName, sc.path)

		case pb.SignallingChannelCmd_SCCMD_S2C_PUSH_NOTIFICATION:
			logger.Info("SignallingChannel[%s %s %s]| CMD -- recviced PUBLISH NOTIFICATION, %+v", sc.businessName, sc.appName, sc.path, resp.CmdPush)
			go sc.handler.Handle(resp.CmdPush)

		case pb.SignallingChannelCmd_SCCMD_S2C_PUSH_ROLLBACK_NOTIFICATION:
			logger.Info("SignallingChannel[%s %s %s]| CMD -- recviced ROLLBACK PUBLISH NOTIFICATION, %+v", sc.businessName, sc.appName, sc.path, resp.CmdRollback)
			go sc.handler.Handle(resp.CmdRollback)

		case pb.SignallingChannelCmd_SCCMD_S2C_PUSH_RELOAD_NOTIFICATION:
			logger.Info("SignallingChannel[%s %s %s]| CMD -- recviced RELOAD PUBLISH NOTIFICATION, %+v", sc.businessName, sc.appName, sc.path, resp.CmdReload)
			go sc.handler.Handle(resp.CmdReload)

		default:
			logger.Error("SignallingChannel[%s %s %s]| unknow signalling channel cmd[%+v]!", sc.businessName, sc.appName, sc.path, resp.Cmd)
		}
	}

	return nil
}

// Setup setups a signalling channel.
func (sc *SignallingChannel) Setup() {
	// don't wait here at first time.
	isFirstTime := true

	ticker := time.NewTicker(sc.viper.GetDuration("sidecar.accessInterval"))
	defer ticker.Stop()

	for {
		if !isFirstTime {
			<-ticker.C
		}
		isFirstTime = false

		// access now.
		nodes, err := sc.access()
		if err != nil {
			logger.Error("SignallingChannel[%s %s %s]| access, %+v", sc.businessName, sc.appName, sc.path, err)
			continue
		}

		// try to setup a signalling channel.
		var client pb.ConnectionClient
		var conn *grpc.ClientConn
		var stream pb.Connection_SignallingChannelClient

		isSetupSuccess := false
		if sc.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(sc.businessName, sc.appName, sc.path))) {
			logger.Info("SignallingChannel[%s %s %s]| signalling stop now!", sc.businessName, sc.appName, sc.path)
			return
		}

		for _, node := range nodes {
			opts := []grpc.DialOption{
				grpc.WithInsecure(),
				grpc.WithTimeout(sc.viper.GetDuration("connserver.dialtimeout")),
			}

			// dial a new connection with the node.
			c, err := grpc.Dial(node, opts...)
			if err != nil {
				logger.Error("SignallingChannel[%s %s %s]| can't dial connserver node[%+v], %+v, try next now.", sc.businessName, sc.appName, sc.path, node, err)
				continue
			}
			conn = c
			client = pb.NewConnectionClient(conn)

			// setup a signalling channel stream now.
			s, err := client.SignallingChannel(context.Background())
			if err != nil {
				logger.Error("SignallingChannel[%s %s %s]| can't setup a signalling channel with node[%+v], %+v", sc.businessName, sc.appName, sc.path, node, err)
				conn.Close()
				continue
			}
			logger.Info("SignallingChannel[%s %s %s]| setup a new signalling channel with node[%+v] success.", sc.businessName, sc.appName, sc.path, node)

			// setup stream success.
			stream = s
			isSetupSuccess = true
			break
		}

		if !isSetupSuccess {
			logger.Error("SignallingChannel[%s %s %s]| can't setup a signalling channel finally, try again later.", sc.businessName, sc.appName, sc.path)
			continue
		}

		// signalling gCoroutines context.
		ctx, cancel := context.WithCancel(context.Background())

		// wait switch signal from two signalling gCoroutines.
		switchCh := make(chan struct{}, 2)

		// keeping signalling channel now.
		go sc.signalling(ctx, switchCh, stream)

		// stream error, switch now.
		<-switchCh
		logger.Info("SignallingChannel[%s %s %s]| cancel signalling gCoroutines and switch stream now.", sc.businessName, sc.appName, sc.path)

		// cancel signalling gCoroutines.
		cancel()
		conn.Close()
	}
}

// Close stops the signalling and handlers.
func (sc *SignallingChannel) Close() {
	sc.viper.Set(fmt.Sprintf("appmod.%s.stop", ModKey(sc.businessName, sc.appName, sc.path)), true)
	logger.Info("SignallingChannel[%s %s %s]| mark signalling stop flag done!", sc.businessName, sc.appName, sc.path)
}
