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
	bcsv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubebkbcs/apis/bk-bcs/v1"
)

const (
	// KubeSystemNamespace is k8s system namespace
	KubeSystemNamespace = "kube-system"
	// BCSSystemNamespace is bcs system namespace
	BCSSystemNamespace = "bcs-system"
	// KubePublicNamespace is k8s public namespace
	KubePublicNamespace = "kube-public"
	// RawDataName is data name of bcs/k8s system log in bkdata
	RawDataName = "bcs_k8s_system_log" // dataname use _ instead of -
	// DeployAPIName is api name for deploy new data access in bkdata
	DeployAPIName = "v3_access_deploy_plan_post"
	// DataCleanAPIName is api name for creat new data clean strategy in bkdata
	DataCleanAPIName = "v3_databus_cleans_post"
	// APIGatewayClusterTunnel is subpath of api-gateway for accessing cluster info
	APIGatewayClusterTunnel = "/tunnels/clusters/"
	// SchemaHTTP http schema
	SchemaHTTP = "http://"
	// SchemaHTTPS https schema
	SchemaHTTPS = "https://"
	// DefaultLogConfigNamespace is default namespace for bcslogconfigs CRD
	DefaultLogConfigNamespace = "default"
)

var (
	// SystemNamspaces includes namespaces described above
	SystemNamspaces = []string{"kube-system", "bcs-system", "kube-public"}
	// BKDataAPIConfigKind is resource name of BkDataAPIConfig
	BKDataAPIConfigKind string
	// BKDataAPIConfigGroupVersion is resouce name of BkDataAPIConfig
	BKDataAPIConfigGroupVersion string
	// LogConfigKind is crd name of bcslogconfigs
	LogConfigKind string
	// LogConfigAPIVersion is api version of bcslogconfigs
	LogConfigAPIVersion = "bkbcs.tencent.com/v1"
)

func init() {
	BKDataAPIConfigKind = reflect.TypeOf(bkdatav1.BKDataApiConfig{}).Name()
	BKDataAPIConfigGroupVersion = fmt.Sprintf("%s/%s", bkdatav1.SchemeGroupVersion.Group, bkdatav1.SchemeGroupVersion.Version)
	LogConfigKind = reflect.TypeOf(bcsv1.BcsLogConfig{}).Name()
	LogConfigAPIVersion = fmt.Sprintf("%s/%s", bcsv1.SchemeGroupVersion.Group, bcsv1.SchemeGroupVersion.Version)
}

// NewManager returns a new log manager
func NewManager(conf *config.ManagerConfig) LogManagerInterface {
	manager := &LogManager{
		stopCh:     conf.StopCh,
		ctx:        conf.Ctx,
		config:     conf,
		logClients: make(map[string]*LogClient),
		// controllers:             make(map[string]*ClusterLogController),
		dataidChMap:             make(map[string]chan string),
		GetLogCollectionTask:    make(chan *RequestMessage),
		AddLogCollectionTask:    make(chan *RequestMessage),
		DeleteLogCollectionTask: make(chan *RequestMessage),
	}
	cli := bcsapi.NewClient(&conf.BcsAPIConfig)
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
	manager.bkDataAPIConfigClientset, err = internalclientset.NewForConfig(restConf)
	if err != nil {
		blog.Errorf("build BKDataApiConfig clientset by kubeconfig %s error %s", conf.KubeConfig, err.Error())
		return nil
	}
	internalFactory := externalversions.NewSharedInformerFactory(manager.bkDataAPIConfigClientset, time.Hour)
	manager.bkDataAPIConfigInformer = internalFactory.Bkbcs().V1().BKDataApiConfigs().Informer()
	internalFactory.Start(conf.StopCh)
	// Wait for all caches to sync.
	internalFactory.WaitForCacheSync(conf.StopCh)
	//add k8s resources event handler functions
	manager.bkDataAPIConfigInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: manager.handleUpdatedBKDataAPIConfig,
		},
	)
	return manager
}

func (m *LogManager) addSystemCollectionConfig() {
	if m.config.SystemDataID == "" {
		dataid, ok := m.obtainDataID(bkdata.BKDataClientConfig{
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

// obtainDataID is used to obtain dataid for unspecified system log dataid
func (m *LogManager) obtainDataID(clientconf bkdata.BKDataClientConfig, bizid int, dataname string) (string, bool) {
	dataname = strings.ToLower(dataname)
	if _, ok := m.dataidChMap[dataname]; ok {
		blog.Errorf("Dataname %s already existed.", dataname)
		return "", false
	}
	deployconfig := bkdata.NewDefaultAccessDeployPlanConfig()
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
	bkdataapiconfig.TypeMeta.APIVersion = BKDataAPIConfigGroupVersion
	bkdataapiconfig.TypeMeta.Kind = BKDataAPIConfigKind
	bkdataapiconfig.SetName(strings.ReplaceAll(dataname, "_", "-"))
	bkdataapiconfig.SetNamespace("default")
	bkdataapiconfig.Spec.ApiName = DeployAPIName
	bkdataapiconfig.Spec.AccessDeployPlanConfig = deployconfig
	blog.Infof("Apply for dataid with bkdataapiconfig : %+v, waitting for response...", *bkdataapiconfig)
	_, err := m.bkDataAPIConfigClientset.BkbcsV1().BKDataApiConfigs(bkdataapiconfig.GetNamespace()).Create(bkdataapiconfig)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		blog.Errorf("Create BKDataApiConfig crd failed: %s, crd info: %+v", err.Error(), *bkdataapiconfig)
		return "", false
	}
	select {
	case dataid, ok := <-dataidCh:
		if !ok {
			blog.Errorf("Obtain dataid failed for system log dataid before create BKDataApiConfig crd")
			return "", false
		}
		blog.Infof("Get response with dataid [%s], crdinfo: %+v", dataid, *bkdataapiconfig)
		close(dataidCh)
		return dataid, true
	case <-time.After(time.Second * 10):
		blog.Errorf("Obtain dataid failed for system log dataid: timeout")
		return "", false
	}
}

// Start start the log manager
func (m *LogManager) Start() {
	go m.run()
}

// start log manager
func (m *LogManager) run() {
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
		blog.Infof("Total Clusters: %n", len(ccinfo))
		blog.Infof("ListAllClusters success")
		var schema string
		if m.config.BcsAPIConfig.TLSConfig != nil {
			schema = SchemaHTTPS
		} else {
			schema = SchemaHTTP
		}
		newClusters := make(map[string]*LogClient)
		// find new clusters and deleted clusters
		m.clientRWMutex.Lock()
		for _, cc := range ccinfo {
			id := strings.ToLower(cc.ClusterID)
			if _, ok := m.logClients[id]; ok {
				continue
			}
			// new cluster
			blog.V(3).Infof("New cluster: %+v", cc)
			restConf := &rest.Config{
				Host:        fmt.Sprintf("%s%s%s%s", schema, m.config.BcsAPIConfig.Hosts[0], APIGatewayClusterTunnel, cc.ClusterID),
				BearerToken: m.config.BcsAPIConfig.AuthToken,
				// TODO TLS security
				TLSClientConfig: rest.TLSClientConfig{
					Insecure: true,
				},
				Timeout: time.Second * 10,
			}
			clientset, err := internalclientset.NewForConfig(restConf)
			if err != nil {
				blog.Errorf("Clientset initialization failed: server %s, cluster %s, %s", restConf.Host, cc.ClusterID, err.Error())
				continue
			}
			m.logClients[id] = &LogClient{
				ClusterInfo: cc,
				Client:      clientset.BkbcsV1().RESTClient(),
			}
			blog.Infof("Create cluster bcslogconfig controller success")
			newClusters[id] = m.logClients[id]
		}
		m.clientRWMutex.Unlock()
		m.distributeAddTasks(m.ctx, newClusters, m.config.CollectionConfigs)
		// delete invalid clusters
		deletedClusters := make(map[string]struct{})
		m.clientRWMutex.RLock()
		for id := range m.logClients {
			deletedClusters[id] = struct{}{}
		}
		m.clientRWMutex.RUnlock()
		for _, cc := range ccinfo {
			delete(deletedClusters, strings.ToLower(cc.ClusterID))
		}
		m.clientRWMutex.Lock()
		for id := range deletedClusters {
			blog.Infof("Delete deleted cluster (%s)", id)
			delete(m.logClients, id)
		}
		m.clientRWMutex.Unlock()
		<-time.After(time.Minute)
	}
}

// get dataid from crd
func (m *LogManager) handleUpdatedBKDataAPIConfig(oldobj, newobj interface{}) {
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
	} else {
		blog.Errorf("Obtain dataid failed")
		close(m.dataidChMap[name])
		return
	}
}

func (m *LogManager) getLogClients() map[string]*LogClient {
	ret := make(map[string]*LogClient)
	m.clientRWMutex.RLock()
	for k, v := range m.logClients {
		ret[k] = v
	}
	m.clientRWMutex.RUnlock()
	return ret
}
