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
	"strings"

	"github.com/go-chi/render"
	"golang.org/x/text/language"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/i18n"
)

const (
	changelogPath = "CHANGELOG"
	featurePath   = "frontend/static/features"
)

var changelogNamePattern = regexp.MustCompile(`^(?P<version>v1.\d+.\d+)_(?P<date>[\w-]+).md$`)

// ChangeLog change log
type ChangeLog struct {
	Content string `json:"content"`
	Date    string `json:"date"`
	Version string `json:"version"`
}

// Feature feature content
type Feature struct {
	Content string `json:"content"`
}

// ReleaseNoteLang map of ReleaseNote
type ReleaseNoteLang map[language.Tag]ReleaseNote

// ReleaseNote release_note
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

// initReleaseNote :
func (s *WebServer) initReleaseNote() error {
	// obtain the folder under CHANGELOG
	entries, err := s.embedWebServer.RootFS().ReadDir(changelogPath)
	if err != nil {
		return err
	}

	// array of directory name
	directoryNames := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			// get available language
			if !i18n.IsAvailableLanguage(e.Name()) {
				continue
			}
			directoryNames = append(directoryNames, e.Name())
		}
	}

	// array of language and content
	releaseNoteLang := make(ReleaseNoteLang, len(directoryNames))
	for _, fn := range directoryNames {
		langEntries, err := s.embedWebServer.RootFS().ReadDir(changelogPath + "/" + fn)
		if err != nil {
			return err
		}

		langTag := i18n.GetAvailableLanguage(fn, "zh")
		// array of file contents
		cls := make([]*ChangeLog, 0, len(langEntries))
		for _, e := range langEntries {
			if e.IsDir() {
				continue
			}

			// get file contents
			c, err := s.embedWebServer.RootFS().ReadFile(path.Join(changelogPath+"/"+fn, e.Name()))
			if err != nil {
				return err
			}

			// get valid changelog name
			result := parseChangelogName(e.Name())
			if len(result) == 0 {
				return fmt.Errorf("not valid changelog name: %s", e.Name())
			}

			cls = append(cls, &ChangeLog{Content: string(c), Date: result["date"], Version: result["version"]})
		}

		sort.Slice(cls, func(i, j int) bool {
			return cls[i].Version > cls[j].Version
		})

		feature := &Feature{}
		// Priority read configured
		if _, ok := config.G.FrontendConf.Features[strings.ToLower(fn)]; ok {
			feature.Content = config.G.FrontendConf.Features[strings.ToLower(fn)]
		} else {
			featureCorrectPath := featurePath
			if langTag == language.English {
				featureCorrectPath += "_en"
			}
			featureCorrectPath += ".md"
			// 读取特性配置
			f, err := s.embedWebServer.RootFS().ReadFile(featureCorrectPath)
			if err != nil {
				return err
			}
			feature.Content = string(f)
		}

		releaseNoteLang[langTag] = ReleaseNote{
			ChangeLogs: cls,
			Feature:    feature,
		}
	}
	s.releaseNote = releaseNoteLang

	return nil
}

func (s *WebServer) getReleaseNote(r *http.Request) (releaseNote ReleaseNote) {
	lang := i18n.GetLangByRequest(r, config.G.Base.LanguageCode)
	langTag := i18n.GetAvailableLanguage(lang, config.G.Base.LanguageCode)
	return s.releaseNote[langTag]
}

// ReleaseNoteHandler 含版本日志和特性说明
func (s *WebServer) ReleaseNoteHandler(w http.ResponseWriter, r *http.Request) {
	releaseNote := s.getReleaseNote(r)
	okResponse := &OKResponse{Message: "OK", Data: releaseNote, RequestID: r.Header.Get("x-request-id")}
	render.JSON(w, r, okResponse)
}
