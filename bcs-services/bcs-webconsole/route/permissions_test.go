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

package route

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

const testClusterId = "BCS-K8S-40000"

func newTestAuthContext(appCode string) *AuthContext {
	return &AuthContext{
		Username:    "user1",
		ProjectId:   "p1",
		ProjectCode: "proj1",
		ClusterId:   testClusterId,
		BindBCS:     &UserClaimsInfo{BKAppCode: appCode},
		BindProject: &bcs.Project{ProjectId: "p1", Code: "proj1"},
		BindCluster: &bcs.Cluster{ProjectId: "p1", ClusterId: testClusterId},
	}
}

func newTestGinContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)
	return c
}

func TestValidateCredContainerNamespace(t *testing.T) {
	credName := "test-cred-container-namespace"
	err := config.G.ReadCred(credName, []byte(`
- credential_type: app_code
  credential: ns-app
  enabled: true
  scopes:
    - cluster_id: `+testClusterId+`
      namespace: EQ_ns-a
- credential_type: app_code
  credential: cluster-app
  enabled: true
  scopes:
    - cluster_id: `+testClusterId+`
`))
	assert.NoError(t, err)
	defer delete(config.G.Credentials, credName)

	tests := []struct {
		name      string
		appCode   string
		namespace string
		wantErr   bool
	}{
		{name: "namespace in scope", appCode: "ns-app", namespace: "ns-a", wantErr: false},
		{name: "container in other namespace", appCode: "ns-app", namespace: "kube-system", wantErr: true},
		{name: "empty namespace", appCode: "ns-app", namespace: "", wantErr: true},
		{name: "cluster scoped app", appCode: "cluster-app", namespace: "kube-system", wantErr: false},
		{name: "unknown app", appCode: "other-app", namespace: "ns-a", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredContainerNamespace(newTestAuthContext(tt.appCode), tt.namespace)
			assert.Equal(t, tt.wantErr, err != nil, "err: %v", err)
		})
	}
}

func TestValidateContainerNamespaceSkip(t *testing.T) {
	authCtx := newTestAuthContext("")

	err := ValidateContainerNamespace(newTestGinContext(), authCtx, "")
	assert.Error(t, err, "empty namespace must be rejected")

	config.G.Base.ManagerMap[authCtx.Username] = struct{}{}
	err = ValidateContainerNamespace(newTestGinContext(), authCtx, "kube-system")
	delete(config.G.Base.ManagerMap, authCtx.Username)
	assert.NoError(t, err, "manager skips permission check")

	config.G.Auth.Disable = true
	err = ValidateContainerNamespace(newTestGinContext(), authCtx, "kube-system")
	config.G.Auth.Disable = false
	assert.NoError(t, err, "auth disabled skips permission check")
}

func TestCheckProjectViewSkip(t *testing.T) {
	err := checkProjectView(newTestGinContext(), &AuthContext{ProjectId: "p1"})
	assert.Error(t, err, "request without username must be rejected")

	authCtx := &AuthContext{Username: "user1", ProjectId: "p1"}

	config.G.Base.ManagerMap[authCtx.Username] = struct{}{}
	err = checkProjectView(newTestGinContext(), authCtx)
	delete(config.G.Base.ManagerMap, authCtx.Username)
	assert.NoError(t, err, "manager skips permission check")

	config.G.Auth.Disable = true
	err = checkProjectView(newTestGinContext(), authCtx)
	config.G.Auth.Disable = false
	assert.NoError(t, err, "auth disabled skips permission check")
}
