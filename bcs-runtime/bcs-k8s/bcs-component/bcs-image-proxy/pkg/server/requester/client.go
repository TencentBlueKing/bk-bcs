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

package requester

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/apiclient"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/recorder"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils"
)

func commonHeaders(ctx context.Context) map[string]string {
	v := ctx.Value(logctx.RequestKey)
	if v == nil {
		return nil
	}
	return map[string]string{
		traceconst.RequestIDHeaderKey: v.(string),
	}
}

// GetManifest get manifest from master
func GetManifest(ctx context.Context, req *apiclient.GetManifestRequest) (string, string, error) {
	op := options.GlobalOptions()
	master := op.CurrentMaster()
	newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	body, err := utils.SendHTTPRequest(newCtx, &utils.HTTPRequest{
		Url:    fmt.Sprintf("http://%s%s", master, apiclient.CustomAPIGetManifest),
		Method: http.MethodPost,
		Body:   req,
		Header: commonHeaders(ctx),
	})
	if err != nil {
		return master, "", errors.Wrapf(err, "get manifest failed")
	}
	manifest := strings.TrimSpace(string(body))
	if manifest == "" {
		return master, manifest, errors.New("empty manifest")
	}
	return master, manifest, nil
}

// DownloadLayerFromMaster download layer from master
func DownloadLayerFromMaster(ctx context.Context, req *apiclient.CommonDownloadLayerRequest, digest string) (
	*apiclient.CommonDownloadLayerResponse, string, error) {
	op := options.GlobalOptions()
	master := op.CurrentMaster()
	recorder.GlobalRecorder().SendObjEvent(ctx, recorder.Normal,
		fmt.Sprintf("Starting get layer-info ‘%s’ from master %s", digest, master))
	body, err := utils.SendHTTPRequest(ctx, &utils.HTTPRequest{
		Url:    fmt.Sprintf("http://%s%s", master, apiclient.CustomAPIDownloadLayerFromMaster),
		Method: http.MethodPost,
		Body:   req,
	})
	if err != nil {
		return nil, master, errors.Wrapf(err, "get layer failed")
	}
	resp := new(apiclient.CommonDownloadLayerResponse)
	if err = json.Unmarshal(body, resp); err != nil {
		return nil, master, errors.Wrapf(err, "unmarshal resp body failed")
	}
	return resp, master, nil
}

// CheckStaticLayer check static layer exist
func CheckStaticLayer(ctx context.Context, target string, req *apiclient.CheckStaticLayerRequest) (
	*apiclient.CheckStaticLayerResponse, error) {
	op := options.GlobalOptions()
	body, err := utils.SendHTTPRequest(ctx, &utils.HTTPRequest{
		Url:    fmt.Sprintf("http://%s:%d%s", target, op.HTTPPort, apiclient.CustomAPICheckStaticLayer), // nolint
		Method: http.MethodGet,
		Body:   req,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "check static-layer failed")
	}
	resp := new(apiclient.CheckStaticLayerResponse)
	if err = json.Unmarshal(body, resp); err != nil {
		return nil, errors.Wrapf(err, "unmarshal resp body failed")
	}
	return resp, nil
}

// CheckOCILayer check oci layer exist
func CheckOCILayer(ctx context.Context, target string, req *apiclient.CheckOCILayerRequest) (
	*apiclient.CheckOCILayerResponse, error) {
	op := options.GlobalOptions()
	body, err := utils.SendHTTPRequest(ctx, &utils.HTTPRequest{
		Url:    fmt.Sprintf("http://%s:%d%s", target, op.HTTPPort, apiclient.CustomAPICheckOCILayer), // nolint
		Method: http.MethodGet,
		Body:   req,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "check oci-layer failed")
	}
	resp := new(apiclient.CheckOCILayerResponse)
	if err = json.Unmarshal(body, resp); err != nil {
		return nil, errors.Wrapf(err, "unmarshal resp body failed")
	}
	return resp, nil
}

// DownloadLayerFromNode download layer from node
func DownloadLayerFromNode(ctx context.Context, target string, req *apiclient.CommonDownloadLayerRequest) (
	*apiclient.CommonDownloadLayerResponse, error) {
	body, err := utils.SendHTTPRequest(ctx, &utils.HTTPRequest{
		Url:    fmt.Sprintf("http://%s%s", target, apiclient.CustomAPIDownloadLayerFromNode), // nolint
		Method: http.MethodGet,
		Body:   req,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "download layer from node failed")
	}
	resp := new(apiclient.CommonDownloadLayerResponse)
	if err = json.Unmarshal(body, resp); err != nil {
		return nil, errors.Wrapf(err, "unmarshal resp body failed")
	}
	return resp, nil
}
