/*
Copyright 2016 The Kubernetes Authors.

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

package kubernetes

import (
	clientv1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	kube_record "k8s.io/client-go/tools/record"

	"k8s.io/klog"
)

const (
	// Rate of refill for the event spam filter in client go
	// 1 per event key per 5 minutes.
	defaultQPS = 1. / 300.
	// Number of events allowed per event key before rate limiting is triggered
	// Has to greater than or equal to 1.
	defaultBurstSize = 1
	// Number of distinct event keys in the rate limiting cache.
	defaultLRUCache = 8192
)

// CreateEventRecorder creates an event recorder to send custom events to Kubernetes to be recorded for targeted Kubernetes objects
func CreateEventRecorder(kubeClient clientset.Interface) kube_record.EventRecorder {
	eventBroadcaster := kube_record.NewBroadcasterWithCorrelatorOptions(getCorrelationOptions())
	eventBroadcaster.StartLogging(klog.V(4).Infof)
	if _, isfake := kubeClient.(*fake.Clientset); !isfake {
		eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubeClient.CoreV1().RESTClient()).Events("")})
	}
	return eventBroadcaster.NewRecorder(scheme.Scheme, clientv1.EventSource{Component: "cluster-autoscaler"})
}

func getCorrelationOptions() kube_record.CorrelatorOptions {
	return kube_record.CorrelatorOptions{
		QPS:          defaultQPS,
		BurstSize:    defaultBurstSize,
		LRUCacheSize: defaultLRUCache,
	}
}
