package aggregation

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/cmd/apiserver/app"
	bcs_storage "github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/bcs-storage"
	"k8s.io/api/core/v1"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"net/http"
	"sigs.k8s.io/apiserver-builder-alpha/pkg/builders"
)

var _ rest.KindProvider = &PodAggregation{}
var _ rest.Storage = &PodAggregation{}
var _ rest.Lister = &PodAggregation{}
var _ rest.TableConvertor = &PodAggregation{}
var _ rest.GetterWithOptions = &PodAggregation{}
var _ rest.Scoper = &PodAggregation{}

func NewPodAggretationREST(getter generic.RESTOptionsGetter) rest.Storage {
	return &PodAggregation{}
}

func (pa *PodAggregation) New() runtime.Object {
	return &PodAggregation{}
}

func (pa *PodAggregation) Kind() string {
	return "PodAggregation"
}

func (pa *PodAggregation) NamespaceScoped() bool {
	return true
}

func (pa *PodAggregation) NewGetOptions() (runtime.Object, bool, string) {
	builders.ParameterScheme.AddKnownTypes(SchemeGroupVersion, &PodAggregation{})
	return &PodAggregation{}, false, ""
}

func (pa *PodAggregation) Get(ctx context.Context, name string, options runtime.Object) (runtime.Object,
	error) {
	var res []PodAggregation

	fullPath, err := GetPodAggGetFullPath(ctx, name, options)
	if err != nil {
		fmt.Printf("Get func GetPodAggGetFullPath failed, %s\n", err)
		return &PodAggregationList{}, err
	}
	fmt.Printf("Get fullPath: %s\n", fullPath)

	client := &http.Client{}
	request, err := http.NewRequest("GET", fullPath, nil)
	if err != nil {
		fmt.Printf("Get func NewRequest failed, %s\n", err)
		return &PodAggregationList{}, err
	}

	if app.GetBcsStorageTokenEable() == "true" {
		var bearer = "Bearer " + app.GetBcsStorageToken()
		request.Header.Add("Authorization", bearer)
	}
	request.Header.Set("Content-type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Get func client.Do failed, %s\n", err)
		return &PodAggregationList{}, err
	}
	defer response.Body.Close()

	responseData, err := bcs_storage.DecodeResp(*response)
	if err != nil {
		fmt.Printf("Get func bcs_storage.DecodeResp failed, %s\n", err)
		return &PodAggregationList{}, err
	}

	for _, rd := range responseData {
		target := &v1.Pod{}
		if err := json.Unmarshal(rd.Data, target); err != nil {
			fmt.Printf("http storage decode data object %s failed, %s\n", "target", err)
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

func (pa *PodAggregation) NewList() runtime.Object {
	return &PodAggregationList{}
}

func (pa *PodAggregation) List(ctx context.Context, options *metainternalversion.ListOptions) (
	runtime.Object, error) {
	var res []PodAggregation

	fullPath, err := GetPodAggListFullPath(ctx, options)
	if err != nil {
		fmt.Printf("List func GetPodAggListFullPath failed, %s\n", err)
		return &PodAggregationList{}, err
	}
	fmt.Printf("List fullPath: %s\n", fullPath)

	client := &http.Client{}
	request, err := http.NewRequest("GET", fullPath, nil)
	if err != nil {
		fmt.Printf("List func http.NewRequest failed, %s\n", err)
		return &PodAggregationList{}, err
	}

	if app.GetBcsStorageTokenEable() == "true" {
		var bearer = "Bearer " + app.GetBcsStorageToken()
		request.Header.Add("Authorization", bearer)
	}
	request.Header.Set("Content-type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("List func client.Do failed, %s\n", err)
		return &PodAggregationList{}, err
	}
	defer response.Body.Close()

	responseData, err := bcs_storage.DecodeResp(*response)
	if err != nil {
		fmt.Printf("List func bcs_storage.DecodeResp failed, %s\n", err)
		return &PodAggregationList{}, err
	}

	for _, rd := range responseData {
		target := &v1.Pod{}
		if err := json.Unmarshal(rd.Data, target); err != nil {
			fmt.Printf("http storage decode data object %s failed, %s\n", "target", err)
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

func (pa *PodAggregation) ConvertToTable(ctx context.Context, object runtime.Object,
	tableOptions runtime.Object) (*metav1beta1.Table, error) {
	var table metav1beta1.Table
	return &table, nil
}

func GetPodAggGetFullPath(ctx context.Context, name string, options runtime.Object) (string, error) {
	var fullPath string

	namespace := genericapirequest.NamespaceValue(ctx)

	if len(app.GetClusterList()) == 0 {
		return "", fmt.Errorf("There is no member cluster info\n")
	}

	fullPath = fmt.Sprintf("%s?%s=%s&%s=%s&%s=%s", app.GetBcsStorageUrlBase(), "clusterId", app.GetClusterList(),
		"namespace", namespace, "resourceName", name)

	return fullPath, nil
}

func GetPodAggListFullPath(ctx context.Context, options *metainternalversion.ListOptions) (string, error) {
	var fullPath string

	namespace := genericapirequest.NamespaceValue(ctx)
	labelSelector := labels.Everything()
	if options != nil && options.LabelSelector != nil {
		labelSelector = options.LabelSelector
	}

	if len(app.GetClusterList()) == 0 {
		return "", fmt.Errorf("There is no member cluster info\n")
	}

	if namespace == "" {
		fullPath = fmt.Sprintf("%s?%s=%s", app.GetBcsStorageUrlBase(), "clusterId", app.GetClusterList())
	} else {
		fullPath = fmt.Sprintf("%s?%s=%s&%s=%s", app.GetBcsStorageUrlBase(), "clusterId", app.GetClusterList(),
			"namespace", namespace)
	}

	if labelSelector.String() != "" {
		fullPath = fmt.Sprintf("%s&%s=%s", fullPath, "labelSelector", labelSelector.String())
	}
	return fullPath, nil
}
