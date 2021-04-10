package configuration

import (
	"context"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
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

//var ClusterInfo AggregationClusterInfo

func (aci *AggregationClusterInfo) SetClusterInfo(acm *AggregationConfigMapInfo) {

	if acm.GetClusterOverride() != "" {
		klog.Infoln("Get memberClusterList from AggregationConfigMapInfo of ClusterOverride.")
		aci.memberClusterList = strings.ToUpper(acm.GetClusterOverride())
	} else {
		klog.Infoln("Get memberClusterList from kubeFederated member cluster.")

		config, err := rest.InClusterConfig()
		if err != nil {
			// fallback to kubeconfig
			kubeconfig := filepath.Join("~", ".kube", "config")
			if envvar := os.Getenv("KUBECONFIG"); len(envvar) > 0 {
				kubeconfig = envvar
			}
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				klog.Errorf("The kubeconfig cannot be loaded: %v\n", err)
				os.Exit(1)
			}
		}

		clientSet, err := generic.New(config)
		if err != nil {
			klog.Errorf("Failed to create clientset: %v", err)
			os.Exit(1)
		}

		clusterList := &fedv1b1.KubeFedClusterList{}

		for {
			aci.memberClusterList = ""

			err = clientSet.List(context.TODO(), clusterList, "kube-federation-system")
			if err != nil {
				klog.Warningf("Error retrieving list of federated clusters: %v\n", err)
			} else {
				if len(clusterList.Items) == 0 {
					klog.Errorln("No federated clusters found, wait for join KubeFed member cluster")
				} else {
					for _, cluster := range clusterList.Items {
						var clusterTmp string
						if acm.GetClusterIgnorePrefix() != "" {
							clusterTmp = strings.TrimPrefix(cluster.Name,
								acm.GetClusterIgnorePrefix())
						} else {
							clusterTmp = cluster.Name
						}
						aci.memberClusterList += strings.ToUpper(clusterTmp) + ","
					}
					aci.memberClusterList = strings.TrimRight(aci.memberClusterList, ",")
					break
				}
			}

			klog.Errorln("Can not get the member cluster list from kubeFederated member cluster, " +
				"wait for 30 seconds for next loop")
			time.Sleep(30 * time.Second)
		}
	}
	klog.Infoln("Get memberClusterList: " + aci.memberClusterList)
}

func (aci *AggregationClusterInfo) GetClusterList() string {
	return aci.memberClusterList
}