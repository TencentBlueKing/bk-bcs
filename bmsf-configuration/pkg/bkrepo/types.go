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

package bkrepo

const (
	// CONFIGSREPONAME is bscp config repository name.
	CONFIGSREPONAME = "bscp-configs"

	// REPOTYPE is repository type.
	REPOTYPE = "GENERIC"

	// CATEGORY is bscp repository category.
	CATEGORY = "LOCAL"

	// REPOCFGTYPE repository type in bkrepo.
	REPOCFGTYPE = "local"

	// GENERICAPIPATH is bkrepo api path for generic.
	GENERICAPIPATH = "api/generic/"

	// BSCPBIZIDPREFIX is bscp bkrepo bizid prefix(bscp-{CMDB biz_id}).
	BSCPBIZIDPREFIX = "bscp-"
)

const (
	// BKRepoUIDHeaderKey is bkrepo uid header key.
	BKRepoUIDHeaderKey = "X-BKREPO-UID"

	// BKRepoSHA256HeaderKey is bkrepo sha256 header key.
	BKRepoSHA256HeaderKey = "X-BKREPO-SHA256"

	// BKRepoOverwriteHeaderKey is bkrepo upload overwrite flag header key.
	BKRepoOverwriteHeaderKey = "X-BKREPO-OVERWRITE"
)

const (
	// BKRepoErrCodeProjectAlreadyExist bk-repo project already exist.
	BKRepoErrCodeProjectAlreadyExist = 251005 // OLD: 251002

	// BKRepoErrCodeRepoAlreadyExist bk-repo repository already exist.
	BKRepoErrCodeRepoAlreadyExist = 251007 // OLD: 251004

	// BKRepoErrCodeNodeAlreadyExist bk-repo node already exist.
	BKRepoErrCodeNodeAlreadyExist = 251012 // OLD: 251008
)

// Auth is bkrepo auth information.
type Auth struct {
	// Token is bkrepo ak:sk base64 string.
	Token string

	// UID is bkrepo user id which same with wechatwork user.
	UID string
}

// CommonResp is common response struct.
type CommonResp struct {
	// Code is http request result code.
	Code int `json:"code"`

	// Message is result message.
	Message string `json:"message"`
}

// CreateProjectReq is bkrepo create project request struct.
type CreateProjectReq struct {
	// Name is project name(business name).
	Name string `json:"name"`

	// DisplayName is display name.
	DisplayName string `json:"displayName"`

	// Description is project memo description.
	Description string `json:"description"`
}

// Configuration is repo configuration.
type Configuration struct {
	// Type is configuration type(local).
	Type string `json:"type"`
}

// CreateRepoReq is bkrepo create repo request struct.
type CreateRepoReq struct {
	// ProjectID is target project id(business name).
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

// DownloadContentOption is download content option.
type DownloadContentOption struct {
	// URL is target file source url which should support ranges bytes mode.
	URL string

	// ContentID is configs content id(sha256).
	ContentID string

	// NewFile is new file path-name for target source.
	NewFile string

	// Concurrent is download gcroutine num.
	Concurrent int

	// LimitBytesInSecond is target limit bytes num in second.
	LimitBytesInSecond int64
}

// UploadContentOption is upload content option.
type UploadContentOption struct {
	// URL is target file source url which should support ranges bytes mode.
	URL string

	// ContentID is configs content id(sha256).
	ContentID string
}
