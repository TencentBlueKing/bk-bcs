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

// Package util xxx
package util

import (
	"context"
	"crypto/tls"
	"fmt"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog"
	"net"
	"net/http"
	"os"
	"time"
)

// GetCtx xxx
func GetCtx(duration time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	go func() {
		time.Sleep(duration)
		cancel()
	}()
	return ctx
}

// GetServerCert xxx
func GetServerCert(domain, ip, port string) (time.Time, error) {
	// 创建自定义的Transport，用于指定IP地址
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证
			ServerName:         domain,
		},
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 2 * time.Second,
			}
			conn, err := dialer.DialContext(ctx, "tcp", ip+":"+port)
			if err != nil {
				return nil, err
			}
			tlsConn := tls.Client(conn, &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         domain,
			})
			if err := tlsConn.Handshake(); err != nil {
				return nil, err
			}
			return tlsConn, nil
		},
	}

	// 创建自定义的Client，使用自定义的Transport
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	// 发起HTTPS请求
	resp, err := client.Get("https://" + domain)
	if err != nil {
		return time.Now(), err
	}
	defer resp.Body.Close()

	// 获取远程证书
	cert := resp.TLS.PeerCertificates[0]

	// 获取证书的过期时间
	expiration := cert.NotAfter

	return expiration, nil
}

// WriteConfigIfNotExist xxx
func WriteConfigIfNotExist(filePath, content string) error {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			file, openErr := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if openErr != nil {
				return err
			}

			defer file.Close()

			// 写入文本信息
			_, err = file.WriteString(content)
			if err != nil {
				return err
			}

			return nil
		} else {
			return err
		}

	} else {
		return nil
	}
}

// ReadorInitConf xxx
func ReadorInitConf(configFilePath string, obj interface{}, initContent string) error {
	_, err := os.Stat(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = WriteConfigIfNotExist(configFilePath, initContent)
			if err != nil {
				return err
			} else {
				return ReadConf(configFilePath, obj)
			}
		}
		return err
	} else {
		return ReadConf(configFilePath, obj)
	}
}

// ReadFromStr xxx
func ReadFromStr(obj interface{}, initContent string) error {
	if err := json.Unmarshal([]byte(initContent), obj); err != nil {
		if err = yaml.Unmarshal([]byte(initContent), obj); err != nil {
			return fmt.Errorf("decode config %s failed, err %s", initContent, err.Error())
		}
	}

	return nil
}

// ReadConf xxx
func ReadConf(configFilePath string, obj interface{}) error {
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(configFileBytes, obj); err != nil {
		if err = yaml.Unmarshal(configFileBytes, obj); err != nil {
			return fmt.Errorf("decode clustercheck config file %s failed, err %s", configFilePath, err.Error())
		}
	}

	return nil
}

// GetHostPath xxx
func GetHostPath() string {
	hostPath := os.Getenv("HOST_PATH")
	if hostPath == "" {
		hostPath = "/"
	}

	return hostPath
}

// GetNodeName xxx
func GetNodeName() string {
	name := os.Getenv("NODE_NAME")
	var err error
	if name == "" {
		name, err = os.Hostname()
		if err != nil {
			klog.Fatal(err.Error())
		}
	}

	return name
}
