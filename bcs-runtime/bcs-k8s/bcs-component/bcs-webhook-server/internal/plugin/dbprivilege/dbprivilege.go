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

// Package dbprivilege xxx
package dbprivilege

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/capabilities"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/apis/bkbcs/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/generated/clientset/versioned"
	informers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/generated/informers/externalversions"
	listers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/generated/listers/bkbcs/v1"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	clientGoCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/plugin/dbprivilege/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/plugin/dbprivilege/pkg"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/plugin/dbprivilege/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/options"
)

const failRetryLimit = 10

// Hooker webhook for db privilege
type Hooker struct {
	stopCh                chan struct{}
	bcsDbPrivConfigLister listers.BcsDbPrivConfigLister
	opt                   *DbPrivOptions
	dbPrivSecret          *corev1.Secret
}

// DBPrivEnv is db privilege info
type DBPrivEnv struct {
	AppName  string `json:"appName"`
	TargetDb string `json:"targetDb"`
	CallUser string `json:"callUser"`
	DbName   string `json:"dbName"`
	CallType string `json:"callType"`
	Operator string `json:"operator"`
	UseCDP   bool   `json:"useCDP"`
}

// AnnotationKey implements plugin interface
func (h *Hooker) AnnotationKey() string {
	return ""
}

// Init implements plugin interface
func (h *Hooker) Init(configFilePath string) error {
	fileBytes, err := ioutil.ReadFile(configFilePath) // nolint
	if err != nil {
		blog.Errorf("read db privilege config file %s failed, err %s", configFilePath, err.Error())
		return fmt.Errorf("read db privilege config file %s failed, err %s", configFilePath, err.Error())
	}
	h.opt = &DbPrivOptions{}
	blog.Errorf("decode db privilege config, fileBytes %s", string(fileBytes))
	if err = json.Unmarshal(fileBytes, h.opt); err != nil {
		blog.Errorf("decode db privilege config failed, fileBytes %s, err %s", string(fileBytes), err.Error())
		return fmt.Errorf("decode db privilege config failed, err %s", err.Error())
	}
	if err = h.opt.Validate(); err != nil {
		return err
	}
	h.stopCh = make(chan struct{})

	if err = h.initKubeClient(); err != nil {
		return err
	}

	go h.initListenerServer()

	return nil
}

// InjectApplicationContent implements mesos plugin interface
func (h *Hooker) InjectApplicationContent(application *commtypes.ReplicaController) (
	*commtypes.ReplicaController, error) {
	return nil, nil
}

// InjectDeployContent implements mesos plugin interface
func (h *Hooker) InjectDeployContent(deploy *commtypes.BcsDeployment) (*commtypes.BcsDeployment, error) {
	return nil, nil
}

// create crd of BcsDbPrivConfig
func (h *Hooker) createBcsDbPrivCrd(clientset apiextensionsclient.Interface) (bool, error) {
	bcsDbPrivConfigPlural := "bcsdbprivconfigs"

	bcsDbPrivConfigFullName := "bcsdbprivconfigs" + "." + bcsv1.SchemeGroupVersion.Group

	var err error
	capabilities, err := capabilities.GetCapabilities(clientset.Discovery())
	if err != nil {
		return false, fmt.Errorf("get kubernetes capabilities failed, err %s", err.Error())
	}
	if !capabilities.APIVersions.Has("apiextensions.k8s.io/v1beta1") {
		blog.Infof("kubernetes doesn't support apiextensions.k8s.io/v1beta1, use v1 instead")
		crd := &apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: bcsDbPrivConfigFullName,
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Group: bcsv1.SchemeGroupVersion.Group, // BcsDbPrivConfigsGroup,
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    bcsv1.SchemeGroupVersion.Version, // BcsDbPrivConfigsVersion,
						Served:  true,
						Storage: true,
						Schema: &apiextensionsv1.CustomResourceValidation{
							OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
								Type: "object",
							},
						},
					},
				},
				Scope: apiextensionsv1.NamespaceScoped,
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural:   bcsDbPrivConfigPlural,
					Kind:     reflect.TypeOf(bcsv1.BcsDbPrivConfig{}).Name(),
					ListKind: reflect.TypeOf(bcsv1.BcsDbPrivConfigList{}).Name(),
				},
			},
		}

		_, err = clientset.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), crd,
			metav1.CreateOptions{})
	} else {
		blog.Infof("kubernetes supports apiextensions.k8s.io/v1beta1")
		crd := &apiextensionsv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: bcsDbPrivConfigFullName,
			},
			Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   bcsv1.SchemeGroupVersion.Group,   // BcsDbPrivConfigsGroup,
				Version: bcsv1.SchemeGroupVersion.Version, // BcsDbPrivConfigsVersion,
				Scope:   apiextensionsv1beta1.NamespaceScoped,
				Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
					Plural:   bcsDbPrivConfigPlural,
					Kind:     reflect.TypeOf(bcsv1.BcsDbPrivConfig{}).Name(),
					ListKind: reflect.TypeOf(bcsv1.BcsDbPrivConfigList{}).Name(),
				},
			},
		}

		_, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.Background(), crd,
			metav1.CreateOptions{})
	}

	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			blog.Infof("crd is already exists: %s", err)
			return false, nil
		}
		blog.Errorf("create crd failed: %s", err)
		return false, err
	}
	return true, nil
}

func (h *Hooker) initKubeClient() error {
	var cfg *restclient.Config
	var err error
	if len(h.opt.KubeMaster) == 0 && len(h.opt.Kubeconfig) == 0 {
		cfg, err = restclient.InClusterConfig()
		if err != nil {
			return fmt.Errorf("build config from in cluster failed, err %s", err.Error())
		}
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(h.opt.KubeMaster, h.opt.Kubeconfig)
		if err != nil {
			return fmt.Errorf("building kubeconfig failed, err %s", err.Error())
		}
	}
	// init extension kubeclient to create db privilege crd
	extensionClient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("buildling extension clientset failed, err %s", err.Error())
	}
	dbPrivCreated, err := h.createBcsDbPrivCrd(extensionClient)
	if err != nil {
		return fmt.Errorf("create db privilege crd failed, err %s", err.Error())
	}
	blog.Infof("created BcsDbPrivConfig crd: %t", dbPrivCreated)

	// init kube client to get db privilege config lister and privilege secret
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("building kubernetes clientset failed, err %s", err.Error())
	}

	clientset, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("build internal clientset failed, err %s", err.Error())
	}

	factory := informers.NewSharedInformerFactory(clientset, 0)
	bcsDbPrivConfigInformer := factory.Bkbcs().V1().BcsDbPrivConfigs()
	h.bcsDbPrivConfigLister = bcsDbPrivConfigInformer.Lister()

	dbPrivSecret, err := kubeClient.CoreV1().Secrets(metav1.NamespaceSystem).
		Get(context.Background(), DbPrivilegeSecretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get db privilege secret in cluster failed, err %s", err.Error())
	}
	h.dbPrivSecret = dbPrivSecret

	// start factory and wait factory synced
	go factory.Start(h.stopCh)

	// start goroutine
	if h.opt.DbmOptimizeEnabled {
		go h.taskAuthDBM(clientset, kubeClient)
	}
	blog.Infof("Waiting for BcsLogConfig inormer caches to sync")
	blog.Infof("sleep 1 seconds to wait for BcsLogConfig crd to be ready")
	time.Sleep(1 * time.Second)
	if ok := clientGoCache.WaitForCacheSync(h.stopCh, bcsDbPrivConfigInformer.Informer().HasSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	return nil
}

// Handle implements plugin interface
func (h *Hooker) Handle(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	// when the kind is not Pod, ignore hook
	if req.Kind.Kind != "Pod" {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	if req.Operation != v1beta1.Create {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	started := time.Now()
	pod := &corev1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		blog.Errorf("cannot decode raw object %s to pod, err %s", string(req.Object.Raw), err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(DBPrivilegePluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}

	// Deal with potential empty fields, e.g., when the pod is created by a deployment
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = req.Namespace
	}

	// do inject
	patches, err := h.createPatch(pod)
	if err != nil {
		blog.Errorf("create path failed, err %s", err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(DBPrivilegePluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Errorf("encoding patches failed, err %s", err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(DBPrivilegePluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	reviewResponse := v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
	metrics.ReportBcsWebhookServerPluginLantency(DBPrivilegePluginName, metrics.StatusSuccess, started)

	return &reviewResponse
}

func (h *Hooker) createPatch(pod *corev1.Pod) ([]types.PatchOperation, error) {
	var patch []types.PatchOperation

	bcsDbPrivConfs, err := h.bcsDbPrivConfigLister.BcsDbPrivConfigs(pod.Namespace).List(labels.Everything())
	if err != nil {
		blog.Errorf("list BcsDbPrivConfig error %s", err.Error())
		return nil, err
	}

	var matchedBdpcs []*bcsv1.BcsDbPrivConfig
	for _, d := range bcsDbPrivConfs {

		labelSelector := &metav1.LabelSelector{
			MatchLabels: d.Spec.PodSelector,
		}
		selector, err := metav1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid label selector: %s", err.Error())
		}
		if selector.Matches(labels.Set(pod.Labels)) {
			matchedBdpcs = append(matchedBdpcs, d)
		}
	}
	if len(matchedBdpcs) > 0 {
		container, err := h.generateInitContainer(matchedBdpcs)
		if err != nil {
			blog.Errorf("generateInitContainer error %s", err.Error())
			return nil, err
		}
		initContainers := append(pod.Spec.InitContainers, container) // nolint
		patch = append(patch, types.PatchOperation{
			Op:    "replace",
			Path:  "/spec/initContainers",
			Value: initContainers,
		})
	}

	return patch, nil
}

// generateInitContainer generate an init-container with BcsDbPrivConfig
func (h *Hooker) generateInitContainer(configs []*bcsv1.BcsDbPrivConfig) (corev1.Container, error) {
	var envs = make([]DBPrivEnv, 0)

	for _, config := range configs {
		var env = DBPrivEnv{
			AppName:  config.Spec.AppName,
			TargetDb: config.Spec.TargetDb,
			CallUser: config.Spec.CallUser,
			DbName:   config.Spec.DbName,
			Operator: config.Spec.Operator,
			UseCDP:   config.Spec.UseCDP,
		}
		if config.Spec.DbType == "mysql" {
			env.CallType = "mysql_ignoreCC"
		} else if config.Spec.DbType == "spider" {
			env.CallType = "spider_ignoreCC"
		}
		envs = append(envs, env)
	}
	envstr, err := json.Marshal(envs)
	if err != nil {
		blog.Errorf("convert DBPrivEnv array to json string failed: %s", err.Error())
		return corev1.Container{}, err
	}
	var fieldPath string
	if h.opt.NetworkType == NetworkTypeOverlay {
		fieldPath = "status.hostIP"
	} else if h.opt.NetworkType == NetworkTypeUnderlay {
		fieldPath = "status.podIP"
	}
	return h.getContainer(fieldPath, envstr)
}

// getContainer
func (h *Hooker) getContainer(fieldPath string, envstr []byte) (corev1.Container, error) {

	initContainer := corev1.Container{
		Name:  "db-privilege",
		Image: h.opt.InitContainerImage,
		Env:   h.buildEnvParam(fieldPath, string(envstr)),
	}

	if h.opt.InitContainerResources != nil {
		resources, err := buildContainerResources(h.opt.InitContainerResources)
		if err != nil {
			blog.Errorf("build container resources failed %s", err.Error())
			return corev1.Container{}, err
		}
		initContainer.Resources = resources
	}

	if h.opt.DbmOptimizeEnabled {
		env := buildContainerEnvVar(h.opt.ServiceName, h.opt.ServiceNamespace, h.opt.ServiceServerPort, h.opt.TicketTimer)
		initContainer.Env = append(initContainer.Env, env...)
	}

	return initContainer, nil
}

func buildContainerEnvVar(serviceName, serviceNamespace string, serviceServerPort, ticketTimer int) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: BcsPodName,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: BcsPodNamespace,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  BcsPrivilegeServiceURL,
			Value: fmt.Sprintf(BcsPrivilegeHost, serviceName, serviceNamespace, serviceServerPort),
		},
		{
			Name:  BcsPrivilegeDbmOptimizeEnabled,
			Value: "true",
		},
		{
			Name:  BcsPrivilegeServiceTicketTimer,
			Value: strconv.Itoa(ticketTimer),
		},
	}
}

func buildContainerResources(containerResources *InitContainerResources) (corev1.ResourceRequirements, error) {

	if containerResources.CpuRequest == "" || containerResources.CpuRequest == "0" ||
		containerResources.MemRequest == "" || containerResources.MemRequest == "0" ||
		containerResources.MemLimit == "" || containerResources.MemLimit == "0" ||
		containerResources.CpuLimit == "" || containerResources.CpuLimit == "0" {
		return corev1.ResourceRequirements{}, fmt.Errorf("initContainerResources failed, Requests or Limits are empty")
	}

	cpuLimit, err := resource.ParseQuantity(containerResources.CpuLimit)
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("invalid CPU limit: %v", err)
	}
	memLimit, err := resource.ParseQuantity(containerResources.MemLimit)
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("invalid Mem limit: %v", err)
	}
	cpuRequest, err := resource.ParseQuantity(containerResources.CpuRequest)
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("invalid CPU Requests: %v", err)
	}
	memRequest, err := resource.ParseQuantity(containerResources.MemRequest)
	if err != nil {
		return corev1.ResourceRequirements{}, fmt.Errorf("invalid Mem Requests: %v", err)
	}

	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    cpuLimit,
			corev1.ResourceMemory: memLimit,
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    cpuRequest,
			corev1.ResourceMemory: memRequest,
		}}, nil
}

func (h *Hooker) buildEnvParam(fieldPath, envstr string) []corev1.EnvVar {

	return []corev1.EnvVar{
		{
			Name: "io_tencent_bcs_privilege_ip",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: fieldPath,
				},
			},
		},
		{
			Name: "io_tencent_bcs_pod_ip",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name:  "external_sys_type",
			Value: h.opt.ExternalSysType,
		},
		{
			Name:  "external_sys_config",
			Value: h.opt.ExternalSysConfig,
		},
		{
			Name:  "io_tencent_bcs_app_code",
			Value: string(h.dbPrivSecret.Data["sdk-appCode"]),
		},
		{
			Name:  "io_tencent_bcs_app_secret",
			Value: string(h.dbPrivSecret.Data["sdk-appSecret"]),
		},
		{
			Name:  "io_tencent_bcs_app_operator",
			Value: string(h.dbPrivSecret.Data["sdk-operator"]),
		},
		{
			Name:  "io_tencent_bcs_db_privilege_env",
			Value: envstr,
		},
	}
}

// Close implements plugin interface
func (h *Hooker) Close() error {
	h.stopCh <- struct{}{}
	return nil
}

// taskAuthDBM 周期性汇聚未授权pod 进行dbm授权，修改状态 并写入到crd 中
func (h *Hooker) taskAuthDBM(clientset *internalclientset.Clientset, kubeClient *kubernetes.Clientset) {
	if h.opt.TicketTimer == 0 {
		h.opt.TicketTimer = 60
	}
	ticker := time.NewTicker(time.Duration(h.opt.TicketTimer) * time.Second)
	defer ticker.Stop()
	apc, aps, opt, err := util.DesDecrypt(
		h.dbPrivSecret.Data["sdk-appCode"][:], // nolint
		h.dbPrivSecret.Data["sdk-appSecret"], h.dbPrivSecret.Data["sdk-operator"])
	if err != nil {
		blog.Errorf("taskAuthDBM DesDecrypt error %s", err.Error())
		return
	}

	for { //nolint
		select {
		case <-ticker.C:
			namespaces, err := kubeClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				blog.Errorf("taskAuthDBM get Namespaces error %s", err.Error())
				continue
			}
			for _, ns := range namespaces.Items {
				bcsDbPrivConfs, err := clientset.BkbcsV1().BcsDbPrivConfigs(ns.Name).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					blog.Errorf("taskAuthDBM list BcsDbPrivConfigs error %s", err.Error())
					continue
				}
				// 根据labelSelector 匹配 pods
				for _, dbPrivConfig := range bcsDbPrivConfs.Items {
					pods, err := searchPodsByLabels(ns.Name, &dbPrivConfig, kubeClient)
					if err != nil {
						continue
					}
					if len(pods.Items) == 0 {
						blog.Errorf("taskAuthDBM Pods empty ns %s ;labels %+v ", ns.Name, dbPrivConfig.Spec.PodSelector)
						updateDbprivconfig(&dbPrivConfig, clientset)
						continue
					}
					pdsMap := make(map[string]*corev1.Pod, len(pods.Items))
					for _, pod := range pods.Items {
						pdsMap[pod.Name] = &pod
					}
					status := &dbPrivConfig.Status
					if reflect.ValueOf(status).IsZero() {
						status = &bcsv1.BcsDbPrivConfigStatus{
							DbPrivConfigStatusMap: make(map[string]*bcsv1.DbPrivConfigStatus),
						}
					}
					configStatusMap := status.DbPrivConfigStatusMap
					if reflect.ValueOf(configStatusMap).IsZero() {
						configStatusMap = make(map[string]*bcsv1.DbPrivConfigStatus)
					}
					hasDeleted := false
					hasDeleted = compareDbPrivConfigCRInfo(hasDeleted, pdsMap, &dbPrivConfig, configStatusMap)
					// 判断 pods是否在 dbPrivConfigStatusMap里，若不在则新增
					for podName, pod := range pdsMap {
						nodeIp := ""
						if h.opt.NetworkType == NetworkTypeOverlay {
							nodeIp = pod.Status.HostIP
						} else if h.opt.NetworkType == NetworkTypeUnderlay {
							nodeIp = pod.Status.PodIP
						}
						podIP := pod.Status.PodIP
						if podIP != "" && podIP != nodeIp {
							nodeIp = fmt.Sprintf("%s,%s", nodeIp, podIP)
						}
						toDbmAuth(podName, nodeIp, podIP, &dbPrivConfig, configStatusMap)
					}
					h.authDbm(hasDeleted, apc, aps, opt, &dbPrivConfig, clientset)
				}
			}
		}
	}
}

// updateDbprivconfig
func updateDbprivconfig(dbPrivConfig *bcsv1.BcsDbPrivConfig, clientset *internalclientset.Clientset) {
	if reflect.ValueOf(dbPrivConfig.Status).IsZero() {
		return
	}

	dbPrivConfig.Status = bcsv1.BcsDbPrivConfigStatus{
		DbPrivConfigStatusMap: nil,
	}
	_, err := clientset.BkbcsV1().BcsDbPrivConfigs(dbPrivConfig.Namespace).
		Update(context.TODO(), dbPrivConfig, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("taskAuthDBM authDbm update dbprivconfig failed config0: %+v, err: %s",
			dbPrivConfig, err.Error())
		return
	}
}

// authDbm 进行dbm授权
func (h *Hooker) authDbm(hasDeleted bool, apc, aps, opt string, dbPrivConfig *bcsv1.BcsDbPrivConfig,
	clientset *internalclientset.Clientset) {
	configStatusMap := dbPrivConfig.Status.DbPrivConfigStatusMap
	if len(configStatusMap) == 0 {
		return
	}
	var wg sync.WaitGroup
	hasUpdated := false
	for _, dbConfig := range configStatusMap {
		// 去授权
		if dbConfig.Status == BcsPrivilegeDBMAuthStatusPending || dbConfig.Status == BcsPrivilegeDBMAuthStatusChanged {
			dbConfig.Status = BcsPrivilegeDBMAuthStatusPending
			configStatusMap[dbConfig.PodName] = dbConfig
			hasUpdated = true
			callType := ""
			if dbPrivConfig.Spec.DbType == "mysql" {
				callType = "mysql_ignoreCC"
			} else if dbPrivConfig.Spec.DbType == "spider" {
				callType = "spider_ignoreCC"
			}
			osEnv := &options.Env{
				PodIp:             dbConfig.PodIp,
				NodeIp:            dbConfig.NodeIp,
				ExternalSysType:   h.opt.ExternalSysType,
				ExternalSysConfig: h.opt.ExternalSysConfig,
				AppCode:           apc,
				AppSecret:         aps,
				AppOperator:       opt,
				CallType:          callType,
			}

			wg.Add(1)
			go func(env *options.Env, dbPrivConfigStatus *bcsv1.DbPrivConfigStatus) {
				defer wg.Done()
				var doPriRetry, checkRetry = 0, 0
				client, err := pkg.InitClient(env)
				if err != nil {
					blog.Errorf("taskAuthDBM failed to init client for external system, %s", err.Error())
					return
				}
				for doPriRetry < failRetryLimit {
					time.Sleep(1 * time.Second)
					err = client.DoPri(env, dbPrivConfigStatus)
					if err == nil {
						break
					}
					blog.Errorf("taskAuthDBM error calling the privilege api err: %s, status: %+v, retry %d",
						err.Error(), dbPrivConfigStatus, doPriRetry)
					doPriRetry++
				}
				if doPriRetry >= failRetryLimit {
					blog.Errorf("taskAuthDBM error calling the privilege api with db: %s, dbname: %s, max retry times reached",
						dbPrivConfigStatus.TargetDb, dbPrivConfigStatus.DbName)
					return
				}
				for checkRetry < failRetryLimit {
					common.WaitForSeveralSeconds()
					err = client.CheckFinalStatus()
					if err == nil {
						break
					}
					blog.Errorf("taskAuthDBM check operation status failed: %s, db: %s, dbname: %s, retry %d",
						err.Error(), dbPrivConfigStatus.TargetDb, dbPrivConfigStatus.DbName, checkRetry)
					checkRetry++
				}
				if checkRetry >= failRetryLimit {
					blog.Errorf("taskAuthDBM check operation status failed with db: %s, dbname: %s, max retry times reached",
						dbPrivConfigStatus.TargetDb, dbPrivConfigStatus.DbName)
					return
				}
				dbPrivConfigStatus.Status = BcsPrivilegeDBMAuthStatusDone
				configStatusMap[dbPrivConfigStatus.PodName] = dbPrivConfigStatus

			}(osEnv, dbConfig)
		}
	}
	wg.Wait()
	updatePrivConfigCR(hasDeleted, hasUpdated, dbPrivConfig, clientset)
}

func updatePrivConfigCR(hasDeleted bool, hasUpdated bool, dbPrivConfig *bcsv1.BcsDbPrivConfig,
	clientset *internalclientset.Clientset) {
	if hasUpdated || hasDeleted {
		blog.Infof("authDbm start Update crName:%s", dbPrivConfig.Name)
		_, err := clientset.BkbcsV1().BcsDbPrivConfigs(dbPrivConfig.Namespace).Update(context.TODO(), dbPrivConfig,
			metav1.UpdateOptions{})
		if err != nil {
			blog.Errorf("taskAuthDBM update failed dbPrivConfig: %+v, err: %s", dbPrivConfig, err.Error())
			return
		}
	}
}

// initListenerServer 开启一个服务监听新的端口，查询pod是否已经授权，用于被 initContainer 调用接口
func (h *Hooker) initListenerServer() {

	var cfg *restclient.Config
	var err error
	if len(h.opt.KubeMaster) == 0 && len(h.opt.Kubeconfig) == 0 {
		cfg, err = restclient.InClusterConfig()
		if err != nil {
			blog.Errorf("initListenerServer build config from in cluster failed, err %s", err.Error())
			return
		}
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(h.opt.KubeMaster, h.opt.Kubeconfig)
		if err != nil {
			blog.Errorf("initListenerServer building kubeconfig failed, err %s", err.Error())
			return
		}
	}

	clientset, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		blog.Errorf("initListenerServer internalclientset NewForConfig error %s", err.Error())
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/check_status", func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		podName := query.Get("podName")
		podNamespace := query.Get("podNameSpace")
		blog.Infof("initListenerServer podName:%s,podNamespace:%s", podName, podNamespace)
		ok, _ := h.checkDbPrivConfigIsOk(podName, podNamespace, clientset)
		if ok {
			_, err := fmt.Fprint(writer, "ok") //nolint
			if err != nil {
				blog.Errorf("initListenerServer check_status podName:%s, podNamespace:%s, err:%s", podName,
					podNamespace, err.Error())
				return
			}
			return
		}

		_, err := fmt.Fprint(writer, "not ok") // nolint
		if err != nil {
			blog.Errorf("initListenerServer podName:%s, podNamespace:%s, err:%s", podName, podNamespace, err.Error())
			return
		}
	})

	// HTTP服务器监听xxxx端口
	err = http.ListenAndServe(fmt.Sprintf(":%d", h.opt.ServiceServerPort), mux)
	if err != nil {
		blog.Errorf("initListenerServer ListenAndServe error %s", err.Error())
		return
	}
}

// checkDbPrivConfigIsOk 查询pod是否已经授权
func (h *Hooker) checkDbPrivConfigIsOk(podName string, namespace string,
	clientset *internalclientset.Clientset) (bool, error) {

	// 查出符合条件的dbprivconfigs
	bcsDbPrivConfs, err := clientset.BkbcsV1().BcsDbPrivConfigs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		blog.Errorf("initListenerServer checkDbPrivConfigIsOk list BcsDbPrivConfigs error %s", err.Error())
		return false, err
	}

	for _, dbPrivConfig := range bcsDbPrivConfs.Items {
		if !reflect.ValueOf(dbPrivConfig.Status).IsZero() && len(dbPrivConfig.Status.DbPrivConfigStatusMap) > 0 {
			if dbPrivConfigStatus, ok := dbPrivConfig.Status.DbPrivConfigStatusMap[podName]; ok {
				if dbPrivConfigStatus.Status == BcsPrivilegeDBMAuthStatusDone {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// searchPodsByLabels 根据namespace 和 labels 查询pods
func searchPodsByLabels(namespace string, dbPrivConfig *bcsv1.BcsDbPrivConfig,
	kubeClient *kubernetes.Clientset) (*corev1.PodList, error) {

	if dbPrivConfig.Spec.PodSelector == nil {
		blog.Errorf("taskAuthDBM PodSelector is empty; namespace:%s ; cr name:%s", namespace, dbPrivConfig.Name)
		return nil, fmt.Errorf("PodSelector is empty")
	}

	labelSelector := &metav1.LabelSelector{
		MatchLabels: dbPrivConfig.Spec.PodSelector,
	}

	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		blog.Errorf("taskAuthDBM LabelSelectorAsSelector error %s", err.Error())
		return nil, err
	}

	// 查找相匹配的Pods
	pods, err := kubeClient.CoreV1().Pods(namespace).List(context.TODO(),
		metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		blog.Errorf("taskAuthDBM list Pods failed; ns:%s; crname:%s;error %s", namespace,
			dbPrivConfig.Name, err.Error())
		return nil, err
	}

	return pods, nil
}

// compareDbPrivConfigCRInfo 对比 dbPrivConfig 信息
func compareDbPrivConfigCRInfo(hasDeleted bool, pdsMap map[string]*corev1.Pod,
	dbPrivConfig *bcsv1.BcsDbPrivConfig, configStatusMap map[string]*bcsv1.DbPrivConfigStatus) bool {

	if len(configStatusMap) == 0 {
		return hasDeleted
	}

	var dltKeys []string
	// 查看对比 db信息是否有改变，如有改变，则更新db信息以及状态
	// 对比该pod信息是否还在 pods里
	for podName, dbPrivConfigStatus := range configStatusMap {
		// 判断pod 是否还在,若不在，则无须往下走
		if _, ok := pdsMap[podName]; !ok {
			dltKeys = append(dltKeys, podName)
			continue
		}

		flag := false
		// 对比db信息
		if dbPrivConfig.Spec.DbName != dbPrivConfigStatus.DbName {
			dbPrivConfigStatus.DbName = dbPrivConfig.Spec.DbName
			flag = true
		}
		if dbPrivConfig.Spec.TargetDb != dbPrivConfigStatus.TargetDb {
			dbPrivConfigStatus.TargetDb = dbPrivConfig.Spec.TargetDb
			flag = true
		}
		if dbPrivConfig.Spec.AppName != dbPrivConfigStatus.AppName {
			dbPrivConfigStatus.AppName = dbPrivConfig.Spec.AppName
			flag = true
		}
		if dbPrivConfig.Spec.DbType != dbPrivConfigStatus.DbType {
			dbPrivConfigStatus.DbType = dbPrivConfig.Spec.DbType
			flag = true
		}
		if dbPrivConfig.Spec.Operator != dbPrivConfigStatus.Operator {
			dbPrivConfigStatus.Operator = dbPrivConfig.Spec.Operator
			flag = true
		}
		if dbPrivConfig.Spec.UseCDP != dbPrivConfigStatus.UseCDP {
			dbPrivConfigStatus.UseCDP = dbPrivConfig.Spec.UseCDP
			flag = true
		}
		if dbPrivConfig.Spec.CallUser != dbPrivConfigStatus.CallUser {
			dbPrivConfigStatus.CallUser = dbPrivConfig.Spec.CallUser
			flag = true
		}

		if flag {
			dbPrivConfigStatus.Status = BcsPrivilegeDBMAuthStatusChanged
		}
	}

	if len(dltKeys) > 0 {
		hasDeleted = true
		for _, key := range dltKeys {
			delete(configStatusMap, key)
		}
	}

	return hasDeleted
}

// toDbmAuth 构造dbprivconfig 信息，并且去dbm授权
func toDbmAuth(podName, nodeIp, podIP string, dbPrivConfig *bcsv1.BcsDbPrivConfig,
	configStatusMap map[string]*bcsv1.DbPrivConfigStatus) {
	// 若不在则新增
	if _, ok := configStatusMap[podName]; !ok {
		dbPrivConfigStatus := &bcsv1.DbPrivConfigStatus{
			AppName:  dbPrivConfig.Spec.AppName,
			TargetDb: dbPrivConfig.Spec.TargetDb,
			DbType:   dbPrivConfig.Spec.DbType,
			CallUser: dbPrivConfig.Spec.CallUser,
			DbName:   dbPrivConfig.Spec.DbName,
			Operator: dbPrivConfig.Spec.Operator,
			UseCDP:   dbPrivConfig.Spec.UseCDP,
			Status:   "pending",
			PodIp:    podIP,
			NodeIp:   nodeIp,
			PodName:  podName,
		}

		configStatusMap[podName] = dbPrivConfigStatus
	} else {
		if podIP != "" && configStatusMap[podName].PodIp != podIP {
			configStatusMap[podName].PodIp = podIP
			configStatusMap[podName].Status = BcsPrivilegeDBMAuthStatusChanged
		}
		if nodeIp != "" && configStatusMap[podName].NodeIp != nodeIp {
			configStatusMap[podName].NodeIp = nodeIp
			configStatusMap[podName].Status = BcsPrivilegeDBMAuthStatusChanged
		}
	}

	dbPrivConfig.Status.DbPrivConfigStatusMap = configStatusMap

}
