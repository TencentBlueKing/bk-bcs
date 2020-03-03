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

package main

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store/etcd"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store/zk"
)

type Options struct {
	conf.FileConfig
	conf.ZkConfig
	conf.LogConfig
	KubeConfig string `json:"kubeconfig" value:"" usage:"kube config for custom resource feature and etcd storage"`
}

//bcs version 1.15.x start support etcd store driver
//this tool can migrate data from zk to etcd
//and make sure the data is not lost
//but can't migrate from etcd to zk
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	op := &Options{}
	conf.Parse(op)
	blog.InitLogs(op.LogConfig)

	//connect zk
	dbzk := zk.NewDbZk(strings.Split(op.BCSZk, ","))
	err := dbzk.Connect()
	if err != nil {
		blog.Errorf("connect zookeeper %s failed: %s", op.BCSZk, err.Error())
		os.Exit(1)
	}
	zkStore := zk.NewManagerStore(dbzk)
	zkStore.InitCacheMgr(false)
	blog.Infof("connect zookeeper %s success", op.BCSZk)

	//connect etcd
	etcdStore, err := etcd.NewEtcdStore(op.KubeConfig)
	if err != nil {
		blog.Errorf("new etcd store failed: %s", err.Error())
		os.Exit(1)
	}
	etcdStore.InitCacheMgr(false)
	blog.Infof("connect kube-apiserver %s success", op.KubeConfig)

	//sync framework
	err = syncFramework(zkStore, etcdStore)
	if err != nil {
		blog.Errorf("sync Framework failed: %s", err.Error())
		os.Exit(1)
	}

	//sync application
	err = syncApplication(zkStore, etcdStore)
	if err != nil {
		blog.Errorf("sync application failed: %s", err.Error())
		os.Exit(1)
	}

	//sync agents
	err = syncAgent(zkStore, etcdStore)
	if err != nil {
		blog.Errorf("sync agents failed: %s", err.Error())
		os.Exit(1)
	}

	//sync configmap
	err = syncConfigmap(zkStore, etcdStore)
	if err != nil {
		blog.Errorf("sync configmap failed: %s", err.Error())
		os.Exit(1)
	}

	//sync secret
	err = syncSecret(zkStore, etcdStore)
	if err != nil {
		blog.Errorf("sync secret failed: %s", err.Error())
		os.Exit(1)
	}

	//sync service
	err = syncService(zkStore, etcdStore)
	if err != nil {
		blog.Errorf("sync service failed: %s", err.Error())
		os.Exit(1)
	}

	//sync deployment
	err = syncDeployment(zkStore, etcdStore)
	if err != nil {
		blog.Errorf("sync deployment failed: %s", err.Error())
		os.Exit(1)
	}

	//sync admission
	err = syncAdmission(zkStore, etcdStore)
	if err != nil {
		blog.Errorf("sync admission failed: %s", err.Error())
		os.Exit(1)
	}

	blog.Infof("sync zk data done")
}

func syncFramework(zkStore store.Store, etcdStore store.Store) error {
	blog.Infof("start sync framework data start...")
	//get framework id
	framework, err := zkStore.FetchFrameworkID()
	if err != nil {
		return fmt.Errorf("FetchFrameworkID failed: %s", err.Error())
	}

	//save framework id
	err = etcdStore.SaveFrameworkID(framework)
	if err != nil {
		return fmt.Errorf("SaveFrameworkID %s failed: %s", framework, err.Error())
	}

	//check etcd framework id
	cframework, err := etcdStore.FetchFrameworkID()
	if err != nil {
		return fmt.Errorf("FetchFrameworkID failed: %s", err.Error())
	}
	if framework != cframework {
		return fmt.Errorf("sync framework %s failed", framework)
	}

	blog.Infof("start sync framework data success")
	return nil
}

//Application
//Version
//Taskgroup
//Task
func syncApplication(zkStore store.Store, etcdStore store.Store) error {
	blog.Infof("start sync application data start...")
	//list all applications
	apps, err := zkStore.ListAllApplications()
	if err != nil {
		return fmt.Errorf("ListAllApplications failed: %s", err.Error())
	}

	for _, app := range apps {
		runAs, appId := app.RunAs, app.ID
		//sync version
		//list all versions of application
		versions, err := zkStore.ListVersions(runAs, appId)
		if err != nil {
			return fmt.Errorf("ListVersions(%s:%s) failed: %s", runAs, appId, err.Error())
		}
		for _, no := range versions {
			//fetch version
			version, err := zkStore.FetchVersion(runAs, appId, no)
			if err != nil {
				return fmt.Errorf("FetchVersion(%s:%s:%s) failed: %s", runAs, appId, no, err.Error())
			}
			//save version
			err = etcdStore.UpdateVersion(version)
			if err != nil {
				return fmt.Errorf("SaveVersion(%s:%s:%s) failed: %s", runAs, appId, no, err.Error())
			}
			//check version
			cversion, err := etcdStore.FetchVersion(runAs, appId, no)
			if err != nil {
				return fmt.Errorf("FetchVersion(%s:%s:%s) failed: %s", runAs, appId, no, err.Error())
			}
			if !reflect.DeepEqual(version, cversion) {
				return fmt.Errorf("sync Version(%s:%s:%s) failed", runAs, appId, no)
			}
			blog.Infof("SaveVersion(%s:%s:%s) success", runAs, appId, no)
		}

		//sync application
		err = etcdStore.SaveApplication(app)
		if err != nil {
			return fmt.Errorf("SaveApplication(%s:%s) failed: %s", app.RunAs, app.ID, err.Error())
		}
		capp, err := etcdStore.FetchApplication(runAs, appId)
		if err != nil {
			return fmt.Errorf("FetchApplication(%s:%s) failed: %s", app.RunAs, app.ID, err.Error())
		}
		if !reflect.DeepEqual(app, capp) {
			return fmt.Errorf("sync Application(%s:%s) failed", app.RunAs, app.ID)
		}
		blog.Infof("SaveApplication(%s:%s) success", app.RunAs, app.ID)

		//sync taskgroup
		taskgs, err := zkStore.ListTaskGroups(runAs, appId)
		if err != nil {
			return fmt.Errorf("ListTaskGroups Application(%s:%s) error %s", app.RunAs, app.ID, err.Error())
		}
		for _, taskg := range taskgs {
			err = etcdStore.SaveTaskGroup(taskg)
			if err != nil {
				return fmt.Errorf("SaveTaskGroup(%s) failed: %s", taskg.ID, err.Error())
			}
			ctaskg, err := etcdStore.FetchTaskGroup(taskg.ID)
			if err != nil {
				return fmt.Errorf("FetchTaskGroup(%s) failed: %s", taskg.ID, err.Error())
			}
			if !reflect.DeepEqual(taskg, ctaskg) {
				if app.ID != capp.ID {
					return fmt.Errorf("sync TaskGroup(%s) failed", taskg.ID)
				}
			}
			blog.Infof("SaveTaskGroup(%s) success", taskg.ID)
		}
	}
	blog.Infof("start sync application data success")

	return nil
}

//sync agent
//Agents
//AgentInfo
//AgentSetting
func syncAgent(zkStore store.Store, etcdStore store.Store) error {
	blog.Infof("start sync agents data start...")
	agents, err := zkStore.ListAllAgents()
	if err != nil {
		return fmt.Errorf("ListAllAgents failed: %s", err.Error())
	}
	for _, agent := range agents {
		err = etcdStore.SaveAgent(agent)
		if err != nil {
			return fmt.Errorf("SaveAgent %s failed: %s", agent.Key, err.Error())
		}

		cagent, err := etcdStore.FetchAgent(agent.Key)
		if err != nil {
			return fmt.Errorf("FetchFrameworkID failed: %s", err.Error())
		}
		if !reflect.DeepEqual(agent, cagent) {
			return fmt.Errorf("sync agent %s failed", agent.Key)
		}
		blog.Infof("SaveAgent(%s) success", agent.Key)
	}

	settings, err := zkStore.ListAgentSettingNodes()
	if err != nil {
		return fmt.Errorf("ListAgentSettingNodes failed: %s", err.Error())
	}
	for _, no := range settings {
		setting, err := zkStore.FetchAgentSetting(no)
		if err != nil {
			return fmt.Errorf("FetchAgentSetting %s failed: %s", no, err.Error())
		}
		err = etcdStore.SaveAgentSetting(setting)
		if err != nil {
			return fmt.Errorf("SaveAgentSetting %s failed: %s", no, err.Error())
		}
		csetting, err := etcdStore.FetchAgentSetting(no)
		if err != nil {
			return fmt.Errorf("FetchAgentSetting %s failed: %s", no, err.Error())
		}
		if !reflect.DeepEqual(setting, csetting) {
			return fmt.Errorf("sync AgentSetting %s failed", no)
		}
		blog.Infof("SaveAgentSetting(%s) success", no)
	}

	blog.Infof("start sync agents data success")
	return nil
}

//Configmap
func syncConfigmap(zkStore store.Store, etcdStore store.Store) error {
	blog.Infof("start sync Configmap data start...")
	cfgs, err := zkStore.ListAllConfigmaps()
	if err != nil {
		return fmt.Errorf("ListAllConfigmaps failed: %s", err.Error())
	}
	for _, cfg := range cfgs {
		err = etcdStore.SaveConfigMap(cfg)
		if err != nil {
			return fmt.Errorf("SaveConfigMap(%s:%s) failed: %s", cfg.NameSpace, cfg.Name, err.Error())
		}
		ccfg, err := etcdStore.FetchConfigMap(cfg.NameSpace, cfg.Name)
		if err != nil {
			return fmt.Errorf("FetchConfigMap(%s:%s) failed: %s", cfg.NameSpace, cfg.Name, err.Error())
		}
		if !reflect.DeepEqual(cfg, ccfg) {
			return fmt.Errorf("sync Configmap(%s:%s) failed", cfg.NameSpace, cfg.Name)
		}
		blog.Infof("SaveConfigMap(%s:%s) success", cfg.NameSpace, cfg.Name)
	}

	blog.Infof("start sync Configmap data success")
	return nil
}

//Secret
func syncSecret(zkStore store.Store, etcdStore store.Store) error {
	blog.Infof("start sync Secret data start...")
	scts, err := zkStore.ListAllSecrets()
	if err != nil {
		return fmt.Errorf("ListAllSecrets failed: %s", err.Error())
	}
	for _, sct := range scts {
		if sct.Name == "paas_image_secret" {
			sct.Name = "paas-image-secret"
		}
		err = etcdStore.SaveSecret(sct)
		if err != nil {
			return fmt.Errorf("SaveSecret(%s:%s) failed: %s", sct.NameSpace, sct.Name, err.Error())
		}
		csct, err := etcdStore.FetchSecret(sct.NameSpace, sct.Name)
		if err != nil {
			return fmt.Errorf("FetchSecret(%s:%s) failed: %s", sct.NameSpace, sct.Name, err.Error())
		}
		if !reflect.DeepEqual(sct, csct) {
			return fmt.Errorf("sync Secret(%s:%s) failed", sct.NameSpace, sct.Name)
		}
		blog.Infof("SaveSecret(%s:%s) success", sct.NameSpace, sct.Name)
	}

	blog.Infof("start sync Secret data success")
	return nil
}

//Service
func syncService(zkStore store.Store, etcdStore store.Store) error {
	blog.Infof("start sync Service data start...")
	svcs, err := zkStore.ListAllServices()
	if err != nil {
		return fmt.Errorf("ListAllServices failed: %s", err.Error())
	}
	for _, svc := range svcs {
		err = etcdStore.SaveService(svc)
		if err != nil {
			return fmt.Errorf("SaveService(%s:%s) failed: %s", svc.NameSpace, svc.Name, err.Error())
		}
		csvc, err := etcdStore.FetchService(svc.NameSpace, svc.Name)
		if err != nil {
			return fmt.Errorf("FetchService(%s:%s) failed: %s", svc.NameSpace, svc.Name, err.Error())
		}
		if !reflect.DeepEqual(svc, csvc) {
			return fmt.Errorf("sync Service(%s:%s) failed", svc.NameSpace, svc.Name)
		}
		blog.Infof("SaveService(%s:%s) success", svc.NameSpace, svc.Name)

		end, err := zkStore.FetchEndpoint(svc.NameSpace, svc.Name)
		if err != nil {
			blog.Warnf("FetchEndpoint(%s:%s) failed: %s", svc.NameSpace, svc.Name)
			continue
		}
		err = etcdStore.SaveEndpoint(end)
		if err != nil {
			return fmt.Errorf("SaveEndpoint(%s:%s) failed: %s", end.NameSpace, end.Name, err.Error())
		}
		cend, err := etcdStore.FetchEndpoint(end.NameSpace, end.Name)
		if err != nil {
			return fmt.Errorf("FetchEndpoint(%s:%s) failed: %s", end.NameSpace, end.Name, err.Error())
		}
		if !reflect.DeepEqual(end, cend) {
			return fmt.Errorf("sync Endpoint(%s:%s) failed", end.NameSpace, end.Name)
		}
		blog.Infof("SaveEndpoint(%s:%s) success", end.NameSpace, end.Name)
	}

	blog.Infof("start sync Service data success")
	return nil
}

//Deployment
func syncDeployment(zkStore store.Store, etcdStore store.Store) error {
	blog.Infof("start sync Deployment data start...")
	deployments, err := zkStore.ListAllDeployments()
	if err != nil {
		return fmt.Errorf("ListAllDeployments failed: %s", err.Error())
	}
	for _, deployment := range deployments {
		ns, name := deployment.ObjectMeta.NameSpace, deployment.ObjectMeta.Name
		err = etcdStore.SaveDeployment(deployment)
		if err != nil {
			return fmt.Errorf("SaveDeployment(%s:%s) failed: %s", ns, name, err.Error())
		}
		cdeployment, err := etcdStore.FetchDeployment(ns, name)
		if err != nil {
			return fmt.Errorf("FetchDeployment(%s:%s) failed: %s", ns, name, err.Error())
		}
		if !reflect.DeepEqual(deployment, cdeployment) {
			return fmt.Errorf("sync Deployment(%s:%s) failed", ns, name)
		}
		blog.Infof("SaveDeployment(%s:%s) success", ns, name)
	}

	blog.Infof("start sync Deployment data success")
	return nil
}

//AdminssionWebhooks
func syncAdmission(zkStore store.Store, etcdStore store.Store) error {
	blog.Infof("start sync Admission data start...")
	admissions, err := zkStore.FetchAllAdmissionWebhooks()
	if err != nil {
		return fmt.Errorf("FetchAllAdmissionWebhooks failed: %s", err.Error())
	}
	for _, admission := range admissions {
		ns, name := admission.NameSpace, admission.Name
		err = etcdStore.SaveAdmissionWebhook(admission)
		if err != nil {
			return fmt.Errorf("SaveAdmissionWebhook(%s:%s) failed: %s", ns, name, err.Error())
		}
		cadmission, err := etcdStore.FetchAdmissionWebhook(ns, name)
		if err != nil {
			return fmt.Errorf("FetchAdmissionWebhook(%s:%s) failed: %s", ns, name, err.Error())
		}
		if !reflect.DeepEqual(cadmission, admission) {
			return fmt.Errorf("sync Admission(%s:%s) failed", ns, name)
		}
		blog.Infof("SaveAdmissionWebhook(%s:%s) success", ns, name)
	}

	blog.Infof("start sync Admission data success")
	return nil
}
