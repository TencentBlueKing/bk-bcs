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
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	pbsidecar "bk-bscp/internal/protocol/sidecar"
	"bk-bscp/internal/safeviper"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// InstanceServer is instance server.
type InstanceServer struct {
	// viper as context.
	viper *safeviper.SafeViper

	// endpoints.
	httpEndpoint string
	grpcEndpoint string

	// app mod manager.
	appModMgr *AppModManager

	// gRPC gateway server mux.
	gwmux *runtime.ServeMux

	// http server mux.
	mux *http.ServeMux

	// instance server gRPC client.
	insSvrConn *grpc.ClientConn
	insSvrCli  pbsidecar.InstanceClient

	// grpc listener.
	lis net.Listener

	// grpc server.
	server *grpc.Server

	// configs reloader.
	reloader *Reloader

	// app reload event chans.
	events   map[string]chan *ReloadSpec
	eventsMu sync.RWMutex
}

// NewInstanceServer creates a new InstanceServer.
func NewInstanceServer(viper *safeviper.SafeViper, httpEndpoint, grpcEndpoint string,
	appModMgr *AppModManager, reloader *Reloader) *InstanceServer {

	return &InstanceServer{
		viper:        viper,
		httpEndpoint: httpEndpoint,
		grpcEndpoint: grpcEndpoint,
		appModMgr:    appModMgr,
		reloader:     reloader,
		events:       make(map[string]chan *ReloadSpec),
	}
}

// Ping handles ping request, and return sideca mod infos.
func (ins *InstanceServer) Ping(ctx context.Context, req *pbsidecar.PingReq) (*pbsidecar.PingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("INSTANCE-Ping[%s]| input[%+v]", req.Seq, req)
	response := &pbsidecar.PingResp{Seq: req.Seq, Code: pbcommon.ErrCode_E_OK, Message: "OK"}

	defer func() {
		cost := common.ToMSTimestamp(time.Now()) - common.ToMSTimestamp(rtime)
		logger.V(2).Infof("INSTANCE-Ping[%s]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if err := verifyProto(req); err != nil {
		response.Code = pbcommon.ErrCode_E_IS_PARAMS_INVALID
		response.Message = err.Error()
		return response, nil
	}

	mods := []*pbsidecar.ModInfo{}

	for _, mod := range ins.appModMgr.AppModInfos() {
		modKey := ModKey(mod.BizID, mod.AppID, mod.Path)

		m := &pbsidecar.ModInfo{
			BizId:   mod.BizID,
			AppId:   mod.AppID,
			CloudId: mod.CloudID,
			Ip:      ins.viper.GetString("appinfo.ip"),
			Path:    mod.Path,
			Labels:  ins.viper.GetStringMapString(fmt.Sprintf("appmod.%s.labels", modKey)),
		}

		// main flag is true or app level flag is true.
		if ins.viper.GetBool("sidecar.readyPullConfigs") ||
			ins.viper.GetBool(fmt.Sprintf("sidecar.%s.readyPullConfigs", modKey)) {
			m.IsReady = true
		}

		mods = append(mods, m)
	}
	response.Mods = mods

	return response, nil
}

// Inject handle inject request, and update instance labels.
func (ins *InstanceServer) Inject(ctx context.Context, req *pbsidecar.InjectReq) (*pbsidecar.InjectResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("INSTANCE-Inject[%s]| input[%+v]", req.Seq, req)
	response := &pbsidecar.InjectResp{Seq: req.Seq, Code: pbcommon.ErrCode_E_OK, Message: "OK"}

	defer func() {
		cost := common.ToMSTimestamp(time.Now()) - common.ToMSTimestamp(rtime)
		logger.V(2).Infof("INSTANCE-Inject[%s]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if err := verifyProto(req); err != nil {
		response.Code = pbcommon.ErrCode_E_IS_PARAMS_INVALID
		response.Message = err.Error()
		return response, nil
	}
	modKey := ModKey(req.BizId, req.AppId, req.Path)

	signallings := ins.appModMgr.AppSignallings()

	signalling := signallings[modKey]
	signalling.Reset(req.Labels)

	ins.viper.Set(fmt.Sprintf("appmod.%s.labels", modKey), req.Labels)

	// ready to pull configs now. All inner mods init coroutines would watch this flag,
	// and pull configs until it is true.
	ins.viper.Set(fmt.Sprintf("sidecar.%s.readyPullConfigs", modKey), true)

	return response, nil
}

// WatchReload is watch stream for reload events.
func (ins *InstanceServer) WatchReload(req *pbsidecar.WatchReloadReq,
	stream pbsidecar.Instance_WatchReloadServer) error {

	rtime := time.Now()
	logger.V(2).Infof("INSTANCE-WatchReload[%s]| input[%+v]", req.Seq, req)

	// stream context.
	var finalErr error
	ctx := stream.Context()

	defer func() {
		cost := common.ToMSTimestamp(time.Now()) - common.ToMSTimestamp(rtime)
		logger.V(2).Infof("INSTANCE-WatchReload[%s]| output[%dms][%+v]", req.Seq, cost, finalErr)
	}()

	if err := verifyProto(req); err != nil {
		finalErr = err
		stream.Send(&pbsidecar.WatchReloadResp{
			Seq:     req.Seq,
			Code:    pbcommon.ErrCode_E_IS_PARAMS_INVALID,
			Message: err.Error(),
		})
		return nil
	}

	modKey := ModKey(req.BizId, req.AppId, req.Path)

	ins.eventsMu.Lock()
	if _, ok := ins.events[modKey]; !ok {
		ins.events[modKey] = make(chan *ReloadSpec, ins.viper.GetInt("instance.reloadChanSize"))
	}
	ch := ins.events[modKey]
	ins.eventsMu.Unlock()

	// watch on ch.
	for {
		select {
		case <-ctx.Done():
			finalErr = errors.New("client watch reload is closing")
			stream.Send(&pbsidecar.WatchReloadResp{
				Seq:     req.Seq,
				Code:    pbcommon.ErrCode_E_IS_SYSTEM_UNKNOWN,
				Message: finalErr.Error(),
			})
			return nil

		case event := <-ch:
			if event.BizID != req.BizId || event.AppID != req.AppId || event.Path != filepath.Clean(req.Path) {
				logger.Warn("INSTANCE-WatchReload[%s]| recv invalid business/app mod events data, %+v", req.Seq, event)
				continue
			}

			response := &pbsidecar.WatchReloadResp{
				Seq:            req.Seq,
				Code:           pbcommon.ErrCode_E_OK,
				Message:        "OK",
				ReleaseId:      event.ReleaseID,
				MultiReleaseId: event.MultiReleaseID,
				ReleaseName:    event.ReleaseName,
				ReloadType:     event.ReloadType,
				RootPath:       ins.viper.GetString(fmt.Sprintf("appmod.%s.path", modKey)),
			}

			metadatas := []*pbsidecar.ConfigsMetadata{}

			for _, configs := range event.Configs {
				md := &pbsidecar.ConfigsMetadata{
					Name:  configs.Name,
					Fpath: configs.Fpath,
				}
				metadatas = append(metadatas, md)
			}
			response.Metadatas = metadatas

			if err := stream.Send(response); err != nil {
				finalErr = fmt.Errorf("send reload event to suber(business app module), %+v, %+v", event, err)
				stream.Send(&pbsidecar.WatchReloadResp{
					Seq:     req.Seq,
					Code:    pbcommon.ErrCode_E_IS_SYSTEM_UNKNOWN,
					Message: finalErr.Error(),
				})
				return nil
			}
			logger.V(2).Infof("INSTANCE-WatchReload[%s]| send reload event success, [%+v]", req.Seq, response)
		}
	}
}

// ReportReload handle configs reload reports.
func (ins *InstanceServer) ReportReload(ctx context.Context,
	req *pbsidecar.ReportReloadReq) (*pbsidecar.ReportReloadResp, error) {

	rtime := time.Now()
	logger.V(2).Infof("INSTANCE-ReportReload[%s]| input[%+v]", req.Seq, req)
	response := &pbsidecar.ReportReloadResp{Seq: req.Seq, Code: pbcommon.ErrCode_E_OK, Message: "OK"}

	defer func() {
		cost := common.ToMSTimestamp(time.Now()) - common.ToMSTimestamp(rtime)
		logger.V(2).Infof("INSTANCE-ReportReload[%s]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if err := verifyProto(req); err != nil {
		response.Code = pbcommon.ErrCode_E_IS_PARAMS_INVALID
		response.Message = err.Error()
		return response, nil
	}

	// make connserver gRPC client now.
	client, conn, err := ins.makeConnectionClient()
	if err != nil {
		response.Code = pbcommon.ErrCode_E_IS_SYSTEM_UNKNOWN
		response.Message = err.Error()
		return response, nil
	}
	defer conn.Close()

	modKey := ModKey(req.BizId, req.AppId, req.Path)

	// marshal sidecar labels.
	sidecarLabels := &strategy.SidecarLabels{
		Labels: ins.viper.GetStringMapString(fmt.Sprintf("appmod.%s.labels", modKey)),
	}

	labels, err := json.Marshal(sidecarLabels)
	if err != nil {
		response.Code = pbcommon.ErrCode_E_IS_SYSTEM_UNKNOWN
		response.Message = err.Error()
		return response, nil
	}

	r := &pb.ReportReq{
		Seq:     req.Seq,
		BizId:   ins.viper.GetString(fmt.Sprintf("appmod.%s.bizid", modKey)),
		AppId:   ins.viper.GetString(fmt.Sprintf("appmod.%s.appid", modKey)),
		CloudId: ins.viper.GetString(fmt.Sprintf("appmod.%s.cloudid", modKey)),
		Ip:      ins.viper.GetString("appinfo.ip"),
		Path:    ins.viper.GetString(fmt.Sprintf("appmod.%s.path", modKey)),
		Labels:  string(labels),
		Infos: []*pbcommon.ReportInfo{&pbcommon.ReportInfo{
			ReleaseId:      req.ReleaseId,
			MultiReleaseId: req.MultiReleaseId,
			ReloadTime:     req.ReloadTime,
			ReloadCode:     req.ReloadCode,
			ReloadMsg:      req.ReloadMsg,
		}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), ins.viper.GetDuration("connserver.callTimeout"))
	defer cancel()

	logger.V(2).Infof("INSTANCE-ReportReload[%s]| request to connserver Report, %+v", req.Seq, r)

	resp, err := client.Report(ctx, r)
	if err != nil {
		response.Code = pbcommon.ErrCode_E_IS_SYSTEM_UNKNOWN
		response.Message = err.Error()
		return response, nil
	}

	if resp.Code != pbcommon.ErrCode_E_OK {
		response.Code = resp.Code
		response.Message = resp.Message
		return response, nil
	}

	return response, nil
}

// makeConnectionClient returns connserver gRPC connection/client.
func (ins *InstanceServer) makeConnectionClient() (pb.ConnectionClient, *grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(ins.viper.GetDuration("connserver.dialTimeout")),
	}

	endpoint := ins.viper.GetString("connserver.hostName") + ":" + ins.viper.GetString("connserver.port")
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, nil, err
	}
	client := pb.NewConnectionClient(conn)
	return client, conn, nil
}

// Init inits base listener and muxs.
func (ins *InstanceServer) Init() error {
	// listen.
	lis, err := net.Listen("tcp", ins.grpcEndpoint)
	if err != nil {
		return fmt.Errorf("instance server grpc endpoint listen: %+v", err)
	}
	ins.lis = lis

	// gRPC server register.
	ins.server = grpc.NewServer()
	pbsidecar.RegisterInstanceServer(ins.server, ins)

	// init gateway mux.
	opt := runtime.WithMarshalerOption(runtime.MIMEWildcard,
		&runtime.JSONPb{EnumsAsInts: true, EmitDefaults: true, OrigName: true})
	ins.gwmux = runtime.NewServeMux(opt)

	// init gateway http server.
	ins.mux = http.NewServeMux()

	ins.mux.Handle("/", ins.gwmux)
	ins.mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join("swagger", strings.TrimPrefix(r.URL.Path, "/swagger/")))
	})

	return nil
}

func (ins *InstanceServer) handleReloadEvents() {
	events := ins.reloader.EventChan()

	for {
		event := <-events
		logger.Info("recv new instance server reload event from reloader, %+v", event)

		ins.eventsMu.Lock()

		modKey := ModKey(event.BizID, event.AppID, event.Path)

		if _, ok := ins.events[modKey]; !ok {
			ins.events[modKey] = make(chan *ReloadSpec, ins.viper.GetInt("instance.reloadChanSize"))
		}

		ch := ins.events[modKey]

		select {
		case ch <- event:
		case <-time.After(ins.viper.GetDuration("instance.reloadChanTimeout")):
			logger.Warn("add reload spec to instance server events channel timeout, event[%+v]", event)
		}

		ins.eventsMu.Unlock()
	}
}

// Run runs grpc server and gateway.
func (ins *InstanceServer) Run() error {
	// grpc serve.
	go func() {
		if err := ins.server.Serve(ins.lis); err != nil {
			logger.Errorf("instance server grpc serve: %+v", err)
		}
	}()

	// init instance server gRPC client.
	conn, err := grpc.Dial(ins.grpcEndpoint, grpc.WithInsecure(),
		grpc.WithTimeout(ins.viper.GetDuration("instance.dialTimeout")))
	if err != nil {
		logger.Errorf("create instance server grpc client, %+v", err)
		return err
	}
	ins.insSvrConn = conn
	ins.insSvrCli = pbsidecar.NewInstanceClient(conn)

	// register instance server handler.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := pbsidecar.RegisterInstanceHandlerClient(ctx, ins.gwmux, ins.insSvrCli); err != nil {
		logger.Errorf("gateway register instance server handler, %+v", err)
		return err
	}

	go ins.handleReloadEvents()
	logger.Info("instance server run success.")

	// gateway service listen and serve.
	if err := http.ListenAndServe(ins.httpEndpoint, ins.mux); err != nil {
		logger.Errorf("gateway listen and serve, %+v", err)
		return err
	}
	return nil
}
