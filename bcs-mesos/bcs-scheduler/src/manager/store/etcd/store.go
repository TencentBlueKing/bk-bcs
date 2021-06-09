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
	"context"
	goerrors "errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	typesplugin "github.com/Tencent/bk-bcs/bcs-common/common/plugin"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/pluginManager"
	internalclientset "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/clientset/versioned"
	bkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/clientset/versioned/typed/bkbcs/v2"

	corev1 "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extensionClientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// ApiversionV2 mesos crd api version
	ApiversionV2 = "v2"
	// DefaultNamespace is mesos crd default namespace
	DefaultNamespace = "bkbcs"
)

const (
	// ObjectVersionNotLatestError error message when object version is not latest
	ObjectVersionNotLatestError = "please apply your changes to the latest version and try again"
)

// bcs mesos custom resources list
const (
	// CrdAdmissionWebhookConfiguration mesos admission webhook configuration crd name
	CrdAdmissionWebhookConfiguration = "AdmissionWebhookConfiguration"
	// CrdAgent mesos agent crd name
	CrdAgent = "Agent"
	// CrdAgentSetting mesos agent setting crd name
	CrdAgentSetting = "BcsClusterAgentSetting"
	// CrdAgentSchedInfo mesos agent scheduled info crd name
	CrdAgentSchedInfo = "AgentSchedInfo"
	// CrdApplication mesos application crd name
	CrdApplication = "Application"
	// CrdBcsCommandInfo bcs command info crd name
	CrdBcsCommandInfo = "BcsCommandInfo"
	// CrdBcsConfigMap bcs configmap crd name
	CrdBcsConfigMap = "BcsConfigMap"
	// CrdCrr mesos custom resource register crd name
	CrdCrr = "Crr"
	// CrdCrd mesos custom resource definition crd name
	CrdCrd = "Crd"
	// CrdDeployment mesos deployment crd name
	CrdDeployment = "Deployment"
	// CrdBcsEndpoint mesos endpoint crd name
	CrdBcsEndpoint = "BcsEndpoint"
	// CrdFramework mesos framework crd name
	CrdFramework = "Framework"
	// CrdBcsSecret mesos secret crd name
	CrdBcsSecret = "BcsSecret"
	// CrdBcsService mesos service crd name
	CrdBcsService = "BcsService"
	// CrdTask mesos task crd name
	CrdTask = "Task"
	// CrdTaskGroup mesos taskgroup crd name
	CrdTaskGroup = "TaskGroup"
	// CrdVersion mesos version crd name
	CrdVersion = "Version"
	// CrdBcsDaemonset mesos daemonset crd name
	CrdBcsDaemonset = "BcsDaemonset"
	// CrdBcsTransaction mesos transaction crd name
	CrdBcsTransaction = "BcsTransaction"
)

const (
	// Default namespace
	defaultRunAs string = "defaultGroup"
	// LabelKVMaxLength object label's key or value max length 63
	LabelKVMaxLength = 63
)

// Store Manager
type managerStore struct {
	BkbcsClient     bkbcsv2.BkbcsV2Interface
	k8sClient       *kubernetes.Clientset
	extensionClient *extensionClientset.Clientset

	regkey   *regexp.Regexp
	regvalue *regexp.Regexp

	ctx    context.Context
	cancel context.CancelFunc

	// plugin manager, ip-resources
	pm        *pluginManager.PluginManager
	clusterId string
}

// init bcs mesos custom resources
// connect kube-apiserver, and create custom resources definition
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
		CrdBcsDaemonset,
		CrdBcsTransaction,
	}

	for _, crd := range crds {
		client := s.extensionClient.ApiextensionsV1beta1().CustomResourceDefinitions()

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
		// create crd definition
		_, err := client.Create(context.Background(), crd, metav1.CreateOptions{})
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
}

// store metrics report prometheus
func (s *managerStore) StartStoreObjectMetrics() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	for {
		time.Sleep(time.Minute)
		if cacheMgr == nil || !cacheMgr.isOK {
			continue
		}
		blog.Infof("start produce metrics")
		store.ObjectResourceInfo.Reset()
		store.TaskgroupInfo.Reset()
		store.AgentCpuResourceRemain.Reset()
		store.AgentCpuResourceTotal.Reset()
		store.AgentMemoryResourceRemain.Reset()
		store.AgentMemoryResourceTotal.Reset()
		store.AgentIpResourceRemain.Reset()
		store.StorageOperatorFailedTotal.Reset()
		store.StorageOperatorLatencyMs.Reset()
		store.StorageOperatorTotal.Reset()
		store.ClusterMemoryResouceRemain.Reset()
		store.ClusterCpuResouceRemain.Reset()
		store.ClusterMemoryResouceTotal.Reset()
		store.ClusterCpuResouceTotal.Reset()

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
				blog.Errorf("list application(%s.%s) taskgroup error %s", app.RunAs, app.Name, err.Error())
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
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceDeployment,
				deployment.ObjectMeta.NameSpace, deployment.ObjectMeta.Name, "")
		}

		// handle configmap metrics
		configmaps, err := s.ListAllConfigmaps()
		if err != nil {
			blog.Errorf("list all configmap error %s", err.Error())
		}
		for _, configmap := range configmaps {
			store.ReportObjectResourceInfoMetrics(store.ObjectResourceConfigmap,
				configmap.NameSpace, configmap.Name, "")
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
		var (
			clusterCpu float64
			clusterMem float64
			remainCpu  float64
			remainMem  float64
		)
		for _, agent := range agents {
			info := agent.GetAgentInfo()
			if info.IP == "" {
				blog.Errorf("agent %s don't have InnerIP attribute", agent.Key)
				continue
			}

			schedInfo, err := s.FetchAgentSchedInfo(info.HostName)
			if err != nil && !goerrors.Is(err, store.ErrNoFound) {
				blog.Infof("failed to to fetch agent sched info of host %s, err %s", info.HostName, err.Error())
				continue
			}

			var ipValue float64
			if s.pm != nil {
				// request netservice to node container ip
				para := &typesplugin.HostPluginParameter{
					Ips:       []string{info.IP},
					ClusterId: s.clusterId,
				}

				outerAttri, err := s.pm.GetHostAttributes(para)
				if err != nil {
					blog.Errorf("Get host(%s) ip-resources failed: %s", info.IP, err.Error())
					continue
				}
				attr, ok := outerAttri[info.IP]
				if !ok {
					blog.Errorf("host(%s) don't have ip-resources attributes", info.IP)
					continue
				}
				ipAttr := attr.Attributes[0]
				blog.Infof("Host(%s) %s Scalar(%f)", info.IP, ipAttr.Name, ipAttr.Scalar.Value)
				ipValue = ipAttr.Scalar.Value
			}

			tmpAgentRemainCpu := info.CpuTotal - info.CpuUsed
			tmpAgentRemainMem := info.MemTotal - info.MemUsed
			if schedInfo != nil {
				tmpAgentRemainCpu = tmpAgentRemainCpu - schedInfo.DeltaCPU
				tmpAgentRemainMem = tmpAgentRemainMem - schedInfo.DeltaMem
			}
			// if ip-resources is zero, then ignore it, the cluster manager do not want to see it
			if s.pm == nil || ipValue > 0 {
				remainCpu += float2Float(tmpAgentRemainCpu)
				remainMem += float2Float(tmpAgentRemainMem)

			}
			store.ReportAgentInfoMetrics(info.IP, s.clusterId, info.CpuTotal, tmpAgentRemainCpu,
				info.MemTotal, tmpAgentRemainMem, ipValue)
			clusterCpu += float2Float(info.CpuTotal)
			clusterMem += float2Float(info.MemTotal)
		}
		store.ReportClusterInfoMetrics(s.clusterId, remainCpu, clusterCpu, remainMem, clusterMem)
	}
}

func float2Float(num float64) float64 {
	float_num, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", num), 64)
	return float_num
}

// etcd store, based on kube-apiserver
func NewEtcdStore(kubeconfig string, pm *pluginManager.PluginManager, clusterId string) (store.Store, error) {
	// build kube-apiserver config
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		blog.Errorf("etcdstore build kubeconfig %s error %s", kubeconfig, err.Error())
		return nil, err
	}
	restConfig.QPS = 1e6
	restConfig.Burst = 2e6
	blog.Infof("etcdstore build kubeconfig %s success", kubeconfig)

	// build kubernetes clientset for kubeconfig
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
	// build internal clientset for kubeconfig
	clientset, err := internalclientset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("etcdstore build clientset error %s", err.Error())
		return nil, err
	}

	m := &managerStore{
		BkbcsClient:     clientset.BkbcsV2(),
		k8sClient:       k8sClientset,
		extensionClient: extensionClient,
		pm:              pm,
		clusterId:       clusterId,
	}

	// fetch application
	clientset.BkbcsV2().Applications("").List(context.Background(), metav1.ListOptions{})

	// watch application
	clientset.BkbcsV2().Applications("").Watch(context.Background(), metav1.ListOptions{})

	// list application
	clientset.BkbcsV2().Applications("").Get(context.Background(), "", metav1.GetOptions{})

	m.regkey, _ = regexp.Compile("^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$")
	m.regvalue, _ = regexp.Compile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$")

	// init clientset crds
	err = m.initKubeCrd()
	if err != nil {
		return nil, err
	}

	// init default namespace
	err = m.checkNamespace(DefaultNamespace)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// check namespace exist, if not exist, then create it
func (store *managerStore) checkNamespace(ns string) error {
	if cacheMgr != nil && cacheMgr.isOK {
		exist := checkCacheNamespaceExist(ns)
		if exist {
			blog.V(3).Infof("check namespace %s exist", ns)
			return nil
		}
	}

	client := store.k8sClient.CoreV1().Namespaces()
	_, err := client.Get(context.Background(), ns, metav1.GetOptions{})
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
		_, err = client.Create(context.Background(), ns, metav1.CreateOptions{})
		if err != nil {
			return err
		}

		syncCacheNamespace(ns.Name)
		return nil
	}

	return nil
}

// list all namespaces
func (store *managerStore) ListRunAs() ([]string, error) {
	client := store.k8sClient.CoreV1().Namespaces()
	nss, err := client.List(context.Background(), metav1.ListOptions{})
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

// filter invalid labels
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

func (store *managerStore) ObjectNotLatestErr(err error) bool {
	return strings.Contains(err.Error(), ObjectVersionNotLatestError)
}
