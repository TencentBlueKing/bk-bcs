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

package argocd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/stream"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
)

// ApplicationDryRunRequest defines the dry-run request
type ApplicationDryRunRequest struct {
	ApplicationName string   `json:"applicationName"`
	Revision        string   `json:"revision"`
	Revisions       []string `json:"revisions"`

	ApplicationManifests string `json:"applicationManifests"`
}

// ApplicationDryRunResponse defines the response of application dry-run
type ApplicationDryRunResponse struct {
	Code      int32                  `json:"code"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"requestID"`
	Data      *ApplicationDryRunData `json:"data"`
}

// ApplicationDryRunData defines the data of application dry-run
type ApplicationDryRunData struct {
	Cluster       string `json:"cluster,omitempty"`
	RepoUrl       string `json:"repoUrl,omitempty"`
	LocalRevision string `json:"localRevision,omitempty"`
	LiveRevision  string `json:"liveRevision,omitempty"`

	Result []*ApplicationDryRunManifest `json:"result,omitempty"`
}

// ApplicationDryRunManifest defines the dry-run result for every resource
type ApplicationDryRunManifest struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Group     string `json:"group,omitempty"`
	Version   string `json:"version,omitempty"`
	Kind      string `json:"kind,omitempty"`

	RenderObj *unstructured.Unstructured `json:"-"`

	// dyr-run result
	Existed    bool                       `json:"existed"`
	IsSucceed  bool                       `json:"isSucceed"`
	ErrMessage string                     `json:"errMessage,omitempty"`
	Merged     *unstructured.Unstructured `json:"merged,omitempty"`
}

// applicationDryRun will generate application manifests with argocd first. And then
// compare every resource with live in Kubernetes cluster.
func (plugin *AppPlugin) applicationDryRun(r *http.Request) (*http.Request, *mw.HttpResponse) {
	req, err := buildDryRunRequest(r)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "build request failed"))
	}
	blog.Infof("RequestID[%s] received dry-run request: %v", ctxutils.RequestID(r.Context()), req)
	application, err := plugin.checkApplicationDryRunRequest(r.Context(), req)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "build application from request failed"))
	}
	result, err := plugin.handleApplicationDryRun(r.Context(), req, application)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "compare application manifests failed"))
	}
	result, err = plugin.manifestDryRun(r.Context(), application, result)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "manifests dry run failed"))
	}
	resp := &ApplicationDryRunResponse{
		Code:      0,
		RequestID: ctxutils.RequestID(r.Context()),
		Data: &ApplicationDryRunData{
			Cluster:       application.Spec.Destination.Server,
			RepoUrl:       application.Spec.GetSource().RepoURL,
			LocalRevision: req.Revision,
			LiveRevision:  application.Spec.GetSource().TargetRevision,
			Result:        result,
		},
	}
	return r, mw.ReturnJSONResponse(resp)
}

func (plugin *AppPlugin) handleApplicationDryRun(ctx context.Context, req *ApplicationDryRunRequest,
	application *v1alpha1.Application) ([]*ApplicationDryRunManifest, error) {
	if application.Spec.HasMultipleSources() {
		for i := range req.Revisions {
			if i < len(application.Spec.Sources) {
				application.Spec.Sources[i].TargetRevision = req.Revisions[i]
			}
		}
	} else if req.Revision != "" {
		application.Spec.Source.TargetRevision = req.Revision
	}
	// 获取应用对应目标 Revision 的 manifest
	manifestMap, err := plugin.applicationManifests(ctx, application)
	if err != nil {
		return nil, errors.Wrapf(err, "handle application dry-run manifests failed")
	}
	results := make([]*ApplicationDryRunManifest, 0, len(manifestMap))
	for _, decodeObj := range manifestMap {
		results = append(results, &ApplicationDryRunManifest{
			Namespace: decodeObj.GetNamespace(),
			Name:      decodeObj.GetName(),
			Group:     decodeObj.GroupVersionKind().Group,
			Version:   decodeObj.GroupVersionKind().Version,
			Kind:      decodeObj.GetKind(),
			RenderObj: decodeObj,
		})
	}
	return results, nil
}

// nolint
func (plugin *AppPlugin) checkApplicationDryRunRequest(ctx context.Context,
	req *ApplicationDryRunRequest) (*v1alpha1.Application, error) {
	argoApplication := new(v1alpha1.Application)
	if req.ApplicationName != "" {
		var err error
		argoApplication, _, err = plugin.permitChecker.CheckApplicationPermission(ctx, req.ApplicationName,
			permitcheck.AppViewRSAction)
		if err != nil {
			return nil, errors.Wrapf(err, "check application permission failed")
		}
	} else if req.ApplicationManifests != "" {
		var err error
		if err = json.Unmarshal([]byte(req.ApplicationManifests), argoApplication); err != nil {
			return nil, errors.Wrapf(err, "unmarshal applicationManifests failed")
		}
		_, err = plugin.permitChecker.CheckApplicationCreate(ctx, argoApplication)
		if err != nil {
			return nil, errors.Wrapf(err, "check create application failed")
		}
		blog.Infof("RequestID[%s] checked create application permission", ctxutils.RequestID(ctx))
	} else {
		return nil, errors.Errorf("request body 'applicationName' or 'applicationManifests' cannot be empty")
	}
	if argoApplication.Spec.HasMultipleSources() {
		if len(req.Revisions) != 0 && len(req.Revisions) > len(argoApplication.Spec.Sources) {
			return nil, errors.Errorf("application has '%d' resources but 'revisions' have '%d' items",
				len(argoApplication.Spec.Sources), len(req.Revisions))
		}
		return argoApplication, nil
	}
	if argoApplication.Spec.Source == nil {
		return nil, errors.Errorf("application spec.source is nil")
	}
	return argoApplication, nil
}

func buildDryRunRequest(r *http.Request) (*ApplicationDryRunRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read body failed")
	}
	req := new(ApplicationDryRunRequest)
	if err = json.Unmarshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal request body failed")
	}
	if req.ApplicationName == "" && req.ApplicationManifests == "" {
		return nil, errors.Errorf("applicationName/revision or applicationManifests should in request")
	}
	return req, nil
}

func (plugin *AppPlugin) applicationManifests(ctx context.Context, application *v1alpha1.Application) (
	map[string]*unstructured.Unstructured, error) {
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
	blog.Infof("RequestID[%s] application get manifest success", ctxutils.RequestID(ctx))
	return manifestMap, nil
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
// nolint
func (plugin *AppPlugin) manifestDryRun(ctx context.Context, app *v1alpha1.Application,
	manifests []*ApplicationDryRunManifest) ([]*ApplicationDryRunManifest, error) {
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

	drHandler := &dryRunHandler{
		manifests:     manifests,
		gvr:           gvr,
		clusterServer: clusterServer,
		dynamicClient: dynamicClient,
		concurrent:    100,
	}
	results := drHandler.Handle(ctx)
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

type dryRunHandler struct {
	manifests     []*ApplicationDryRunManifest
	gvr           map[string]metav1.APIResource
	clusterServer string
	dynamicClient dynamic.Interface

	concurrent int
}

func (h *dryRunHandler) Handle(ctx context.Context) []*ApplicationDryRunManifest {
	results := make([]*ApplicationDryRunManifest, 0, len(h.manifests))
	s := stream.New().WithMaxGoroutines(h.concurrent)
	for _, elem := range h.manifests {
		newElem := elem
		s.Go(func() stream.Callback {
			mf := h.dryRunManifest(ctx, newElem)
			results = append(results, mf)
			return func() {}
		})
	}
	s.Wait()
	return results
}

func (h *dryRunHandler) dryRunManifest(ctx context.Context,
	manifest *ApplicationDryRunManifest) *ApplicationDryRunManifest {
	local := manifest.RenderObj
	dryRunManifest := &ApplicationDryRunManifest{
		Namespace: local.GetNamespace(),
		Name:      local.GetName(),
		Group:     local.GroupVersionKind().Group,
		Version:   local.GroupVersionKind().Version,
		Kind:      local.GetKind(),
	}
	gvk := local.GroupVersionKind().GroupVersion().String() + "/" + local.GetKind()
	resource, ok := h.gvr[gvk]
	if !ok {
		dryRunManifest.IsSucceed = false
		dryRunManifest.ErrMessage = fmt.Sprintf("cluster '%s' get resource from gvk '%s' not found",
			h.clusterServer, gvk)
		return dryRunManifest
	}

	localGVR := local.GroupVersionKind().GroupVersion().WithResource(resource.Name)
	localName := local.GetName()
	localNamespace := local.GetNamespace()
	// NOTE: there should set namespace empty if resource is not namespaced
	if !resource.Namespaced {
		localNamespace = ""
		dryRunManifest.Namespace = localNamespace
	}
	_, err := h.dynamicClient.Resource(localGVR).Namespace(localNamespace).
		Get(ctx, localName, metav1.GetOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		dryRunManifest.IsSucceed = false
		dryRunManifest.ErrMessage = fmt.Sprintf("get resource '%s' with gvr '%v' from namespace '%s' failed: %s",
			local.GetName(), localGVR, localNamespace, err.Error())
		return dryRunManifest
	}
	updatedObj, isExisted, dryRunError := h.dryRunActualResource(ctx, local, !k8serrors.IsNotFound(err),
		resource, localGVR, localNamespace)
	dryRunManifest.Existed = isExisted
	if dryRunError == nil {
		if updatedObj != nil {
			updatedObj.SetManagedFields(nil)
		}
		dryRunManifest.IsSucceed = true
		dryRunManifest.Merged = updatedObj
	} else {
		dryRunManifest.IsSucceed = false
		dryRunManifest.ErrMessage = dryRunError.Error()
	}
	return dryRunManifest
}

func (h *dryRunHandler) dryRunActualResource(ctx context.Context, local *unstructured.Unstructured, exist bool,
	localResource metav1.APIResource,
	localGVR schema.GroupVersionResource,
	localNamespace string) (*unstructured.Unstructured, bool, error) {
	var updatedObj *unstructured.Unstructured
	var dryRunError error
	var isExisted bool
	// there will patch resource with dry-run if resource exist
	if exist {
		localBS, err := local.MarshalJSON()
		if err != nil {
			dryRunError = errors.Wrapf(err, "marshal local object json failed")
		} else {
			isExisted = true
			if updatedObj, err = h.dynamicClient.
				Resource(localGVR).
				Namespace(localNamespace).
				Patch(ctx, local.GetName(), types.MergePatchType,
					localBS, metav1.PatchOptions{
						DryRun:       []string{metav1.DryRunAll},
						FieldManager: "kubectl-client-side-apply",
					}); err != nil {
				dryRunError = errors.Wrapf(err, "dry-run with exist resource '%s' with gvr '%v' "+
					"and namespace '%s' failed", local.GetName(), localGVR, localNamespace)
			}
		}
		return updatedObj, isExisted, dryRunError
	}

	// there will create resource with dry-run if resource not exist
	isExisted = false
	var err error
	if !localResource.Namespaced {
		// dry-run object when resource not namespaced
		updatedObj, err = h.dynamicClient.Resource(localGVR).Namespace(localNamespace).
			Create(ctx, local,
				metav1.CreateOptions{
					DryRun:       []string{metav1.DryRunAll},
					FieldManager: "kubectl-client-side-apply",
				})
		if err != nil {
			dryRunError = errors.Wrapf(err, "dry-run not exist resource '%s' with gvr '%v' "+
				"and namespace '%s' failed", local.GetName(), localGVR, localNamespace)
		}
	} else {
		if _, err = h.dynamicClient.Resource(schema.GroupVersionResource{Group: "", Version: "v1",
			Resource: "namespaces"}).Get(ctx, localNamespace, metav1.GetOptions{}); err != nil {
			// return local if namespace not exist, no need dry-run it
			if k8serrors.IsNotFound(err) {
				updatedObj = local.DeepCopy()
			} else {
				dryRunError = errors.Wrapf(err, "get namespace '%s' failed", localNamespace)
			}
		} else {
			// dry-run object with exist namespace
			updatedObj, err = h.dynamicClient.Resource(localGVR).Namespace(localNamespace).
				Create(ctx, local,
					metav1.CreateOptions{
						DryRun:       []string{metav1.DryRunAll},
						FieldManager: "kubectl-client-side-apply",
					})
			if err != nil {
				dryRunError = errors.Wrapf(err, "dry-run not exist resource '%s' with gvr '%v' "+
					"and namespace '%s' failed", local.GetName(), localGVR, localNamespace)
			}
		}
	}
	return updatedObj, isExisted, dryRunError
}
