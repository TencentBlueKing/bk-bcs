package app

import (
	"context"
	"fmt"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"time"
)

type AggregationInfo struct {
	bcsStorageUrlBase               string
	bcsStorageTokenEnable           string
	bcsStorageToken                 string
	memberClusterOverrideEnable     string
	memberClusterOverride           string
	memberClusterIgnorePrefixEnable string
	memberClusterIgnorePrefix       string
}

var Aggregation AggregationInfo

func GetAggregationInfo() {
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

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to create clientset: %v\n", err)
		os.Exit(1)
	}

	for {
		Aggregation.bcsStorageUrlBase = ""
		BcsStorageAddressConfig, err := clientset.CoreV1().ConfigMaps("bcs-system").Get(context.TODO(),
			"bcs-federated-apiserver",
			metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				fmt.Printf("failed to query configmap: %v\n", err)
			} else if kerrors.IsUnauthorized(err) {
				fmt.Printf("Unauthorized to query configmap: %v\n", err)
			} else {
				fmt.Printf("failed to query kubeadm's configmap: %v\n", err)
			}

			fmt.Println("can not get the bcs-strorage configmap, wait for 30 seconds for next loop")
			time.Sleep(30 * time.Second)
		} else {
			var address string
			var podURI string

			for k, data := range BcsStorageAddressConfig.Data {
				switch k {
				case "bcs-storage-address":
					address = data
				case "bcs-storage-pod-uri":
					podURI = data
				case "bcs-storage-token-enable":
					Aggregation.bcsStorageTokenEnable = data
				case "bcs-storage-token":
					Aggregation.bcsStorageToken = data
				case "member-cluster-override-enable":
					Aggregation.memberClusterOverrideEnable = data
				case "member-cluster-override":
					Aggregation.memberClusterOverride = data
				case "member-cluster-ignore-prefix-enable":
					Aggregation.memberClusterIgnorePrefixEnable = data
				case "member-cluster-ignore-prefix":
					Aggregation.memberClusterIgnorePrefix = data
				default:
					fmt.Println("no need to parse it: ", k, data)
				}
			}
			fmt.Printf("Aggregation: %+v\n", Aggregation)

			if address == "" || podURI == "" {
				fmt.Println("bcs-storage address or podURI is null, please check your configmap")
				continue
			}
			Aggregation.bcsStorageUrlBase = address + podURI
			fmt.Println("bcs-storage base uri: " + Aggregation.bcsStorageUrlBase)
			break
		}
	}
}

func GetBcsStorageUrlBase() string {
	return Aggregation.bcsStorageUrlBase
}

func GetBcsStorageTokenEable() string {
	return Aggregation.bcsStorageTokenEnable
}

func GetBcsStorageToken() string {
	return Aggregation.bcsStorageToken
}

func GetClusterOverrideEnable() string {
	return Aggregation.memberClusterOverrideEnable
}

func GetClusterOverride() string {
	return Aggregation.memberClusterOverride
}

func GetClusterIgnorePrefixEnable() string {
	return Aggregation.memberClusterIgnorePrefixEnable
}

func GetClusterIgnorePrefix() string {
	return Aggregation.memberClusterIgnorePrefix
}
