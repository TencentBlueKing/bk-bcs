/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	clusterops "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	cmcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	etcdlock "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock/etcd"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/passcc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/user"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tkehandler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tunnel"
	k8stunnel "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tunnelhandler/k8s"
	mesostunnel "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tunnelhandler/mesos"
	mesoswebconsole "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tunnelhandler/mesoswebconsole"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	restful "github.com/emicklei/go-restful"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	microsvc "github.com/micro/go-micro/v2/service"
	microgrpcsvc "github.com/micro/go-micro/v2/service/grpc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	grpccred "google.golang.org/grpc/credentials"
)

// ClusterManager cluster manager
type ClusterManager struct {
	// options for cluster manager
	opt *options.ClusterManagerOptions

	// mongo DB options
	mongoOptions *mongo.Options

	// tls config for cluster manager server
	tlsConfig *tls.Config

	// tls config for cluster manager as client role
	clientTLSConfig *tls.Config

	// server handler
	serverHandler *handler.ClusterManager

	// http mux
	mux *http.ServeMux

	// http server
	httpServer *http.Server

	// extra module server, [pprof, metrics, swagger]
	extraServer *http.Server

	// discovery
	disc *discovery.ModuleDiscovery

	// IAM client
	iamClient iam.PermClient

	// locker
	locker lock.DistributedLock

	// tunnel peer manager
	tunnelPeerManager *tunnel.PeerManager

	// service registry
	microRegistry registry.Registry

	// micro service
	microService microsvc.Service

	model store.ClusterManagerModel

	k8sops *clusterops.K8SOperator

	// tke handler
	tkeHandler *tkehandler.Handler

	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	stopCh        chan struct{}
}

// NewClusterManager create cluster manager
func NewClusterManager(opt *options.ClusterManagerOptions) *ClusterManager {
	ctx, cancel := context.WithCancel(context.Background())
	options.SetGlobalCMOptions(opt)
	return &ClusterManager{
		opt:           opt,
		ctx:           ctx,
		ctxCancelFunc: cancel,
		stopCh:        make(chan struct{}),
	}
}

// init server and client tls config
func (cm *ClusterManager) initTLSConfig() error {
	if len(cm.opt.ServerCert) != 0 && len(cm.opt.ServerKey) != 0 && len(cm.opt.ServerCa) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(cm.opt.ServerCa, cm.opt.ServerCert,
			cm.opt.ServerKey, static.ServerCertPwd)
		if err != nil {
			blog.Errorf("load cluster manager server tls config failed, err %s", err.Error())
			return err
		}
		cm.tlsConfig = tlsConfig
		blog.Infof("load cluster manager server tls config successfully")
	}

	if len(cm.opt.ClientCert) != 0 && len(cm.opt.ClientKey) != 0 && len(cm.opt.ClientCa) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(cm.opt.ClientCa, cm.opt.ClientCert,
			cm.opt.ClientKey, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("load cluster manager client tls config failed, err %s", err.Error())
			return err
		}
		cm.clientTLSConfig = tlsConfig
		blog.Infof("load cluster manager client tls config successfully")
	}
	return nil
}

// init lock
func (cm *ClusterManager) initLocker() error {
	etcdEndpoints := utils.SplitAddrString(cm.opt.Etcd.EtcdEndpoints)
	var opts []lock.Option
	opts = append(opts, lock.Endpoints(etcdEndpoints...))
	opts = append(opts, lock.Prefix("clustermanager"))
	var etcdTLS *tls.Config
	var err error
	if len(cm.opt.Etcd.EtcdCa) != 0 && len(cm.opt.Etcd.EtcdCert) != 0 && len(cm.opt.Etcd.EtcdKey) != 0 {
		etcdTLS, err = ssl.ClientTslConfVerity(cm.opt.Etcd.EtcdCa, cm.opt.Etcd.EtcdCert, cm.opt.Etcd.EtcdKey, "")
		if err != nil {
			return err
		}
	}
	if etcdTLS != nil {
		opts = append(opts, lock.TLS(etcdTLS))
	}
	locker, err := etcdlock.New(opts...)
	if err != nil {
		blog.Errorf("init locker failed, err %s", err.Error())
		return err
	}
	blog.Infof("init locker successfullly")
	cm.locker = locker
	return nil
}

// init mongo client
func (cm *ClusterManager) initModel() error {
	if len(cm.opt.Mongo.Address) == 0 {
		return fmt.Errorf("mongo address cannot be empty")
	}
	if len(cm.opt.Mongo.Database) == 0 {
		return fmt.Errorf("mongo database cannot be empty")
	}
	password := cm.opt.Mongo.Password
	if password != "" {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
		password = string(realPwd)
	}
	mongoOptions := &mongo.Options{
		Hosts:                 strings.Split(cm.opt.Mongo.Address, ","),
		ConnectTimeoutSeconds: int(cm.opt.Mongo.ConnectTimeout),
		Database:              cm.opt.Mongo.Database,
		Username:              cm.opt.Mongo.Username,
		Password:              password,
		MaxPoolSize:           uint64(cm.opt.Mongo.MaxPoolSize),
		MinPoolSize:           uint64(cm.opt.Mongo.MinPoolSize),
	}
	cm.mongoOptions = mongoOptions

	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		blog.Errorf("init mongo db failed, err %s", err.Error())
		return err
	}
	if err = mongoDB.Ping(); err != nil {
		blog.Errorf("ping mongo db failed, err %s", err.Error())
		return err
	}
	blog.Infof("init mongo db successfully")
	modelSet := store.NewModelSet(mongoDB)
	cm.model = modelSet
	blog.Infof("init store successfully")
	return nil
}

// init task server
func (cm *ClusterManager) initTaskServer() error {
	cloudprovider.InitStorageModel(cm.model)
	//get taskserver and init
	taskMgr := taskserver.GetTaskServer()

	if err := taskMgr.Init(&cm.opt.Broker, cm.mongoOptions); err != nil {
		blog.Errorf("cluster-manager init task server failed, %s", err.Error())
		return err
	}
	blog.Infof("cluster-manager init task server successfully")
	return nil
}

// init remote client for cloud dependent data client, client may be disable or empty
func (cm *ClusterManager) initRemoteClient() error {
	// init tags client
	err := cmdb.SetCmdbClient(cmdb.Options{
		Enable:     cm.opt.Cmdb.Enable,
		AppCode:    cm.opt.Cmdb.AppCode,
		BKUserName: cm.opt.Cmdb.BkUserName,
		AppSecret:  cm.opt.Cmdb.AppSecret,
		Server:     cm.opt.Cmdb.Server,
		Debug:      cm.opt.Cmdb.Debug,
	})
	if err != nil {
		return err
	}
	// init ssm client
	err = auth.SetSSMClient(auth.Options{
		Server:    cm.opt.Ssm.Server,
		AppCode:   cm.opt.Ssm.AppCode,
		AppSecret: cm.opt.Ssm.AppSecret,
		Enable:    cm.opt.Ssm.Enable,
		Debug:     cm.opt.Ssm.Debug,
	})
	if err != nil {
		return err
	}

	// init pass-cc client
	err = passcc.SetCCClient(passcc.Options{
		Server:    cm.opt.Passcc.Server,
		AppCode:   cm.opt.BKOps.AppCode,
		AppSecret: cm.opt.BKOps.AppSecret,
		Enable:    cm.opt.Passcc.Enable,
		Debug:     cm.opt.Passcc.Debug,
	})
	if err != nil {
		return err
	}

	// init user-manager config
	user.SetUserManagerClient(&user.Options{
		Enable:          cm.opt.UserManager.Enable,
		GateWay:         cm.opt.UserManager.GateWay,
		IsVerifyTLS:     cm.opt.UserManager.IsVerifyTLS,
		Token:           cm.opt.UserManager.Token,
		EtcdRegistry:    cm.microRegistry,
		ClientTLSConfig: cm.clientTLSConfig,
	})

	return nil
}

// init bk-ops client
func (cm *ClusterManager) initBKOpsClient() error {
	err := common.SetBKOpsClient(common.Options{
		AppCode:       cm.opt.BKOps.AppCode,
		AppSecret:     cm.opt.BKOps.AppSecret,
		Debug:         cm.opt.BKOps.Debug,
		External:      cm.opt.BKOps.External,
		CreateTaskURL: cm.opt.BKOps.CreateTaskURL,
		TaskStatusURL: cm.opt.BKOps.TaskStatusURL,
		StartTaskURL:  cm.opt.BKOps.StartTaskURL,
	})
	if err != nil {
		blog.Errorf("initBKOpsClient failed: %v", err)
		return err
	}

	return nil
}

// init iam client for perm
func (cm *ClusterManager) initIAMClient() error {
	var err error
	cm.iamClient, err = iam.NewIamClient(&iam.Options{
		SystemID:    cm.opt.IAM.SystemID,
		AppCode:     cm.opt.IAM.AppCode,
		AppSecret:   cm.opt.IAM.AppSecret,
		External:    cm.opt.IAM.External,
		GateWayHost: cm.opt.IAM.GatewayServer,
		IAMHost:     cm.opt.IAM.IAMServer,
		BkiIAMHost:  cm.opt.IAM.BkiIAMServer,
		Metric:      cm.opt.IAM.Metric,
		Debug:       cm.opt.IAM.Debug,
	})

	if err != nil {
		return err
	}

	return nil
}

func (cm *ClusterManager) initCloudTemplateConfig() error {
	if cm.opt.CloudTemplatePath == "" {
		return fmt.Errorf("cloud template path empty, please manual build cloud")
	}

	cloudList := &options.CloudTemplateList{}
	cloudBytes, err := ioutil.ReadFile(cm.opt.CloudTemplatePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(cloudBytes, cloudList)
	if err != nil {
		blog.Errorf("initCloudTemplateConfig Unmarshal err: %v", err)
		return err
	}

	// init cloud config
	for i := range cloudList.CloudList {
		err = cm.updateCloudConfig(cloudList.CloudList[i])
		if err != nil {
			blog.Errorf("initCloudTemplateConfig[%s] failed %v", cloudList.CloudList[i].CloudID, err)
		}
	}

	return nil
}

func (cm *ClusterManager) updateCloudConfig(cloud *cmproto.Cloud) error {
	timeStr := time.Now().Format(time.RFC3339)
	cloud.UpdateTime = timeStr

	destCloud, err := cm.model.GetCloud(cm.ctx, cloud.CloudID)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		return err
	}

	// generate new cloud config
	if destCloud == nil {
		cloud.CreatTime = timeStr
		err = cm.model.CreateCloud(cm.ctx, cloud)
		if err != nil {
			return err
		}

		return nil
	}

	// update existed cloud config
	destCloud.UpdateTime = timeStr
	destCloud.Editable = cloud.Editable
	if cloud.Name != "" {
		destCloud.Name = cloud.Name
	}
	if cloud.OpsPlugins != nil {
		destCloud.OpsPlugins = cloud.OpsPlugins
	}
	if cloud.ExtraPlugins != nil {
		destCloud.ExtraPlugins = cloud.ExtraPlugins
	}
	if cloud.CloudCredential != nil {
		destCloud.CloudCredential = cloud.CloudCredential
	}
	if cloud.OsManagement != nil {
		destCloud.OsManagement = cloud.OsManagement
	}
	if cloud.ClusterManagement != nil {
		destCloud.ClusterManagement = cloud.ClusterManagement
	}
	if cloud.NodeGroupManagement != nil {
		destCloud.NodeGroupManagement = cloud.NodeGroupManagement
	}
	if len(cloud.CloudProvider) > 0 {
		destCloud.CloudProvider = cloud.CloudProvider
	}
	if len(cloud.Config) > 0 {
		destCloud.Config = cloud.Config
	}
	if len(cloud.Updater) > 0 {
		destCloud.Updater = cloud.Updater
	}
	if len(cloud.EngineType) > 0 {
		destCloud.EngineType = cloud.EngineType
	}
	if len(cloud.Enable) > 0 {
		destCloud.Enable = cloud.Enable
	}
	if len(cloud.Description) > 0 {
		destCloud.Description = cloud.Description
	}
	if cloud.NetworkInfo != nil {
		destCloud.NetworkInfo = cloud.NetworkInfo
	}
	if cloud.ConfInfo != nil {
		destCloud.ConfInfo = cloud.ConfInfo
	}
	if cloud.PlatformInfo != nil {
		destCloud.PlatformInfo = cloud.PlatformInfo
	}

	err = cm.model.UpdateCloud(cm.ctx, destCloud)
	if err != nil {
		return err
	}

	return nil
}

// init k8s operator
func (cm *ClusterManager) initK8SOperator() {
	cm.k8sops = clusterops.NewK8SOperator(cm.opt, cm.model)
	blog.Infof("init k8s cluster operator successfully")
}

func (cm *ClusterManager) initRegistry() error {
	etcdEndpoints := utils.SplitAddrString(cm.opt.Etcd.EtcdEndpoints)
	etcdSecure := false
	var etcdTLS *tls.Config
	var err error
	if len(cm.opt.Etcd.EtcdCa) != 0 && len(cm.opt.Etcd.EtcdCert) != 0 && len(cm.opt.Etcd.EtcdKey) != 0 {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(cm.opt.Etcd.EtcdCa, cm.opt.Etcd.EtcdCert, cm.opt.Etcd.EtcdKey, "")
		if err != nil {
			return err
		}
	}
	cm.microRegistry = etcd.NewRegistry(
		registry.Addrs(etcdEndpoints...),
		registry.Secure(etcdSecure),
		registry.TLSConfig(etcdTLS),
	)
	if err := cm.microRegistry.Init(); err != nil {
		return err
	}
	return nil
}

func (cm *ClusterManager) initDiscovery() {
	cm.disc = discovery.NewModuleDiscovery(cmcommon.ClusterManagerServiceDomain, cm.microRegistry)
	blog.Infof("init discovery for cluster manager successfully")
}

func (cm *ClusterManager) initTkeHandler(router *mux.Router) error {
	tkeHandler := tkehandler.NewTkeHandler(cm.model, cm.locker)
	cm.tkeHandler = tkeHandler

	tkeContainer := restful.NewContainer()
	tkeHandlerURL := "/clustermanager/v1/tke/cidr/{uri:.*}"
	tkeWebService := new(restful.WebService).
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	tkeWebService.Route(tkeWebService.POST("/clustermanager/v1/tke/cidr/add_cidr").To(tkeHandler.AddTkeCidr))
	tkeWebService.Route(tkeWebService.POST("/clustermanager/v1/tke/cidr/apply_cidr").To(tkeHandler.ApplyTkeCidr))
	tkeWebService.Route(tkeWebService.POST("/clustermanager/v1/tke/cidr/release_cidr").To(tkeHandler.ReleaseTkeCidr))
	tkeWebService.Route(tkeWebService.POST("/clustermanager/v1/tke/cidr/list_count").To(tkeHandler.ListTkeCidrCount))
	tkeContainer.Add(tkeWebService)
	router.Handle(tkeHandlerURL, tkeContainer)
	blog.Infof("register tke handler to path %s", tkeHandlerURL)
	return nil
}

func (cm *ClusterManager) initTunnelServer(router *mux.Router) error {
	tunnelServerCallback := tunnel.NewWsTunnelServerCallback(cm.model)
	cm.tunnelPeerManager = tunnel.NewPeerManager(
		cm.opt, cm.clientTLSConfig, tunnelServerCallback.GetTunnelServer(), cm.disc)
	if err := cm.tunnelPeerManager.Start(); err != nil {
		return err
	}
	tunnelProxyDispatcher := k8stunnel.NewTunnelProxyDispatcher(
		"cluster_id", "sub_path", cm.model, tunnelServerCallback.GetTunnelServer())

	// register tunnel server handler
	tunnelServerURL := "/clustermanager/v1/websocket/connect"
	router.Handle(tunnelServerURL, tunnelServerCallback.GetTunnelServer())
	blog.Infof("register tunnel server handler to path %s", tunnelServerURL)

	// register websocket tunnel cluster request handler
	clusterTunnelURL := "/clustermanager/clusters/{cluster_id}/{sub_path:.*}"
	router.Handle(clusterTunnelURL, tunnelProxyDispatcher)
	blog.Infof("register cluster tunnel handler to path %s", clusterTunnelURL)

	// init mesos tunnel interface
	mesosTunnelHandlerDispatcher := mesostunnel.NewWsTunnelDispatcher(cm.model, tunnelServerCallback.GetTunnelServer())
	mesosTunnelHander := mesostunnel.NewTunnelHandler(cm.clientTLSConfig, mesosTunnelHandlerDispatcher)
	mesosTunnelContainer := restful.NewContainer()
	mesosTunnelWebService := new(restful.WebService)
	mesosTunnelWebService.Route(mesosTunnelWebService.GET("{uri:*}").
		To(mesosTunnelHander.HandleGetActions))
	mesosTunnelWebService.Route(mesosTunnelWebService.POST("{uri:*}").
		To(mesosTunnelHander.HandlePostActions))
	mesosTunnelWebService.Route(mesosTunnelWebService.PUT("{uri:*}").
		To(mesosTunnelHander.HandlePutActions))
	mesosTunnelWebService.Route(mesosTunnelWebService.DELETE("{uri:*}").
		To(mesosTunnelHander.HandleDeleteActions))
	mesosTunnelContainer.Add(mesosTunnelWebService)
	mesosTunnelURL := "/clustermanager/mesosdriver/v4/{uri:.*}"
	router.Handle(mesosTunnelURL, mesosTunnelContainer)
	blog.Infof("register mesos tunnel handler to path %s", mesosTunnelURL)

	// init mesos websocket webconsole tunnel
	mesosWebconsoleURL := "/mesosdriver/v4/webconsole/{sub_path:.*}"
	router.Handle(mesosWebconsoleURL, mesoswebconsole.NewWebconsoleProxy(
		cm.clientTLSConfig, cm.model, tunnelServerCallback.GetTunnelServer()))
	blog.Infof("register mesos webconsole handler to path %s", mesosWebconsoleURL)
	return nil
}

// CustomMatcher for http header
func CustomMatcher(key string) (string, bool) {
	switch key {
	case "X-Request-Id":
		return "X-Request-Id", true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// init http grpc gateway
func (cm *ClusterManager) initHTTPGateway(router *mux.Router) error {
	gwmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(CustomMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			OrigName:     true,
			EmitDefaults: true,
		}),
	)
	grpcDialOpts := []grpc.DialOption{}
	if cm.tlsConfig != nil && cm.clientTLSConfig != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(grpccred.NewTLS(cm.clientTLSConfig)))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}
	err := cmproto.RegisterClusterManagerGwFromEndpoint(
		context.TODO(),
		gwmux,
		cm.opt.Address+":"+strconv.Itoa(int(cm.opt.Port)),
		grpcDialOpts)
	if err != nil {
		blog.Errorf("register http gateway failed, err %s", err.Error())
		return fmt.Errorf("register http gateway failed, err %s", err.Error())
	}
	router.Handle("/{uri:.*}", gwmux)
	blog.Info("register grpc gateway handler to path /")
	return nil
}

func (cm *ClusterManager) initHTTPService() error {
	router := mux.NewRouter()
	// init tke cidr handler
	if err := cm.initTkeHandler(router); err != nil {
		return err
	}
	// init tunnel server
	if err := cm.initTunnelServer(router); err != nil {
		return err
	}
	// init micro http gateway
	if err := cm.initHTTPGateway(router); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", router)
	cm.initSwagger(mux)

	httpAddr := cm.opt.Address + ":" + strconv.Itoa(int(cm.opt.HTTPPort))
	cm.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}
	go func() {
		var err error
		blog.Infof("start http gateway server on address %s", httpAddr)
		if cm.tlsConfig != nil {
			cm.httpServer.TLSConfig = cm.tlsConfig
			err = cm.httpServer.ListenAndServeTLS("", "")
		} else {
			err = cm.httpServer.ListenAndServe()
		}
		if err != nil {
			blog.Errorf("start http gateway server failed, err %s", err.Error())
			cm.stopCh <- struct{}{}
		}
	}()

	return nil
}

func (cm *ClusterManager) initPProf(mux *http.ServeMux) {
	if !cm.opt.Debug {
		blog.Infof("pprof is disabled")
		return
	}
	blog.Infof("pprof is enabled")
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func (cm *ClusterManager) initSwagger(mux *http.ServeMux) {
	if len(cm.opt.Swagger.Dir) != 0 {
		blog.Infof("swagger doc is enabled")
		mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(cm.opt.Swagger.Dir, strings.TrimPrefix(r.URL.Path, "/swagger/")))
		})
	}
}

func (cm *ClusterManager) initMetric(mux *http.ServeMux) {
	blog.Infof("init metric handler")
	mux.Handle("/metrics", promhttp.Handler())
}

func (cm *ClusterManager) initExtraModules() {
	extraMux := http.NewServeMux()
	cm.initPProf(extraMux)
	cm.initMetric(extraMux)
	extraServerEndpoint := cm.opt.Address + ":" + strconv.Itoa(int(cm.opt.MetricPort))
	cm.extraServer = &http.Server{
		Addr:    extraServerEndpoint,
		Handler: extraMux,
	}

	go func() {
		var err error
		blog.Infof("start extra modules [pprof, metric] server %s", extraServerEndpoint)
		err = cm.extraServer.ListenAndServe()
		if err != nil {
			blog.Errorf("extra modules server listen failed, err %s", err.Error())
			cm.stopCh <- struct{}{}
		}
	}()
}

func (cm *ClusterManager) initMicro() error {
	// New Service
	microService := microgrpcsvc.NewService(
		microsvc.Name(cmcommon.ClusterManagerServiceDomain),
		microsvc.Metadata(map[string]string{
			cmcommon.MicroMetaKeyHTTPPort: strconv.Itoa(int(cm.opt.HTTPPort)),
		}),
		microgrpcsvc.WithTLS(cm.tlsConfig),
		microsvc.Address(cm.opt.Address+":"+strconv.Itoa(int(cm.opt.Port))),
		microsvc.Registry(cm.microRegistry),
		microsvc.Version(version.BcsVersion),
		microsvc.RegisterTTL(30*time.Second),
		microsvc.RegisterInterval(25*time.Second),
		microsvc.Context(cm.ctx),
		microsvc.BeforeStart(func() error {
			return nil
		}),
		microsvc.AfterStart(func() error {
			return cm.disc.Start()
		}),
		microsvc.BeforeStop(func() error {
			cm.disc.Stop()
			return nil
		}),
	)
	microService.Init()

	// create cluster manager server handler
	cm.serverHandler = handler.NewClusterManager(&handler.ControllerOptions{
		Model:      cm.model,
		KubeClient: cm.k8sops,
		Locker:     cm.locker,
		IAMClient:  cm.iamClient,
	})
	// Register handler
	cmproto.RegisterClusterManagerHandler(microService.Server(), cm.serverHandler)
	cm.microService = microService
	return nil
}

func (cm *ClusterManager) initSignalHandler() {
	// listen system signal
	// to run in the container, should not trap SIGTERM
	interupt := make(chan os.Signal, 10)
	signal.Notify(interupt, syscall.SIGINT)
	go func() {
		select {
		case e := <-interupt:
			blog.Infof("receive interupt %s, do close", e.String())
			cm.close()
		case <-cm.stopCh:
			blog.Infof("stop channel, do close")
			cm.close()
		}
	}()
}

func (cm *ClusterManager) close() {
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closeCancel()
	cm.extraServer.Shutdown(closeCtx)
	cm.httpServer.Shutdown(closeCtx)
	cm.ctxCancelFunc()
}

// Init init cluster manager server
func (cm *ClusterManager) Init() error {
	// init server and client tls config
	if err := cm.initTLSConfig(); err != nil {
		return err
	}
	// init locker
	if err := cm.initLocker(); err != nil {
		return err
	}
	// init model
	if err := cm.initModel(); err != nil {
		return err
	}
	// init kube operator
	cm.initK8SOperator()
	// init registry
	if err := cm.initRegistry(); err != nil {
		return err
	}
	// init IAM client
	if err := cm.initIAMClient(); err != nil {
		return err
	}

	// init core micro service
	if err := cm.initMicro(); err != nil {
		return err
	}

	// init discovery
	cm.initDiscovery()
	// init http service
	if err := cm.initHTTPService(); err != nil {
		return err
	}
	// init task server
	err := cm.initTaskServer()
	if err != nil {
		return err
	}
	// init bk-ops client
	err = cm.initBKOpsClient()
	if err != nil {
		return err
	}
	// init remote cloud depend client
	err = cm.initRemoteClient()
	if err != nil {
		return err
	}
	// init cloud template config
	err = cm.initCloudTemplateConfig()
	if err != nil {
		blog.Errorf("initCloudTemplateConfig failed: %v", err)
	}

	// init metric, pprof
	cm.initExtraModules()
	// init system signal handler
	cm.initSignalHandler()

	return nil
}

// Run run cluster manager server
func (cm *ClusterManager) Run() error {
	// run the service
	if err := cm.microService.Run(); err != nil {
		blog.Fatal(err)
	}
	blog.CloseLogs()
	return nil
}
