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

package service

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"time"

	"github.com/gofrs/flock"

	"bk-bscp/pkg/common"
	dl "bk-bscp/pkg/downloader"
)

var (
	// ErrorFLockFailed is error of file lock failed.
	ErrorFLockFailed = errors.New("can't get flock, try again later")

	// tryLockWaitTime is wait duration for try lock.
	tryLockWaitTime = time.Second

	// maxTryLockCount is max try lock count.
	maxTryLockCount = 3600

	// sidecarDownloadOperator is operator used for sidecars to download bkrepo content.
	sidecarDownloadOperator = "sidecar-admin"

	// sidecarDownloadAppCode is bk appcode used for sidecars to download bkrepo content.
	sidecarDownloadAppCode = "sidecar-admin"
)

// GetAppModInfoValue parses app mod info and return string value.
func GetAppModInfoValue(v interface{}) string {
	if v == nil {
		return ""
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Int:
		return fmt.Sprintf("%d", v.(int))

	case reflect.String:
		return fmt.Sprintf("%s", v.(string))

	default:
		return ""
	}
}

// ModKey returns key string for target app mod.
func ModKey(bizID, appID, path string) string {
	return fmt.Sprintf("%s_%s_%s", bizID, appID, filepath.Clean(path))
}

// LockFile locks target file.
func LockFile(file string, needBlock bool) (*flock.Flock, error) {
	fl := flock.New(file)

	isLocked, err := fl.TryLock()
	if err != nil {
		return nil, err
	}

	if isLocked {
		return fl, nil
	}

	if !needBlock {
		return nil, ErrorFLockFailed
	}

	for i := 0; i < maxTryLockCount; i++ {
		time.Sleep(tryLockWaitTime)

		fl := flock.New(file)

		isLocked, err := fl.TryLock()
		if err != nil {
			return nil, err
		}

		if isLocked {
			return fl, nil
		}
	}

	return nil, ErrorFLockFailed
}

// UnlockFile unlocks target file lock.
func UnlockFile(fl *flock.Flock) {
	if fl != nil {
		fl.Unlock()
	}
}

// DownloadConfigsOption is download configs option.
type DownloadConfigsOption struct {
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

const (
	// defaultDownloadConcurrent is default download concurrent num.
	defaultDownloadConcurrent = 5

	// defaultDownloadLimitBytesInSecond is default download limit in second, 1MB.
	defaultDownloadLimitBytesInSecond = int64(1024 * 1024 * 1)
)

// DownloadConfigs downloads target configs content base on cid.
func DownloadConfigs(option *DownloadConfigsOption, timeout time.Duration) (time.Duration, error) {
	timenow := time.Now()
	sequence := common.Sequence()

	headers := make(map[string]string)
	headers[common.RidHeaderKey] = sequence
	headers[common.UserHeaderKey] = sidecarDownloadOperator
	headers[common.ContentIDHeaderKey] = option.ContentID
	headers[common.AppCodeHeaderKey] = sidecarDownloadAppCode

	if option.Concurrent == 0 {
		option.Concurrent = defaultDownloadConcurrent
	}
	if option.LimitBytesInSecond == 0 {
		option.LimitBytesInSecond = defaultDownloadLimitBytesInSecond
	}

	downloader := dl.NewDownloader(option.URL, option.Concurrent, headers, option.NewFile)
	downloader.SetRateLimiterOption(dl.NewSimpleRateLimiter(option.LimitBytesInSecond))

	// download.
	if err := downloader.Download(timeout); err != nil {
		downloader.Clean()
		return time.Since(timenow), fmt.Errorf("download failed, %+v", err)
	}

	// check cid.
	contentID, err := common.FileSHA256(option.NewFile)
	if err != nil {
		downloader.Clean()
		return time.Since(timenow), fmt.Errorf("check file cid failed, %+v", err)
	}

	if contentID != option.ContentID {
		downloader.Clean()
		return time.Since(timenow), errors.New("download invalid cid")
	}

	return time.Since(timenow), nil
}

// PermissionOption descs file permission details for effecting base on content.
type PermissionOption struct {
	// User config file user.
	User string

	// UserGroup config file user group.
	UserGroup string

	// FilePrivilege config file privilege.
	FilePrivilege string

	// FileFormat config file Format.
	FileFormat string

	// FileMode config file mode.
	FileMode int32
}

// FlushFilePermission flushs config file privilege/user/user group.
func FlushFilePermission(file string, option *PermissionOption) error {
	if option == nil {
		return nil
	}

	// chmod file.
	fileMode := common.ToFileMode(option.FilePrivilege)
	if fileMode != 0 {
		if err := os.Chmod(file, fileMode); err != nil {
			return err
		}
	}

	if len(option.User) == 0 || len(option.UserGroup) == 0 {
		// ingore user group flush action.
		return nil
	}

	// get target file user.
	targetUser, err := user.Lookup(option.User)
	if err != nil {
		return err
	}
	targetUID := targetUser.Uid

	targetGroup, err := user.LookupGroup(option.UserGroup)
	if err != nil {
		return err
	}
	targetGroupid := targetGroup.Gid

	if targetUser.Gid != targetGroupid {
		return errors.New("user and user group inconsistent")
	}

	// chown file.
	if err := os.Chown(file, common.ToInt(targetUID), common.ToInt(targetGroupid)); err != nil {
		return err
	}
	return nil
}
