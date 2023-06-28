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
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/thirdparty/repo"
)

const (
	// defaultWriteBufferSize is default write buffer size, 4KB.
	defaultWriteBufferSize = 4 << 10

	// defaultReadBufferSize is default read buffer size, 4KB.
	defaultReadBufferSize = 4 << 10
)

var (
	// The transport used to perform proxy requests. If nil,
	// http.DefaultTransport is used.
	defaultTransport http.RoundTripper = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                (&net.Dialer{Timeout: 10 * time.Second}).Dial,
		MaxConnsPerHost:     200,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     time.Minute,
		WriteBufferSize:     defaultWriteBufferSize,
		ReadBufferSize:      defaultReadBufferSize,
	}

	// notImplementedErr
	notImplementedErr = errors.New("notImplemented")
)

// ObjectMetadata 文件元数据
type ObjectMetadata struct {
	ByteSize int64  `json:"byte_size"`
	Sha256   string `json:"sha256"`
}

// DecoratorInter ..
type DecoratorInter interface {
	Root() string
	RepoName() string
	Path(sign string) string
	RelativePath(sign string) string
	Url() string
}

// ObjectDownloader 文件下载
type ObjectDownloader interface {
	DownloadLink(kt *kit.Kit, sign string, fetchLimit uint32) (string, error)
	AsyncDownload(kt *kit.Kit, sign string) (string, error)
	AsyncDownloadStatus(kt *kit.Kit, sign string, taskID string) (bool, error)
	URIDecorator(bizID uint32) DecoratorInter
}

// Provider repo provider interface
type Provider interface {
	ObjectDownloader
	Upload(kt *kit.Kit, sign string, body io.Reader) (*ObjectMetadata, error)
	Download(kt *kit.Kit, sign string) (io.ReadCloser, int64, error)
	Metadata(kt *kit.Kit, sign string) (*ObjectMetadata, error)
}

// GetFileSign get file sha256
func GetFileSign(r *http.Request) (string, error) {
	sign := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	if len(sign) != 64 {
		return "", errors.New("not valid X-Bkapi-File-Content-Id in header")
	}

	return sign, nil
}

type uriDecoratorInter struct {
	bizID uint32
}

// Root ..
func (u *uriDecoratorInter) Root() string {
	return ""
}

// RepoName ..
func (u *uriDecoratorInter) RepoName() string {
	name, _ := repo.GenRepoName(u.bizID) // nolint
	return name
}

// Path ..
func (u *uriDecoratorInter) Path(sign string) string {
	p, _ := repo.GenS3NodeFullPath(u.bizID, sign) // nolint
	return p

}

// RelativePath ..
func (u *uriDecoratorInter) RelativePath(sign string) string {
	p, _ := repo.GenNodeFullPath(sign) // nolint
	return p
}

// Url ..
func (u *uriDecoratorInter) Url() string {
	return ""
}

// newUriDecoratorInter ..
func newUriDecoratorInter(bizID uint32) DecoratorInter {
	return &uriDecoratorInter{bizID: bizID}
}

// NewProvider init provider factory by storage type
func NewProvider(conf cc.Repository) (Provider, error) {
	switch strings.ToUpper(string(conf.StorageType)) {
	case string(cc.S3):
		return newCosProvider(conf.S3)
	case string(cc.BkRepo):
		return newBKRepoProvider(conf)
	}
	return nil, fmt.Errorf("store with type %s is not supported", conf.StorageType)
}
