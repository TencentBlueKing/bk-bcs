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

package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/signals"
	"bk-bcs/bcs-services/bcs-log-webhook-server/config"
	"bk-bcs/bcs-services/bcs-log-webhook-server/options"
	bcsv2 "bk-bcs/bcs-services/bcs-log-webhook-server/pkg/apis/bk-bcs/v2"
	internalclientset "bk-bcs/bcs-services/bcs-log-webhook-server/pkg/client/clientset/versioned"
	informers "bk-bcs/bcs-services/bcs-log-webhook-server/pkg/client/informers/externalversions"
	"bk-bcs/bcs-services/bcs-log-webhook-server/pkg/inject"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientGoCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

//Run bcs log webhook server
func Run(op *options.ServerOption) {

	conf := parseConfig(op)

	whSvr, err := NewWebhookServer(conf)
	if err != nil {
		blog.Errorf("create webhook server error %s, and exit", err.Error())
		os.Exit(1)
	}

	// start webhook server in new routine
	go func() {
		if err := whSvr.Server.ListenAndServeTLS("", ""); err != nil {
			blog.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	blog.Infof("webhook server started")

	//pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Errorf("fail to save pid: err:%s", err.Error())
	}

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	blog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	whSvr.Server.Shutdown(context.Background())

	return
}

func NewWebhookServer(conf *config.BcsLogWhsConfig) (*inject.WebhookServer, error) {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	pair, err := tls.LoadX509KeyPair(conf.ServerCertFile, conf.ServerKeyFile)
	if err != nil {
		return nil, err
	}

	whsvr := &inject.WebhookServer{
		EngineType: conf.EngineType,
		Server: &http.Server{
			Addr:      fmt.Sprintf("%s:%v", conf.Address, conf.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}

	cfg, err := clientcmd.BuildConfigFromFlags(conf.KubeMaster, conf.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kube config: %s\n", err.Error())
	}
	externalClient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error building external clientset: %s", err.Error())
	}

	created, err := createBcsLogConfig(externalClient)
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %s", err.Error())
	}
	blog.Infof("created crd: %t", created)

	clientset, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("build clientset error %s", err.Error())
	}
	whsvr.ClientSet = clientset

	factory := informers.NewSharedInformerFactory(clientset, 0)
	bcsLogConfigInformer := factory.Bkbcs().V2().BcsLogConfigs()
	whsvr.BcsLogConfigLister = bcsLogConfigInformer.Lister()

	go factory.Start(stopCh)

	blog.Infof("Waiting for inormer caches to sync")
	blog.Infof("sleep 5 seconds to wait for crd to be ready")
	time.Sleep(5 * time.Second)
	if ok := clientGoCache.WaitForCacheSync(stopCh, bcsLogConfigInformer.Informer().HasSynced); !ok {
		return nil, fmt.Errorf("failed to wait for caches to sync")
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/bcs/log_inject/v1/k8s", whsvr.K8sLogInject)
	mux.HandleFunc("/bcs/log_inject/v1/mesos", whsvr.MesosLogInject)
	whsvr.Server.Handler = mux

	return whsvr, nil
}

func createBcsLogConfig(clientset apiextensionsclient.Interface) (created bool, err error) {
	bcsLogConfigPlural := "bcslogconfigs"

	bcsLogConfigFullName := "bcslogconfigs" + "." + bcsv2.SchemeGroupVersion.Group

	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: bcsLogConfigFullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   bcsv2.SchemeGroupVersion.Group,   // BcsLogConfigsGroup,
			Version: bcsv2.SchemeGroupVersion.Version, // BcsLogConfigsVersion,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   bcsLogConfigPlural,
				Kind:     reflect.TypeOf(bcsv2.BcsLogConfig{}).Name(),
				ListKind: reflect.TypeOf(bcsv2.BcsLogConfigList{}).Name(),
			},
		},
	}

	_, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
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

func parseConfig(op *options.ServerOption) *config.BcsLogWhsConfig {
	bcsLogWhsConfig := config.NewBcsLogWhsConfig()

	bcsLogWhsConfig.Address = op.Address
	bcsLogWhsConfig.Port = op.Port
	bcsLogWhsConfig.ServerCertFile = op.ServerCertFile
	bcsLogWhsConfig.ServerKeyFile = op.ServerKeyFile
	bcsLogWhsConfig.MetricPort = op.MetricPort
	bcsLogWhsConfig.EngineType = op.EngineType
	bcsLogWhsConfig.Kubeconfig = op.KubeConfig
	bcsLogWhsConfig.KubeMaster = op.KubeMaster

	return bcsLogWhsConfig
}
