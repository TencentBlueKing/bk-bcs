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

package release

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chartutil"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewReleasePreviewAction return a new ReleasePreviewAction instance
func NewReleasePreviewAction(model store.HelmManagerModel, platform repo.Platform,
	releaseHandler release.Handler) *ReleasePreviewAction {
	return &ReleasePreviewAction{
		model:          model,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// ReleasePreviewAction provides the action to get release preview
type ReleasePreviewAction struct { // nolint
	ctx context.Context

	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	createBy string

	req  *helmmanager.ReleasePreviewReq
	resp *helmmanager.ReleasePreviewResp
}

// Handle the release preview process
func (r *ReleasePreviewAction) Handle(ctx context.Context,
	req *helmmanager.ReleasePreviewReq, resp *helmmanager.ReleasePreviewResp) error {
	r.ctx = ctx
	r.req = req
	r.resp = resp

	if err := r.req.Validate(); err != nil {
		blog.Errorf("get release preview failed, invalid request, %s, param: %v", err.Error(), r.req)
		r.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	old, err := r.model.GetRelease(r.ctx, r.req.GetClusterID(), r.req.GetNamespace(), r.req.GetName())
	if err == nil {
		r.createBy = old.CreateBy
	} else {
		r.createBy = auth.GetUserFromCtx(r.ctx)
	}

	preview, err := r.getReleasePreview()
	if err != nil {
		blog.Errorf("get release preview failed, %s", err.Error())
		r.setResp(common.ErrHelmManagerPreviewActionFailed, err.Error(), nil)
		return nil
	}
	r.setResp(common.ErrHelmManagerSuccess, "ok", preview)
	return nil
}

func (r *ReleasePreviewAction) getReleasePreview() (*helmmanager.ReleasePreview, error) {
	// get manifest from helm
	currentRelease, err := r.releaseHandler.Cluster(r.req.GetClusterID()).Get(r.ctx, release.GetOption{
		Namespace: r.req.GetNamespace(), Name: r.req.GetName()})
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		return nil, fmt.Errorf("get current releasefailed, err %s", err.Error())
	}

	// revision 之间对比，用于回滚
	if r.req.GetRevision() != 0 {
		newRelease, err := r.releaseHandler.Cluster(r.req.GetClusterID()).Get(r.ctx, release.GetOption{ // nolint
			Namespace: r.req.GetNamespace(), Name: r.req.GetName(), Revision: int(r.req.GetRevision())})
		if err != nil {
			return nil, fmt.Errorf("get release revision %d failed, err %s", r.req.GetRevision(), err.Error())
		}
		return r.GenerateReleasePreview(currentRelease.Transfer2Release(), newRelease.Transfer2Release())
	}

	// helm template, get new manifest
	username := auth.GetUserFromCtx(r.ctx)
	projectCode := contextx.GetProjectCodeFromCtx(r.ctx)
	contents, err := getChartContent(r.model, r.platform, projectCode, r.req.GetRepository(),
		r.req.GetChart(), r.req.GetVersion())
	if err != nil {
		return nil, fmt.Errorf("get release preview, get contents failed, %s", err.Error())
	}
	var reuse bool
	// 过滤掉不支持的参数并判断是否需要--reuse-values
	reuse, r.req.Args = filtArgs(r.req.GetArgs())
	// 支持--reuse-values参数
	r.req.Values, err = reuseValues(reuse, currentRelease, r.req.GetValues())
	if err != nil {
		return nil, fmt.Errorf("reuse values failed, %s", err.Error())
	}
	result, err := release.InstallRelease(r.releaseHandler, contextx.GetProjectIDFromCtx(r.ctx), projectCode,
		r.req.GetClusterID(), r.req.GetName(), r.req.GetNamespace(), r.req.GetChart(), r.req.GetVersion(),
		r.createBy, username, r.req.GetArgs(), nil, contents, r.req.GetValues(), true, true, true)
	if err != nil {
		return nil, fmt.Errorf("get release preview, get helm template failed, %s", err.Error())
	}
	newRelease := result.Release

	return r.GenerateReleasePreview(currentRelease.Transfer2Release(), newRelease)
}

// 过滤掉不支持的参数并判断是否需要--reuse-values
func filtArgs(args []string) (bool, []string) {
	var reuse bool
	// 黑名单参数
	filtContent := map[string]struct{}{
		"--force":        {},
		"--reuse-values": {},
	}
	result := []string{}
	for _, value := range args {
		s := strings.Split(value, "=")
		if value == "--reuse-values" {
			reuse = true
		}
		if len(s) > 0 {
			if _, ok := filtContent[s[0]]; ok {
				continue
			}
			result = append(result, value)
		}
	}
	return reuse, result
}

// GenerateReleasePreview generate release preview
func (r *ReleasePreviewAction) GenerateReleasePreview(oldRelease,
	newRelease *helmrelease.Release) (*helmmanager.ReleasePreview, error) {
	preview := &helmmanager.ReleasePreview{
		NewContent: common.GetStringP(""),
		OldContent: common.GetStringP(""),
	}
	if newRelease == nil {
		return preview, nil
	}

	manifest := newRelease.Manifest
	for _, v := range newRelease.Hooks {
		manifest += "\n---\n" + v.Manifest
	}
	preview.NewContent = removeCustomAnnotations(manifest)
	if oldRelease != nil {
		oldManifest := oldRelease.Manifest
		for _, v := range oldRelease.Hooks {
			oldManifest += "\n---\n" + v.Manifest
		}
		preview.OldContent = removeCustomAnnotations(oldManifest)
	}

	// get contents
	var err error
	preview.NewContents, err = generateFileContents(manifest)
	if err != nil {
		return nil, err
	}
	preview.OldContents, err = generateFileContents(preview.GetOldContent())
	if err != nil {
		return nil, err
	}
	return preview, nil
}

func (r *ReleasePreviewAction) setResp(err common.HelmManagerError, message string, rp *helmmanager.ReleasePreview) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	r.resp.Code = &code
	r.resp.Message = &msg
	r.resp.Result = err.OK()
	r.resp.Data = rp
}

// reuseValues copies values from the current release to a new release
// if there is a new value, overwrite the current value
func reuseValues(reuseValues bool, release *release.Release, values []string) ([]string, error) {
	if release == nil {
		return values, nil
	}

	if reuseValues {
		// old value
		var oldVar map[string]interface{}
		err := yaml.Unmarshal([]byte(release.Values), &oldVar)
		if err != nil {
			return nil, err
		}

		// new value, if there is a new value, overwrite the current value
		newVar := make(map[string]interface{}, 0)
		for _, data := range values {
			var temp map[string]interface{}
			err = yaml.Unmarshal([]byte(data), &temp)
			if err != nil {
				return nil, err
			}
			newVar = chartutil.CoalesceTables(temp, newVar)
		}
		// if there is a new value, overwrite the current value
		newVar = chartutil.CoalesceTables(newVar, oldVar)
		b, err := yaml.Marshal(newVar)
		if err != nil {
			return nil, err
		}
		return []string{string(b)}, nil
	}
	return values, nil
}
