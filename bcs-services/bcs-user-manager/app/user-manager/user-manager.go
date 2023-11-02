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

package usermanager

import (
	"bytes"
	"crypto/tls"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encryptv2"
	bcshttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"github.com/emicklei/go-restful"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go-micro.dev/v4/registry"

	i18n2 "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/permission"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v3http"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/migrations"
)

var (
	// ErrHTTPServerNotInit server not init
	ErrHTTPServerNotInit = errors.New("UserManager server not init")
)

// UserManager http interface of user-manager
type UserManager struct {
	config   *config.UserMgrConfig
	httpServ *httpserver.HttpServer

	IamPermClient iam.PermMigrateClient
	EtcdRegistry  registry.Registry

	permService *permission.PermVerifyClient
}

// NewUserManager creates an UserManager object
func NewUserManager(conf *config.UserMgrConfig) *UserManager {
	userManager := &UserManager{
		config:   conf,
		httpServ: httpserver.NewIPv6HttpServer(conf.Port, conf.Address, conf.IPv6Address, conf.Sock),
	}

	if conf.ServCert.IsSSL {
		userManager.httpServ.SetSsl(conf.ServCert.CAFile, conf.ServCert.CertFile, conf.ServCert.KeyFile,
			conf.ServCert.CertPasswd)
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
	v1http.InitV1Routers(u.httpServ.NewWebService("/usermanager", nil), u.permService)
	v3http.InitV3Routers(u.httpServ.NewWebService("/usermanager/v3", nil))

	router := u.httpServ.GetRouter()
	webContainer := u.httpServ.GetWebContainer()
	// set recover handler
	webContainer.RecoverHandler(responseOnRecover)
	webContainer.DoNotRecover(false)

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
		_ = resp.WriteHeaderAndEntity(http.StatusUnauthorized, bcshttp.APIRespone{
			Result:  false,
			Code:    common.BcsErrApiUnauthorized,
			Message: "must provide admin token to request with websocket",
			Data:    nil,
		})
		return
	}

	chain.ProcessFilter(req, resp)
}

func (u *UserManager) initPermService() error {
	permService := permission.NewPermVerifyClient(u.config.PermissionSwitch, u.IamPermClient)
	u.permService = permService

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
		Debug:       u.config.IAMConfig.ServerDebug,
	}
	iamCli, err := iam.NewIamMigrateClient(opt)
	if err != nil {
		blog.Errorf("initIamPermClient failed: %v", err)
		return err
	}

	u.IamPermClient = iamCli
	config.GloablIAMClient = iamCli
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

func (u *UserManager) initCryptor() error {
	if !u.config.Encrypt.Enable {
		return nil
	}
	conf := &encryptv2.Config{
		Enabled:   u.config.Encrypt.Enable,
		Algorithm: encryptv2.Algorithm(u.config.Encrypt.Algorithm),
	}
	switch conf.Algorithm {
	case encryptv2.Sm4:
		conf.Sm4 = &encryptv2.Sm4Conf{
			Key: u.config.Encrypt.Secret.Key,
			Iv:  u.config.Encrypt.Secret.Secret,
		}
	case encryptv2.AesGcm:
		conf.AesGcm = &encryptv2.AesGcmConf{
			Key:   u.config.Encrypt.Secret.Key,
			Nonce: u.config.Encrypt.Secret.Secret,
		}
	case encryptv2.Normal:
		conf.Normal = &encryptv2.NormalConf{
			PriKey: static.EncryptionKey,
		}
	}
	cryptor, err := encryptv2.NewCrypto(conf)
	if err != nil {
		return fmt.Errorf("init cryptor failed, %s", err.Error())
	}
	config.GlobalCryptor = cryptor
	blog.Info("init cryptor successfully")
	return nil
}

// Migrate migrates something.
//
// op is a pointer to options.Migration.
// It is the description of the parameter.
// The function does not return anything.
func (u *UserManager) migrate() {
	go func() {
		blog.Info("start iam migration")
		tempVar := map[string]string{
			"BK_IAM_SYSTEM_ID": u.config.IAMConfig.SystemID,
			"APP_CODE":         u.config.IAMConfig.AppCode,
			"BCS_HOST":         u.config.BcsAPI.Host,
		}
		d, err := iofs.New(migrations.MigrationFS, ".")
		if err != nil {
			blog.Errorf("get migrations files error, %s", err.Error())
			return
		}
		if err := u.IamPermClient.Migrate(sqlstore.GCoreDB.DB(), d, "bk_iam_migrations",
			5*time.Minute, tempVar); err != nil {
			if strings.Contains(err.Error(), "no change") {
				blog.Info("iam migration success")
				return
			}
			blog.Errorf("migrate iam failed, %s", err.Error())
			return
		}
		blog.Info("iam migration success")
	}()
}

// initI18n init i18n
func (u *UserManager) initI18n() {
	i18n.Instance()
	// 加载翻译文件路径
	i18n.SetPath([]embed.FS{i18n2.Assets})
	// 设置默认语言
	// 默认是 zh
	i18n.SetLanguage("zh")
}

func (u *UserManager) initUserManagerServer() error {
	var err error
	err = u.initCryptor()
	if err != nil {
		return err
	}

	err = u.initEtcdRegistry()
	if err != nil {
		return err
	}

	err = u.initIamPermClient()
	if err != nil {
		return err
	}

	err = u.initPermService()
	if err != nil {
		return err
	}

	u.migrate()
	u.initI18n()

	return nil
}

// responseOnRecover response on recover
func responseOnRecover(panicReason interface{}, httpWriter http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	blog.Error(buffer.String())
	httpWriter.WriteHeader(http.StatusInternalServerError)
	httpWriter.Write([]byte(`{"code": 500, "message": "server error"}`))
}
