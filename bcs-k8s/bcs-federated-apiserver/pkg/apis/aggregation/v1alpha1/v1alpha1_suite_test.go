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

package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/apiserver-builder-alpha/pkg/test"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/client/clientset_generated/clientset"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/openapi"
)

var testenv *test.TestEnvironment
var config *rest.Config
var cs *clientset.Clientset

func TestV1alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "v1 Suite", []Reporter{test.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	testenv = test.NewTestEnvironment(apis.GetAllApiBuilders(), openapi.GetOpenAPIDefinitions)
	config = testenv.Start()
	cs = clientset.NewForConfigOrDie(config)
})

var _ = AfterSuite(func() {
	testenv.Stop()
})
