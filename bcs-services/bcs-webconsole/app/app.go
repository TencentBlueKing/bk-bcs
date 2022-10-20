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
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	yaml "github.com/go-micro/plugins/v4/config/encoder/yaml"
	etcd "github.com/go-micro/plugins/v4/registry/etcd"
	mhttp "github.com/go-micro/plugins/v4/server/http"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	microConf "go-micro.dev/v4/config"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/config/reader/json"
	"go-micro.dev/v4/config/source/file"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/util/cmd"
	"golang.org/x/sync/errgroup"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/web"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
)

var (
	// 变量, 编译后覆盖
	service              = "webconsole.bkbcs.tencent.com"
	appName              = "bcs-webconsole"
	versionTag           = "latest"
	credentialConfigPath = cli.StringSlice{}
	serverAddressFlag    = "server_address" // 默认启动ip:port
	podIPsEnv            = "POD_IPs"        // 双栈监听环境变量
	ipv6Interface        = "IPV6_INTERFACE" // ipv6本地网关地址
)

// WebConsoleManager is an console struct
type WebConsoleManager struct {
	ctx           context.Context
	opt           *options.WebConsoleManagerOption
	microService  micro.Service
	microConfig   microConf.Config
	multiCredConf *options.MultiCredConf
	serverAddress string
	listenPort    string
}

// NewWebConsoleManager xxx
func NewWebConsoleManager(ctx context.Context, opt *options.WebConsoleManagerOption) *WebConsoleManager {
	return &WebConsoleManager{
		ctx: ctx,
		opt: opt,
	}
}

// Init 初始化
func (c *WebConsoleManager) Init() error {
	// 初始化服务注册, 配置文件等
	microService, microConfig, multiCredConf := c.initMicroService()
	c.microService = microService
	c.microConfig = microConfig
	c.multiCredConf = multiCredConf

	// etcd 服务发现注册
	etcdRegistry, err := c.initEtcdRegistry()
	if err != nil {
		return err
	}

	metadata := map[string]string{}
	dualStackListener := listener.NewDualStackListener()

	if err := dualStackListener.AddListenerWithAddr(c.serverAddress); err != nil {
		return err
	}

	ipv6Addr := getIPv6AddrFromEnv(c.listenPort)
	if ipv6Addr != "" {
		metadata[types.IPV6] = ipv6Addr
		if err := dualStackListener.AddListenerWithAddr(ipv6Addr); err != nil {
			return err
		}
		logger.Infof("dualStackListener with ipv6: %s", ipv6Addr)
	}

	microService.Init(
		micro.Server(mhttp.NewServer(mhttp.Listener(dualStackListener))),
		micro.AfterStop(func() error {
			// 会让 websocket 发送 EndOfTransmission, 不能保证一定发送成功
			logger.Info("receive interput, gracefully shutdown")
			<-c.ctx.Done()
			return nil
		}),
	)

	// 服务注册需要单独处理
	if etcdRegistry != nil {
		microService.Init(
			micro.Name(service),
			micro.Version(versionTag),
			micro.RegisterTTL(time.Second*30),
			micro.RegisterInterval(time.Second*15),
			micro.Registry(etcdRegistry),
			micro.Metadata(metadata),
		)
	}

	// http 路由注册
	router, err := c.initHTTPService()
	if err != nil {
		return err
	}

	if err := micro.RegisterHandler(microService.Server(), router); err != nil {
		return err
	}

	return nil
}

func (m *WebConsoleManager) initMicroService() (micro.Service, microConf.Config, *options.MultiCredConf) {
	var (
		configPath string
	)

	// new config
	conf, _ := microConf.NewConfig(
		microConf.WithReader(json.NewReader(reader.WithEncoder(yaml.NewEncoder()))),
	)
	var multiCredConf *options.MultiCredConf

	cmdOptions := []cmd.Option{
		cmd.Description("bcs webconsole micro service"),
		cmd.Version(versionTag),
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(appName+",", version.GetVersion())
	}

	microCmd := cmd.NewCmd(cmdOptions...)
	microCmd.App().Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    serverAddressFlag,
			EnvVars: []string{"MICRO_SERVER_ADDRESS"},
			Usage:   "Bind address for the server. 127.0.0.1:8080",
		},
		&cli.StringFlag{
			Name:        "config",
			Usage:       "config file path",
			Required:    true,
			Destination: &configPath,
		},
		&cli.StringSliceFlag{
			Name:        "credential-config",
			Usage:       "credential config file path",
			Required:    false,
			Destination: &credentialConfigPath,
		},
	}

	microCmd.App().Action = func(c *cli.Context) error {
		if err := conf.Load(file.NewSource(file.WithPath(configPath))); err != nil {
			return err
		}

		// 解析端口地址
		m.listenPort = parseListenPort(c)
		m.serverAddress = c.Value(serverAddressFlag).(string)

		// 初始化配置文件
		if err := config.G.ReadFrom(conf.Bytes()); err != nil {
			logger.Errorf("config not valid, err: %s, exited", err)
			os.Exit(1)
		}

		logger.Infof("load conf from %s", configPath)

		// 授权信息
		if len(credentialConfigPath.Value()) > 0 {
			credConf, err := options.NewMultiCredConf(credentialConfigPath.Value())
			if err != nil {
				logger.Errorf("config not valid, err: %s, exited", err)
				os.Exit(1)
			}
			multiCredConf = credConf

		}
		return nil
	}

	srv := micro.NewService()
	opts := []micro.Option{
		micro.Name(service),
		micro.Version(versionTag),
		micro.Cmd(microCmd),
	}

	// 配置文件, 日志这里才设置完成
	srv.Init(opts...)

	return srv, conf, multiCredConf
}

// initHTTPService 初始化 gin Http 配置
func (c *WebConsoleManager) initHTTPService() (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger(), cors.Default())
	router.Use(i18n.Localize())

	// 注册模板和静态资源
	router.SetHTMLTemplate(web.WebTemplate())

	// 静态资源
	routePrefix := config.G.Web.RoutePrefix
	if routePrefix == "" {
		routePrefix = "/webconsole"
	}

	// 支持路径 prefix 透传和 rewrite 的场景
	router.Group(routePrefix).StaticFS("/web/static", http.FS(web.WebStatic()))
	router.Group("").StaticFS("/web/static", http.FS(web.WebStatic()))

	handlerOpts := &route.Options{
		RoutePrefix: routePrefix,
		Client:      c.microService.Client(),
		Router:      router,
	}

	// 注册 HTTP 请求
	for _, r := range []route.Registrar{
		web.NewRouteRegistrar(handlerOpts),
		api.NewRouteRegistrar(handlerOpts),
	} {
		r.RegisterRoute(router.Group(routePrefix))
		r.RegisterRoute(router.Group(""))
	}

	return router, nil
}

// initEtcdRegistry etcd 服务注册
func (c *WebConsoleManager) initEtcdRegistry() (registry.Registry, error) {
	endpoints := c.microConfig.Get("etcd", "endpoints").String("")
	if endpoints == "" {
		return nil, nil
	}

	etcdRegistry := etcd.NewRegistry(registry.Addrs(strings.Split(endpoints, ",")...))

	ca := c.microConfig.Get("etcd", "ca").String("")
	cert := c.microConfig.Get("etcd", "cert").String("")
	key := c.microConfig.Get("etcd", "key").String("")
	if ca != "" && cert != "" && key != "" {
		tlsConfig, err := ssl.ClientTslConfVerity(ca, cert, key, "")
		if err != nil {
			return nil, err
		}
		etcdRegistry.Init(registry.TLSConfig(tlsConfig))
	}

	return etcdRegistry, nil
}

// checkVersion refer to https://github.com/urfave/cli/blob/main/help.go#L318 but use os.args check
func checkVersion() bool {
	if len(os.Args) < 2 {
		return false
	}
	arg := os.Args[1]
	for _, name := range cli.VersionFlag.Names() {
		if arg == "-"+name || arg == "--"+name {
			return true
		}
	}
	return false
}

// parseListenPort 解析端口
func parseListenPort(c *cli.Context) string {
	// 解析端口地址
	ipv4, ok := c.Value(serverAddressFlag).(string)
	if !ok || ipv4 == "" {
		return ""
	}

	_, port, _ := net.SplitHostPort(ipv4)
	return port
}

// getIPv6AddrFromEnv 解析ipv6
func getIPv6AddrFromEnv(listenPort string) string {
	if listenPort == "" {
		return ""
	}

	podIPs := os.Getenv(podIPsEnv)
	if podIPs == "" {
		return ""
	}

	ipv6 := util.GetIPv6Address(podIPs)
	if ipv6 == "" {
		return ""
	}

	// 在实际中，ipv6不能是回环地址
	if v := net.ParseIP(ipv6); v == nil || v.IsLoopback() {
		return ""
	}

	// local link ipv6 需要带上 interface， 格式如::%eth0
	ipv6Interface := os.Getenv(ipv6Interface)
	if ipv6Interface != "" {
		ipv6 = ipv6 + "%" + ipv6Interface
	}

	return net.JoinHostPort(ipv6, listenPort)
}

// Run create a pid
func (c *WebConsoleManager) Run() error {
	if checkVersion() {
		return nil
	}

	logger.Info("starting bcs-webconsole.")

	eg, ctx := errgroup.WithContext(c.ctx)

	podCleanUpMgr := podmanager.NewCleanUpManager(ctx)
	eg.Go(func() error {
		return podCleanUpMgr.Run()
	})

	if c.multiCredConf != nil {
		c.microService.Init(micro.AfterStop(func() error {
			c.multiCredConf.Stop()
			return nil
		}))

		eg.Go(func() error {
			return c.multiCredConf.Watch()
		})
	}

	eg.Go(func() error {
		if err := c.microService.Run(); err != nil {
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		defer logger.CloseLogs()
		return err
	}
	return nil
}
