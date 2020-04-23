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

package user_manager

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/http/httpserver"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"
	"bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http"
	"bk-bcs/bcs-services/bcs-user-manager/config"
	"github.com/emicklei/go-restful"
)

type UserManager struct {
	config   *config.UserMgrConfig
	httpServ *httpserver.HttpServer
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

func (u *UserManager) Start() error {
	err := SetupStore(u.config)
	if err != nil {
		return err
	}

	ws := u.httpServ.NewWebService("/usermanager", nil)
	u.initRouters(ws)

	router := u.httpServ.GetRouter()
	webContainer := u.httpServ.GetWebContainer()
	router.Handle("/usermanager/{sub_path:.*}", webContainer)
	if err := u.httpServ.ListenAndServeMux(u.config.VerifyClientTLS); err != nil {
		return fmt.Errorf("http ListenAndServe error %s", err.Error())
	}

	return nil
}

func (u *UserManager) initRouters(ws *restful.WebService) {
	v1http.InitV1Routers(ws)
}

// RegDiscover register and discovery to zk
func (u *UserManager) RegDiscover() {
	rd := RegisterDiscover.NewRegDiscoverEx(u.config.RegDiscvSrv, 10*time.Second)
	//start regdiscover
	if err := rd.Start(); err != nil {
		blog.Error("fail to start register and discover serv. err:%s", err.Error())
	}
	//register user-manager
	userMgrServInfo := new(types.BcsUserMgrServInfo)

	userMgrServInfo.IP = u.config.LocalIp
	userMgrServInfo.Port = u.config.InsecurePort
	userMgrServInfo.Scheme = "http"
	userMgrServInfo.MetricPort = u.config.MetricPort
	if u.config.ServCert.IsSSL {
		userMgrServInfo.Scheme = "https"
		userMgrServInfo.Port = u.config.Port
	}
	userMgrServInfo.Version = version.GetVersion()
	userMgrServInfo.Pid = os.Getpid()

	data, err := json.Marshal(userMgrServInfo)
	if err != nil {
		blog.Error("fail to marshal userMgrServInfo to json. err:%s", err.Error())
	}

	path := types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_USERMGR + "/" + u.config.LocalIp

	blog.Infof("register key %s user-manager %s", path, string(data))
	_ = rd.RegisterAndWatchService(path, data)
}
