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
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/metricmanager"
	pluginmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/pluginmanager"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

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
	opt            *Options            // 插件配置选项
	testYamlString string              // 测试YAML字符串
	pluginmanager.ClusterPlugin        // 嵌入集群插件接口
}

// define plugin vars
var (
	// clusterAvailabilityLabels 集群可用性指标标签
	clusterAvailabilityLabels = []string{"target", "bk_biz_id", "status"}
	// clusterWebhookCrossLabels Webhook跨命名空间指标标签
	clusterWebhookCrossLabels = []string{"target", "bk_biz_id", "webhook"}
	// clusterCheckDurationLabels 集群检查耗时指标标签
	clusterCheckDurationLabels = []string{"target", "bk_biz_id", "step"}
	// clusterApiserverCertificateExpirationLabels APIServer证书过期指标标签
	clusterApiserverCertificateExpirationLabels = []string{"target", "bk_biz_id", "type"}
	// clusterVersionLabels 集群版本指标标签
	clusterVersionLabels = []string{"target", "bk_biz_id", "version"}
	
	// clusterAvailability 集群可用性指标
	clusterAvailability = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterAvailabilityCheckMetricName,
		Help: ClusterAvailabilityCheckMetricName,
	}, clusterAvailabilityLabels)
	
	// clusterWebhookCross Webhook跨命名空间检查指标
	clusterWebhookCross = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterWebhookCrossMetricName,
		Help: ClusterWebhookCrossMetricName,
	}, clusterAvailabilityLabels)
	
	// clusterVersion 集群版本信息指标
	clusterVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterVersionMetricName,
		Help: ClusterVersionMetricName,
	}, clusterVersionLabels)
	
	// clusterCheckDuration 集群检查各阶段耗时指标
	clusterCheckDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterCheckDurationMeticName,
		Help: ClusterCheckDurationMeticName,
	}, clusterCheckDurationLabels)

	// clusterApiserverCertificateExpiration APIServer证书过期时间指标
	clusterApiserverCertificateExpiration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: ClusterApiserverCertExpirationMetricName,
		Help: ClusterApiserverCertExpirationMetricName,
	}, clusterApiserverCertificateExpirationLabels)
	
	// unstructuredObj 非结构化对象，用于创建测试工作负载
	unstructuredObj = &unstructured.Unstructured{}

	// clusterCheckGaugeVecSetList 存储集群检查结果指标集合
	clusterCheckGaugeVecSetList = make(map[string][]*metricmanager.GaugeVecSet)
	// clusterWebhookGaugeVecSetList 存储Webhook检查结果指标集合
	clusterWebhookGaugeVecSetList = make(map[string][]*metricmanager.GaugeVecSet)
	// clusterCheckDurationGaugeVecSetList 存储检查耗时指标集合
	clusterCheckDurationGaugeVecSetList = make(map[string][]*metricmanager.GaugeVecSet)
	// certificateExpirationGVSList 存储证书过期指标集合
	certificateExpirationGVSList = make(map[string][]*metricmanager.GaugeVecSet)
	// clusterVersionGaugeVecSetList 存储集群版本指标集合
	clusterVersionGaugeVecSetList = make(map[string][]*metricmanager.GaugeVecSet)

	// gvr GroupVersionResource，用于动态客户端资源操作
	gvr schema.GroupVersionResource
)

func init() {
	// register plugin metric
	metricmanager.Register(clusterAvailability)
	metricmanager.Register(clusterWebhookCross)
	metricmanager.Register(clusterCheckDuration)
	metricmanager.Register(clusterApiserverCertificateExpiration)
	metricmanager.Register(clusterVersion)
}

// Setup 初始化插件配置
// 该函数负责：
// 1. 读取并解析配置文件（支持JSON和YAML格式）
// 2. 验证配置有效性
// 3. 初始化插件状态
// 4. 解析测试工作负载配置
// 5. 根据运行模式启动检查任务
//
// 参数:
//   - configFilePath: 配置文件路径
//   - runMode: 运行模式，支持RunModeDaemon（守护进程）和RunModeOnce（单次执行）
// 返回:
//   - error: 初始化过程中的错误
func (p *Plugin) Setup(configFilePath string, runMode string) error {
	// 读取配置文件
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("read clustercheck config file %s failed, err %s", configFilePath, err.Error())
	}
	
	// 尝试JSON格式解析
	p.opt = &Options{}
	if err = json.Unmarshal(configFileBytes, p.opt); err != nil {
		// JSON解析失败，尝试YAML格式解析
		if err = yaml.Unmarshal(configFileBytes, p.opt); err != nil {
			return fmt.Errorf("decode clustercheck config file %s failed, err %s", configFilePath, err.Error())
		}
	}

	// 验证配置
	if err = p.opt.Validate(); err != nil {
		return err
	}

	// 初始化插件状态
	p.Result = make(map[string]pluginmanager.CheckResult)
	p.ReadyMap = make(map[string]bool)

	// 解析测试工作负载配置
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
		// 守护进程模式：定期执行检查
		go func() {
			for {
				// 尝试获取检查锁，避免并发检查
				if p.CheckLock.TryLock() {
					p.CheckLock.Unlock()
					// 如果配置了同步模式，则获取插件管理器锁
					if p.opt.Synchronization {
						pluginmanager.Pm.Lock()
					}
					// 异步执行检查
					go p.Check(pluginmanager.CheckOption{})
				} else {
					klog.Infof("the former clustercheck didn't over, skip in this loop")
				}
				// 等待下一个检查周期或停止信号
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
		// 单次执行模式
		p.Check(pluginmanager.CheckOption{})
	}

	return nil
}

// Stop 停止插件
// 该函数通过向StopChan发送停止信号来优雅地停止插件运行
// 返回:
//   - error: 停止过程中的错误（当前始终返回nil）
func (p *Plugin) Stop() error {
	// 发送停止信号
	p.StopChan <- 1
	klog.Infof("plugin %s stopped", p.Name())
	return nil
}

// Name 返回插件名称
// 该函数实现了ClusterPlugin接口，返回插件标识名称
// 返回:
//   - string: 插件名称 "clustercheck"
func (p *Plugin) Name() string {
	return "clustercheck"
}

// int64Ptr 辅助函数，返回int64指针
func int64Ptr(i int64) *int64 { return &i }

// Check 执行集群检查，包括APIServer证书、控制面板可用性和存储结果
// 该函数会检查所有配置的集群，并执行以下检查项：
// 1. APIServer证书过期时间检查
// 2. 集群可用性黑名单检查
// 3. 通过创建测试工作负载来验证集群功能
// 4. Webhook跨命名空间检查
// 检查结果将存储在p.Result中，并通过Prometheus指标暴露
func (p *Plugin) Check(option pluginmanager.CheckOption) {
	// 记录检查开始时间
	start := time.Now()
	// 获取检查锁，确保同一时间只有一个检查在运行
	p.CheckLock.Lock()
	klog.Infof("start %s", p.Name())
	defer func() {
		klog.Infof("end %s", p.Name())
		// 如果配置了同步模式，释放插件管理器锁
		if p.opt.Synchronization {
			pluginmanager.Pm.UnLock()
		}
		// 释放检查锁
		p.CheckLock.Unlock()
		// 记录检查总耗时
		metricmanager.SetCommonDurationMetric([]string{"clustercheck", "", "", ""}, start)
	}()

	// 获取需要检查的集群配置
	clusterConfigs := make(map[string]*pluginmanager.ClusterConfig)
	// 如果未指定集群ID，则检查所有配置的集群
	if option.ClusterIDs == nil || len(option.ClusterIDs) == 0 {
		clusterConfigs = pluginmanager.Pm.GetConfig().ClusterConfigs
		// clean deleted cluster metric data
	} else {
		// 只检查指定的集群
		for _, clusterID := range option.ClusterIDs {
			if cluster, ok := pluginmanager.Pm.GetConfig().ClusterConfigs[clusterID]; ok {
				clusterConfigs[clusterID] = cluster
			}
		}
	}

	// 用于等待所有集群检查完成的WaitGroup
	wg := sync.WaitGroup{}

	// 遍历所有集群，并发执行检查
	for _, cluster := range clusterConfigs {
		wg.Add(1)
		// 增加等待计数
		if len(option.ClusterIDs) > 0 {
			pluginmanager.Pm.AddCheck()
		} else {
			pluginmanager.Pm.Add()
		}

		// 启动协程检查单个集群
		go func(cluster *pluginmanager.ClusterConfig) {
			defer func() {
				wg.Done()
				// 减少等待计数
				if len(option.ClusterIDs) > 0 {
					pluginmanager.Pm.DoneCheck()
				} else {
					pluginmanager.Pm.Done()
				}
			}()

			clusterId := cluster.ClusterID
			// 初始化检查结果
			clusterResult := pluginmanager.CheckResult{
				Items:        make([]pluginmanager.CheckItem, 0, 0),
				InfoItemList: make([]pluginmanager.InfoItem, 0, 0),
			}

			klog.Infof("start clustercheck for %s", clusterId)

			// 标记集群检查未就绪
			p.WriteLock.Lock()
			p.ReadyMap[cluster.ClusterID] = false
			p.WriteLock.Unlock()

			loopClusterChecktGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopClusterCheckDurationGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopCertificateExpirationGVSList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopClusterVersionGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)
			loopClusterWebhookGaugeVecSetList := make([]*metricmanager.GaugeVecSet, 0, 0)

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

			// check apiserver cert（异步执行，结果存入局部变量，Wait 后再合并，避免并发 append）
			getServerCertWG := sync.WaitGroup{}
			var certCheckItemList []pluginmanager.CheckItem
			var certGvsList []*metricmanager.GaugeVecSet
			var webhookGvs *metricmanager.GaugeVecSet
			if len(cluster.Master) > 0 {
				getServerCertWG.Add(1)
				go func() {
					defer func() {
						getServerCertWG.Done()
					}()
					var err error
					certCheckItemList, certGvsList, err = getApiserverCert(cluster)
					if err != nil {
						klog.Errorf("%s check apiserver cert expiration failed: %s", cluster.ClusterID, err.Error())
						certCheckItemList = nil
						certGvsList = nil
					}

					_, webhookList, err := CheckWebhookNamespaceCrossHook(cluster)
					if err != nil {
						klog.Errorf("%s CheckWebhookNamespaceCrossHook failed: %s", cluster.ClusterID, err.Error())
					} else if len(webhookList) > 0 {
						webhookGvs = &metricmanager.GaugeVecSet{
							Labels: []string{cluster.ClusterID, cluster.BusinessID, strings.Join(webhookList, ",")},
							Value:  1,
						}
					}

				}()
			}

			// blackbox check
			checkItemList, infoItemList, gvs, gvsList, err := testClusterByCreateUnstructuredObj(unstructuredObj, cluster)
			if err != nil {
				klog.Errorf("%s testClusterByCreateUnstructuredObj failed: %s", clusterId, err.Error())
			}

			// 等待异步证书检查执行完成，之后所有 append 在单协程内进行，避免并发数据竞争
			getServerCertWG.Wait()

			clusterResult.Items = append(clusterResult.Items, certCheckItemList...)
			clusterResult.Items = append(clusterResult.Items, checkItemList...)
			clusterResult.InfoItemList = append(clusterResult.InfoItemList, infoItemList...)
			loopClusterChecktGaugeVecSetList = append(loopClusterChecktGaugeVecSetList, gvs)
			loopClusterCheckDurationGaugeVecSetList = append(loopClusterCheckDurationGaugeVecSetList, gvsList...)
			loopClusterVersionGaugeVecSetList = append(loopClusterVersionGaugeVecSetList, &metricmanager.GaugeVecSet{
				Labels: []string{clusterId, cluster.BusinessID, cluster.Version},
				Value:  1,
			})
			loopCertificateExpirationGVSList = append(loopCertificateExpirationGVSList, certGvsList...)
			if webhookGvs != nil {
				loopClusterWebhookGaugeVecSetList = append(loopClusterWebhookGaugeVecSetList, webhookGvs)
			}
			klog.Infof("end clustercheck for %s", clusterId)

			// 获取写锁，更新指标和结果
			p.WriteLock.Lock()

			// 删除旧指标并设置新指标
			if _, ok := clusterCheckGaugeVecSetList[cluster.ClusterID]; !ok {
				// 首次检查，初始化指标列表
				clusterCheckGaugeVecSetList[clusterId] = make([]*metricmanager.GaugeVecSet, 0, 0)
				clusterWebhookGaugeVecSetList[clusterId] = make([]*metricmanager.GaugeVecSet, 0, 0)
				clusterCheckDurationGaugeVecSetList[clusterId] = make([]*metricmanager.GaugeVecSet, 0, 0)
				certificateExpirationGVSList[clusterId] = make([]*metricmanager.GaugeVecSet, 0, 0)
				clusterVersionGaugeVecSetList[clusterId] = make([]*metricmanager.GaugeVecSet, 0, 0)

				// 设置所有指标
				metricmanager.SetMetric(clusterAvailability, loopClusterChecktGaugeVecSetList)
				metricmanager.SetMetric(clusterWebhookCross, loopClusterWebhookGaugeVecSetList)
				metricmanager.SetMetric(clusterCheckDuration, loopClusterCheckDurationGaugeVecSetList)
				metricmanager.SetMetric(clusterApiserverCertificateExpiration, loopCertificateExpirationGVSList)
				metricmanager.SetMetric(clusterVersion, loopClusterVersionGaugeVecSetList)
			} else {
				// 非首次检查，先删除旧指标再设置新指标
				metricmanager.DeleteMetric(clusterAvailability, clusterCheckGaugeVecSetList[clusterId])
				metricmanager.SetMetric(clusterAvailability, loopClusterChecktGaugeVecSetList)
				metricmanager.DeleteMetric(clusterWebhookCross, clusterWebhookGaugeVecSetList[clusterId])
				metricmanager.SetMetric(clusterWebhookCross, loopClusterWebhookGaugeVecSetList)
				metricmanager.DeleteMetric(clusterCheckDuration, clusterCheckDurationGaugeVecSetList[clusterId])
				metricmanager.SetMetric(clusterCheckDuration, loopClusterCheckDurationGaugeVecSetList)
				metricmanager.DeleteMetric(clusterApiserverCertificateExpiration, certificateExpirationGVSList[clusterId])
				metricmanager.SetMetric(clusterApiserverCertificateExpiration, loopCertificateExpirationGVSList)
				metricmanager.DeleteMetric(clusterVersion, clusterVersionGaugeVecSetList[clusterId])
				metricmanager.SetMetric(clusterVersion, loopClusterVersionGaugeVecSetList)
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
			clusterWebhookGaugeVecSetList[clusterId] = loopClusterWebhookGaugeVecSetList
			clusterCheckDurationGaugeVecSetList[clusterId] = loopClusterCheckDurationGaugeVecSetList
			certificateExpirationGVSList[clusterId] = loopCertificateExpirationGVSList
			clusterVersionGaugeVecSetList[clusterId] = loopClusterVersionGaugeVecSetList

			p.ReadyMap[clusterId] = true
			p.WriteLock.Unlock()
		}(cluster)
	}

	wg.Wait()

	// 加锁清理已删除集群数据，避免与 Ready()/GetResult() 并发访问 map
	p.WriteLock.Lock()

	// clean deleted cluster data
	for clusterID := range p.ReadyMap {
		if _, ok := clusterConfigs[clusterID]; !ok {
			p.ReadyMap[clusterID] = false
			klog.Infof("delete cluster %s", clusterID)
		}
	}

	// 从readymap和指标中清理已删除集群
	for clusterID := range p.ReadyMap {
		if _, ok := clusterConfigs[clusterID]; !ok {
			delete(p.ReadyMap, clusterID)
			metricmanager.DeleteMetric(clusterAvailability, clusterCheckGaugeVecSetList[clusterID])
			metricmanager.DeleteMetric(clusterWebhookCross, clusterWebhookGaugeVecSetList[clusterID])
			metricmanager.DeleteMetric(clusterCheckDuration, clusterCheckDurationGaugeVecSetList[clusterID])
			metricmanager.DeleteMetric(clusterApiserverCertificateExpiration, certificateExpirationGVSList[clusterID])
			metricmanager.DeleteMetric(clusterVersion, clusterVersionGaugeVecSetList[clusterID])
			delete(clusterCheckGaugeVecSetList, clusterID)
			delete(clusterWebhookGaugeVecSetList, clusterID)
			delete(clusterCheckDurationGaugeVecSetList, clusterID)
			delete(certificateExpirationGVSList, clusterID)
			delete(clusterVersionGaugeVecSetList, clusterID)
			delete(p.Result, clusterID)
			klog.Infof("delete cluster %s", clusterID)
		}
	}

	p.WriteLock.Unlock()
}

// getApiserverCert 通过API端口获取APIServer证书过期时间
// 该函数会检查两种证书：
// 1. APIServer自签证书（使用apiserver-loopback-client作为域名）
// 2. APIServer证书（使用master地址作为域名）
// 参数:
//   - clusterConfig: 集群配置，包含Master节点地址信息
// 返回:
//   - []pluginmanager.CheckItem: 检查项列表
//   - []*metricmanager.GaugeVecSet: 指标集合列表
//   - error: 错误信息
func getApiserverCert(clusterConfig *pluginmanager.ClusterConfig) ([]pluginmanager.CheckItem, []*metricmanager.GaugeVecSet, error) {
	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	gvsList := make([]*metricmanager.GaugeVecSet, 0, 0)
	var expirationDate *time.Time
	// 检查apiserver自签证书，可能无法在cluster-reporter直通
	for _, master := range clusterConfig.Master {
		expiration, err := util.GetServerCert("apiserver-loopback-client", master, "60002")
		if err != nil {
			expiration, err = util.GetServerCert("apiserver-loopback-client", master, "6443")
			if err != nil {
				klog.Errorf("%s check apiserver self-signed cert expiration failed: %s", clusterConfig.ClusterID, err.Error())
				return checkItemList, gvsList, err
			}
		}

		if expirationDate == nil {
			expirationDate = &expiration
		} else if expiration.Before(*expirationDate) {
			expirationDate = &expiration
		}
	}

	// 证书检查结果
	checkItem := pluginmanager.CheckItem{
		ItemName:   ClusterApiserverCertExpirationCheckItem,
		ItemTarget: ApiserverTarget,
		Normal:     true,
		Status:     NormalStatus,
		Detail:     fmt.Sprintf(StringMap[AboutToExpireDetail], clusterConfig.ClusterID, expirationDate.Sub(time.Now())/time.Second),
		Level:      pluginmanager.WARNLevel,
		Tags:       nil,
	}

	// 时间在1周以内则返回异常
	if expirationDate.Sub(time.Now()) < 604800*time.Second {
		checkItem.Normal = false
		checkItem.Status = AboutToExpireStatus
		checkItem.SetDetail(fmt.Sprintf(StringMap[AboutToExpireDetail], clusterConfig.ClusterID, expirationDate.Sub(time.Now())/time.Second))
	}

	checkItemList = append(checkItemList, checkItem)

	gvsList = append(gvsList, &metricmanager.GaugeVecSet{
		Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, "self signed"},
		Value:  float64(expirationDate.Sub(time.Now()) / time.Second),
	})

	// 检查apiserver证书，可能无法在cluster-reporter直通
	expirationDate = nil
	for _, master := range clusterConfig.Master {
		expiration, err := util.GetServerCert(master, master, "60002")
		if err != nil {
			expiration, err = util.GetServerCert(master, master, "6443")
			if err != nil {
				klog.Errorf("%s check apiserver cert expiration failed: %s", clusterConfig.ClusterID, err.Error())
				return checkItemList, gvsList, err
			}
		}
		if expirationDate == nil {
			expirationDate = &expiration
		} else if expiration.Before(*expirationDate) {
			expirationDate = &expiration
		}
	}

	// 创建APIServer证书检查结果
	checkItem = pluginmanager.CheckItem{
		ItemName:   ClusterApiserverCertExpirationCheckItem,
		ItemTarget: ApiserverTarget,
		Normal:     true,
		Status:     NormalStatus,
		Detail:     fmt.Sprintf(StringMap[AboutToExpireDetail], clusterConfig.ClusterID, expirationDate.Sub(time.Now())/time.Second),
		Level:      pluginmanager.WARNLevel,
		Tags:       nil,
	}

	// 时间在1周以内则返回异常
	if expirationDate.Sub(time.Now()) < 604800*time.Second {
		checkItem.Normal = false
		checkItem.Status = AboutToExpireStatus
		checkItem.SetDetail(fmt.Sprintf(StringMap[AboutToExpireDetail], clusterConfig.ClusterID, expirationDate.Sub(time.Now())/time.Second))
	}

	checkItemList = append(checkItemList, checkItem)

	// 记录APIServer证书过期时间指标
	gvsList = append(gvsList, &metricmanager.GaugeVecSet{
		Labels: []string{clusterConfig.ClusterID, clusterConfig.BusinessID, "apiserver"},
		Value:  float64(expirationDate.Sub(time.Now()) / time.Second),
	})

	// 返回检查结果和指标
	return checkItemList, gvsList, nil
}

// testClusterByCreateUnstructuredObj 通过创建非结构化工作负载来测试集群可用性
// 该函数会执行以下操作：
// 1. 获取K8s集群版本，确认集群是否可访问
// 2. 创建测试命名空间（如果不存在）
// 3. 创建测试工作负载（Job）
// 4. 通过Watch机制监控Pod状态，记录各阶段耗时
// 5. 清理测试资源
//
// 参数:
//   - unstructuredObj: 非结构化对象，用于创建测试工作负载
//   - clusterConfig: 集群配置
// 返回:
//   - []pluginmanager.CheckItem: 检查项列表
//   - []pluginmanager.InfoItem: 信息项列表
//   - *metricmanager.GaugeVecSet: 集群可用性指标
//   - []*metricmanager.GaugeVecSet: 检查耗时指标列表
//   - error: 错误信息
func testClusterByCreateUnstructuredObj(unstructuredObj *unstructured.Unstructured, clusterConfig *pluginmanager.ClusterConfig,
) ([]pluginmanager.CheckItem, []pluginmanager.InfoItem, *metricmanager.GaugeVecSet, []*metricmanager.GaugeVecSet, error) {
	// 初始化返回结果
	checkItemList := make([]pluginmanager.CheckItem, 0, 0)
	infoItemList := make([]pluginmanager.InfoItem, 0, 0)
	gvsList := make([]*metricmanager.GaugeVecSet, 0, 0)
	var gvs *metricmanager.GaugeVecSet

	// 定义各阶段耗时
	var workloadToScheduleCost time.Duration  // 工作负载到调度耗时
	var workloadToPodCost time.Duration       // 工作负载到Pod创建耗时
	var worloadToRunningCost time.Duration   // 工作负载到运行耗时

	// 深拷贝非结构化对象，避免修改原始对象
	clusterUnstructuredObj := unstructuredObj.DeepCopy()
	// 随机workload名，避免重复导致的问题
	clusterUnstructuredObj.SetName("bcs-blackbox-job-" + time.Now().Format("2006150405"))
	var status string  // 检查状态

	// 初始化检查项
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

	// 延迟执行资源清理函数
	defer func() {
		// 清理残留resource
		go func() {
			backgroundDeletion := metav1.DeletePropagationBackground
			labelSelector := "bcs-cluster-reporter=bcs-cluster-reporter"

			// 1. 先删除 Job，阻止 Job Controller 继续创建新 Pod
			jobList, listErr := clusterConfig.ClientSet.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{
				ResourceVersion: "0",
				LabelSelector:   labelSelector,
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

			// 2. 显式清理残留 Pod，避免已完成（Completed）的 Pod 残留
			//    Background 级联删除依赖 GC 异步清理，可能出现 Pod 残留；
			//    上次清理失败时也会产生孤儿 Pod，此处统一兜底清理
			podList, podListErr := clusterConfig.ClientSet.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
				ResourceVersion: "0",
				LabelSelector:   labelSelector,
			})
			if podListErr != nil {
				klog.Errorf("%s get pod failed: %s", clusterConfig.ClusterID, podListErr.Error())
			} else if len(podList.Items) > 0 {
				klog.Infof("%s found %d residual pod(s) to delete", clusterConfig.ClusterID, len(podList.Items))
				for _, pod := range podList.Items {
					if podErr := clusterConfig.ClientSet.CoreV1().Pods(namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{
						GracePeriodSeconds: int64Ptr(5),
					}); podErr != nil {
						klog.Errorf("%s delete pod %s failed: %s", clusterConfig.ClusterID, pod.Name, podErr.Error())
					} else {
						klog.Infof("%s delete pod %s success", clusterConfig.ClusterID, pod.Name)
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

// getResourceInterface 获取动态资源接口
// 该函数会执行以下操作：
// 1. 检查或创建测试命名空间
// 2. 获取Discovery客户端
// 3. 创建动态客户端
// 4. 返回指定命名空间的资源接口
//
// 参数:
//   - clusterConfig: 集群配置
//   - clusterUnstructuredObj: 非结构化对象（未使用）
//   - status: 状态指针，用于返回错误状态
// 返回:
//   - dynamic.ResourceInterface: 动态资源接口
//   - error: 错误信息
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

// getWatchStatus 获取工作负载的Pod状态并返回
// 该函数通过Watch机制监控Pod的生命周期事件：
// 1. 启动Informer监听Pod事件
// 2. 创建测试工作负载
// 3. 监控Pod的创建、调度、运行状态
// 4. 记录各阶段耗时
// 5. 检测并处理各种异常情况
//
// 参数:
//   - config: Kubernetes客户端接口
//   - clusterUnstructuredObj: 非结构化对象
//   - dri: 动态资源接口
//   - namespace: 命名空间
//   - clusterID: 集群ID
// 返回:
//   - status: 检查状态
//   - workloadToScheduleCost: 工作负载到调度耗时
//   - workloadToPodCost: 工作负载到Pod创建耗时
//   - worloadToRunningCost: 工作负载到运行耗时
//   - err: 错误信息
func getWatchStatus(config kubernetes.Interface, clusterUnstructuredObj *unstructured.Unstructured,
	dri dynamic.ResourceInterface, namespace string, clusterID string) (status string,
	workloadToScheduleCost, workloadToPodCost, worloadToRunningCost time.Duration, err error) {
	// 记录函数开始时间
	startTime := time.Now()

	defer func() {
		klog.Infof("%s getWatchStatus duration %.2f s", clusterID, float64(time.Now().Sub(startTime)/time.Second))
	}()

	// 标记Pod是否已创建
	createPodFlag := false
	// 记录工作负载创建开始时间
	var createStartTime time.Time

	// 启动watch，观察对应任务的pod的所有事件
	factory := informers.NewSharedInformerFactoryWithOptions(config, 0, informers.WithNamespace(namespace))
	informer := factory.Core().V1().Pods().Informer()
	stopchan := make(chan struct{})
	safechan := util.NewSafeChannel(stopchan)
	
	// 添加事件处理器，监听Pod的增删改事件
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Pod创建事件
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			// 只处理与测试工作负载相关的Pod
			if strings.Contains(pod.Name, clusterUnstructuredObj.GetName()) {
				// 首次检测到Pod创建
				if !createPodFlag {
					klog.Infof("cluster %s create pod successful", clusterID)
					// pod创建成功时间
					workloadToPodCost = pod.CreationTimestamp.Sub(createStartTime)
					createPodFlag = true
				}
				// Pod已被调度到节点
				if pod.Spec.NodeName != "" {
					if !safechan.Done() {
						status = NormalStatus
						// pod调度成功耗时，调度成功直接返回
						klog.Infof("cluster %s schedule pod successful", clusterID)
						workloadToScheduleCost, worloadToRunningCost = getPodLifeCycleTimePoint(pod, createStartTime)
						safechan.SafeClose()
					}
				} else {
					// 判断是否因为没有可供调度的node，是则返回没有node，区分未调度
					for _, condition := range pod.Status.Conditions {
						if strings.Contains(condition.Message, "nodes are available") {
							if !safechan.Done() {
								status = AvailabilityNoNodeErrorStatus
								klog.Infof("%s scheduled pod failed: %s", clusterID, condition.Message)
								safechan.SafeClose()
							}
						}
					}
				}
			}
		},
		// Pod更新事件
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*v1.Pod)
			// 只处理与测试工作负载相关的Pod
			if strings.Contains(pod.Name, clusterUnstructuredObj.GetName()) {
				// Pod已被调度到节点
				if pod.Spec.NodeName != "" {
					if !safechan.Done() {
						status = NormalStatus
						// pod调度成功耗时，调度成功直接返回
						klog.Infof("cluster %s schedule pod successful", clusterID)
						workloadToScheduleCost, worloadToRunningCost = getPodLifeCycleTimePoint(pod, createStartTime)

						safechan.SafeClose()
					}
				} else {
					// 判断是否因为没有可供调度的node，是则返回没有node，区分未调度
					for _, condition := range pod.Status.Conditions {
						if strings.Contains(condition.Message, "nodes are available") {
							if !safechan.Done() {
								status = AvailabilityNoNodeErrorStatus
								klog.Infof("%s scheduled pod failed: %s", clusterID, condition.Message)
								safechan.SafeClose()
							}
						}
					}
				}
			}

		},
		// Pod删除事件（空实现）
		DeleteFunc: func(obj interface{}) {
		},
	})

	klog.Infof("%s start informer", clusterID)
	go informer.Run(stopchan)

	defer func() {
		// 关闭informer
		if !safechan.Done() {
			safechan.SafeClose()
		}
	}()

	// 记录发起创建workload的时间
	createStartTime = time.Now()
	// 创建workload，设置90秒超时
	ctx := util.GetCtx(90 * time.Second)
	klog.Infof("%s start create workload", clusterID)
	testObj, err := dri.Create(ctx, clusterUnstructuredObj, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("%s Create failed %s", clusterID, err.Error())
		// 处理创建失败的各种情况
		if strings.Contains(err.Error(), "already exists") {
			// 工作负载已存在，等待5秒后返回
			time.Sleep(5 * time.Second)
			status = AvailabilityWorkloadExistStatus
		} else {
			// 其他创建错误
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
			// 根据Pod创建情况判断超时原因
			if status != NormalStatus && status != AvailabilityNoNodeErrorStatus {
				if !createPodFlag {
					// Pod未创建成功
					status = AvailabilityCreatePodTimeoutStatus
					klog.Errorf("%s create pod timeout", clusterID)
				} else {
					// Pod已创建但未调度成功
					status = AvailabilitySchedulePodTimeoutStatus
					klog.Errorf("%s wait scheduled watch event timeout", clusterID)
				}
			}
			return

		case <-stopchan:
			// Informer停止
			klog.Infof("%s informer stopped", clusterID)
			// 根据Pod创建情况判断停止原因
			if status != NormalStatus && status != AvailabilityNoNodeErrorStatus {
				if !createPodFlag {
					// Pod未创建成功
					status = AvailabilityCreatePodTimeoutStatus
					klog.Errorf("%s create pod timeout", clusterID)
				} else {
					// Pod已创建但未调度成功
					status = AvailabilitySchedulePodTimeoutStatus
					klog.Errorf("%s wait scheduled watch event timeout", clusterID)
				}
			}
			return
		}
	}
}

// getPodLifeCycleTimePoint 获取Pod生命周期各阶段的耗时
// 该函数会检查Pod的Conditions，提取以下时间点：
// 1. PodScheduled: Pod被调度到节点的时间
// 2. PodReady: Pod准备就绪的时间
//
// 参数:
//   - pod: Pod对象
//   - createStartTime: 工作负载创建开始时间
// 返回:
//   - time.Duration: 调度耗时（从创建到调度）
//   - time.Duration: 运行耗时（从创建到就绪）
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

// Ready 如果集群检查已完成，返回true
// 该函数用于检查指定集群的检查是否已结束
// 参数:
//   - clusterID: 集群ID
// 返回:
//   - bool: 检查是否完成
func (p *Plugin) Ready(clusterID string) bool {
	p.WriteLock.Lock()
	defer p.WriteLock.Unlock()
	return p.ReadyMap[clusterID]
}

// GetResult 根据集群ID返回检查结果
// 该函数用于获取指定集群的最近一次检查结果
// 参数:
//   - clusterID: 集群ID
// 返回:
//   - pluginmanager.CheckResult: 检查结果
func (p *Plugin) GetResult(clusterID string) pluginmanager.CheckResult {
	p.WriteLock.Lock()
	defer p.WriteLock.Unlock()
	return p.Result[clusterID]
}
