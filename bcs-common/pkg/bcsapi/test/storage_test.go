package test

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"testing"
)

func TestStorageCli_QueryK8SGameDeployment(t *testing.T) {
	tlsconfig, err := ssl.ClientTslConfVerity(
		"xxx",
		"xxx",
		"xxx",
		"xxx")

	if err != nil {
		t.Errorf("ssl.ClientTslConfVerity err: %v", err)
	}

	config := &bcsapi.Config{
		Hosts:     []string{"xxx:xxx"},
		TLSConfig: tlsconfig,
	}

	client := bcsapi.NewClient(config)
	s := client.Storage()
	gamedeployments, err := s.QueryK8SGameDeployment("xxx")
	if err != nil {
		return
	}
	t.Logf("gamedeployments : %v", gamedeployments)

	for _, ns := range gamedeployments {
		t.Log(ns)
	}
}
