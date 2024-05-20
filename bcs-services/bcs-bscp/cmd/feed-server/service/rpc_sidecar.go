/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	prm "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/bcs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/gse"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbkv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/kv"
	pbfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/feed-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

var (
	// LabelKeyAgentID is the key of agent id in bcs node labels.
	LabelKeyAgentID = "bkcmdb.tencent.com/bk-agent-id"
)

// Handshake received handshake from sidecar to validate the app instance's authorization and legality.
func (s *Service) Handshake(ctx context.Context, hm *pbfs.HandshakeMessage) (*pbfs.HandshakeResp, error) {

	im, err := sfs.ParseFeedIncomingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = hm.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid handshake message "+err.Error())
	}

	// check if the sidecar's version can be accepted.
	if !sfs.IsAPIVersionMatch(hm.ApiVersion) {
		return nil, status.Error(codes.InvalidArgument, "sdk's api version is too low, should be upgraded")
	}

	// check if the sidecar's version can be accepted.
	if !sfs.IsSidecarVersionMatch(hm.Spec.Version) {
		return nil, status.Error(codes.InvalidArgument, "sdk's version is too low, should be upgraded")
	}

	ra := &meta.ResourceAttribute{Basic: meta.Basic{Type: meta.Sidecar, Action: meta.Access}, BizID: hm.Spec.BizId}
	authorized, err := s.bll.Auth().Authorize(im.Kit, ra)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	if !authorized {
		return nil, status.Errorf(codes.PermissionDenied, "no permission to access bscp server")
	}

	// Note:
	// 1. get the basic configurations for sidecar if necessary, and send it back to sidecar later.
	// 2. collect the basic info for the app with biz dimension.

	decorator := s.provider.URIDecorator(hm.Spec.BizId)
	payload := &sfs.SidecarHandshakePayload{
		ServiceInfo: &sfs.ServiceInfo{
			Name: s.name,
		},
		RuntimeOption: &sfs.SidecarRuntimeOption{
			BounceIntervalHour: s.dsSetting.BounceIntervalHour,
			Repository: &sfs.RepositoryV1{
				Root: decorator.Root(),
				Url:  decorator.Url(),
			},
			EnableAsyncDownload: cc.FeedServer().GSE.Enabled,
		},
	}

	payloadBytes, err := jsoni.Marshal(payload)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "encode payload failed, err: %v", err)
	}

	return &pbfs.HandshakeResp{ApiVersion: sfs.CurrentAPIVersion, Payload: payloadBytes}, nil
}

// Watch the change message from feed server for sidecar.
func (s *Service) Watch(swm *pbfs.SideWatchMeta, fws pbfs.Upstream_WatchServer) error {

	// check if the sidecar's version can be accepted.
	if !sfs.IsAPIVersionMatch(swm.ApiVersion) {
		return status.Error(codes.InvalidArgument, "sidecar's api version is too low, should be upgraded")
	}

	im, err := sfs.ParseFeedIncomingContext(fws.Context())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	ra := &meta.ResourceAttribute{Basic: meta.Basic{Type: meta.Sidecar, Action: meta.Access}, BizID: im.Meta.BizID}
	authorized, err := s.bll.Auth().Authorize(im.Kit, ra)
	if err != nil {
		return status.Errorf(codes.Aborted, "do authorization failed, %s", err.Error())
	}

	if !authorized {
		return status.Error(codes.PermissionDenied, "no permission to access bscp server")
	}

	// parse the payload according the different api version.
	payload := new(sfs.SideWatchPayload)
	if err := jsoni.Unmarshal(swm.Payload, payload); err != nil {
		return status.Errorf(codes.Aborted, "parse request payload failed, %s", err.Error())
	}

	for i := range payload.Applications {
		appID, err := s.bll.AppCache().GetAppID(im.Kit, payload.BizID, payload.Applications[i].App)
		if err != nil {
			return status.Errorf(codes.Aborted, "get app id failed, %s", err.Error())
		}
		payload.Applications[i].AppID = appID
	}

	if err := payload.Validate(); err != nil {
		return status.Errorf(codes.Aborted, "invalid payload, err: %s", err.Error())
	}

	var msg string
	for _, one := range payload.Applications {
		msg += fmt.Sprintf("biz: %d, app: %s, uid: %s, labels: %s, ", payload.BizID, one.App, one.Uid, one.Labels)
	}

	logs.Infof("received sidecar watch request, biz: %d, %s fingerprint: %s, rid: %s.", im.Meta.BizID, msg,
		im.Meta.Fingerprint, im.Kit.Rid)

	s.mc.watchTotal.With(prm.Labels{"biz": tools.Itoa(im.Meta.BizID)}).Inc()
	defer s.mc.watchTotal.With(prm.Labels{"biz": tools.Itoa(im.Meta.BizID)}).Dec()
	s.mc.watchCounter.With(prm.Labels{"biz": tools.Itoa(im.Meta.BizID)}).Inc()

	if err := s.bll.Release().Watch(im, payload, fws); err != nil {
		logs.Errorf("sidecar watch failed, err: %v, rid: %s.", err, im.Kit.Rid)
		return status.Errorf(codes.Aborted, "do watch job failed, err: %v", err)
	}

	logs.Infof("finished watch job from sidecar, rid: %s", im.Kit.Rid)

	return nil
}

// Messaging received messages delivered from sidecar.
func (s *Service) Messaging(ctx context.Context, msg *pbfs.MessagingMeta) (*pbfs.MessagingResp, error) { // nolint
	im, err := sfs.ParseFeedIncomingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ra := &meta.ResourceAttribute{Basic: meta.Basic{Type: meta.Sidecar, Action: meta.Access}, BizID: im.Meta.BizID}
	authorized, err := s.bll.Auth().Authorize(im.Kit, ra)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	if !authorized {
		return nil, status.Errorf(codes.PermissionDenied, "no permission to access bscp server")
	}

	clientMetricData := make(map[uint32]*sfs.ClientMetricData)
	// 按照服务级别上报数据
	// 上报的事件分两种 心跳事件、变更事件
	switch sfs.MessagingType(msg.Type) {
	case sfs.VersionChangeMessage:
		vc := new(sfs.VersionChangePayload)
		err = vc.Decode(msg.Payload)
		if err != nil {
			logs.Errorf("version change message decoding failed, %s", err.Error())
			return nil, err
		}

		if vc.BasicData.BizID != 0 {
			appID, errApp := s.bll.AppCache().GetAppID(im.Kit, im.Meta.BizID, vc.Application.App)
			if errApp != nil {
				logs.Errorf("get app id failed, %s", errApp.Error())
				return nil, errApp
			}
			vc.Application.AppID = appID

			// pull 首次是需要获取app meta, 会出现权限等问题导致失败，
			// 因此TargetReleaseID会出现0的情况，
			// 获取TargetReleaseID时出现错误直接忽略
			if vc.Application.TargetReleaseID == 0 {
				meta := &types.AppInstanceMeta{
					BizID:  vc.BasicData.BizID,
					App:    vc.Application.App,
					AppID:  appID,
					Uid:    vc.Application.Uid,
					Labels: vc.Application.Labels,
				}
				cancel := im.Kit.CtxWithTimeoutMS(1500)
				defer cancel()
				metas, _ := s.bll.Release().ListAppLatestReleaseMeta(im.Kit, meta)
				vc.Application.TargetReleaseID = metas.ReleaseId
			}

			// 处理 心跳时间和在线状态
			vc.BasicData.HeartbeatTime = time.Now().Local().UTC()
			vc.BasicData.OnlineStatus = sfs.Online
			payload, errE := vc.Encode()
			if errE != nil {
				logs.Errorf("version change message encoding failed, %s", errE.Error())
				return nil, err
			}
			s.handleResourceUsageMetrics(vc.BasicData.BizID, vc.Application.App, vc.ResourceUsage)
			clientMetricData[appID] = &sfs.ClientMetricData{
				MessagingType: msg.Type,
				Payload:       payload,
			}
		}
	case sfs.Heartbeat:
		hb := new(sfs.HeartbeatPayload)
		err = hb.Decode(msg.Payload)
		if err != nil {
			return nil, err
		}

		if hb.BasicData.BizID != 0 {
			heartbeatTime := time.Now().UTC()
			onlineStatus := sfs.Online
			for _, item := range hb.Applications {
				if item.CursorID != "" {
					appID, errApp := s.bll.AppCache().GetAppID(im.Kit, im.Meta.BizID, item.App)
					if errApp != nil {
						logs.Errorf("get app id failed, %s", errApp.Error())
						return nil, errApp
					}
					item.AppID = appID
					s.handleResourceUsageMetrics(hb.BasicData.BizID, item.App, hb.ResourceUsage)
					hb.BasicData.HeartbeatTime = heartbeatTime
					hb.BasicData.OnlineStatus = onlineStatus
					oneData := sfs.HeartbeatItem{
						BasicData:     hb.BasicData,
						Application:   item,
						ResourceUsage: hb.ResourceUsage,
					}
					marshal, errHb := oneData.Encode()
					if errHb != nil {
						return nil, errHb
					}
					clientMetricData[appID] = &sfs.ClientMetricData{
						MessagingType: msg.Type,
						Payload:       marshal,
					}
				}
			}
		}
	}

	for appID, v := range clientMetricData {
		payload, err := jsoni.Marshal(v)
		if err != nil {
			logs.Errorf("failed to serialize clientMetricData, err: %s", err.Error())
			continue
		}
		if im.Meta.BizID != 0 && len(payload) != 0 {
			err = s.bll.ClientMetric().Set(im.Kit, im.Meta.BizID, appID, payload)
			if err != nil {
				logs.Errorf("send %d biz %s message, payload: %s, rid: %s", im.Meta.BizID, im.Meta.Fingerprint,
					payload, msg.Rid)
				continue
			}
		}
	}
	logs.V(3).Infof("receive %d biz %s sidecar %s message, payload: %s, rid: %s", im.Meta.BizID, im.Meta.Fingerprint,
		sfs.MessagingType(msg.Type).String(), msg.Payload, msg.Rid)
	return new(pbfs.MessagingResp), nil
}

// PullAppFileMeta pull an app's latest release metadata only when the app's configures is file type.
func (s *Service) PullAppFileMeta(ctx context.Context, req *pbfs.PullAppFileMetaReq) ( // nolint
	*pbfs.PullAppFileMetaResp, error) {

	// check if the sidecar's version can be accepted.
	if !sfs.IsAPIVersionMatch(req.ApiVersion) {
		st := status.New(codes.FailedPrecondition, "sdk's api version is too low, should be upgraded")
		st, err := st.WithDetails(&pbbase.ErrDetails{
			PrimaryError:   uint32(sfs.VersionIsTooLowFailed),
			SecondaryError: uint32(sfs.SDKVersionIsTooLowFailed),
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "grpc status with details failed")
		}
		return nil, st.Err()
	}

	im, err := sfs.ParseFeedIncomingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ra := &meta.ResourceAttribute{Basic: meta.Basic{Type: meta.Sidecar, Action: meta.Access}, BizID: im.Meta.BizID}
	authorized, err := s.bll.Auth().Authorize(im.Kit, ra)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "do authorization failed, %s", err.Error())
	}

	if !authorized {
		return nil, status.Error(codes.PermissionDenied, "no permission to access bscp server")
	}

	if req.AppMeta == nil {
		return nil, status.Error(codes.InvalidArgument, "app meta is empty")
	}

	appID, err := s.bll.AppCache().GetAppID(im.Kit, req.BizId, req.GetAppMeta().App)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "get app id failed, %s", err.Error())
	}
	meta := &types.AppInstanceMeta{
		BizID:  req.BizId,
		App:    req.GetAppMeta().App,
		AppID:  appID,
		Uid:    req.AppMeta.Uid,
		Labels: req.AppMeta.Labels,
	}

	cancel := im.Kit.CtxWithTimeoutMS(1500)
	defer cancel()

	metas, err := s.bll.Release().ListAppLatestReleaseMeta(im.Kit, meta)
	if err != nil {
		// appid等未找到, 刷新缓存, 客户端重试请求
		if isAppNotExistErr(err) {
			s.bll.AppCache().RemoveCache(im.Kit, req.BizId, req.GetAppMeta().App)
		}
		return nil, err
	}

	fileMetas := make([]*pbfs.FileMeta, 0, len(metas.ConfigItems))
	for _, ci := range metas.ConfigItems {
		ok, _ := tools.MatchConfigItem(req.Key, ci.ConfigItemSpec.Path, ci.ConfigItemSpec.Name)
		if req.Key != "" && !ok {
			continue
		}
		app, err := s.bll.AppCache().GetMeta(im.Kit, req.BizId, ci.ConfigItemAttachment.AppId)
		if err != nil {
			return nil, status.Errorf(codes.Aborted, "get app meta failed, %s", err.Error())
		}
		if match, err := s.bll.Auth().CanMatchCI(im.Kit, req.BizId, app.Name, req.Token,
			ci.ConfigItemSpec.Path, ci.ConfigItemSpec.Name); err != nil || !match {
			logs.Errorf("no permission to access config item %d, err: %v", ci.RciId, err)
			return nil, status.Errorf(codes.PermissionDenied, "no permission to access config item %d", ci.RciId)
		}
		fileMetas = append(fileMetas, &pbfs.FileMeta{
			Id:                   ci.RciId,
			CommitId:             ci.CommitID,
			CommitSpec:           ci.CommitSpec,
			ConfigItemSpec:       ci.ConfigItemSpec,
			ConfigItemAttachment: ci.ConfigItemAttachment,
			ConfigItemRevision:   ci.ConfigItemRevision,
			RepositorySpec: &pbfs.RepositorySpec{
				Path: ci.RepositorySpec.Path,
			},
		})
	}
	resp := &pbfs.PullAppFileMetaResp{
		ReleaseId:   metas.ReleaseId,
		ReleaseName: metas.ReleaseName,
		Repository: &pbfs.Repository{
			Root: metas.Repository.Root,
		},
		FileMetas: fileMetas,
		PreHook:   metas.PreHook,
		PostHook:  metas.PostHook,
	}

	return resp, nil
}

// GetDownloadURL get the download url of the file.
func (s *Service) GetDownloadURL(ctx context.Context, req *pbfs.GetDownloadURLReq) (
	*pbfs.GetDownloadURLResp, error) {
	// check if the sidecar's version can be accepted.
	if !sfs.IsAPIVersionMatch(req.ApiVersion) {
		st := status.New(codes.FailedPrecondition, "sdk's api version is too low, should be upgraded")
		st, err := st.WithDetails(&pbbase.ErrDetails{
			PrimaryError:   uint32(sfs.VersionIsTooLowFailed),
			SecondaryError: uint32(sfs.SDKVersionIsTooLowFailed),
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "grpc status with details failed")
		}
		return nil, st.Err()
	}

	im, err := sfs.ParseFeedIncomingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	app, err := s.bll.AppCache().GetMeta(im.Kit, req.BizId, req.FileMeta.ConfigItemAttachment.AppId)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "get app meta failed, %s", err.Error())
	}

	// validate can file be downloaded by credential.
	match, err := s.bll.Auth().CanMatchCI(
		im.Kit, req.BizId, app.Name, req.Token, req.FileMeta.ConfigItemSpec.Path, req.FileMeta.ConfigItemSpec.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "do authorization failed, %s", err.Error())
	}

	if !match {
		return nil, status.Error(codes.PermissionDenied, "no permission to download file")
	}

	// range download swap buffer size is 2MB, so we need to set the permits to byteSize / 2MB,
	// and then set permits to twice to left space for retry.
	fetchLimit := uint32(req.FileMeta.CommitSpec.Content.ByteSize/1024) + 1

	// 生成下载链接
	im.Kit.BizID = req.BizId
	downloadLink, err := s.provider.DownloadLink(im.Kit, req.FileMeta.CommitSpec.Content.Signature, fetchLimit)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "generate temp download url failed, %s", err.Error())
	}

	return &pbfs.GetDownloadURLResp{Url: downloadLink}, nil
}

// PullKvMeta pull an app's latest release metadata only when the app's configures is kv type.
func (s *Service) PullKvMeta(ctx context.Context, req *pbfs.PullKvMetaReq) (*pbfs.PullKvMetaResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if req.GetAppMeta() == nil || req.GetAppMeta().App == "" {
		return nil, status.Error(codes.InvalidArgument, "app_meta is required")
	}

	credential := getCredential(ctx)
	if !credential.MatchApp(req.AppMeta.App) {
		return nil, status.Errorf(codes.PermissionDenied, "not have app %s permission", req.AppMeta.App)
	}

	appID, err := s.bll.AppCache().GetAppID(kt, req.BizId, req.AppMeta.App)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "get app id failed, %s", err.Error())
	}

	app, err := s.bll.AppCache().GetMeta(kt, req.BizId, appID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "get app failed, %s", err.Error())
	}

	if app.ConfigType != table.KV {
		return nil, status.Errorf(codes.Aborted, "app not %s type", table.KV)
	}

	meta := &types.AppInstanceMeta{
		BizID:  req.BizId,
		App:    req.AppMeta.App,
		AppID:  appID,
		Uid:    req.AppMeta.Uid,
		Labels: req.AppMeta.Labels,
	}

	metas, err := s.bll.Release().ListAppLatestReleaseKvMeta(kt, meta)
	if err != nil {
		// appid等未找到, 刷新缓存, 客户端重试请求
		if isAppNotExistErr(err) {
			s.bll.AppCache().RemoveCache(kt, req.BizId, req.AppMeta.App)
		}
		return nil, err
	}

	kvMetas := make([]*pbfs.KvMeta, 0, len(metas.Kvs))
	for _, kv := range metas.Kvs {
		// 只返回有权限的kv
		if !credential.MatchKv(req.AppMeta.App, kv.Key) {
			continue
		}

		// 客户端匹配
		if !matchPattern(kv.Key, req.Match) {
			continue
		}

		kvMetas = append(kvMetas, &pbfs.KvMeta{
			Key:      kv.Key,
			KvType:   kv.KvType,
			Revision: kv.Revision,
			KvAttachment: &pbkv.KvAttachment{
				BizId: kv.KvAttachment.BizId,
				AppId: kv.KvAttachment.AppId,
			},
			ContentSpec: kv.ContentSpec,
		})
	}

	resp := &pbfs.PullKvMetaResp{
		ReleaseId: metas.ReleaseId,
		KvMetas:   kvMetas,
	}

	return resp, nil
}

// GetKvValue get kv value
func (s *Service) GetKvValue(ctx context.Context, req *pbfs.GetKvValueReq) (*pbfs.GetKvValueResp, error) {
	kt := kit.FromGrpcContext(ctx)
	if req.GetAppMeta() == nil || req.GetAppMeta().App == "" {
		return nil, status.Error(codes.InvalidArgument, "app_meta is required")
	}

	credential := getCredential(ctx)
	if !credential.MatchApp(req.AppMeta.App) {
		return nil, status.Errorf(codes.PermissionDenied, "not have app %s permission", req.AppMeta.App)
	}

	if !credential.MatchKv(req.AppMeta.App, req.Key) {
		return nil, status.Error(codes.PermissionDenied, "no permission get value")
	}

	appID, err := s.bll.AppCache().GetAppID(kt, req.BizId, req.GetAppMeta().App)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "get app id failed, %s", err.Error())
	}

	app, err := s.bll.AppCache().GetMeta(kt, req.BizId, appID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "get app failed, %s", err.Error())
	}

	if app.ConfigType != table.KV {
		return nil, status.Errorf(codes.Aborted, "app not %s type", table.KV)
	}

	meta := &types.AppInstanceMeta{
		BizID:  req.BizId,
		App:    req.GetAppMeta().App,
		AppID:  appID,
		Uid:    req.AppMeta.Uid,
		Labels: req.AppMeta.Labels,
	}

	metas, err := s.bll.Release().ListAppLatestReleaseKvMeta(kt, meta)
	if err != nil {
		return nil, err
	}

	rkv, err := s.bll.RKvCache().GetKvValue(kt, req.BizId, appID, metas.ReleaseId, req.Key)
	if err != nil {
		// appid等未找到, 刷新缓存, 客户端重试请求
		if isAppNotExistErr(err) {
			s.bll.AppCache().RemoveCache(kt, req.BizId, req.GetAppMeta().App)
		}

		return nil, status.Errorf(codes.Aborted, "get rkv failed, %s", err.Error())
	}

	kv := &pbfs.GetKvValueResp{
		KvType: rkv.KvType,
		Value:  rkv.Value,
	}

	return kv, nil
}

// isAppNotExistErr 检测app不存在错误, 有grpc，目前通过 msg 判断
// msg = rpc error: code = Code(4000005) desc = app %d not exist
func isAppNotExistErr(err error) bool {
	e := err.Error()

	if !strings.Contains(e, fmt.Sprintf("Code(%d)", errf.RecordNotFound)) {
		return false
	}

	if strings.Contains(e, "app") && strings.Contains(e, "not exist") {
		return true
	}

	return false
}

// ListApps 获取服务列表
func (s *Service) ListApps(ctx context.Context, req *pbfs.ListAppsReq) (*pbfs.ListAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp, err := s.bll.AppCache().ListApps(kt, &pbcs.ListAppsReq{BizId: req.BizId})
	if err != nil {
		return nil, err
	}

	credential := getCredential(ctx)
	apps := make([]*pbfs.App, 0, len(resp.Details))
	for _, d := range resp.Details {
		// 过滤无权限的 app
		if !credential.MatchApp(d.Spec.Name) {
			continue
		}

		// 客户端匹配
		if !matchPattern(d.Spec.Name, req.Match) {
			continue
		}

		apps = append(apps, &pbfs.App{
			Id:         d.Id,
			Name:       d.Spec.Name,
			ConfigType: d.Spec.ConfigType,
			Revision:   d.Revision,
		})
	}

	r := &pbfs.ListAppsResp{Apps: apps}
	return r, nil
}

// AsyncDownload 异步 p2p 下载，文件名为 sha256
func (s *Service) AsyncDownload(ctx context.Context, req *pbfs.AsyncDownloadReq) (*pbfs.AsyncDownloadResp, error) {
	kit := kit.FromGrpcContext(ctx)

	// TODO: 下发版本时包含是否支持 p2p 下载的标记，客户端根据标记决定是否通过 p2p 下载
	// TODO: 大文件下载会导致接口超时，需要优化：
	// 1. 服务端创建一个 task，返回 bscp 定义的 taskID 而不是 GSE taskID
	// 2. task 包含两部分：下载文件到本地、p2p 传输文件到客户端

	// 1. 鉴权
	credential := getCredential(ctx)
	app, err := s.bll.AppCache().GetMeta(kit, req.BizId, req.FileMeta.ConfigItemAttachment.AppId)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "get app %d metadata failed, %s",
			req.FileMeta.ConfigItemAttachment.AppId, err.Error())
	}
	if !credential.MatchApp(app.Name) {
		return nil, status.Errorf(codes.PermissionDenied, "not have app %s permission", app.Name)
	}

	if !credential.MatchConfigItem(app.Name, req.FileMeta.ConfigItemSpec.Path, req.FileMeta.ConfigItemSpec.Name) {
		return nil, status.Error(codes.PermissionDenied, "no permission download file")
	}

	gseConf := cc.FeedServer().GSE
	if !gseConf.Enabled {
		return nil, status.Error(codes.FailedPrecondition, "p2p download was disabled in server")
	}

	// 2. 获取服务端 agent_id 和 container_id
	serverAgentID, serverContainerID, err := s.getAsyncDownloadServerInfo(ctx, gseConf)
	if err != nil {
		return nil, err
	}

	// 3. 获取客户端信息，check 是否支持 p2p 下载
	clientAgentID, clientContainerID, err := s.getAsyncDownloadClientInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	// 4. 下载文件到本地
	sourceDir := path.Join(cc.FeedServer().GSE.SourceDir, strconv.Itoa(int(req.BizId)))
	if err = os.MkdirAll(sourceDir, os.ModePerm); err != nil {
		return nil, err
	}
	// filepath = source/{biz_id}/{sha256}
	signature := req.FileMeta.CommitSpec.Content.Signature
	serverFilePath := path.Join(sourceDir, signature)
	if err = s.checkAndDownloadFile(kit, serverFilePath, signature); err != nil {
		return nil, err
	}

	// 5. 创建文件传输任务
	taskID, err := gse.CreateTransferFileTask(ctx, serverAgentID, serverContainerID, sourceDir, gseConf.AgentUser,
		signature, clientAgentID, clientContainerID, req.FileDir, req.FileMeta.ConfigItemSpec.Permission.User)
	if err != nil {
		return nil, fmt.Errorf("create transfer file task failed, %s", err.Error())
	}

	// 6. 任务 ID 及其相关信息存入 Redis
	task := &types.AsyncDownloadTask{
		BizID:    req.BizId,
		AppID:    req.FileMeta.ConfigItemAttachment.AppId,
		TaskID:   taskID,
		FileName: req.FileMeta.ConfigItemSpec.Name,
		FilePath: req.FileMeta.ConfigItemSpec.Path,
	}
	if err := s.bll.AsyncDownload().CreateAsyncDownloadTask(kit, task); err != nil {
		return nil, err
	}

	r := &pbfs.AsyncDownloadResp{
		TaskId: taskID,
	}
	return r, nil
}

func (s *Service) getAsyncDownloadClientInfo(ctx context.Context, req *pbfs.AsyncDownloadReq) (
	agentID string, containerID string, err error) {
	if req.BkAgentId != "" {
		// target is node
		return req.BkAgentId, "", nil
	}
	// target is container
	if req.ClusterId == "" || req.PodId == "" {
		return "", "", status.Error(codes.InvalidArgument, "client agnet_id or (cluster_id and pod_id) is required")
	}
	pod, qErr := bcs.QueryPod(ctx, req.ClusterId, req.PodId)
	if qErr != nil {
		return "", "", qErr
	}
	for _, initContainer := range pod.Status.InitContainerStatuses {
		if initContainer.Name == req.ContainerName {
			containerID = tools.SplitContainerID(initContainer.ContainerID)
		}
	}
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == req.ContainerName {
			containerID = tools.SplitContainerID(container.ContainerID)
		}
	}
	if containerID == "" {
		return "", "", status.Errorf(codes.InvalidArgument, "client container %s not found in pod %s/%s",
			req.ContainerName, req.ClusterId, req.PodId)
	}
	node, qErr := bcs.QueryNode(ctx, req.ClusterId, pod.Spec.NodeName)
	if qErr != nil {
		return "", "", qErr
	}
	agentID = node.Labels[LabelKeyAgentID]
	if agentID == "" {
		return "", "", status.Errorf(codes.InvalidArgument, "bk-agent-id not found in client node %s/%s",
			req.ClusterId, pod.Spec.NodeName)
	}
	return agentID, containerID, nil
}

func (s *Service) getAsyncDownloadServerInfo(ctx context.Context, gseConf cc.GSE) (
	agentID string, containerID string, err error) {
	if gseConf.NodeAgentID != "" {
		// if serverAgentID configured, it measn feed server was deployed in binary mode, source is node
		agentID = gseConf.NodeAgentID
		return agentID, "", nil
	}
	// if serverAgentID not configured, it means feed server was deployed in container mode, source is container
	if gseConf.ClusterID == "" || gseConf.PodID == "" {
		return "", "", status.Error(codes.Internal, "server agent_id or (cluster_id and pod_id is required")
	}
	pod, qErr := bcs.QueryPod(ctx, gseConf.ClusterID, gseConf.PodID)
	if qErr != nil {
		return "", "", qErr
	}
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == gseConf.ContainerName {
			containerID = tools.SplitContainerID(container.ContainerID)
		}
	}
	if containerID == "" {
		return "", "", status.Errorf(codes.Internal, "server container %s not found in pod %s/%s",
			gseConf.ContainerName, gseConf.ClusterID, gseConf.PodID)
	}
	node, qErr := bcs.QueryNode(ctx, gseConf.ClusterID, pod.Spec.NodeName)
	if qErr != nil {
		return "", "", qErr
	}
	agentID = node.Labels[LabelKeyAgentID]
	if agentID == "" {
		return "", "", status.Errorf(codes.Internal, "bk-agent-id not found in server node %s/%s",
			gseConf.ClusterID, pod.Spec.NodeName)
	}
	return agentID, containerID, nil
}

func (s *Service) checkAndDownloadFile(kit *kit.Kit, filePath, signature string) error {
	// block until file download
	s.fileLock.Lock(filePath)
	defer s.fileLock.Unlock(filePath)
	if _, iErr := os.Stat(filePath); iErr != nil {
		if !os.IsNotExist(iErr) {
			return iErr
		}
		// not exists in feed server, download to local disk
		file, iErr := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if iErr != nil {
			return iErr
		}
		defer file.Close()

		reader, _, iErr := s.provider.Download(kit, signature)
		if iErr != nil {
			return iErr
		}
		defer reader.Close()
		if _, e := io.Copy(file, reader); e != nil {
			return e
		}
		if e := file.Sync(); e != nil {
			return e
		}
	}
	return nil
}

// AsyncDownloadStatus 查询异步 p2p 下载任务状态
func (s *Service) AsyncDownloadStatus(ctx context.Context, req *pbfs.AsyncDownloadStatusReq) (
	*pbfs.AsyncDownloadStatusResp, error) {
	kit := kit.FromGrpcContext(ctx)
	// 1.1 从 Redis 获取到任务对应的服务、文件信息，用token鉴权
	task, err := s.bll.AsyncDownload().GetAsyncDownloadTask(kit, req.BizId, req.TaskId)
	if err != nil {
		return nil, err
	}
	// 1. 鉴权
	credential := getCredential(ctx)
	app, err := s.bll.AppCache().GetMeta(kit, task.BizID, task.AppID)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "get app %d metadata failed, %s",
			task.AppID, err.Error())
	}
	if !credential.MatchApp(app.Name) {
		return nil, status.Errorf(codes.PermissionDenied, "have not app %s permission", app.Name)
	}

	if !credential.MatchConfigItem(app.Name, task.FilePath, task.FileName) {
		return nil, status.Error(codes.PermissionDenied, "no permission download file")
	}

	// 2. 获取GSE任务状态
	status, err := gse.TransferFileResult(ctx, task.TaskID)
	if err != nil {
		return nil, err
	}
	// TODO: 是否要保存任务开始时间，用以判断超时
	return &pbfs.AsyncDownloadStatusResp{
		Status: status,
	}, nil
}

// 匹配
func matchPattern(name string, match []string) bool {
	if len(match) == 0 {
		return true
	}

	for _, m := range match {
		ok, _ := path.Match(m, name)
		if ok {
			return true
		}
	}
	return false
}

func (s *Service) handleResourceUsageMetrics(bizID uint32, appName string, resource sfs.ResourceUsage) {
	s.mc.clientMaxCPUUsage.WithLabelValues(strconv.Itoa(int(bizID)), appName).Set(resource.CpuMaxUsage)
	s.mc.clientCurrentCPUUsage.WithLabelValues(strconv.Itoa(int(bizID)), appName).Set(resource.CpuUsage)
	s.mc.clientMaxMemUsage.WithLabelValues(strconv.Itoa(int(bizID)), appName).Set(float64(resource.MemoryMaxUsage))
	s.mc.clientCurrentMemUsage.WithLabelValues(strconv.Itoa(int(bizID)), appName).Set(float64(resource.MemoryUsage))

}
