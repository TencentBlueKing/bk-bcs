package app

import (
	"context"
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	"sigs.k8s.io/kubefed/pkg/client/generic"
	"strings"
	"time"
)

type AggregationClusterInfo struct {
	memberClusterList string
}

var ClusterInfo AggregationClusterInfo

func GetClusterInfo() {
	config, err := rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig
		kubeconfig := filepath.Join("~", ".kube", "config")
		if envvar := os.Getenv("KUBECONFIG"); len(envvar) > 0 {
			kubeconfig = envvar
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Printf("The kubeconfig cannot be loaded: %v\n", err)
			os.Exit(1)
		}
	}

	clientset, err := generic.New(config)
	if err != nil {
		fmt.Printf("Failed to create clientset: %v", err)
		os.Exit(1)
	}

	clusterList := &fedv1b1.KubeFedClusterList{}

	for {
		ClusterInfo.memberClusterList = ""

		err = clientset.List(context.TODO(), clusterList, "kube-federation-system")
		if err != nil {
			fmt.Printf("Error retrieving list of federated clusters: %v\n", err)
		}
		if len(clusterList.Items) == 0 {
			fmt.Println("No federated clusters found")
		} else {
			for _, cluster := range clusterList.Items {
				var clusterTmp string
				if GetClusterIgnorePrefixEnable() == "true" && GetClusterIgnorePrefix() != "" {
					clusterTmp = strings.TrimPrefix(cluster.Name,
						GetClusterIgnorePrefix())
				} else {
					clusterTmp = cluster.Name
				}
				ClusterInfo.memberClusterList += strings.ToUpper(clusterTmp) + ","
			}
			ClusterInfo.memberClusterList = strings.TrimRight(ClusterInfo.memberClusterList, ",")

			if GetClusterOverrideEnable() == "true" && GetClusterOverride() != "" {
				ClusterInfo.memberClusterList = strings.ToUpper(GetClusterOverride())
			}

			fmt.Println("member cluster list: " + ClusterInfo.memberClusterList)
			break
		}

		fmt.Println("can not get the member cluster list, wait for 30 seconds for next loop")
		time.Sleep(30 * time.Second)
	}
}

func GetClusterList() string {
	return ClusterInfo.memberClusterList
}
