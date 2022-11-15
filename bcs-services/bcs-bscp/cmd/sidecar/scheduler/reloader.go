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

package scheduler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/jsoni"
	sfs "bscp.io/pkg/sf-share"
)

// Reloader implements all the supported operations which used to notify
// app to reload config file.
type Reloader interface {
	NotifyReload(vas *kit.Vas, desc *sfs.ReleaseEventMetaV1) error
}

// reloader is used to notify app to reload config items.
type reloader struct {
	appReloads map[uint32]*appReload
	workspace  *RuntimeWorkspace
}

// NewReloader new reloader.
func NewReloader(ws *RuntimeWorkspace, appReloads map[uint32]*sfs.Reload) (Reloader, error) {
	if ws == nil {
		return nil, errors.New("runtime workspace is nil")
	}

	if len(appReloads) == 0 {
		return nil, errors.New("appReload is nil")
	}

	appReloadMap := make(map[uint32]*appReload)
	for appID, reload := range appReloads {
		if err := reload.ReloadType.Validate(); err != nil {
			return nil, err
		}

		switch reload.ReloadType {
		case table.ReloadWithFile:
			if reload.FileReloadSpec != nil && len(reload.FileReloadSpec.ReloadFilePath) == 0 {
				return nil, errors.New("reload file path is nil")
			}

			if err := prepareReloadDirectory(reload.FileReloadSpec.ReloadFilePath); err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("reload type: %s not support", reload.ReloadType)
		}

		appWorkspace, err := ws.AppFileReleaseWorkspace(appID)
		if err != nil {
			return nil, fmt.Errorf("prepare app workspace failed, err: %v", err)
		}

		appReloadMap[appID] = &appReload{
			reload:    reload,
			workspace: appWorkspace,
		}
	}

	reloader := &reloader{
		appReloads: appReloadMap,
		workspace:  ws,
	}

	return reloader, nil
}

// NotifyReload notify app to reload config items.
func (r *reloader) NotifyReload(vas *kit.Vas, desc *sfs.ReleaseEventMetaV1) error {
	ar, exist := r.appReloads[desc.AppID]
	if !exist {
		return fmt.Errorf("app: %d reload not exist", desc.AppID)
	}

	switch ar.reload.ReloadType {
	case table.ReloadWithFile:
		if err := ar.reloadWithFile(vas, desc); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown app reload type: %s", ar.reload.ReloadType)
	}

	return nil
}

type appReload struct {
	reload    *sfs.Reload
	workspace *AppFileWorkspace
}

// reloadMetadataVersion is reload metadata version.
const reloadMetadataVersion = "v1"

// reloadMetadataV1 defines notify application to reload config file's metadata.
type reloadMetadataV1 struct {
	Version       string   `json:"version"`
	Timestamp     string   `json:"timestamp"`
	AppID         uint32   `json:"app_id"`
	ReleaseID     uint32   `json:"release_id"`
	RootDirectory string   `json:"root_directory"`
	ConfigItem    []string `json:"config_item"`
}

// convReleaseEventMeta to reloader metadata.
func (r *appReload) convReleaseEventMeta(desc *sfs.ReleaseEventMetaV1) *reloadMetadataV1 {

	ciSubPathList := make([]string, 0)
	for _, meta := range desc.CIMetas {
		ciSubPath := filepath.Clean("/" + meta.ConfigItemSpec.Path + "/" + meta.ConfigItemSpec.Name)
		ciSubPathList = append(ciSubPathList, ciSubPath)
	}

	meta := &reloadMetadataV1{
		Version:       reloadMetadataVersion,
		Timestamp:     time.Now().Format(constant.TimeStdFormat),
		AppID:         desc.AppID,
		ReleaseID:     desc.ReleaseID,
		RootDirectory: r.workspace.ConfigItemRootDirectory(desc.ReleaseID),
		ConfigItem:    ciSubPathList,
	}

	return meta
}

// reloadWithFile write reload metadata info to reload file.
func (r *appReload) reloadWithFile(vas *kit.Vas, desc *sfs.ReleaseEventMetaV1) error {

	logs.Infof("app: %d, release: %d, start to exec reload with file", desc.AppID, desc.ReleaseID)

	meta := r.convReleaseEventMeta(desc)

	reloadFile := r.reload.FileReloadSpec.ReloadFilePath

	content, err := jsoni.MarshalIndent(meta, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal app reload metadata failed, err: %v", err)
	}

	metaFile, err := os.OpenFile(reloadFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open the app reload metadata file failed, err: %v", err)
	}
	defer metaFile.Close()

	if _, err := metaFile.Write(content); err != nil {
		return fmt.Errorf("write app reload metadata content to file failed, err: %v", err)
	}

	logs.Infof("app: %d, release: %d, write reload metadata to reload file: %s successfully, metadata: %s", desc.AppID,
		desc.ReleaseID, reloadFile, regexp.MustCompile("\\s+").ReplaceAllString(string(content), ""))

	return nil
}
