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
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/controllers/hook"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/constants"
	clientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned"
	informers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions"
	_ "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/metrics/restclient"
	_ "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/metrics/workqueue"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/api/core/v1"
	api "k8s.io/api/core/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
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
	DefaultResyncPeriod = 15 * 60
)

var (
	kubeConfig   string
	masterURL    string
	resyncPeriod int64
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

// http server config options
var (
	address    string
	metricPort uint
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
		kubeClient.CoordinationV1(),
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
	flag.Int64Var(&resyncPeriod, "resync-period", DefaultResyncPeriod, "Time period in seconds for resync.")
	flag.BoolVar(&LeaderElect, "leader-elect", false, "Enable leader election")
	flag.StringVar(&LockNameSpace, "leader-elect-namespace", "bcs-system", "The resourcelock namespace")
	flag.StringVar(&LockName, "leader-elect-name", "bcs-hook-operator", "The resourcelock name")
	flag.StringVar(&LockComponentName, "leader-elect-componentname", "bcs-hook-operator", "The component name for event resource")
	flag.DurationVar(&LeaseDuration, "leader-elect-lease-duration", 15*time.Second, "The leader-elect LeaseDuration")
	flag.DurationVar(&RenewDeadline, "leader-elect-renew-deadline", 10*time.Second, "The leader-elect RenewDeadline")
	flag.DurationVar(&RetryPeriod, "leader-elect-retry-period", 2*time.Second, "The leader-elect RetryPeriod")
	flag.StringVar(&address, "address", "0.0.0.0", "http server address")
	flag.UintVar(&metricPort, "metric-port", 10251, "prometheus metrics port")
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

	apiextensionClient, err := apiextension.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building apiextension clientset: %s", err.Error())
	}
	fmt.Println("Operator builds apiextension client success...")

	tkexClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building hook clientset: %s", err.Error())
	}
	fmt.Println("Operator builds tkex client success...")

	resyncDuration := time.Duration(resyncPeriod) * time.Second
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, resyncDuration)
	hookInformerFactory := informers.NewSharedInformerFactory(tkexClient, resyncDuration)

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: constants.OperatorName})

	hrController := hook.NewHookController(
		kubeClient,
		apiextensionClient,
		tkexClient,
		hookInformerFactory.Tkex().V1alpha1().HookRuns(),
		recorder,
	)

	kubeInformerFactory.Start(stopCh)
	fmt.Println("Operator starting kube Informer Factory success...")
	hookInformerFactory.Start(stopCh)
	fmt.Println("Operator starting tkex Informer factory success...")
	runPrometheusMetricsServer()
	fmt.Println("run prometheus server metrics success...")

	if err = hrController.Run(1, stopCh); err != nil {
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

//runPrometheusMetrics starting prometheus metrics handler
func runPrometheusMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	addr := address + ":" + strconv.Itoa(int(metricPort))
	go http.ListenAndServe(addr, nil)
}

func hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return hostname
}
