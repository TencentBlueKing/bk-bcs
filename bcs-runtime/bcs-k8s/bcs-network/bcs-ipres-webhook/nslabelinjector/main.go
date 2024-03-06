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

package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	nsListFlag  string
	injectKey   string
	injectValue string
)

func main() {
	flag.StringVar(&nsListFlag, "namespace_list", "kube-system,bcs-system", "namespace list for webhook to ignore")
	flag.StringVar(&injectKey, "inject_key", "ignore-bcs-ipres-webhook", "key to be injected into namespace label")
	flag.StringVar(&injectValue, "inject_value", "true", "value to be injected into namespace label")
	flag.Parse()

	restConfig, err1 := rest.InClusterConfig()
	if err1 != nil {
		blog.Fatalf("get incluster config failed, err %s", err1.Error())
	}
	k8sCli, err2 := kubernetes.NewForConfig(restConfig)
	if err2 != nil {
		blog.Fatalf("create k8s client set failed, err %s", err2.Error())
	}
	hasErr := false
	nsList := strings.Split(nsListFlag, ",")
	for _, ns := range nsList {
		nsObj, err := k8sCli.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
		if err != nil {
			blog.Errorf("get ns %s failed, err %s", ns, err.Error())
			hasErr = true
			continue
		}
		nsObj.Labels[injectKey] = injectValue
		if _, err = k8sCli.CoreV1().Namespaces().Update(
			context.Background(), nsObj, metav1.UpdateOptions{}); err != nil {
			blog.Errorf("update ns %v failed, err %s", nsObj, err.Error())
			hasErr = true
			continue
		}
	}
	if hasErr {
		os.Exit(1)
	}
	blog.Infof("inject label successfully")
}
