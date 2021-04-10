package kubectl_agg

import (
	"context"
	"flag"
	"fmt"
	aggregation_api_v1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/aggregation/v1alpha1"
	clientset "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"math"
	"os"
	"path/filepath"
	"time"
)

type AggPodOptions struct {
	ResourceName  string
	Namespace     string
	Selector      string
	AllNamespaces bool
	WideMessage   string
	HelpMessage   bool
}

func GetPodRestartCount(pod aggregation_api_v1alpha1.PodAggregation) int32 {
	var restartCount int32 = 0
	for _, v := range pod.Status.ContainerStatuses {
		restartCount = restartCount + v.RestartCount
	}
	return restartCount
}

func GetContainerReadyStatus(pod aggregation_api_v1alpha1.PodAggregation) string {
	var containerCount int32 = 0
	var containerReadyCount int32 = 0

	for _, v := range pod.Status.ContainerStatuses {
		containerCount++
		if v.Ready {
			containerReadyCount++
		}
	}
	return fmt.Sprintf("%d/%d", containerReadyCount, containerCount)
}

func GetPodAge(pod aggregation_api_v1alpha1.PodAggregation) string {
	var createAgeHour float64

	createAgeHour = time.Since(pod.CreationTimestamp.Time).Hours()
	if createAgeHour < 24.0 {
		return fmt.Sprintf("%dh", int64(math.Ceil(createAgeHour)))
	} else {
		createAgeDay := createAgeHour / 24.0
		if createAgeDay < 365.0 {
			return fmt.Sprintf("%dd", int64(math.Ceil(createAgeDay)))
		} else {
			createAgeYear := createAgeDay / 365.0
			return fmt.Sprintf("%dy", int64(math.Ceil(createAgeYear)))
		}
	}
}

func GetReadinessGateStatus(pod aggregation_api_v1alpha1.PodAggregation) string {
	var readinessGateCount int32
	var readinessGateReadyCount int32
	for _, v := range pod.Status.Conditions {
		if v.Type != "Initialized" && v.Type != "Ready" && v.Type != "ContainersReady" && v.Type != "PodScheduled" {
			readinessGateCount++
			if v.Status == "true" {
				readinessGateReadyCount++
			}
		}
	}

	if readinessGateCount == 0 {
		return "<none>"
	} else {
		return fmt.Sprintf("%d/%d", readinessGateReadyCount, readinessGateCount)
	}
}

func GetNominatedNode(pod aggregation_api_v1alpha1.PodAggregation) string {
	if pod.Status.NominatedNodeName == "" {
		return "<none>"
	} else {
		return pod.Status.NominatedNodeName
	}
}

func Usage() {
	klog.Infoln("Usage:\n  kubectl agg pod\n[(-o|--output=)wide\n [NAME | -l label] [flags] [options]")
}

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

	switch os.Args[1] {
	case "pod":
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

func NewClientSet() (clientSet *clientset.Clientset, err error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		var kubeconfig string
		if envHome := os.Getenv("HOME"); len(envHome) > 0 {
			kubeconfig = filepath.Join(envHome, ".kube", "config")
			if envVar := os.Getenv("KUBECONFIG"); len(envVar) > 0 {
				kubeconfig = envVar
			}
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Errorf("The kubeconfig cannot be loaded: %v\n", err)
			return nil, err
		}
	}

	clientSet, err = clientset.NewForConfig(config)
	if err != nil {
		klog.Errorln("Failed to create clientset")
		return nil, err
	}
	return clientSet, nil
}

func GetPodAggregationList(clientSet *clientset.Clientset, o *AggPodOptions) (pods *aggregation_api_v1alpha1.PodAggregationList, err error) {
	if o.ResourceName != "" {
		pods, err = clientSet.AggregationV1alpha1().PodAggregations(o.Namespace).Get(context.TODO(),
			o.ResourceName, v1.GetOptions{})
		if err != nil {
			klog.Errorf("Error: failed to get pod: %s/%s: %s\n", o.Namespace, o.ResourceName, err)
			return &aggregation_api_v1alpha1.PodAggregationList{}, nil
		}
	} else {
		if o.AllNamespaces {
			o.Namespace = ""
		}

		selector := labels.Everything()
		if len(o.Selector) > 0 {
			selector, err = labels.Parse(o.Selector)
			if err != nil {
				return &aggregation_api_v1alpha1.PodAggregationList{}, nil
			}
		}

		pods, err = clientSet.AggregationV1alpha1().PodAggregations(o.Namespace).List(context.TODO(),
			v1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			klog.Errorf("Error: failed to list pods: %s\n", err)
			return &aggregation_api_v1alpha1.PodAggregationList{}, nil
		}
	}
	return pods, nil
}

func PrintPodAggregation(o *AggPodOptions, pods *aggregation_api_v1alpha1.PodAggregationList) {
	if len(pods.Items) == 0 {
		fmt.Println("No resources found")
		return
	}
	if o.WideMessage == "wide" {
		fmt.Printf("%-16s%-64s%-8s%-16s%-10s%-10s%-20s%-40s%-20s%-16s\n", "NAMESPACE", "NAME", "READY",
			"STATUS",
			"RESTARTS", "AGE", "IP", "NODE", "NOMINATED NODE", "READINESS GATES")
	} else {
		fmt.Printf("%-16s%-64s%-8s%-16s%-10s%-8s\n", "NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS",
			"AGE")
	}

	for _, v := range pods.Items {
		if o.WideMessage == "wide" {
			//readinessGateStatus := kubectl_agg.GetReadinessGateStatus(v)
			fmt.Printf("%-16s%-64s%-8s%-16s%-10d%-10s%-20s%-40s%-20s%-16s\n", v.Namespace, v.Name,
				GetContainerReadyStatus(v), string(v.Status.Phase),
				GetPodRestartCount(v),
				GetPodAge(v), v.Status.PodIP, v.Spec.NodeName,
				GetNominatedNode(v),
				GetReadinessGateStatus(v))
		} else {
			fmt.Printf("%-16s%-64s%-8s%-16s%-10d%-8s\n", v.Namespace, v.Name,
				GetContainerReadyStatus(v), string(v.Status.Phase),
				GetPodRestartCount(v),
				GetPodAge(v))
		}
	}
}