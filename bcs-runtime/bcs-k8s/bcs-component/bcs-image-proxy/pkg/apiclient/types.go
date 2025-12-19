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

package apiclient

import (
	"fmt"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/google/uuid"

	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/internal/logctx"
)

const (
	CustomAPIGetManifest      = "/custom_api/get_manifest"
	CustomAPIGetLayerInfo     = "/custom_api/get_layer_info"
	CustomAPICheckStaticLayer = "/custom_api/check_static_layer"
	CustomAPICheckOCILayer    = "/custom_api/check_oci_layer"

	CustomAPIDownloadLayerFromMaster = "/custom_api/download_layer_from_master"
	CustomAPIDownloadLayerFromNode   = "/custom_api/download_layer_from_node"
	CustomAPITransferLayerTCP        = "/custom_api/transfer_layer_tcp"

	CustomAPIRecorder      = "/custom_api/recorder"
	CustomAPITorrentStatus = "/custom_api/torrent_status"

	RegistryAuthenticateHeader = "WWW-Authenticate"
)

// SetContext set the request context
func SetContext(req *http.Request) *http.Request {
	requestID := req.Header.Get(traceconst.RequestIDHeaderKey)
	if requestID == "" {
		requestID = uuid.New().String()
	}
	reqCtx := logctx.SetRequestID(req.Context(), requestID)
	// rw.Header().Add(traceconst.RequestIDHeaderKey, requestID)
	return req.WithContext(reqCtx)
}

var (
	// TwentyMB 20MB
	TwentyMB int64 = 20971520
	// TwoHundredMB 200MB
	TwoHundredMB int64 = 209715200
)

// HostLayerDB defines the host oci layers
type HostLayerDB struct {
	HostIP           string            `json:"hostIP"`
	DockerdLayers    map[string]string `json:"dockerdLayers"`
	ContainerdLayers map[string]string `json:"containerdLayers"`
}

// GetManifestRequest defines the request of GetManifest
type GetManifestRequest struct {
	OriginalHost string `json:"originalHost"`
	ManifestUrl  string `json:"manifestUrl"`
	Repo         string `json:"repo"`
	Tag          string `json:"tag"`
	BearerToken  string `json:"bearerToken"`
}

// GetLayerInfoRequest defines the request of GetLayerInfo
type GetLayerInfoRequest struct {
	OriginalHost  string `json:"originalHost"`
	LayerUrl      string `json:"layerUrl"`
	BearerToken   string `json:"bearerToken"`
	ContentLength int64  `json:"fileSize"`
}

// GetLayerInfoResponse defines the response of GetLayerInfo
type GetLayerInfoResponse struct {
	ContentLength int64 `json:"contentLength"`
}

// CheckStaticLayerRequest defines the request of check static layer
type CheckStaticLayerRequest struct {
	Digest                string `json:"digest"`
	LayerPath             string `json:"path"`
	ExpectedContentLength int64  `json:"expectedContentLength"`
}

// CheckStaticLayerResponse defines the response of CheckStaticLayer
type CheckStaticLayerResponse struct {
	Located       string `json:"located"`
	LayerPath     string `json:"layerPath"`
	TorrentBase64 string `json:"torrentBase64"`
	FileSize      int64  `json:"fileSize"`
}

// CheckOCILayerRequest defines the request of CheckOCILayer
type CheckOCILayerRequest struct {
	Digest  string `json:"digest"`
	OCIType string `json:"ociType"`
}

// CheckOCILayerResponse defines the response of CheckOCILayer
type CheckOCILayerResponse struct {
	Located   string `json:"located"`
	LayerPath string `json:"layerPath"`
	FileSize  int64  `json:"fileSize"`
}

// CommonDownloadLayerRequest defines the request of download layer
type CommonDownloadLayerRequest struct {
	OriginalHost string `json:"originalHost"`
	LayerUrl     string `json:"layerUrl"`
	BearerToken  string `json:"bearerToken"`
}

// CommonDownloadLayerResponse defines the response of download layer
type CommonDownloadLayerResponse struct {
	TorrentBase64 string `json:"torrentBase64"`
	Located       string `json:"located"`
	FilePath      string `json:"filePath"`
	FileSize      int64  `json:"fileSize"`
}

// String defines the download layer response string
func (resp *CommonDownloadLayerResponse) String() string {
	if resp.TorrentBase64 != "" {
		return fmt.Sprintf(`{"torrent": "(too long)", "located": "%s", "filePath": "%s", "fileSize": %s}`,
			resp.Located, resp.FilePath, humanize.Bytes(uint64(resp.FileSize)))
	}
	return fmt.Sprintf(`{"located": "%s", "filePath": "%s", "fileSize": %s}`,
		resp.Located, resp.FilePath, humanize.Bytes(uint64(resp.FileSize)))
}
