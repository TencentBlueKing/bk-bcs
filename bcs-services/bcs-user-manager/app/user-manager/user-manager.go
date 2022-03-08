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

package usermanager

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcshttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/cmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/permission"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"

	"github.com/emicklei/go-restful"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
)

var (
	// ErrHTTPServerNotInit server not init
	ErrHTTPServerNotInit = errors.New("UserManager server not init")
)

// UserManager http interface of user-manager
type UserManager struct {
	config   *config.UserMgrConfig
	httpServ *httpserver.HttpServer

	IamPermClient iam.PermClient
	EtcdRegistry  registry.Registry
	CmClient      *cmanager.ClusterManagerClient

	permService *permission.PermVerifyClient
}

// NewUserManager creates an UserManager object
func NewUserManager(conf *config.UserMgrConfig) *UserManager {
	userManager := &UserManager{
		config:   conf,
		httpServ: httpserver.NewHttpServer(conf.Port, conf.Address, conf.Sock),
	}

	if conf.ServCert.IsSSL {
		userManager.httpServ.SetSsl(conf.ServCert.CAFile, conf.ServCert.CertFile, conf.ServCert.KeyFile, conf.ServCert.CertPasswd)
	}

	userManager.httpServ.SetInsecureServer(conf.InsecureAddress, conf.InsecurePort)

	return userManager
}

// Start entry point for user-manager
func (u *UserManager) Start() error {
	// init redis
	if err := cache.InitRedis(u.config); err != nil {
		return err
	}

	if err := SetupStore(u.config); err != nil {
		return err
	}

	// init usermanager role and cache permission
	go permission.InitCache()
	time.Sleep(1 * time.Second)

	err := u.initUserManagerServer()
	if err != nil {
		blog.Errorf("initUserManagerServer failed: %v", err)
		return err
	}

	// usermanager api
	ws := u.httpServ.NewWebService("/usermanager", nil)
	u.initRouters(ws)

	router := u.httpServ.GetRouter()
	webContainer := u.httpServ.GetWebContainer()

	// handle user and cluster manager request
	router.Handle("/usermanager/{sub_path:.*}", webContainer)

	if err := u.httpServ.ListenAndServeMux(u.config.VerifyClientTLS); err != nil {
		return fmt.Errorf("http ListenAndServe error %s", err.Error())
	}

	return nil
}

// Filter authenticate the request
func Filter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// first authenticate the request, only admin user be allowed
	auth := utils.Authenticate(req.Request)
	if !auth {
		resp.WriteHeaderAndEntity(http.StatusUnauthorized, bcshttp.APIRespone{
			Result:  false,
			Code:    common.BcsErrApiUnauthorized,
			Message: "must provide admin token to request with websocket",
			Data:    nil,
		})
		return
	}

	chain.ProcessFilter(req, resp)
}

// initRouters init usermanager http router
func (u *UserManager) initRouters(ws *restful.WebService) {
	v1http.InitV1Routers(ws, u.permService)
	// register pull resource API
}

func (u *UserManager) initPermService() error {
	permService := permission.NewPermVerifyClient(u.config.PermissionSwitch, u.IamPermClient, u.CmClient)
	u.permService = permService

	return nil
}

func (u *UserManager) initClusterManager() error {
	opts := &cmanager.Options{
		Module:          u.config.ClusterConfig.Module,
		EtcdRegistry:    u.EtcdRegistry,
		ClientTLSConfig: u.config.TlsClientConfig,
	}

	cli := cmanager.NewClusterManagerClient(opts)
	if cli == nil {
		errMsg := fmt.Errorf("initClusterManager failed")
		return errMsg
	}

	u.CmClient = cli
	return nil
}

func (u *UserManager) initIamPermClient() error {

	opt := &iam.Options{
		SystemID:    u.config.IAMConfig.SystemID,
		AppCode:     u.config.IAMConfig.AppCode,
		AppSecret:   u.config.IAMConfig.AppSecret,
		External:    u.config.IAMConfig.External,
		GateWayHost: u.config.IAMConfig.GateWayHost,
		IAMHost:     u.config.IAMConfig.IAMHost,
		BkiIAMHost:  u.config.IAMConfig.BkiIAMHost,
		Metric:      u.config.IAMConfig.Metric,
	}
	iamCli, err := iam.NewIamClient(opt)
	if err != nil {
		blog.Errorf("initIamPermClient failed: %v", err)
		return err
	}

	u.IamPermClient = iamCli
	return nil
}

func (u *UserManager) initEtcdRegistry() error {
	if !u.config.EtcdConfig.Feature {
		return fmt.Errorf("etcd feature is off")
	}

	if len(u.config.EtcdConfig.Address) == 0 {
		errMsg := fmt.Errorf("etcdServers invalid")
		return errMsg
	}
	servers := strings.Split(u.config.EtcdConfig.Address, ";")

	var (
		secureEtcd bool
		etcdTLS    *tls.Config
		err        error
	)

	if len(u.config.EtcdConfig.CA) != 0 && len(u.config.EtcdConfig.Cert) != 0 && len(u.config.EtcdConfig.Key) != 0 {
		secureEtcd = true

		etcdTLS, err = ssl.ClientTslConfVerity(u.config.EtcdConfig.CA, u.config.EtcdConfig.Cert,
			u.config.EtcdConfig.Key, "")
		if err != nil {
			return err
		}
	}

	u.EtcdRegistry = etcd.NewRegistry(
		registry.Addrs(servers...),
		registry.Secure(secureEtcd),
		registry.TLSConfig(etcdTLS),
	)
	if err := u.EtcdRegistry.Init(); err != nil {
		return err
	}

	return nil
}

func (u *UserManager) initUserManagerServer() error {
	err := u.initEtcdRegistry()
	if err != nil {
		return err
	}

	err = u.initIamPermClient()
	if err != nil {
		return err
	}

	err = u.initClusterManager()
	if err != nil {
		return err
	}

	err = u.initPermService()
	if err != nil {
		return err
	}

	return nil
}
