package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
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
	port            = 8201
	bindAddr string = "0.0.0.0"
	confPath string
)

func serverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "vault-sidecar server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runServerCmd(); err != nil {
				klog.ErrorS(err, "run server failed")
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&bindAddr, "bind-addr", "0.0.0.0", "the IP address on which to listen")
	cmd.Flags().IntVar(&port, "port", 8201, "listen http/metrics port")
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

	addr := net.JoinHostPort(bindAddr, getPort())
	klog.InfoS("listening for requests and metrics", "addr", addr)
	confIn, err := os.ReadFile(confPath)
	if err != nil {
		return err
	}
	conf := VaultConf{}
	yaml.Unmarshal(confIn, &conf)

	go func() {
		tick := time.NewTicker(time.Second * 5)
		defer tick.Stop()

		for range tick.C {
			klog.InfoS("try unseal vault")
			if err := tryUnseal(conf); err != nil {
				klog.Warningf("unseal vault failed,err: %s", err)
			}

			// already unsealed
			if err := checkVaultStatus(); err != nil {
				klog.InfoS("check vault status not ready", "reason", err)
				continue
			}

			klog.Info("unseal vault done")
			return
		}
	}()

	return http.ListenAndServe(addr, r)
}

// HealthyHandler Healthz 接口
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// HealthyHandler 健康检查
func HealthyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
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

	w.Write([]byte("OK"))
}

func tryUnseal(conf VaultConf) error {
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
	if len(conf.Vault.UnsealKeys) == 0 {
		return fmt.Errorf("empty unseal keys")
	}

	for idx, k := range conf.Vault.UnsealKeys {
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
