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
	"errors"
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
	yaml2 "gopkg.in/yaml.v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/api"
	consoleAudit "github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/audit"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/audit/record"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/audit/replay"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/perf"
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
	configPath           = ""
	serverAddressFlag    = "server-address" // 默认启动ip
	serverPortFlag       = "server-port"    // 默认启动port
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

	// 绑定主端口
	dualStackListener := listener.NewDualStackListener()
	if err = dualStackListener.AddListenerWithAddr(getListenAddr(c.serverAddress, c.listenPort)); err != nil {
		return err
	}

	// IPv6, metadata 都metadata信息
	metadata := map[string]string{}
	ipv6Addr := getIPv6AddrFromEnv()
	if ipv6Addr != "" {
		metadata[types.IPV6] = getListenAddr(ipv6Addr, c.listenPort)
	}

	// 单栈IPv6 可能重复
	if ipv6Addr != "" && ipv6Addr != c.serverAddress {
		listenAddr := getListenAddr(ipv6Addr, c.listenPort)
		if err = dualStackListener.AddListenerWithAddr(listenAddr); err != nil {
			return err
		}
		logger.Infof("dualStackListener with ipv6: %s", listenAddr)
	}

	microService.Init(
		micro.Server(mhttp.NewServer(mhttp.Listener(dualStackListener))),
		micro.AfterStop(func() error {
			// close audit client
			consoleAudit.GetAuditClient().Close()
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
	router := c.initHTTPService()

	if err := micro.RegisterHandler(microService.Server(), router); err != nil {
		return err
	}

	return nil
}

func (c *WebConsoleManager) initMicroService() (micro.Service, microConf.Config, *options.MultiCredConf) {
	// new config
	conf, _ := microConf.NewConfig(microConf.WithReader(json.NewReader(reader.WithEncoder(yaml.NewEncoder()))))
	var multiCredConf *options.MultiCredConf
	cmdOptions := []cmd.Option{
		cmd.Description("bcs webconsole micro service"),
		cmd.Version(versionTag),
	}
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(appName+",", version.GetVersion())
	}
	microCmd := cmd.NewCmd(cmdOptions...)
	microCmd.App().Flags = buildFlags()
	microCmd.App().Commands = buildCommands()
	microCmd.App().Action = func(ctx *cli.Context) error {
		if ctx.Bool("confinfo") {
			encoder := yaml2.NewEncoder(os.Stdout)
			encoder.SetIndent(2)
			if err := encoder.Encode(config.G); err != nil {
				os.Exit(1)
				return err
			}
			os.Exit(0)
			return nil
		}
		if ctx.String(serverAddressFlag) == "" || ctx.String("config") == "" {
			logger.Error("--config and --server-address not set")
			os.Exit(1)
		}
		if err := conf.Load(file.NewSource(file.WithPath(configPath))); err != nil {
			return err
		}
		c.listenPort = ctx.Value(serverPortFlag).(string)
		c.serverAddress = ctx.Value(serverAddressFlag).(string)
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
func (c *WebConsoleManager) initHTTPService() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
	}))
	router.Use(i18n.Localize())

	// 注册模板和静态资源
	router.SetHTMLTemplate(web.WebTemplate())

	// 静态资源
	routePrefix := config.G.Web.RoutePrefix
	if routePrefix == "" {
		routePrefix = "/webconsole"
	}
	// 回放文件
	replayPath := config.G.Audit.DataDir

	// 支持路径 prefix 透传和 rewrite 的场景
	router.Group(routePrefix).StaticFS("/web/static", http.FS(web.WebStatic()))
	router.Group("").StaticFS("/web/static", http.FS(web.WebStatic()))
	router.Group(routePrefix).StaticFS("/casts", http.Dir(replayPath))
	router.Group("").StaticFS("/casts", http.Dir(replayPath))

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

	return router
}

// initEtcdRegistry etcd 服务注册
func (c *WebConsoleManager) initEtcdRegistry() (registry.Registry, error) {
	endpoints := c.microConfig.Get("etcd", "endpoints").String("")

	// 添加环境变量
	if endpoints == "" {
		endpoints = config.BCS_ETCD_HOST
	}

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
		if err := etcdRegistry.Init(registry.TLSConfig(tlsConfig)); err != nil {
			return nil, err
		}
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

// getIPv6AddrFromEnv 解析ipv6
func getIPv6AddrFromEnv() string {
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
	return ipv6
}

// getListenAddr
func getListenAddr(addr, port string) string {
	if ip := net.ParseIP(addr); ip == nil {
		return ""
	}

	if util.IsIPv6(addr) {
		// local link ipv6 需要带上 interface， 格式如::%eth0
		value := os.Getenv(ipv6Interface)
		if value != "" {
			addr = addr + "%" + value
		}
	}

	return net.JoinHostPort(addr, port)
}

func buildFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  serverAddressFlag,
			Usage: "Bind ip address for the server. 127.0.0.1",
		},
		&cli.StringFlag{
			Name:  serverPortFlag,
			Value: "8083",
			Usage: "Bind port for the server",
		},
		&cli.StringFlag{
			Name:        "config",
			Usage:       "config file path",
			Destination: &configPath,
		},
		&cli.StringSliceFlag{
			Name:        "credential-config",
			Usage:       "credential config file path",
			Required:    false,
			Destination: &credentialConfigPath,
		},
		&cli.BoolFlag{
			Name:    "confinfo",
			Usage:   "print init confinfo to stdout",
			Aliases: []string{"o"},
		},
	}
}

// 命令行子命令
func buildCommands() []*cli.Command {
	return cli.Commands{
		&cli.Command{
			Name:    "replay",
			Usage:   "replay terminal session record",
			Aliases: []string{"r"},
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					logger.Error("replay file not set")
					return errors.New("replay file not set")
				}
				if err := replay.Replay(c.Args().First()); err != nil {
					logger.Errorf("replay failure, err: %s, exited", err)
					return err
				}
				os.Exit(0)
				return nil
			},
		},
	}
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

	// 定时上报 cast
	uploader := record.GetGlobalUploader()
	eg.Go(func() error {
		return uploader.IntervalUpload(ctx)
	})

	// 定时上报用户延迟命令列表数据
	performance := perf.GetGlobalPerformance()
	eg.Go(func() error {
		return performance.Run(ctx)
	})

	if c.multiCredConf != nil {
		c.microService.Init(micro.AfterStop(func() error {
			return c.multiCredConf.Stop()
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
