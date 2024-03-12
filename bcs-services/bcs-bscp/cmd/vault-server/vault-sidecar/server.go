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

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

var (
	port     = 8202
	bindAddr = "0.0.0.0"
	confPath string
)

func serverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "vault sidecar server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runServerCmd(); err != nil {
				klog.ErrorS(err, "run server failed")
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&bindAddr, "bind-addr", "0.0.0.0", "the IP address on which to listen")
	cmd.Flags().IntVar(&port, "port", 8202, "listen http/metrics port")
	cmd.Flags().StringVar(&confPath, "config", "", "config path")
	return cmd
}

func getPort() string {
	p := os.Getenv("PORT")
	if p != "" {
		return p
	}

	return strconv.Itoa(port)
}

func runServerCmd() error {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 注册 HTTP 请求
	r.Get("/-/healthy", HealthyHandler)
	r.Get("/-/ready", ReadyHandler)
	r.Get("/healthz", HealthzHandler)

	confIn, err := os.ReadFile(confPath)
	if err != nil {
		return err
	}
	conf := vaultConf{}
	if e := yaml.Unmarshal(confIn, &conf); e != nil {
		return e
	}

	plugins, err := getPlugins(conf)
	if err != nil {
		return err
	}

	// try auto unseal
	go func() {
		tick := time.NewTicker(time.Second * 5)
		defer tick.Stop()

		for range tick.C {
			// already unsealed
			if err := checkVaultStatus(); err == nil {
				klog.InfoS("check vault status already initialized and unsealed, continue")
				continue
			}

			klog.InfoS("try unseal vault")
			if err := tryUnseal(conf); err != nil {
				klog.Warningf("unseal vault failed, err: %s", err)
			}

			klog.Info("unseal vault done")
		}
	}()

	// try auto register plugin
	go func() {
		tick := time.NewTicker(time.Second * 5)
		defer tick.Stop()

		for range tick.C {
			klog.InfoS("try register plugin")

			// ensure already unsealed
			if err := checkVaultStatus(); err != nil {
				klog.InfoS("check vault status not ready", "reason", err)
				continue
			}

			if err := tryRegisterPlugin(conf, plugins); err != nil {
				klog.Warningf("register failed, err: %s", err)
				continue
			}

			klog.InfoS("register plugin done", "plugins", plugins)
			return
		}
	}()

	addr := net.JoinHostPort(bindAddr, getPort())
	klog.InfoS("listening for requests and metrics", "addr", addr)
	return http.ListenAndServe(addr, r)
}

// HealthzHandler Healthz 接口
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK")) //nolint
}

// HealthyHandler 健康检查
func HealthyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK")) //nolint
}

func checkVaultStatus() error {
	c, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	ok, err := c.Sys().InitStatusWithContext(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("not initialized")
	}

	status, err := c.Sys().SealStatusWithContext(ctx)
	if err != nil {
		return err
	}

	if status.Sealed {
		return fmt.Errorf("sealed")
	}

	return nil
}

// ReadyHandler 健康检查
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	if err := checkVaultStatus(); err != nil {
		klog.InfoS("check vault status not ready", "reason", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Write([]byte("OK")) //nolint
}

// tryUnseal auto unseal by keys
func tryUnseal(conf vaultConf) error {
	c, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	status, err := c.Sys().SealStatusWithContext(ctx)
	if err != nil {
		return err
	}
	if !status.Initialized {
		return fmt.Errorf("not initialized")
	}
	if !status.Sealed {
		return nil
	}
	if len(conf.UnsealKeys) == 0 {
		return fmt.Errorf("empty unseal keys")
	}

	for idx, k := range conf.UnsealKeys {
		s, err := c.Sys().UnsealWithContext(ctx, k)
		if err != nil {
			klog.InfoS(fmt.Sprintf("unseal with key %d, failed", idx))
			continue
		}

		p := s.Progress
		if p == 0 {
			p = s.T
		}
		klog.InfoS(fmt.Sprintf("unseal with key %d, progress | %d/%d", idx, p, s.T))

		if !s.Sealed {
			return nil
		}
	}

	return fmt.Errorf("unseal with all keys failed")
}

func getPlugins(conf vaultConf) (map[string]string, error) {
	dir, err := os.ReadDir(conf.PluginDir)
	if err != nil {
		return nil, err
	}

	// plugins name:sha256 prepare for register
	plugins := map[string]string{}
	for _, v := range dir {
		if v.IsDir() {
			continue
		}
		info, err := v.Info()
		if err != nil {
			return nil, err
		}

		f, err := os.Open(path.Join(conf.PluginDir, info.Name()))
		if err != nil {
			return nil, err
		}
		defer f.Close() //nolint

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return nil, err
		}
		plugins[info.Name()] = hex.EncodeToString(h.Sum(nil))
	}

	return plugins, nil
}

// tryRegisterPlugin auto register plugin in pluginDir
func tryRegisterPlugin(conf vaultConf, plugins map[string]string) error {
	c, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}
	c.SetToken(conf.RootToken)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	for name, hash := range plugins {
		pluginInput := &api.RegisterPluginInput{
			Name:    name,
			Type:    api.PluginTypeSecrets,
			SHA256:  hash,
			Command: name,
		}
		if err := c.Sys().RegisterPluginWithContext(ctx, pluginInput); err != nil {
			return err
		}
	}

	return nil
}
