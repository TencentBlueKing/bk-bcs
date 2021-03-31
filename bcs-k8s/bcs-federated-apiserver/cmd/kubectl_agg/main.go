package main

import (
	"fmt"
	kubectl_agg "github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/kubectl_agg"
	"os"
)

func main() {
	var o kubectl_agg.AggPodOptions

	err := kubectl_agg.ParseKubectlArgs(os.Args, &o)
	if err != nil {
		fmt.Println("ParseKubectlArgs error.")
		return
	}

	clientSet, err := kubectl_agg.NewClientSet()
	if err != nil {
		fmt.Println("new clientSet error.")
		return
	}

	pods, err := kubectl_agg.GetPodAggregationList(clientSet, &o)
	if err != nil {
		fmt.Println("GetPodAggregationList error.")
		return
	}

	kubectl_agg.PrintPodAggregation(&o, pods)
}
