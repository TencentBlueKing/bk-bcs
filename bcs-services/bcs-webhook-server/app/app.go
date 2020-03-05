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
	"syscall"
	"time"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/signals"
	"bk-bcs/bcs-services/bcs-webhook-server/config"
	"bk-bcs/bcs-services/bcs-webhook-server/options"
	internalclientset "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/clientset/versioned"
	informers "bk-bcs/bcs-services/bcs-webhook-server/pkg/client/informers/externalversions"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/k8s"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/mesos"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientGoCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// DbPrivilegeSecretName the name of secret to store db privilege info
	DbPrivilegeSecretName = "bcs-db-privilege"
	// EngineTypeKubernetes kubernetes engine type
	EngineTypeKubernetes = "kubernetes"
	// EngineTypeMesos mesos engine type
	EngineTypeMesos = "mesos"
)

//Run bcs log webhook server
func Run(op *options.ServerOption) {

	conf := parseConfig(op)

	whSvr, err := NewWebhookServer(conf)
	if err != nil {
		blog.Errorf("create webhook server error: %s, and exit", err.Error())
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

// NewWebhookServer new WebHookServer object
func NewWebhookServer(conf *config.BcsWhsConfig) (*inject.WebhookServer, error) {
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
		Injects: conf.Injects,
	}

	cfg, err := clientcmd.BuildConfigFromFlags(conf.KubeMaster, conf.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kube config: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}
	whsvr.KubeClient = kubeClient

	externalClient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error building external clientset: %s", err.Error())
	}

	clientset, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("build clientset error %s", err.Error())
	}
	whsvr.ClientSet = clientset

	factory := informers.NewSharedInformerFactory(clientset, 0)

	if conf.Injects.LogConfEnv {
		logCreated, err := createBcsLogConfig(externalClient)
		if err != nil {
			return nil, fmt.Errorf("error creating crd: %s", err.Error())
		}
		blog.Infof("created BcsLogConfig crd: %t", logCreated)

		bcsLogConfigInformer := factory.Bkbcs().V1().BcsLogConfigs()
		whsvr.BcsLogConfigLister = bcsLogConfigInformer.Lister()

		switch whsvr.EngineType {
		case EngineTypeKubernetes:
			k8sLogConfInject := k8s.NewLogConfInject(whsvr.BcsLogConfigLister)
			whsvr.K8sLogConfInject = k8sLogConfInject
		case EngineTypeMesos:
			mesosLogConfInject := mesos.NewLogConfInject(whsvr.BcsLogConfigLister)
			whsvr.MesosLogConfInject = mesosLogConfInject
		}

		go factory.Start(stopCh)

		blog.Infof("Waiting for BcsLogConfig inormer caches to sync")
		blog.Infof("sleep 1 seconds to wait for BcsLogConfig crd to be ready")
		time.Sleep(1 * time.Second)
		if ok := clientGoCache.WaitForCacheSync(stopCh, bcsLogConfigInformer.Informer().HasSynced); !ok {
			return nil, fmt.Errorf("failed to wait for caches to sync")
		}
	}

	if conf.Injects.DbPriv.DbPrivInject {
		dbPrivCreated, err := createBcsDbPrivConfig(externalClient)
		if err != nil {
			return nil, fmt.Errorf("error creating crd: %s", err.Error())
		}
		blog.Infof("created BcsDbPrivConfig crd: %t", dbPrivCreated)

		bcsDbPrivConfigInformer := factory.Bkbcs().V1().BcsDbPrivConfigs()
		whsvr.BcsDbPrivConfigLister = bcsDbPrivConfigInformer.Lister()

		dbPrivSecret, err := whsvr.KubeClient.CoreV1().Secrets(metav1.NamespaceSystem).Get(DbPrivilegeSecretName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("error when get db privilege secret in cluster: %s", err.Error())
		}

		switch whsvr.EngineType {
		case EngineTypeKubernetes:
			k8sDbPrivConfInject := k8s.NewDbPrivConfInject(whsvr.BcsDbPrivConfigLister, whsvr.Injects, dbPrivSecret)

			whsvr.K8sDbPrivConfInject = k8sDbPrivConfInject
		case EngineTypeMesos:
			mesosDbPrivConfInject := mesos.NewDbPrivConfInject(whsvr.BcsDbPrivConfigLister)
			whsvr.MesosDbPrivConfInject = mesosDbPrivConfInject
		}

		go factory.Start(stopCh)

		blog.Infof("Waiting for BcsDbPrivConfig inormer caches to sync")
		blog.Infof("sleep 1 seconds to wait for BcsDbPrivConfig crd to be ready")
		time.Sleep(1 * time.Second)
		if ok := clientGoCache.WaitForCacheSync(stopCh, bcsDbPrivConfigInformer.Informer().HasSynced); !ok {
			return nil, fmt.Errorf("failed to wait for caches to sync")
		}
	}

	// if bscp_inject is true, init bscp inject
	if conf.Injects.Bscp.BscpInject {
		switch whsvr.EngineType {
		case EngineTypeKubernetes:
			bscpInject := k8s.NewBscpInject()
			if err := bscpInject.InitTemplate(conf.Injects.Bscp.BscpTemplatePath); err != nil {
				blog.Fatal(err.Error())
			}
			whsvr.K8sBscpInject = bscpInject
			blog.Info("create bscp k8s inject module success")
		case EngineTypeMesos:
			bscpInject := mesos.NewBscpInject()
			if err := bscpInject.InitTemplate(conf.Injects.Bscp.BscpTemplatePath); err != nil {
				blog.Fatal(err.Error())
			}
			whsvr.MesosBscpInject = bscpInject
			blog.Info("create bscp mesos inject module success")
		}
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/bcs/webhook/inject/v1/k8s", whsvr.K8sInject)
	mux.HandleFunc("/bcs/webhook/inject/v1/mesos", whsvr.MesosInject)
	whsvr.Server.Handler = mux

	return whsvr, nil
}

func parseConfig(op *options.ServerOption) *config.BcsWhsConfig {
	bcsWhsConfig := config.NewBcsLogWhsConfig()

	bcsWhsConfig.Address = op.Address
	bcsWhsConfig.Port = op.Port
	bcsWhsConfig.ServerCertFile = op.ServerCertFile
	bcsWhsConfig.ServerKeyFile = op.ServerKeyFile
	bcsWhsConfig.MetricPort = op.MetricPort
	bcsWhsConfig.EngineType = op.EngineType
	bcsWhsConfig.Kubeconfig = op.KubeConfig
	bcsWhsConfig.KubeMaster = op.KubeMaster
	bcsWhsConfig.Injects = op.Injects

	return bcsWhsConfig
}
