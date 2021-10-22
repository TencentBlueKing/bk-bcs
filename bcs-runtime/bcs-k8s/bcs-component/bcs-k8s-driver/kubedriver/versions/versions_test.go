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

package versions

import (
	"fmt"
	"os"
	"testing"
)

func TestSplitUrlByVersion(t *testing.T) {
	a, b, err := SplitUrlByVersion("/apis/apps/v1beta1/namespaces/{namespace}/controllerrevisions")
	if err != nil {
		t.Error("xxxxxx")
	}
	if a != "/apis/apps/v1beta1/" {
		t.Error("former string: ", a)
	}
	if b != "namespaces/{namespace}/controllerrevisions" {
		t.Error("later string: ", b)
	}
}

func TestFormatURI(t *testing.T) {
	formatString := FormatURI("namespaces/xxxxxxx/secrets/xxxxx")
	if formatString != "namespaces/{namespace}/secrets/{name}" {
		t.Error("string ", formatString)
	}
	formatString = FormatURI("namespaces/asdfasdfa/services/dfdfd/proxy/eeeee")
	if formatString != "namespaces/{namespace}/services/{name}/proxy/{path}" {
		t.Error("string ", formatString)
	}
	formatString = FormatURI("namespaces/asdfasdfa/services/dfdfd/proxy")
	if formatString != "namespaces/{namespace}/services/{name}/proxy" {
		t.Error("string ", formatString)
	}
	formatString = FormatURI("nodes/asdfasd/proxy/idididi")
	if formatString != "nodes/{name}/proxy/{path}" {
		t.Error("string", formatString)
	}
	formatString = FormatURI("namespaces/xxxxx")
	if formatString != "namespaces/{namespace}" {
		t.Error("string", formatString)
	}
}

func TestGetClientSet(t *testing.T) {
	os.Chdir("../../cmd/")
	apiPrefer := map[string]string{
		"apps":                      "v1beta1",
		"authentication.k8s.io":     "v1beta1",
		"authorization.k8s.io":      "v1beta1",
		"autoscaling":               "v1",
		"batch":                     "v1",
		"certificates.k8s.io":       "v1alpha1",
		"extensions":                "v1beta1",
		"policy":                    "v1beta1",
		"rbac.authorization.k8s.io": "v1alpha1",
		"storage.k8s.io":            "v1beta1",
	}
	cs := ClientSetter{}

	err := cs.GetClientSetUrl("namespaces/asdfasdfa/services/dfdfd/proxy/eeeee", "1.5", apiPrefer)
	if err != nil {
		t.Error(fmt.Sprintf("target: %s, error: %s", cs.ClientSet, err))
	}
	if cs.ClientSet != "/api/v1/" {
		t.Error(fmt.Sprintf("not match expected: %s", cs.ClientSet))
	}

	err = cs.GetClientSetUrl("namespaces", "1.7", apiPrefer)
	if err != nil {
		t.Error(fmt.Sprintf("target: %s, error: %s", cs.ClientSet, err))
	}
	if cs.ClientSet != "/api/v1/" {
		t.Error(fmt.Sprintf("not match expected: %s", cs.ClientSet))
	}

	err = cs.GetClientSetUrl("namespaces/xxxx", "1.7", apiPrefer)
	if err != nil {
		t.Error(fmt.Sprintf("target: %s, error: %s", cs.ClientSet, err))
	}
	if cs.ClientSet != "/api/v1/" {
		t.Error(fmt.Sprintf("not match expected: %s", cs.ClientSet))
	}

	err = cs.GetClientSetUrl("jobs", "1.5", apiPrefer)
	if err != nil {
		t.Error(fmt.Sprintf("target: %s, error: %s", cs.ClientSet, err))
	}
	if cs.ClientSet != "/apis/batch/v1/" {
		t.Error(fmt.Sprintf("not match expected: %s", cs.ClientSet))
	}

	err = cs.GetClientSetUrl("deployments", "1.5", apiPrefer)
	if err != nil {
		t.Error(fmt.Sprintf("target: %s, error: %s", cs.ClientSet, err))
	}
	if cs.ClientSet != "/apis/extensions/v1beta1/" {
		t.Error(fmt.Sprintf("not match expected: %s", cs.ClientSet))
	}

	err = cs.GetClientSetUrl("endpoints", "1.5", apiPrefer)
	if err != nil {
		t.Error(fmt.Sprintf("target: %s, error: %s", cs.ClientSet, err))
	}
	if cs.ClientSet != "/api/v1/" {
		t.Error(fmt.Sprintf("not match expected: %s", cs.ClientSet))
	}

}

func TestAddVersionIntoBody(t *testing.T) {
	type TestJson struct {
		ID   int    `json:"id"`
		Body string `json:"body"`
	}
	testJson := TestJson{
		ID:   1,
		Body: "test",
	}
	b, _ := json.Marshal(testJson)
	cs := ClientSetter{
		BodyContent: &b,
		ClientSet:   "hahahaha",
	}
	cs.AddVersionIntoBody()
	if json.Get(*cs.BodyContent, "body").ToString() != "test" {
		t.Error("original struct has been changed")
	}

	versionPrefix := json.Get(*cs.BodyContent, "apiVersion").ToString()
	if versionPrefix != "hahahaha" {
		t.Error("fetch version failed: ", versionPrefix)
	}

}
