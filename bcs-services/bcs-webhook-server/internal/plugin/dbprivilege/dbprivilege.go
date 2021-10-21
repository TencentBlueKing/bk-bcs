/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dbprivilege

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	clientGoCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/apis/bk-bcs/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/client/clientset/versioned"
	informers "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/client/informers/externalversions"
	listers "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/client/listers/bk-bcs/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/pluginutil"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/types"
)

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
}

// AnnotationKey implements plugin interface
func (h *Hooker) AnnotationKey() string {
	return ""
}

// Init implements plugin interface
func (h *Hooker) Init(configFilePath string) error {
	fileBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		blog.Errorf("read db privilege config file %s failed, err %s", configFilePath, err.Error())
		return fmt.Errorf("read db privilege config file %s failed, err %s", configFilePath, err.Error())
	}
	h.opt = &DbPrivOptions{}
	if err = json.Unmarshal(fileBytes, h.opt); err != nil {
		blog.Errorf("decode db privilege config failed, err %s", err.Error())
		return fmt.Errorf("decode db privilege config failed, err %s", err.Error())
	}
	if err = h.opt.Validate(); err != nil {
		return err
	}
	h.stopCh = make(chan struct{})

	if err = h.initKubeClient(); err != nil {
		return err
	}
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

	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
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
		Get(DbPrivilegeSecretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get db privilege secret in cluster failed, err %s", err.Error())
	}
	h.dbPrivSecret = dbPrivSecret

	// start factory and wait factory synced
	go factory.Start(h.stopCh)
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
		initContainers := append(pod.Spec.InitContainers, container)
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
	var fieldPath string

	for _, config := range configs {
		var env = DBPrivEnv{
			AppName:  config.Spec.AppName,
			TargetDb: config.Spec.TargetDb,
			CallUser: config.Spec.CallUser,
			DbName:   config.Spec.DbName,
			Operator: config.Spec.Operator,
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

	if h.opt.NetworkType == NetworkTypeOverlay {
		fieldPath = "status.hostIP"
	} else if h.opt.NetworkType == NetworkTypeUnderlay {
		fieldPath = "status.podIP"
	}

	initContainer := corev1.Container{
		Name:  "db-privilege",
		Image: h.opt.InitContainerImage,
		Env: []corev1.EnvVar{
			{
				Name: "io_tencent_bcs_privilege_ip",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: fieldPath,
					},
				},
			},
			{
				Name:  "io_tencent_bcs_esb_url",
				Value: h.opt.EsbURL,
			},
			{
				Name:  "io_tencent_bcs_app_code",
				Value: string(h.dbPrivSecret.Data["sdk-appCode"][:]),
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
				Value: string(envstr),
			},
		},
	}
	return initContainer, nil
}

// Close implements plugin interface
func (h *Hooker) Close() error {
	h.stopCh <- struct{}{}
	return nil
}
