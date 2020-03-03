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

package etcd

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"bk-bcs/bcs-mesos/pkg/client/internalclientset"
	bkbcsv2 "bk-bcs/bcs-mesos/pkg/client/internalclientset/typed/bkbcs/v2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	extensionClientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/internalclientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ApiversionV2     = "v2"
	DefaultNamespace = "bkbcs"
)

const (
	CrdAdmissionWebhookConfiguration = "AdmissionWebhookConfiguration"
	CrdAgent                         = "Agent"
	CrdAgentSetting                  = "BcsClusterAgentSetting"
	CrdAgentSchedInfo                = "AgentSchedInfo"
	CrdApplication                   = "Application"
	CrdBcsCommandInfo                = "BcsCommandInfo"
	CrdBcsConfigMap                  = "BcsConfigMap"
	CrdCrr                           = "Crr"
	CrdCrd                           = "Crd"
	CrdDeployment                    = "Deployment"
	CrdBcsEndpoint                   = "BcsEndpoint"
	CrdFramework                     = "Framework"
	CrdBcsSecret                     = "BcsSecret"
	CrdBcsService                    = "BcsService"
	CrdTask                          = "Task"
	CrdTaskGroup                     = "TaskGroup"
	CrdVersion                       = "Version"
)

const (
	// Default namespace
	defaultRunAs string = "defaultGroup"
	//object label's key or value max length 63
	LabelKVMaxLength = 63
)

// Store Manager
type managerStore struct {
	BkbcsClient     bkbcsv2.BkbcsV2Interface
	k8sClient       *kubernetes.Clientset
	extensionClient *extensionClientset.Clientset

	regkey   *regexp.Regexp
	regvalue *regexp.Regexp

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *managerStore) initKubeCrd() error {
	crds := []string{
		CrdAdmissionWebhookConfiguration,
		CrdAgent,
		CrdAgentSetting,
		CrdAgentSchedInfo,
		CrdApplication,
		CrdBcsCommandInfo,
		CrdBcsConfigMap,
		CrdCrr,
		CrdCrd,
		CrdDeployment,
		CrdBcsEndpoint,
		CrdFramework,
		CrdBcsSecret,
		CrdBcsService,
		CrdTask,
		CrdTaskGroup,
		CrdVersion,
	}

	for _, crd := range crds {
		client := s.extensionClient.Apiextensions().CustomResourceDefinitions()

		crd := &apiextensions.CustomResourceDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: "v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: strings.ToLower(fmt.Sprintf("%ss.bkbcs.tencent.com", crd)),
			},
			Spec: apiextensions.CustomResourceDefinitionSpec{
				Group: "bkbcs.tencent.com",
				Names: apiextensions.CustomResourceDefinitionNames{
					Kind:     crd,
					Plural:   strings.ToLower(fmt.Sprintf("%ss", crd)),
					ListKind: fmt.Sprintf("%sList", crd),
				},
				Scope: apiextensions.NamespaceScoped,
				Versions: []apiextensions.CustomResourceDefinitionVersion{
					{
						Name:    ApiversionV2,
						Served:  true,
						Storage: true,
					},
				},
			},
		}

		_, err := client.Create(crd)
		if err != nil && !errors.IsAlreadyExists(err) {
			blog.Errorf("etcdstore register Crds failed:%s", err.Error())
			return err
		}
	}

	return nil
}

func (s *managerStore) StopStoreMetrics() {
	if s.cancel == nil {
		return
	}
	s.cancel()

	time.Sleep(time.Second)
	s.wg.Wait()
}

func (s *managerStore) StartStoreObjectMetrics() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	for {
		time.Sleep(time.Minute)

		select {
		case <-s.ctx.Done():
			blog.Infof("stop scheduler store metrics")
			return

		default:
			s.wg.Add(1)
			store.ObjectResourceInfo.Reset()
		}

		// handle service metrics
		services, err := s.ListAllServices()
		if err != nil {
			blog.Errorf("list all services error %s", err.Error())
		}
		for _, service := range services {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceService, service.NameSpace, service.Name, "")
		}

		// handle application metrics
		apps, err := s.ListAllApplications()
		if err != nil {
			blog.Errorf("list all applications error %s", err.Error())
		}
		for _, app := range apps {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceApplication, app.RunAs, app.Name, app.Status)

			// handle taskgroup metrics
			taskgroups, err := s.ListTaskGroups(app.RunAs, app.Name)
			if err != nil {
				blog.Errorf("list all services error %s", err.Error())
			}
			for _, taskgroup := range taskgroups {
				store.ReportTaskgroupInfoMetrics(taskgroup.RunAs, taskgroup.AppID, taskgroup.ID, taskgroup.Status)
			}
		}

		// handle deployment metrics
		deployments, err := s.ListAllDeployments()
		if err != nil {
			blog.Errorf("list all deployment error %s", err.Error())
		}
		for _, deployment := range deployments {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceDeployment, deployment.ObjectMeta.NameSpace, deployment.ObjectMeta.Name, "")
		}

		// handle configmap metrics
		configmaps, err := s.ListAllConfigmaps()
		if err != nil {
			blog.Errorf("list all configmap error %s", err.Error())
		}
		for _, configmap := range configmaps {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceConfigmap, configmap.NameSpace, configmap.Name, "")
		}

		// handle secrets metrics
		secrets, err := s.ListAllConfigmaps()
		if err != nil {
			blog.Errorf("list all secret error %s", err.Error())
		}
		for _, secret := range secrets {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceSecret, secret.NameSpace, secret.Name, "")
		}

		// handle agents metrics
		agents, err := s.ListAllAgents()
		if err != nil {
			blog.Errorf("list all agent error %s", err.Error())
		}

		for _, agent := range agents {
			info := agent.GetAgentInfo()
			if info.IP == "" {
				blog.Errorf("agent %s don't have InnerIP attribute", agent.Key)
				continue
			}

			store.ReportAgentInfoMetrics(info.IP, info.CpuTotal, info.CpuTotal-info.CpuUsed,
				info.MemTotal, info.MemTotal-info.MemUsed)
		}

		s.wg.Done()
	}
}

func NewEtcdStore(kubeconfig string) (store.Store, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		blog.Errorf("etcdstore build kubeconfig %s error %s", kubeconfig, err.Error())
		return nil, err
	}
	blog.Infof("etcdstore build kubeconfig %s success", kubeconfig)

	k8sClientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("etcdstore build clientset error %s", err.Error())
		return nil, err
	}

	extensionClient, err := extensionClientset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("etcdstore build clientset error %s", err.Error())
		return nil, err
	}

	bkbcsClientset, err := internalclientset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("etcdstore build clientset error %s", err.Error())
		return nil, err
	}

	m := &managerStore{
		BkbcsClient:     bkbcsClientset.BkbcsV2(),
		k8sClient:       k8sClientset,
		extensionClient: extensionClient,
	}

	m.regkey, _ = regexp.Compile("^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$")
	m.regvalue, _ = regexp.Compile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$")

	//init clientset crds
	err = m.initKubeCrd()
	if err != nil {
		return nil, err
	}

	//init default namespace
	err = m.checkNamespace(DefaultNamespace)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (store *managerStore) checkNamespace(ns string) error {
	if cacheMgr != nil && cacheMgr.isOK {
		exist := checkCacheNamespaceExist(ns)
		if exist {
			blog.V(3).Infof("check namespace %s exist", ns)
			return nil
		}
	}

	client := store.k8sClient.CoreV1().Namespaces()
	_, err := client.Get(ns, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		blog.Warnf("clientset namespace %s %s", ns, err.Error())
		ns := &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}
		_, err = client.Create(ns)
		if err != nil {
			return err
		}

		syncCacheNamespace(ns.Name)
		return nil
	}

	return nil
}

func (store *managerStore) ListRunAs() ([]string, error) {
	client := store.k8sClient.CoreV1().Namespaces()
	nss, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	runAses := make([]string, 0, len(nss.Items))
	for _, ns := range nss.Items {
		runAses = append(runAses, ns.Name)
	}

	return runAses, nil
}

func (store *managerStore) ListDeploymentRunAs() ([]string, error) {

	return store.ListRunAs()
}

func (store *managerStore) filterSpecialLabels(oriLabels map[string]string) map[string]string {
	if oriLabels == nil {
		return nil
	}

	labels := make(map[string]string)
	for k, v := range oriLabels {
		if !store.regkey.MatchString(k) {
			continue
		}
		if !store.regvalue.MatchString(v) {
			continue
		}
		if len(k) > LabelKVMaxLength || len(v) > LabelKVMaxLength {
			continue
		}

		labels[k] = v
	}
	return labels
}
