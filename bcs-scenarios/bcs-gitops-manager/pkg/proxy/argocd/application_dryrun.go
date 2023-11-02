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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
)

// ApplicationDiffOrDryRunRequest defines the request for application diff or dry-run.
// User can specify application exist, or transfer application manifests which not
// created.
type ApplicationDiffOrDryRunRequest struct {
	ApplicationName string `json:"applicationName"`
	Revision        string `json:"revision"`

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
	application, result, err := plugin.compareApplicationManifests(r.Context(), req, false)
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
	application, result, err := plugin.compareApplicationManifests(r.Context(), req, true)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "compare application manifests failed"))
	}
	result, err = plugin.manifestDryRun(r.Context(), application, result)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "manifests dry run failed"))
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

func buildDiffOrDryRunRequest(r *http.Request) (*ApplicationDiffOrDryRunRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read body failed")
	}
	req := new(ApplicationDiffOrDryRunRequest)
	if err := json.Unmarshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal request body failed")
	}
	if req.ApplicationName == "" && req.ApplicationManifests == "" {
		return nil, errors.Errorf("applicationName/revision or applicationManifests should in request")
	}
	return req, nil
}

func (plugin *AppPlugin) compareApplicationManifests(ctx context.Context,
	req *ApplicationDiffOrDryRunRequest, isDryRun bool) (*v1alpha1.Application, []*Manifest, error) {
	var application *v1alpha1.Application
	var manifestResponse *apiclient.ManifestResponse
	var err error
	if req.ApplicationName != "" {
		application, manifestResponse, err = plugin.buildApplicationManifestsWithName(ctx,
			req.ApplicationName, req.Revision)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "build application manfiests failed")
		}
	} else if req.ApplicationManifests != "" {
		application, manifestResponse, err = plugin.buildApplicationManifestsWithAppYaml(ctx, req.ApplicationManifests)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "build application manifests failed")
		}
	}
	manifestMap, err := decodeManifest(application, manifestResponse)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "decode application manifest failed")
	}
	// 未指定 revision 或指定 manifest，表示不需要对比，直接返回
	if req.Revision == "" || req.ApplicationManifests != "" || isDryRun {
		results := make([]*Manifest, 0, len(manifestMap))
		for _, decodeObj := range manifestMap {
			results = append(results, &Manifest{
				Namespace: decodeObj.GetNamespace(),
				Name:      decodeObj.GetName(),
				Kind:      decodeObj.GetKind(),
				Local:     decodeObj,
			})
		}
		return application, results, nil
	}

	// 获取已存在的 application 的 revision 结果
	liveManifestResp, err := plugin.storage.GetApplicationManifestsFromRepoServer(ctx, application)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "get application manifests with live failed")
	}
	liveManifestMap, err := decodeManifest(application, liveManifestResp)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "decode live application manifest failed")
	}
	results := make([]*Manifest, 0)
	// 对比目标 revision 和已保存的 revision
	for k, decodeObj := range manifestMap {
		liveDecodeObj, ok := liveManifestMap[k]
		if !ok {
			continue
		}
		if err = plugin.storage.ApplicationNormalizeWhenDiff(application,
			decodeObj, liveDecodeObj, true); err != nil {
			return nil, nil, errors.Wrapf(err, "application normlize failed when diff")
		}
		results = append(results, &Manifest{
			Namespace: decodeObj.GetNamespace(),
			Name:      decodeObj.GetName(),
			Kind:      decodeObj.GetKind(),
			Local:     decodeObj,
			Live:      liveDecodeObj,
		})
		delete(manifestMap, k)
		delete(liveManifestMap, k)
	}
	for _, obj := range manifestMap {
		results = append(results, &Manifest{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
			Kind:      obj.GetKind(),
			Local:     obj,
		})
	}
	for _, obj := range liveManifestMap {
		results = append(results, &Manifest{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
			Kind:      obj.GetKind(),
			Live:      obj,
		})
	}
	return application, results, nil
}

func (plugin *AppPlugin) buildApplicationManifestsWithName(ctx context.Context,
	appName, revision string) (*v1alpha1.Application, *apiclient.ManifestResponse, error) {
	application, statusCode, err := plugin.middleware.CheckApplicationPermission(ctx, appName, iam.ProjectView)
	if statusCode != http.StatusOK {
		return nil, nil, errors.Wrapf(err, "check application permission failed")
	}
	originalRevision := application.Spec.Source.TargetRevision
	if revision != "" {
		application.Spec.Source.TargetRevision = revision
	}
	blog.Infof("RequestID[%s] checked application permission", mw.RequestID(ctx))
	var manifestResp *apiclient.ManifestResponse
	manifestResp, err = plugin.storage.GetApplicationManifestsFromRepoServer(ctx, application)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "get application manifests failed")
	}

	application.Spec.Source.TargetRevision = originalRevision
	return application, manifestResp, nil
}

func (plugin *AppPlugin) buildApplicationManifestsWithAppYaml(ctx context.Context, appYaml string) (
	*v1alpha1.Application, *apiclient.ManifestResponse, error) {
	application := new(v1alpha1.Application)
	if err := json.Unmarshal([]byte(appYaml), application); err != nil {
		return nil, nil, errors.Wrapf(err, "unmarshal applicationManifests failed")
	}
	statusCode, err := plugin.middleware.CheckCreateApplication(ctx, application)
	if statusCode != http.StatusOK {
		return nil, nil, errors.Wrapf(err, "check create application failed")
	}
	blog.Infof("RequestID[%s] checked create application permission", mw.RequestID(ctx))
	var manifestResp *apiclient.ManifestResponse
	manifestResp, err = plugin.storage.GetApplicationManifestsFromRepoServer(ctx, application)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "get application manifests without created failed")
	}
	return application, manifestResp, nil
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
		gvk := local.GroupVersionKind().GroupVersion().String() + "/" + local.GetKind()
		resource, ok := gvr[gvk]
		if !ok {
			return nil, errors.Errorf("cluster '%s' get resource from gvk '%s' not found", clusterServer, gvk)
		}
		_, err = dynamicClient.
			Resource(local.GroupVersionKind().GroupVersion().WithResource(resource)).
			Namespace(local.GetNamespace()).
			Get(context.Background(), local.GetName(), metav1.GetOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return nil, errors.Wrapf(err, "get resource '%s/%s/%s' failed",
				resource, local.GetNamespace(), local.GetName())
		}
		var updatedObj *unstructured.Unstructured
		var dryRunError error
		var isExisted bool
		if k8serrors.IsNotFound(err) {
			isExisted = false
			updatedObj, err = dynamicClient.
				Resource(local.GroupVersionKind().GroupVersion().WithResource(resource)).
				Namespace(local.GetNamespace()).Create(context.Background(), local, metav1.CreateOptions{
				DryRun:       []string{metav1.DryRunAll},
				FieldManager: "kubectl-client-side-apply",
			})
			if err != nil {
				dryRunError = errors.Wrapf(err, "dry-run with resource not exist failed")
			}
		} else {
			var localBS []byte
			localBS, err = local.MarshalJSON()
			if err != nil {
				return nil, errors.Wrapf(err, "marshal local object json failed")
			}
			isExisted = true
			updatedObj, err = dynamicClient.
				Resource(local.GroupVersionKind().GroupVersion().WithResource(resource)).
				Namespace(local.GetNamespace()).Patch(context.Background(), local.GetName(), types.ApplyPatchType,
				localBS, metav1.PatchOptions{
					DryRun:       []string{metav1.DryRunAll},
					FieldManager: "kubectl-client-side-apply",
				})
			if err != nil {
				dryRunError = errors.Wrapf(err, "dry-run with resource exist failed")
			}
		}
		dryRunManifest := &Manifest{
			Namespace: updatedObj.GetNamespace(),
			Name:      updatedObj.GetName(),
			Kind:      updatedObj.GetKind(),
			Existed:   isExisted,
		}
		if dryRunError == nil {
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
func getGroupVersionKindResource(config *rest.Config) (map[string]string, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "create discovery client failed")
	}
	_, resources, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return nil, errors.Wrapf(err, "get server groups and resources failed")
	}
	result := make(map[string]string)
	for _, resourceList := range resources {
		for i := range resourceList.APIResources {
			res := resourceList.APIResources[i]
			if strings.Contains(res.Name, "/") {
				continue
			}
			result[resourceList.GroupVersion+"/"+res.Kind] = res.Name
		}
	}
	return result, nil
}
