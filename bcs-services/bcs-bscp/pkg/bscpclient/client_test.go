package bscpclient

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"bscp.io/pkg/criteria/uuid"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/tools"
)

var (
	// cli is api request client.
	cli *Client
	// logCfg is log config
	logCfg LogConfig
)

// LogConfig is log config
type LogConfig struct {
	// Verbosity is verbosity of log
	Verbosity uint
}

// SetLogger set logger
func SetLogger(logCfg LogConfig) {
	logs.InitLogger(
		logs.LogConfig{
			ToStdErr:       true,
			LogLineMaxSize: 2,
			Verbosity:      logCfg.Verbosity,
		},
	)
}

func init() {
	var clientCfg Config

	flag.StringVar(&clientCfg.ApiHost, "api-host", "http://127.0.0.1:8080", "api http server address")
	flag.UintVar(&logCfg.Verbosity, "log-verbosity", 5, "log verbosity")

	SetLogger(logCfg)

	var err error
	if cli, err = NewClient(clientCfg); err != nil {
		fmt.Printf("new client err: %v", err)
		os.Exit(0)
	}
}

func TestContentUpload(t *testing.T) {
	header := Header(uuid.UUID())
	bizID := uint32(2)
	tmplSpaceID := uint32(1)
	content := "test"
	sign := tools.SHA256(content)
	resp, err := cli.ApiClient.Content.Upload(context.Background(), header, bizID, 0, tmplSpaceID, sign, content)
	if err != nil {
		t.Errorf("upload content err: %v", err)
	}
	t.Logf("upload content resp: %v", resp)
}

func TestContentDownload(t *testing.T) {
	header := Header(uuid.UUID())
	bizID := uint32(2)
	tmplSpaceID := uint32(1)
	content := "test"
	sign := tools.SHA256(content)
	resp, err := cli.ApiClient.Content.Download(context.Background(), header, bizID, 0, tmplSpaceID, sign)
	if err != nil {
		t.Errorf("download content err: %v", err)
	}
	t.Logf("download content is: %s", resp)
}
