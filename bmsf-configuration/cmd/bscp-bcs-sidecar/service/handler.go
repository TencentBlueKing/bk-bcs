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
	"sync"
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

// Handler handles all commands from connserver.
type Handler struct {
	// configs handler.
	viper *viper.Viper

	businessName string
	appName      string

	// event channel.
	ch chan interface{}

	// config handler.
	configHandler *ConfigHandler
}

// NewHandler creates new Handler.
func NewHandler(viper *viper.Viper, businessName, appName string, configHandler *ConfigHandler) *Handler {
	return &Handler{
		viper:         viper,
		businessName:  businessName,
		appName:       appName,
		configHandler: configHandler,
		ch:            make(chan interface{}, viper.GetInt("sidecar.handlerChSize")),
	}
}

// handlePub handles publish notifications.
func (h *Handler) handlePub(notification *pb.SCCMDPushNotification) error {
	if notification == nil {
		return errors.New("invalid publish notification struct: nil")
	}

	if notification.Bid != h.viper.GetString(fmt.Sprintf("appmod.%s_%s.bid", h.businessName, h.appName)) ||
		notification.Appid != h.viper.GetString(fmt.Sprintf("appmod.%s_%s.appid", h.businessName, h.appName)) {
		return fmt.Errorf("invalid publish notification organization: bid/appid")
	}

	// handle config publishing.
	if err := h.configHandler.Handle(notification); err != nil {
		return err
	}
	return nil
}

// handleRoll handles rollback publish notifications.
func (h *Handler) handleRoll(notification *pb.SCCMDPushRollbackNotification) error {
	if notification == nil {
		return errors.New("invalid rollback notification struct: nil")
	}

	if notification.Bid != h.viper.GetString(fmt.Sprintf("appmod.%s_%s.bid", h.businessName, h.appName)) ||
		notification.Appid != h.viper.GetString(fmt.Sprintf("appmod.%s_%s.appid", h.businessName, h.appName)) {
		return fmt.Errorf("invalid rollback notification organization: bid/appid")
	}

	// handle config rollback publishing.
	if err := h.configHandler.Handle(notification); err != nil {
		return err
	}
	return nil
}

// handleReload handles reload publish notifications.
func (h *Handler) handleReload(notification *pb.SCCMDPushReloadNotification) error {
	if notification == nil {
		return errors.New("invalid reload notification struct: nil")
	}

	if notification.Bid != h.viper.GetString(fmt.Sprintf("appmod.%s_%s.bid", h.businessName, h.appName)) ||
		notification.Appid != h.viper.GetString(fmt.Sprintf("appmod.%s_%s.appid", h.businessName, h.appName)) {
		return fmt.Errorf("invalid reload notification organization: bid/appid")
	}

	// handle release reload publishing.
	if err := h.configHandler.Handle(notification); err != nil {
		return err
	}
	return nil
}

// signalling keeps processing signalling from connserver.
func (h *Handler) signalling() {
	for {
		if h.viper.GetBool(fmt.Sprintf("appmod.%s_%s.stop", h.businessName, h.appName)) {
			logger.Info("handler[%s %s]| stop signalling now!", h.businessName, h.appName)
			return
		}

		cmd := <-h.ch

		switch cmd.(type) {
		case *pb.SCCMDPushNotification:
			notification := cmd.(*pb.SCCMDPushNotification)
			if err := h.handlePub(notification); err != nil {
				logger.Error("handler[%s %s]| handle publish notification, %+v", h.businessName, h.appName, err)
			}

		case *pb.SCCMDPushRollbackNotification:
			notification := cmd.(*pb.SCCMDPushRollbackNotification)
			if err := h.handleRoll(notification); err != nil {
				logger.Error("handler[%s %s]| handle rollback publish notification, %+v", h.businessName, h.appName, err)
			}

		case *pb.SCCMDPushReloadNotification:
			notification := cmd.(*pb.SCCMDPushReloadNotification)
			if err := h.handleReload(notification); err != nil {
				logger.Error("handler[%s %s]| handle reload publish notification, %+v", h.businessName, h.appName, err)
			}

		default:
			logger.Error("handler[%s %s]| unknow command[%+v]", h.businessName, h.appName, cmd)
		}
	}
}

// Handle handles the commands from connserver.
func (h *Handler) Handle(cmd interface{}) {
	select {
	case h.ch <- cmd:
	case <-time.After(h.viper.GetDuration("sidecar.handlerChTimeout")):
		logger.Error("handler[%s %s]| send cmd to handler channel timeout, %+v", h.businessName, h.appName, cmd)
	}
}

// Run runs the handlers.
func (h *Handler) Run() {
	// run config handler.
	h.configHandler.Run()

	// keep processing signalling.
	go h.signalling()
}

// ConfigHandler is config publishing handler.
type ConfigHandler struct {
	// viper as context here.
	viper *viper.Viper

	businessName string
	appName      string

	// config release effect cache.
	effectCache *EffectCache

	// config content cache.
	contentCache *ContentCache

	// publish event channel.
	ch chan interface{}

	// config set pullers(cfgsetid -> puller).
	pullers map[string]*Puller

	// mu for config set pullers.
	mu sync.RWMutex

	// configs reloader.
	reloader *Reloader
}

// NewConfigHandler creates a new config handler.
func NewConfigHandler(viper *viper.Viper, businessName, appName string, effectCache *EffectCache,
	contentCache *ContentCache, reloader *Reloader) *ConfigHandler {
	return &ConfigHandler{
		viper:        viper,
		businessName: businessName,
		appName:      appName,
		effectCache:  effectCache,
		contentCache: contentCache,
		reloader:     reloader,
		pullers:      make(map[string]*Puller),
		ch:           make(chan interface{}, viper.GetInt("sidecar.configHandlerChSize")),
	}
}

// makeConnectionClient returns connserver gRPC connection/client.
func (h *ConfigHandler) makeConnectionClient() (pb.ConnectionClient, *grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(h.viper.GetDuration("connserver.dialtimeout")),
	}

	endpoint := h.viper.GetString("connserver.hostname") + ":" + h.viper.GetString("connserver.port")
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, nil, err
	}
	client := pb.NewConnectionClient(conn)
	return client, conn, nil
}

// sidecarLabels marshals sidecar labels to string base on strategy protocol.
func (h *ConfigHandler) sidecarLabels() (string, error) {
	sidecarLabels := &strategy.SidecarLabels{Labels: h.viper.GetStringMapString(fmt.Sprintf("appmod.%s_%s.labels", h.businessName, h.appName))}

	labels, err := json.Marshal(sidecarLabels)
	if err != nil {
		return "", err
	}
	return string(labels), nil
}

// report reports the effected release information of all configsets.
func (h *ConfigHandler) report(cfgsetids []string) error {
	if cfgsetids == nil {
		return errors.New("invalid cfgsetids: nil")
	}

	reportInfos := []*pbcommon.ReportInfo{}

	for _, cfgsetid := range cfgsetids {
		md, err := h.effectCache.LocalRelease(cfgsetid)
		if err != nil {
			continue
		}
		if md != nil && md.Releaseid != "" && md.EffectTime != "" {
			reportInfo := &pbcommon.ReportInfo{
				Cfgsetid:   cfgsetid,
				Releaseid:  md.Releaseid,
				EffectTime: md.EffectTime,
				EffectCode: 0,
				EffectMsg:  "SUCCESS",
			}
			reportInfos = append(reportInfos, reportInfo)
		}
	}

	if len(reportInfos) == 0 {
		return nil
	}

	// make connserver gRPC client now.
	client, conn, err := h.makeConnectionClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	// marshal sidecar labels.
	labels, err := h.sidecarLabels()
	if err != nil {
		return err
	}

	r := &pb.ReportReq{
		Seq:       common.Sequence(),
		Bid:       h.viper.GetString(fmt.Sprintf("appmod.%s_%s.bid", h.businessName, h.appName)),
		Appid:     h.viper.GetString(fmt.Sprintf("appmod.%s_%s.appid", h.businessName, h.appName)),
		Clusterid: h.viper.GetString(fmt.Sprintf("appmod.%s_%s.clusterid", h.businessName, h.appName)),
		Zoneid:    h.viper.GetString(fmt.Sprintf("appmod.%s_%s.zoneid", h.businessName, h.appName)),
		Dc:        h.viper.GetString(fmt.Sprintf("appmod.%s_%s.dc", h.businessName, h.appName)),
		IP:        h.viper.GetString("appinfo.ip"),
		Labels:    labels,
		Infos:     reportInfos,
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.viper.GetDuration("connserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("ConfigHandler[%s %s]| request to connserver Report, %+v", h.businessName, h.appName, r)

	resp, err := client.Report(ctx, r)
	if err != nil {
		return err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return errors.New(resp.ErrMsg)
	}
	return nil
}

// pullConfigSetList pulls configset list from connserver.
func (h *ConfigHandler) pullConfigSetList() ([]string, error) {
	// make connserver gRPC client now.
	client, conn, err := h.makeConnectionClient()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// query in page mode.
	cfgsetids := []string{}
	index := 0
	limit := h.viper.GetInt("sidecar.configSetListSize")

	for {
		r := &pb.PullConfigSetListReq{
			Seq:   common.Sequence(),
			Bid:   h.viper.GetString(fmt.Sprintf("appmod.%s_%s.bid", h.businessName, h.appName)),
			Appid: h.viper.GetString(fmt.Sprintf("appmod.%s_%s.appid", h.businessName, h.appName)),
			Index: int32(index),
			Limit: int32(limit),
		}

		ctx, cancel := context.WithTimeout(context.Background(), h.viper.GetDuration("connserver.calltimeout"))
		defer cancel()

		logger.V(2).Infof("ConfigHandler[%s %s]| request to connserver PullConfigSetList, %+v", h.businessName, h.appName, r)

		// pull config set list from connserver.
		resp, err := client.PullConfigSetList(ctx, r)
		if err != nil {
			return nil, err
		}
		if resp.ErrCode != pbcommon.ErrCode_E_OK {
			return nil, errors.New(resp.ErrMsg)
		}
		if len(resp.ConfigSets) == 0 {
			break
		}

		for _, cfgset := range resp.ConfigSets {
			cfgsetids = append(cfgsetids, cfgset.Cfgsetid)
		}

		if len(resp.ConfigSets) < limit {
			break
		}
		index += len(resp.ConfigSets)
	}

	return cfgsetids, nil
}

func (h *ConfigHandler) getPuller(cfgsetid string) *Puller {
	h.mu.Lock()
	defer h.mu.Unlock()

	if v, ok := h.pullers[cfgsetid]; !ok || v == nil {
		newPuller := NewPuller(h.viper, h.businessName, h.appName, cfgsetid, h.effectCache, h.contentCache)

		h.pullers[cfgsetid] = newPuller
		newPuller.Run()
	}
	puller := h.pullers[cfgsetid]
	return puller
}

// pulling keeps pulling release.
func (h *ConfigHandler) pulling() {
	for {
		if h.viper.GetBool(fmt.Sprintf("appmod.%s_%s.stop", h.businessName, h.appName)) {
			logger.Info("ConfigHandler[%s %s]| stop pulling now!", h.businessName, h.appName)
			return
		}

		notification := <-h.ch

		switch notification.(type) {
		case *pb.SCCMDPushNotification:
			msg := notification.(*pb.SCCMDPushNotification)
			puller := h.getPuller(msg.Cfgsetid)

			// send publishing notification to target puller.
			if err := puller.HandlePub(msg); err != nil {
				logger.Error("ConfigHandler[%s %s] | pulling, handle publish notification to puller, %+v", h.businessName, h.appName, err)
			}

		case *pb.SCCMDPushRollbackNotification:
			msg := notification.(*pb.SCCMDPushRollbackNotification)
			puller := h.getPuller(msg.Cfgsetid)

			// send publishing notification to target puller.
			if err := puller.HandleRoll(msg); err != nil {
				logger.Error("ConfigHandler[%s %s] | pulling, handle rollback notification to puller, %+v", h.businessName, h.appName, err)
			}

		case *pb.SCCMDPushReloadNotification:
			msg := notification.(*pb.SCCMDPushReloadNotification)

			// send publishing notification to target puller.
			if err := h.handleReload(msg); err != nil {
				logger.Error("ConfigHandler[%s %s] | pulling, handle reload notification to puller, %+v", h.businessName, h.appName, err)
			}

		default:
			logger.Error("ConfigHandler[%s %s]| unknow command[%+v]", h.businessName, h.appName, notification)
		}
	}
}

// handleReload handles reload event, you may not know why it's here,
// but here is the only interface to handle all puller of configsets.
func (h *ConfigHandler) handleReload(msg *pb.SCCMDPushReloadNotification) error {
	if !h.viper.GetBool("instance.open") {
		// instance service is not open.
		logger.Warnf("ConfigHandler[%s %s]| instance service is not open, can't do reload action", h.businessName, h.appName)
		return nil
	}

	// check base message.
	if msg == nil {
		return errors.New("invalid struct: nil")
	}
	if msg.ReloadSpec == nil {
		return errors.New("invalid struct ReloadSpec: nil")
	}
	if len(msg.ReloadSpec.Info) == 0 {
		return errors.New("empty reload spec.")
	}

	// handle all reload spec info.
	var referenceMetadata *ReleaseMetadata
	metadatas := []*ReleaseMetadata{}

	for _, eInfo := range msg.ReloadSpec.Info {
		md, err := h.effectCache.LocalRelease(eInfo.Cfgsetid)
		if err != nil || md == nil {
			// suppose no effected release.
			return fmt.Errorf("can't reload this release now, configset[%s] suppose not effectting release[%s] this moment", eInfo.Cfgsetid, eInfo.Releaseid)
		}

		if md.Releaseid != eInfo.Releaseid {
			// not effectting target release this moment.
			return fmt.Errorf("can't reload this release now, configset[%s] not effectting release[%s] this moment", eInfo.Cfgsetid, eInfo.Releaseid)
		}

		// mark reference metadata to judge event type.
		referenceMetadata = md
		metadatas = append(metadatas, md)

		// NOTE: may other release is coming, but there should be a lock in user level.
		// Reload is just check local releases and send notification to business, you should
		// know all actions from your operators.
	}

	// all configsets are effectting target release, reload now.
	spec := &ReloadSpec{
		BusinessName: h.businessName,
		AppName:      h.appName,

		// all releases have the same name even under multi release.
		ReleaseName: referenceMetadata.ReleaseName,
	}

	// config reload specs.
	configSpec := []ConfigSpec{}
	for _, md := range metadatas {
		configSpec = append(configSpec, ConfigSpec{Name: md.CfgsetName, Fpath: md.CfgsetFpath})
	}
	spec.Configs = configSpec

	// reload mode.
	if len(msg.ReloadSpec.MultiReleaseid) != 0 {
		// multi release reload mode.
		spec.MultiReleaseid = msg.ReloadSpec.MultiReleaseid
	} else {
		// single release reload mode.
		spec.Releaseid = referenceMetadata.Releaseid
	}

	// reload type, 0: update  1: rollback.
	if referenceMetadata.isRollback {
		spec.ReloadType = 1
	}

	// sync reload event.
	h.reloader.Reload(spec)

	return nil
}

// reporting keeps reporting local release effected information of
// all configsets to connserver.
func (h *ConfigHandler) reporting() {
	ticker := time.NewTicker(h.viper.GetDuration("sidecar.reportInfoInterval"))
	defer ticker.Stop()

	for {
		if h.viper.GetBool(fmt.Sprintf("appmod.%s_%s.stop", h.businessName, h.appName)) {
			logger.Info("ConfigHandler[%s %s]| stop reporting now!", h.businessName, h.appName)
			return
		}

		<-ticker.C

		h.mu.RLock()
		cfgsetids := []string{}
		for cfgsetid := range h.pullers {
			cfgsetids = append(cfgsetids, cfgsetid)
		}
		h.mu.RUnlock()

		if err := h.report(cfgsetids); err != nil {
			logger.Error("ConfigHandler[%s %s]| reporting, report local releases effected information, %+v", h.businessName, h.appName, err)
		}
		logger.Warn("ConfigHandler[%s %s]| reporting, report local releases effected information succcess, %+v", h.businessName, h.appName, cfgsetids)
	}
}

// syncConfigSetList keeps syncing configset list from connserver.
func (h *ConfigHandler) syncConfigSetList() {
	// don't wait here at first time.
	isFirstTime := true

	ticker := time.NewTicker(h.viper.GetDuration("sidecar.syncConfigsetListInterval"))
	defer ticker.Stop()

	for {
		if h.viper.GetBool(fmt.Sprintf("appmod.%s_%s.stop", h.businessName, h.appName)) {
			logger.Info("ConfigHandler[%s %s]| stop syncing configset list now!", h.businessName, h.appName)
			return
		}

		if !isFirstTime {
			<-ticker.C
		}
		isFirstTime = false

		// pull configset list from connserver.
		cfgsetids, err := h.pullConfigSetList()
		if err != nil {
			logger.Error("ConfigHandler[%s %s]| syncConfigSetList, sync configset list, %+v", h.businessName, h.appName, err)
			continue
		}
		if len(cfgsetids) == 0 {
			continue
		}

		// update local pullers.
		h.mu.Lock()
		newCfgsetids := make(map[string]string)
		for _, cfgsetid := range cfgsetids {
			if v, ok := h.pullers[cfgsetid]; !ok || v == nil {
				newPuller := NewPuller(h.viper, h.businessName, h.appName, cfgsetid, h.effectCache, h.contentCache)
				h.pullers[cfgsetid] = newPuller
				newPuller.Run()
			}
			newCfgsetids[cfgsetid] = cfgsetid
		}

		for cfgsetid, puller := range h.pullers {
			if _, ok := newCfgsetids[cfgsetid]; !ok {
				puller.Stop()
			}
		}
		h.mu.Unlock()
		logger.Warn("ConfigHandler[%s %s]| syncConfigSetList, sync configset list success, %+v", h.businessName, h.appName, cfgsetids)
	}
}

// Handle handles publishing notification.
func (h *ConfigHandler) Handle(notification interface{}) error {
	if notification == nil {
		return errors.New("invalid notification struct: nil")
	}

	select {
	case h.ch <- notification:
	case <-time.After(h.viper.GetDuration("sidecar.configHandlerChTimeout")):
		return fmt.Errorf("send cmd to config handler main channel timeout, %+v", notification)
	}

	return nil
}

// Debug prints the debug information.
func (h *ConfigHandler) Debug() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		if h.viper.GetBool(fmt.Sprintf("appmod.%s_%s.stop", h.businessName, h.appName)) {
			logger.Info("ConfigHandler[%s %s]| stop debuging now!", h.businessName, h.appName)
			return
		}

		<-ticker.C

		h.mu.RLock()
		cfgsetids := []string{}
		for cfgsetid := range h.pullers {
			cfgsetids = append(cfgsetids, cfgsetid)
		}
		h.mu.RUnlock()

		for _, cfgsetid := range cfgsetids {
			logger.V(3).Infof("ConfigHandler[%s %s]| debug, %s", h.businessName, h.appName, h.effectCache.Debug(cfgsetid))
		}
	}
}

// Run runs the config handler.
func (h *ConfigHandler) Run() {
	// keep pulling release.
	go h.pulling()

	// keep reporting effected information.
	go h.reporting()

	// keep syncing config set list.
	go h.syncConfigSetList()

	// debug.
	go h.Debug()
}
