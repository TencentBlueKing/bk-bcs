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

package sidecar

import (
	"reflect"
	"strings"
	"time"

	"bk-bcs/bcs-common/common/blog"
	bcsv1 "bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
	internalclientset "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/clientset/versioned"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/client/informers/externalversions"

	docker "github.com/fsouza/go-dockerclient"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

//connect to kube-apiserver, and init BcsLogConfig crd controller
func (s *SidecarController) initKubeconfig() error {
	cfg, err := clientcmd.BuildConfigFromFlags("", s.conf.Kubeconfig)
	if err != nil {
		blog.Errorf("build kubeconfig %s error %s", s.conf.Kubeconfig, err.Error())
		return err
	}
	stopCh := make(chan struct{})
	//kubernetes clientset
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		blog.Errorf("build kubeclient by kubeconfig %s error %s", s.conf.Kubeconfig, err.Error())
		return err
	}
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	s.podLister = factory.Core().V1().Pods().Lister()
	factory.Start(stopCh)
	// Wait for all caches to sync.
	factory.WaitForCacheSync(stopCh)
	blog.Infof("build kubeclient for config %s success", s.conf.Kubeconfig)

	//apiextensions clientset for creating BcsLogConfig Crd
	s.extensionClientset, err = apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		blog.Errorf("build apiextension client by kubeconfig % error %s", s.conf.Kubeconfig, err.Error())
		return err
	}
	//create BcsLogConfig Crd
	err = s.createBcsLogConfig()
	if err != nil {
		return err
	}

	//internal clientset for informer BcsLogConfig Crd
	internalClientset, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		blog.Errorf("build internal clientset by kubeconfig %s error %s", s.conf.Kubeconfig, err.Error())
		return err
	}
	internalFactory := externalversions.NewSharedInformerFactory(internalClientset, time.Hour)
	s.bcsLogConfigInformer = internalFactory.Bkbcs().V1().BcsLogConfigs().Informer()
	s.bcsLogConfigLister = internalFactory.Bkbcs().V1().BcsLogConfigs().Lister()
	internalFactory.Start(stopCh)
	// Wait for all caches to sync.
	internalFactory.WaitForCacheSync(stopCh)
	//add k8s resources event handler functions
	s.bcsLogConfigInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    s.handleChangedBcsLogConfig,
			UpdateFunc: s.handleUpdatedBcsLogConfig,
			DeleteFunc: s.handleChangedBcsLogConfig,
		},
	)
	blog.Infof("build internalClientset for config %s success", s.conf.Kubeconfig)
	return nil
}

// create crd of BcsLogConf
func (s *SidecarController) createBcsLogConfig() error {
	bcsLogConfigPlural := "bcslogconfigs"
	bcsLogConfigFullName := "bcslogconfigs" + "." + bcsv1.SchemeGroupVersion.Group
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsLogConfigFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   bcsv1.SchemeGroupVersion.Group,   // BcsLogConfigsGroup,
			Version: bcsv1.SchemeGroupVersion.Version, // BcsLogConfigsVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   bcsLogConfigPlural,
				Kind:     reflect.TypeOf(bcsv1.BcsLogConfig{}).Name(),
				ListKind: reflect.TypeOf(bcsv1.BcsLogConfigList{}).Name(),
			},
		},
	}

	_, err := s.extensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			blog.Infof("BcsLogConfig Crd is already exists")
			return nil
		}
		blog.Errorf("create BcsLogConfig Crd error %s", err.Error())
		return err
	}
	blog.Infof("create BcsLogConfig Crd success")
	return nil
}

// InjectContent inject log envs to pod
func (s *SidecarController) getPodLogConfigCrd(container *docker.Container, pod *corev1.Pod) *bcsv1.BcsLogConfig {
	//fetch cluster all BcsLogConfig
	bcsLogConfs, err := s.bcsLogConfigLister.List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcslogconfig error %s", err.Error())
		return nil
	}
	if len(bcsLogConfs)==0 {
		blog.Warnf("The container clusters don't have any BcsLogConfig")
		return nil
	}

	var highLogConfig *bcsv1.BcsLogConfig
	var highScore int
	for _, conf := range bcsLogConfs {
		blog.Infof("BcsLogConfig(%s) check pod(%s) container(%s)", conf.Name, pod.Name, container.ID)
		score := scoreBcsLogConfig(container, pod, conf)
		if score > highScore {
			highScore = score
			highLogConfig = conf
			blog.Infof("container %s pod(%s) BcsLogConfig(%s) higher score(%d)",
				container.ID, pod.Name, highLogConfig.Name, score)
		}
	}
	if highLogConfig == nil {
		blog.Warnf("container %s pod(%s) not match BcsLogConfigs", container.ID, pod.Name)
	} else {
		blog.Infof("container %s pod(%s) match BcsLogConfig(%s)", container.ID, pod.Name, highLogConfig.Name)
	}

	return highLogConfig
}

//function scoreBcsLogConfig score the BcsLogConfig, the highest score will match the container
//no matched, 0 score
//the default BcsLogConfig, 1 score
//BcsLogConfig parameter WorkloadType、WorkloadName、WorkloadNamespace matched, increased 2 score
//BcsLogConfig parameter ContainerName matched, increased 10 score
//finally, the above scores will be accumulated to be the BcsLogConfig final score
func scoreBcsLogConfig(container *docker.Container, pod *corev1.Pod, bcsLogConf *bcsv1.BcsLogConfig) int {
	//the default BcsLogConfig, 1 score
	if bcsLogConf.Spec.ConfigType == bcsv1.DefaultConfigType {
		return 1
	}
	//the bcs-system component BcsLogConfig, 0 score
	if bcsLogConf.Spec.ConfigType == bcsv1.BcsSystemConfigType {
		return 0
	}
	//the BcsLogConfig scores
	score := 0
	//each match BcsLogConfig parameters, if matched, then increased score
	//BcsLogConfig parameter WorkloadType、WorkloadName、WorkloadNamespace matched, increased 2 score
	//else not matched, return 0 score
	if bcsLogConf.Spec.WorkloadType != "" {
		//if pod don't belong any workload
		if len(pod.OwnerReferences)==0 {
			blog.Warnf("container %s pod(%s:%s) not match BcsLogConfig(%s:%s) WorkloadType",
				container.ID, pod.Name, pod.Namespace, bcsLogConf.Name, bcsLogConf.Spec.WorkloadType)
			return 0
		}

		matched := false
		if pod.OwnerReferences[0].Kind == "ReplicaSet" {
			if strings.ToLower(bcsLogConf.Spec.WorkloadType) == strings.ToLower("Deployment") &&
				strings.HasPrefix(pod.OwnerReferences[0].Name, bcsLogConf.Spec.WorkloadName){
				score += 2
				matched = true
			}
		} else if strings.ToLower(pod.OwnerReferences[0].Kind) == strings.ToLower(bcsLogConf.Spec.WorkloadType) {
			score += 2
			matched = true
		}
		//not matched, return 0 score
		if !matched {
			blog.Warnf("container %s pod(%s:%s) not match BcsLogConfig(%s:%s) WorkloadType",
				container.ID, pod.Name, pod.OwnerReferences[0].Kind, bcsLogConf.Name, bcsLogConf.Spec.WorkloadType)
			return 0
		}
	}
	if bcsLogConf.Spec.WorkloadNamespace != "" {
		if pod.Namespace == bcsLogConf.Spec.WorkloadNamespace {
			score += 2
			//not matched, return 0 score
		} else {
			blog.Warnf("container %s pod(%s:%s) not match BcsLogConfig(%s:%s) WorkloadNamespace",
				container.ID, pod.Name, pod.Namespace, bcsLogConf.Name, bcsLogConf.Spec.WorkloadNamespace)
			return 0
		}
	}
	if bcsLogConf.Spec.WorkloadName != "" {
		matched := false
		if pod.OwnerReferences[0].Kind == "ReplicaSet" {
			if strings.HasPrefix(pod.OwnerReferences[0].Name, bcsLogConf.Spec.WorkloadName) {
				score += 2
				matched = true
			}
		} else if pod.OwnerReferences[0].Name == bcsLogConf.Spec.WorkloadName {
			score += 2
			matched = true
		}
		//not matched, return 0 score
		if !matched {
			blog.Warnf("container %s pod(%s:%s) not match BcsLogConfig(%s:%s) WorkloadName",
				container.ID, pod.Name, pod.OwnerReferences[0].Name, bcsLogConf.Name, bcsLogConf.Spec.WorkloadName)
			return 0
		}
	}
	//BcsLogConfig parameter ContainerName matched, increased 10 score
	matched := false
	for _, conf := range bcsLogConf.Spec.ContainerConfs {
		if conf.ContainerName == container.Config.Labels[ContainerLabelK8sContainerName] {
			score += 10
			matched = true
		}
	}
	//not matched, return 0 score
	if len(bcsLogConf.Spec.ContainerConfs) != 0 && !matched {
		blog.Warnf("container(%s:%s) pod(%s:%s) not match BcsLogConfig(%s) containerName",
			container.ID, container.Config.Labels[ContainerLabelK8sContainerName], pod.Name, pod.Name, bcsLogConf.Name)
		return 0
	}

	return score
}

func (s *SidecarController) handleChangedBcsLogConfig(obj interface{}) {
	conf, ok := obj.(*bcsv1.BcsLogConfig)
	if !ok {
		blog.Errorf("cannot convert to *bcsv1.BcsLogConfig: %v", obj)
		return
	}
	blog.Infof("handle kubernetes AddOrDelete event BcsLogConfig(%s:%s)", conf.Name, conf.Namespace)
	s.syncLogConfs()
}

func (s *SidecarController) handleUpdatedBcsLogConfig(oldObj, newObj interface{}) {
	conf, ok := newObj.(*bcsv1.BcsLogConfig)
	if !ok {
		blog.Errorf("cannot convert to *bcsv1.BcsLogConfig: %v", newObj)
		return
	}
	blog.Infof("handle kubernetes Update event BcsLogConfig(%s:%s)", conf.Name, conf.Namespace)
	s.syncLogConfs()
}
