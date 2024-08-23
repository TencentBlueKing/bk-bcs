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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/yaml.v3"
)

// ScenarioInfo info of monitor scenario
type ScenarioInfo struct {
	Name       string         `json:"name"`
	Readme     string         `json:"readme"`
	OptionInfo string         `json:"option_info"`
	Config     ScenarioConfig `json:"config"`
}

// ScenarioConfig config
type ScenarioConfig struct {
	Label string `yaml:"label" json:"label"`
	// if true, not visible for users
	Invisible bool `yaml:"invisible" json:"invisible"`
}

// gitRepo clone git repo and pull in certain freq
type gitRepo struct {
	URL       string
	Directory string
	Username  string
	Password  string
	Frequency time.Duration

	Key string

	// TargetRevision defines the revision of the source to sync the application to.
	// In case of Git, this can be commit, tag, or branch. If omitted, will equal to HEAD.
	TargetRevision string `json:"targetRevision,omitempty"`

	ScenarioInfos []*ScenarioInfo

	auth *http.BasicAuth
}

// doto 适配ssh
// doto 测试同repo 不同targetRevision
func newGitRepo(url, username, password, targetRevision, bathPath string) (*gitRepo, error) {
	repoKey := genRepoKey(url, targetRevision)
	repo := &gitRepo{
		URL:            url,
		Directory:      filepath.Join(bathPath, base64.URLEncoding.EncodeToString([]byte(repoKey))),
		Username:       username,
		Password:       password,
		TargetRevision: targetRevision,
		Key:            repoKey,
		auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
	}
	if err := repo.Clone(); err != nil {
		return nil, fmt.Errorf("clone git repo(%+v) failed, err: %s", repo, err.Error())
	}

	return repo, nil
}

func (gr *gitRepo) GetURL() string {
	return gr.URL
}
func (gr *gitRepo) GetDirectory() string {
	return gr.Directory
}

func (gr *gitRepo) GetRepoKey() string {
	return gr.Key
}

func (gr *gitRepo) Clone() error {
	_, err := git.PlainClone(gr.Directory, false, &git.CloneOptions{
		URL:           gr.URL,
		ReferenceName: plumbing.NewBranchReferenceName(gr.TargetRevision),
		SingleBranch:  true,
		Auth:          gr.auth,
	})
	return err
}

func (gr *gitRepo) Pull() error {
	repo, err := git.PlainOpen(gr.Directory)
	if err != nil {
		blog.Errorf("Internal error! open repo failed, err: %s", err.Error())
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		blog.Errorf("Internal error! get repo's work tree failed, err: %s", err.Error())
		return err
	}

	err = worktree.Pull(&git.PullOptions{
		ReferenceName: plumbing.NewBranchReferenceName(gr.TargetRevision),
		SingleBranch:  true,
		Auth:          gr.auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		blog.Errorf("Internal error! pull repo failed, err: %s", err.Error())
		return err
	}
	return nil
}

// Reload Pull the latest branch, and return change directories
func (gr *gitRepo) Reload() ([]string, error) {
	if err := gr.Pull(); err != nil {
		// DOTO send event
		return nil, fmt.Errorf("pull git repo[%s] failed! err: %s", gr.URL,
			err.Error())
	}
	if err := gr.refreshScenarioInfos(); err != nil {
		return nil, fmt.Errorf("refresh scenario info for repo[%s] failed, err: %s", gr.Key, err.Error())
	}
	changeDirs, err := gr.getChangedDirs()
	if err != nil {
		// DOTO send event
		return nil, fmt.Errorf("get git repo[%s] change dirs failed! err: %s", gr.URL, err.Error())
	}
	return changeDirs, nil
	// if len(changeDirs) > 0 {
	// 	blog.Infof("found update scenario: %s", utils.ToJsonString(changeDirs))
	// 	_ = gr.resolveChangeScenario(changeDirs)
	// }
}

func (gr *gitRepo) getChangedDirs() ([]string, error) {
	repo, err := git.PlainOpen(gr.Directory)
	if err != nil {
		blog.Errorf("open repo failed, err: %s", err.Error())
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		blog.Errorf("get repo head failed, err: %s", err.Error())
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		blog.Errorf("get repo's head commit failed, err: %s", err.Error())
		return nil, err
	}

	// 获取上次拉取的HEAD
	prevHead, err := repo.Reference("refs/prev_head", true)
	if err != nil && err != plumbing.ErrReferenceNotFound {
		blog.Errorf("get repo ref failed, err: %s", err.Error())
		return nil, err
	}

	// 保存当前HEAD作为下次拉取的上次HEAD
	err1 := repo.Storer.SetReference(plumbing.NewHashReference("refs/prev_head", ref.Hash()))
	if err1 != nil {
		blog.Errorf("set repo ref failed, err: %s", err1.Error())
		return nil, err1
	}

	// 如果上次拉取的HEAD不存在，则只比较当前HEAD
	if err == plumbing.ErrReferenceNotFound {
		return compareCommitAndParent(commit)
	}

	// 遍历并比较所有新拉取的提交
	changedDirs := make(map[string]struct{})
	iter, err := repo.Log(&git.LogOptions{From: commit.Hash})
	if err != nil {
		blog.Errorf("get repo log failed, err: %s", err.Error())
		return nil, err
	}

	err = iter.ForEach(func(c *object.Commit) error {
		if c.Hash == prevHead.Hash() {
			return errStop
		}

		dirs, inErr := compareCommitAndParent(c)
		if inErr != nil {
			return inErr
		}

		for _, dir := range dirs {
			changedDirs[dir] = struct{}{}
		}

		return nil
	})
	if err != nil && err != errStop {
		return nil, err
	}

	dirs := make([]string, 0, len(changedDirs))
	for dir := range changedDirs {
		dirs = append(dirs, dir)
	}

	return dirs, nil
}

func (gr *gitRepo) refreshScenarioInfos() error {
	f, err := os.Open(gr.Directory)
	if err != nil {
		return fmt.Errorf("os.Open repo[%s] dir[%s] failed", gr.Key, gr.Directory)
	}
	files, err := f.Readdir(-1)
	if err != nil {
		return fmt.Errorf("read repo[%s] dir[%s] failed", gr.Key, gr.Directory)
	}

	scenarioList := make([]*ScenarioInfo, 0)
	for _, file := range files {
		if strInSlice(filterScenarios, file.Name()) {
			// 筛掉测试场景
			continue
		}
		// 所有场景以目录形式存在
		if file.IsDir() {
			config := ScenarioConfig{}
			configStr := readFileContent(filepath.Join(gr.Directory, file.Name(), "config.yaml"))
			// 解析config失败时正常返回
			if err1 := yaml.Unmarshal([]byte(configStr), &config); err1 != nil {
				blog.Errorf("Scene[%s] unmarshal config failed, err: %s", file.Name(), err1.Error)
			}
			if config.Invisible {
				continue
			}
			scenarioList = append(scenarioList, &ScenarioInfo{
				Name:       file.Name(),
				Readme:     readFileContent(filepath.Join(gr.Directory, file.Name(), "README.md")),
				OptionInfo: readFileContent(filepath.Join(gr.Directory, file.Name(), "option.yaml")),
				Config:     config,
			})
		}
	}
	gr.ScenarioInfos = scenarioList
	return nil
}

// GetScenarioInfos return scenario info of related repo
func (gr *gitRepo) GetScenarioInfos() []*ScenarioInfo {
	return gr.ScenarioInfos
}

func compareCommitAndParent(commit *object.Commit) ([]string, error) {
	parents := commit.Parents()
	parent, err := parents.Next()
	if err != nil {
		blog.Errorf("get commit[%s] next parent failed, err: %s", commit.ID(), err.Error())
		return nil, err
	}

	commitFiles, err := commit.Files()
	if err != nil {
		blog.Errorf("get commit[%s] files failed, err: %s", commit.ID(), err.Error())
		return nil, err
	}

	parentFiles, err := parent.Files()
	if err != nil {
		blog.Errorf("open repo failed, err: %s", err.Error())
		return nil, err
	}

	parentFilesMap := make(map[string]*object.File)
	err = parentFiles.ForEach(func(file *object.File) error {
		parentFilesMap[file.Name] = file
		return nil
	})
	if err != nil {
		return nil, err
	}

	var changedDirs []string
	err = commitFiles.ForEach(func(file *object.File) error {
		parentFile, ok := parentFilesMap[file.Name]
		if !ok {
			dir, inErr := getTopDir(file.Name)
			if inErr != nil {
				return inErr
			}
			if !strInSlice(changedDirs, dir) {
				changedDirs = append(changedDirs, dir)
			}
		} else if file.Hash != parentFile.Hash {
			dir, inErr := getTopDir(file.Name)
			if inErr != nil {
				return inErr
			}
			if !strInSlice(changedDirs, dir) {
				changedDirs = append(changedDirs, dir)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return changedDirs, nil
}
