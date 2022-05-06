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

package main

import (
	"flag"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-multi-ns-proxy/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-multi-ns-proxy/internal/proxy"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-multi-ns-proxy/pkg/filewatcher"
)

func main() {
	pflag.String(constant.FlagKeyKubeconfigMode,
		"file", "mode for proxy to get all kubeconfigs, available [secret, file]")
	pflag.String(constant.FlagKeyKubeconfigSecretName,
		"", "k8s secret name for proxy to get all kubeconfigs when use secret mode")
	pflag.String(constant.FlagKeyKubeconfigSecretNamespace,
		"", "k8s secret namespace for proxy to get all kubeconfigs when use secret mode")
	pflag.String(constant.FlagKeyKubeconfigDir,
		"", "the directory which holds all kubeconfigs for different namespaces")
	pflag.String(constant.FlagKeyKubeconfigDefaultNs,
		"", "the default namespace to use for non-namespaced api resource")
	pflag.Duration(constant.FlagKeyKubeconfigCheckDuration,
		10*time.Second, "interval for checking kubeconfig directory")
	pflag.Uint(constant.FlagKeyProxyPort, 8080, "listening port for proxy server")
	pflag.String(constant.FlagKeyProxyAddress, "127.0.0.1", "listening address for proxy server")
	pflag.String(constant.FlagKeyProxyServerCert, "", "cert file path for proxy server")
	pflag.String(constant.FlagKeyProxyServerKey, "", "key file path for proxy server")

	var configName string
	pflag.StringVar(&configName, constant.FlagKeyConfigName,
		"config.yaml", "The config file name of bcs-multi-ns-proxy")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	viper.AutomaticEnv()
	viper.BindPFlags(pflag.CommandLine)
	if len(configName) != 0 {
		viper.SetConfigFile(configName)
	} else {
		// Search config in home directory with name (without extension).
		viper.AddConfigPath("/etc")
		viper.AddConfigPath("/data/bcs/bcs-multi-ns-proxy")
		viper.AddConfigPath(".")
		viper.SetConfigName("bcs-multi-ns-proxy")
		viper.SetConfigType("yaml")
	}
	pflag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			zap.L().Info("not config file found, use default value")
		} else {
			zap.L().Fatal("read config failed", zap.Error(err))
		}
	}

	handler, err := proxy.NewHandler(viper.GetString(constant.FlagKeyKubeconfigDefaultNs))
	if err != nil {
		zap.L().Fatal("create proxy handler failed", zap.Error(err))
	}

	var conflister filewatcher.Lister
	kubeconfigMode := viper.GetString(constant.FlagKeyKubeconfigMode)
	switch kubeconfigMode {
	case constant.KubeconfigModeFile:
		conflister = filewatcher.NewFileLister(viper.GetString(constant.FlagKeyKubeconfigDir))
	case constant.KubeconfigModeSecret:
		secretName := viper.GetString(constant.FlagKeyKubeconfigSecretName)
		secretNamespace := viper.GetString(constant.FlagKeyKubeconfigSecretNamespace)
		if len(secretName) == 0 || len(secretNamespace) == 0 {
			zap.L().Fatal("secret name or namespace cannot be empty",
				zap.String("secretname", secretName), zap.String("secretns", secretNamespace))
		}
		conflister, err = filewatcher.NewSecretLister("", secretName, secretNamespace)
	default:
		zap.L().Fatal("invalid kubeconfig mode", zap.String("mode", kubeconfigMode))
	}
	if err != nil {
		zap.L().Fatal("create config lister failed", zap.Error(err))
	}

	watcher := filewatcher.NewWatcher(conflister, viper.GetDuration(constant.FlagKeyKubeconfigCheckDuration))
	watcher.RegisterHandler(handler)
	go watcher.WatchLoop()
	defer watcher.Stop()

	httpServer := httpserver.NewHttpServer(
		viper.GetUint(constant.FlagKeyProxyPort),
		viper.GetString(constant.FlagKeyProxyAddress), "")

	serverCertFile := viper.GetString(constant.FlagKeyProxyServerCert)
	serverKeyFile := viper.GetString(constant.FlagKeyProxyServerKey)
	if len(serverCertFile) != 0 && len(serverKeyFile) != 0 {
		httpServer.SetSsl("", serverCertFile, serverKeyFile, "")
	}

	router := httpServer.GetRouter()
	router.Handle("/{uri:.*}", handler)
	if err := httpServer.ListenAndServeMux(false); err != nil {
		blog.Errorf("http listen and serve failed, err %s", err.Error())
		os.Exit(1)
	}

	ch := make(chan int)
	<-ch
}
