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

package sidecar

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/capabilities"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/apis/bkbcs/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/generated/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/generated/informers/externalversions"
	dockertypes "github.com/docker/docker/api/types"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// initKubeconfig xxx
// connect to kube-apiserver, and init BcsLogConfig crd controller
func (s *SidecarController) initKubeconfig() error {
	cfg, err := clientcmd.BuildConfigFromFlags("", s.conf.Kubeconfig)
	if err != nil {
		blog.Errorf("build kubeconfig %s error %s", s.conf.Kubeconfig, err.Error())
		return err
	}
	stopCh := make(chan struct{})
	// kubernetes clientset
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		blog.Errorf("build kubeclient by kubeconfig %s error %s", s.conf.Kubeconfig, err.Error())
		return err
	}

	nodeName := s.conf.NodeName
	if nodeName == "" {
		return fmt.Errorf("node_name is not set properly")
	}
	labelOptions := informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.FieldSelector = fmt.Sprintf("spec.nodeName=%s", nodeName)
	})
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0, labelOptions)
	blog.Infof("build informer for pod lister with nodeName filter, node_name: %s", nodeName)

	s.podLister = factory.Core().V1().Pods().Lister()
	factory.Start(stopCh)
	// Wait for all caches to sync.
	factory.WaitForCacheSync(stopCh)
	blog.Infof("build kubeclient for config %s success", s.conf.Kubeconfig)

	// apiextensions clientset for creating BcsLogConfig Crd
	s.extensionClientset, err = apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		blog.Errorf("build apiextension client by kubeconfig % error %s", s.conf.Kubeconfig, err.Error())
		return err
	}
	// create BcsLogConfig Crd
	err = s.createBcsLogConfig()
	if err != nil {
		return err
	}

	// internal clientset for informer BcsLogConfig Crd
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
	// add k8s resources event handler functions
	s.bcsLogConfigInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    s.handleAddedBcsLogConfig,
			UpdateFunc: s.handleUpdatedBcsLogConfig,
			DeleteFunc: s.handleDeletedBcsLogConfig,
		},
	)
	blog.Infof("build internalClientset for config %s success", s.conf.Kubeconfig)
	return nil
}

// createBcsLogConfig xxx
// create crd of BcsLogConf
func (s *SidecarController) createBcsLogConfig() error {
	bcsLogConfigPlural := "bcslogconfigs"
	bcsLogConfigFullName := "bcslogconfigs" + "." + bcsv1.SchemeGroupVersion.Group

	capabilities, err := capabilities.GetCapabilities(s.extensionClientset.Discovery())
	if err != nil {
		return fmt.Errorf("get kubernetes capabilities failed, err %s", err.Error())
	}
	blog.Infof("kubernetes capabilities %+v", capabilities.APIVersions)

	if !capabilities.APIVersions.Has("apiextensions.k8s.io/v1beta1") {
		blog.Infof("capabilities does not has apiextensions.k8s.io/v1beta1, create v1 version")
		err = s.createCrdV1(bcsLogConfigPlural, bcsLogConfigFullName)
	} else {
		blog.Infof("capabilities has apiextensions.k8s.io/v1beta1, create v1 beta version")
		err = s.createCrdV1Beta1(bcsLogConfigPlural, bcsLogConfigFullName)
	}

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

// getPodLogConfigCrd xxx
// InjectContent inject log envs to pod
func (s *SidecarController) getPodLogConfigCrd(container *dockertypes.ContainerJSON,
	pod *corev1.Pod) *bcsv1.BcsLogConfig {
	// fetch cluster all BcsLogConfig
	bcsLogConfs, err := s.bcsLogConfigLister.List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcslogconfig error %s", err.Error())
		return nil
	}
	if len(bcsLogConfs) == 0 {
		blog.Warnf("The container clusters don't have any BcsLogConfig")
		return nil
	}

	var highLogConfig *bcsv1.BcsLogConfig
	var highScore int
	for _, conf := range bcsLogConfs {
		blog.V(4).Infof("BcsLogConfig(%s) check pod(%s) container(%s)", conf.Name, pod.Name, container.ID)
		score := s.scoreBcsLogConfig(container, pod, conf)
		if score > highScore {
			highScore = score
			highLogConfig = conf
			blog.Infof("container %s pod(%s) BcsLogConfig(%s) higher score(%d)",
				container.ID, pod.Name, highLogConfig.Name, score)
		}
	}
	if highLogConfig == nil {
		blog.Warnf("container %s pod(%s) don't match any BcsLogConfigs", container.ID, pod.Name)
	} else {
		blog.Infof("container %s pod(%s) match BcsLogConfig(%s.%s)", container.ID, pod.Name, highLogConfig.Namespace,
			highLogConfig.Name)
	}

	return highLogConfig
}

// scoreBcsLogConfig xxx
// function scoreBcsLogConfig score the BcsLogConfig, the highest score will match the container
// no matched, 0 score
// the default BcsLogConfig, 1 score
// BcsLogConfig parameter WorkloadType、WorkloadName、WorkloadNamespace matched, increased 2 score
// BcsLogConfig parameter ContainerName matched, increased 10 score
// finally, the above scores will be accumulated to be the BcsLogConfig final score
// nolint funlen
func (s *SidecarController) scoreBcsLogConfig(container *dockertypes.ContainerJSON, pod *corev1.Pod,
	bcsLogConf *bcsv1.BcsLogConfig) int {
	// do not select ConfigType == host
	if bcsLogConf.Spec.ConfigType == bcsv1.HostConfigType {
		return 0
	}
	// selector match
	podLabelSet := apilabels.Set(pod.GetLabels())
	podSelector, err := buildSelector(bcsLogConf.Spec.Selector)
	if err != nil {
		blog.Errorf("build pod selector for bcslogconfig(%s/%s) failed: %s", bcsLogConf.GetNamespace(), bcsLogConf.GetName(),
			err.Error())
		return 0
	}
	if !podSelector.Matches(podLabelSet) {
		blog.V(4).Infof("container %s pod(%s:%s) labels(%+v) not match BcsLogConfig(%s:%s) pod selector(%+v)",
			container.ID, pod.Name, pod.Namespace, pod.GetLabels(), bcsLogConf.GetNamespace(), bcsLogConf.GetName(), podSelector)
		return 0
	}
	blog.V(4).Infof("container %s pod(%s:%s) labels(%+v) match BcsLogConfig(%s:%s) pod selector(%+v)",
		container.ID, pod.Name, pod.Namespace, pod.GetLabels(), bcsLogConf.GetNamespace(), bcsLogConf.GetName(), podSelector)
	// the default BcsLogConfig, 1 score
	if bcsLogConf.Spec.ConfigType == bcsv1.DefaultConfigType {
		return 1
	}
	// the bcs-system component BcsLogConfig, 0 score
	if bcsLogConf.Spec.ConfigType == bcsv1.BcsSystemConfigType {
		return 0
	}
	// the BcsLogConfig scores
	score := 2
	if bcsLogConf.Spec.PodNamePattern != "" {
		matched := false
		r, err := regexp.Compile(bcsLogConf.Spec.PodNamePattern)
		if err == nil && r.MatchString(pod.GetName()) {
			score += 2
			matched = true
		} else if pod.GetName() == bcsLogConf.Spec.PodNamePattern {
			score += 2
			matched = true
		}
		// not matched, return 0 score
		if !matched {
			blog.V(4).Infof("container %s pod(%s:%s) not match BcsLogConfig(%s:%s) StaticPodNamePattern %s",
				container.ID, pod.Namespace, pod.Name, bcsLogConf.Namespace, bcsLogConf.Name, bcsLogConf.Spec.PodNamePattern)
			return 0
		}
	}
	// each match BcsLogConfig parameters, if matched, then increased score
	// BcsLogConfig parameter WorkloadType、WorkloadName、WorkloadNamespace matched, increased 2 score
	// else not matched, return 0 score
	if bcsLogConf.Spec.WorkloadType != "" {
		if len(pod.OwnerReferences) == 0 {
			blog.Warnf(
				"container %s pod(%s) not match BcsLogConfig(%s:%s) WorkloadType %s, "+
					"because of lacking onwer reference information",
				container.ID, pod.Name, bcsLogConf.Namespace, bcsLogConf.Name, bcsLogConf.Spec.WorkloadType)
			return 0
		}
		matched := false
		lowerWorkloadType := strings.ToLower(bcsLogConf.Spec.WorkloadType)
		if pod.OwnerReferences[0].Kind == "ReplicaSet" {
			if lowerWorkloadType == "deployment" {
				score += 2
				matched = true
			}
		} else if pod.OwnerReferences[0].Kind == bcsLogConf.Spec.WorkloadType {
			score += 2
			matched = true
		}
		// not matched, return 0 score
		if !matched {
			blog.V(4).Infof("container %s pod(%s) OwnerReferencesKind(%s) not match BcsLogConfig(%s:%s) WorkloadType %s",
				container.ID, pod.Name, pod.OwnerReferences[0].Kind, bcsLogConf.Namespace, bcsLogConf.Name,
				bcsLogConf.Spec.WorkloadType)
			return 0
		}
	}
	if bcsLogConf.Spec.WorkloadNamespace != "" {
		// nolint
		if pod.Namespace == bcsLogConf.Spec.WorkloadNamespace {
			score += 2
			// not matched, return 0 score
		} else {
			blog.V(4).Infof("container %s pod(%s) namespace(%s) not match BcsLogConfig(%s:%s) WorkloadNamespace %s",
				container.ID, pod.Name, pod.Namespace, bcsLogConf.Namespace, bcsLogConf.Name, bcsLogConf.Spec.WorkloadNamespace)
			return 0
		}
	}
	if bcsLogConf.Spec.WorkloadName != "" {
		if len(pod.OwnerReferences) == 0 {
			blog.Warnf(
				"container %s pod(%s) not match BcsLogConfig(%s:%s) WorkloadName %s, "+
					"because of lacking onwer reference information",
				container.ID, pod.Name, bcsLogConf.Namespace, bcsLogConf.Name, bcsLogConf.Spec.WorkloadName)
			return 0
		}
		matched := false
		var workloadName string
		if pod.OwnerReferences[0].Kind == "ReplicaSet" {
			index := strings.LastIndex(pod.OwnerReferences[0].Name, "-")
			workloadName = pod.OwnerReferences[0].Name[:index]
		} else {
			workloadName = pod.OwnerReferences[0].Name
		}
		// match
		r, err := regexp.Compile(bcsLogConf.Spec.WorkloadName)
		if err == nil && r.MatchString(workloadName) {
			score += 2
			matched = true
		} else if workloadName == bcsLogConf.Spec.WorkloadName {
			score += 2
			matched = true
		}
		// not matched, return 0 score
		if !matched {
			blog.V(4).Infof("container %s pod(%s) OwnerReferencesName(%s) not match BcsLogConfig(%s:%s) WorkloadName %s",
				container.ID, pod.Name, pod.OwnerReferences[0].Name, bcsLogConf.Namespace, bcsLogConf.Name,
				bcsLogConf.Spec.WorkloadName)
			return 0
		}
	}
	// BcsLogConfig parameter ContainerName matched, increased 10 score
	matched := false
	for _, conf := range bcsLogConf.Spec.ContainerConfs {
		if conf.ContainerName == container.Config.Labels[ContainerLabelK8sContainerName] {
			score += 10
			matched = true
		}
	}
	// not matched, return 0 score
	if len(bcsLogConf.Spec.ContainerConfs) != 0 && !matched {
		blog.V(4).Infof("container(%s) pod(%s) containerName(%s) not match BcsLogConfig(%s:%s) ContainerConfs(%+v)",
			container.ID, pod.Name,
			container.Config.Labels[ContainerLabelK8sContainerName], bcsLogConf.Namespace, bcsLogConf.Name,
			bcsLogConf.Spec.ContainerConfs)
		return 0
	}

	return score
}

func (s *SidecarController) handleAddedBcsLogConfig(obj interface{}) {
	conf, ok := obj.(*bcsv1.BcsLogConfig)
	if !ok {
		blog.Errorf("cannot convert to *bcsv1.BcsLogConfig: %v", obj)
		return
	}
	by, _ := json.Marshal(conf)
	blog.Infof("handle kubernetes Add event BcsLogConfig(%s:%s) data(%s)", conf.Namespace, conf.Name, string(by))
	s.syncLogConfs()
}

func (s *SidecarController) handleDeletedBcsLogConfig(obj interface{}) {
	conf, ok := obj.(*bcsv1.BcsLogConfig)
	if !ok {
		blog.Errorf("cannot convert to *bcsv1.BcsLogConfig: %v", obj)
		return
	}
	if conf.Spec.ConfigType == bcsv1.HostConfigType {
		delete(s.logConfs, s.getHostLogConfKey(conf))
	}
	by, _ := json.Marshal(conf)
	blog.Infof("handle kubernetes Delete event BcsLogConfig(%s:%s) data(%s)", conf.Namespace, conf.Name, string(by))
	s.syncLogConfs()
}

func (s *SidecarController) handleUpdatedBcsLogConfig(oldObj, newObj interface{}) {
	conf, ok := newObj.(*bcsv1.BcsLogConfig)
	if !ok {
		blog.Errorf("cannot convert to *bcsv1.BcsLogConfig: %v", newObj)
		return
	}
	by, _ := json.Marshal(conf)
	blog.Infof("handle kubernetes Update event BcsLogConfig(%s:%s) data(%s)", conf.Namespace, conf.Name, string(by))
	s.syncLogConfs()
}

func buildSelector(selector bcsv1.PodSelector) (apilabels.Selector, error) {
	podSelector := apilabels.NewSelector()
	for key, val := range selector.MatchLabels {
		require, _ := apilabels.NewRequirement(key, selection.Equals, []string{val})
		podSelector = podSelector.Add(*require)
	}
	for _, exp := range selector.MatchExpressions {
		var op selection.Operator
		switch strings.ToLower(exp.Operator) {
		case "in", "notin", "exists":
			op = selection.Operator(strings.ToLower(exp.Operator))
		case "doesnotexist":
			op = selection.DoesNotExist
		default:
			return nil, fmt.Errorf("build pod selector error: operator(%s) is not valid", exp.Operator)
		}
		require, err := apilabels.NewRequirement(exp.Key, op, exp.Values)
		if err != nil {
			return nil, err
		}
		podSelector = podSelector.Add(*require)
	}
	return podSelector, nil
}

func (s *SidecarController) createCrdV1(bcsLogConfigPlural, bcsLogConfigFullName string) error {
	T := true
	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsLogConfigFullName,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: bcsv1.SchemeGroupVersion.Group, // BcsLogConfigsGroup,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{
				Name:    bcsv1.SchemeGroupVersion.Version,
				Served:  true,
				Storage: true,
				Schema: &apiextensionsv1.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
						Properties: map[string]apiextensionsv1.JSONSchemaProps{
							"apiVersion": {Type: "string"},
							"kind":       {Type: "string"},
							"metadata":   {Type: "object"},
							"spec":       {Type: "object", XPreserveUnknownFields: &T},
							"status":     {Type: "object", XPreserveUnknownFields: &T},
						},
						Required: []string{"metadata", "spec"},
						Type:     "object",
					},
				},
			}}, // BcsLogConfigsVersion,
			Scope: apiextensionsv1.NamespaceScoped,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   bcsLogConfigPlural,
				Kind:     reflect.TypeOf(bcsv1.BcsLogConfig{}).Name(),
				ListKind: reflect.TypeOf(bcsv1.BcsLogConfigList{}).Name(),
			},
		},
	}
	_, err := s.extensionClientset.ApiextensionsV1().CustomResourceDefinitions().Create(context.Background(), crd,
		metav1.CreateOptions{TypeMeta: crd.TypeMeta})
	return err
}

func (s *SidecarController) createCrdV1Beta1(bcsLogConfigPlural, bcsLogConfigFullName string) error {
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
	_, err := s.extensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.Background(),
		crd, metav1.CreateOptions{TypeMeta: crd.TypeMeta})
	return err
}
