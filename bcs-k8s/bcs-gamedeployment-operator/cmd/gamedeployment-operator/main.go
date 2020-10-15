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

package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	clientset "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/clientset/versioned"
	informers "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/informers/externalversions"
	gamedeploy "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/controllers"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/server"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

const (
	metricsEndpoint = "0.0.0.0:8080"
)

var (
	kubeConfig string
	masterURL  string
	//MinResyncPeriod period definition
	MinResyncPeriod metav1.Duration
)

// leader-election config options
var (
	LockNameSpace     string
	LockName          string
	LockComponentName string
	LeaderElect       bool
	LeaseDuration     time.Duration
	RenewDeadline     time.Duration
	RetryPeriod       time.Duration
)

func main() {

	flag.Parse()

	if !LeaderElect {
		fmt.Println("No leader election, stand alone running...")
		run()
		return
	}
	clientConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println("Operator build client configuration success...")
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println("Operator building kube client for election success...")
	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: corev1.New(kubeClient.CoreV1().RESTClient()).Events(LockNameSpace)})
	recorder := broadcaster.NewRecorder(scheme.Scheme, api.EventSource{Component: LockComponentName})

	rl, err := resourcelock.New(
		resourcelock.EndpointsResourceLock,
		LockNameSpace,
		LockName,
		kubeClient.CoreV1(),
		resourcelock.ResourceLockConfig{
			Identity:      hostname(),
			EventRecorder: recorder,
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("Operator try leader election RunOrDie...")
	// Try and become the leader and start cloud controller manager loops
	leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: LeaseDuration,
		RenewDeadline: RenewDeadline,
		RetryPeriod:   RetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: StartedLeading,
			OnStoppedLeading: StoppedLeading,
			OnNewLeader:      NewLeader,
		},
	})

}

func init() {
	flag.StringVar(&kubeConfig, "kubeConfig", "", "Path to a kubeConfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeConfig. Only required if out-of-cluster.")
	flag.DurationVar(&MinResyncPeriod.Duration, "min-resync-period", 10*time.Minute, "The resync period in reflectors will be random between MinResyncPeriod and 2*MinResyncPeriod.")
	flag.BoolVar(&LeaderElect, "leader-elect", true, "Enable leader election")
	flag.StringVar(&LockNameSpace, "leader-elect-namespace", "bcs-system", "The resourcelock namespace")
	flag.StringVar(&LockName, "leader-elect-name", "gamedeployment", "The resourcelock name")
	flag.StringVar(&LockComponentName, "leader-elect-componentname", "gamedeployment", "The component name for event resource")
	flag.DurationVar(&LeaseDuration, "leader-elect-lease-duration", 15*time.Second, "The leader-elect LeaseDuration")
	flag.DurationVar(&RenewDeadline, "leader-elect-renew-deadline", 10*time.Second, "The leader-elect RenewDeadline")
	flag.DurationVar(&RetryPeriod, "leader-elect-retry-period", 2*time.Second, "The leader-elect RetryPeriod")
}

// resyncPeriod computes the time interval a shared informer waits before
// resyncing with the api server.
func resyncPeriod(MinResyncPeriod metav1.Duration) func() time.Duration {
	return func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(float64(MinResyncPeriod.Nanoseconds()) * factor)
	}
}

func run() {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := server.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)

	fmt.Printf("Rest Client Config: %v\n", cfg)

	if err != nil {
		klog.Fatalf("Error building kubeConfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	fmt.Println("Operator builds kube client success...")
	tkexClient, err := clientset.NewForConfig(cfg)

	if err != nil {
		klog.Fatalf("Error building gamedeployment clientset: %s", err.Error())
	}
	fmt.Println("Operator builds tkex client success...")
	//kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, resyncPeriod(MinResyncPeriod)())
	//gameDeploymentInformerFactory := informers.NewSharedInformerFactory(tkexClient, resyncPeriod(MinResyncPeriod)())
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, 0)
	gameDeploymentInformerFactory := informers.NewSharedInformerFactory(tkexClient, 0)

	gdController := gamedeploy.NewGameDeploymentController(
		kubeInformerFactory.Core().V1().Pods(),
		gameDeploymentInformerFactory.Tkex().V1alpha1().GameDeployments(),
		kubeInformerFactory.Apps().V1().ControllerRevisions(),
		kubeClient,
		tkexClient)

	go kubeInformerFactory.Start(stopCh)
	fmt.Println("Operator starting kube Informer Factory success...")
	go gameDeploymentInformerFactory.Start(stopCh)
	fmt.Println("Operator starting tkex Informer factory success...")

	if err = gdController.Run(1, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
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
