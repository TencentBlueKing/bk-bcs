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
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/argoproj/argo-cd/v2/util/db"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/utils"
)

// Manager 管理所有Repo， 通过轮训定时拉取最新资源。 当发生更新时， 通过patch注解的方式通知appMonitor进入reconcile流程
type Manager struct {
	// 定时pull所有仓库， 判断场景是否更新
	ScenarioRefreshFrequency time.Duration
	// 根据集群内AppMonitor刷新Repo缓存 (如某个Repo已经不被任何AppMonitor引用，则从缓存中删除)
	RepoRefreshFrequency time.Duration
	Cli                  client.Client
	argoDB               db.ArgoDB

	repoStore map[string]Repo
	repoMutex sync.RWMutex
	BasePath  string
}

// NewRepoManager return new repo manager
func NewRepoManager(cli client.Client, opt *option.ControllerOption) (*Manager, error) {
	manager := &Manager{
		ScenarioRefreshFrequency: opt.ScenarioGitRefreshFreq,
		RepoRefreshFrequency:     opt.RepoRefreshFreq,
		Cli:                      cli,
		BasePath:                 opt.ScenarioPath,
		repoStore:                make(map[string]Repo),
		repoMutex:                sync.RWMutex{},
	}
	if opt.EnableArgo {
		argoDB, _, err := NewArgoDB(context.Background(), opt.ArgoAdminNamespace)
		if err != nil {
			return nil, fmt.Errorf("connect to argo failed, err: %w", err)
		}
		manager.argoDB = argoDB
	}

	defaultRepoURL, defaultUserName, defaultSecret := loadEnv()
	defaultRepo, err := newGitRepo(defaultRepoURL, defaultUserName, defaultSecret, "master",
		manager.BasePath)
	if err != nil {
		return nil, fmt.Errorf("generate default repo failed, err: %s", err.Error())
	}

	manager.repoStore[RepoKeyDefault] = defaultRepo

	go manager.startAutoUpdate()

	return manager, nil
}

// RegisterRepoFromArgo register repo into store
func (m *Manager) RegisterRepoFromArgo(repoURL, targetRevision string) error {
	if m.argoDB == nil {
		return fmt.Errorf("argo db not initialized")
	}
	blog.Infof("register repo [%s/%s]", repoURL, targetRevision)
	argoRepo, err := m.argoDB.GetRepository(context.Background(), repoURL)
	if err != nil {
		return fmt.Errorf("get repo[%s] from argo failed, err: %s", repoURL, err.Error())
	}

	repo, err := newGitRepo(repoURL, argoRepo.Username, argoRepo.Password, targetRevision, m.BasePath)
	if err != nil {
		return err
	}

	m.repoMutex.Lock()
	defer m.repoMutex.Unlock()

	m.repoStore[genRepoKey(repoURL, targetRevision)] = repo
	return nil
}

// GetRepo get repo from store
func (m *Manager) GetRepo(repoKey string) (Repo, bool) {
	m.repoMutex.RLock()
	defer m.repoMutex.RUnlock()

	repo, ok := m.repoStore[repoKey]
	return repo, ok
}

func loadEnv() (string, string, string) {
	repoURL := os.Getenv(EnvNameGitRepoURL)
	username := os.Getenv(EnvNameGitUserName)
	secret := os.Getenv(EnvNameGitSecret)

	return repoURL, username, secret
}

// startAutoUpdate 定时轮训所有仓库， 判断仓库下场景是否发生更新
func (m *Manager) startAutoUpdate() {
	scenarioTicker := time.NewTicker(m.ScenarioRefreshFrequency)
	repoTicker := time.NewTicker(m.RepoRefreshFrequency)

	for {
		select {
		case <-scenarioTicker.C:
			blog.Infof("start reload repo...")

			// 遍历repo缓存， 判断仓库下是否有更新
			for key, repo := range m.repoStore {
				blog.Infof("reload repo[%s]...", key)
				changeDirs, err := repo.Reload()
				if err != nil {
					blog.Errorf("repo[%s] reload failed, err: %s", key, err.Error())
					continue
				}

				if len(changeDirs) != 0 {
					blog.Infof("repo[%s] changed, related scenarios: %s", repo.GetRepoKey(),
						utils.ToJsonString(changeDirs))
					if inErr := m.resolveChangeScenario(key, changeDirs); inErr != nil {
						blog.Errorf("repo[%s] resolveChangeScenario failed , err: %s", key, inErr.Error())
						continue
					}
				}
			}
		case <-repoTicker.C:
			blog.Infof("start refresh repo store...")
			usingRepo := make(map[string]struct{})
			//  根据集群内AppMonitor刷新缓存
			appMonitorList := &monitorextensionv1.AppMonitorList{}
			if err := m.Cli.List(context.Background(), appMonitorList); err != nil {
				blog.Errorf("read appmonitor list from apiserver failed, err: %s", err.Error())
				continue
			}
			for _, appMonitor := range appMonitorList.Items {
				usingRepo[GenRepoKeyFromAppMonitor(&appMonitor)] = struct{}{}
			}

			m.repoMutex.Lock()
			for repoKey := range m.repoStore {
				// 默认场景仓库一直保留
				if repoKey == RepoKeyDefault {
					continue
				}
				// 如果repoStore中的仓库没被任何appMonitor使用， 则移除
				_, ok := usingRepo[repoKey]
				if !ok {
					blog.Infof("repo[%s] is not used by any app monitor, released... ", repoKey)
					delete(m.repoStore, repoKey)
				}
			}
			m.repoMutex.Unlock()
		}
	}
}

// resolveChangeScenario repo下场景更新时， 寻找使用对应场景的appMonitor, 通过变更anno的方式触发reconcile
func (m *Manager) resolveChangeScenario(repoKey string, scenarios []string) error {
	for _, scenario := range scenarios {
		selector, err := metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(map[string]string{
			monitorextensionv1.LabelKeyForScenarioName: scenario,
			// monitorextensionv1.LabelKeyForScenarioRepo: repoKey,
		}))
		if err != nil {
			blog.Errorf("generate selector for scenario'%s' failed, err: %s", scenario, err.Error())
			return err
		}
		appMonitorList := &monitorextensionv1.AppMonitorList{}
		if err = m.Cli.List(context.Background(), appMonitorList, &client.ListOptions{
			LabelSelector: selector,
		}); err != nil {
			blog.Errorf("list app monitor for scenario'%s' failed, err: %s", scenario, err.Error())
			return err
		}

		for _, appMonitor := range appMonitorList.Items {
			if GenRepoKeyFromAppMonitor(&appMonitor) != repoKey {
				continue
			}
			// 通过修改注解触发reconcile
			if inErr := utils.PatchAppMonitorAnnotation(context.Background(), m.Cli, &appMonitor, map[string]interface{}{
				monitorextensionv1.AnnotationScenarioUpdateTimestamp: time.Now().Format(time.RFC3339Nano),
			}); inErr != nil {
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
