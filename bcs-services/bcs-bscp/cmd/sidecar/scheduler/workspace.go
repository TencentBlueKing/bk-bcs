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

package scheduler

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	pbci "bscp.io/pkg/protocol/core/config-item"
)

const bscpWorkspaceDir = constant.SideWorkspaceDir

// NewWorkspace create a runtime workspace instance
func NewWorkspace(ws cc.SidecarWorkspace, spec cc.SidecarAppSpec) (*RuntimeWorkspace, error) {
	if err := ws.Validate(); err != nil {
		return nil, fmt.Errorf("invalid sidecar worksapce, err: %v", err)
	}

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid sidecar app spec, err: %v", err)
	}

	absDirectory, err := filepath.Abs(ws.RootDirectory)
	if err != nil {
		return nil, fmt.Errorf("invalid absolute workspace's root directory, err: %v", err)
	}
	absDirectory = strings.TrimRight(absDirectory, "/")

	rw := &RuntimeWorkspace{
		BizID:     spec.BizID,
		Workspace: fmt.Sprintf("%s/%s", absDirectory, bscpWorkspaceDir),
	}

	if err := rw.prepare(); err != nil {
		return nil, fmt.Errorf("prepare workspace failed, err: %v", err)
	}

	return rw, nil
}

// Note: update this landscape.
// sidecar's workspace landscape.
// ${workspaceRootDirectory}/bk-bscp
// ├── file-release-v1
// │     └──${bizID}
// │         └──${appID}
// │             └──${releaseID}
// │                 ├── config-items
// │                 │     ├── mongodb.yaml
// │                 │     ├── /data/bscp/etc/redis/redis.yaml
// │                 │     └── ......
// │                 ├── file.lock
// │                 └── metadata
// └── logs

// RuntimeWorkspace defines the sidecar's workspace information.
type RuntimeWorkspace struct {
	BizID     uint32
	Workspace string
}

// prepare the sidecar's runtime workspace
func (rw RuntimeWorkspace) prepare() error {
	if len(rw.Workspace) == 0 {
		return errors.New("invalid sidecar workspace root directory, is empty")
	}

	if err := os.MkdirAll(rw.Workspace, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir root directory failied, err: %v", err)
	}

	return nil
}

// AppFileReleaseWorkspace return the app works at file mode's release workspace instance.
func (rw RuntimeWorkspace) AppFileReleaseWorkspace(appID uint32) (*AppFileWorkspace, error) {
	af := &AppFileWorkspace{
		BizID:     rw.BizID,
		AppID:     appID,
		Workspace: fmt.Sprintf("%s/fileReleaseV1/%d/%d", rw.Workspace, rw.BizID, appID),
	}

	if err := af.prepare(); err != nil {
		return nil, fmt.Errorf("prepare app file workspace failed, err: %v", err)
	}

	return af, nil
}

// AppFileWorkspace defines the app which works at file mode's release workspace.
type AppFileWorkspace struct {
	BizID     uint32
	AppID     uint32
	Workspace string
}

// prepare the app which works at file mode's release workspace.
func (af AppFileWorkspace) prepare() error {
	if err := os.MkdirAll(af.Workspace, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %s failed, err: %v", af.Workspace, err)
	}

	return nil
}

// PrepareReleaseDirectory create the application release's directory.
func (af AppFileWorkspace) PrepareReleaseDirectory(releaseID uint32) error {
	dir := fmt.Sprintf("%s/%d/", af.Workspace, releaseID)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %s failed, err: %v", dir, err)
	}

	return nil
}

// MetadataFile returns the app release's absolute metadata file name.
func (af AppFileWorkspace) MetadataFile(releaseID uint32) string {
	return fmt.Sprintf("%s/%d/metadata.json", af.Workspace, releaseID)
}

// LockFile returns the application release's absolute lock file name.
func (af AppFileWorkspace) LockFile(releaseID uint32) string {
	return fmt.Sprintf("%s/%d/file.lock", af.Workspace, releaseID)
}

// PrepareCIDirectory create the application CI's directory.
func (af AppFileWorkspace) PrepareCIDirectory(releaseID uint32, ciPath string) error {

	dir := af.ConfigItemsDirectory(releaseID, ciPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %s failed, err: %v", dir, err)
	}

	return nil
}

// ConfigItemRootDirectory return the directory where to store the user's all config items content.
func (af AppFileWorkspace) ConfigItemRootDirectory(releaseID uint32) string {
	return fmt.Sprintf("%s/%d/configItems", af.Workspace, releaseID)
}

// ConfigItemsDirectory return the directory where to store the user's configure items content with path.
func (af AppFileWorkspace) ConfigItemsDirectory(releaseID uint32, ciPath string) string {
	if len(ciPath) == 0 {
		return af.ConfigItemRootDirectory(releaseID)
	}

	ciPath = strings.Trim(ciPath, " ")
	ciPath = strings.Trim(ciPath, "/")

	return fmt.Sprintf("%s/%s", af.ConfigItemRootDirectory(releaseID), ciPath)
}

// ConfigItemFile return the config item's absolute file name.
func (af AppFileWorkspace) ConfigItemFile(releaseID uint32, ciPath string, ciFilename string) string {
	return fmt.Sprintf("%s/%s", af.ConfigItemsDirectory(releaseID, ciPath), ciFilename)
}

// SetFilePermission set config item file's permission.
func (af AppFileWorkspace) SetFilePermission(ciPath string, pm *pbci.FilePermission) error {
	file, err := os.Open(ciPath)
	if err != nil {
		return fmt.Errorf("open the target file failed, err: %v", err)
	}
	defer file.Close()

	mode, err := strconv.ParseInt("0"+pm.Privilege, 8, 64)
	if err != nil {
		return fmt.Errorf("parse %s privilege to int failed, err: %v", pm.Privilege, err)
	}

	if err = file.Chmod(os.FileMode(mode)); err != nil {
		return fmt.Errorf("file chmod %o failed, err: %v", mode, err)
	}

	ur, err := user.Lookup(pm.User)
	if err != nil {
		return fmt.Errorf("look up %s user failed, err: %v", pm.User, err)
	}

	uid, err := strconv.Atoi(ur.Uid)
	if err != nil {
		return fmt.Errorf("atoi %s uid failed, err: %v", ur.Uid, err)
	}

	gp, err := user.LookupGroup(pm.UserGroup)
	if err != nil {
		return fmt.Errorf("look up %s group failed, err: %v", pm.User, err)
	}

	gid, err := strconv.Atoi(gp.Gid)
	if err != nil {
		return fmt.Errorf("atoi %s gid failed, err: %v", gp.Gid, err)
	}

	if err := file.Chown(uid, gid); err != nil {
		return fmt.Errorf("file chown %s %s failed, err: %v", ur.Uid, gp.Gid, err)
	}

	return nil
}
