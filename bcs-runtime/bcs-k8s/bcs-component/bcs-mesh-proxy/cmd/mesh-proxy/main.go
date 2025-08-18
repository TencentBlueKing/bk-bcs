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

// Package main is the main package for the mesh proxy.
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mesh-proxy/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mesh-proxy/pkg/proxy"
)

func main() {
	var (
		configFile = flag.String("config", "/etc/mesh-proxy/config.yaml", "配置文件路径")
		certFile   = flag.String("cert", "", "TLS证书文件路径")
		keyFile    = flag.String("key", "", "TLS密钥文件路径")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configFile)
	if err != nil {
		klog.Fatalf("加载配置失败: %v", err)
	}

	// 创建服务器
	server, actualCertFile, actualKeyFile, err := startServer(cfg, cfg.Proxy.Port, *certFile, *keyFile)
	if err != nil {
		klog.Fatalf("创建服务器失败: %v", err)
	}

	// 启动服务器
	go func() {
		klog.Infof("mesh-proxy 服务启动在端口 %d", cfg.Proxy.Port)
		var serveErr error
		if server.TLSConfig != nil {
			serveErr = server.ListenAndServeTLS(actualCertFile, actualKeyFile)
		} else {
			klog.Info("使用HTTP模式")
			serveErr = server.ListenAndServe()
		}
		if serveErr != nil && serveErr != http.ErrServerClosed {
			klog.Fatalf("服务器启动失败: %v", serveErr)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	klog.Info("正在关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		klog.Fatalf("服务器强制关闭: %v", err)
	}

	klog.Info("服务器已关闭")
}

// createTLSServer 创建TLS服务器配置
func createTLSServer(cfg *config.Config) (*http.Server, error) {
	// 创建自定义TLS配置
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// 根据配置设置客户端认证模式
	switch cfg.Proxy.TLS.ClientAuth {
	case "NoClientCert":
		tlsConfig.ClientAuth = tls.NoClientCert
		klog.Info("TLS配置: 不验证客户端证书")
	case "RequestClientCert":
		tlsConfig.ClientAuth = tls.RequestClientCert
		klog.Info("TLS配置: 请求但不验证客户端证书")
	case "RequireAnyClientCert":
		tlsConfig.ClientAuth = tls.RequireAnyClientCert
		klog.Info("TLS配置: 要求任何客户端证书")
	case "VerifyClientCertIfGiven":
		tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
		klog.Info("TLS配置: 如果提供则验证客户端证书")
	case "RequireAndVerifyClientCert":
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		klog.Info("TLS配置: 要求并验证客户端证书")
	default:
		// 默认不验证客户端证书
		tlsConfig.ClientAuth = tls.NoClientCert
		klog.Info("TLS配置: 默认不验证客户端证书")
	}

	// 设置CA证书（如果提供）
	if cfg.Proxy.TLS.CAFile != "" {
		caCert, readErr := os.ReadFile(cfg.Proxy.TLS.CAFile)
		if readErr != nil {
			return nil, fmt.Errorf("读取CA证书失败: %v", readErr)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("解析CA证书失败")
		}
		tlsConfig.ClientCAs = caCertPool
		klog.Infof("已加载CA证书: %s", cfg.Proxy.TLS.CAFile)
	}

	return &http.Server{
		TLSConfig: tlsConfig,
	}, nil
}

// startServer 启动服务器
func startServer(cfg *config.Config, port int, certFile, keyFile string) (*http.Server, string, string, error) {
	// 创建代理服务器
	proxyServer := proxy.NewServer(cfg)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      proxyServer,
		ReadTimeout:  cfg.TargetCluster.Timeout,
		WriteTimeout: cfg.TargetCluster.Timeout,
		IdleTimeout:  cfg.TargetCluster.Timeout,
	}

	// 优先使用命令行参数，如果没有则使用配置文件
	if certFile == "" && keyFile == "" && cfg.Proxy.TLS.Enabled {
		certFile = cfg.Proxy.TLS.CertFile
		keyFile = cfg.Proxy.TLS.KeyFile
	}

	if certFile != "" && keyFile != "" {
		klog.Infof("使用TLS证书: %s, %s", certFile, keyFile)

		tlsServer, err := createTLSServer(cfg)
		if err != nil {
			return nil, "", "", fmt.Errorf("创建TLS服务器失败: %v", err)
		}
		server.TLSConfig = tlsServer.TLSConfig
	}

	return server, certFile, keyFile, nil
}
