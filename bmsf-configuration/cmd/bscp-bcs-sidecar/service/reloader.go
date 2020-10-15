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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	pbsidecar "bk-bscp/internal/protocol/sidecar"
	"bk-bscp/internal/strategy"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

type ReloadType int32

const (
	// ReloadTypeUpdate is update reload type.
	ReloadTypeUpdate ReloadType = 0

	// ReloadTypeRollback is rollback reload type.
	ReloadTypeRollback ReloadType = 1

	// ReloadTypeFirstReload is first reload type when new sidecar instance setup.
	ReloadTypeFirstReload ReloadType = 2
)

// ConfigSpec is config spec.
type ConfigSpec struct {
	// config set name
	Name string

	// config set fpath.
	Fpath string
}

// ReloadSpec specs how to reload.
type ReloadSpec struct {
	// business name.
	BusinessName string

	// app name.
	AppName string

	// app configs root path.
	Path string

	// release id.
	Releaseid string

	// multi release id .
	MultiReleaseid string

	// release name.
	ReleaseName string

	// reload type.
	ReloadType int32

	// config specs.
	Configs []ConfigSpec
}

// Reloader is configs reloader.
type Reloader struct {
	viper  *viper.Viper
	events chan *ReloadSpec
}

// NewReloader creates a new Reloader.
func NewReloader(viper *viper.Viper) *Reloader {
	return &Reloader{viper: viper, events: make(chan *ReloadSpec, viper.GetInt("instance.reloadChanSize"))}
}

// Init inits new Reloader.
func (r *Reloader) Init() {
	if r.viper.GetBool("sidecar.fileReloadMode") {
		// file reload handler.
		go r.handleFileReload()
	}
}

// Reload handle configs reload.
func (r *Reloader) Reload(spec *ReloadSpec) {
	if spec != nil {
		go r.reload(spec)
	}
}

func (r *Reloader) reload(spec *ReloadSpec) {
	select {
	case r.events <- spec:
	case <-time.After(r.viper.GetDuration("instance.reloadChanTimeout")):
		logger.Warn("send reload spec to reload events channel timeout, spec[%+v]", spec)
	}
}

// EventChan is reload events channel.
func (r *Reloader) EventChan() chan *ReloadSpec {
	return r.events
}

// handleFileReload handles file reload in file reload mode.
// can't use filereload and instance server in the same time.
func (r *Reloader) handleFileReload() chan *ReloadSpec {
	for {
		event := <-r.events
		logger.Info("[%s][%s][%s]| recv new file reload event from reloader, %+v", event.BusinessName, event.AppName, event.Path, event)

		// touch file to notify reload.
		fReloadFName := fmt.Sprintf("%s/%s", r.viper.GetString(fmt.Sprintf("appmod.%s.path", ModKey(event.BusinessName, event.AppName, event.Path))),
			r.viper.GetString("sidecar.fileReloadFName"))

		fReload, err := os.OpenFile(fReloadFName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			logger.Errorf("[%s][%s][%s]| filereload, touch reload file failed, %+v", event.BusinessName, event.AppName, event.Path, err)
			continue
		}

		// write reload content.
		// unixts + space + reloadtype, eg: 11212423423 1.
		reloadSpecContent := fmt.Sprintf("%d %d", time.Now().Unix(), event.ReloadType)
		for _, configs := range event.Configs {
			reloadSpecContent += fmt.Sprintf("\n%s/%s", configs.Fpath, configs.Name)
		}

		if _, err := fReload.WriteString(reloadSpecContent); err != nil {
			logger.Errorf("[%s][%s][%s]| filereload, write reload file content failed, %+v", event.BusinessName, event.AppName, event.Path, err)
		} else {
			logger.Infof("[%s][%s][%s]| filereload, notify reload success!", event.BusinessName, event.AppName, event.Path)
		}

		// close fd.
		fReload.Close()

		if event.ReloadType != int32(ReloadTypeFirstReload) {

			// reload success, and report reload result now.
			reportReloadReq := &pbsidecar.ReportReloadReq{
				Seq:            common.Sequence(),
				BusinessName:   event.BusinessName,
				AppName:        event.AppName,
				Releaseid:      event.Releaseid,
				MultiReleaseid: event.MultiReleaseid,
				ReloadTime:     time.Now().Format("2006-01-02 15:04:05"),
				ReloadCode:     1,
				ReloadMsg:      "SUCCESS",
			}

			if event.ReloadType == int32(ReloadTypeRollback) {
				// rollback reload.
				reportReloadReq.ReloadCode = 2
				reportReloadReq.ReloadMsg = "ROLLBACK SUCCESS"
			}

			if err := r.reportReload(reportReloadReq, event.Path); err != nil {
				logger.Infof("[%s][%s][%s]| filereload, report reload failed, %+v", event.BusinessName, event.AppName, event.Path, err)
			}
		}
	}
}

// makeConnectionClient returns connserver gRPC connection/client.
func (r *Reloader) makeConnectionClient() (pb.ConnectionClient, *grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(r.viper.GetDuration("connserver.dialtimeout")),
	}

	endpoint := r.viper.GetString("connserver.hostname") + ":" + r.viper.GetString("connserver.port")
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, nil, err
	}
	client := pb.NewConnectionClient(conn)
	return client, conn, nil
}

func (r *Reloader) reportReload(req *pbsidecar.ReportReloadReq, path string) error {
	// make connserver gRPC client now.
	client, conn, err := r.makeConnectionClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	modKey := ModKey(req.BusinessName, req.AppName, path)

	// marshal sidecar labels.
	sidecarLabels := &strategy.SidecarLabels{
		Labels: r.viper.GetStringMapString(fmt.Sprintf("appmod.%s.labels", modKey)),
	}

	labels, err := json.Marshal(sidecarLabels)
	if err != nil {
		return err
	}

	rr := &pb.ReportReq{
		Seq:       req.Seq,
		Bid:       r.viper.GetString(fmt.Sprintf("appmod.%s.bid", modKey)),
		Appid:     r.viper.GetString(fmt.Sprintf("appmod.%s.appid", modKey)),
		Clusterid: r.viper.GetString(fmt.Sprintf("appmod.%s.clusterid", modKey)),
		Zoneid:    r.viper.GetString(fmt.Sprintf("appmod.%s.zoneid", modKey)),
		Dc:        r.viper.GetString(fmt.Sprintf("appmod.%s.dc", modKey)),
		IP:        r.viper.GetString("appinfo.ip"),
		Labels:    string(labels),
		Infos: []*pbcommon.ReportInfo{&pbcommon.ReportInfo{
			Releaseid:      req.Releaseid,
			MultiReleaseid: req.MultiReleaseid,
			ReloadTime:     req.ReloadTime,
			ReloadCode:     req.ReloadCode,
			ReloadMsg:      req.ReloadMsg,
		}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.viper.GetDuration("connserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("[%s][%s][%s][%d]| filereload, request to connserver Report, %+v", req.BusinessName, req.AppName, path, req.Seq, rr)

	resp, err := client.Report(ctx, rr)
	if err != nil {
		return err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return errors.New(resp.ErrMsg)
	}
	logger.Infof("[%s][%s][%s][%d]| filereload, report reload success, %+v", req.BusinessName, req.AppName, path, req.Seq, rr.Infos)

	return nil
}
