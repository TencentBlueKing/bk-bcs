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

package apiserver

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/pkg/storage"

	aggregationinstall "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/pkg/apis/aggregation/install"
	aggregationapi "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/pkg/apis/aggregation/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-federated-apiserver/pkg/template"
	autoscalingapiv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/admission"
	genericapi "k8s.io/apiserver/pkg/endpoints"
	genericdiscovery "k8s.io/apiserver/pkg/endpoints/discovery"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/storageversion"
	"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
)

var (
	// Scheme defines methods for serializing and deserializing API objects.
	Scheme = runtime.NewScheme()
	// Codecs provides methods for retrieving codecs and serializers for specific
	// versions and content types.
	Codecs = serializer.NewCodecFactory(Scheme)
	// ParameterCodec handles versioning of objects that are converted to query parameters.
	ParameterCodec = runtime.NewParameterCodec(Scheme)
)

const (
	aggregationGroupSuffix = ".federated.bkbcs.tencent.com"
)

func init() {
	aggregationinstall.Install(Scheme)

	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	metav1.AddToGroupVersion(Scheme, metav1.SchemeGroupVersion)

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
		&metav1.List{},
	)

	Scheme.AddUnversionedTypes(autoscalingapiv1.SchemeGroupVersion,
		&autoscalingapiv1.Scale{},
	)
}

// AggregationAPIServer will make a shadow copy for all the APIs
type AggregationAPIServer struct {
	GenericAPIServer    *genericapiserver.GenericAPIServer
	maxRequestBodyBytes int64
	minRequestTimeout   int

	// admissionControl performs deep inspection of a given request (including content)
	// to set values and determine whether its allowed
	admissionControl admission.Interface

	kubeRESTClient restclient.Interface

	config     *config.Config
	bcsStorage *storage.BcsStorage
}

func NewAggregationAPIServer(apiserver *genericapiserver.GenericAPIServer, config *config.Config,
	maxRequestBodyBytes int64, minRequestTimeout int,
	admissionControl admission.Interface,
	kubeRESTClient restclient.Interface,
	bcsStorage *storage.BcsStorage) *AggregationAPIServer {
	return &AggregationAPIServer{
		GenericAPIServer:    apiserver,
		maxRequestBodyBytes: maxRequestBodyBytes,
		minRequestTimeout:   minRequestTimeout,
		admissionControl:    admissionControl,
		kubeRESTClient:      kubeRESTClient,
		config:              config,
		bcsStorage:          bcsStorage,
	}
}

func (as *AggregationAPIServer) InstallShadowAPIGroups(stopCh <-chan struct{}, cl discovery.DiscoveryInterface) error {
	apiGroupResources, err := restmapper.GetAPIGroupResources(cl)
	if err != nil {
		return err
	}

	aggregationv1alpha1storage := map[string]rest.Storage{}
	for _, apiGroupResource := range apiGroupResources {
		// no need to duplicate xxx.federated.bkbcs.tencent.com
		if strings.HasSuffix(apiGroupResource.Group.Name, aggregationGroupSuffix) {
			continue
		}

		// skip shadow group to avoid getting nested
		if apiGroupResource.Group.Name == aggregationapi.GroupName {
			continue
		}

		for _, apiresource := range normalizeAPIGroupResources(apiGroupResource) {
			//  安需注册
			var needInstallAPIResource bool
			for _, configResource := range as.config.APIResources {
				if configResource.Group == apiresource.Group && configResource.Version == apiresource.Version && configResource.Kind == apiresource.Kind {
					needInstallAPIResource = true
					break
				}
			}
			if !needInstallAPIResource {
				continue
			}

			// register scheme for original GVK
			Scheme.AddKnownTypeWithName(schema.GroupVersion{Group: apiGroupResource.Group.Name, Version: apiresource.Version}.WithKind(apiresource.Kind),
				&unstructured.Unstructured{},
			)
			resourceRest := template.NewREST(as.kubeRESTClient, ParameterCodec, as.bcsStorage, as.config)
			resourceRest.SetNamespaceScoped(apiresource.Namespaced)
			// name 换个名称
			name := fmt.Sprintf("%saggregations", strings.ToLower(apiresource.Kind))
			resourceRest.SetName(name)
			//resourceRest.SetShortNames(apiresource.ShortNames)
			resourceRest.SetKind(apiresource.Kind)
			resourceRest.SetGroup(apiresource.Group)
			resourceRest.SetVersion(apiresource.Version)
			switch {
			case strings.HasSuffix(apiresource.Name, "/scale"):
				//TODO 单独处理scaler
			default:
				aggregationv1alpha1storage[name] = resourceRest
			}
		}
	}

	aggertationAPIGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(aggregationapi.GroupName, Scheme, ParameterCodec, Codecs)
	aggertationAPIGroupInfo.PrioritizedVersions = []schema.GroupVersion{
		{
			Group:   aggregationapi.GroupName,
			Version: aggregationapi.SchemeGroupVersion.Version,
		},
	}
	aggertationAPIGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = aggregationv1alpha1storage
	return as.installAPIGroups(&aggertationAPIGroupInfo)
}

// Exposes given api groups in the API.
// copied from k8s.io/apiserver/pkg/server/genericapiserver.go and modified
func (as *AggregationAPIServer) installAPIGroups(apiGroupInfos ...*genericapiserver.APIGroupInfo) error {
	for _, apiGroupInfo := range apiGroupInfos {
		// Do not register empty group or empty version.  Doing so claims /apis/ for the wrong entity to be returned.
		// Catching these here places the error  much closer to its origin
		if len(apiGroupInfo.PrioritizedVersions[0].Group) == 0 {
			return fmt.Errorf("cannot register handler with an empty group for %#v", *apiGroupInfo)
		}
		if len(apiGroupInfo.PrioritizedVersions[0].Version) == 0 {
			return fmt.Errorf("cannot register handler with an empty version for %#v", *apiGroupInfo)
		}
	}

	for _, apiGroupInfo := range apiGroupInfos {
		if err := as.installAPIResources(genericapiserver.APIGroupPrefix, apiGroupInfo); err != nil {
			return fmt.Errorf("unable to install api resources: %v", err)
		}

		if apiGroupInfo.PrioritizedVersions[0].String() == aggregationapi.SchemeGroupVersion.String() {
			var found bool
			for _, ws := range as.GenericAPIServer.Handler.GoRestfulContainer.RegisteredWebServices() {
				if ws.RootPath() == path.Join(genericapiserver.APIGroupPrefix, aggregationapi.SchemeGroupVersion.String()) {
					//TODO crd handler
					//as.crdHandler.SetRootWebService(ws)
					found = true
				}
			}
			if !found {
				klog.WarningDepth(2, fmt.Sprintf("failed to find a root WebServices for %s", aggregationapi.SchemeGroupVersion))
			}
		}

		// Install the version handler.
		// Add a handler at /apis/<groupName> to enumerate all versions supported by this group.
		apiVersionsForDiscovery := []metav1.GroupVersionForDiscovery{}
		for _, groupVersion := range apiGroupInfo.PrioritizedVersions {
			// Check the config to make sure that we elide versions that don't have any resources
			if len(apiGroupInfo.VersionedResourcesStorageMap[groupVersion.Version]) == 0 {
				continue
			}
			apiVersionsForDiscovery = append(apiVersionsForDiscovery, metav1.GroupVersionForDiscovery{
				GroupVersion: groupVersion.String(),
				Version:      groupVersion.Version,
			})
		}
		preferredVersionForDiscovery := metav1.GroupVersionForDiscovery{
			GroupVersion: apiGroupInfo.PrioritizedVersions[0].String(),
			Version:      apiGroupInfo.PrioritizedVersions[0].Version,
		}
		apiGroup := metav1.APIGroup{
			Name:             apiGroupInfo.PrioritizedVersions[0].Group,
			Versions:         apiVersionsForDiscovery,
			PreferredVersion: preferredVersionForDiscovery,
		}
		as.GenericAPIServer.DiscoveryGroupManager.AddGroup(apiGroup)
		as.GenericAPIServer.Handler.GoRestfulContainer.Add(genericdiscovery.NewAPIGroupHandler(as.GenericAPIServer.Serializer, apiGroup).WebService())
	}
	return nil
}

// installAPIResources is a private method for installing the REST storage backing each api groupversionresource
// copied from k8s.io/apiserver/pkg/server/genericapiserver.go and modified
func (as *AggregationAPIServer) installAPIResources(apiPrefix string, apiGroupInfo *genericapiserver.APIGroupInfo) error {
	var resourceInfos []*storageversion.ResourceInfo
	for _, groupVersion := range apiGroupInfo.PrioritizedVersions {
		if len(apiGroupInfo.VersionedResourcesStorageMap[groupVersion.Version]) == 0 {
			klog.Warningf("Skipping API %v because it has no resources.", groupVersion)
			continue
		}

		apiGroupVersion := as.getAPIGroupVersion(apiGroupInfo, groupVersion, apiPrefix)
		if apiGroupInfo.OptionsExternalVersion != nil {
			apiGroupVersion.OptionsExternalVersion = apiGroupInfo.OptionsExternalVersion
		}

		apiGroupVersion.MaxRequestBodyBytes = as.maxRequestBodyBytes
		r, err := apiGroupVersion.InstallREST(as.GenericAPIServer.Handler.GoRestfulContainer)
		if err != nil {
			return fmt.Errorf("unable to setup API %v: %v", apiGroupInfo, err)
		}

		resourceInfos = append(resourceInfos, r...)
	}

	return nil
}

// a private method that copied from k8s.io/apiserver/pkg/server/genericapiserver.go and modified
func (as *AggregationAPIServer) getAPIGroupVersion(apiGroupInfo *genericapiserver.APIGroupInfo, groupVersion schema.GroupVersion, apiPrefix string) *genericapi.APIGroupVersion {
	storage := make(map[string]rest.Storage)
	for k, v := range apiGroupInfo.VersionedResourcesStorageMap[groupVersion.Version] {
		storage[strings.ToLower(k)] = v
	}
	version := as.newAPIGroupVersion(apiGroupInfo, groupVersion)
	version.Root = apiPrefix
	version.Storage = storage
	return version
}

// a private method that copied from k8s.io/apiserver/pkg/server/genericapiserver.go and modified
func (as *AggregationAPIServer) newAPIGroupVersion(apiGroupInfo *genericapiserver.APIGroupInfo, groupVersion schema.GroupVersion) *genericapi.APIGroupVersion {
	return &genericapi.APIGroupVersion{
		GroupVersion:     groupVersion,
		MetaGroupVersion: apiGroupInfo.MetaGroupVersion,

		ParameterCodec:        apiGroupInfo.ParameterCodec,
		Serializer:            apiGroupInfo.NegotiatedSerializer,
		Creater:               apiGroupInfo.Scheme, //nolint:misspell
		Convertor:             apiGroupInfo.Scheme,
		ConvertabilityChecker: apiGroupInfo.Scheme,
		UnsafeConvertor:       runtime.UnsafeObjectConvertor(apiGroupInfo.Scheme),
		Defaulter:             apiGroupInfo.Scheme,
		Typer:                 apiGroupInfo.Scheme,
		Linker:                runtime.SelfLinker(meta.NewAccessor()),

		EquivalentResourceRegistry: as.GenericAPIServer.EquivalentResourceRegistry,

		Admit:               as.admissionControl,
		MinRequestTimeout:   time.Duration(as.minRequestTimeout) * time.Second,
		Authorizer:          as.GenericAPIServer.Authorizer,
		MaxRequestBodyBytes: as.maxRequestBodyBytes,
	}
}
