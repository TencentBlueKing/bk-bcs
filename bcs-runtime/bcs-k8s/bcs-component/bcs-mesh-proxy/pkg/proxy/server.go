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

// Package proxy provides the proxy server for the mesh proxy.
package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mesh-proxy/pkg/config"
)

// Server 代理服务器
type Server struct {
	config     *config.Config
	k8sClient  *kubernetes.Clientset
	httpClient *http.Client
	restConfig *rest.Config
}

// NewServer 创建新的代理服务器
func NewServer(cfg *config.Config) *Server {
	// 创建k8s客户端配置
	var k8sConfig *rest.Config
	var err error

	if cfg.TargetCluster.UseInClusterConfig {
		// 使用client-go内置的in-cluster配置，自动支持token轮转
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("获取in-cluster配置失败: %v", err)
		}
		cfg.TargetCluster.APIServer = k8sConfig.Host
		klog.Info("使用in-cluster配置")
	} else {
		k8sConfig = &rest.Config{
			Host: cfg.TargetCluster.APIServer,
		}

		// 设置认证
		if cfg.TargetCluster.BearerToken != "" {
			k8sConfig.BearerToken = cfg.TargetCluster.BearerToken
		} else if cfg.TargetCluster.ClientCertPath != "" && cfg.TargetCluster.ClientKeyPath != "" {
			k8sConfig.TLSClientConfig = rest.TLSClientConfig{
				CertFile: cfg.TargetCluster.ClientCertPath,
				KeyFile:  cfg.TargetCluster.ClientKeyPath,
			}
		}

		// 设置CA证书
		if cfg.TargetCluster.CACertPath != "" {
			k8sConfig.TLSClientConfig.CAFile = cfg.TargetCluster.CACertPath
		}
	}

	// 创建k8s客户端
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		klog.Fatalf("创建k8s客户端失败: %v", err)
	}

	// 使用rest.HTTPWrappersForConfig创建HTTP传输层，内置token轮转支持
	transport, err := rest.HTTPWrappersForConfig(k8sConfig, &http.Transport{
		TLSClientConfig: &tls.Config{
			// nolint:gosec // 这是配置选项，允许跳过TLS验证
			InsecureSkipVerify: cfg.Proxy.InsecureSkipTLSVerify,
		},
	})
	if err != nil {
		klog.Fatalf("创建HTTP传输层失败: %v", err)
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   cfg.Proxy.RequestTimeout,
	}

	return &Server{
		config:     cfg,
		k8sClient:  k8sClient,
		httpClient: httpClient,
		restConfig: k8sConfig,
	}
}

// ServeHTTP 处理HTTP请求
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取客户端IP地址
	clientIP := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		clientIP = forwardedFor
	} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		clientIP = realIP
	}

	// 获取User-Agent
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		userAgent = "unknown"
	}

	// 记录详细的请求信息
	klog.Infof("收到请求 - 来源: %s, 方法: %s, URL: %s, User-Agent: %s",
		clientIP, r.Method, r.URL.String(), userAgent)

	// 处理健康检查和就绪检查
	switch r.URL.Path {
	case "/healthz":
		s.handleHealthCheck(w, r)
		return
	case "/readyz":
		s.handleReadyCheck(w, r)
		return
	}

	// 检查请求路径
	if !s.isAllowedRequest(r.URL.Path) {
		klog.Warningf("拒绝请求 - 来源: %s, 路径: %s (不在允许的API组中)", clientIP, r.URL.Path)
		http.Error(w, "不允许的API请求", http.StatusForbidden)
		return
	}

	// 构建目标URL
	targetURL, err := s.buildTargetURL(r.URL)
	if err != nil {
		klog.Errorf("构建目标URL失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}

	klog.Infof("代理请求 - 目标URL: %s", targetURL.String())

	// 创建代理请求
	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		klog.Errorf("创建代理请求失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// 发送请求
	resp, err := s.httpClient.Do(proxyReq)
	if err != nil {
		klog.Errorf("代理请求失败: %v", err)
		http.Error(w, "代理请求失败", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	klog.Infof("收到响应 - 状态码: %d, 内容类型: %s",
		resp.StatusCode, resp.Header.Get("Content-Type"))

	// 复制响应头
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 设置响应状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		klog.Errorf("复制响应体失败: %v", err)
	}
}

// handleHealthCheck 处理健康检查
func (s *Server) handleHealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// handleReadyCheck 处理就绪检查
func (s *Server) handleReadyCheck(w http.ResponseWriter, _ *http.Request) {
	// 检查k8s API版本信息
	_, err := s.k8sClient.Discovery().ServerVersion()
	if err != nil {
		klog.Errorf("就绪检查失败: %v", err)
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// isAllowedRequest 检查请求是否被允许
func (s *Server) isAllowedRequest(path string) bool {
	// 检查是否是istio相关的API
	if strings.Contains(path, "/apis/networking.istio.io/") {
		return true
	}

	// 检查其他允许的API组
	for _, allowedGroup := range s.config.Proxy.AllowedAPIGroups {
		if strings.Contains(path, "/apis/"+allowedGroup+"/") {
			return true
		}
	}

	return false
}

// buildTargetURL 构建目标URL
func (s *Server) buildTargetURL(originalURL *url.URL) (*url.URL, error) {
	baseURL := s.config.TargetCluster.APIServer

	targetURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("解析基础URL失败: %v", err)
	}

	// 设置路径
	targetURL.Path = originalURL.Path
	targetURL.RawQuery = originalURL.RawQuery

	return targetURL, nil
}
