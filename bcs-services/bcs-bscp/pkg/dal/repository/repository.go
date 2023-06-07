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

package repository

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbas "bscp.io/pkg/protocol/auth-server"
)

const (
	// repoRecordCacheExpiration repo created record cache expiration.
	RepoRecordCacheExpiration = time.Hour
)

// FileApiType file api type
type FileApiType interface {
	DownloadFile(w http.ResponseWriter, r *http.Request)
	FileMetadata(w http.ResponseWriter, r *http.Request)
	UploadFile(w http.ResponseWriter, r *http.Request)
}

// ObjectMetadata 文件元数据
type ObjectMetadata struct {
	ByteSize int64  `json:"byte_size"`
	Sha256   string `json:"sha256"`
}

// AuthResp http response with need apply permission.
type AuthResp struct {
	Code       int32               `json:"code"`
	Message    string              `json:"message"`
	Permission *pbas.IamPermission `json:"permission,omitempty"`
}

// GetBizIDAndAppID get biz_id and app_id from req path.
func GetBizIDAndAppID(kt *kit.Kit, req *http.Request) (uint32, uint32, error) {
	bizIDStr := chi.URLParam(req, "biz_id")
	bizID, err := strconv.ParseUint(bizIDStr, 10, 64)
	if err != nil {
		logs.Errorf("biz id parse uint failed, err: %v, rid: %s", err, kt.Rid)
		return 0, 0, err
	}

	if bizID == 0 {
		return 0, 0, errf.New(errf.InvalidParameter, "biz_id should > 0")
	}

	appIDStr := chi.URLParam(req, "app_id")
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		logs.Errorf("app id parse uint failed, err: %v, rid: %s", err, kt.Rid)
		return 0, 0, err
	}

	if appID == 0 {
		return 0, 0, errf.New(errf.InvalidParameter, "app_id should > 0")
	}

	return uint32(bizID), uint32(appID), nil
}
