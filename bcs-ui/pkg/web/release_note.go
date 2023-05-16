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
 *
 */

package web

import (
	"fmt"
	"net/http"
	"path"
	"regexp"
	"sort"

	"github.com/go-chi/render"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
)

const (
	changelogPath = "CHANGELOG/zh_CN"
	featurePath   = "frontend/static/features.md"
	defaultLang   = "zh_cn" // viper 默认不区分大小写
)

var changelogNamePattern = regexp.MustCompile(`^(?P<version>v1.\d+.\d+)_(?P<date>[\w-]+).md$`)

// ChangeLog
type ChangeLog struct {
	Content string `json:"content"`
	Date    string `json:"date"`
	Version string `json:"version"`
}

// Feature
type Feature struct {
	Content string `json:"content"`
}

// ReleaseNote
type ReleaseNote struct {
	ChangeLogs []*ChangeLog `json:"changelog"`
	Feature    *Feature     `json:"feature"`
}

func parseChangelogName(value string) map[string]string {
	result := make(map[string]string)

	match := changelogNamePattern.FindStringSubmatch(value)
	if len(match) == 0 {
		return result
	}

	for i, name := range changelogNamePattern.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}

// initReleaseNote
func (s *WebServer) initReleaseNote() error {
	entries, err := s.embedWebServer.RootFS().ReadDir(changelogPath)
	if err != nil {
		return err
	}

	cls := make([]*ChangeLog, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		c, err := s.embedWebServer.RootFS().ReadFile(path.Join(changelogPath, e.Name()))
		if err != nil {
			return err
		}

		result := parseChangelogName(e.Name())
		if len(result) == 0 {
			return fmt.Errorf("not valid changelog name: %s", e.Name())
		}

		cls = append(cls, &ChangeLog{Content: string(c), Date: result["date"], Version: result["version"]})
	}

	sort.Slice(cls, func(i, j int) bool {
		return cls[i].Version > cls[j].Version
	})

	// 读取特性配置
	f, err := s.embedWebServer.RootFS().ReadFile(featurePath)
	if err != nil {
		return err
	}

	feature := &Feature{Content: string(f)}
	if config.G.FrontendConf.Features[defaultLang] != "" {
		feature.Content = config.G.FrontendConf.Features[defaultLang]
	}

	s.releaseNote = &ReleaseNote{ChangeLogs: cls, Feature: feature}

	return nil
}

// ReleaseNoteHandler 含版本日志和特性说明
func (s *WebServer) ReleaseNoteHandler(w http.ResponseWriter, r *http.Request) {
	okResponse := &OKResponse{Message: "OK", Data: s.releaseNote, RequestID: r.Header.Get("x-request-id")}
	render.JSON(w, r, okResponse)
}
