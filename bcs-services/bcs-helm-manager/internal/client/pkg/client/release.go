/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	urlReleaseList      = "/helmmanager/v1/release/%s"
	urlReleaseInstall   = "/helmmanager/v1/release/%s/%s/%s/install"
	urlReleaseUninstall = "/helmmanager/v1/release/%s/%s/%s/uninstall"
	urlReleaseUpgrade   = "/helmmanager/v1/release/%s/%s/%s/upgrade"
	urlReleaseRollback  = "/helmmanager/v1/release/%s/%s/%s/rollback"
)

// Release return a pkg.ReleaseClient instance
func (c *Client) Release() pkg.ReleaseClient {
	return &release{Client: c}
}

type release struct {
	*Client
}

// List release
func (rl *release) List(ctx context.Context, req *helmmanager.ListReleaseReq) (*helmmanager.ReleaseListData, error) {
	if req == nil {
		return nil, fmt.Errorf("list release request is empty")
	}

	clusterID := req.GetClusterID()
	if clusterID == "" {
		return nil, fmt.Errorf("release clusterID can not be empty")
	}

	resp, err := rl.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseList, clusterID)+"?"+rl.listReleaseQuery(req).Encode(),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.ListReleaseResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("list release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

func (rl *release) listReleaseQuery(req *helmmanager.ListReleaseReq) url.Values {
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
func (rl *release) Install(ctx context.Context, req *helmmanager.InstallReleaseReq) (
	*helmmanager.ReleaseDetail, error) {
	if req == nil {
		return nil, fmt.Errorf("install release request is empty")
	}

	req.Operator = common.GetStringP(rl.conf.Operator)
	clusterID := req.GetClusterID()
	if clusterID == "" {
		return nil, fmt.Errorf("install release clusterID can not be empty")
	}
	namespace := req.GetNamespace()
	if namespace == "" {
		return nil, fmt.Errorf("install release namespace can not be empty")
	}
	name := req.GetName()
	if name == "" {
		return nil, fmt.Errorf("install release name can not be empty")
	}

	var data []byte
	_ = codec.EncJson(req, &data)

	resp, err := rl.post(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseInstall, clusterID, namespace, name),
		nil,
		data,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.InstallReleaseResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("install release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// Uninstall release
func (rl *release) Uninstall(ctx context.Context, req *helmmanager.UninstallReleaseReq) error {
	if req == nil {
		return fmt.Errorf("uninstall release request is empty")
	}

	req.Operator = common.GetStringP(rl.conf.Operator)
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

	var data []byte
	_ = codec.EncJson(req, &data)

	resp, err := rl.post(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseUninstall, clusterID, namespace, name),
		nil,
		data,
	)
	if err != nil {
		return err
	}

	var r helmmanager.UninstallReleaseResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("uninstall release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}

// Upgrade release
func (rl *release) Upgrade(ctx context.Context, req *helmmanager.UpgradeReleaseReq) (
	*helmmanager.ReleaseDetail, error) {
	if req == nil {
		return nil, fmt.Errorf("upgrade release request is empty")
	}

	req.Operator = common.GetStringP(rl.conf.Operator)
	clusterID := req.GetClusterID()
	if clusterID == "" {
		return nil, fmt.Errorf("upgrade release clusterID can not be empty")
	}
	namespace := req.GetNamespace()
	if namespace == "" {
		return nil, fmt.Errorf("upgrade release namespace can not be empty")
	}
	name := req.GetName()
	if name == "" {
		return nil, fmt.Errorf("upgrade release name can not be empty")
	}

	var data []byte
	_ = codec.EncJson(req, &data)

	resp, err := rl.post(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseUpgrade, clusterID, namespace, name),
		nil,
		data,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.UpgradeReleaseResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("upgrade release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// Rollback release
func (rl *release) Rollback(ctx context.Context, req *helmmanager.RollbackReleaseReq) error {
	if req == nil {
		return fmt.Errorf("rollback release request is empty")
	}

	req.Operator = common.GetStringP(rl.conf.Operator)
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

	var data []byte
	_ = codec.EncJson(req, &data)

	resp, err := rl.post(
		ctx,
		urlPrefix+fmt.Sprintf(urlReleaseRollback, clusterID, namespace, name),
		nil,
		data,
	)
	if err != nil {
		return err
	}

	var r helmmanager.RollbackReleaseResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("rollback release get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}
