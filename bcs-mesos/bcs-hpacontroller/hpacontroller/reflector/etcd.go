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

package reflector

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/signals"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	schedtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/config"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"
	internalclientset "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/clientset/versioned"
	informers "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/informers/externalversions"
	listers "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/listers/bkbcs/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientGoCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ApiversionV2      = "v2"
	autoscalerCrdKind = "autoscaler"
	crdCrd            = "Crd"
)

type etcdReflector struct {
	//hpa controller config
	config            *config.Config
	bkbcsClientSet    *internalclientset.Clientset //kube bkbcs clientset
	crdLister         listers.CrdLister
	deploymentLister  listers.DeploymentLister
	applicationLister listers.ApplicationLister
	taskgroupLister   listers.TaskGroupLister
}

func NewEtcdReflector(conf *config.Config) Reflector {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	reflector := &etcdReflector{
		config: conf,
	}

	if conf.KubeConfig == "" {
		blog.Errorf("kubeconfig not provided, exit")
		os.Exit(1)
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", conf.KubeConfig)
	if err != nil {
		blog.Errorf("etcd reflector build kubeconfig %s error %s", conf.KubeConfig, err.Error())
		os.Exit(1)
	}
	blog.Infof("etcd reflector build kubeconfig %s success", conf.KubeConfig)

	clientset, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		blog.Errorf("etcd reflector build clientset error %s", err.Error())
		os.Exit(1)
	}
	reflector.bkbcsClientSet = clientset

	factory := informers.NewSharedInformerFactory(clientset, 0)
	crdInformer := factory.Bkbcs().V2().Crds()
	deploymentInformer := factory.Bkbcs().V2().Deployments()
	applicationInformer := factory.Bkbcs().V2().Applications()
	taskgroupInformer := factory.Bkbcs().V2().TaskGroups()
	reflector.crdLister = crdInformer.Lister()
	reflector.deploymentLister = deploymentInformer.Lister()
	reflector.applicationLister = applicationInformer.Lister()
	reflector.taskgroupLister = taskgroupInformer.Lister()

	go factory.Start(stopCh)

	blog.Infof("Waiting for informer caches to sync")
	if ok := clientGoCache.WaitForCacheSync(stopCh, crdInformer.Informer().HasSynced, deploymentInformer.Informer().HasSynced, applicationInformer.Informer().HasSynced, taskgroupInformer.Informer().HasSynced); !ok {
		blog.Errorf("failed to wait for caches to sync")
		os.Exit(1)
	}

	return reflector
}

//list all namespace autoscaler
func (reflector *etcdReflector) ListAutoscalers() ([]*commtypes.BcsAutoscaler, error) {
	bcsCrdList, err := reflector.crdLister.List(labels.Everything())
	if err != nil {
		blog.Errorf("store kube-api list crds error %s", err.Error())
		return nil, err
	}
	scalers := make([]*commtypes.BcsAutoscaler, 0)
	for _, bcsCrd := range bcsCrdList {
		if strings.Contains(bcsCrd.Namespace, autoscalerCrdKind) {
			crd := bcsCrd.Spec
			var scaler *commtypes.BcsAutoscaler
			scaler.TypeMeta = crd.TypeMeta
			scaler.ObjectMeta = crd.ObjectMeta
			scalerSpec := crd.Spec.(commtypes.BcsAutoscalerSpec)
			scaler.Spec = &scalerSpec
			scalerStatus := crd.Status.(commtypes.BcsAutoscalerStatus)
			scaler.Status = &scalerStatus
			scalers = append(scalers, scaler)
		}
	}

	return scalers, nil
}

//crd namespace = crd.kind-crd.namespace
func getCrdNamespace(kind, ns string) string {
	return fmt.Sprintf("%s-%s", kind, ns)
}

func (reflector *etcdReflector) CheckCustomResourceDefinitionExist(crd *commtypes.Crd) (string, bool) {
	//client := reflector.bkbcsClientSet.BkbcsV2().Crds(getCrdNamespace(string(crd.Kind), crd.NameSpace))
	//v2Crd, _ := client.Get(crd.Name, metav1.GetOptions{})
	v2Crd, _ := reflector.crdLister.Crds(getCrdNamespace(string(crd.Kind), crd.NameSpace)).Get(crd.Name)
	if v2Crd != nil {
		return v2Crd.ResourceVersion, true
	}

	return "", false
}

func (reflector *etcdReflector) StoreAutoscaler(autoscaler *commtypes.BcsAutoscaler) error {
	//crd namespace = crd.kind-crd.namespace
	realNs := getCrdNamespace(autoscalerCrdKind, autoscaler.NameSpace)
	client := reflector.bkbcsClientSet.BkbcsV2().Crds(realNs)

	var crd commtypes.Crd
	crd.TypeMeta = autoscaler.TypeMeta
	crd.ObjectMeta = autoscaler.ObjectMeta
	crd.Spec = autoscaler.Spec
	crd.Status = autoscaler.Status
	v2Crd := &v2.Crd{
		TypeMeta: metav1.TypeMeta{
			Kind:       crdCrd,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      crd.Name,
			Namespace: realNs,
		},
		Spec: v2.CrdSpec{
			Crd: crd,
		},
	}

	rv, exist := reflector.CheckCustomResourceDefinitionExist(&crd)
	var err error
	if exist {
		v2Crd.ResourceVersion = rv
		_, err = client.Update(context.Background(), v2Crd, metav1.UpdateOptions{})
	} else {
		_, err = client.Create(context.Background(), v2Crd, metav1.CreateOptions{})
	}
	return err
}

func (reflector *etcdReflector) UpdateAutoscaler(autoscaler *commtypes.BcsAutoscaler) error {
	etcdScaler, err := reflector.FetchAutoscalerByUuid(autoscaler.GetUuid())
	if err != nil {
		return err
	}
	if etcdScaler.GetUuid() != autoscaler.GetUuid() {
		return fmt.Errorf("autoscaler %s not found", autoscaler.GetUuid())
	}

	reflector.StoreAutoscaler(autoscaler)
	return err
}

// fetch autoscaler from kube-api
func (reflector *etcdReflector) FetchAutoscalerByUuid(uuid string) (*commtypes.BcsAutoscaler, error) {
	uids := strings.Split(uuid, "_")
	if len(uids) != 3 {
		return nil, fmt.Errorf("uuid %s is invalid", uuid)
	}

	//client := reflector.bkbcsClientSet.BkbcsV2().Crds(getCrdNamespace(autoscalerCrdKind, uids[0]))
	//v2Crd, err := client.Get(uids[1], metav1.GetOptions{})
	v2Crd, err := reflector.crdLister.Crds(getCrdNamespace(autoscalerCrdKind, uids[0])).Get(uids[1])
	if err != nil {
		return nil, err
	}

	crd := v2Crd.Spec
	var scaler *commtypes.BcsAutoscaler
	scaler.TypeMeta = crd.TypeMeta
	scaler.ObjectMeta = crd.ObjectMeta
	scalerSpec := crd.Spec.(commtypes.BcsAutoscalerSpec)
	scaler.Spec = &scalerSpec
	scalerStatus := crd.Status.(commtypes.BcsAutoscalerStatus)
	scaler.Status = &scalerStatus

	return scaler, nil
}

//fetch deployment info, if deployment status is not Running, then can't autoscale this deployment
func (reflector *etcdReflector) FetchDeploymentInfo(namespace, name string) (*schedtypes.Deployment, error) {
	bcsDeployment, err := reflector.deploymentLister.Deployments(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("get bcsDeployment failed, %s", err.Error())
	}
	deploy := bcsDeployment.Spec.Deployment
	return &deploy, nil
}

//fetch application info, if application status is not Running or Abnormal, then can't autoscale this application
func (reflector *etcdReflector) FetchApplicationInfo(namespace, name string) (*schedtypes.Application, error) {
	bcsApplication, err := reflector.applicationLister.Applications(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("get bcsApplication failed, %s", err.Error())
	}
	application := bcsApplication.Spec.Application
	return &application, nil
}

//list selectorRef deployment taskgroup
func (reflector *etcdReflector) ListTaskgroupRefDeployment(namespace, name string) ([]*schedtypes.TaskGroup, error) {
	bcsDeployment, err := reflector.deploymentLister.Deployments(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("get bcsDeployment failed, %s", err.Error())
	}
	deploy := bcsDeployment.Spec.Deployment

	return reflector.ListTaskgroupRefApplication(namespace, deploy.Application.ApplicationName)
}

//list selectorRef application taskgroup
func (reflector *etcdReflector) ListTaskgroupRefApplication(namespace, name string) ([]*schedtypes.TaskGroup, error) {
	bcsTaskgroupList, err := reflector.taskgroupLister.TaskGroups(namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("list bcsTaskgroup failed, %s", err.Error())
	}
	taskgroups := make([]*schedtypes.TaskGroup, 0)
	for _, bcsTaskgroup := range bcsTaskgroupList {
		if bcsTaskgroup.Spec.AppID == name {
			taskGroup := bcsTaskgroup.Spec.TaskGroup
			taskgroups = append(taskgroups, &taskGroup)
		}
	}

	return taskgroups, nil
}
