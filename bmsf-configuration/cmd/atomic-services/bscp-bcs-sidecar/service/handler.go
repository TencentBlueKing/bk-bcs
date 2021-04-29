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
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/internal/safeviper"
	"bk-bscp/internal/strategy"
	"bk-bscp/internal/types"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// Handler handles all commands from connserver.
type Handler struct {
	viper *safeviper.SafeViper

	bizID string
	appID string
	path  string

	// event channel.
	ch chan interface{}

	// config handler.
	configHandler *ConfigHandler
}

// NewHandler creates new Handler.
func NewHandler(viper *safeviper.SafeViper, bizID, appID, path string, configHandler *ConfigHandler) *Handler {
	return &Handler{
		viper:         viper,
		bizID:         bizID,
		appID:         appID,
		path:          filepath.Clean(path),
		configHandler: configHandler,
		ch:            make(chan interface{}, viper.GetInt("sidecar.handlerChSize")),
	}
}

// handlePub handles publish notifications.
func (h *Handler) handlePub(notification *pb.SCCMDPushNotification) error {
	if notification == nil {
		return errors.New("invalid publish notification struct: nil")
	}
	modKey := ModKey(h.bizID, h.appID, h.path)

	if notification.BizId != h.viper.GetString(fmt.Sprintf("appmod.%s.bizid", modKey)) ||
		notification.AppId != h.viper.GetString(fmt.Sprintf("appmod.%s.appid", modKey)) {
		return fmt.Errorf("invalid publish notification organization: bizid/appid")
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
	modKey := ModKey(h.bizID, h.appID, h.path)

	if notification.BizId != h.viper.GetString(fmt.Sprintf("appmod.%s.bizid", modKey)) ||
		notification.AppId != h.viper.GetString(fmt.Sprintf("appmod.%s.appid", modKey)) {
		return fmt.Errorf("invalid rollback notification organization: bizid/appid")
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
	modKey := ModKey(h.bizID, h.appID, h.path)

	if notification.BizId != h.viper.GetString(fmt.Sprintf("appmod.%s.bizid", modKey)) ||
		notification.AppId != h.viper.GetString(fmt.Sprintf("appmod.%s.appid", modKey)) {
		return fmt.Errorf("invalid reload notification organization: bizid/appid")
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
		if h.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(h.bizID, h.appID, h.path))) {
			logger.Info("handler[%s %s %s]| stop signalling now!", h.bizID, h.appID, h.path)
			return
		}

		var cmd interface{}

		select {
		case cmd = <-h.ch:
		case <-time.After(time.Second):
			continue
		}

		switch cmd.(type) {
		case *pb.SCCMDPushNotification:
			notification := cmd.(*pb.SCCMDPushNotification)
			if err := h.handlePub(notification); err != nil {
				logger.Error("handler[%s %s %s]| handle publish notification, %+v", h.bizID, h.appID, h.path, err)
			}

		case *pb.SCCMDPushRollbackNotification:
			notification := cmd.(*pb.SCCMDPushRollbackNotification)
			if err := h.handleRoll(notification); err != nil {
				logger.Error("handler[%s %s %s]| handle rollback notification, %+v", h.bizID, h.appID, h.path, err)
			}

		case *pb.SCCMDPushReloadNotification:
			notification := cmd.(*pb.SCCMDPushReloadNotification)
			if err := h.handleReload(notification); err != nil {
				logger.Error("handler[%s %s %s]| handle reload notification, %+v", h.bizID, h.appID, h.path, err)
			}

		default:
			logger.Error("handler[%s %s %s]| unknow command[%+v]", h.bizID, h.appID, h.path, cmd)
		}
	}
}

// Reset resets the app runtime data for new instance.
func (h *Handler) Reset() {
	if h != nil {
		h.configHandler.Reset()
	}
}

// Handle handles the commands from connserver.
func (h *Handler) Handle(cmd interface{}) {
	select {
	case h.ch <- cmd:
	case <-time.After(h.viper.GetDuration("sidecar.handlerChTimeout")):
		logger.Error("handler[%s %s %s]| send cmd to handler channel timeout, %+v", h.bizID, h.appID, h.path, cmd)
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
	viper *safeviper.SafeViper

	bizID string
	appID string
	path  string

	connSvrConn *grpc.ClientConn
	connSvrCli  pb.ConnectionClient

	// config release effect cache.
	effectCache *EffectCache

	// config content cache.
	contentCache *ContentCache

	// publish event channel.
	ch chan interface{}

	// config pullers(cfgid -> puller).
	pullers map[string]*Puller

	// mu for config pullers.
	mu sync.RWMutex

	// configs reloader.
	reloader *Reloader

	// if first reload handled success.
	isFirstReloadSucc bool
}

// NewConfigHandler creates a new config handler.
func NewConfigHandler(viper *safeviper.SafeViper, bizID, appID, path string, effectCache *EffectCache,
	contentCache *ContentCache, reloader *Reloader) *ConfigHandler {
	return &ConfigHandler{
		viper:        viper,
		bizID:        bizID,
		appID:        appID,
		path:         filepath.Clean(path),
		effectCache:  effectCache,
		contentCache: contentCache,
		reloader:     reloader,
		pullers:      make(map[string]*Puller),
		ch:           make(chan interface{}, viper.GetInt("sidecar.configHandlerChSize")),
	}
}

// makeConnectionClient returns connserver gRPC connection/client.
func (h *ConfigHandler) makeConnectionClient() error {
	if h.connSvrConn != nil {
		h.connSvrConn.Close()
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(h.viper.GetDuration("connserver.dialTimeout")),
	}

	endpoint := h.viper.GetString("connserver.hostName") + ":" + h.viper.GetString("connserver.port")
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	h.connSvrConn = conn
	h.connSvrCli = pb.NewConnectionClient(conn)
	return nil
}

// sidecarLabels marshals sidecar labels to string base on strategy protocol.
func (h *ConfigHandler) sidecarLabels() (string, error) {
	sidecarLabels := &strategy.SidecarLabels{
		Labels: h.viper.GetStringMapString(fmt.Sprintf("appmod.%s.labels", ModKey(h.bizID, h.appID, h.path))),
	}

	labels, err := json.Marshal(sidecarLabels)
	if err != nil {
		return "", err
	}
	return string(labels), nil
}

// report reports the effected release information of all configs.
func (h *ConfigHandler) report(cfgIDs []string) error {
	// report effect result in batch mode.
	reportInfos := []*pbcommon.ReportInfo{}

	// marshal sidecar labels.
	labels, err := h.sidecarLabels()
	if err != nil {
		return err
	}
	modKey := ModKey(h.bizID, h.appID, h.path)

	for idx, cfgID := range cfgIDs {
		md, _ := h.effectCache.LocalRelease(cfgID)

		if md != nil && len(md.ReleaseID) != 0 && len(md.EffectTime) != 0 {
			reportInfos = append(reportInfos, &pbcommon.ReportInfo{
				CfgId:      cfgID,
				ReleaseId:  md.ReleaseID,
				EffectTime: md.EffectTime,
				EffectCode: types.EffectCodeSuccess,
				EffectMsg:  types.EffectMsgSuccess,
			})
		}

		if len(reportInfos) >= h.viper.GetInt("sidecar.reportInfoLimit") || idx == (len(cfgIDs)-1) {
			r := &pb.ReportReq{
				Seq:     common.Sequence(),
				BizId:   h.viper.GetString(fmt.Sprintf("appmod.%s.bizid", modKey)),
				AppId:   h.viper.GetString(fmt.Sprintf("appmod.%s.appid", modKey)),
				CloudId: h.viper.GetString(fmt.Sprintf("appmod.%s.cloudid", modKey)),
				Ip:      h.viper.GetString("appinfo.ip"),
				Path:    h.viper.GetString(fmt.Sprintf("appmod.%s.path", modKey)),
				Labels:  labels,
				Infos:   reportInfos,
			}

			ctx, cancel := context.WithTimeout(context.Background(), h.viper.GetDuration("connserver.callTimeout"))
			defer cancel()

			logger.V(4).Infof("ConfigHandler[%s %s %s]| request to connserver Report, %+v", h.bizID, h.appID, h.path, r)

			h.connSvrCli.Report(ctx, r)

			reportInfos = []*pbcommon.ReportInfo{}
		}
	}

	return nil
}

// pullConfigList pulls config list from connserver.
func (h *ConfigHandler) pullConfigList() ([]string, error) {
	cfgIDs := []string{}

	index := 0
	limit := h.viper.GetInt("sidecar.configListPageSize")

	modKey := ModKey(h.bizID, h.appID, h.path)

	for {
		r := &pb.PullConfigListReq{
			Seq:   common.Sequence(),
			BizId: h.viper.GetString(fmt.Sprintf("appmod.%s.bizid", modKey)),
			AppId: h.viper.GetString(fmt.Sprintf("appmod.%s.appid", modKey)),
			Page:  &pbcommon.Page{Start: int32(index), Limit: int32(limit)},
		}

		ctx, cancel := context.WithTimeout(context.Background(), h.viper.GetDuration("connserver.callTimeout"))
		defer cancel()

		logger.V(4).Infof("ConfigHandler[%s %s %s]| request to connserver, %+v", h.bizID, h.appID, h.path, r)

		// pull config list from connserver.
		resp, err := h.connSvrCli.PullConfigList(ctx, r)
		if err != nil {
			return nil, err
		}
		if resp.Code != pbcommon.ErrCode_E_OK {
			return nil, errors.New(resp.Message)
		}

		for _, cfg := range resp.Configs {
			cfgIDs = append(cfgIDs, cfg.CfgId)
		}

		if len(resp.Configs) < limit {
			break
		}
		index += len(resp.Configs)
	}

	return cfgIDs, nil
}

func (h *ConfigHandler) getPuller(cfgID string) *Puller {
	h.mu.Lock()
	defer h.mu.Unlock()

	if v, ok := h.pullers[cfgID]; !ok || v == nil {
		newPuller := NewPuller(h.viper, h.bizID, h.appID, h.path, cfgID, h.connSvrCli,
			h.effectCache, h.contentCache)

		h.pullers[cfgID] = newPuller
		newPuller.Run()
	}

	puller := h.pullers[cfgID]
	return puller
}

// pulling keeps pulling release.
func (h *ConfigHandler) pulling() {
	for {
		if h.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(h.bizID, h.appID, h.path))) {
			logger.Info("ConfigHandler[%s %s %s]| stop pulling now!", h.bizID, h.appID, h.path)
			return
		}

		var notification interface{}

		select {
		case notification = <-h.ch:
		case <-time.After(time.Second):
			continue
		}

		switch notification.(type) {
		case *pb.SCCMDPushNotification:
			msg := notification.(*pb.SCCMDPushNotification)
			puller := h.getPuller(msg.CfgId)

			// send publishing notification to target puller.
			if err := puller.HandlePub(msg); err != nil {
				logger.Error("ConfigHandler[%s %s %s] | pulling, handle publish notification to puller, %+v",
					h.bizID, h.appID, h.path, err)
			}

		case *pb.SCCMDPushRollbackNotification:
			msg := notification.(*pb.SCCMDPushRollbackNotification)
			puller := h.getPuller(msg.CfgId)

			// send publishing notification to target puller.
			if err := puller.HandleRoll(msg); err != nil {
				logger.Error("ConfigHandler[%s %s %s] | pulling, handle rollback notification to puller, %+v",
					h.bizID, h.appID, h.path, err)
			}

		case *pb.SCCMDPushReloadNotification:
			msg := notification.(*pb.SCCMDPushReloadNotification)

			// send publishing notification to target puller.
			if err := h.handleReload(msg); err != nil {
				logger.Error("ConfigHandler[%s %s %s] | pulling, handle reload notification to puller, %+v",
					h.bizID, h.appID, h.path, err)
			}

		default:
			logger.Error("ConfigHandler[%s %s %s]| unknow command[%+v]", h.bizID, h.appID, h.path, notification)
		}
	}
}

// handleReload handles reload event, you may not know why it's here,
// but here is the only interface to handle all puller of configs.
func (h *ConfigHandler) handleReload(msg *pb.SCCMDPushReloadNotification) error {
	if !h.viper.GetBool("instance.open") && !h.viper.GetBool("sidecar.fileReloadMode") {
		// instance service is not open and file reload mode is not open.
		logger.Warnf("ConfigHandler[%s %s %s]| instance-service and file-reload aren't open, can't do reload action",
			h.bizID, h.appID, h.path)
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
		return errors.New("empty reload spec")
	}

	// handle all reload spec info.
	var referenceReleaseName string
	var referenceReleaseID string

	metadatas := []*ReleaseMetadata{}

	for _, eInfo := range msg.ReloadSpec.Info {
		md, err := h.effectCache.LocalRelease(eInfo.CfgId)
		if err != nil || md == nil {
			// suppose no effected release.
			return fmt.Errorf("can't reload this release, config[%s] suppose not effected release[%s] this moment",
				eInfo.CfgId, eInfo.ReleaseId)
		}

		// check release.
		if !msg.ReloadSpec.Rollback {
			// normal reload.
			if md.ReleaseID != eInfo.ReleaseId {
				// not effectting target release this moment.
				return fmt.Errorf("can't reload this release, config[%s] not effected release[%s] this moment",
					eInfo.CfgId, eInfo.ReleaseId)
			}
		} else {
			// rollback reload.
			if md.ReleaseID == eInfo.ReleaseId {
				return fmt.Errorf("can't rollback reload this release, config[%s] not rollbacked "+
					"release[%s] this moment, %+v", eInfo.CfgId, eInfo.ReleaseId, md)
			}
		}

		referenceReleaseName = md.ReleaseName
		referenceReleaseID = eInfo.ReleaseId

		metadatas = append(metadatas, md)

		// NOTE: may other release is coming, but there should be a lock in user level.
		// Reload is just check local releases and send notification to business, you should
		// know all actions from your operators.
	}

	// all configs are effectting target release, reload now.
	spec := &ReloadSpec{
		BizID:       h.bizID,
		AppID:       h.appID,
		Path:        h.path,
		ReleaseName: referenceReleaseName,
	}

	// config reload specs.
	configSpec := []ConfigSpec{}
	for _, md := range metadatas {
		configSpec = append(configSpec, ConfigSpec{Name: md.CfgName, Fpath: md.CfgFpath})
	}
	spec.Configs = configSpec

	// reload mode.
	if len(msg.ReloadSpec.MultiReleaseId) != 0 {
		// multi release reload mode.
		spec.MultiReleaseID = msg.ReloadSpec.MultiReleaseId
	} else {
		// single release reload mode.
		spec.ReleaseID = referenceReleaseID
	}

	// reload type, 0: update  1: rollback  2.first reload.
	if msg.ReloadSpec.Rollback {
		spec.ReloadType = int32(ReloadTypeRollback)
	}

	// sync reload event.
	h.reloader.Reload(spec)

	return nil
}

// reporting keeps reporting local release effected information of
// all configs to connserver.
func (h *ConfigHandler) reporting() {
	ticker := time.NewTicker(h.viper.GetDuration("sidecar.reportInfoInterval"))
	defer ticker.Stop()

	for {
		if h.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(h.bizID, h.appID, h.path))) {
			logger.Info("ConfigHandler[%s %s %s]| stop reporting now!", h.bizID, h.appID, h.path)
			return
		}

		<-ticker.C

		h.mu.RLock()
		cfgIDs := []string{}
		for cfgID := range h.pullers {
			cfgIDs = append(cfgIDs, cfgID)
		}
		h.mu.RUnlock()

		if err := h.report(cfgIDs); err != nil {
			logger.Error("ConfigHandler[%s %s %s]| reporting, report local releases effected information, %+v",
				h.bizID, h.appID, h.path, err)
		}
		logger.Warn("ConfigHandler[%s %s %s]| reporting, report local releases effected information succcess, %+v",
			h.bizID, h.appID, h.path, cfgIDs)
	}
}

// syncConfigList keeps syncing config list from connserver.
func (h *ConfigHandler) syncConfigList() {
	// don't wait here at first time.
	isFirstTime := true

	ticker := time.NewTicker(h.viper.GetDuration("sidecar.syncConfigListInterval"))
	defer ticker.Stop()

	for {
		if h.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(h.bizID, h.appID, h.path))) {
			logger.Info("ConfigHandler[%s %s %s]| stop syncing config list now!", h.bizID, h.appID, h.path)
			return
		}

		if !isFirstTime {
			<-ticker.C
		} else {
			common.DelayRandomMS(1000)
		}

		// pull config list from connserver.
		cfgIDs, err := h.pullConfigList()
		if err != nil {
			logger.Error("ConfigHandler[%s %s %s]| syncConfigList, sync config list, %+v",
				h.bizID, h.appID, h.path, err)
			continue
		}

		// update local pullers.
		h.mu.Lock()
		newCfgIDs := make(map[string]string)
		for _, cfgID := range cfgIDs {
			if v, ok := h.pullers[cfgID]; !ok || v == nil {
				newPuller := NewPuller(h.viper, h.bizID, h.appID, h.path, cfgID, h.connSvrCli,
					h.effectCache, h.contentCache)

				h.pullers[cfgID] = newPuller
				newPuller.Run()
			}
			newCfgIDs[cfgID] = cfgID
		}

		for cfgID, puller := range h.pullers {
			if _, ok := newCfgIDs[cfgID]; !ok {
				puller.Stop()
			}
		}
		h.mu.Unlock()
		logger.Warn("ConfigHandler[%s %s %s]| syncConfigList, sync config list success, %+v",
			h.bizID, h.appID, h.path, cfgIDs)

		// handle instance start first reload notification.
		if isFirstTime {
			go h.handleFirstReload()
		}
		isFirstTime = false
	}
}

func (h *ConfigHandler) handleFirstReload() {
	for {
		// wait for pullers.
		time.Sleep(h.viper.GetDuration("sidecar.firstReloadCheckInterval"))

		if h.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(h.bizID, h.appID, h.path))) {
			return
		}

		if h.isFirstReloadSucc {
			// first reload already success.
			return
		}

		// handle first reload.
		h.mu.RLock()
		cfgIDs := []string{}
		for cfgID := range h.pullers {
			cfgIDs = append(cfgIDs, cfgID)
		}
		h.mu.RUnlock()

		// check cfgids effected status.
		isAllCfgsEffectedSucc := true
		metadatas := []*ReleaseMetadata{}

		for _, cfgID := range cfgIDs {

			// check local release.
			md, _ := h.effectCache.LocalRelease(cfgID)
			if md == nil {
				// no need to check others, just wait and check next round.
				logger.Warn("ConfigHandler[%s %s %s]| handleFirstReload, check local release for %+v, no effected "+
					"release this moment", h.bizID, h.appID, h.path, cfgID)

				isAllCfgsEffectedSucc = false
				break
			}
			metadatas = append(metadatas, md)
		}

		if !isAllCfgsEffectedSucc {
			// check next round.
			continue
		}

		// all configs already effected success this moment.
		// send reload notification now.
		spec := &ReloadSpec{
			BizID:       h.bizID,
			AppID:       h.appID,
			Path:        h.path,
			ReleaseName: FirstReloadReleaseName,
			ReloadType:  int32(ReloadTypeFirstReload),
		}

		configSpec := []ConfigSpec{}
		for _, md := range metadatas {
			configSpec = append(configSpec, ConfigSpec{Name: md.CfgName, Fpath: md.CfgFpath})
		}
		spec.Configs = configSpec

		// send reload event.
		h.reloader.Reload(spec)
		logger.Warn("ConfigHandler[%s %s %s]| handleFirstReload success!", h.bizID, h.appID, h.path)

		// mark first reload handled success.
		h.isFirstReloadSucc = true
		return
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
		if h.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(h.bizID, h.appID, h.path))) {
			logger.Info("ConfigHandler[%s %s %s]| stop debuging now!", h.bizID, h.appID, h.path)
			return
		}

		<-ticker.C

		h.mu.RLock()
		cfgIDs := []string{}
		for cfgID := range h.pullers {
			cfgIDs = append(cfgIDs, cfgID)
		}
		h.mu.RUnlock()

		for _, cfgID := range cfgIDs {
			logger.V(4).Infof("ConfigHandler[%s %s %s]| debug %s", h.bizID, h.appID, h.path, h.effectCache.Debug(cfgID))
		}
	}
}

func (h *ConfigHandler) processConnectionClient() {
	for {
		if err := h.makeConnectionClient(); err != nil {
			logger.Warnf("ConfigHandler[%s %s %s]| create new client for handler, %+v", h.bizID, h.appID, h.path, err)

			time.Sleep(time.Second)
			continue
		}
		logger.Infof("ConfigHandler[%s %s %s]| create client for new config handler success", h.bizID, h.appID, h.path)
		break
	}

	go func() {
		for {
			time.Sleep(time.Second)

			if h.viper.GetBool(fmt.Sprintf("appmod.%s.stop", ModKey(h.bizID, h.appID, h.path))) {
				if h.connSvrConn != nil {
					h.connSvrConn.Close()
				}
				break
			}
		}
	}()
}

// Reset resets the app runtime data for new instance.
func (h *ConfigHandler) Reset() {
	if h != nil {
		h.effectCache.Reset()
		logger.Warnf("ConfigHandler[%s %s %s]| reset effect cache success", h.bizID, h.appID, h.path)
	}
}

// Run runs the config handler.
func (h *ConfigHandler) Run() {
	// process connection client.
	h.processConnectionClient()

	// keep pulling release.
	go h.pulling()

	// keep reporting effected information.
	go h.reporting()

	// keep syncing config list.
	go h.syncConfigList()

	// debug.
	go h.Debug()
}
