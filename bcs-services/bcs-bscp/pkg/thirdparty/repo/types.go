/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package repo

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/runtime/jsoni"
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
	// errCodeProjectAlreadyExist repo project already exist.
	errCodeProjectAlreadyExist = 251005
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
		s.index = s.index + 1
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

// GenRepoName generate repo repository name, like "bscp-{version}-{biz_id}".
func GenRepoName(bizID uint32) (string, error) {
	if bizID == 0 {
		return "", errf.New(errf.InvalidParameter, "biz_id should > 0")
	}

	return fmt.Sprintf("bscp-%s-biz-%d", version, bizID), nil
}

// GenNodeFullPath generate node full path, like "/file/c7d78b78205a2619eb2b80558f85ee188836ef5f4f317f8587ee38bc3712a8a"
func GenNodeFullPath(sign string) (string, error) {
	if len(sign) == 0 {
		return "", errf.New(errf.InvalidParameter, "sign is required")
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
		return "", errf.New(errf.InvalidParameter, "option is nil")
	}

	if len(opt.Project) == 0 {
		return "", errf.New(errf.InvalidParameter, "project should > 0")
	}

	if opt.BizID == 0 {
		return "", errf.New(errf.InvalidParameter, "biz_id should > 0")
	}

	if len(opt.Sign) != 64 {
		return "", errf.New(errf.InvalidParameter, "file sha256 is not standard format")
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
	BizID uint32   `json:"biz_id"`
	AppID []uint32 `json:"app_id"`
}

// String get content meta repo request format.
func (c NodeMeta) String() (string, error) {
	appIDs, err := jsoni.Marshal(c.AppID)
	if err != nil {
		return "", fmt.Errorf("marshal node metadata failed, err: %v", err)
	}

	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("biz_id=%d&app_id=%s", c.BizID, appIDs))), nil
}

// NewUriDecorator create a uri decorator instance.
func NewUriDecorator(r cc.Repository) (UriDecoratorInter, error) {
	switch strings.ToUpper(string(r.StorageType)) {
	case string(cc.S3):
		if len(r.S3.SecretAccessKey) == 0 {
			return nil, errors.New("s3 repository secretAccessKey is empty")
		}
		if len(r.S3.AccessKeyID) == 0 {
			return nil, errors.New("s3 repository accessKeyID is empty")
		}
		if len(r.S3.AppID) == 0 {
			return nil, errors.New("s3 repository appID is empty")
		}
		return &UriS3Decorator{
			Endpoints:       r.S3.Endpoint,
			AccessKeyID:     r.S3.AccessKeyID,
			SecretAccessKey: r.S3.SecretAccessKey,
			AppID:           r.S3.AppID,
		}, nil

	case string(cc.BK_REPO):

		if len(r.BkRepo.Project) == 0 {
			return nil, errors.New("repository project is empty")
		}

		if len(r.BkRepo.Endpoints) == 0 {
			return nil, errors.New("repository endpoints is empty")
		}

		return &UriDecorator{
			index:     0,
			count:     len(r.BkRepo.Endpoints),
			Endpoints: r.BkRepo.Endpoints,
			Project:   r.BkRepo.Project,
		}, nil
	}
	return nil, errors.New("create a uri decorator instance fail")
}

// UriDecorator is used to generate configure item's download uri
type UriDecorator struct {
	index     int
	count     int
	Endpoints []string
	Project   string
	rrLock    sync.Mutex
}

// Init initialize an new uri decorator
func (ud *UriDecorator) Init(bizID uint32) DecoratorInter {
	repoName := fmt.Sprintf("bscp-%s-biz-%d", version, bizID)

	return &Decorator{
		root:       ud.root() + repoName,
		pathPrefix: nodeFrontPath,
	}
}

// root get repo request root path by UriDecorator.
func (ud *UriDecorator) root() string {
	ud.rrLock.Lock()
	defer ud.rrLock.Unlock()

	if ud.index < ud.count-1 {
		ud.index = ud.index + 1
	} else {
		ud.index = 0
	}
	host := ud.Endpoints[ud.index]

	return fmt.Sprintf("%s/generic/%s/", host, ud.Project)
}

// Decorator defines how to generate a repository uri, an configure item download uri: root + Path(sign).
type Decorator struct {
	root       string
	pathPrefix string
}

// Root return Decorator root that is repo download root.
func (de *Decorator) Root() string {
	return de.root
}

// Path generate the download sub path for an configure item.
func (de *Decorator) Path(sign string) string {
	return de.pathPrefix + sign
}

func (de *Decorator) Url() string {
	return ""
}
func (de *Decorator) AccessKeyID() string {
	return ""
}
func (de *Decorator) SecretAccessKey() string {
	return ""
}
func (de *Decorator) GetRepositoryType() cc.StorageMode {
	return cc.BK_REPO
}

type DecoratorInter interface {
	Root() string
	Path(sign string) string
	Url() string
	AccessKeyID() string
	SecretAccessKey() string
	GetRepositoryType() cc.StorageMode
}

type RepositoryTypeInter interface {
	//IsProjectExist(ctx context.Context) error
	CreateRepo(ctx context.Context, req *CreateRepoReq) error
	DeleteRepo(ctx context.Context, bizID uint32, forced bool) error
	//IsNodeExist(ctx context.Context, nodePath string) (bool, error)
	//DeleteNode(ctx context.Context, nodePath string) error
	QueryMetadata(ctx context.Context, opt *NodeOption) (map[string]string, error)
}

func GenS3Name(bizID uint32) (string, error) {
	if bizID == 0 {
		return "", errf.New(errf.InvalidParameter, "biz_id should > 0")
	}
	appID := cc.ApiServer().Repo.S3.AppID

	return fmt.Sprintf("bscp-%s-biz-%d-%s", version, bizID, appID), nil
}

func GenS3NodeFullPath(sign string) (string, error) {
	if len(sign) == 0 {
		return "", errf.New(errf.InvalidParameter, "sign is required")
	}
	if len(sign) != 64 {
		return "", errf.New(errf.InvalidParameter, "file sha256 is not standard format")
	}

	return fmt.Sprintf("%s%s", nodeFrontPath, sign), nil
}

type UriDecoratorInter interface {
	Init(bizID uint32) DecoratorInter
}

// UriS3Decorator is used to generate configure item's download uri
type UriS3Decorator struct {
	Endpoints       string
	AccessKeyID     string
	SecretAccessKey string
	AppID           string
}

// Init initialize an new uri decorator
func (ud *UriS3Decorator) Init(bizID uint32) DecoratorInter {
	repoName := fmt.Sprintf("bscp-%s-biz-%d-%s", version, bizID, ud.AppID)

	return &DecoratorS3{
		root:            repoName,
		url:             ud.root(),
		pathPrefix:      nodeFrontPath,
		accessKeyID:     ud.AccessKeyID,
		secretAccessKey: ud.SecretAccessKey,
	}
}

// root get repo request root path by UriDecorator.
func (ud *UriS3Decorator) root() string {
	return ud.Endpoints
}

// Decorator defines how to generate a repository uri, an configure item download uri: root + Path(sign).
type DecoratorS3 struct {
	root            string
	url             string
	pathPrefix      string
	accessKeyID     string
	secretAccessKey string
}

func (de *DecoratorS3) Url() string {
	return de.url
}
func (de *DecoratorS3) AccessKeyID() string {
	return de.accessKeyID
}
func (de *DecoratorS3) SecretAccessKey() string {
	return de.secretAccessKey
}

// Root return Decorator root that is repo download root.
func (de *DecoratorS3) Root() string {
	return de.root
}

// Path generate the download sub path for an configure item.
func (de *DecoratorS3) Path(sign string) string {
	return de.pathPrefix + sign
}

func (de *DecoratorS3) GetRepositoryType() cc.StorageMode {
	return cc.S3
}
