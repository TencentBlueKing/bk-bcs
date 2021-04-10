package main

import (
	kubectlAgg "github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/kubectl-agg"
	"k8s.io/klog"
	"os"
)

func main() {
	var o kubectlAgg.AggPodOptions

	err := kubectlAgg.ParseKubectlArgs(os.Args, &o)
	if err != nil {
		klog.Errorln("ParseKubectlArgs error.")
		return
	}

	clientSet, err := kubectlAgg.NewClientSet()
	if err != nil {
		klog.Errorln("new clientSet error.")
		return
	}

	pods, err := kubectlAgg.GetPodAggregationList(clientSet, &o)
	if err != nil {
		klog.Errorln("GetPodAggregationList error.")
		return
	}

	kubectlAgg.PrintPodAggregation(&o, pods)
}
