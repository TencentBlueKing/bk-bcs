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

package capabilities

import (
	"path"

	"github.com/pkg/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

var (
	// DefaultVersionSet is the default version set, which includes only Core V1 ("v1").
	defaultVersionSet = allKnownVersions()
)

// Capabilities describes the capabilities of the Kubernetes cluster.
type Capabilities struct {
	// KubeVersion is the Kubernetes version.
	KubeVersion KubeVersion
	// APIversions are supported Kubernetes API versions.
	APIVersions VersionSet
}

// GetCapabilities returns the capabilities of the Kubernetes cluster.
func GetCapabilities(dc discovery.DiscoveryInterface) (*Capabilities, error) {
	kubeVersion, err := dc.ServerVersion()
	if err != nil {
		return nil, errors.Wrap(err, "could not get server version from Kubernetes")
	}
	// Issue #6361:
	// Client-Go emits an error when an API service is registered but unimplemented.
	// We trap that error here and print a warning. But since the discovery client continues
	// building the API object, it is correctly populated with all valid APIs.
	// See https://github.com/kubernetes/kubernetes/issues/72051#issuecomment-521157642
	apiVersions, err := GetVersionSet(dc)
	if err != nil {
		if discovery.IsGroupDiscoveryFailedError(err) {
			blog.Infof("WARNING: The Kubernetes server has an orphaned API service. Server reports: %s", err)
			blog.Infof("WARNING: To fix this, kubectl delete apiservice <service-name>")
		} else {
			return nil, errors.Wrap(err, "could not get apiVersions from Kubernetes")
		}
	}
	return &Capabilities{
		KubeVersion: KubeVersion{
			Major:   kubeVersion.Major,
			Minor:   kubeVersion.Minor,
			Version: kubeVersion.GitVersion,
		},
		APIVersions: apiVersions,
	}, nil
}

// KubeVersion is the Kubernetes version.
type KubeVersion struct {
	Version string // Kubernetes version
	Major   string // Kubernetes major version
	Minor   string // Kubernetes minor version
}

// VersionSet is a set of Kubernetes API versions.
type VersionSet []string

// Has returns true if the version string is in the set.
//
//	vs.Has("apps/v1")
func (v VersionSet) Has(apiVersion string) bool {
	for _, x := range v {
		if x == apiVersion {
			return true
		}
	}
	return false
}

func allKnownVersions() VersionSet {
	// We should register the built in extension APIs as well so CRDs are
	// supported in the default version set. This has caused problems with `helm
	// template` in the past, so let's be safe
	apiextensionsv1beta1.AddToScheme(scheme.Scheme)
	apiextensionsv1.AddToScheme(scheme.Scheme)

	groups := scheme.Scheme.PrioritizedVersionsAllGroups()
	vs := make(VersionSet, 0, len(groups))
	for _, gv := range groups {
		vs = append(vs, gv.String())
	}
	return vs
}

// GetVersionSet retrieves a set of available k8s API versions
func GetVersionSet(client discovery.ServerResourcesInterface) (VersionSet, error) {
	groups, resources, err := client.ServerGroupsAndResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return defaultVersionSet, errors.Wrap(err, "could not get apiVersions from Kubernetes")
	}

	// FIXME: The Kubernetes test fixture for cli appears to always return nil
	// for calls to Discovery().ServerGroupsAndResources(). So in this case, we
	// return the default API list. This is also a safe value to return in any
	// other odd-ball case.
	if len(groups) == 0 && len(resources) == 0 {
		return defaultVersionSet, nil
	}

	versionMap := make(map[string]interface{})
	versions := []string{}

	// Extract the groups
	for _, g := range groups {
		for _, gv := range g.Versions {
			versionMap[gv.GroupVersion] = struct{}{}
		}
	}

	// Extract the resources
	var id string
	var ok bool
	for _, r := range resources {
		for _, rl := range r.APIResources {

			// A Kind at a GroupVersion can show up more than once. We only want
			// it displayed once in the final output.
			id = path.Join(r.GroupVersion, rl.Kind)
			if _, ok = versionMap[id]; !ok {
				versionMap[id] = struct{}{}
			}
		}
	}

	// Convert to a form that NewVersionSet can use
	for k := range versionMap {
		versions = append(versions, k)
	}

	return versions, nil
}
