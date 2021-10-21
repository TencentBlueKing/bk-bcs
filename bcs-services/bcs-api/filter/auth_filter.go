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

package filter

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/auth/bkiam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"

	"github.com/emicklei/go-restful"
)

const (
	// BcsClusterIDHeaderKey key for http header
	BcsClusterIDHeaderKey = "BCS-ClusterID"
	// ApiPrefix bcs-api prefix
	ApiPrefix = `/bcsapi/[^/]+/[^/]+`
)

// NewAuthFilter filter creator
func NewAuthFilter(conf *config.ApiServConfig) (RequestFilterFunction, error) {
	myAuth, err := bkiam.NewAuth(conf)
	if err != nil {
		return nil, err
	}

	return &AuthFilter{
		conf: conf,
		auth: myAuth,
	}, nil
}

// AuthFilter auth filter for all bcs-api request
type AuthFilter struct {
	conf *config.ApiServConfig
	auth auth.BcsAuth
}

// Execute check authorization
func (af *AuthFilter) Execute(req *restful.Request) (errCode int, err error) {
	token, err := af.auth.GetToken(req.Request.Header)
	if err != nil {
		return common.BcsErrApiAuthCheckFail, fmt.Errorf("%s: %v", common.BcsErrApiAuthCheckFailStr, err)
	}

	clusterID := req.Request.Header.Get(BcsClusterIDHeaderKey)
	method := req.Request.Method
	uri := req.Request.URL.Path

	var authRuleList []*AuthURLRule

	switch AuthRuleRegex.ReplaceAllString(uri, "$1") {
	case "storage":
		authRuleList = StorageAuthRule
	case "metric":
		authRuleList = MetricAuthRule
	case "scheduler":
		if AuthRuleRegex.ReplaceAllString(uri, "$2") == "mesos" {
			authRuleList = MesosAuthRule
			break
		}
		fallthrough
	default:
		// no rule match the uri prefix, then just let it pass
		return 0, nil
	}

	for _, rule := range authRuleList {
		match, action, resource := rule.Match(clusterID, "", uri, method)
		if match {
			// no cluster id means can not be check auth, just let it pass
			if resource.ClusterID == "" {
				return 0, nil
			}
			ok, err := af.auth.Allow(token, action, resource)
			if err != nil {
				blog.Errorf("AuthFilter Execute get auth allow failed: %v", err)
				return common.BcsErrApiAuthCheckFail, fmt.Errorf("%s: %s", common.BcsErrApiAuthCheckFailStr, err.Error())
			}
			if !ok {
				return common.BcsErrApiAuthCheckNoAuthority, fmt.Errorf(common.BcsErrApiAuthCheckFailStr)
			}
			return 0, nil
		}
	}

	// no rule list match the uri, then just let it pass
	return 0, nil
}

const (
	// ClusterIDSignTag tag for bkbcs cluster ID
	ClusterIDSignTag = "{clusterId}"
	// NamespaceSignTag tag for bkbcs namespace in URL
	NamespaceSignTag = "{namespace}"
)

// AuthURLRule URL rule for dispatch
type AuthURLRule struct {
	Rule   string
	Method string
	Action auth.Action

	regex            *regexp.Regexp
	clusterIDSubSign string
	namespaceSubSign string
}

func (aur *AuthURLRule) init() {
	aur.Rule = fmt.Sprintf(`^%s/*$`, path.Join(ApiPrefix, aur.Rule))
	clusterIDIndex := strings.Index(aur.Rule, ClusterIDSignTag)
	namespaceIndex := strings.Index(aur.Rule, NamespaceSignTag)

	if clusterIDIndex > -1 {
		aur.clusterIDSubSign = "$1"
	}

	if namespaceIndex > -1 {
		if clusterIDIndex > namespaceIndex {
			aur.namespaceSubSign = "$1"
			aur.clusterIDSubSign = "$2"
		} else if clusterIDIndex > -1 {
			aur.namespaceSubSign = "$2"
		} else {
			aur.namespaceSubSign = "$1"
		}
	}

	raw := strings.Replace(aur.Rule, ClusterIDSignTag, `([^/]+)`, 1)
	raw = strings.Replace(raw, NamespaceSignTag, `([^/]+)`, 1)
	aur.regex = regexp.MustCompile(raw)
}

// Match request match for module
func (aur *AuthURLRule) Match(clusterID, namespace, uri, method string) (match bool, action auth.Action, resource auth.Resource) {
	if aur.Method != method {
		return
	}

	if !aur.regex.MatchString(uri) {
		return
	}

	if aur.clusterIDSubSign != "" && clusterID == "" {
		clusterID = aur.regex.ReplaceAllString(uri, aur.clusterIDSubSign)
	}

	if aur.namespaceSubSign != "" {
		namespace = aur.regex.ReplaceAllString(uri, aur.namespaceSubSign)
	}

	resource = auth.Resource{
		ClusterID: clusterID,
		Namespace: namespace,
	}

	action = aur.Action

	return true, action, resource
}

var (
	// AuthRuleRegex regex for rule match
	AuthRuleRegex = regexp.MustCompile(`^/bcsapi/[^/]+/([^/]+)/([^/]+).*`)
	// StorageAuthRule rule for module bkbcs storage
	StorageAuthRule = []*AuthURLRule{
		// storage dynamic query
		{Rule: `/query/(?:mesos|k8s)/dynamic/clusters/{clusterId}/[^/]+`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/query/(?:mesos|k8s)/dynamic/clusters/{clusterId}/[^/]+`, Method: "POST", Action: auth.ActionRead},

		// storage dynamic query(old)
		{Rule: `/(?:mesos|k8s)/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/[^/]+/[^/]+`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/(?:mesos|k8s)/dynamic/namespace_resources/clusters/{clusterId}/namespaces/{namespace}/[^/]+`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/(?:mesos|k8s)/dynamic/cluster_resources/clusters/{clusterId}/[^/]+/[^/]+`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/(?:mesos|k8s)/dynamic/cluster_resources/clusters/{clusterId}/[^/]+`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/(?:mesos|k8s)/dynamic/all_resources/clusters/{clusterId}/[^/]+`, Method: "GET", Action: auth.ActionRead},

		// storage metric
		{Rule: `/metric/clusters/{clusterId}/namespaces/{namespace}/[^/]+/[^/]+`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/metric/clusters/{clusterId}/namespaces/{namespace}/[^/]+/[^/]+`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/metric/clusters/{clusterId}/namespaces/{namespace}/[^/]+/[^/]+`, Method: "DELETE", Action: auth.ActionManage},
		{Rule: `/metric/clusters/{clusterId}`, Method: "GET", Action: auth.ActionRead},
	}
	// MetricAuthRule rule for metric
	MetricAuthRule = []*AuthURLRule{
		// metric
		{Rule: `/metric/clustertype/[^/]+/clusters/{clusterId}/namespaces/{namespace}/metrics`, Method: "DELETE", Action: auth.ActionManage},
		{Rule: `/metric/tasks/clusters/{clusterId}`, Method: "GET", Action: auth.ActionRead},

		// metric task
		{Rule: `/metric/tasks/clusters/{clusterId}/namespaces/{namespace}/name/[^/]+`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/metric/tasks/clusters/{clusterId}/namespaces/{namespace}/name/[^/]+`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/metric/tasks/clusters/{clusterId}/namespaces/{namespace}/name/[^/]+`, Method: "DELETE", Action: auth.ActionManage},
	}
	// MesosAuthRule mesosdriver rule
	MesosAuthRule = []*AuthURLRule{
		// application and process
		{Rule: `/mesos/namespaces/{namespace}/(?:applications|processes)`, Method: "POST", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/(?:applications|processes)`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/(?:applications|processes)/[^/]+`, Method: "DELETE", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/(?:applications|processes)/rollback`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/(?:applications|processes)/[^/]+/scale/[^/]+`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/(?:applications|processes)`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/mesos/namespaces/{namespace}/(?:applications|processes)/[^/]+`, Method: "GET", Action: auth.ActionRead},

		// message
		{Rule: `/mesos/namespaces/{namespace}/applications/[^/]+/message`, Method: "POST", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/applications/[^/]+/taskgroups/[^/]+/message`, Method: "POST", Action: auth.ActionManage},

		// task
		{Rule: `/mesos/namespaces/{namespace}/applications/[^/]+/tasks`, Method: "GET", Action: auth.ActionRead},

		// taskgroup
		{Rule: `/mesos/namespaces/{namespace}/applications/[^/]+/taskgroups`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/mesos/namespaces/{namespace}/applications/[^/]+/taskgroups/[^/]+/rescheduler`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/applications/[^/]+/taskgroups/[^/]+/restart`, Method: "POST", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/applications/[^/]+/taskgroups/[^/]+/reload`, Method: "POST", Action: auth.ActionManage},

		// version
		{Rule: `/mesos/namespaces/{namespace}/applications/[^/]+/versions`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/mesos/namespaces/{namespace}/applications/[^/]+/versions/[^/]+`, Method: "GET", Action: auth.ActionRead},

		// configmaps secrets and services
		{Rule: `/mesos/namespaces/{namespace}/(?:configmaps|secrets|services)`, Method: "POST", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/(?:configmaps|secrets|services)`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/(?:configmaps|secrets|services)/[^/]+`, Method: "DELETE", Action: auth.ActionManage},

		// cluster
		{Rule: `/mesos/cluster/resources`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/mesos/cluster/endpoints`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/mesos/cluster/current/offers`, Method: "GET", Action: auth.ActionRead},

		// deployment
		{Rule: `/mesos/namespaces/{namespace}/deployments`, Method: "POST", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/deployments`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/deployments/[^/]+`, Method: "DELETE", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/deployments/[^/]+/cancelupdate`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/deployments/[^/]+/pauseupdate`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/deployments/[^/]+/resumeupdate`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/mesos/namespaces/{namespace}/deployments/[^/]+/scale/[^/]+`, Method: "PUT", Action: auth.ActionManage},

		// agent setting
		{Rule: `/mesos/agentsettings`, Method: "GET", Action: auth.ActionRead},
		{Rule: `/mesos/agentsettings`, Method: "DELETE", Action: auth.ActionManage},
		{Rule: `/mesos/agentsettings`, Method: "POST", Action: auth.ActionManage},
		{Rule: `/mesos/agentsettings/update`, Method: "POST", Action: auth.ActionManage},
		{Rule: `/mesos/agentsettings/enable`, Method: "POST", Action: auth.ActionManage},
		{Rule: `/mesos/agentsettings/disable`, Method: "POST", Action: auth.ActionManage},

		// custom resource
		{Rule: `/mesos/crr/register`, Method: "POST", Action: auth.ActionManage},
		{Rule: `/mesos/crd/namespaces/{namespace}/[^/]+`, Method: "POST", Action: auth.ActionManage},
		{Rule: `/mesos/crd/namespaces/{namespace}/[^/]+`, Method: "PUT", Action: auth.ActionManage},
		{Rule: `/mesos/crd/namespaces/{namespace}/[^/]+/[^/]+`, Method: "DELETE", Action: auth.ActionManage},

		// image
		{Rule: `/mesos/image/commit/[^/]+`, Method: "POST", Action: auth.ActionManage},

		// definition
		{Rule: `/mesos/definition/(?:application|deployment)/{namespace}/[^/]+`, Method: "GET", Action: auth.ActionRead},
	}
)

func init() {
	for _, rule := range StorageAuthRule {
		rule.init()
	}
	for _, rule := range MetricAuthRule {
		rule.init()
	}
	for _, rule := range MesosAuthRule {
		rule.init()
	}
}
