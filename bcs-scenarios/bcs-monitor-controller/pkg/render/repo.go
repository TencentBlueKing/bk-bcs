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

package render

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/utils"
)

const (
	EnvNameGitRepoURL  = "GIT_URL"
	EnvNameGitUserName = "GIT_USERNAME"
	EnvNameGitSecret   = "GIT_SECRET"
)

var errStop = errors.New("stop")

// gitRepo clone git repo and pull in certain freq
type gitRepo struct {
	URL       string
	Directory string
	Username  string
	Password  string
	Frequency time.Duration

	cli client.Client
}

// newGitRepo return new git repo
func newGitRepo(cli client.Client, opt *option.ControllerOption) (*gitRepo, error) {
	repo := &gitRepo{
		Directory: opt.ScenarioPath,
		Frequency: opt.ScenarioGitRefreshFreq,
		cli:       cli,
	}
	repo.loadEnv()

	err := repo.Clone()
	if err != nil {
		blog.Errorf("clone git repo(%+v) failed, err: %s", repo, err.Error())
		return nil, err
	}

	ticker := time.NewTicker(repo.Frequency)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			repo.StartAutoUpdate()
		}
	}()

	return repo, nil
}

func (gr *gitRepo) loadEnv() {
	repoURL := os.Getenv(EnvNameGitRepoURL)
	username := os.Getenv(EnvNameGitUserName)
	secret := os.Getenv(EnvNameGitSecret)

	gr.URL = repoURL
	gr.Username = username
	gr.Password = secret
}

func (gr *gitRepo) Clone() error {
	_, err := git.PlainClone(gr.Directory, false, &git.CloneOptions{
		URL: gr.URL,
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
		RemoteURL: gr.URL,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		blog.Errorf("Internal error! pull repo failed, err: %s", err.Error())
		return err
	}
	return nil
}

func (gr *gitRepo) StartAutoUpdate() {
	if err := gr.Pull(); err != nil {
		blog.Errorf("pull git repo failed! err: %s", err.Error())
		// DOTO send event
		return
	}

	changeDirs, err := gr.getChangedDirs()
	if err != nil {
		blog.Errorf("get git repo change dirs failed! err: %s", err.Error())
		// DOTO send event
		return
	}
	if len(changeDirs) > 0 {
		blog.Infof("found update scenario: %s", utils.ToJsonString(changeDirs))
		_ = gr.resolveChangeScenario(changeDirs)
	}
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

func (gr *gitRepo) resolveChangeScenario(scenarios []string) error {
	for _, scenario := range scenarios {
		selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
			monitorextensionv1.LabelKeyForScenarioName: scenario,
		}))
		if err != nil {
			blog.Errorf("generate selector for scenario'%s' failed, err: %s", scenario, err.Error())
			// DOTO 是否continue后 统一成multierror上报
			return err
		}
		appMonitorList := &monitorextensionv1.AppMonitorList{}
		if inErr := gr.cli.List(context.Background(), appMonitorList, &client.ListOptions{
			LabelSelector: selector,
		}); inErr != nil {
			blog.Errorf("list app monitor for scenario'%s' failed, err: %s", scenario, inErr.Error())
			// DOTO 同上
			return inErr
		}

		for _, appMonitor := range appMonitorList.Items {
			if inErr := gr.patchAppMonitorAnnotation(&appMonitor, time.Now()); inErr != nil {
				blog.Errorf("patch app monitor'%s/%s' annotation failed, err: %s", appMonitor.GetNamespace(),
					appMonitor.GetName(), inErr.Error())
				return inErr
			}

			blog.Infof("scenario '%s' related app monitor '%s/%s' updated", scenario, appMonitor.GetNamespace(),
				appMonitor.GetName())
		}
		blog.Infof("scenario '%s' related app monitor all updated", scenario)
	}
	return nil
}

func (gr *gitRepo) patchAppMonitorAnnotation(
	monitor *monitorextensionv1.AppMonitor, updateTime time.Time,
) error {
	patchStruct := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				monitorextensionv1.AnnotationScenarioUpdateTimestamp: updateTime.Format(time.RFC3339Nano),
			},
		},
	}
	patchBytes, err := json.Marshal(patchStruct)
	if err != nil {
		return errors.Wrapf(err, "marshal patchStruct for app monitor '%s/%s' failed", monitor.GetNamespace(),
			monitor.GetName())
	}
	rawPatch := client.RawPatch(k8stypes.MergePatchType, patchBytes)
	updateAppMonitor := &monitorextensionv1.AppMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      monitor.GetName(),
			Namespace: monitor.GetNamespace(),
		},
	}
	if inErr := gr.cli.Patch(context.Background(), updateAppMonitor, rawPatch, &client.PatchOptions{}); inErr != nil {
		return errors.Wrapf(err, "patch app monitor %s/%s annotation failed, patcheStruct: %s",
			monitor.GetNamespace(), monitor.GetName(), string(patchBytes))
	}
	return nil
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
			if !contains(changedDirs, dir) {
				changedDirs = append(changedDirs, dir)
			}
		} else {
			if file.Hash != parentFile.Hash {
				dir, inErr := getTopDir(file.Name)
				if inErr != nil {
					return inErr
				}
				if !contains(changedDirs, dir) {
					changedDirs = append(changedDirs, dir)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return changedDirs, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getTopDir(path string) (string, error) {
	dir := filepath.Dir(path)
	topDir := strings.Split(dir, string(filepath.Separator))
	if len(topDir) == 0 {
		return "", errors.Errorf("unknown error, update file: %s, has not related scenario", path)
	}
	return topDir[0], nil
}
