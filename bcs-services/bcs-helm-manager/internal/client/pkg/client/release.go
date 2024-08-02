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

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	// urlReleaseList
	urlReleaseList = "/release/%s" // nolint
	// urlReleaseListV1
	urlReleaseListV1 = "/projects/%s/clusters/%s/releases"
	// urlReleaseInstall
	urlReleaseInstall = "/release/%s/%s/%s/install" // nolint
	// urlReleaseUninstall
	urlReleaseUninstall = "/release/%s/%s/%s/uninstall" // nolint
	// urlReleaseUpgrade
	urlReleaseUpgrade = "/release/%s/%s/%s/upgrade" // nolint
	// urlReleaseRollback
	urlReleaseRollback = "/release/%s/%s/%s/rollback" // nolint
	// urlReleaseDetailV1Get
	urlReleaseDetailV1Get = "/projects/%s/clusters/%s/namespaces/%s/releases/%s"
	// urlReleaseDetailV1Install
	urlReleaseDetailV1Install = "/projects/%s/clusters/%s/namespaces/%s/releases/%s"
	// urlReleaseDetailV1Uninstall
	urlReleaseDetailV1Uninstall = "/projects/%s/clusters/%s/namespaces/%s/releases/%s"
	// urlReleaseDetailV1Upgrade
	urlReleaseDetailV1Upgrade = "/projects/%s/clusters/%s/namespaces/%s/releases/%s"
	// urlReleaseDetailV1Rollback
	urlReleaseDetailV1Rollback = "/projects/%s/clusters/%s/namespaces/%s/releases/%s/rollback"
	// urlReleaseHistoryGet
	urlReleaseHistoryGet = "/projects/%s/clusters/%s/namespaces/%s/releases/%s/history"
	// urlReleasePreview
	urlReleasePreview = "/projects/%s/clusters/%s/namespaces/%s/releases/%s/preview"
	// urlReleaseManifestGet
	urlReleaseManifestGet = "/projects/%s/clusters/%s/namespaces/%s/releases/%s/revisions/%d/manifest"
)

// Release return a pkg.ReleaseClient instance
func (c *Client) Release() pkg.ReleaseClient {
	return &release{Client: c}
}

type release struct {
	*Client
}

// GetReleaseDetail get release detail
func (rl *release) GetReleaseDetail(ctx context.Context, req *helmmanager.GetReleaseDetailV1Req) (
	*helmmanager.ReleaseDetail, error) {
	if req == nil {
		return nil, fmt.Errorf("get release detail request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return nil, fmt.Errorf("release projectCode can not be empty")
	}

	clusterID := req.GetClusterID()
	if clusterID == "" {
		return nil, fmt.Errorf("release clusterID can not be empty")
	}

	namespace := req.GetNamespace()
	if namespace == "" {
		return nil, fmt.Errorf("release namespace can not be empty")
	}

	name := req.GetName()
	if name == "" {
		return nil, fmt.Errorf("release name can not be empty")
	}
	resp, err := rl.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseDetailV1Get, projectCode, clusterID, namespace, name),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.GetReleaseDetailV1Resp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("get release detail get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// List release
func (rl *release) List(ctx context.Context, req *helmmanager.ListReleaseV1Req) (*helmmanager.ReleaseListData, error) {
	if req == nil {
		return nil, fmt.Errorf("list release request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return nil, fmt.Errorf("release projectCode can not be empty")
	}

	clusterID := req.GetClusterID()
	if clusterID == "" {
		return nil, fmt.Errorf("release clusterID can not be empty")
	}

	resp, err := rl.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseListV1, projectCode, clusterID)+"?"+rl.listReleaseQuery(req).Encode(),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.ListReleaseV1Resp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("list release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// listReleaseQuery
func (rl *release) listReleaseQuery(req *helmmanager.ListReleaseV1Req) url.Values {
	query := url.Values{}
	if req.Page != nil {
		query.Set("page", strconv.FormatInt(int64(req.GetPage()), 10))
	}
	if req.Size != nil {
		query.Set("size", strconv.FormatInt(int64(req.GetSize()), 10))
	}
	if req.Namespace != nil {
		query.Set("namespace", req.GetNamespace())
	}
	if req.Name != nil {
		query.Set("name", req.GetName())
	}
	return query
}

// Install release
func (rl *release) Install(ctx context.Context, req *helmmanager.InstallReleaseV1Req) error {
	if req == nil {
		return fmt.Errorf("install release request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return fmt.Errorf("install release projectCode can not be empty")
	}
	clusterID := req.GetClusterID()
	if clusterID == "" {
		return fmt.Errorf("install release clusterID can not be empty")
	}
	namespace := req.GetNamespace()
	if namespace == "" {
		return fmt.Errorf("install release namespace can not be empty")
	}
	name := req.GetName()
	if name == "" {
		return fmt.Errorf("install release name can not be empty")
	}

	data, _ := json.Marshal(req)
	resp, err := rl.post(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseDetailV1Install, projectCode, clusterID, namespace, name),
		nil,
		data,
	)
	if err != nil {
		return err
	}

	var r helmmanager.InstallReleaseV1Resp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("install release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}

// Uninstall release
func (rl *release) Uninstall(ctx context.Context, req *helmmanager.UninstallReleaseV1Req) error {
	if req == nil {
		return fmt.Errorf("uninstall release request is empty")
	}
	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return fmt.Errorf("uninstall release projectCode can not be empty")
	}
	clusterID := req.GetClusterID()
	if clusterID == "" {
		return fmt.Errorf("uninstall release clusterID can not be empty")
	}
	namespace := req.GetNamespace()
	if namespace == "" {
		return fmt.Errorf("uninstall release namespace can not be empty")
	}
	name := req.GetName()
	if name == "" {
		return fmt.Errorf("uninstall release name can not be empty")
	}

	data, _ := json.Marshal(req)

	resp, err := rl.delete(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseDetailV1Uninstall, projectCode, clusterID, namespace, name),
		nil,
		data,
	)
	if err != nil {
		return err
	}

	var r helmmanager.UninstallReleaseV1Resp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("uninstall release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}

// Upgrade release
func (rl *release) Upgrade(ctx context.Context, req *helmmanager.UpgradeReleaseV1Req) error {
	if req == nil {
		return fmt.Errorf("upgrade release request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return fmt.Errorf("upgrade release projectCode can not be empty")
	}
	clusterID := req.GetClusterID()
	if clusterID == "" {
		return fmt.Errorf("upgrade release clusterID can not be empty")
	}
	namespace := req.GetNamespace()
	if namespace == "" {
		return fmt.Errorf("upgrade release namespace can not be empty")
	}
	name := req.GetName()
	if name == "" {
		return fmt.Errorf("upgrade release name can not be empty")
	}

	data, _ := json.Marshal(req)

	resp, err := rl.put(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseDetailV1Upgrade, projectCode, clusterID, namespace, name),
		nil,
		data,
	)
	if err != nil {
		return err
	}

	var r helmmanager.UpgradeReleaseV1Resp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("upgrade release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}

// Rollback release
func (rl *release) Rollback(ctx context.Context, req *helmmanager.RollbackReleaseV1Req) error {
	if req == nil {
		return fmt.Errorf("rollback release request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return fmt.Errorf("upgrade release projectCode can not be empty")
	}
	clusterID := req.GetClusterID()
	if clusterID == "" {
		return fmt.Errorf("rollback release clusterID can not be empty")
	}
	namespace := req.GetNamespace()
	if namespace == "" {
		return fmt.Errorf("rollback release namespace can not be empty")
	}
	name := req.GetName()
	if name == "" {
		return fmt.Errorf("rollback release name can not be empty")
	}

	data, _ := json.Marshal(req)

	resp, err := rl.put(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseDetailV1Rollback, projectCode, clusterID, namespace, name),
		nil,
		data,
	)
	if err != nil {
		return err
	}

	var r helmmanager.RollbackReleaseV1Resp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("rollback release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}

// GetReleaseHistory get release history
func (rl *release) GetReleaseHistory(ctx context.Context, req *helmmanager.GetReleaseHistoryReq) (
	[]*helmmanager.ReleaseHistory, error) {
	if req == nil {
		return nil, fmt.Errorf("get release history request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return nil, fmt.Errorf("get release history projectCode can not be empty")
	}
	clusterID := req.GetClusterID()
	if clusterID == "" {
		return nil, fmt.Errorf("get release history clusterID can not be empty")
	}
	namespace := req.GetNamespace()
	if namespace == "" {
		return nil, fmt.Errorf("get release history namespace can not be empty")
	}
	name := req.GetName()
	if name == "" {
		return nil, fmt.Errorf("get release history name can not be empty")
	}

	var data []byte
	data, _ = json.Marshal(req)

	resp, err := rl.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseHistoryGet, projectCode, clusterID, namespace, name),
		nil,
		data,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.GetReleaseHistoryResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("rollback release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// ReleasePreview release preview
func (rl *release) ReleasePreview(ctx context.Context, req *helmmanager.ReleasePreviewReq) (
	*helmmanager.ReleasePreview, error) {
	if req == nil {
		return nil, fmt.Errorf("release preview request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return nil, fmt.Errorf("release preview projectCode can not be empty")
	}
	clusterID := req.GetClusterID()
	if clusterID == "" {
		return nil, fmt.Errorf("release preview clusterID can not be empty")
	}
	namespace := req.GetNamespace()
	if namespace == "" {
		return nil, fmt.Errorf("release preview namespace can not be empty")
	}
	name := req.GetName()
	if name == "" {
		return nil, fmt.Errorf("release preview name can not be empty")
	}

	data, _ := json.Marshal(req)
	resp, err := rl.post(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleasePreview, projectCode, clusterID, namespace, name),
		nil,
		data,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.ReleasePreviewResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("release preview get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// GetReleaseManifest get release manifest
func (rl *release) GetReleaseManifest(ctx context.Context, req *helmmanager.GetReleaseManifestReq) (
	map[string]*helmmanager.FileContent, error) {
	if req == nil {
		return nil, fmt.Errorf("get release manifest request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return nil, fmt.Errorf("get release manifest projectCode can not be empty")
	}
	clusterID := req.GetClusterID()
	if clusterID == "" {
		return nil, fmt.Errorf("get release manifest clusterID can not be empty")
	}
	namespace := req.GetNamespace()
	if namespace == "" {
		return nil, fmt.Errorf("get release manifest namespace can not be empty")
	}
	name := req.GetName()
	if name == "" {
		return nil, fmt.Errorf("get release manifest name can not be empty")
	}
	// 上层做了校验
	revision := *req.Revision

	data, _ := json.Marshal(req)
	resp, err := rl.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseManifestGet, projectCode, clusterID, namespace, name, revision),
		nil,
		data,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.GetReleaseManifestResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("get release manifest get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}
