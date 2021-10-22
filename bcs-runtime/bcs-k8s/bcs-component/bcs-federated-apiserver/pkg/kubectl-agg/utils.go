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

// Package kubectl_agg offer some functions used by kubectl-agg command line.
package kubectl_agg

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/client/clientset_generated/clientset"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

// KubeConfig is set by the order of "inCluster" "~/.kube/config" and the env "KUBECONFIG"
func newConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		var kubeConfig string

		// fallback to kubeConfig
		if envHome := os.Getenv("HOME"); len(envHome) > 0 {
			kubeConfig = filepath.Join(envHome, ".kube", "config")
		}
		if envVar := os.Getenv("KUBECONFIG"); len(envVar) > 0 {
			kubeConfig = envVar
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			klog.Errorf("The kubeConfig cannot be loaded: %v\n", err)
			return nil, err
		}
	}
	return config, err
}

// NewFedApiServerClientSet to generate a new NewFedApiServerClientSet
func NewFedApiServerClientSet() (*clientset.Clientset, error) {
	config, err := newConfig()
	if err != nil {
		klog.Errorf("The kubeConfig cannot be loaded: %v\n", err)
		return nil, err
	}

	clientSet, err := clientset.NewForConfig(config)
	if err != nil {
		klog.Errorf("Failed to create clientSet: %v\n", err)
		return nil, err
	}
	return clientSet, err
}

// Usage function print the 'kubectl agg pod' command-line tools basic usage.
func Usage() {
	klog.Infoln("Usage:\n  kubectl agg pod\n[(-o|--output=)wide\n [NAME | -l label] [flags] [options]")
}

// ParseKubectlArgs function parse the kubectl-agg commandline,
// which support the static "kubectl agg pod" prefix commands.
func ParseKubectlArgs(args []string, o *AggPodOptions) (err error) {
	flag.Usage = Usage

	if len(args) < 2 {
		klog.Errorln("expected 'pod' subcommands")
		return err
	}

	podCmd := flag.NewFlagSet("pod", flag.ExitOnError)

	podCmd.BoolVar(&o.AllNamespaces, "all-namespaces", o.AllNamespaces, "--all-namespaces=false: If present, list the requested object(s) across all namespaces. Namespace in current\ncontext is ignored even if specified with --namespace.")
	podCmd.BoolVar(&o.AllNamespaces, "A", o.AllNamespaces, "  -A, If present, list the requested object(s) across all namespaces. Namespace in current\ncontext is ignored even if specified with --namespace.")
	podCmd.StringVar(&o.Namespace, "n", "default", "Namespace")
	podCmd.StringVar(&o.Namespace, "namespace", "default", "Namespace")
	podCmd.StringVar(&o.WideMessage, "o", o.WideMessage, "Output format. One of: wide")
	podCmd.StringVar(&o.Selector, "l", o.Selector, " Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	podCmd.BoolVar(&o.LabelsMessage, "show-labels", o.LabelsMessage, "--show-labels=false: If present, "+
		"list the requested object(s) show it label messages.")

	switch os.Args[1] {
	case "pod", "po":
		podCmd.Parse(os.Args[2:])
	default:
		klog.Infoln("Usage:\n  kubectl agg pod [(-o|--output=)wide [NAME | -l label] [flags] [options]")
		return err
	}

	o.ResourceName = podCmd.Arg(0)

	if podCmd.Arg(0) != "" {
		podCmd.Parse(podCmd.Args()[1:])
	}
	return nil
}
