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

package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
)

// ApplicationDiffOrDryRunRequest defines the request for application diff or dry-run.
// User can specify application exist, or transfer application manifests which not
// created.
type ApplicationDiffOrDryRunRequest struct {
	ApplicationName string   `json:"applicationName"`
	Revision        string   `json:"revision"`
	Revisions       []string `json:"revisions"`

	ApplicationManifests string `json:"applicationManifests"`
}

// ApplicationDiffOrDryRunResponse defines the response of application diff or dry-run
type ApplicationDiffOrDryRunResponse struct {
	Code      int32                        `json:"code"`
	Message   string                       `json:"message"`
	RequestID string                       `json:"requestID"`
	Data      *ApplicationDiffOrDryRunData `json:"data"`
}

// ApplicationDiffOrDryRunData defines the data of application diff or dry-run
type ApplicationDiffOrDryRunData struct {
	Cluster       string `json:"cluster,omitempty"`
	RepoUrl       string `json:"repoUrl,omitempty"`
	LocalRevision string `json:"localRevision,omitempty"`
	LiveRevision  string `json:"liveRevision,omitempty"`

	Result []*Manifest `json:"result,omitempty"`
}

// Manifest defines the dry-run result for every resource
type Manifest struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Group     string `json:"group,omitempty"`
	Version   string `json:"version,omitempty"`
	Kind      string `json:"kind,omitempty"`

	// diff result
	Local *unstructured.Unstructured `json:"local,omitempty"`
	Live  *unstructured.Unstructured `json:"live,omitempty"`

	// dyr-run result
	Existed    bool                       `json:"existed"`
	IsSucceed  bool                       `json:"isSucceed"`
	ErrMessage string                     `json:"errMessage,omitempty"`
	Merged     *unstructured.Unstructured `json:"merged,omitempty"`
}

func (plugin *AppPlugin) applicationDiff(r *http.Request) (*http.Request, *mw.HttpResponse) {
	req, err := buildDiffOrDryRunRequest(r)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "build request failed"))
	}
	blog.Infof("RequestID[%s] received diff request: %v", mw.RequestID(r.Context()), req)
	application, err := plugin.buildApplicationByDifOrDryRunRequest(r.Context(), req)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "build application from request failed"))
	}
	result, err := plugin.compareApplicationManifests(r.Context(), req, application.DeepCopy(), false)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "compare application manifests failed"))
	}
	return r, mw.ReturnJSONResponse(&ApplicationDiffOrDryRunResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data: &ApplicationDiffOrDryRunData{
			Cluster:       application.Spec.Destination.Server,
			RepoUrl:       application.Spec.Source.RepoURL,
			LocalRevision: req.Revision,
			LiveRevision:  application.Spec.Source.TargetRevision,
			Result:        result,
		},
	})
}

// applicationDryRun will generate application manifests with argocd first. And then
// compare every resource with live in Kubernetes cluster.
func (plugin *AppPlugin) applicationDryRun(r *http.Request) (*http.Request, *mw.HttpResponse) {
	req, err := buildDiffOrDryRunRequest(r)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "build request failed"))
	}
	blog.Infof("RequestID[%s] received dry-run request: %v", mw.RequestID(r.Context()), req)
	application, err := plugin.buildApplicationByDifOrDryRunRequest(r.Context(), req)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "build application from request failed"))
	}
	result, err := plugin.compareApplicationManifests(r.Context(), req, application, true)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "compare application manifests failed"))
	}
	result, err = plugin.manifestDryRun(r.Context(), application, result)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "manifests dry run failed"))
	}
	resp := &ApplicationDiffOrDryRunResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data: &ApplicationDiffOrDryRunData{
			Cluster:       application.Spec.Destination.Server,
			RepoUrl:       application.Spec.Source.RepoURL,
			LocalRevision: req.Revision,
			LiveRevision:  application.Spec.Source.TargetRevision,
			Result:        result,
		},
	}
	return r, mw.ReturnJSONResponse(resp)
}

// nolint
func (plugin *AppPlugin) buildApplicationByDifOrDryRunRequest(ctx context.Context,
	req *ApplicationDiffOrDryRunRequest) (*v1alpha1.Application, error) {
	application := new(v1alpha1.Application)
	if req.ApplicationName != "" {
		var statusCode int
		var err error
		application, statusCode, err = plugin.middleware.CheckApplicationPermission(ctx, req.ApplicationName,
			iamnamespace.NameSpaceScopedView)
		if statusCode != http.StatusOK {
			return nil, errors.Wrapf(err, "check application permission failed")
		}
	} else if req.ApplicationManifests != "" {
		var statusCode int
		var err error
		if err = json.Unmarshal([]byte(req.ApplicationManifests), application); err != nil {
			return nil, errors.Wrapf(err, "unmarshal applicationManifests failed")
		}
		statusCode, err = plugin.middleware.CheckCreateApplication(ctx, application)
		if statusCode != http.StatusOK {
			return nil, errors.Wrapf(err, "check create application failed")
		}
		blog.Infof("RequestID[%s] checked create application permission", mw.RequestID(ctx))
	} else {
		return nil, errors.Errorf("request body 'applicationName' or 'applicationManifests' cannot be empty")
	}
	if application.Spec.HasMultipleSources() {
		if len(req.Revisions) != 0 && len(req.Revisions) != len(application.Spec.Sources) {
			return nil, errors.Errorf("application has '%d' resources, but request body 'revisions' only "+
				"have '%d'. Or you can set empty 'revisions', we will use default revisions in application",
				len(application.Spec.Sources), len(req.Revisions))
		}
		return application, nil
	}
	if application.Spec.Source == nil {
		return nil, errors.Errorf("application spec.source is nil")
	}
	return application, nil
}

func buildDiffOrDryRunRequest(r *http.Request) (*ApplicationDiffOrDryRunRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read body failed")
	}
	req := new(ApplicationDiffOrDryRunRequest)
	if err = json.Unmarshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal request body failed")
	}
	if req.ApplicationName == "" && req.ApplicationManifests == "" {
		return nil, errors.Errorf("applicationName/revision or applicationManifests should in request")
	}
	return req, nil
}

func (plugin *AppPlugin) compareApplicationManifests(ctx context.Context, req *ApplicationDiffOrDryRunRequest,
	application *v1alpha1.Application, isDryRun bool) ([]*Manifest, error) {
	originalApplication := application.DeepCopy()
	// 使用 Request 的 Revision 版本覆盖应用的版本
	if len(req.Revisions) != 0 {
		for i := range application.Spec.Sources {
			application.Spec.Sources[i].TargetRevision = req.Revisions[i]
		}
	} else if req.Revision != "" {
		application.Spec.Source.TargetRevision = req.Revision
	}
	resp, err := plugin.storage.GetApplicationManifestsFromRepoServerWithMultiSources(ctx, application)
	if err != nil {
		return nil, errors.Wrapf(err, "get application manifests failed")
	}
	if len(resp) == 0 {
		return nil, errors.Errorf("application manifests response length is 0")
	}
	manifestResponse := resp[0]
	manifestMap, err := decodeManifest(application, manifestResponse)
	if err != nil {
		return nil, errors.Wrapf(err, "decode application manifest failed")
	}
	// 未指定 revision 或只指定 manifest，表示只是 DryRun 不需要对比，直接返回
	if (req.Revision == "" && len(req.Revisions) == 0) || req.ApplicationManifests != "" || isDryRun {
		results := make([]*Manifest, 0, len(manifestMap))
		for _, decodeObj := range manifestMap {
			results = append(results, &Manifest{
				Namespace: decodeObj.GetNamespace(),
				Name:      decodeObj.GetName(),
				Group:     decodeObj.GroupVersionKind().Group,
				Version:   decodeObj.GroupVersionKind().Version,
				Kind:      decodeObj.GetKind(),
				Local:     decodeObj,
			})
		}
		return results, nil
	}

	// 获取已存在的 application 的 revision 结果
	liveManifestResp, err := plugin.storage.
		GetApplicationManifestsFromRepoServerWithMultiSources(ctx, originalApplication)
	if err != nil {
		return nil, errors.Wrapf(err, "get application manifests with live failed")
	}
	if len(liveManifestResp) == 0 {
		return nil, errors.Errorf("application manifests response length is 0")
	}
	liveManifestMap, err := decodeManifest(originalApplication, liveManifestResp[0])
	if err != nil {
		return nil, errors.Wrapf(err, "decode live application manifest failed")
	}
	results := make([]*Manifest, 0)
	// 对比目标 revision 和已保存的 revision
	for k, decodeObj := range manifestMap {
		liveDecodeObj, ok := liveManifestMap[k]
		if !ok {
			continue
		}
		if err = plugin.storage.ApplicationNormalizeWhenDiff(originalApplication,
			decodeObj, liveDecodeObj, true); err != nil {
			return nil, errors.Wrapf(err, "application normlize failed when diff")
		}
		results = append(results, &Manifest{
			Namespace: decodeObj.GetNamespace(),
			Name:      decodeObj.GetName(),
			Group:     decodeObj.GroupVersionKind().Group,
			Version:   decodeObj.GroupVersionKind().Version,
			Kind:      decodeObj.GetKind(),
			Local:     decodeObj,
			Live:      liveDecodeObj,
		})
		delete(manifestMap, k)
		delete(liveManifestMap, k)
	}
	for _, obj := range manifestMap {
		obj.SetManagedFields(nil)
		results = append(results, &Manifest{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
			Group:     obj.GroupVersionKind().Group,
			Version:   obj.GroupVersionKind().Version,
			Kind:      obj.GetKind(),
			Local:     obj,
		})
	}
	for _, obj := range liveManifestMap {
		obj.SetManagedFields(nil)
		results = append(results, &Manifest{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
			Group:     obj.GroupVersionKind().Group,
			Version:   obj.GroupVersionKind().Version,
			Kind:      obj.GetKind(),
			Live:      obj,
		})
	}
	return results, nil
}

func decodeManifest(application *v1alpha1.Application,
	resp *apiclient.ManifestResponse) (map[string]*unstructured.Unstructured, error) {
	result := make(map[string]*unstructured.Unstructured)
	decode := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer().Decode
	for i := range resp.Manifests {
		manifest := resp.Manifests[i]
		obj, _, err := decode([]byte(manifest), nil, &unstructured.Unstructured{})
		if err != nil {
			return nil, errors.Wrapf(err, "decode manifest '%s' failed", manifest)
		}
		manifestObj := obj.(*unstructured.Unstructured)
		manifestObj.SetNamespace(application.Spec.Destination.Namespace)
		if manifestObj.GetNamespace() == "" {
			manifestObj.SetNamespace("default")
		}
		key := manifestObj.GroupVersionKind().String() + manifestObj.GetNamespace() + manifestObj.GetName()
		result[key] = manifestObj
	}
	return result, nil
}

// manifestDryRun will check every resource's manifest existed in Kubernetes cluster.
// If not exist, just save the manifest and no need to compare. If existed, there need
// to compare with live.
func (plugin *AppPlugin) manifestDryRun(ctx context.Context, app *v1alpha1.Application,
	manifests []*Manifest) ([]*Manifest, error) {
	clusterServer := app.Spec.Destination.Server
	cluster, err := plugin.storage.GetClusterFromDB(ctx, clusterServer)
	if err != nil {
		return nil, errors.Wrapf(err, "get cluster '%s' failed", clusterServer)
	}
	config := &rest.Config{
		Host:        app.Spec.Destination.Server,
		BearerToken: cluster.Config.BearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	gvr, err := getGroupVersionKindResource(config)
	if err != nil {
		return nil, errors.Wrapf(err, "get cluster '%s' gvr failed", clusterServer)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create dynamic client for cluster '%s' failed", clusterServer)
	}

	results := make([]*Manifest, 0, len(manifests))
	for _, manifest := range manifests {
		local := manifest.Local
		dryRunManifest := &Manifest{
			Namespace: local.GetNamespace(),
			Name:      local.GetName(),
			Group:     local.GroupVersionKind().Group,
			Version:   local.GroupVersionKind().Version,
			Kind:      local.GetKind(),
		}
		gvk := local.GroupVersionKind().GroupVersion().String() + "/" + local.GetKind()
		resource, ok := gvr[gvk]
		if !ok {
			dryRunManifest.IsSucceed = false
			dryRunManifest.ErrMessage = fmt.Sprintf("cluster '%s' get resource from gvk '%s' not found",
				clusterServer, gvk)
			results = append(results, dryRunManifest)
			continue
		}

		localGVR := local.GroupVersionKind().GroupVersion().WithResource(resource.Name)
		localName := local.GetName()
		localNamespace := local.GetNamespace()
		// NOTE: there should set namespace empty if resource is not namespaced
		if !resource.Namespaced {
			localNamespace = ""
			dryRunManifest.Namespace = localNamespace
		}
		_, err = dynamicClient.Resource(localGVR).Namespace(localNamespace).
			Get(context.Background(), localName, metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			dryRunManifest.IsSucceed = false
			dryRunManifest.ErrMessage = fmt.Sprintf("get resource '%s' with gvr '%v' from namespace '%s' failed: %s",
				local.GetName(), localGVR, localNamespace, err.Error())
			results = append(results, dryRunManifest)
			continue
		}
		var updatedObj *unstructured.Unstructured
		var dryRunError error
		var isExisted bool
		if k8serrors.IsNotFound(err) {
			isExisted = false
			updatedObj, err = dynamicClient.Resource(localGVR).Namespace(localNamespace).
				Create(context.Background(), local,
					metav1.CreateOptions{
						DryRun:       []string{metav1.DryRunAll},
						FieldManager: "kubectl-client-side-apply",
					})
			if err != nil {
				dryRunError = errors.Wrapf(err, "dry-run not exist resource '%s' with gvr '%v' "+
					"and namespace '%s' failed", local.GetName(), localGVR, localNamespace)
			}
		} else {
			var localBS []byte
			localBS, err = local.MarshalJSON()
			if err != nil {
				return nil, errors.Wrapf(err, "marshal local object json failed")
			}
			isExisted = true
			updatedObj, err = dynamicClient.
				Resource(localGVR).
				Namespace(localNamespace).
				Patch(context.Background(), local.GetName(), types.MergePatchType,
					localBS, metav1.PatchOptions{
						DryRun:       []string{metav1.DryRunAll},
						FieldManager: "kubectl-client-side-apply",
					})
			if err != nil {
				dryRunError = errors.Wrapf(err, "dry-run with exist resource '%s' with gvr '%v' "+
					"and namespace '%s' failed", local.GetName(), localGVR, localNamespace)
			}
		}
		dryRunManifest.Existed = isExisted
		if dryRunError == nil {
			updatedObj.SetManagedFields(nil)
			dryRunManifest.IsSucceed = true
			dryRunManifest.Merged = updatedObj
		} else {
			dryRunManifest.IsSucceed = false
			dryRunManifest.ErrMessage = dryRunError.Error()
		}
		results = append(results, dryRunManifest)
	}
	return results, nil
}

// getGroupVersionKindResource returns all GroupVersionResource from cluster
func getGroupVersionKindResource(config *rest.Config) (map[string]metav1.APIResource, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create discovery client failed")
	}
	_, resources, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return nil, errors.Wrapf(err, "get server groups and resources failed")
	}
	result := make(map[string]metav1.APIResource)
	for _, resourceList := range resources {
		for i := range resourceList.APIResources {
			res := resourceList.APIResources[i]
			if strings.Contains(res.Name, "/") {
				continue
			}
			result[resourceList.GroupVersion+"/"+res.Kind] = res
		}
	}
	return result, nil
}
