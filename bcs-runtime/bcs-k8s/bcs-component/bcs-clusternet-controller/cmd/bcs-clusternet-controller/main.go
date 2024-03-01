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

// Package main xxx
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	clusternet "github.com/clusternet/clusternet/pkg/generated/clientset/versioned"
	informers "github.com/clusternet/clusternet/pkg/generated/informers/externalversions"
	"k8s.io/apiserver/pkg/server"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/controller-manager/pkg/clientbuilder"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-clusternet-controller/pkg/controllers/manifest"
)

const (
	defaultResyncPeriod = 15 * 60
)

var (
	kubeConfig   string
	masterURL    string
	resyncPeriod int64
)

// leader-election config options
var (
	lockNameSpace     string
	lockName          string
	lockComponentName string
	leaderElect       bool
	leaseDuration     time.Duration
	renewDeadline     time.Duration
	retryPeriod       time.Duration
)

// http server config options
var (
	address    string
	metricPort uint
)

func init() {
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Path to a kubeConfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeConfig. Only required if out-of-cluster.")
	flag.BoolVar(&leaderElect, "leader-elect", true, "Enable leader election")
	flag.StringVar(&lockNameSpace, "leader-elect-namespace", "bcs-system", "The resourcelock namespace")
	flag.StringVar(&lockName, "leader-elect-name", "bcs-clusternet-controller", "The resourcelock name")
	flag.StringVar(&lockComponentName, "leader-elect-componentname",
		"bcs-clusternet-controller", "The component name for event resource")
	flag.DurationVar(&leaseDuration, "leader-elect-lease-duration", 35*time.Second, "The leader-elect LeaseDuration")
	flag.DurationVar(&renewDeadline, "leader-elect-renew-deadline", 25*time.Second, "The leader-elect RenewDeadline")
	flag.DurationVar(&retryPeriod, "leader-elect-retry-period", 15*time.Second, "The leader-elect RetryPeriod")
	flag.StringVar(&address, "address", "0.0.0.0", "http server address")
	flag.UintVar(&metricPort, "metric-port", 10251, "prometheus metrics port")
	flag.Int64Var(&resyncPeriod, "resync-period", defaultResyncPeriod, "Time period in seconds for resync.")

}

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()
	flag.Parse()

	if !leaderElect {
		fmt.Println("No leader election, stand alone running...")
		run()
		return
	}
	clientConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		panic(err)
	}
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println("Operator build client configuration success...")

	rl, err := resourcelock.New(
		resourcelock.EndpointsResourceLock,
		lockNameSpace,
		lockName,
		kubeClient.CoreV1(),
		kubeClient.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: hostname(),
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("Operator try leader election RunOrDie...")
	// Try and become the leader and start cloud controller manager loops
	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: leaseDuration,
		RenewDeadline: renewDeadline,
		RetryPeriod:   retryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: StartedLeading,
			OnStoppedLeading: StoppedLeading,
			OnNewLeader:      NewLeader,
		},
	})
}

func run() {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := server.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)

	fmt.Printf("Rest Client Config: %v\n", cfg)

	if err != nil {
		klog.Fatalf("Error building kubeConfig: %s", err.Error())
	}

	// creating the clientset
	rootClientBuilder := clientbuilder.SimpleControllerClientBuilder{
		ClientConfig: cfg,
	}
	resyncDuration := time.Duration(resyncPeriod) * time.Second
	// kubeClient := kubernetes.NewForConfigOrDie(rootClientBuilder.ConfigOrDie("kube-client"))
	clusternetClient := clusternet.NewForConfigOrDie(rootClientBuilder.ConfigOrDie("bcs-clusternet-controller-client"))
	clusternetInformerFactory := informers.NewSharedInformerFactory(clusternetClient, resyncDuration)

	kubeClient := kubernetes.NewForConfigOrDie(rootClientBuilder.ConfigOrDie("bcs-clusternet-controller-client"))
	kubeInformerFactory := k8sinformers.NewSharedInformerFactory(kubeClient, resyncDuration)

	manifestController, err := manifest.NewController(
		clusternetClient, clusternetInformerFactory.Apps().V1alpha1().Manifests(),
		kubeInformerFactory.Core().V1().Namespaces())
	if err != nil {
		klog.Fatalf("error create manifest controller : %s", err.Error())
	}
	clusternetInformerFactory.Start(stopCh)
	kubeInformerFactory.Start(stopCh)
	manifestController.Run(1, stopCh)
}

// StartedLeading callback function
func StartedLeading(ctx context.Context) {
	fmt.Printf("%s: started leading\n", hostname())
	run()
}

// StoppedLeading invoked when this node stops being the leader
func StoppedLeading() {
	fmt.Printf("%s: stopped leading\n", hostname())
}

// NewLeader invoked when a new leader is elected
func NewLeader(id string) {
	fmt.Printf("%s: new leader: %s\n", hostname(), id)
}

func hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return hostname
}
