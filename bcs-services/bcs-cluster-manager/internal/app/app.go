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

// Package app xxx
package app

import (
	"context"
	"crypto/tls"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net"
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
	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	commonutil "github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/micro"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	restful "github.com/emicklei/go-restful"
	"github.com/go-micro/plugins/v4/registry/etcd"
	microgrpcserver "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	microsvc "go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"google.golang.org/grpc"
	grpccred "google.golang.org/grpc/credentials"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	i18n2 "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	clusterops "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	cmcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/commonhandler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/daemon"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	etcdlock "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock/etcd"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/alarm/bkmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/alarm/tmp"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/audit"
	ssmAuth "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/gse"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/addons"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/helm"
	installTypes "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/job"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/nodeman"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/passcc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	resource "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/user"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tkehandler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tunnel"
	k8stunnel "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tunnelhandler/k8s"
	mesostunnel "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tunnelhandler/mesos"
	mesoswebconsole "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tunnelhandler/mesoswebconsole"
	itypes "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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
	mux *http.ServeMux // nolint

	// http server
	httpServer *ipv6server.IPv6Server

	// extra module server, [pprof, metrics, swagger]
	extraServer *ipv6server.IPv6Server

	// discovery
	disc *discovery.ModuleDiscovery

	// resource discovery
	resourceDisc *discovery.ModuleDiscovery

	// project discovery
	projectDisc *discovery.ModuleDiscovery

	// cidr discovery
	cidrDisc *discovery.ModuleDiscovery

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

	// model store
	model store.ClusterManagerModel

	// etcd store
	etcdModel store.EtcdStoreInterface

	// k8s cluster operator
	k8sops *clusterops.K8SOperator

	// tke handler
	tkeHandler *tkehandler.Handler

	// daemon process
	daemon daemon.DaemonInterface

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
	// client tls config
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

	// server tls config
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
func (cm *ClusterManager) initEtcdLockerStore() error {
	etcdEndpoints := utils.SplitAddrString(cm.opt.Etcd.EtcdEndpoints)
	var opts []itypes.Option
	opts = append(opts, itypes.Endpoints(etcdEndpoints...))
	opts = append(opts, itypes.Prefix("clustermanager"))
	var etcdTLS *tls.Config
	var err error
	if len(cm.opt.Etcd.EtcdCa) != 0 && len(cm.opt.Etcd.EtcdCert) != 0 && len(cm.opt.Etcd.EtcdKey) != 0 {
		etcdTLS, err = ssl.ClientTslConfVerity(cm.opt.Etcd.EtcdCa, cm.opt.Etcd.EtcdCert, cm.opt.Etcd.EtcdKey, "")
		if err != nil {
			return err
		}
	}
	if etcdTLS != nil {
		opts = append(opts, itypes.TLS(etcdTLS))
	}

	// register etcd distributed lock
	locker, err := etcdlock.New(opts...)
	if err != nil {
		blog.Errorf("init locker failed, err %s", err.Error())
		return err
	}
	blog.Infof("init locker successfully")

	// init etcd store client
	etcdClient, err := store.NewModelEtcd(opts...)
	if err != nil {
		blog.Errorf("init etcd store client failed: %v", err.Error())
		return err
	}
	blog.Infof("init etcdClient successfully")

	cm.etcdModel = etcdClient
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

	modelSet, err := store.NewModelSet(mongoOptions)
	if err != nil {
		return err
	}
	cm.model = modelSet

	blog.Infof("init store successfully")
	return nil
}

// init task server
func (cm *ClusterManager) initTaskServer() error {
	cloudprovider.InitStorageModel(cm.model)
	cloudprovider.InitEtcdModel(cm.etcdModel)
	cloudprovider.InitDistributeLock(cm.locker)
	// get taskserver and init
	taskMgr := taskserver.GetTaskServer()

	if err := taskMgr.Init(&cm.opt.Broker, cm.mongoOptions); err != nil {
		blog.Errorf("cluster-manager init task server failed, %s", err.Error())
		return err
	}
	blog.Infof("cluster-manager init task server successfully")
	return nil
}

// init remote client for cloud dependent data client, client may be disable or empty
func (cm *ClusterManager) initRemoteClient() error { // nolint
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
	// init perm client
	err = ssmAuth.SetAccessClient(ssmAuth.Options{
		Server: cm.opt.Access.Server,
		Debug:  cm.opt.Access.Debug,
	})
	if err != nil {
		return err
	}

	// init pass-cc client
	err = passcc.SetCCClient(passcc.Options{
		Server:    cm.opt.Passcc.Server,
		AppCode:   cm.opt.Passcc.AppCode,
		AppSecret: cm.opt.Passcc.AppSecret,
		Enable:    cm.opt.Passcc.Enable,
		Debug:     cm.opt.Passcc.Debug,
	})
	if err != nil {
		return err
	}

	// init alarm client
	err = cm.initAlarmClient()
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

	// init nodeman client
	err = nodeman.SetNodeManClient(nodeman.Options{
		Enable:     cm.opt.NodeMan.Enable,
		AppCode:    cm.opt.NodeMan.AppCode,
		BKUserName: cm.opt.NodeMan.BkUserName,
		AppSecret:  cm.opt.NodeMan.AppSecret,
		Server:     cm.opt.NodeMan.Server,
		Debug:      cm.opt.NodeMan.Debug,
	})
	if err != nil {
		return err
	}

	// init gse client
	if err = gse.SetGseClient(gse.Options{
		Enable:        cm.opt.Gse.Enable,
		AppCode:       cm.opt.Gse.AppCode,
		AppSecret:     cm.opt.Gse.AppSecret,
		BKUserName:    cm.opt.Gse.BkUserName,
		EsbServer:     cm.opt.Gse.EsbServer,
		GatewayServer: cm.opt.Gse.GatewayServer,
		Debug:         cm.opt.Gse.Debug,
	}); err != nil {
		return err
	}

	// init job client
	if err = job.SetJobClient(job.Options{
		AppCode:    cm.opt.Job.AppCode,
		AppSecret:  cm.opt.Job.AppSecret,
		BKUserName: cm.opt.Job.BkUserName,
		Server:     cm.opt.Job.Server,
		Debug:      cm.opt.Job.Debug,
	}); err != nil {
		return err
	}

	// init helm client
	err = helm.SetHelmManagerClient(&installTypes.Options{
		Enable:          cm.opt.Helm.Enable,
		GateWay:         cm.opt.Helm.GateWay,
		Token:           cm.opt.Helm.Token,
		Module:          cm.opt.Helm.Module,
		EtcdRegistry:    cm.microRegistry,
		ClientTLSConfig: cm.clientTLSConfig,
	})
	if err != nil {
		return err
	}
	// init addons client
	err = addons.SetAddonsClient(&installTypes.Options{
		Enable:          cm.opt.Helm.Enable,
		GateWay:         cm.opt.Helm.GateWay,
		Token:           cm.opt.Helm.Token,
		Module:          cm.opt.Helm.Module,
		EtcdRegistry:    cm.microRegistry,
		ClientTLSConfig: cm.clientTLSConfig,
	})
	if err != nil {
		return err
	}

	// init encrypt client
	err = encrypt.SetEncryptClient(cm.opt.Encrypt)
	if err != nil {
		return err
	}

	return nil
}

// init different alarm client
func (cm *ClusterManager) initAlarmClient() error {
	// init alarm client
	err := tmp.SetBKAlarmClient(tmp.Options{
		AppCode:   cm.opt.Alarm.AppCode,
		AppSecret: cm.opt.Alarm.AppSecret,
		Enable:    cm.opt.Alarm.Enable,
		Server:    cm.opt.Alarm.Server,
		Debug:     cm.opt.Alarm.Debug,
	})
	if err != nil {
		return err
	}

	// bkmonitor client
	err = bkmonitor.SetMonitorClient(bkmonitor.Options{
		AppCode:   cm.opt.Alarm.AppCode,
		AppSecret: cm.opt.Alarm.AppSecret,
		Enable:    cm.opt.Alarm.Enable,
		Server:    cm.opt.Alarm.MonitorServer,
		Debug:     cm.opt.Alarm.Debug,
	})
	if err != nil {
		return err
	}

	return nil
}

// init bk-ops client
func (cm *ClusterManager) initBKOpsClient() error {
	err := common.SetBKOpsClient(common.Options{
		EsbServer:  cm.opt.BKOps.EsbServer,
		Server:     cm.opt.BKOps.Server,
		AppCode:    cm.opt.BKOps.AppCode,
		AppSecret:  cm.opt.BKOps.AppSecret,
		BKUserName: cm.opt.BKOps.BkUserName,
		Debug:      cm.opt.BKOps.Debug,
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

	// init perm client
	auth.InitPermClient(cm.iamClient)

	return nil
}

// init jwt client for perm
func (cm *ClusterManager) initJWTClient() error {
	return auth.InitJWTClient(cm.opt)
}

// initCache for cache init
func (cm *ClusterManager) initCache() error {
	cache.InitCache()
	return nil
}

// init client permissions
func (cm *ClusterManager) initClientPermissions() error {
	auth.ClientPermissions = make(map[string][]string, 0)
	if len(cm.opt.Auth.ClientPermissions) == 0 {
		return nil
	}

	err := json.Unmarshal([]byte(cm.opt.Auth.ClientPermissions), &auth.ClientPermissions)
	if err != nil {
		return fmt.Errorf("parse ClientPermissions error: %s", err.Error())
	}
	return nil
}

// init no auth method
func (cm *ClusterManager) initNoAuthMethod() error {
	if len(cm.opt.Auth.NoAuthMethod) == 0 {
		return nil
	}

	methods := strings.Split(cm.opt.Auth.NoAuthMethod, ",")
	auth.NoAuthMethod = append(auth.NoAuthMethod, methods...)
	return nil
}

// initCloudTemplateConfig cloud template config
func (cm *ClusterManager) initCloudTemplateConfig() error {
	if cm.opt.CloudTemplatePath == "" {
		return fmt.Errorf("cloud template path empty, please manual build cloud")
	}

	blog.Infof("initCloudTemplateConfig %s", cm.opt.CloudTemplatePath)

	cloudList := &options.CloudTemplateList{}
	cloudBytes, err := os.ReadFile(cm.opt.CloudTemplatePath)
	if err != nil {
		blog.Errorf("initCloudTemplateConfig readFile[%s] failed: %v", cm.opt.CloudTemplatePath, err)
		return err
	}

	err = json.Unmarshal(cloudBytes, cloudList)
	if err != nil {
		blog.Errorf("initCloudTemplateConfig Unmarshal err: %v", err)
		return err
	}

	blog.Infof("initCloudTemplateConfig cloudList %+v", cloudList)

	// init cloud config
	for i := range cloudList.CloudList {
		err = cm.updateCloudConfig(cloudList.CloudList[i])
		if err != nil {
			blog.Errorf("initCloudTemplateConfig[%s] failed %v", cloudList.CloudList[i].CloudID, err)
		}
	}

	return nil
}

// updateCloudConfig update cloud config template
func (cm *ClusterManager) updateCloudConfig(cloud *cmproto.Cloud) error {
	timeStr := time.Now().Format(time.RFC3339)
	cloud.UpdateTime = timeStr

	destCloud, err := cm.model.GetCloud(cm.ctx, cloud.CloudID)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) &&
		!errors.Is(err, util.ErrDecryptCloudCredential) {
		blog.Errorf("updateCloudConfig GetCloud[%s] failed: %v", cloud.CloudID, err)
		return err
	}

	// generate new cloud config
	if destCloud == nil {
		cloud.CreatTime = timeStr
		err = cm.model.CreateCloud(cm.ctx, cloud)
		if err != nil {
			blog.Errorf("updateCloudConfig CreateCloud[%s] failed: %v", cloud.CloudID, err)
			return err
		}

		blog.Infof("updateCloudConfig[%s] success", cloud.CloudID)
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
		blog.Errorf("updateCloudConfig UpdateCloud[%s] failed: %v", cloud.CloudID, err)
		return err
	}

	blog.Infof("updateCloudConfig[%s] success", cloud.CloudID)
	return nil
}

// init k8s operator
func (cm *ClusterManager) initK8SOperator() {
	cm.k8sops = clusterops.NewK8SOperator(cm.opt, cm.model)
	blog.Infof("init k8s cluster operator successfully")
}

// init daemon
func (cm *ClusterManager) initDaemon() {
	cm.daemon = daemon.NewDaemon(0, cm.model, cm.locker, daemon.DaemonOptions{
		EnableDaemon:             cm.opt.Daemon.Enable,
		EnableAllocateCidrDaemon: cm.opt.Daemon.EnableAllocateCidr,
		EnableInsTypeUsage:       cm.opt.Daemon.EnableInsTypeUsage,
	})
}

// initRegistry etcd registry
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
	if err = cm.microRegistry.Init(); err != nil {
		return err
	}
	return nil
}

// initDiscovery discovery client
func (cm *ClusterManager) initDiscovery() {
	cm.disc = discovery.NewModuleDiscovery(cmcommon.ClusterManagerServiceDomain, cm.microRegistry)
	blog.Infof("init discovery for cluster manager successfully")

	// enable discovery resource module
	if cm.opt.ResourceManager.Enable {
		cm.resourceDisc = discovery.NewModuleDiscovery(cm.opt.ResourceManager.Module, cm.microRegistry)
		blog.Infof("init discovery for resource manager successfully")

		resource.SetResourceClient(&resource.Options{
			Enable:    cm.opt.ResourceManager.Enable,
			Module:    cm.opt.ResourceManager.Module,
			TLSConfig: cm.clientTLSConfig,
		}, cm.resourceDisc)
	}

	// enable discovery project module
	if cm.opt.ProjectManager.Enable {
		cm.projectDisc = discovery.NewModuleDiscovery(cm.opt.ProjectManager.Module, cm.microRegistry)
		blog.Infof("init discovery for project manager successfully")

		project.SetProjectClient(&project.Options{
			Module:    cm.opt.ProjectManager.Module,
			TLSConfig: cm.clientTLSConfig,
		}, cm.projectDisc)
	}
}

// initTkeHandler tke cidr handler
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

// initTunnelServer init tunnel server
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
	case middleware.CustomUsernameHeaderKey:
		return middleware.CustomUsernameHeaderKey, true
	case middleware.InnerClientHeaderKey:
		return middleware.InnerClientHeaderKey, true
	case constants.Traceparent:
		return constants.GrpcTraceparent, true
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
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure()) // nolint
	}
	grpcDialOpts = append(grpcDialOpts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(utils.MaxBodySize), grpc.MaxCallSendMsgSize(utils.MaxBodySize)))

	err := cmproto.RegisterClusterManagerGwFromEndpoint(
		context.TODO(),
		gwmux,
		net.JoinHostPort(cm.opt.Address, strconv.Itoa(int(cm.opt.Port))),
		grpcDialOpts)
	if err != nil {
		blog.Errorf("register http gateway failed, err %s", err.Error())
		return fmt.Errorf("register http gateway failed, err %s", err.Error())
	}
	router.Handle("/{uri:.*}", gwmux)
	blog.Info("register grpc gateway handler to path /")
	return nil
}

// initCommonHandler common handler
func (cm *ClusterManager) initCommonHandler(router *mux.Router) error {
	commonHandler := commonhandler.NewCommonHandler(cm.model, cm.locker)
	commonContainer := restful.NewContainer()
	commonHandlerURL := "/clustermanager/v1/common/{uri:.*}"
	commonWebService := new(restful.WebService).
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	commonWebService.Route(commonWebService.GET("/clustermanager/v1/common/downloadtaskrecords").
		To(commonHandler.DownloadTaskRecords))
	commonContainer.Add(commonWebService)
	router.Handle(commonHandlerURL, commonContainer)
	blog.Infof("register common handler to path %s", commonHandlerURL)
	return nil
}

// initHTTPService init http service
func (cm *ClusterManager) initHTTPService() error {
	router := mux.NewRouter()
	// init tke cidr handler
	if err := cm.initTkeHandler(router); err != nil {
		return err
	}
	// init common http handler
	if err := cm.initCommonHandler(router); err != nil {
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

	muxServe := http.NewServeMux()
	muxServe.Handle("/", router)
	cm.initSwagger(muxServe)

	// server address
	addresses := []string{cm.opt.Address}
	if len(cm.opt.Ipv6Address) > 0 && (cm.opt.Ipv6Address != cm.opt.Address) {
		addresses = append(addresses, cm.opt.Ipv6Address)
	}
	cm.httpServer = ipv6server.NewIPv6Server(addresses, strconv.Itoa(int(cm.opt.HTTPPort)), "", muxServe)

	go func() {
		var err error
		blog.Infof("start http gateway server on address %+v", addresses)
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

// initPProf pprof
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

// initSwagger init swagger
func (cm *ClusterManager) initSwagger(mux *http.ServeMux) {
	if len(cm.opt.Swagger.Dir) != 0 {
		blog.Infof("swagger doc is enabled")
		mux.HandleFunc("/clustermanager/swagger/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(cm.opt.Swagger.Dir, strings.TrimPrefix(r.URL.Path, "/clustermanager/swagger/")))
		})
	}
}

// initMetric metric
func (cm *ClusterManager) initMetric(mux *http.ServeMux) {
	blog.Infof("init metric handler")
	mux.Handle("/metrics", promhttp.Handler())
}

func (cm *ClusterManager) initExtraModules() {
	extraMux := http.NewServeMux()
	cm.initPProf(extraMux)
	cm.initMetric(extraMux)

	ips := []string{cm.opt.Address}
	if len(cm.opt.Ipv6Address) > 0 && (cm.opt.Ipv6Address != cm.opt.Address) {
		ips = append(ips, cm.opt.Ipv6Address)
	}
	cm.extraServer = ipv6server.NewIPv6Server(ips, strconv.Itoa(int(cm.opt.MetricPort)), "", extraMux)

	go func() {
		var err error
		blog.Infof("start extra modules [pprof, metric] server %+v", ips)
		err = cm.extraServer.ListenAndServe()
		if err != nil {
			blog.Errorf("extra modules server listen failed, err %s", err.Error())
			cm.stopCh <- struct{}{}
		}
	}()
}

func (cm *ClusterManager) initMicro() error { // nolint
	// server listen ip
	ipv4 := cm.opt.Address
	ipv6 := cm.opt.Ipv6Address
	port := strconv.Itoa(int(cm.opt.Port))

	// service inject metadata to discovery center
	metadata := make(map[string]string)
	metadata[cmcommon.MicroMetaKeyHTTPPort] = strconv.Itoa(int(cm.opt.HTTPPort))

	// 适配单栈环境（ipv6注册地址不能是本地回环地址）
	if v := net.ParseIP(ipv6); v != nil && !v.IsLoopback() {
		metadata[types.IPV6] = net.JoinHostPort(ipv6, port)
	}

	authWrapper := middleware.NewGoMicroAuth(auth.GetJWTClient()).
		EnableSkipHandler(auth.SkipHandler).
		EnableSkipClient(auth.SkipClient).
		SetCheckUserPerm(auth.CheckUserPerm)

	// with tls
	grpcSvr := microgrpcserver.NewServer(microgrpcserver.AuthTLS(cm.tlsConfig))

	// New Service
	microService := microsvc.NewService(
		microsvc.Server(grpcSvr),
		microsvc.Cmd(commonutil.NewDummyMicroCmd()),
		microsvc.Name(cmcommon.ClusterManagerServiceDomain),
		microsvc.Metadata(metadata),
		microsvc.Address(net.JoinHostPort(ipv4, port)),
		microsvc.Registry(cm.microRegistry),
		microsvc.Version(version.BcsVersion),
		microsvc.RegisterTTL(30*time.Second),
		microsvc.RegisterInterval(25*time.Second),
		microsvc.Context(cm.ctx),
		utils.MaxMsgSize(utils.MaxBodySize),
		microsvc.BeforeStart(func() error {
			return nil
		}),
		microsvc.AfterStart(func() error {
			if cm.resourceDisc != nil {
				cm.resourceDisc.Start() // nolint
			}
			if cm.projectDisc != nil {
				cm.projectDisc.Start() // nolint
			}
			if cm.cidrDisc != nil {
				cm.cidrDisc.Start() // nolint
			}
			return cm.disc.Start()
		}),
		microsvc.BeforeStop(func() error {
			if cm.resourceDisc != nil {
				cm.resourceDisc.Stop()
			}
			if cm.projectDisc != nil {
				cm.projectDisc.Stop()
			}
			if cm.cidrDisc != nil {
				cm.cidrDisc.Stop()
			}
			cm.disc.Stop()
			return nil
		}),
		microsvc.AfterStop(func() error {
			audit.GetAuditClient().Close()
			return nil
		}),
		microsvc.WrapHandler(
			utils.HandleLanguageWrapper,
			utils.RequestLogWarpper,
			utils.ResponseWrapper,
			authWrapper.AuthenticationFunc,
			authWrapper.AuthorizationFunc,
			utils.NewAuditWrapper,
			micro.NewTracingWrapper(),
		),
	)
	microService.Init()

	// create cluster manager server handler
	cm.serverHandler = handler.NewClusterManager(&handler.ControllerOptions{
		Model:      cm.model,
		KubeClient: cm.k8sops,
		Locker:     cm.locker,
		IAMClient:  cm.iamClient,
	})
	// 创建双栈监听
	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListener(ipv4, port); err != nil { // 添加主地址监听
		return err
	}
	if ipv6 != ipv4 {
		err := dualStackListener.AddListener(ipv6, port) // 添加副地址监听
		if err != nil {
			return err
		}
	}

	// grpc server
	grpcServer := microService.Server()
	if err := grpcServer.Init(microgrpcserver.Listener(dualStackListener)); err != nil {
		return err
	}
	// Register handler
	cmproto.RegisterClusterManagerHandler(grpcServer, cm.serverHandler) // nolint
	cm.microService = microService
	return nil
}

func (cm *ClusterManager) initSignalHandler() {
	// listen system signal
	// to run in the container, should not trap SIGTERM
	interrupt := make(chan os.Signal, 10)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case e := <-interrupt:
			blog.Infof("receive interrupt %s, do close", e.String())
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
	helm.GetHelmManagerClient().Stop()
	addons.GetAddonsClient().Stop()
	cm.extraServer.Shutdown(closeCtx) // nolint
	cm.httpServer.Shutdown(closeCtx)  // nolint
	cm.daemon.Stop()
	cm.ctxCancelFunc()
}

// Init init cluster manager server
func (cm *ClusterManager) Init() error {
	// init server and client tls config
	if err := cm.initTLSConfig(); err != nil {
		return err
	}
	// init locker
	if err := cm.initEtcdLockerStore(); err != nil {
		return err
	}
	// init registry
	if err := cm.initRegistry(); err != nil {
		return err
	}
	// init remote cloud depend client
	if err := cm.initRemoteClient(); err != nil {
		return err
	}
	// init model
	if err := cm.initModel(); err != nil {
		return err
	}
	// init kube operator
	cm.initK8SOperator()
	// init IAM client
	if err := cm.initIAMClient(); err != nil {
		return err
	}
	// init cache
	if err := cm.initCache(); err != nil {
		return err
	}

	// init jwt client
	if err := cm.initJWTClient(); err != nil {
		return err
	}

	// init client permissions
	if err := cm.initClientPermissions(); err != nil {
		return err
	}

	// init no auth methods
	if err := cm.initNoAuthMethod(); err != nil {
		return err
	}

	// init core micro service
	if err := cm.initMicro(); err != nil {
		return err
	}
	// init daemon
	cm.initDaemon()
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
	// init cloud template config
	err = cm.initCloudTemplateConfig()
	if err != nil {
		blog.Errorf("initCloudTemplateConfig failed: %v", err)
	}

	// init metric, pprof
	cm.initExtraModules()
	// init system signal handler
	cm.initSignalHandler()
	// init i18n
	cm.initI18n()
	return nil
}

func (cm *ClusterManager) initI18n() {
	i18n.Instance()
	// 加载翻译文件路径
	i18n.SetPath([]embed.FS{i18n2.Assets})
	// 设置默认语言
	// 默认是 zh
	i18n.SetLanguage("zh")
}

// Run run cluster manager server
func (cm *ClusterManager) Run() error {
	// run daemon
	go cm.daemon.InitDaemon(cm.ctx)
	// run the service
	if err := cm.microService.Run(); err != nil {
		blog.Fatal(err)
	}
	blog.CloseLogs()
	return nil
}
