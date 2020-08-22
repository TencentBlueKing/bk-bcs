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
	"os"
	"path/filepath"
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

// Puller is config puller.
type Puller struct {
	// viper as context here.
	viper *viper.Viper

	businessName string
	appName      string

	// configset id.
	cfgsetid string

	// config release effect cache.
	effectCache *EffectCache

	// config content cache.
	contentCache *ContentCache

	// stop pulling sig channel.
	stopCh chan bool

	// publish event channel.
	ch chan interface{}
}

// NewPuller creates new Puller.
func NewPuller(viper *viper.Viper, businessName, appName, cfgsetid string,
	effectCache *EffectCache, contentCache *ContentCache) *Puller {
	return &Puller{
		viper:        viper,
		businessName: businessName,
		appName:      appName,
		cfgsetid:     cfgsetid,
		effectCache:  effectCache,
		contentCache: contentCache,
		stopCh:       make(chan bool, 1),
		ch:           make(chan interface{}, viper.GetInt("sidecar.configHandlerChSize")),
	}
}

// makeConnectionClient returns connserver gRPC connection/client.
func (p *Puller) makeConnectionClient() (pb.ConnectionClient, *grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(p.viper.GetDuration("connserver.dialtimeout")),
	}

	endpoint := p.viper.GetString("connserver.hostname") + ":" + p.viper.GetString("connserver.port")
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, nil, err
	}
	client := pb.NewConnectionClient(conn)
	return client, conn, nil
}

// sidecarLabels marshals sidecar labels to string base on strategy protocol.
func (p *Puller) sidecarLabels() (string, error) {
	sidecarLabels := &strategy.SidecarLabels{Labels: p.viper.GetStringMapString(fmt.Sprintf("appmod.%s_%s.labels", p.businessName, p.appName))}
	labels, err := json.Marshal(sidecarLabels)
	if err != nil {
		return "", err
	}
	return string(labels), nil
}

func (p *Puller) configSetFpath(fpath string) string {
	return filepath.Clean(fmt.Sprintf("%s/%s", p.viper.GetString(fmt.Sprintf("appmod.%s_%s.path", p.businessName, p.appName)), fpath))
}

// effect effects the release in notification.
func (p *Puller) effect(metadata *ReleaseMetadata) error {
	// effect app real config.
	if err := p.contentCache.Effect(metadata.Cid, metadata.CfgsetName, p.configSetFpath(metadata.CfgsetFpath)); err != nil {
		if err := p.report(p.cfgsetid, metadata.Releaseid, metadata.EffectTime, err); err != nil {
			logger.Warn("Puller[%s %s][%+v]| report configs local effect, %+v", p.businessName, p.appName, p.cfgsetid, err)
		}
		return err
	}

	// add effect cache.
	if err := p.effectCache.Effect(metadata); err != nil {
		if err := p.report(p.cfgsetid, metadata.Releaseid, metadata.EffectTime, err); err != nil {
			logger.Warn("Puller[%s %s][%+v]| report configs local effect, %+v", p.businessName, p.appName, p.cfgsetid, err)
		}
		return err
	}

	// file reload mode option.
	if err := p.fileReload(filepath.Clean(fmt.Sprintf("%s/%s", metadata.CfgsetFpath, metadata.CfgsetName))); err != nil {
		logger.Warn("Puller[%s %s][%+v]| file reload mode, reload failed, %+v", p.businessName, p.appName, p.cfgsetid, err)
	}

	// report local effected release information right now.
	if err := p.report(p.cfgsetid, metadata.Releaseid, metadata.EffectTime, nil); err != nil {
		logger.Warn("Puller[%s %s][%+v]| report configs local effected, %+v", p.businessName, p.appName, p.cfgsetid, err)
	}
	logger.Warn("Puller[%s %s][%+v]| effect the release success, %+v", p.businessName, p.appName, p.cfgsetid, metadata)

	return nil
}

// fileReload execs the file mode reload action.
func (p *Puller) fileReload(configSetName string) error {
	if !p.viper.GetBool("sidecar.fileReloadMode") {
		return nil
	}

	// touch file to notify reload.
	fReloadFName := fmt.Sprintf("%s/%s", p.viper.GetString(fmt.Sprintf("appmod.%s_%s.path", p.businessName, p.appName)),
		p.viper.GetString("sidecar.fileReloadFName"))

	fReload, err := os.OpenFile(fReloadFName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("touch file failed, %+v", err)
	}
	defer fReload.Close()

	// write reload content.
	if _, err := fReload.WriteString(fmt.Sprintf("%s\n%d", configSetName, time.Now().Unix())); err != nil {
		return fmt.Errorf("write file content failed, %+v", err)
	}
	logger.Infof("Puller[%s %s][%+v]| file reload mode, notify reload success!", p.businessName, p.appName, p.cfgsetid)

	return nil
}

// pullTarget returns target release information.
func (p *Puller) pullTarget(targetRelease string) (bool, *pbcommon.Release, string, string, error) {
	return p.pullRelease(targetRelease)
}

// pullNewest returns the newest release information.
func (p *Puller) pullNewest() (bool, *pbcommon.Release, string, string, error) {
	return p.pullRelease("")
}

// pullRelease pulls release information from connserver.
func (p *Puller) pullRelease(target string) (bool, *pbcommon.Release, string, string, error) {
	// make connserver gRPC client now.
	client, conn, err := p.makeConnectionClient()
	if err != nil {
		return false, nil, "", "", err
	}
	defer conn.Close()

	// marshal sidecar labels.
	labels, err := p.sidecarLabels()
	if err != nil {
		return false, nil, "", "", err
	}

	// local releaseid.
	md, err := p.effectCache.LocalRelease(p.cfgsetid)
	if err != nil {
		return false, nil, "", "", err
	}

	if md == nil {
		md = &ReleaseMetadata{}
	}

	r := &pb.PullReleaseReq{
		Seq:            common.Sequence(),
		Bid:            p.viper.GetString(fmt.Sprintf("appmod.%s_%s.bid", p.businessName, p.appName)),
		Appid:          p.viper.GetString(fmt.Sprintf("appmod.%s_%s.appid", p.businessName, p.appName)),
		Clusterid:      p.viper.GetString(fmt.Sprintf("appmod.%s_%s.clusterid", p.businessName, p.appName)),
		Zoneid:         p.viper.GetString(fmt.Sprintf("appmod.%s_%s.zoneid", p.businessName, p.appName)),
		Dc:             p.viper.GetString(fmt.Sprintf("appmod.%s_%s.dc", p.businessName, p.appName)),
		IP:             p.viper.GetString("appinfo.ip"),
		Labels:         labels,
		Cfgsetid:       p.cfgsetid,
		LocalReleaseid: md.Releaseid,
		Releaseid:      target,
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.viper.GetDuration("connserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Puller[%s %s][%+v]| request to connserver PullRelease, %+v", p.businessName, p.appName, p.cfgsetid, r)

	resp, err := client.PullRelease(ctx, r)
	if err != nil {
		return false, nil, "", "", err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return false, nil, "", "", errors.New(resp.ErrMsg)
	}
	return resp.NeedEffect, resp.Release, resp.Cid, resp.CfgLink, nil
}

// pullReleaseConfigs pulls release configs from connserver.
func (p *Puller) pullReleaseConfigs(releaseid, cfgsetid, cid string) (string, string, []byte, error) {
	// make connserver gRPC client now.
	client, conn, err := p.makeConnectionClient()
	if err != nil {
		return "", "", nil, err
	}
	defer conn.Close()

	r := &pb.PullReleaseConfigsReq{
		Seq:       common.Sequence(),
		Bid:       p.viper.GetString(fmt.Sprintf("appmod.%s_%s.bid", p.businessName, p.appName)),
		Appid:     p.viper.GetString(fmt.Sprintf("appmod.%s_%s.appid", p.businessName, p.appName)),
		Clusterid: p.viper.GetString(fmt.Sprintf("appmod.%s_%s.clusterid", p.businessName, p.appName)),
		Zoneid:    p.viper.GetString(fmt.Sprintf("appmod.%s_%s.zoneid", p.businessName, p.appName)),
		IP:        p.viper.GetString("appinfo.ip"),
		Cfgsetid:  cfgsetid,
		Releaseid: releaseid,
		Cid:       cid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.viper.GetDuration("connserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Puller[%s %s][%+v]| request to connserver PullReleaseConfigs, %+v", p.businessName, p.appName, cfgsetid, r)

	resp, err := client.PullReleaseConfigs(ctx, r)
	if err != nil {
		return "", "", nil, err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return "", "", nil, errors.New(resp.ErrMsg)
	}
	if resp.Cid != cid {
		return "", "", nil, fmt.Errorf("inconsistent cid[%+v][%+v]", cid, resp.Cid)
	}
	return resp.Cid, resp.CfgLink, resp.Content, nil
}

// handle target release in notification.
func (p *Puller) handlePubNotification(notification *pb.SCCMDPushNotification) (*ReleaseMetadata, error) {
	// check release publishing metadata.
	if notification == nil || notification.Cfgsetid != p.cfgsetid {
		return nil, fmt.Errorf("invalid notification, %+v", notification)
	}

	// pull target release metadata in notification.
	needEffect, release, cid, cfgLink, err := p.pullTarget(notification.Releaseid)
	if err != nil {
		return nil, fmt.Errorf("can't pull target release[%s], %+v", notification.Releaseid, err)
	}
	if !needEffect {
		return nil, fmt.Errorf("pull target release[%s], no need to effect.", notification.Releaseid)
	}

	// check pull back release.
	if notification.Cfgsetid != release.Cfgsetid || notification.Serialno != release.ID {
		return nil, fmt.Errorf("pull target release, invalid release information, %+v, %+v", notification, release)
	}

	// release metadata pull back base on notification.
	metadata := &ReleaseMetadata{
		Cfgsetid:       release.Cfgsetid,
		CfgsetName:     release.CfgsetName,
		Serialno:       release.ID,
		Releaseid:      notification.Releaseid,
		ReleaseName:    release.Name,
		MultiReleaseid: release.MultiReleaseid,
		Cid:            cid,
		CfgLink:        cfgLink,
		CfgsetFpath:    release.CfgsetFpath,
		EffectTime:     time.Now().Format("2006-01-02 15:04:05"),
	}
	return metadata, nil
}

// pull newest to get right release now, even have not effected in local.
// The intention is just make the release version correct.
func (p *Puller) handleNewestRelease() (*ReleaseMetadata, error) {
	needEffect, release, cid, cfgLink, err := p.pullNewest()
	if err != nil {
		return nil, fmt.Errorf("can't pull newest release, %+v", err)
	}
	if !needEffect {
		return nil, errors.New("pull newest release, no newest need to effect.")
	}

	// release metadata.
	metadata := &ReleaseMetadata{
		Cfgsetid:       release.Cfgsetid,
		CfgsetName:     release.CfgsetName,
		Serialno:       release.ID,
		Releaseid:      release.Releaseid,
		ReleaseName:    release.Name,
		MultiReleaseid: release.MultiReleaseid,
		Cid:            cid,
		CfgLink:        cfgLink,
		CfgsetFpath:    release.CfgsetFpath,
		EffectTime:     time.Now().Format("2006-01-02 15:04:05"),
	}
	return metadata, nil
}

// pulling keeps pulling configs.
func (p *Puller) pulling() {
	isFirstPulling := true

	for {
		if p.viper.GetBool(fmt.Sprintf("appmod.%s_%s.stop", p.businessName, p.appName)) {
			logger.Info("Puller[%s %s][%+v]| stop pulling now!", p.businessName, p.appName, p.cfgsetid)
			return
		}

		var metadata *ReleaseMetadata

		// effect with serial num, unless rollback or newest logic.
		needEffectWithSerialNo := true

		autoPullInterval := p.viper.GetDuration("sidecar.pullConfigInterval")
		if isFirstPulling {
			autoPullInterval = time.Second
		}

		select {
		// stop pulling signal.
		case <-p.stopCh:
			logger.Warn("Puller[%s %s][%+v]-pulling| stop pulling now", p.businessName, p.appName, p.cfgsetid)
			// just return, Puller is asked to exist.
			return

		// handle publishing notifications.
		case notification := <-p.ch:
			isFirstPulling = false

			// handle multi type notifications.
			switch notification.(type) {
			case *pb.SCCMDPushNotification:
				// normal release publish notification.
				pubNotification := notification.(*pb.SCCMDPushNotification)

				md, err := p.handlePubNotification(pubNotification)
				if err != nil {
					logger.Warn("Puller[%s %s][%+v]-pulling| handle publish, %+v", p.businessName, p.appName, p.cfgsetid, err)
					continue
				}
				metadata = md

				logger.Warn("Puller[%s %s][%+v]-pulling| recviced publishing notification, metadata %+v",
					p.businessName, p.appName, p.cfgsetid, metadata)

			case *pb.SCCMDPushRollbackNotification:
				// release rollback publishing notification.
				rollbackNotification := notification.(*pb.SCCMDPushRollbackNotification)

				logger.Info("Puller[%s %s][%+v]-pulling| recviced rollback publishing notification, %+v",
					p.businessName, p.appName, p.cfgsetid, rollbackNotification)

				// need effect without serial num(rollback event).
				needEffectWithSerialNo = false

				md, err := p.handleNewestRelease()
				if err != nil {
					logger.Warn("Puller[%s %s][%+v]-pulling| handle rollback publish, %+v", p.businessName, p.appName, p.cfgsetid, err)
					continue
				}
				metadata = md
				metadata.isRollback = true

				logger.Warn("Puller[%s %s][%+v]-pulling| rollback publish, newest release, %+v",
					p.businessName, p.appName, p.cfgsetid, metadata)

			default:
				logger.Error("Puller[%s %s][%+v]-pulling| unknow notification[%+v]",
					p.businessName, p.appName, p.cfgsetid, notification)
			}

		case <-time.After(autoPullInterval):
			isFirstPulling = false

			// newest logic need effect without serial num(rollback event).
			needEffectWithSerialNo = false

			md, err := p.handleNewestRelease()
			if err != nil {
				logger.Warn("Puller[%s %s][%+v]-pulling| handle pull newest, %+v", p.businessName, p.appName, p.cfgsetid, err)
				continue
			}
			metadata = md

			logger.Warn("Puller[%s %s][%+v]-pulling| recviced newest release, %+v",
				p.businessName, p.appName, p.cfgsetid, metadata)
		}

		// check if need to effect this release.
		// Newest release and rollback logic no need to effect with serial num.
		if needEffectWithSerialNo {
			// compare local release serial num.
			needEffect, err := p.effectCache.NeedEffected(metadata.Cfgsetid, metadata.Serialno)
			if err != nil {
				logger.Error("Puller[%s %s][%+v]-pulling| check local effect information, %+v",
					p.businessName, p.appName, p.cfgsetid, err)
				continue
			}
			if !needEffect {
				logger.Warn("Puller[%s %s][%+v]-pulling| finally, no need to effect the release, %+v",
					p.businessName, p.appName, p.cfgsetid, metadata)
				continue
			}
		}

		// mark event type.
		lmd, err := p.effectCache.LocalRelease(metadata.Cfgsetid)
		if err != nil {
			logger.Warn("Puller[%s %s][%+v]-pulling| mark event type, %+v", p.businessName, p.appName, p.cfgsetid, err)
		} else {
			if lmd != nil {
				if metadata.isRollback || metadata.Serialno < lmd.Serialno {
					// recved a rollback publishing or pull newest on time get an old release.
					metadata.isRollback = true
				}
			}
		}

		// check local file content cache.
		if cached, err := p.contentCache.Has(metadata.Cid); err == nil && cached {
			logger.Warn("Puller[%s %s][%+v]-pulling| has the content cache[%+v], and effect right now.",
				p.businessName, p.appName, p.cfgsetid, metadata.Cid)

			if err := p.effect(metadata); err != nil {
				logger.Error("Puller[%s %s][%+v]-pulling| after cache checking, can't effect release, %+v",
					p.businessName, p.appName, p.cfgsetid, err)
			}
			continue
		}

		// has no cache, try to pull back.
		cid, cfgLink, content, err := p.pullReleaseConfigs(metadata.Releaseid, p.cfgsetid, metadata.Cid)
		if err != nil {
			logger.Error("Puller[%s %s][%+v]-pulling| can't pull the release configs, %+v", p.businessName, p.appName, p.cfgsetid, err)
			continue
		}

		// TODO download base on cfglink.

		// add config content cache.
		logger.Warn("Puller[%s %s][%+v]-pulling| pull release[%+v] back, add content cache now.",
			p.businessName, p.appName, p.cfgsetid, metadata.Releaseid)

		if err := p.contentCache.Add(&Content{Cid: cid, CfgLink: cfgLink, Metadata: content}); err != nil {
			logger.Error("Puller[%s %s][%+v]-pulling| add config content cache, %+v.",
				p.businessName, p.appName, p.cfgsetid, err)
			continue
		}

		// effect this release now.
		if err := p.effect(metadata); err != nil {
			logger.Error("Puller[%s %s][%+v]-pulling| after adding cache, can't effect release, %+v",
				p.businessName, p.appName, p.cfgsetid, err)
		}
	}
}

// report reports release effect information.
func (p *Puller) report(cfgsetid, releaseid, effectTime string, effectErr error) error {
	reportInfos := []*pbcommon.ReportInfo{}

	reportInfo := &pbcommon.ReportInfo{
		Cfgsetid:   cfgsetid,
		Releaseid:  releaseid,
		EffectTime: effectTime,
	}

	if effectErr == nil {
		reportInfo.EffectCode = 0
		reportInfo.EffectMsg = "SUCCESS"
	} else {
		reportInfo.EffectCode = 1
		reportInfo.EffectMsg = effectErr.Error()
	}
	reportInfos = append(reportInfos, reportInfo)

	// make connserver gRPC client now.
	client, conn, err := p.makeConnectionClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	// marshal sidecar labels.
	labels, err := p.sidecarLabels()
	if err != nil {
		return err
	}

	r := &pb.ReportReq{
		Seq:       common.Sequence(),
		Bid:       p.viper.GetString(fmt.Sprintf("appmod.%s_%s.bid", p.businessName, p.appName)),
		Appid:     p.viper.GetString(fmt.Sprintf("appmod.%s_%s.appid", p.businessName, p.appName)),
		Clusterid: p.viper.GetString(fmt.Sprintf("appmod.%s_%s.clusterid", p.businessName, p.appName)),
		Zoneid:    p.viper.GetString(fmt.Sprintf("appmod.%s_%s.zoneid", p.businessName, p.appName)),
		Dc:        p.viper.GetString(fmt.Sprintf("appmod.%s_%s.dc", p.businessName, p.appName)),
		IP:        p.viper.GetString("appinfo.ip"),
		Labels:    labels,
		Infos:     reportInfos,
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.viper.GetDuration("connserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("Puller[%s %s][%+v]| request to connserver Report, %+v", p.businessName, p.appName, p.cfgsetid, r)

	resp, err := client.Report(ctx, r)
	if err != nil {
		return err
	}
	if resp.ErrCode != pbcommon.ErrCode_E_OK {
		return errors.New(resp.ErrMsg)
	}
	return nil
}

// HandlePub handles release publishing notification.
func (p *Puller) HandlePub(notification *pb.SCCMDPushNotification) error {
	if notification == nil {
		return errors.New("invalid publish notification struct: nil")
	}

	select {
	case p.ch <- notification:
	case <-time.After(p.viper.GetDuration("sidecar.configHandlerChTimeout")):
		return fmt.Errorf("send cmd to config handler puller channel timeout, %+v", notification)
	}

	return nil
}

// HandleRoll handles release rollback publishing notification.
func (p *Puller) HandleRoll(notification *pb.SCCMDPushRollbackNotification) error {
	if notification == nil {
		return errors.New("invalid rollback notification struct: nil")
	}

	select {
	case p.ch <- notification:
	case <-time.After(p.viper.GetDuration("sidecar.configHandlerChTimeout")):
		return fmt.Errorf("send cmd to config handler puller channel timeout, %+v", notification)
	}

	return nil
}

// Stop stops the puller.
func (p *Puller) Stop() {
	select {
	case p.stopCh <- true:
	case <-time.After(time.Second):
		logger.Warn("Puller[%s %s][%+v]| stop puller timeout.", p.businessName, p.appName, p.cfgsetid)
	}
}

// Run runs the puller.
func (p *Puller) Run() {
	go p.pulling()
}
