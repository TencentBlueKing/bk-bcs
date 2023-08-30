package main

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"
)

const (
	initialized     = "Initialized"
	secretShares    = 5 // Number of key shares to split the generated root key into
	secretThreshold = 3 // Number of key shares required to reconstruct the root key
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "take vault init operator",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(runVaultInitCmd())
		},
	}

	return cmd
}

var secret = `global:
  vault:
    unsealKeys:
    {{- range .KeysB64 }}
      - {{ . }}
    {{- end }}
    rootToken: {{ .RootToken }}`

// VaultConf ..
type VaultConf struct {
	UnsealKeys      []string `yaml:"unsealKeys"`
	RootToken       string   `yaml:"rootToken"`
	SecretShares    int      `yaml:"secretShares"`
	SecretThreshold int      `yaml:"secretThreshold"`
}

// runVaultInitCmd 初始化
// 如果已经初始化, 返回 Initialized
// 如果未初始化, 返回结构化的 secret
// 其他返回错误信息
func runVaultInitCmd() string {
	c, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err.Error()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ok, err := c.Sys().InitStatusWithContext(ctx)
	if err != nil {
		return err.Error()
	}

	if ok {
		return initialized
	}

	resp, err := c.Sys().InitWithContext(ctx, &api.InitRequest{
		SecretShares:    secretShares,
		SecretThreshold: secretThreshold,
	})
	if err != nil {
		return err.Error()
	}

	tmpl, err := template.New("").Parse(secret)
	if err != nil {
		return err.Error()
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, resp); err != nil {
		return err.Error()
	}

	return tpl.String()
}
