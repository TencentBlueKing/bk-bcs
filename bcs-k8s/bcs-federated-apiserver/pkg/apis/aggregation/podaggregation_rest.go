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

package aggregation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	bcs_storage "github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/bcs-storage"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/configuration"

	"k8s.io/api/core/v1"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-builder-alpha/pkg/builders"
)

type PodAggregationRest struct {
	acm configuration.AggregationConfigMapInfo
	aci configuration.AggregationClusterInfo
	asi configuration.AggregationBcsStorageInfo
}

var _ rest.KindProvider = &PodAggregationRest{}
var _ rest.Storage = &PodAggregationRest{}
var _ rest.Lister = &PodAggregationRest{}
var _ rest.TableConvertor = &PodAggregationRest{}
var _ rest.GetterWithOptions = &PodAggregationRest{}
var _ rest.Scoper = &PodAggregationRest{}

func NewPodAggretationREST(getter generic.RESTOptionsGetter) rest.Storage {
	var par PodAggregationRest

	// sync at background for the latest value
	go func() {
		for {
			par.acm.SetAggregationInfo()
			par.aci.SetClusterInfo(&par.acm)
			par.asi.SetBcsStorageInfo(&par.acm)
			klog.Infof("PodAggretationREST: [ %+v ]\n", par)
			time.Sleep(120 * time.Second)
		}
	}()

	// The asi and aci must be filled in at first time.
	for par.asi.GetBcsStoragePodUrlBase() == "" || par.aci.GetClusterList() == "" {
		klog.Infof("Waiting for clusterInfo and bcs-storageInfo ready, sleep...")
		time.Sleep(3 * time.Second)
	}

	return &par
}

func (pa *PodAggregationRest) New() runtime.Object {
	return &PodAggregation{}
}

func (pa *PodAggregationRest) Kind() string {
	return "PodAggregationRest"
}

func (pa *PodAggregationRest) NamespaceScoped() bool {
	return true
}

func (pa *PodAggregationRest) NewGetOptions() (runtime.Object, bool, string) {
	builders.ParameterScheme.AddKnownTypes(SchemeGroupVersion, &PodAggregation{})
	return &PodAggregation{}, false, ""
}

func (pa *PodAggregationRest) Get(ctx context.Context, name string, options runtime.Object) (runtime.Object,
	error) {
	var res []PodAggregation

	// http fullPath
	fullPath, err := GetPodAggGetFullPath(pa, ctx, name, options)
	if err != nil {
		klog.Errorf("Get func GetPodAggGetFullPath failed, %s\n", err)
		return &PodAggregationList{}, err
	}
	klog.Infof("Get fullPath: %s\n", fullPath)

	// request to bcs-storage
	response, err := bcs_storage.DoBcsStorageGetRequest(fullPath, pa.asi.GetBcsStorageToken(),
		"application/json")
	if err != nil {
		klog.Errorf("DoBcsStorageGetRequest failed, Err: %s\n", err)
		return &PodAggregationList{}, err
	}
	defer response.Body.Close()

	// Decode response json to PodAggregationList
	responseData, err := bcs_storage.DecodeResp(*response)
	if err != nil {
		klog.Errorf("Get func bcs_storage.DecodeResp failed, %s\n", err)
		return &PodAggregationList{}, err
	}

	for _, rd := range responseData {
		target := &v1.Pod{}
		if err := json.Unmarshal(rd.Data, target); err != nil {
			klog.Errorf("http storage decode data object %s failed, %s\n", "target", err)
			return &PodAggregationList{}, fmt.Errorf("json decode: %s", err)
		}

		res = append(res, PodAggregation{
			TypeMeta:   target.TypeMeta,
			ObjectMeta: target.ObjectMeta,
			Spec:       target.Spec,
			Status:     target.Status})
	}

	return &PodAggregationList{Items: res}, nil
}

func (pa *PodAggregationRest) NewList() runtime.Object {
	return &PodAggregationList{}
}

func (pa *PodAggregationRest) List(ctx context.Context, options *metainternalversion.ListOptions) (
	runtime.Object, error) {
	var res []PodAggregation

	// http fullPath
	fullPath, err := GetPodAggListFullPath(pa, ctx, options)
	if err != nil {
		klog.Errorf("List func GetPodAggListFullPath failed, %s\n", err)
		return &PodAggregationList{}, err
	}
	klog.Infof("List fullPath: %s\n", fullPath)

	// request to bcs-storage
	response, err := bcs_storage.DoBcsStorageGetRequest(fullPath, pa.asi.GetBcsStorageToken(),
		"application/json")
	if err != nil {
		klog.Errorf("DoBcsStorageGetRequest failed, Err: %s\n", err)
		return &PodAggregationList{}, err
	}
	defer response.Body.Close()

	// Decode response json to PodAggregationList
	responseData, err := bcs_storage.DecodeResp(*response)
	if err != nil {
		klog.Errorf("List func bcs_storage.DecodeResp failed, %s\n", err)
		return &PodAggregationList{}, err
	}

	for _, rd := range responseData {
		target := &v1.Pod{}
		if err := json.Unmarshal(rd.Data, target); err != nil {
			klog.Errorf("http storage decode data object %s failed, %s\n", "target", err)
			return &PodAggregationList{}, fmt.Errorf("json decode: %s", err)
		}
		res = append(res, PodAggregation{
			TypeMeta:   target.TypeMeta,
			ObjectMeta: target.ObjectMeta,
			Spec:       target.Spec,
			Status:     target.Status})
	}
	return &PodAggregationList{Items: res}, nil
}

func (pa *PodAggregationRest) ConvertToTable(ctx context.Context, object runtime.Object,
	tableOptions runtime.Object) (*metav1beta1.Table, error) {
	var table metav1beta1.Table
	return &table, nil
}

func GetPodAggGetFullPath(pa *PodAggregationRest, ctx context.Context, name string,
	options runtime.Object) (string, error) {
	var fullPath string

	namespace := genericapirequest.NamespaceValue(ctx)

	if len(pa.aci.GetClusterList()) == 0 {
		return "", fmt.Errorf("There is no member cluster info\n")
	}

	fullPath = fmt.Sprintf("%s?%s=%s&%s=%s&%s=%s", pa.asi.GetBcsStoragePodUrlBase(), "clusterId",
		pa.aci.GetClusterList(),
		"namespace", namespace, "resourceName", name)

	return fullPath, nil
}

func GetPodAggListFullPath(pa *PodAggregationRest, ctx context.Context, options *metainternalversion.ListOptions) (string, error) {
	var fullPath string

	namespace := genericapirequest.NamespaceValue(ctx)
	labelSelector := labels.Everything()
	if options != nil && options.LabelSelector != nil {
		labelSelector = options.LabelSelector
	}

	if len(pa.aci.GetClusterList()) == 0 {
		return "", fmt.Errorf("There is no member cluster info\n")
	}

	if namespace == "" {
		fullPath = fmt.Sprintf("%s?%s=%s", pa.asi.GetBcsStoragePodUrlBase(), "clusterId", pa.aci.GetClusterList())
	} else {
		fullPath = fmt.Sprintf("%s?%s=%s&%s=%s", pa.asi.GetBcsStoragePodUrlBase(), "clusterId",
			pa.aci.GetClusterList(),
			"namespace", namespace)
	}

	if labelSelector.String() != "" {
		fullPath = fmt.Sprintf("%s&%s=%s", fullPath, "labelSelector", labelSelector.String())
	}
	return fullPath, nil
}
