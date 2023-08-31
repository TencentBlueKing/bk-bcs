/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"fmt"

	"bscp.io/cmd/feed-server/bll/types"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbfs "bscp.io/pkg/protocol/feed-server"
	"bscp.io/pkg/runtime/jsoni"
	sfs "bscp.io/pkg/sf-share"
	"bscp.io/pkg/tools"

	prm "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	ra := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Sidecar, Action: meta.Access}, BizID: hm.Spec.BizId}
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
		},
	}

	payloadBytes, err := jsoni.Marshal(payload)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "encode payload failed, err: %v", err)
	}

	return &pbfs.HandshakeResp{ApiVersion: sfs.CurrentAPIVersion, Payload: payloadBytes}, nil
}

// getAppReload get app reload option.
func (s *Service) getAppReload(kt *kit.Kit, bizID uint32, appIDs []uint32) (map[uint32]*sfs.Reload, error) {
	appReloadList := make(map[uint32]*sfs.Reload)

	for _, appID := range appIDs {
		appMeta, err := s.bll.AppCache().GetMeta(kt, bizID, appID)
		if err != nil {
			return nil, err
		}

		appReloadList[appID] = &sfs.Reload{
			ReloadType: appMeta.Reload.ReloadType,
			FileReloadSpec: &sfs.FileReloadSpec{
				ReloadFilePath: appMeta.Reload.FileReloadSpec.ReloadFilePath,
			},
		}
	}

	return appReloadList, nil
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

	ra := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Sidecar, Action: meta.Access}, BizID: im.Meta.BizID}
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
		msg += fmt.Sprintf("biz: %d, app: %s, uid: %s, ", payload.BizID, one.App, one.Uid)
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
func (s *Service) Messaging(ctx context.Context, msg *pbfs.MessagingMeta) (*pbfs.MessagingResp, error) {
	im, err := sfs.ParseFeedIncomingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ra := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Sidecar, Action: meta.Access}, BizID: im.Meta.BizID}
	authorized, err := s.bll.Auth().Authorize(im.Kit, ra)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, err.Error())
	}

	if !authorized {
		return nil, status.Errorf(codes.PermissionDenied, "no permission to access bscp server")
	}

	logs.V(3).Infof("receive %d biz %s sidecar %s message, payload: %s, rid: %s", im.Meta.BizID, im.Meta.Fingerprint,
		sfs.MessagingType(msg.Type).String(), msg.Payload, msg.Rid)
	return new(pbfs.MessagingResp), nil
}

// PullAppFileMeta pull an app's latest release metadata only when the app's configures is file type.
func (s *Service) PullAppFileMeta(ctx context.Context, req *pbfs.PullAppFileMetaReq) (
	*pbfs.PullAppFileMetaResp, error) {

	// check if the sidecar's version can be accepted.
	if !sfs.IsAPIVersionMatch(req.ApiVersion) {
		return nil, status.Error(codes.InvalidArgument, "sdk's api version is too low, should be upgraded")
	}

	im, err := sfs.ParseFeedIncomingContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ra := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Sidecar, Action: meta.Access}, BizID: im.Meta.BizID}
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

	appID, err := s.bll.AppCache().GetAppID(im.Kit, req.BizId, req.AppMeta.App)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "get app id failed, %s", err.Error())
	}
	meta := &types.AppInstanceMeta{
		BizID:  req.BizId,
		App:    req.AppMeta.App,
		AppID:  appID,
		Uid:    req.AppMeta.Uid,
		Labels: req.AppMeta.Labels,
	}

	cancel := im.Kit.CtxWithTimeoutMS(1500)
	defer cancel()

	metas, err := s.bll.Release().ListAppLatestReleaseMeta(im.Kit, meta)
	if err != nil {
		return nil, err
	}

	fileMetas := make([]*pbfs.FileMeta, 0, len(metas.ConfigItems))
	for _, ci := range metas.ConfigItems {
		if req.Key != "" && !tools.MatchConfigItem(req.Key, ci.ConfigItemSpec.Path, ci.ConfigItemSpec.Name) {
			continue
		}
		app, err := s.bll.AppCache().GetMeta(im.Kit, req.BizId, ci.ConfigItemAttachment.AppId)
		if err != nil {
			return nil, status.Errorf(codes.Aborted, "get app meta failed, %s", err.Error())
		}
		if match, err := s.bll.Auth().CanMatchCI(im.Kit, req.BizId, app.Name, req.Token, ci.ConfigItemSpec); err != nil || !match {
			logs.Errorf("no permission to access config item %d, err: %v", ci.RciId, err)
			return nil, status.Errorf(codes.PermissionDenied, "no permission to access config item %d", ci.RciId)
		}
		fileMetas = append(fileMetas, &pbfs.FileMeta{
			Id:                   ci.RciId,
			CommitId:             ci.CommitID,
			CommitSpec:           ci.CommitSpec,
			ConfigItemSpec:       ci.ConfigItemSpec,
			ConfigItemAttachment: ci.ConfigItemAttachment,
			RepositorySpec: &pbfs.RepositorySpec{
				Path: ci.RepositorySpec.Path,
			},
		})
	}
	resp := &pbfs.PullAppFileMetaResp{
		ReleaseId: metas.ReleaseId,
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
		return nil, status.Error(codes.InvalidArgument, "sdk's api version is too low, should be upgraded")
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
	match, err := s.bll.Auth().CanMatchCI(im.Kit, req.BizId, app.Name, req.Token, req.FileMeta.ConfigItemSpec)
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
