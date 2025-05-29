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

// Package clustercheck xxx
package clustercheck

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	pluginmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"math/rand"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Plugin define cluster check plugin
type Plugin struct {
	opt            *Options
	testYamlString string
	pluginmanager.ClusterPlugin
}

// define plugin vars
var (
	clusterAvailabilityLabels                   = []string{"target", "bk_biz_id", "status"}
	clusterCheckDurationLabels                  = []string{"target", "bk_biz_id", "step"}
	clusterApiserverCertificateExpirationLabels = []string{"target", "bk_biz_id", "type"}
	clusterVersionLabels                        = []string{"target", "bk_biz_id", "version"}
	clusterAvailability                         = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterAvailabilityCheckMetricName,
		Help: ClusterAvailabilityCheckMetricName,
	}, clusterAvailabilityLabels)
	clusterVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterVersionMetricName,
		Help: ClusterVersionMetricName,
	}, clusterVersionLabels)
	clusterCheckDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterCheckDurationMeticName,
		Help: ClusterCheckDurationMeticName,
	}, clusterCheckDurationLabels)

	clusterApiserverCertificateExpiration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterApiserverCertExpirationMetricName,
		Help: ClusterApiserverCertExpirationMetricName,
	}, clusterApiserverCertificateExpirationLabels)
	unstructuredObj = &unstructured.Unstructured{}

	clusterCheckGaugeVecSetList         = make(map[string][]*metricmanager.GaugeVecSet)
	clusterCheckDurationGaugeVecSetList = make(map[string][]*metricmanager.GaugeVecSet)
	certificateExpirationGVSList        = make(map[string][]*metricmanager.GaugeVecSet)
	clusterVersionGaugeVecSetList       = make(map[string][]*metricmanager.GaugeVecSet)

	gvr schema.GroupVersionResource
)

func init() {
	// register plugin metric
	metricmanager.Register(clusterAvailability)
	metricmanager.Register(clusterCheckDuration)
	metricmanager.Register(clusterApiserverCertificateExpiration)
	metricmanager.Register(clusterVersion)
}

// Setup xxx
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("read clustercheck config file %s failed, err %s", configFilePath, err.Error())
	}
	p.opt = &Options{}
	if err = json.Unmarshal(configFileBytes, p.opt); err != nil {
		if err = yaml.Unmarshal(configFileBytes, p.opt); err != nil {
			return fmt.Errorf("decode clustercheck config file %s failed, err %s", configFilePath, err.Error())
		}
	}

	if err = p.opt.Validate(); err != nil {
		return err
	}

	p.Result = make(map[string]pluginmanager.CheckResult)
	p.ReadyMap = make(map[string]bool)

	decode := scheme.Codecs.UniversalDeserializer().Decode
	yamlData, _ := yaml.Marshal(p.opt.TestYaml)
	obj, gKV, _ := decode(yamlData, nil, nil)

	// 给测试workload添加标签
	switch gKV.Kind {
	case "Job":
		job := obj.(*batchv1.Job)
		job.Spec.Template.ObjectMeta.Labels["bcs-cluster-reporter"] = "bcs-cluster-reporter"
		job.ObjectMeta.Namespace = p.opt.Namespace
		objMap, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(job)
		unstructuredObj.SetUnstructuredContent(objMap)
		gvr.Group = "batch"
		gvr.Version = "v1"
		gvr.Resource = "jobs"
	default:
		klog.Fatalf("workload %s type is %s, not supported, please use job, deployment, replicaset",
			unstructuredObj.GetName(), gKV.Kind)
	}

	interval := p.opt.Interval

	// 运行模式
	if runMode == pluginmanager.RunModeDaemon {
		go func() {
			for {
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()
					if p.opt.Synchronization {
						pluginmanager.Pm.Lock()
					}
					go p.Check(pluginmanager.CheckOption{})
				} else {
					klog.Infof("the former clustercheck didn't over, skip in this loop")
				}
				select {
				case result := <-p.StopChan:
					klog.Infof("stop plugin %s by signal %d", p.Name(), result)
					return
				case <-time.After(time.Duration(interval) * time.Second):
					continue
				}
			}
		}()
	} else if runMode == pluginmanager.RunModeOnce {
		p.Check(pluginmanager.CheckOption{})
	}

	return nil
}

// Stop xxx
func (p *Plugin) Stop() error {
	p.StopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	return nil
}

// Name xxx
func (p *Plugin) Name() string {
	return "clustercheck"
}

func int64Ptr(i int64) *int64 { return &i }

// Check check for cluster apiserver cert, control panael availability and store result
func (p *Plugin) Check(option pluginmanager.CheckOption) {
	start := time.Now()
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		if p.opt.Synchronization {
			pluginmanager.Pm.UnLock()
		}
		p.CheckLock.Unlock()
		metricmanager.SetCommonDurationMetric([]string{"clustercheck", "", "", ""}, start)
	}()

	clusterConfigs := make(map[string]*pluginmanager.ClusterConfig)
	if option.ClusterIDs == nil || len(option.ClusterIDs) == 0 {
		clusterConfigs = pluginmanager.Pm.GetConfig().ClusterConfigs
		// clean deleted cluster metric data
	} else {
		for _, clusterID := range option.ClusterIDs {
			if cluster, ok := pluginmanager.Pm.GetConfig().ClusterConfigs[clusterID]; ok {
				clusterConfigs[clusterID] = cluster
			}
		}
	}

	wg := sync.WaitGroup{}

	// 遍历所有集群
	for _, cluster := range clusterConfigs {
		wg.Add(1)
		if len(option.ClusterIDs) > 0 {
			pluginmanager.Pm.AddCheck()
		} else {
			pluginmanager.Pm.Add()
		}

		go func(cluster *pluginmanager.ClusterConfig) {
			defer func() {
				wg.Done()
				if len(option.ClusterIDs) > 0 {
					pluginmanager.Pm.DoneCheck()
				} else {
					pluginmanager.Pm.Done()
				}
			}()

			clusterId := cluster.ClusterID
			clusterResult := pluginmanager.CheckResult{
				Items:        make([]pluginmanager.CheckItem, 0, 0),
				InfoItemList: make([]pluginmanager.InfoItem, 0, 0),
			}

			klog.Infof("start clustercheck for %s", clusterId)

			p.WriteLock.Lock()
			p.ReadyMap[cluster.ClusterID] = false
			p.WriteLock.Unlock()

			loopClusterChecktGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopClusterCheckDurationGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopCertificateExpirationGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopClusterVersionGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)

			defer func() {
				if r := recover(); r != nil {
					klog.Errorf("%s clustercheck failed: %s, stack: %v\n", clusterId, r, string(debug.Stack()))
					var responseContentType string
					body, _ := cluster.ClientSet.RESTClient().Get().
						AbsPath("/apis").
						SetHeader("Accept", "application/json").
						Do(context.TODO()).
						ContentType(&responseContentType).
						Raw()
					klog.Infof("Try get apis for %s: %s", clusterId, string(body))
				}
			}()

			// check apiserver cert
			getServerCertWG := sync.WaitGroup{}
			if len(cluster.Master) > 0 {
				getServerCertWG.Add(1)
				go func() {
					defer func() {
						getServerCertWG.Done()
					}()
					checkItemList, gvsList, err := getApiserverCert(cluster)
					if err != nil {
						klog.Errorf("%s check apiserver cert expiration failed: %s", cluster.ClusterID, err.Error())
					} else {
						clusterResult.Items = append(clusterResult.Items, checkItemList...)
						loopCertificateExpirationGVSList = append(loopCertificateExpirationGVSList, gvsList...)
					}
				}()
			}

			// blackbox check
			checkItemList, infoItemList, gvs, gvsList, err := testClusterByCreateUnstructuredObj(unstructuredObj, cluster)
			if err != nil {
				klog.Errorf("%s testClusterByCreateUnstructuredObj failed: %s", clusterId, err.Error())
			}

			clusterResult.Items = append(clusterResult.Items, checkItemList...)
			clusterResult.InfoItemList = append(clusterResult.InfoItemList, infoItemList...)
			loopClusterChecktGaugeVecSetList = append(loopClusterChecktGaugeVecSetList, gvs)
			loopClusterCheckDurationGaugeVecSetList = append(loopClusterCheckDurationGaugeVecSetList, gvsList...)
			loopClusterVersionGaugeVecSetList = append(loopClusterVersionGaugeVecSetList, &metricmanager.GaugeVecSet{
				Labels: []string{clusterId, cluster.BusinessID, cluster.Version},
				Value:  1,
			})

			klog.Infof("end clustercheck for %s", clusterId)

			p.WriteLock.Lock()

			// delete former metric
			if _, ok := clusterCheckGaugeVecSetList[cluster.ClusterID]; !ok {
				clusterCheckGaugeVecSetList[clusterId] = make([]*metricmanager.GaugeVecSet, 0, 0)
				clusterCheckDurationGaugeVecSetList[clusterId] = make([]*metricmanager.GaugeVecSet, 0, 0)
				certificateExpirationGVSList[clusterId] = make([]*metricmanager.GaugeVecSet, 0, 0)
				clusterVersionGaugeVecSetList[clusterId] = make([]*metricmanager.GaugeVecSet, 0, 0)
			} else {
				metricmanager.DeleteMetric(clusterAvailability, clusterCheckGaugeVecSetList[clusterId])
				metricmanager.DeleteMetric(clusterCheckDuration, clusterCheckDurationGaugeVecSetList[clusterId])
				metricmanager.DeleteMetric(clusterApiserverCertificateExpiration, certificateExpirationGVSList[clusterId])
				metricmanager.DeleteMetric(clusterVersion, clusterVersionGaugeVecSetList[clusterId])
			}

			// refresh new metric data
			for key, val := range clusterResult.Items {
				val.ItemName = StringMap[val.ItemName]
				val.ItemTarget = StringMap[val.ItemTarget]
				val.Status = StringMap[val.Status]
				clusterResult.Items[key] = val
			}
			p.Result[clusterId] = clusterResult

			clusterCheckGaugeVecSetList[clusterId] = loopClusterChecktGaugeVecSetList
			clusterCheckDurationGaugeVecSetList[clusterId] = loopClusterCheckDurationGaugeVecSetList
			certificateExpirationGVSList[clusterId] = loopCertificateExpirationGVSList
			clusterVersionGaugeVecSetList[clusterId] = loopClusterVersionGaugeVecSetList

			metricmanager.SetMetric(clusterAvailability, clusterCheckGaugeVecSetList[clusterId])
			metricmanager.SetMetric(clusterCheckDuration, clusterCheckDurationGaugeVecSetList[clusterId])
			metricmanager.SetMetric(clusterApiserverCertificateExpiration, certificateExpirationGVSList[clusterId])
			metricmanager.SetMetric(clusterVersion, clusterVersionGaugeVecSetList[clusterId])

			p.ReadyMap[clusterId] = true
			p.WriteLock.Unlock()
			getServerCertWG.Wait()
		}(cluster)
	}

	wg.Wait()

	// clean deleted cluster data
	for clusterID, _ := range p.ReadyMap {
		if _, ok := clusterConfigs[clusterID]; !ok {
			p.ReadyMap[clusterID] = false
			klog.Infof("delete cluster %s", clusterID)
		}
	}

	// 从readymap和指标中清理已删除集群
	for clusterID, _ := range p.ReadyMap {
		if _, ok := clusterConfigs[clusterID]; !ok {
			delete(p.ReadyMap, clusterID)
			metricmanager.DeleteMetric(clusterAvailability, clusterCheckGaugeVecSetList[clusterID])
			metricmanager.DeleteMetric(clusterCheckDuration, clusterCheckDurationGaugeVecSetList[clusterID])
			metricmanager.DeleteMetric(clusterApiserverCertificateExpiration, certificateExpirationGVSList[clusterID])
			metricmanager.DeleteMetric(clusterVersion, clusterVersionGaugeVecSetList[clusterID])
			delete(clusterCheckGaugeVecSetList, clusterID)
			delete(clusterCheckDurationGaugeVecSetList, clusterID)
			delete(certificateExpirationGVSList, clusterID)
			delete(clusterVersionGaugeVecSetList, clusterID)
			delete(p.Result, clusterID)
			klog.Infof("delete cluster %s", clusterID)
		}
	}

}

// getApiserverCert get apsierver cert expiration through api port
func getApiserverCert(clusterConfig *pluginmanager.ClusterConfig) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, error) {
	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	gvsList := make([]*metricmanager.GaugeVecSet, 0, 0)
	// 检查apiserver自签证书，可能无法在cluster-reporter直通
	index := rand.Intn(len(clusterConfig.Master))
	expiration, err := util.GetServerCert("apiserver-loopback-client", clusterConfig.Master[index], "60002")
	if err != nil {
		expiration, err = util.GetServerCert("apiserver-loopback-client", clusterConfig.Master[index], "6443")
		if err != nil {
			klog.Errorf("%s check apiserver self-signed cert expiration failed: %s", clusterConfig.ClusterID, err.Error())
			return checkItemList, gvsList, err
		}
	}

	// 证书检查结果
	checkItem := pluginmanager.CheckItem{
		ItemName:   ClusterApiserverCertExpirationCheckItem,
		ItemTarget: ApiserverTarget,
		Normal:     true,
		Status:     NormalStatus,
		Detail:     fmt.Sprintf(StringMap[AboutToExpireDetail], clusterConfig.ClusterID, expiration.Sub(time.Now())/time.Second),
		Level:      pluginmanager.WARNLevel,
		Tags:       nil,
	}

	// 时间在1周以内则返回异常
	if expiration.Sub(time.Now()) < 604800*time.Second {
		checkItem.Normal = false
		checkItem.Status = AboutToExpireStatus
		checkItem.SetDetail(fmt.Sprintf(StringMap[AboutToExpireDetail], clusterConfig.ClusterID, expiration.Sub(time.Now())/time.Second))
	}

	checkItemList = append(checkItemList, checkItem)

	gvsList = append(gvsList, &metricmanager.GaugeVecSet{
		Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, "self signed"},
		Value:  float64(expiration.Sub(time.Now()) / time.Second),
	})

	// 检查apiserver证书，可能无法在cluster-reporter直通
	expiration, err = util.GetServerCert(clusterConfig.Master[index], clusterConfig.Master[index], "60002")
	if err != nil {
		expiration, err = util.GetServerCert(clusterConfig.Master[index], clusterConfig.Master[index], "6443")
		if err != nil {
			klog.Errorf("%s check apiserver cert expiration failed: %s", clusterConfig.ClusterID, err.Error())
			return checkItemList, gvsList, err
		}
	}

	checkItem = pluginmanager.CheckItem{
		ItemName:   ClusterApiserverCertExpirationCheckItem,
		ItemTarget: ApiserverTarget,
		Normal:     true,
		Status:     NormalStatus,
		Detail:     fmt.Sprintf(StringMap[AboutToExpireDetail], clusterConfig.ClusterID, expiration.Sub(time.Now())/time.Second),
		Level:      pluginmanager.WARNLevel,
		Tags:       nil,
	}

	// 时间在1周以内则返回异常
	if expiration.Sub(time.Now()) < 604800*time.Second {
		checkItem.Normal = false
		checkItem.Status = AboutToExpireStatus
		checkItem.SetDetail(fmt.Sprintf(StringMap[AboutToExpireDetail], clusterConfig.ClusterID, expiration.Sub(time.Now())/time.Second))
	}

	checkItemList = append(checkItemList, checkItem)

	gvsList = append(gvsList, &metricmanager.GaugeVecSet{
		Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, "apiserver"},
		Value:  float64(expiration.Sub(time.Now()) / time.Second),
	})

	return checkItemList, gvsList, err
}

// testClusterByCreateUnstructuredObj test cluster by create a unstructuredObj workload and watch what will happen
func testClusterByCreateUnstructuredObj(unstructuredObj *unstructured.Unstructured, clusterConfig *pluginmanager.ClusterConfig,
) ([]pluginmanager.CheckItem, []pluginmanager.InfoItem, *metricmanager.GaugeVecSet, []*metricmanager.GaugeVecSet, error) {
	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	infoItemList := make([]pluginmanager.InfoItem, 0, 0)
	gvsList := make([]*metricmanager.GaugeVecSet, 0, 0)
	var gvs *metricmanager.GaugeVecSet

	var workloadToScheduleCost time.Duration
	var workloadToPodCost time.Duration
	var worloadToRunningCost time.Duration

	clusterUnstructuredObj := unstructuredObj.DeepCopy()
	// 随机workload名，避免重复导致的问题
	clusterUnstructuredObj.SetName("bcs-blackbox-job-" + time.Now().Format("2006150405"))
	var status string

	checkItem := pluginmanager.CheckItem{
		ItemName:   ClusterAvailabilityItem,
		ItemTarget: ApiserverTarget,
		Detail:     "",
		Level:      pluginmanager.WARNLevel,
		Tags:       nil,
	}

	// 获取k8s集群version,确认集群是否可访问
	version, err := k8s.GetK8sVersion(clusterConfig.ClientSet)
	if err != nil {
		// 如果失败则直接返回
		status = AvailabilityClusterFailStatus
		gvs = &metricmanager.GaugeVecSet{
			Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, status},
			Value:  1,
		}
		checkItem.Status = status
		checkItem.Normal = NormalStatus == status
		if !checkItem.Normal {
			checkItem.Detail = fmt.Sprintf(StringMap[ClusterAvailabilityDetail], clusterConfig.ClusterID, status)
		}
		err = fmt.Errorf("GetK8sVersion failed: %s", err.Error())
		return checkItemList, infoItemList, gvs, gvsList, err
	}

	// store version info
	infoItem := pluginmanager.InfoItem{
		ItemName: ClusterVersionItem,
		Result:   version,
	}
	clusterConfig.Version = version
	gvsList = append(gvsList, &metricmanager.GaugeVecSet{
		Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, version},
		Value:  1,
	})
	infoItemList = append(infoItemList, infoItem)

	// 获取dynamic resource interface
	dri, err := getResourceInterface(clusterConfig, unstructuredObj, &status)
	if err != nil {
		gvs = &metricmanager.GaugeVecSet{
			Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, status},
			Value:  1,
		}
		checkItem.Status = status
		checkItem.Normal = NormalStatus == status
		if !checkItem.Normal {
			checkItem.Detail = fmt.Sprintf(StringMap[ClusterAvailabilityDetail], clusterConfig.ClusterID, status)
		}
		return checkItemList, infoItemList, gvs, gvsList, err
	}

	defer func() {
		// 清理残留resource
		go func() {
			backgroundDeletion := metav1.DeletePropagationBackground

			// 获取所有的匹配job，避免历史残留
			jobList, listErr := clusterConfig.ClientSet.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{
				ResourceVersion: "0",
				LabelSelector:   "bcs-cluster-reporter=bcs-cluster-reporter",
			})
			if listErr != nil {
				klog.Errorf("%s get job failed %s", clusterConfig.ClusterID, listErr.Error())
			} else {
				// 避免过快删除导致异常事件
				time.Sleep(5 * time.Second)
				for _, job := range jobList.Items {
					klog.Infof("%s start to delete job %s", clusterConfig.ClusterID, job.Name)
					err = clusterConfig.ClientSet.BatchV1().Jobs(namespace).Delete(context.Background(), job.Name, metav1.DeleteOptions{
						GracePeriodSeconds: int64Ptr(5),
						PropagationPolicy:  &backgroundDeletion,
					})
					if err != nil {
						klog.Errorf("%s delete job failed %s", clusterConfig.ClusterID, err.Error())
					}
				}
			}

		}()
	}()

	// watch并判断创建clusterUnstructuredObj中发生的各种情况
	status, workloadToScheduleCost, workloadToPodCost, worloadToRunningCost, err =
		getWatchStatus(clusterConfig.ClientSet, clusterUnstructuredObj, dri, namespace, clusterConfig.ClusterID)

	infoItemList = append(infoItemList,
		pluginmanager.InfoItem{ItemName: worloadToRunningItem, Result: worloadToRunningCost},
		pluginmanager.InfoItem{ItemName: workloadToScheduleItem, Result: workloadToScheduleCost},
		pluginmanager.InfoItem{ItemName: workloadToPodItem, Result: workloadToPodCost})

	gvsList = append(gvsList, &metricmanager.GaugeVecSet{
		Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, workloadToPod},
		Value:  float64(workloadToPodCost) / float64(time.Second)}, &metricmanager.GaugeVecSet{
		Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, workloadToSchedule},
		Value:  float64(workloadToScheduleCost) / float64(time.Second),
	}, &metricmanager.GaugeVecSet{
		Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, worloadToRunning},
		Value:  float64(worloadToRunningCost) / float64(time.Second),
	})

	// 集群可用性检测结果单独一个指标
	gvs = &metricmanager.GaugeVecSet{
		Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, status},
		Value:  1,
	}
	checkItem.Status = status
	checkItem.Normal = NormalStatus == status
	if !checkItem.Normal {
		checkItem.Detail = fmt.Sprintf(StringMap[ClusterAvailabilityDetail], clusterConfig.ClusterID, status)
	}
	return checkItemList, infoItemList, gvs, gvsList, err
}

// getResourceInterface get dynamic resource interface
func getResourceInterface(clusterConfig *pluginmanager.ClusterConfig, clusterUnstructuredObj *unstructured.Unstructured, status *string) (dynamic.ResourceInterface, error) {
	ctx := util.GetCtx(10 * time.Second)

	_, err := clusterConfig.ClientSet.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		*status = AvailabilityNamespaceFailStatus
		_, createError := clusterConfig.ClientSet.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:            namespace,
				ResourceVersion: "0",
			},
		}, metav1.CreateOptions{})
		if createError != nil {
			klog.Errorf("%s create namespace failed: %s", clusterConfig.ClusterID, createError.Error())
		}
		err = fmt.Errorf("get target resource namespace failed: %s", err.Error())
		return nil, err
	}

	discoveryInterface := clusterConfig.ClientSet.Discovery().WithLegacy()
	if discoveryInterface == nil {
		*status = AvailabilityConfigFailStatus
		return nil, fmt.Errorf("Get discoveryInterface failed %s", err.Error())
	}

	dynamicConfig, err := dynamic.NewForConfig(clusterConfig.Config)
	if err != nil {
		*status = AvailabilityConfigFailStatus
		return nil, fmt.Errorf("%s Create dynamicConfig %s", clusterConfig.ClusterID, err.Error())
	}

	dri := dynamicConfig.Resource(gvr).Namespace(namespace)
	return dri, nil
}

// getWatchStatus get pod status of the workload, and return it.
func getWatchStatus(config kubernetes.Interface, clusterUnstructuredObj *unstructured.Unstructured,
	dri dynamic.ResourceInterface, namespace string, clusterID string) (status string,
	workloadToScheduleCost, workloadToPodCost, worloadToRunningCost time.Duration, err error) {
	startTime := time.Now()

	defer func() {
		klog.Infof("%s getWatchStatus duration %.2f s", clusterID, float64(time.Now().Sub(startTime)/time.Second))
	}()

	createPodFlag := false
	var createStartTime time.Time

	// 启动watch，观察对应任务的pod的所有事件
	factory := informers.NewSharedInformerFactoryWithOptions(config, 0, informers.WithNamespace(namespace))
	informer := factory.Core().V1().Pods().Informer()
	stopchan := make(chan struct{})
	closed := false
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			if strings.Contains(pod.Name, clusterUnstructuredObj.GetName()) {
				if !createPodFlag {
					klog.Infof("cluster %s create pod successful", clusterID)
					// pod创建成功时间
					workloadToPodCost = pod.CreationTimestamp.Sub(createStartTime)
					createPodFlag = true
				}
				if pod.Spec.NodeName != "" {
					if !closed {
						status = NormalStatus
						// pod调度成功耗时，调度成功直接返回
						klog.Infof("cluster %s schedule pod successful", clusterID)
						workloadToScheduleCost, worloadToRunningCost = getPodLifeCycleTimePoint(pod, createStartTime)

						closed = true
						close(stopchan)
					}
				} else {
					// 判断是否因为没有可供调度的node，是则返回没有node，区分未调度
					for _, condition := range pod.Status.Conditions {
						if strings.Contains(condition.Message, "nodes are available") {
							if !closed {
								status = AvailabilityNoNodeErrorStatus
								klog.Infof("%s scheduled pod failed: %s", clusterID, condition.Message)
								closed = true
								close(stopchan)
							}
						}
					}
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*v1.Pod)
			if strings.Contains(pod.Name, clusterUnstructuredObj.GetName()) {
				if pod.Spec.NodeName != "" {
					if !closed {
						status = NormalStatus
						// pod调度成功耗时，调度成功直接返回
						klog.Infof("cluster %s schedule pod successful", clusterID)
						workloadToScheduleCost, worloadToRunningCost = getPodLifeCycleTimePoint(pod, createStartTime)

						closed = true
						close(stopchan)
					}
				} else {
					// 判断是否因为没有可供调度的node，是则返回没有node，区分未调度
					for _, condition := range pod.Status.Conditions {
						if strings.Contains(condition.Message, "nodes are available") {
							if !closed {
								status = AvailabilityNoNodeErrorStatus
								klog.Infof("%s scheduled pod failed: %s", clusterID, condition.Message)
								closed = true
								close(stopchan)
							}
						}
					}
				}
			}

		},
		DeleteFunc: func(obj interface{}) {
		},
	})

	klog.Infof("%s start informer", clusterID)
	go informer.Run(stopchan)

	defer func() {
		// 关闭informer
		if !closed {
			closed = true
			close(stopchan)
		}
	}()

	// 记录发起创建workload的时间
	createStartTime = time.Now()
	// 创建workload
	ctx := util.GetCtx(90 * time.Second)
	klog.Infof("%s start create workload", clusterID)
	testObj, err := dri.Create(ctx, clusterUnstructuredObj, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("%s Create failed %s", clusterID, err.Error())
		if strings.Contains(err.Error(), "already exists") {
			time.Sleep(5 * time.Second)
			status = AvailabilityWorkloadExistStatus
		} else {
			status = AvailabilityCreateWorkloadErrorStatus
		}
		return
	}

	// 校验testObj创建时间，以检测apiserver时间是否有偏差
	createTS := testObj.GetCreationTimestamp()
	if createStartTime.Sub(createTS.Local()) > time.Second*5 || createStartTime.Sub(createTS.Local()) < 0-time.Second*5 {
		klog.Errorf("%s createtime %s, workload createtime %s", clusterID, createStartTime, createTS)
		status = AvailabilityTimeOffsetStatus
		return
	}

	klog.Infof("%s start wait for pod status", clusterID)
	// 创建workload后等待100s或观察到pod调度成功
	ctx = util.GetCtx(100 * time.Second)
	for {
		select {
		// 状态未调度或创建超时
		case <-ctx.Done():
			klog.Infof("%s context timeout", clusterID)
			if status != NormalStatus && status != AvailabilityNoNodeErrorStatus {
				if !createPodFlag {
					status = AvailabilityCreatePodTimeoutStatus
					klog.Errorf("%s create pod timeout", clusterID)
				} else {
					status = AvailabilitySchedulePodTimeoutStatus
					klog.Errorf("%s wait scheduled watch event timeout", clusterID)
				}
			}
			return

		case <-stopchan:
			klog.Infof("%s informer stopped", clusterID)
			if status != NormalStatus && status != AvailabilityNoNodeErrorStatus {
				if !createPodFlag {
					status = AvailabilityCreatePodTimeoutStatus
					klog.Errorf("%s create pod timeout", clusterID)
				} else {
					status = AvailabilitySchedulePodTimeoutStatus
					klog.Errorf("%s wait scheduled watch event timeout", clusterID)
				}
			}
			return
		}
	}
}

// getPodLifeCycleTimePoint get the time costed of every stag after the workload is created
func getPodLifeCycleTimePoint(pod *v1.Pod, createStartTime time.Time) (time.Duration, time.Duration) {
	var workloadToScheduleCost, worloadToRunningCost time.Duration
	for _, condition := range pod.Status.Conditions {
		// 获取pod调度的时间
		if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionTrue {
			if workloadToScheduleCost == 0 {
				workloadToScheduleCost = condition.LastTransitionTime.Sub(createStartTime)
			}

		} else if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			if worloadToRunningCost == 0 {
				worloadToRunningCost = condition.LastTransitionTime.Sub(createStartTime)
			}
		}
	}

	return workloadToScheduleCost, worloadToRunningCost
}

// Ready return true if cluster check is over
func (p *Plugin) Ready(clusterID string) bool {
	p.WriteLock.Lock()
	defer p.WriteLock.Unlock()
	return p.ReadyMap[clusterID]
}

// GetResult return check result by cluster ID
func (p *Plugin) GetResult(clusterID string) pluginmanager.CheckResult {
	return p.Result[clusterID]
}
