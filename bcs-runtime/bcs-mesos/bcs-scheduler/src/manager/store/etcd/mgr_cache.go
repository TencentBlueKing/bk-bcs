/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http:// opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package etcd

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// Cache Manager for store cache
// include: application、namespace、configmap、secret
type cacheManager struct {
	// Current cached application
	// key = {app.RunAs}.{app.ID}
	Applications map[string]*types.Application
	// key = taskgroup.ID
	Taskgroups map[string]*types.TaskGroup
	// key = {version.RunAs}.{version.ID}
	// value is version list
	Versions   map[string]*cacheVersions
	Namespaces map[string]struct{}
	Configmaps map[string]*commtypes.BcsConfigMap
	Secrets    map[string]*commtypes.BcsSecret
	// key = {deployment.namespace}.{deployment.name}
	Deployments map[string]*types.Deployment
	// agent, key = agent.Key
	Agents map[string]*types.Agent
	// services, key = {service.namespace}.{service.name}
	Services map[string]*commtypes.BcsService
	// key=agent.InnerIP
	Agentsettings map[string]*commtypes.BcsClusterAgentSetting
	// key = {daemonset.namespace}.{daemonset.name}
	Daemonsets map[string]*types.BcsDaemonset
	// key = {transaction.namespace}.{transaction.name}
	Transactions map[string]*types.Transaction
	// command
	Commands map[string]*commtypes.BcsCommandInfo
	// Manager currently is OK
	isOK    bool
	mapLock *sync.RWMutex
}

type cacheVersions struct {
	namespace string
	name      string
	objs      []*types.Version
}

var cacheMgr *cacheManager

// Init cache manager, use cache or not
// if isUsed = true, then will use cache manager to improve performance
// else isUsed = false, managerStore not work
func (store *managerStore) InitCacheMgr(isUsed bool) error {
	// etcd store always use cache
	isUsed = true
	if isUsed {
		blog.Infof("init cache begin: use cache for storage")
	} else {
		blog.Infof("init cache begin: not use cache for storage")
	}

	cacheMgr = new(cacheManager)
	cacheMgr.mapLock = new(sync.RWMutex)

	cacheMgr.mapLock.Lock()

	cacheMgr.Applications = make(map[string]*types.Application)
	cacheMgr.Taskgroups = make(map[string]*types.TaskGroup)
	cacheMgr.Versions = make(map[string]*cacheVersions)
	cacheMgr.Namespaces = make(map[string]struct{})
	cacheMgr.Configmaps = make(map[string]*commtypes.BcsConfigMap)
	cacheMgr.Secrets = make(map[string]*commtypes.BcsSecret)
	cacheMgr.Agentsettings = make(map[string]*commtypes.BcsClusterAgentSetting)
	cacheMgr.Deployments = make(map[string]*types.Deployment)
	cacheMgr.Services = make(map[string]*commtypes.BcsService)
	cacheMgr.Agents = make(map[string]*types.Agent)
	cacheMgr.Daemonsets = make(map[string]*types.BcsDaemonset)
	cacheMgr.Commands = make(map[string]*commtypes.BcsCommandInfo)
	cacheMgr.Transactions = make(map[string]*types.Transaction)
	// when isUsed=true, then init cache
	if isUsed {
		// init namespace in cache
		err := store.initCacheNamespaces()
		if err != nil {
			return err
		}
		// init taskgroups in cache
		err = store.initCacheTaskgroups()
		if err != nil {
			return err
		}
		// init versions in cache
		err = store.initCacheVersions()
		if err != nil {
			return err
		}
		// init application in cache
		err = store.initCacheApplications()
		if err != nil {
			return err
		}
		// init configmap in cache
		err = store.initCacheConfigmaps()
		if err != nil {
			return err
		}
		// init secret in cache
		err = store.initCacheSecrets()
		if err != nil {
			return err
		}
		// init agentsetting
		err = store.initCacheAgentsettings()
		if err != nil {
			return err
		}
		// init agent
		err = store.initCacheAgents()
		if err != nil {
			return err
		}
		// init deployment
		err = store.initCacheDeployments()
		if err != nil {
			return err
		}
		// init services
		err = store.initCacheServices()
		if err != nil {
			return err
		}
		// init daemonsets
		err = store.initCacheDaemonsets()
		if err != nil {
			return err
		}
		// init transactions
		err = store.initCacheTransactions()
		if err != nil {
			return err
		}
		// init commands
		err = store.initCacheCommands()
		if err != nil {
			return err
		}
	}
	cacheMgr.isOK = isUsed
	cacheMgr.mapLock.Unlock()
	blog.Infof("init cache end")
	return nil
}

// UnInitCacheMgr unint cache manager
// clear all cache
func (store *managerStore) UnInitCacheMgr() error {

	blog.Infof("uninit cache begin")

	if cacheMgr != nil {
		cacheMgr.mapLock.Lock()

		cacheMgr.Applications = nil
		cacheMgr.Namespaces = nil
		cacheMgr.Configmaps = nil
		cacheMgr.Secrets = nil
		cacheMgr.mapLock = nil
		cacheMgr.isOK = false
		cacheMgr.mapLock.Unlock()
	}
	cacheMgr = nil

	blog.Infof("uninit cache end")
	return nil
}

// init namespace in cache
func (store *managerStore) initCacheNamespaces() error {
	nss, err := store.ListRunAs()
	if err != nil {
		blog.Errorf("cacheManager init namespace failed: %s", err.Error())
		return err
	}

	for _, ns := range nss {
		cacheMgr.Namespaces[ns] = struct{}{}
		blog.V(3).Infof("cacheManager sync namespace %s in cache", ns)
	}
	blog.Infof("cacheMgr init cache namespaces done")
	return nil
}

// init configmaps in cache
func (store *managerStore) initCacheConfigmaps() error {
	cfgs, err := store.ListAllConfigmaps()
	if err != nil {
		blog.Errorf("cacheManager init configmaps failed: %s", err.Error())
		return err
	}

	for _, cfg := range cfgs {
		key := fmt.Sprintf("%s.%s", cfg.NameSpace, cfg.Name)
		cacheMgr.Configmaps[key] = cfg.DeepCopy()
		blog.V(3).Infof("cacheManager sync configmap %s in cache", key)
	}
	blog.Infof("cacheMgr init cache configmaps done")
	return nil
}

// init secret in cache
func (store *managerStore) initCacheSecrets() error {
	secs, err := store.ListAllSecrets()
	if err != nil {
		blog.Errorf("cacheManager init secrets failed: %s", err.Error())
		return err
	}

	for _, sec := range secs {
		key := fmt.Sprintf("%s.%s", sec.NameSpace, sec.Name)
		cacheMgr.Secrets[key] = sec.DeepCopy()
		blog.V(3).Infof("cacheManager sync secret %s in cache", key)
	}
	blog.Infof("cacheMgr init cache secrets done")
	return nil
}

// init deployment in cache
func (store *managerStore) initCacheDeployments() error {
	deps, err := store.ListAllDeployments()
	if err != nil {
		blog.Errorf("cacheManager ListAllDeployments failed: %s", err.Error())
		return err
	}

	for _, dep := range deps {
		key := fmt.Sprintf("%s.%s", dep.ObjectMeta.NameSpace, dep.ObjectMeta.Name)
		cacheMgr.Deployments[key] = dep.DeepCopy()
		blog.V(3).Infof("cacheManager sync deployment %s in cache", key)
	}
	blog.Infof("cacheMgr init cache deployments done")
	return nil
}

// init agent in cache
func (store *managerStore) initCacheAgents() error {
	agents, err := store.ListAllAgents()
	if err != nil {
		blog.Errorf("cacheManager ListAllAgents failed: %s", err.Error())
		return err
	}

	for _, agent := range agents {
		cacheMgr.Agents[agent.Key] = agent.DeepCopy()
		blog.V(3).Infof("cacheManager sync agent %s in cache", agent.Key)
	}
	blog.Infof("cacheMgr init cache agents done")
	return nil
}

// init services in cache
func (store *managerStore) initCacheServices() error {
	svcs, err := store.ListAllServices()
	if err != nil {
		blog.Errorf("cacheManager ListAllServices failed: %s", err.Error())
		return err
	}

	for _, svc := range svcs {
		key := fmt.Sprintf("%s.%s", svc.NameSpace, svc.Name)
		cacheMgr.Services[key] = svc.DeepCopy()
		blog.V(3).Infof("cacheManager sync service %s in cache", key)
	}
	blog.Infof("cacheMgr init cache services done")
	return nil
}

// init daemonsets in cache
func (store *managerStore) initCacheDaemonsets() error {
	dms, err := store.ListAllDaemonset()
	if err != nil {
		blog.Errorf("cacheManager ListAllDaemonsets failed: %s", err.Error())
		return err
	}

	for _, dm := range dms {
		cacheMgr.Daemonsets[dm.GetUuid()] = dm.DeepCopy()
	}
	blog.Infof("cacheMgr init cache daemonsets done")
	return nil
}

// init transaction in cache
func (store *managerStore) initCacheTransactions() error {
	transs, err := store.ListAllTransaction()
	if err != nil {
		blog.Errorf("cacheManager ListAllTransaction failed, err %s", err.Error())
		return err
	}

	for _, trans := range transs {
		cacheMgr.Transactions[trans.GetUuid()] = trans.DeepCopy()
	}
	blog.Infof("cacheMgr init cache transaction done")
	return nil
}

// init commands in cache
func (store *managerStore) initCacheCommands() error {
	cmds, err := store.listAllCommands()
	if err != nil {
		blog.Errorf("cacheManager listAllCommands failed: %s", err.Error())
		return err
	}

	for _, cmd := range cmds {
		cacheMgr.Commands[cmd.Id] = cmd.DeepCopy()
	}
	blog.Infof("cacheMgr init cache commands done")
	return nil
}

// init agentsetting in cache
func (store *managerStore) initCacheAgentsettings() error {
	agents, err := store.ListAgentsettings()
	if err != nil {
		blog.Errorf("cacheManager ListAgentsettings failed: %s", err.Error())
		return err
	}

	for _, agent := range agents {
		cacheMgr.Agentsettings[agent.InnerIP] = agent.DeepCopy()
		blog.V(3).Infof("cacheManager sync agentsettings %s in cache", agent.InnerIP)
	}
	blog.Infof("cacheMgr init cache agentsettings done")
	return nil
}

// init application in cache
func (store *managerStore) initCacheApplications() error {
	apps, err := store.ListAllApplications()
	if err != nil {
		blog.Errorf("cacheManager init application failed: %s", err.Error())
		return err
	}

	for _, app := range apps {
		cacheMgr.Applications[app.GetUuid()] = app.DeepCopy()
		blog.V(3).Infof("cacheManager sync application %s in cache", app.GetUuid())
	}
	blog.Infof("cacheMgr init cache application done")
	return nil
}

// init application in cache
func (store *managerStore) initCacheVersions() error {
	// cache versions
	versions, err := store.listClusterVersions()
	if err != nil {
		return err
	}
	for _, version := range versions {
		blog.V(3).Infof("cacheManager instance(%s:%s) sync version(%s) in cache", version.RunAs, version.ID, version.Name)
		tmpData := version.DeepCopy()
		vns, ok := cacheMgr.Versions[fmt.Sprintf("%s.%s", version.RunAs, version.ID)]
		if !ok {
			vns = &cacheVersions{
				namespace: version.RunAs,
				name:      version.ID,
				objs:      make([]*types.Version, 0),
			}
			cacheMgr.Versions[fmt.Sprintf("%s.%s", version.RunAs, version.ID)] = vns
		}
		vns.objs = append(vns.objs, tmpData)
	}
	blog.Infof("cacheMgr init cache versions done")
	return nil
}

func (store *managerStore) initCacheTaskgroups() error {
	taskgroups, err := store.listTaskgroupsInDB()
	if err != nil {
		blog.Errorf("cacheManager init taskgroup failed: %s", err.Error())
		return err
	}

	for _, taskgroup := range taskgroups {
		cacheMgr.Taskgroups[taskgroup.ID] = taskgroup.DeepCopy()
		blog.V(3).Infof("cacheManager sync taskgroup %s in cache", taskgroup.ID)
	}
	blog.Infof("cacheMgr init cache application done")
	return nil
}

// ns = namespcae,
// if exist, then return true
// else return false
func checkCacheNamespaceExist(ns string) bool {
	cacheMgr.mapLock.RLock()
	_, ok := cacheMgr.Namespaces[ns]
	cacheMgr.mapLock.RUnlock()
	return ok
}

// cache namespace in cache
func syncCacheNamespace(ns string) {
	cacheMgr.mapLock.Lock()
	cacheMgr.Namespaces[ns] = struct{}{}
	cacheMgr.mapLock.Unlock()
}

// delete application in cache
func deleteAppCacheNode(runAs, appID string) error {
	key := runAs + "." + appID
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Applications, key)
	cacheMgr.mapLock.Unlock()
	return nil
}

// save configmap in cache
func saveCacheConfigmap(obj *commtypes.BcsConfigMap) error {
	key := fmt.Sprintf("%s.%s", obj.NameSpace, obj.Name)
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Configmaps[key] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

// get configmap in cache
// if not exist, then return nil
func getCacheConfigmap(ns, name string) *commtypes.BcsConfigMap {
	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Configmaps[key]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	return obj.DeepCopy()
}

// delete configmap in cache
func deleteCacheConfigmap(ns, name string) error {
	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Configmaps, key)
	cacheMgr.mapLock.Unlock()
	return nil
}

// save command in cache
func saveCacheCommand(obj *commtypes.BcsCommandInfo) error {
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Commands[obj.Id] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

// get command in cache
// if not exist, then return nil
func getCacheCommand(key string) *commtypes.BcsCommandInfo {
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Commands[key]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	return obj.DeepCopy()
}

// delete command in cache
func deleteCacheCommand(key string) error {
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Commands, key)
	cacheMgr.mapLock.Unlock()
	return nil
}

// save daemonset in cache
func saveCacheDaemonset(obj *types.BcsDaemonset) error {
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Daemonsets[obj.GetUuid()] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

// get daemonset in cache
// if not exist, then return nil
func getCacheDaemonset(ns, name string) *types.BcsDaemonset {
	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Daemonsets[key]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	return obj.DeepCopy()
}

// delete daemonset in cache
func deleteCacheDaemonset(ns, name string) error {
	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Daemonsets, key)
	cacheMgr.mapLock.Unlock()
	return nil
}

// save secret in cache
func saveCacheSecret(obj *commtypes.BcsSecret) error {
	key := fmt.Sprintf("%s.%s", obj.NameSpace, obj.Name)
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Secrets[key] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

// get secret in cache
func getCacheSecret(ns, name string) *commtypes.BcsSecret {
	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Secrets[key]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	return obj.DeepCopy()
}

// delete secret in cache
func deleteCacheSecret(ns, name string) error {
	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Secrets, key)
	cacheMgr.mapLock.Unlock()
	return nil
}

// save agentsettings in cache
func saveCacheAgentsetting(obj *commtypes.BcsClusterAgentSetting) error {
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Agentsettings[tmpData.InnerIP] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

// get Agentsetting in cache
func getCacheAgentsetting(InnerIp string) *commtypes.BcsClusterAgentSetting {
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Agentsettings[InnerIp]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	return obj.DeepCopy()
}

// delete agentsetting in cache
func deleteCacheAgentsetting(InnerIp string) error {
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Agentsettings, InnerIp)
	cacheMgr.mapLock.Unlock()
	return nil
}

// save version in cache
func saveCacheVersion(runAs, appID string, obj *types.Version) error {
	cacheMgr.mapLock.Lock()
	vns, ok := cacheMgr.Versions[fmt.Sprintf("%s.%s", runAs, appID)]
	if !ok {
		vns = &cacheVersions{
			namespace: runAs,
			name:      appID,
			objs:      make([]*types.Version, 0),
		}
		cacheMgr.Versions[fmt.Sprintf("%s.%s", runAs, appID)] = vns
	}
	cacheMgr.mapLock.Unlock()
	tmpData := obj.DeepCopy()
	vns.objs = append(vns.objs, tmpData)
	return nil
}

// get version from cache
func getCacheVersion(runAs, versionId, versionNo string) (*types.Version, error) {
	cacheMgr.mapLock.RLock()
	versions, ok := cacheMgr.Versions[fmt.Sprintf("%s.%s", runAs, versionId)]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil, nil
	}
	for _, version := range versions.objs {
		if version.Name == versionNo {
			return version.DeepCopy(), nil
		}
	}

	return nil, nil
}

// list version from cache
// runAs=namespace, versionId=appid
func listCacheVersions(runAs, versionId string) ([]*types.Version, error) {
	cacheMgr.mapLock.RLock()
	iversions, ok := cacheMgr.Versions[fmt.Sprintf("%s.%s", runAs, versionId)]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		blog.Warnf("listCache versions(%s) is empty", fmt.Sprintf("%s.%s", runAs, versionId))
		return nil, nil
	}

	var versions []*types.Version
	for _, version := range iversions.objs {
		tmpData := version.DeepCopy()
		versions = append(versions, tmpData)
	}
	return versions, nil
}

// delete version in cache
func deleteCacheVersion(runAs, versionId string) error {
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Versions, fmt.Sprintf("%s.%s", runAs, versionId))
	blog.Infof("delete cache version(%s)", fmt.Sprintf("%s.%s", runAs, versionId))
	cacheMgr.mapLock.Unlock()
	return nil
}

// save application in cache
func saveCacheApplication(runAs, appID string, obj *types.Application) error {
	app := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Applications[app.GetUuid()] = app
	cacheMgr.mapLock.Unlock()

	return nil
}

// get application from cache
func getCacheApplication(runAs, appID string) (*types.Application, error) {
	cacheMgr.mapLock.RLock()
	app, ok := cacheMgr.Applications[fmt.Sprintf("%s.%s", runAs, appID)]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil, nil
	}

	return app.DeepCopy(), nil
}

// list cache applications
func listCacheApplications() ([]*types.Application, error) {
	cacheMgr.mapLock.RLock()
	apps := make([]*types.Application, 0, len(cacheMgr.Applications))
	for _, node := range cacheMgr.Applications {
		apps = append(apps, node.DeepCopy())
	}
	cacheMgr.mapLock.RUnlock()

	return apps, nil
}

func listCacheRunAsApplications(runAs string) ([]*types.Application, error) {
	cacheMgr.mapLock.RLock()
	apps := make([]*types.Application, 0)
	for _, node := range cacheMgr.Applications {
		if node.RunAs == runAs {
			apps = append(apps, node.DeepCopy())
		}
	}
	cacheMgr.mapLock.RUnlock()

	return apps, nil
}

func listCacheRunAsDeployment(namespace string) ([]*types.Deployment, error) {
	cacheMgr.mapLock.RLock()
	deps := make([]*types.Deployment, 0)
	for _, dep := range cacheMgr.Deployments {
		if dep.ObjectMeta.NameSpace == namespace {
			deps = append(deps, dep.DeepCopy())
		}
	}
	cacheMgr.mapLock.RUnlock()

	return deps, nil
}

func listCacheConfigmaps() ([]*commtypes.BcsConfigMap, error) {
	cacheMgr.mapLock.RLock()
	cfgs := make([]*commtypes.BcsConfigMap, 0, len(cacheMgr.Configmaps))
	for _, cfg := range cacheMgr.Configmaps {
		cfgs = append(cfgs, cfg.DeepCopy())
	}
	cacheMgr.mapLock.RUnlock()

	return cfgs, nil
}

// list all daemonsets from cache
func listCacheDaemonsets() ([]*types.BcsDaemonset, error) {
	cacheMgr.mapLock.RLock()
	dms := make([]*types.BcsDaemonset, 0, len(cacheMgr.Daemonsets))
	for _, dm := range cacheMgr.Daemonsets {
		dms = append(dms, dm.DeepCopy())
	}
	cacheMgr.mapLock.RUnlock()

	return dms, nil
}

func listCacheSecrets() ([]*commtypes.BcsSecret, error) {
	cacheMgr.mapLock.RLock()
	secs := make([]*commtypes.BcsSecret, 0, len(cacheMgr.Secrets))
	for _, sec := range cacheMgr.Secrets {
		secs = append(secs, sec.DeepCopy())
	}
	cacheMgr.mapLock.RUnlock()

	return secs, nil
}

// save taskgroup in cache
func saveCacheTaskGroup(taskGroup *types.TaskGroup) error {
	obj := taskGroup.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Taskgroups[obj.ID] = obj
	cacheMgr.mapLock.Unlock()

	return nil
}

// get taskgroup from cache
func fetchCacheTaskGroup(taskGroupID string) (*types.TaskGroup, error) {
	cacheMgr.mapLock.RLock()
	taskgroup, ok := cacheMgr.Taskgroups[taskGroupID]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil, nil
	}

	return taskgroup.DeepCopy(), nil
}

// delete taskgroup in cache
func deleteCacheTaskGroup(taskGroupID string) error {
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Taskgroups, taskGroupID)
	cacheMgr.mapLock.Unlock()

	return nil
}

// save task in cache
func saveCacheTask(task *types.Task) error {
	taskGroupID := types.GetTaskGroupID(task.ID)
	if taskGroupID == "" {
		return errors.New("task id error")
	}
	cacheMgr.mapLock.RLock()
	taskGroup, ok := cacheMgr.Taskgroups[taskGroupID]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}
	cacheData := task.DeepCopy()
	isExist := false
	for index, cacheTask := range taskGroup.Taskgroup {
		if task.ID == cacheTask.ID {
			taskGroup.Taskgroup[index] = cacheData
			isExist = true
		}
	}

	if !isExist {
		blog.Warnf("insert task(%s) in cache", task.ID)
		taskGroup.Taskgroup = append(taskGroup.Taskgroup, cacheData)
	}

	return nil
}

// delete task in cache
func deleteCacheTask(taskId string) error {
	taskGroupID := types.GetTaskGroupID(taskId)
	if taskGroupID == "" {
		return errors.New("task id error")
	}
	cacheMgr.mapLock.RLock()
	taskGroup, ok := cacheMgr.Taskgroups[taskGroupID]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	for index, cacheTask := range taskGroup.Taskgroup {
		if taskId == cacheTask.ID {
			blog.Infof("delete task(%s) in cache", taskId)
			taskGroup.Taskgroup = append(taskGroup.Taskgroup[:index], taskGroup.Taskgroup[index+1:]...)
		}
	}
	return nil
}

// get task from cache
func fetchCacheTask(taskId string) (*types.Task, error) {
	taskGroupID := types.GetTaskGroupID(taskId)
	if taskGroupID == "" {
		return nil, errors.New("task id error")
	}
	cacheMgr.mapLock.RLock()
	taskGroup, ok := cacheMgr.Taskgroups[taskGroupID]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil, nil
	}

	for _, cacheTask := range taskGroup.Taskgroup {
		if taskId == cacheTask.ID {
			tmpData := cacheTask.DeepCopy()
			return tmpData, nil
		}
	}
	blog.Warnf("fetch task(%s) in cache return nil", taskId)
	return nil, nil
}

// save deployment in cache
func saveCacheDeployment(obj *types.Deployment) error {
	key := fmt.Sprintf("%s.%s", obj.ObjectMeta.NameSpace, obj.ObjectMeta.Name)
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Deployments[key] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

// get deployment in cache
// if not exist, then return nil
func getCacheDeployment(ns, name string) *types.Deployment {
	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Deployments[key]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	return obj.DeepCopy()
}

// delete deployment in cache
func deleteCacheDeployment(ns, name string) error {
	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Deployments, key)
	cacheMgr.mapLock.Unlock()
	return nil
}

func listCacheDeployments() ([]*types.Deployment, error) {
	cacheMgr.mapLock.RLock()
	deps := make([]*types.Deployment, 0, len(cacheMgr.Deployments))
	for _, dep := range cacheMgr.Deployments {
		deps = append(deps, dep.DeepCopy())
	}
	cacheMgr.mapLock.RUnlock()

	return deps, nil
}

// save agent in cache
func saveCacheAgent(obj *types.Agent) error {
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Agents[obj.Key] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

// get agent in cache
// if not exist, then return nil
// key = agent.key
func getCacheAgent(key string) *types.Agent {
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Agents[key]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	return obj.DeepCopy()
}

// delete agent in cache, key = agent.key
func deleteCacheAgent(key string) error {
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Agents, key)
	cacheMgr.mapLock.Unlock()
	return nil
}

func listCacheAgents() ([]*types.Agent, error) {
	cacheMgr.mapLock.RLock()
	agents := make([]*types.Agent, 0, len(cacheMgr.Agents))
	for _, agent := range cacheMgr.Agents {
		agents = append(agents, agent.DeepCopy())
	}
	cacheMgr.mapLock.RUnlock()

	return agents, nil
}

func listCacheAgentsettings() ([]*commtypes.BcsClusterAgentSetting, error) {
	cacheMgr.mapLock.RLock()
	agents := make([]*commtypes.BcsClusterAgentSetting, 0, len(cacheMgr.Agentsettings))
	for _, agent := range cacheMgr.Agentsettings {
		agents = append(agents, agent.DeepCopy())
	}
	cacheMgr.mapLock.RUnlock()

	return agents, nil
}

// save service in cache
func saveCacheService(obj *commtypes.BcsService) error {
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Services[fmt.Sprintf("%s.%s", obj.NameSpace, obj.Name)] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

// get service in cache
// if not exist, then return nil
func getCacheService(ns, name string) *commtypes.BcsService {
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Services[fmt.Sprintf("%s.%s", ns, name)]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	return obj.DeepCopy()
}

// delete service in cache
func deleteCacheService(ns, name string) error {
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Services, fmt.Sprintf("%s.%s", ns, name))
	cacheMgr.mapLock.Unlock()
	return nil
}

func listCacheServices() ([]*commtypes.BcsService, error) {
	cacheMgr.mapLock.RLock()
	svcs := make([]*commtypes.BcsService, 0, len(cacheMgr.Services))
	for _, svc := range cacheMgr.Services {
		svcs = append(svcs, svc.DeepCopy())
	}
	cacheMgr.mapLock.RUnlock()

	return svcs, nil
}

func listCacheTaskgroups() ([]*types.TaskGroup, error) {
	cacheMgr.mapLock.RLock()
	taskgroups := make([]*types.TaskGroup, 0, len(cacheMgr.Taskgroups))
	for _, o := range cacheMgr.Taskgroups {
		taskgroups = append(taskgroups, o.DeepCopy())
	}
	cacheMgr.mapLock.RUnlock()

	return taskgroups, nil
}

func saveCacheTransaction(obj *types.Transaction) error {
	if obj == nil {
		return fmt.Errorf("transaction to be saved into cache cannot be empty")
	}
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Transactions[obj.GetUuid()] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

func getCacheTransaction(ns, name string) *types.Transaction {
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Transactions[fmt.Sprintf("%s/%s", ns, name)]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}
	return obj.DeepCopy()
}

func deleteCacheTransaction(ns, name string) error {
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Transactions, fmt.Sprintf("%s/%s", ns, name))
	cacheMgr.mapLock.Unlock()
	return nil
}

func listCacheTransactions() ([]*types.Transaction, error) {
	cacheMgr.mapLock.RLock()
	trans := make([]*types.Transaction, 0, len(cacheMgr.Transactions))
	for _, tran := range cacheMgr.Transactions {
		trans = append(trans, tran.DeepCopy())
	}
	cacheMgr.mapLock.RUnlock()
	return trans, nil
}

func listCacheRunAsTransaction(namespace string) ([]*types.Transaction, error) {
	cacheMgr.mapLock.RLock()
	tranList := make([]*types.Transaction, 0)
	for _, trans := range cacheMgr.Transactions {
		if trans.Namespace == namespace {
			tranList = append(tranList, trans.DeepCopy())
		}
	}
	cacheMgr.mapLock.RUnlock()

	return tranList, nil
}
