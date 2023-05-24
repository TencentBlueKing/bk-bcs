/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package release

import (
	"context"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	releaseDefaultTimeout = 10 * time.Minute
)

func getChartContent(model store.HelmManagerModel, platform repo.Platform,
	projectID, repoName, chart, version string) ([]byte, error) {
	// 获取对应的仓库信息
	repository, err := model.GetProjectRepository(context.Background(), projectID, repoName)
	if err != nil {
		return nil, err
	}

	// 下载到具体的chart version信息
	contents, err := platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		Chart(chart).
		Download(context.Background(), version)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

// ReleasesSortByUpdateTime sort releases by update time
type ReleasesSortByUpdateTime []*helmmanager.Release

// Len xxx
func (r ReleasesSortByUpdateTime) Len() int { return len(r) }

// Less xxx
func (r ReleasesSortByUpdateTime) Less(i, j int) bool {
	return r[i].GetUpdateTime() > r[j].GetUpdateTime()
}

// Swap xxx
func (r ReleasesSortByUpdateTime) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

func filterIndex(offset, limit int, release []*helmmanager.Release) []*helmmanager.Release {
	if offset >= len(release) {
		return nil
	}

	if limit < 0 {
		limit = 0
	}

	if offset+limit > len(release) {
		return release[offset:]
	}

	return release[offset : offset+limit]
}

var shouldRemoveAnnotations = []string{
	"io.tencent.paas.creator",
	"io.tencent.paas.updator",
	"io.tencent.paas.version",
}

func removeCustomAnnotations(manifest string) *string {
	// Split the input string into lines
	lines := strings.Split(manifest, "\n")

	// Create a new slice to store the filtered lines
	filteredLines := make([]string, 0)

	// Loop through the lines and remove the specified lines
	for _, line := range lines {
		shouldRemove := false
		for _, lineToRemove := range shouldRemoveAnnotations {
			if strings.Contains(line, lineToRemove) {
				shouldRemove = true
				break
			}
		}

		if !shouldRemove {
			filteredLines = append(filteredLines, line)
		}
	}

	// Join the filtered lines into a single string
	result := strings.Join(filteredLines, "\n")
	return &result
}
