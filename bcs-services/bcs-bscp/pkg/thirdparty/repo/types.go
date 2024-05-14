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

package repo

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
)

const (
	// version the current bscp uses the version of the repo scheme, including
	// project, warehouse, Node and other names to design and store content
	version = "v1"

	// nodeFrontPath is node full path's front path, reserved for future expansion.
	nodeFrontPath = "/file/"
)

const (
	// RepositoryType is repository type.
	RepositoryType = "GENERIC"
	// CategoryType is bscp repository category.
	CategoryType = "LOCAL"
	// RepositoryCfgType repository configuration type.
	RepositoryCfgType = "local"
)

// header key.
const (
	// HeaderKeyMETA file metadata in the format base64 (key1=value1&key2=value2). key is case sensitive.
	HeaderKeyMETA = "X-BKREPO-META"
	// HeaderKeyUID is repo uid header key.
	HeaderKeyUID = "X-BKREPO-UID"
	// HeaderKeySHA256 is repo file sha256 header key.
	HeaderKeySHA256 = "X-BKREPO-SHA256"
	// HeaderKeyOverwrite is repo upload overwrite flag header key.
	HeaderKeyOverwrite = "X-BKREPO-OVERWRITE"
)

// error code.
const (
	// errCodeRepoAlreadyExist repo repository already exist.
	errCodeRepoAlreadyExist = 251007
	// ErrCodeNodeAlreadyExist repo node already exist.
	ErrCodeNodeAlreadyExist = 251012
	// ErrCodeNodeNotExist repo node not exist.
	ErrCodeNodeNotExist = 251010
)

// Config repo config.
type Config struct {
	// Addrs repo server addresses.
	Addrs []string
	// Token repo auth token.
	Token string
	// Project repo bscp project name.
	Project string
	// User repo bscp project admin user name.
	User string
}

// repoDiscovery repo server discovery.
type repoDiscovery struct {
	servers []string
	index   int
	sync.Mutex
}

// GetServers get repo server host.
func (s *repoDiscovery) GetServers() ([]string, error) {
	s.Lock()
	defer s.Unlock()

	num := len(s.servers)
	if num == 0 {
		return []string{}, errors.New("repo is no server can be used")
	}

	if s.index < num-1 {
		s.index++
		return append(s.servers[s.index-1:], s.servers[:s.index-1]...), nil
	}

	s.index = 0
	return append(s.servers[num-1:], s.servers[:num-1]...), nil
}

// CreateRepoReq is repo create repo request struct.
type CreateRepoReq struct {
	// ProjectID is bscp project name in repo.
	ProjectID string `json:"projectId"`
	// Name is name of new repo.
	Name string `json:"name"`
	// Type is type of new repo(GENERIC).
	Type string `json:"type"`
	// Category is category type of new repo(LOCAL).
	Category string `json:"category"`
	// Public is repo public flag, default false not public to download.
	Public bool `json:"public"`
	// Configuration is configuration for new repo.
	Configuration Configuration `json:"configuration"`
	// Description is repo memo description.
	Description string `json:"description"`
}

// UploadFileReq is repo upload file request struct
type UploadFileReq struct {
	// BizID is business ID
	BizID uint32 `json:"bizID"`
	// AppID is application ID
	AppID uint32 `json:"appID"`
	// Content is base64 encoded content of file
	Content string `json:"content"`
}

// UploadResp upload response
// Docs https://github.com/TencentBlueKing/bk-repo/blob/master/docs/apidoc/generic/simple.md
type UploadResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    *UploadData `json:"data"`
}

// UploadData upload data
type UploadData struct {
	Size   int64  `json:"size"` // bkrepo always return 0
	Sha256 string `json:"sha256"`
}

// DownloadFileReq is repo download file request struct
type DownloadFileReq struct {
	// BizID is business ID
	BizID uint32 `json:"bizID"`
	// AppID is application ID
	AppID uint32 `json:"appID"`
	// Sign is sha256 encoding of file
	Sign string `json:"sign"`
}

// Configuration is repo configuration.
type Configuration struct {
	// Type is configuration type(local).
	Type string `json:"type"`
}

// GenerateTempDownloadURLReq is repo generate temp download url request struct.
type GenerateTempDownloadURLReq struct {
	// ProjectID is bscp project name in repo.
	ProjectID string `json:"projectId"`
	// RepoName is name of new repo.
	RepoName string `json:"repoName"`
	// FullPathSet is node full path set.
	FullPathSet []string `json:"fullPathSet"`
	// ExpireSeconds is expire seconds.
	ExpireSeconds uint32 `json:"expireSeconds"`
	// Permits is count limit for download.
	Permits uint32 `json:"permits"`
	// Type is download type.
	Type string `json:"type"`
}

// GenerateTempDownloadURLResp is repo generate temp download url response struct.
type GenerateTempDownloadURLResp struct {
	// Code is response code.
	Code int `json:"code"`
	// Message is response message.
	Message string `json:"message"`
	// Data is response data.
	Data []GenerateTempDownloadURLData `json:"data"`
	// TraceID is trace id.
	TraceID string `json:"traceId"`
}

// GenerateTempDownloadURLData is repo generate temp download url response data struct.
type GenerateTempDownloadURLData struct {
	// ProjectID is bscp project name in repo.
	ProjectID string `json:"projectId"`
	// RepoName is name of new repo.
	RepoName string `json:"repoName"`
	// FullPath is node full path.
	FullPath string `json:"fullPath"`
	// URL is temp download url.
	URL string `json:"url"`
	// AuthorizedUserList is authorized user list.
	AuthorizedUserList []string `json:"authorizedUserList"`
	// AuthorizedIpList is authorized ip list.
	AuthorizedIpList []string `json:"authorizedIpList"`
	// ExpireDate is expire date.
	ExpireDate string `json:"expireDate"`
	// Permits is count limit for download.
	Permits uint32 `json:"permits"`
	// Type is download type.
	Type string `json:"type"`
}

// FileMetadataValue ..
type FileMetadataValue struct {
	ByteSize int64  `json:"byte_size"`
	Sha256   string `json:"sha256"`
}

// GenRepoName generate repo repository name, like "bscp-{version}-{biz_id}".
func GenRepoName(bizID uint32) (string, error) {
	if bizID == 0 {
		return "", errors.New("biz_id should > 0")
	}

	return fmt.Sprintf("bscp-%s-biz-%d", version, bizID), nil
}

// GenNodeFullPath generate node full path, like "/file/c7d78b78205a2619eb2b80558f85ee188836ef5f4f317f8587ee38bc3712a8a"
func GenNodeFullPath(sign string) (string, error) {
	if len(sign) == 0 {
		return "", errors.New("sign is required")
	}

	return fmt.Sprintf("%s%s", nodeFrontPath, sign), nil
}

// NodeOption used to generate node path.
type NodeOption struct {
	// Project bscp project in repo, optional. auth method is not required this parameter.
	Project string
	// BizID biz id.
	BizID uint32
	// Sign file sha256.
	Sign string
}

// GenNodePath generate node upload/download path by download method.
// repo path format: /generic/{project}/{repoName}/{fullPath}
// normal path format: /generic/{project}/bscp-{version}-{biz_id}/file/{file sha256}
func GenNodePath(opt *NodeOption) (string, error) {
	if opt == nil {
		return "", errors.New("option is nil")
	}

	if len(opt.Project) == 0 {
		return "", errors.New("project should > 0")
	}

	if opt.BizID == 0 {
		return "", errors.New("biz_id should > 0")
	}

	if len(opt.Sign) != 64 {
		return "", errors.New("file sha256 is not standard format")
	}

	repoName, err := GenRepoName(opt.BizID)
	if err != nil {
		return "", err
	}

	fullPath, err := GenNodeFullPath(opt.Sign)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/generic/%s/%s%s", opt.Project, repoName, fullPath), nil
}

// NodeMeta node metadata info.
type NodeMeta struct {
	BizID       uint32   `json:"biz_id"`
	AppID       []uint32 `json:"app_id"`
	TmplSpaceID []uint32 `json:"template_space_id"`
}

// String get content meta repo request format.
func (c NodeMeta) String() (string, error) {
	var (
		appIDs, tmplSpaceIDs []byte
		err                  error
	)

	appIDs, err = jsoni.Marshal(c.AppID)
	if err != nil {
		return "", fmt.Errorf("marshal node metadata app ids failed, err: %v", err)
	}

	tmplSpaceIDs, err = jsoni.Marshal(c.TmplSpaceID)
	if err != nil {
		return "", fmt.Errorf("marshal node metadata tmplate space ids failed, err: %v", err)
	}

	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("biz_id=%d&app_id=%s&template_space_id=%s",
		c.BizID, appIDs, tmplSpaceIDs))), nil
}

// GenS3NodeFullPath ..
func GenS3NodeFullPath(bizID uint32, sign string) (string, error) {
	if len(sign) == 0 {
		return "", errors.New("sign is required")
	}
	if len(sign) != 64 {
		return "", errors.New("file sha256 is not standard format")
	}

	repoName := fmt.Sprintf("bscp-%s-biz-%d", version, bizID)
	return fmt.Sprintf("/%s%s%s", repoName, nodeFrontPath, sign), nil
}
