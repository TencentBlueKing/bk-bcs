# kubernetes自定义调度扩展

BCS通过不同层面整合kubernetes编排调度的同时，也最大程度保留了kubernetes原生的扩展能力，
可以借助kubernetes现有的扩展能力完成：

* 基于调度机制实现自定义调度扩展
* 基于CSI实现存储扩展
* 基于CNI实现网络扩展
* 基于CRI实现容器运行时扩展

以下是简单案例演示说明，如何基于kubernetes调度机制扩展调度机制，满足业务调度需求。

## 自定义调度扩展方式

kubernetes提供了三种便于扩展调度方式的机制

* 基于kubernetes开发的扩展规则，直接在原有代码进行扩展
* 实现自定义调度器，接管所有容器调度，或者与kubernetes原生调度器并行
* 实现调度扩展，实现自定义调度算法，接受kube-scheduler调度请求，可以查看[这里](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scheduling/scheduler_extender.md)

第一种方式直接调整kubernetes源码，需要重新编译kube-scheduler，代码耦合存在一定风险。
具体扩展可以参照[这里](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-scheduling/scheduler.md)。

第二种方式对golang较为友好，可以参照kube-scheduler逻辑，复用kubernetes基础库，并
借助社区工具kube-builder可以加速代码构建，[kube-builder](https://github.com/kubernetes-sigs/kubebuilder)。

第三种方式构建http(s)服务，与kube-scheduler进行集成，当有相关Pod需要调度的时候，
kube-scheduler会通过http(s)调用自定义服务的filter和prioritize接口决定最终Node
选择。该方式通过http协议集成，适合多语言环境。

本文通过第二种方式实现自定义调度扩展案例。案例仅用于演示**构建工具，构建思路**，资源
调度项**不一定具备普适性**，仅供参考。

## 扩展信息

**假设背景**

业务某类容器采用underlay方案，但是由于网络资源受限，容器节点仅能启动10个underlay容器。
CPU与Memory资源相对空闲，对于指定业务容器，调度时需要考虑underlay IP资源。

**假设方案**

利用golang实现自定义调度器ip-scheduler，针对特定业务容器，设置调度器为ip-scheduler，保障指定节点所消耗underlay资源不能超过10个。

**演示工具**

* golang
* kube-builder

```shell
$ kubebuilder version
Version: version.Version{KubeBuilderVersion:"1.0.7", KubernetesVendor:"1.11", GitCommit:"63bd3604767ddb5042fe76b67d097840a7a282c2", BuildDate:"2018-12-20T18:41:54Z", GoOs:"unknown", GoArch:"unknown"}
```

详细与完整kube-builder使用方式请参照其[详细文档](https://github.com/kubernetes-sigs/kubebuilder)

## 项目构建

**下载kubebuilder**

```shell
curl -sL https://github.com/kubernetes-sigs/kubebuilder/releases/download/v1.0.7/kubebuilder_1.0.7_linux_amd64.tar.gz -o kubebuilder_1.0.7_linux_amd64.tar.gz
tar -xzf kubebuilder_1.0.7_linux_amd64.tar.gz 
mv kubebuilder_1.0.7_linux_amd64 /usr/local/
export PATH=$PATH:/usr/local/kubebuilder_1.0.7_linux_amd64/bin
```

**创建项目ip-scheduler**

```shell
cd $GOPATH/src
mkdir ip-scheduler && cd ip-scheduler
#初始化项目，大概耗时耗时4-5minute
kubebuilder init --domain blueking.io --license apache2 --owner "The BlueKing Authors"
Run `dep ensure` to fetch dependencies (Recommended) [y/n]? y
#dep ensure
#Running make...
#make
#go generate ./pkg/... ./cmd/...
#go fmt ./pkg/... ./cmd/...
#go vet ./pkg/... ./cmd/...
#go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all
#> CRD manifests generated under '/Users/Workspace/go/src/ip-scheduler/config/crds'
#> RBAC manifests generated under '/Users/Workspace/go/src/ip-scheduler/config/rbac'
#go test ./pkg/... ./cmd/... -coverprofile cover.out
#?       ip-scheduler/pkg/apis   [no test files]
#?       ip-scheduler/pkg/controller     [no test files]
#?       ip-scheduler/pkg/webhook        [no test files]
#?       ip-scheduler/cmd/manager        [no test files]
#go build -o bin/manager ip-scheduler/cmd/manager
mkdir -p cmd/scheduler

touch cmd/scheduler/main.go
touch cmd/scheduler/scheduler.go
```

**项目目录说明**

```text
├── Dockerfile
├── Gopkg.lock
├── Gopkg.toml
├── Makefile
├── PROJECT
├── bin
│   └── manager
├── cmd
│   ├── manager
│   └── scheduler
├── config
│   ├── crds
│   ├── default
│   ├── manager
│   └── rbac
├── cover.out
├── hack
│   └── boilerplate.go.txt
├── pkg
│   ├── apis
│   ├── controller
│   └── webhook
└── vendor
    ├── cloud.google.com
    ├── github.com
    ├── go.uber.org
    ├── golang.org
    ├── google.golang.org
    ├── gopkg.in
    ├── k8s.io
    └── sigs.k8s.io
```

目录结构是典型k8s项目结构：
* cmd/scheduler目录：ip-scheduler程序
* pkg/apis：如果调度需要引入crd资源，生成的crd对象在该目录中，**本次演示中不包含crd部分**

代码文件说明：
* cmd/scheduler/main.go：scheduler main入口
* cmd/scheduler/scheduler.go：简易的scheduler封装

## main.go

```golang
/*
Copyright 2019 The BlueKing Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/informers"
	kubeClientSet "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	log     logr.Logger
	cfgFile string
	ipNum   int
)

func init() {
	logf.SetLogger(logf.ZapLogger(false))
	log = logf.Log.WithName("ip-scheduler")
	flag.StringVar(&cfgFile, "config", "./kube.yaml", "ip-scheduler kubernetes configuration")
	flag.IntVar(&ipNum, "ip-resource", 10, "ip resource for every node, only for demonstration")
}

func main() {
	flag.Parse()
	//building clientset configuration
	restConfig, err := clientcmd.BuildConfigFromFlags("", cfgFile)
	if err != nil {
		log.Error(err, "blueking ip-scheduler builds clientConfig with kubeconfig ./kube.yaml")
		os.Exit(1)
	}
	//process kubernetes clientSet for kubeInformerFactory
	clientset, err := kubeClientSet.NewForConfig(restConfig)
	if err != nil {
		log.Error(err, "bmsf-networkpolicy-controller create kubernetes clientset with restConfig failed")
		os.Exit(1)
	}
	//create SharedInformerFactory with ClientSet
	factories := informers.NewSharedInformerFactory(clientset, time.Second*180)
	s := newscheduler(clientset.CoreV1(), factories.Core().V1().Pods(), factories.Core().V1().Nodes())
	handleSignal(s)
	if err := s.run(factories); err != nil {
		log.Error(err, "ip-scheduler force to exit")
		os.Exit(1)
	}
	//wait for exit signal
	s.wait()
}

// HandleSignal handle system signal for exit
func handleSignal(s *scheduler) {
	signalChan := make(chan os.Signal, 5)
	signal.Notify(signalChan, syscall.SIGTRAP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-signalChan:
			log.V(0).Info("blueking ip-scheduler was killed.", "signal", sig.String())
			s.stop()
			time.Sleep(time.Second * 3)
		}
	}()
}
```

## 简单scheduler封装

功能目标：
* 监听集群所有node，针对node设置IP资源计数
* 根据IP资源计数，将Pod调度到IP资源做大的节点

```golang
/*
Copyright 2019 The BlueKing Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/informers/core/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	schedulerName = "blueking-ip-scheduler"
)

//newscheduler initialize custom scheduler for IP resource demonstration
func newscheduler(pc coreclient.CoreV1Interface, pods corev1.PodInformer, nodes corev1.NodeInformer) *scheduler {
	s := &scheduler{
		stopCh:       make(chan struct{}),
		podClient:    pc,
		podInformer:  pods,
		nodeInformer: nodes,
		queue:        workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		ipResource:   make(map[string]int),
	}
	//register Pod event callback to Informer
	s.podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: s.onPodAdd,
		//UpdateFunc: s.onPodUpdate,
		DeleteFunc: s.onPodDelete,
	})
	s.nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: s.onNodeAdd,
		//UpdateFunc: s.onNodeUpdate,
		DeleteFunc: s.onNodeDelete,
	})
	return s
}

type scheduler struct {
	stopCh chan struct{}
	//client for Pod Operation
	podClient coreclient.CoreV1Interface
	//informers for event watching
	nodeInformer corev1.NodeInformer
	podInformer  corev1.PodInformer
	//queue for retry
	queue       workqueue.RateLimitingInterface
	retryLocker sync.Mutex
	//ipResource cachew
	ipResource map[string]int
	ipLocker   sync.Mutex
}

func (s *scheduler) run(factory informers.SharedInformerFactory) error {
	factory.Start(s.stopCh)
	kubeFlags := factory.WaitForCacheSync(s.stopCh)
	for t, state := range kubeFlags {
		if !state {
			return fmt.Errorf("%s informer cache init failure", t.Name())
		}
	}
	go wait.Until(s.handleQueue, time.Millisecond*100, s.stopCh)
	return nil
}

//handleQueue when ip resource limited, push to
// queue and retry again
func (s *scheduler) handleQueue() {
	keyObj, shutdown := s.queue.Get()
	if shutdown {
		return
	}
	s.retryLocker.Lock()
	defer s.retryLocker.Unlock()
	key := keyObj.(string)
	//check local cache
	obj, exist, err := s.podInformer.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		log.Error(err, "retry Pod failed, get from local cache failed.", "podID", key)
		return
	}
	if !exist {
		//Pod deleted, no scheduling needed
		s.queue.Forget(keyObj)
		s.queue.Done(keyObj)
		return
	}
	pod := obj.(*v1.Pod)
	defer s.queue.Done(keyObj)
	if err := s.schedule(pod); err != nil {
		log.Error(err, "scheduler re-process Pod stil failed.", "podId", key)
		//check reschedule limit
		retry := s.queue.NumRequeues(keyObj)
		if retry < 100 {
			s.queue.AddRateLimited(keyObj)
			log.Info("Pod retry is stil effective, push back to queue again", "podID", key, "retryCnt", retry)
		} else {
			log.Info("Pod retry is beyond limit system setting, forget scheduling", "podID", key, "retryCnt", retry)
			s.queue.Forget(keyObj)
		}
	}
}

func (s *scheduler) wait() {
	<-s.stopCh
}

func (s *scheduler) stop() {
	close(s.stopCh)
	s.queue.ShutDown()
}

func (s *scheduler) onPodAdd(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return
	}
	// Only schedule pending pods, also ignore abnormal pod
	if !(pod.Spec.NodeName == "" && pod.Spec.SchedulerName == schedulerName &&
		pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed) {
		log.V(0).Info("Pod is not under ip-scheduler control", "podId", key)
		return
	}
	log.Info("ip-scheduler Ready to schedule pod first onPodAddEvent", "podId", key)
	if err := s.schedule(pod); err != nil {
		log.Error(err, "ip-scheduler first schedule pod failed, push to retry queue to recover", "podId", key)
		//push to reschedule
		key, _ := cache.MetaNamespaceKeyFunc(pod)
		s.queue.Add(key)
	}
}

func (s *scheduler) onPodUpdate(old, cur interface{}) {

}

func (s *scheduler) onPodDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}
	podName := fmt.Sprintf("%s/%s", pod.GetNamespace(), pod.GetName())
	nodeName := pod.Spec.NodeName
	if len(nodeName) == 0 {
		log.Info("Pod have no Node scheduling, skip IP resource release", "podId", podName)
		return
	}
	s.retryLocker.Lock()
	defer s.retryLocker.Unlock()
	if pod.Spec.SchedulerName == schedulerName {
		//schedule management scope
		log.Info("Pod is under deletion, ip-scheduler release ip resource for this pod", "podId", podName, "nodeName", nodeName)
		s.releaseIPResource(nodeName)
		return
	}
	log.Info("Pod is not under ip-scheduler control", "podId", podName)
}

func (s *scheduler) onNodeAdd(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		return
	}
	//check node role
	if _, isMaster := node.Labels["node-role.kubernetes.io/master"]; isMaster {
		log.Info("Node is master role, it's not under scheduling scope. skip", "nodeName", node.GetName())
		return
	}
	s.ipLocker.Lock()
	defer s.ipLocker.Unlock()
	key := node.GetName()
	if _, found := s.ipResource[key]; !found {
		log.Info("Node is adding, initialize IP resource...", "nodeName", key)
		s.ipResource[key] = ipNum
	}
}

func (s *scheduler) onNodeUpdate(old, cur interface{}) {

}

func (s *scheduler) onNodeDelete(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		return
	}
	key := node.GetName()
	s.ipLocker.Lock()
	defer s.ipLocker.Unlock()
	if _, found := s.ipResource[key]; found {
		delete(s.ipResource, key)
		log.Info("Node is deleting, clean resource from local cache...", "nodeName", key)
	}
}

func (s *scheduler) schedule(pod *v1.Pod) error {
	nodeName := s.acquireIPResource()
	if len(nodeName) == 0 {
		log.Error(fmt.Errorf("ip-resource runs out"), "Pod can not be scheduled because of ip resource running out", "namespace", pod.GetNamespace(), "podName", pod.GetName())
		return fmt.Errorf("lack of resource")
	}
	//assign Pod to this Node
	binding := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{Namespace: pod.GetNamespace(), Name: pod.GetName(), UID: pod.UID},
		Target: v1.ObjectReference{
			Kind: "Node",
			Name: nodeName,
		},
	}
	log.Info("Attempting to bind pod to node", "namespace", pod.GetNamespace(), "podName", pod.GetName(), "nodeName", nodeName)
	if err := s.podClient.Pods(binding.Namespace).Bind(binding); err != nil {
		log.Error(err, "Pod is binding to Node failed", "namespace", pod.GetNamespace(), "podName", pod.GetName(), "nodeName", nodeName)
		s.releaseIPResource(nodeName)
		return fmt.Errorf("binding failed")
	}
	log.Info("Binding pod to node successfully", "namespace", pod.GetNamespace(), "podName", pod.GetName(), "nodeName", nodeName)
	return nil
}

//releaseIPResource release IP resource to local cache
func (s *scheduler) releaseIPResource(node string) {
	s.ipLocker.Lock()
	defer s.ipLocker.Unlock()
	ipNum, found := s.ipResource[node]
	if !found {
		log.Info("Node do not exist in local cache", "nodeName", node)
		return
	}
	s.ipResource[node] = ipNum + 1
	log.Info("Node IP resource release successfully", "nodeName", node)
}

//acquireIPResource get max IP resource node from local cache
func (s *scheduler) acquireIPResource() string {
	s.ipLocker.Lock()
	defer s.ipLocker.Unlock()
	//get max resource node
	ipNum := 0
	node := ""
	for name, ipResource := range s.ipResource {
		if ipResource > ipNum {
			node = name
			ipNum = ipResource
		}
	}
	if ipNum == 0 {
		return ""
	}
	s.ipResource[node] = ipNum - 1
	log.Info("Node IP resource acquire successfully", "nodeName", node)
	return node
}
```

## makefile调整与编译

```makefile
# Build custom ip-scheduler binary
scheduler: generate fmt vet
	go build -o bin/bk-ip-scheduler ip-scheduler/cmd/scheduler
```

```shell
make scheduler
go generate ./pkg/... ./cmd/...
go fmt ./pkg/... ./cmd/...
go vet ./pkg/... ./cmd/...
go build -o bin/bk-ip-scheduler ip-scheduler/cmd/scheduler
```

## 配置与启动

配置文件kube.yaml
```yaml
apiVersion: v1
clusters:
- cluster:
    server: http://127.0.0.1:8080
  name: scheduler
contexts:
- context:
    cluster: scheduler
    user: ""
  name: scheduler
current-context: scheduler
```

启动
```shell
./bk-ip-scheduler --config ./kube.yaml
```

## 容器配置说明

启动容器时，需要制定调度器为blueking-ip-scheduler。
测试文件demo.yaml:

```yaml
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  labels:
    app: blueking-demo
  name: bk-demo
  namespace: blueking
spec:
  replicas: 1
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: blueking-demo
    spec:
      containers:
      - name: bk-demo
        command: ['python']
        args:
        - -m
        - SimpleHTTPServer
        - "8080"
        image: python:v2.7.10
        imagePullPolicy: IfNotPresent
        resources: {}
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: blueking-ip-scheduler
```

```shell
kubectl create -f demo.yaml
```
