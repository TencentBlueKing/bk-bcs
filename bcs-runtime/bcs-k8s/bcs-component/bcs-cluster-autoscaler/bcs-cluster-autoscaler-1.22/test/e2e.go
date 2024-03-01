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

// Package main implements e2e test function
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v12 "k8s.io/client-go/listers/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	klog "k8s.io/klog/v2"
)

var (
	desiredReplicas = flag.Int("desired-replicas", 1000, "desired replicas")
	round           = flag.Int("round", 10, "how many round to test")
	cpu             = flag.String("cpu", "10", "cpu setting")
	externalIP      = flag.Bool("external-ip", false, "if we need eni ip")
	sleepTime       = flag.Duration("sleep-time", 2*time.Minute, "if we need eni ip")
	namespace       = flag.String("namespace", "default", "if we need eni ip")
	address         = flag.String("address", ":8086", "metrics address")
	uid             = flag.String("uid", "", "uid of this pod")
	name            = flag.String("name", "", "name of this pod")
)

// testConfig config of test
type testConfig struct {
	desiredReplicas int32
	client          kubernetes.Interface
	name            string
	namespace       string
	podTemplateSpec v1.PodTemplateSpec
}

// newConfig news testConfig
func newConfig() *testConfig {
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		panic(err)
	}
	client := kubernetes.NewForConfigOrDie(kubeConfig)
	cpuRes := resource.MustParse(*cpu)
	pod := v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:            "nginx",
				Image:           "nginx:latest",
				ImagePullPolicy: v1.PullIfNotPresent,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU: cpuRes,
					},
					Limits: v1.ResourceList{
						v1.ResourceCPU: cpuRes,
					},
				},
			},
		},
	}
	if *externalIP {
		pod.Containers[0].Resources.Requests["tke.cloud.tencent.com/direct-eni"] = resource.MustParse("1")
		pod.Containers[0].Resources.Limits["tke.cloud.tencent.com/direct-eni"] = resource.MustParse("1")
	}
	return &testConfig{
		desiredReplicas: int32(*desiredReplicas),
		client:          client,
		namespace:       *namespace,
		podTemplateSpec: v1.PodTemplateSpec{
			Spec: pod,
		},
	}
}

// produceName produce name
func (tc *testConfig) produceName() {
	name := fmt.Sprintf("ca-%v", string(uuid.NewUUID()))
	tc.name = name
	tc.podTemplateSpec.ObjectMeta.Labels = map[string]string{"ca-test": name}
}

// ScaleUpWorkLoad xxx
func (tc *testConfig) ScaleUpWorkLoad(deploy *appsv1.Deployment, lister v12.PodLister) error {
	klog.Infof("E2E start scale up workload %v", tc.name)
	var failed bool
	var changeScale bool
	defer func() {
		if failed {
			failedScaleUpCount.Inc()
		}
		if changeScale {
			scaleUpCount.Inc()
		}
		klog.Infof("E2E finish scale up workload %v", tc.name)
	}()
	now := time.Now()
	defer func() {
		end := time.Now()
		gap := end.Sub(now)
		scaleUpSeconds.Observe(gap.Seconds())
		klog.Infof("Finish scale up, cost: %v", gap.String())
	}()
	if err := tc.changeScale(deploy, tc.desiredReplicas); err != nil {
		klog.Errorf("ScaleUpWorkLoad err: %v", err)
		return err
	}
	klog.Info("Scale Up WorkLoad operation success")
	changeScale = true
	if err := wait.PollImmediate(5*time.Second, 10*time.Minute, func() (done bool, err error) {
		return tc.ReconcileScaleUp(lister), nil
	}); err != nil {
		failed = true
		klog.Errorf("ScaleUpWorkLoad err: %v", err)
		return err
	}
	return nil
}

// ScaleDownWorkLoad trys to scale down the workload
func (tc *testConfig) ScaleDownWorkLoad(deploy *appsv1.Deployment, lister v12.NodeLister, desired int) error {
	klog.Infof("E2E start scale down workload %v", tc.name)
	var failed bool
	var changeScale bool
	defer func() {
		if failed {
			failedScaleDownCount.Inc()
		}
		if changeScale {
			scaleDownCount.Inc()
		}
		klog.Infof("E2E finish scale down workload %v", tc.name)
	}()
	now := time.Now()
	defer func() {
		end := time.Now()
		gap := end.Sub(now)
		scaleDownSeconds.Observe(gap.Seconds())
		klog.Infof("Finish scale down, cost: %v", gap.String())
	}()
	if err := tc.changeScale(deploy, 0); err != nil {
		klog.Errorf("ScaleDownWorkLoad err: %v", err)
		return err
	}
	klog.Info("Scale Down WorkLoad operation success")
	changeScale = true
	if err := wait.PollImmediate(5*time.Second, 10*time.Minute, func() (done bool, err error) {
		return tc.ReconcileScaleDown(lister, desired), nil
	}); err != nil {
		failed = true
		klog.Error(err)
		return err
	}
	return nil
}

// changeScale change scale
func (tc *testConfig) changeScale(deploy *appsv1.Deployment, desired int32) error {
	return wait.PollImmediate(1*time.Second, 5*time.Second, func() (done bool, err error) {
		scale, err := tc.client.AppsV1().Deployments(deploy.Namespace).GetScale(context.TODO(), deploy.Name,
			metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		scale.Spec.Replicas = desired
		if _, err := tc.client.AppsV1().Deployments(deploy.Namespace).UpdateScale(
			context.TODO(), deploy.Name, scale, metav1.UpdateOptions{}); err != nil {
			return false, nil

		}
		return true, nil
	})
}

// CreateWorkLoad creates a workload
func (tc *testConfig) CreateWorkLoad() *appsv1.Deployment {
	klog.Infof("E2E start create workload %v", tc.name)
	defer func() {
		klog.Infof("E2E finish create workload %v", tc.name)
	}()
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tc.name,
			Namespace: tc.namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       *name,
					UID:        types.UID(*uid),
				},
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: tc.podTemplateSpec.Labels},
			Replicas: new(int32),
			Template: tc.podTemplateSpec,
		},
	}
	if deployment, err := tc.client.AppsV1().Deployments(tc.namespace).Create(
		context.TODO(), deploy, metav1.CreateOptions{}); err != nil {
		panic(err)
	} else {
		return deployment
	}
}

// DeleteWorkLoad deletes the workload
func (tc *testConfig) DeleteWorkLoad() {
	klog.Infof("E2E start delete workload %v", tc.name)
	defer func() {
		klog.Infof("E2E finish delete workload %v", tc.name)
	}()
	// nolint
	wait.PollImmediate(1*time.Second, 5*time.Second, func() (done bool, err error) {
		if err := tc.client.AppsV1().Deployments(tc.namespace).Delete(context.TODO(), tc.name,
			metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
			klog.Errorf("Delete deploy %v", err)
			return false, nil
		}
		return true, nil
	})
}

// ReconcileScaleUp checks the count and desiredReplicas
func (tc *testConfig) ReconcileScaleUp(lister v12.PodLister) bool {
	labelSelector := labels.SelectorFromSet(tc.podTemplateSpec.Labels)
	pods, err := lister.Pods(tc.namespace).List(labelSelector)
	if err != nil {
		panic(err)
	}
	count := 0
	for _, pod := range pods {
		if pod.Spec.NodeName == "" {
			continue
		}
		count++
	}
	return count >= int(tc.desiredReplicas)
}

// ReconcileScaleDown checks the real and desired number of nodes.
func (tc *testConfig) ReconcileScaleDown(lister v12.NodeLister, desired int) bool {
	nodes, err := lister.List(labels.Everything())
	if err != nil {
		panic(err)
	}
	return len(nodes) <= desired
}

// getNodeCount return count of nodes
func (tc *testConfig) getNodeCount(lister v12.NodeLister) int {
	nodes, err := lister.List(labels.Everything())
	if err != nil {
		panic(err)
	}
	return len(nodes)
}

func main() {
	defer klog.Flush()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(*address, nil)
		klog.Fatalf("Failed to start metrics: %v", err)
	}()
	registerAll()
	stop := server.SetupSignalHandler()
	tc := newConfig()
	coreFactory := informers.NewSharedInformerFactoryWithOptions(tc.client, 0,
		informers.WithNamespace(tc.namespace))
	podSynced := coreFactory.Core().V1().Pods().Informer().HasSynced
	nodeSynced := coreFactory.Core().V1().Nodes().Informer().HasSynced
	podLister := coreFactory.Core().V1().Pods().Lister()
	nodeLister := coreFactory.Core().V1().Nodes().Lister()
	coreFactory.Start(stop)
	cache.WaitForCacheSync(stop, podSynced, nodeSynced)
	successScaleUp := 0
	successScaleDown := 0
	for i := 0; i < *round; i++ {
		finished := make(chan struct{})
		func() {
			go func() {
				select {
				case <-stop:
					tc.DeleteWorkLoad()
				case <-finished:
					return

				}
			}()
			tc.produceName()
			klog.Infof("----------------------------E2E start round %v-------------------------", i+1)
			defer klog.Infof("----------------------------E2E stop round %v-------------------------", i+1)
			deploy := tc.CreateWorkLoad()
			defer tc.DeleteWorkLoad()
			defer close(finished)
			deployCopy := deploy.DeepCopy()
			nodeNum := tc.getNodeCount(nodeLister)
			if err := tc.ScaleUpWorkLoad(deployCopy, podLister); err != nil {
				klog.Error(err)
				return
			}
			successScaleUp++
			scaleUpSuccessRate.Set(float64(successScaleUp*100) / float64(i+1))
			time.Sleep(3 * time.Minute)
			if err := tc.ScaleDownWorkLoad(deployCopy, nodeLister, nodeNum); err != nil {
				klog.Error(err)
				return
			}
			successScaleDown++
			scaleDownSuccessRate.Set(float64(successScaleDown*100) / float64(i+1))
		}()
		time.Sleep(*sleepTime)
		klog.Infof("Finial success scale up rate: %v/100", successScaleUp*100/(i+1))
		klog.Infof("Finial success scale down rate: %v/100", successScaleDown*100/(i+1))
	}
}
