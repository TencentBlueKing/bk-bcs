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

package k8s

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/apigateway/bkdata"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	bkdatav1 "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/apis/bkbcs.tencent.com/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/informers/externalversions"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
)

const (
	KubeSystemNamespace = "kube-system"
	BCSSystemNamespace  = "bcs-system"
	KubePublicNamespace = "kube-public"
	RawDataName         = "bcs_k8s_system_log" // dataname use _ instead of -
	DeployAPIName       = "v3_access_deploy_plan_post"
	DataCleanAPIName    = "v3_databus_cleans_post"
)

var (
	SystemNamspaces             = []string{"kube-system", "bcs-system", "kube-public"}
	BKDataApiConfigKind         string
	BKDataApiConfigGroupVersion string
)

type RequestMessage struct {
	Data   interface{}
	RespCh chan interface{}
}

type LogManager struct {
	GetLogCollectionTask    chan *RequestMessage
	AddLogCollectionTask    chan *RequestMessage
	DeleteLogCollectionTask chan *RequestMessage

	userManagerCli           *bcsapi.UserManagerCli
	config                   *config.ManagerConfig
	controllers              map[string]*ClusterLogController
	dataidChMap              map[string]chan string
	currCollectionConfigInd  int
	bkDataApiConfigClientset *internalclientset.Clientset
	bkDataApiConfigInformer  cache.SharedIndexInformer
	stopCh                   chan struct{}
}

func init() {
	BKDataApiConfigKind = reflect.TypeOf(bkdatav1.BKDataApiConfig{}).Name()
	BKDataApiConfigGroupVersion = fmt.Sprintf("%s/%s", bkdatav1.SchemeGroupVersion.Group, bkdatav1.SchemeGroupVersion.Version)
}

func NewManager(conf *config.ManagerConfig) *LogManager {
	manager := &LogManager{
		stopCh:                  conf.StopCh,
		config:                  conf,
		controllers:             make(map[string]*ClusterLogController),
		dataidChMap:             make(map[string]chan string),
		GetLogCollectionTask:    make(chan *RequestMessage),
		AddLogCollectionTask:    make(chan *RequestMessage),
		DeleteLogCollectionTask: make(chan *RequestMessage),
	}
	cli := bcsapi.NewClient(&conf.BcsApiConfig)
	manager.userManagerCli = cli.UserManager()

	var restConf *rest.Config
	var err error
	if conf.KubeConfig != "" {
		restConf, err = clientcmd.BuildConfigFromFlags("", conf.KubeConfig)
	} else {
		restConf, err = rest.InClusterConfig()
	}
	if err != nil {
		blog.Errorf("build kubeconfig %s error :%s", conf.KubeConfig, err.Error())
		return nil
	}

	//internal clientset for informer BKDataApiConfig Crd
	manager.bkDataApiConfigClientset, err = internalclientset.NewForConfig(restConf)
	if err != nil {
		blog.Errorf("build BKDataApiConfig clientset by kubeconfig %s error %s", conf.KubeConfig, err.Error())
		return nil
	}
	internalFactory := externalversions.NewSharedInformerFactory(manager.bkDataApiConfigClientset, time.Hour)
	manager.bkDataApiConfigInformer = internalFactory.Bkbcs().V1().BKDataApiConfigs().Informer()
	internalFactory.Start(conf.StopCh)
	// Wait for all caches to sync.
	internalFactory.WaitForCacheSync(conf.StopCh)
	//add k8s resources event handler functions
	manager.bkDataApiConfigInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: manager.handleUpdatedBKDataApiConfig,
		},
	)
	return manager
}

func (m *LogManager) addSystemCollectionConfig() {
	if m.config.SystemDataID == "" {
		dataid, ok := m.ObtainDataId(bkdata.BKDataClientConfig{
			BkAppCode:   m.config.BkAppCode,
			BkAppSecret: m.config.BkAppSecret,
			BkUsername:  m.config.BkUsername,
		}, m.config.BkBizID, RawDataName)
		if !ok {
			return
		}
		m.config.SystemDataID = dataid
	}
	for _, namespace := range SystemNamspaces {
		collectConfig := config.CollectionConfig{
			ConfigNamespace: "default",
			ConfigName:      strings.ToLower(fmt.Sprintf("%s-%s", LogConfigKind, namespace)),
			ConfigSpec: bcsv1.BcsLogConfigSpec{
				ConfigType:        "custom",
				Stdout:            true,
				StdDataId:         m.config.SystemDataID,
				WorkloadNamespace: namespace,
				PodLabels:         true,
				LogTags: map[string]string{
					"platform": "bcs",
				},
			},
			ClusterIDs: "",
		}
		m.config.CollectionConfigs = append(m.config.CollectionConfigs, collectConfig)
	}
	blog.Infof("log config list: %+v", m.config.CollectionConfigs)
	blog.Infof("System log configs ready to create")
}

func (m *LogManager) ObtainDataId(clientconf bkdata.BKDataClientConfig, bizid int, dataname string) (string, bool) {
	dataname = strings.ToLower(dataname)
	if _, ok := m.dataidChMap[dataname]; ok {
		blog.Errorf("Dataname %s already existed.", dataname)
		return "", false
	}
	deployconfig := bkdata.NewDefaultCustomAccessDeployPlanConfig()
	deployconfig.BkAppCode = clientconf.BkAppCode
	deployconfig.BkAppSecret = clientconf.BkAppSecret
	deployconfig.BkUsername = clientconf.BkUsername
	deployconfig.BkBizID = bizid
	deployconfig.AccessRawData.Maintainer = clientconf.BkUsername
	deployconfig.AccessRawData.RawDataName = dataname
	deployconfig.AccessRawData.RawDataAlias = dataname
	dataidCh := make(chan string)
	m.dataidChMap[dataname] = dataidCh
	defer delete(m.dataidChMap, dataname)

	// create BKDataApiConfig crd
	bkdataapiconfig := &bkdatav1.BKDataApiConfig{}
	bkdataapiconfig.TypeMeta.APIVersion = BKDataApiConfigGroupVersion
	bkdataapiconfig.TypeMeta.Kind = BKDataApiConfigKind
	bkdataapiconfig.SetName(strings.ReplaceAll(dataname, "_", "-"))
	bkdataapiconfig.SetNamespace("default")
	bkdataapiconfig.Spec.ApiName = DeployAPIName
	bkdataapiconfig.Spec.AccessDeployPlanConfig = deployconfig
	blog.Infof("Apply for dataid with bkdataapiconfig : %+v, waitting for response...", *bkdataapiconfig)
	_, err := m.bkDataApiConfigClientset.BkbcsV1().BKDataApiConfigs(bkdataapiconfig.GetNamespace()).Create(bkdataapiconfig)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		blog.Errorf("Create BKDataApiConfig crd failed: %s, crd info: %+v", err.Error(), *bkdataapiconfig)
		return "", false
	}
	dataid, ok := <-dataidCh
	if !ok {
		blog.Errorf("Obtain dataid failed for system log dataid before create BKDataApiConfig crd")
		return "", false
	}
	blog.Infof("Get response with dataid [%s], crdinfo: %+v", dataid, *bkdataapiconfig)
	close(dataidCh)
	return dataid, true
}

func (m *LogManager) Start() {
	go m.run()
	go m.apiService()
}

func (m *LogManager) run() {
	var cnt int64
	cnt = 0
	m.addSystemCollectionConfig()
	for {
		select {
		case _, ok := <-m.stopCh:
			if !ok {
				blog.Errorf("Stop channel closed, cluster manager stop working")
				return
			}
		default:
			break
		}
		// sync cluster infos
		blog.Infof("Begin to sync clusers info")
		ccinfo, err := m.userManagerCli.ListAllClusters()
		if err != nil {
			blog.Errorf("ListAllClusters failed: %s", err.Error())
			<-time.After(time.Minute)
			continue
		}
		blog.Infof("Clusters: %+v", ccinfo)
		blog.Infof("ListAllClusters success")
		cnt++
		newClusters := make(map[string]*ClusterLogController)
		// find new clusters and deleted clusters
		for _, cc := range ccinfo {
			if _, ok := m.controllers[cc.ClusterID]; ok {
				m.controllers[cc.ClusterID].SetTick(cnt)
				continue
			}
			// new cluster
			blog.Infof("New cluster: %+v", cc)
			controller, err := NewClusterLogController(&config.ControllerConfig{Credential: cc, CAFile: m.config.CAFile})
			if err != nil {
				blog.Errorf("Create Cluster Log Controller failed, Cluster Id: %s, Cluster Domain: %s, error info: %s", cc.ClusterID, cc.ClusterDomain, err.Error())
				continue
			}
			blog.Infof("Create cluster bcslogconfig controller success")
			controller.SetTick(cnt)
			m.controllers[cc.ClusterID] = controller
			newClusters[cc.ClusterID] = controller
			controller.Start()
		}
		// delete invalid clusters
		for k, v := range m.controllers {
			if v.GetTick() == cnt {
				continue
			}
			m.controllers[k].Stop()
			blog.Infof("Stop deleted cluster (%s) controller", k)
			delete(m.controllers, k)
		}
		m.distributeTasks(newClusters, m.config.CollectionConfigs)
		<-time.After(time.Minute)
	}
}

func (m *LogManager) apiService() {
	for {
		select {
		// get log configs
		case msg, ok := <-m.GetLogCollectionTask:
			if !ok {
				blog.Errorf("Get request data from api server failed, API service crashed")
				return
			}
			switch conf := msg.Data.(type) {
			case *config.CollectionFilterConfig:
				blog.Infof("Get CollectionFilterConfig for GetLogCollectionTask: %+v", conf)
				confsList := m.getLogCollectionTaskByFilter(conf)
				for _, confs := range confsList {
					for _, c := range confs {
						msg.RespCh <- c
					}
				}
				msg.RespCh <- "termination"
			default:
				blog.Errorf("Unrecognized data type received from api server while get log collection tasks, data value (%+v)", conf)
				msg.RespCh <- fmt.Errorf("Unrecognized data type received from api server while get log collection tasks, data value (%+v)", conf)
			}
		// create log config
		case msg, ok := <-m.AddLogCollectionTask:
			if !ok {
				blog.Errorf("Get request data from api server failed, API service crashed")
				return
			}
			switch conf := msg.Data.(type) {
			case *config.CollectionConfig:
				blog.Infof("Get CollectionConfig for AddLogCollectionTask: %+v", conf)
				m.distributeTasks(m.controllers, []config.CollectionConfig{*conf})
				msg.RespCh <- "termination"
			default:
				blog.Errorf("Unrecognized data type received from api server while get log collection tasks, data value (%+v)", conf)
				msg.RespCh <- fmt.Errorf("Unrecognized data type received from api server while get log collection tasks, data value (%+v)", conf)
			}
		// delete log config
		case msg, ok := <-m.DeleteLogCollectionTask:
			if !ok {
				blog.Errorf("Get request data from api server failed, API service crashed")
				return
			}
			switch conf := msg.Data.(type) {
			case *config.CollectionFilterConfig:
				blog.Infof("Get CollectionFilterConfig for DeleteLogCollectionTask: %+v", conf)
				m.distributeDeleteTasks(conf)
				msg.RespCh <- "termination"
			default:
				blog.Errorf("Unrecognized data type received from api server while get log collection tasks, data value (%+v)", conf)
				msg.RespCh <- fmt.Errorf("Unrecognized data type received from api server while get log collection tasks, data value (%+v)", conf)
			}
		}
	}
}

func (m *LogManager) distributeTasks(newClusters map[string]*ClusterLogController, confs []config.CollectionConfig) {
	blog.Infof("Start distribute log configs to clusters")
	blog.Infof("log config list: %+v", confs)
	for _, logconf := range confs {
		blog.Infof("distribute config : %+v", logconf)
		if logconf.ClusterIDs == "" {
			for k, ctrl := range newClusters {
				ctrl.AddCollectionTask <- logconf
				blog.Infof("Send logconf to cluster %s", k)
			}
			continue
		}
		clusters := strings.Split(strings.ToLower(logconf.ClusterIDs), ",")
		for _, clusterid := range clusters {
			if _, ok := newClusters[clusterid]; !ok {
				blog.Errorf("Wrong cluster ID %s of collection config %+v", clusterid, logconf)
				continue
			}
			newClusters[clusterid].AddCollectionTask <- logconf
			blog.Infof("Send logconf to cluster %s", clusterid)
		}
	}
}

func (m *LogManager) getLogCollectionTaskByFilter(filter *config.CollectionFilterConfig) [][]config.CollectionConfig {
	var ret [][]config.CollectionConfig
	if filter.ClusterIDs == "" {
		ret = make([][]config.CollectionConfig, len(m.controllers))
		var i = 0
		for _, ctl := range m.controllers {
			ret[i] = ctl.getLogCollectionTaskByFilter(filter)
			i++
		}
	} else {
		clusters := strings.Split(filter.ClusterIDs, ",")
		for _, id := range clusters {
			if ctl, ok := m.controllers[id]; !ok {
				blog.Warnf("No cluster id (%s)", id)
				continue
			} else {
				ret = append(ret, ctl.getLogCollectionTaskByFilter(filter))
				// to be deleted
				blog.Errorf("get log collection configs: %+v", ctl.getLogCollectionTaskByFilter(filter))
			}
		}
	}
	return ret
}

func (m *LogManager) distributeDeleteTasks(filter *config.CollectionFilterConfig) {
	if filter.ClusterIDs == "" {
		for _, ctl := range m.controllers {
			ctl.DeleteCollectionTask <- filter
		}
	} else {
		clusters := strings.Split(filter.ClusterIDs, ",")
		for _, id := range clusters {
			if ctl, ok := m.controllers[id]; !ok {
				blog.Warnf("No cluster id (%s)", id)
				continue
			} else {
				ctl.DeleteCollectionTask <- filter
			}
		}
	}
}

func (m *LogManager) handleUpdatedBKDataApiConfig(oldobj, newobj interface{}) {
	config, ok := newobj.(*bkdatav1.BKDataApiConfig)
	if !ok {
		blog.Errorf("Convert object to BKDataApiConfig failed")
		return
	}
	if config.Spec.ApiName != DeployAPIName {
		blog.Info("Not deploy plan config, ignore")
		return
	}
	name := config.Spec.AccessDeployPlanConfig.AccessRawData.RawDataName
	if _, ok = m.dataidChMap[name]; !ok {
		blog.Warnf("No dataid channel named %s, ignore", name)
		return
	}
	if config.Spec.Response.Result {
		var obj map[string]interface{}
		err := json.Unmarshal([]byte(config.Spec.Response.Data), &obj)
		if err != nil {
			blog.Errorf("Convert from BKDataApi response to interface failed: %s", err.Error())
			close(m.dataidChMap[name])
			return
		}
		val, ok := obj["dataid"]
		if !ok {
			blog.Errorf("BKDataApi response does not contain dataid field")
			close(m.dataidChMap[name])
			return
		}
		dataidf, ok := val.(float64)
		if !ok {
			blog.Errorf("Parse dataid from BKDataApi response failed: type assertion failed")
			close(m.dataidChMap[name])
			return
		}
		dataid := int(dataidf)
		blog.Info("Obtain dataid [%d] success of dataname: %s", dataid, name)
		m.dataidChMap[name] <- strconv.Itoa(dataid)
		m.newDataCleanStrategy(config, dataid)
	} else {
		blog.Errorf("Obtain dataid failed")
		close(m.dataidChMap[name])
		return
	}
}

func (m *LogManager) newDataCleanStrategy(config *bkdatav1.BKDataApiConfig, dataid int) {
	// new data clean strategy crd
	strategy := bkdata.NewDefaultLogCollectionDataCleanStrategy()
	strategy.BkAppCode = config.Spec.AccessDeployPlanConfig.BkAppCode
	strategy.BkAppSecret = config.Spec.AccessDeployPlanConfig.BkAppSecret
	strategy.BkUsername = config.Spec.AccessDeployPlanConfig.BkUsername
	strategy.BkBizID = config.Spec.AccessDeployPlanConfig.BkBizID
	strategy.RawDataID = int(dataid)

	bkdataapiconfig := &bkdatav1.BKDataApiConfig{}
	bkdataapiconfig.Spec.ApiName = DataCleanAPIName
	bkdataapiconfig.SetName(fmt.Sprintf("%s-data-clean-strategy", config.GetName()))
	bkdataapiconfig.SetNamespace("default")
	bkdataapiconfig.Spec.DataCleanStrategyConfig = strategy
	// delete successful obtain dataid crd
	m.bkDataApiConfigClientset.BkbcsV1().BKDataApiConfigs(config.GetNamespace()).Delete(config.GetName(), &metav1.DeleteOptions{})
	// apply data clean strategy
	_, err := m.bkDataApiConfigClientset.BkbcsV1().BKDataApiConfigs(bkdataapiconfig.GetNamespace()).Create(bkdataapiconfig)
	if err != nil {
		blog.Errorf("Create BKDataApiConfig crd failed: %s, crd info: %+v", err.Error(), *bkdataapiconfig)
	}
}
